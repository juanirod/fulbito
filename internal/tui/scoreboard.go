package tui

import (
	"fmt"
	"sync"
	"time"

	"fulbito/internal/espn"
	"fulbito/internal/leagues"
	"fulbito/internal/matchdata"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// ESPN's public scoreboard doesn't update faster than this in practice;
// 20s keeps scores fresh without hammering the endpoint.
const pollInterval = 20 * time.Second

// Switching the sidebar selection replaces the scoreboard model outright, so a
// matchesMsg/tickMsg already in flight for the previous selection can still arrive
// afterwards. Tagging messages with the selection's key (its sidebar title, which
// is unique) lets the new model recognize and discard those stale deliveries.
type matchesMsg struct {
	key    string
	events []espn.Event
	rows   []matchdata.MatchRow
	err    error
}

type tickMsg struct {
	key string
	t   time.Time
}

func fetchMatchesCmd(key string, targets []leagues.League, dateStr string) tea.Cmd {
	return func() tea.Msg {
		type fetchResult struct {
			events []espn.Event
			err    error
		}
		results := make([]fetchResult, len(targets))

		var wg sync.WaitGroup
		for i, league := range targets {
			wg.Add(1)
			go func(i int, slug string) {
				defer wg.Done()
				resp, err := espn.Fetch(slug, dateStr)
				if err != nil {
					results[i] = fetchResult{err: err}
					return
				}
				results[i] = fetchResult{events: resp.Events}
			}(i, league.Slug)
		}
		wg.Wait()

		var events []espn.Event
		var firstErr error
		for _, r := range results {
			if r.err != nil {
				if firstErr == nil {
					firstErr = r.err
				}
				continue
			}
			events = append(events, r.events...)
		}
		if len(events) == 0 && firstErr != nil {
			return matchesMsg{key: key, err: firstErr}
		}

		matchdata.SortEventsByKickoff(events)
		rows := make([]matchdata.MatchRow, len(events))
		for i, event := range events {
			rows[i] = matchdata.BuildMatchRow(event)
		}
		return matchesMsg{key: key, events: events, rows: rows}
	}
}

func tickCmd(key string) tea.Cmd {
	return tea.Tick(pollInterval, func(t time.Time) tea.Msg {
		return tickMsg{key: key, t: t}
	})
}

func shiftAndFormatDate(base time.Time, days int) string {
	return base.AddDate(0, 0, days).Format("20060102")
}

var weekdaysEs = map[time.Weekday]string{
	time.Sunday:    "Dom",
	time.Monday:    "Lun",
	time.Tuesday:   "Mar",
	time.Wednesday: "Mié",
	time.Thursday:  "Jue",
	time.Friday:    "Vie",
	time.Saturday:  "Sáb",
}

var monthsEs = map[time.Month]string{
	time.January:   "Ene",
	time.February:  "Feb",
	time.March:     "Mar",
	time.April:     "Abr",
	time.May:       "May",
	time.June:      "Jun",
	time.July:      "Jul",
	time.August:    "Ago",
	time.September: "Sep",
	time.October:   "Oct",
	time.November:  "Nov",
	time.December:  "Dic",
}

func formatDateEs(t time.Time) string {
	return fmt.Sprintf("%s %d %s %d", weekdaysEs[t.Weekday()], t.Day(), monthsEs[t.Month()], t.Year())
}

type scoreboardModel struct {
	title         string
	targets       []leagues.League
	date          time.Time
	table         table.Model
	loading       bool
	err           error
	matches       []espn.Event
	selectedMatch *espn.Event
}

func newScoreboardModel(title string, targets []leagues.League) scoreboardModel {
	columns := []table.Column{
		{Title: "Time", Width: 7},
		{Title: "Home", Width: 20},
		{Title: "Score", Width: 9},
		{Title: "Away", Width: 20},
		{Title: "Status", Width: 22},
		{Title: "Key Events", Width: 40},
	}
	t := table.New(table.WithColumns(columns), table.WithHeight(15), table.WithFocused(true))
	return scoreboardModel{title: title, targets: targets, date: time.Now(), table: t, loading: true}
}

func (m scoreboardModel) Init() tea.Cmd {
	return tea.Batch(fetchMatchesCmd(m.title, m.targets, shiftAndFormatDate(m.date, 0)), tickCmd(m.title))
}

func (m scoreboardModel) Update(msg tea.Msg) (scoreboardModel, tea.Cmd) {
	switch msg := msg.(type) {
	case matchesMsg:
		if msg.key != m.title {
			return m, nil
		}
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.matches = msg.events
			m.table.SetRows(rowsToTableRows(msg.rows))
		}
		return m, nil
	case tickMsg:
		if msg.key != m.title {
			return m, nil
		}
		return m, tea.Batch(fetchMatchesCmd(m.title, m.targets, shiftAndFormatDate(m.date, 0)), tickCmd(m.title))
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			m.date = m.date.AddDate(0, 0, -1)
			m.loading = true
			return m, fetchMatchesCmd(m.title, m.targets, shiftAndFormatDate(m.date, 0))
		case "right", "l":
			m.date = m.date.AddDate(0, 0, 1)
			m.loading = true
			return m, fetchMatchesCmd(m.title, m.targets, shiftAndFormatDate(m.date, 0))
		case "enter":
			if idx := m.table.Cursor(); idx >= 0 && idx < len(m.matches) {
				match := m.matches[idx]
				m.selectedMatch = &match
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func rowsToTableRows(rows []matchdata.MatchRow) []table.Row {
	out := make([]table.Row, len(rows))
	for i, r := range rows {
		out[i] = table.Row{r.Kickoff, r.Home, r.Score, r.Away, r.Status, r.Events}
	}
	return out
}

func (m scoreboardModel) View() string {
	header := fmt.Sprintf("%s — %s (refreshing every 20s)\n\n", m.title, formatDateEs(m.date))
	if m.loading {
		return header + "Loading matches..."
	}
	if m.err != nil {
		return header + fmt.Sprintf("Error fetching matches: %v", m.err)
	}
	return header + m.table.View()
}
