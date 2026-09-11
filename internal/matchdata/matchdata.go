package matchdata

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"fulbito/internal/espn"
)

type MatchRow struct {
	Kickoff string
	Home    string
	Away    string
	Score   string
	Status  string
	Events  string
}

// ESPN's "date" field is ISO-8601 UTC but omits seconds (e.g. "2026-09-11T19:00Z"),
// which time.RFC3339 doesn't accept, hence the second layout.
func parseEventDate(dateStr string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02T15:04Z", dateStr)
}

func FormatKickoff(dateStr string) string {
	t, err := parseEventDate(dateStr)
	if err != nil {
		return "--:--"
	}
	return t.Local().Format("15:04")
}

func SortEventsByKickoff(events []espn.Event) {
	sort.SliceStable(events, func(i, j int) bool {
		ti, erri := parseEventDate(events[i].Date)
		tj, errj := parseEventDate(events[j].Date)
		if erri != nil || errj != nil {
			return events[i].Date < events[j].Date
		}
		return ti.Before(tj)
	})
}

func HomeAway(comp espn.Competition) (home, away espn.Competitor) {
	for _, c := range comp.Competitors {
		switch c.HomeAway {
		case "home":
			home = c
		case "away":
			away = c
		}
	}
	return home, away
}

func BuildMatchRow(event espn.Event) MatchRow {
	if len(event.Competitions) == 0 {
		return MatchRow{}
	}
	comp := event.Competitions[0]

	home, away := HomeAway(comp)

	homeScore, awayScore := home.Score, away.Score
	if comp.Status.Type.State == "pre" {
		homeScore, awayScore = "-", "-"
	}

	status := comp.Status.Type.ShortDetail
	if status == "" {
		status = "TBD"
	}

	var events []string
	for _, d := range comp.Details {
		if !d.ScoringPlay && !d.RedCard {
			continue
		}
		player := "Unknown"
		if len(d.AthletesInvolved) > 0 {
			player = d.AthletesInvolved[0].ShortName
		}
		events = append(events, fmt.Sprintf("%s %s - %s", d.Clock.DisplayValue, d.Type.Text, player))
	}

	return MatchRow{
		Kickoff: FormatKickoff(event.Date),
		Home:    home.Team.DisplayName,
		Away:    away.Team.DisplayName,
		Score:   fmt.Sprintf("%s : %s", homeScore, awayScore),
		Status:  status,
		Events:  strings.Join(events, "; "),
	}
}

type TimelineRow struct {
	Minute string
	Home   []string
	Away   []string
}

// ESPN's stoppage-time format has an apostrophe after both halves (e.g.
// "90'+8'"), not just at the end, so both must be stripped before splitting.
func parseClockOrder(display string) int {
	display = strings.ReplaceAll(display, "'", "")
	base, extra := display, ""
	if i := strings.Index(display, "+"); i >= 0 {
		base, extra = display[:i], display[i+1:]
	}
	baseN, _ := strconv.Atoi(base)
	extraN, _ := strconv.Atoi(extra)
	return baseN*100 + extraN
}

// Penalties and own goals have no assist by football convention, so the
// second athlete listed (if any) is only treated as an assist for regular
// goals.
func detailLines(d espn.Detail) []string {
	player := "Unknown"
	if len(d.AthletesInvolved) > 0 {
		player = d.AthletesInvolved[0].ShortName
	}

	switch {
	case d.ScoringPlay && d.OwnGoal:
		return []string{"⚽ (OG) " + player}
	case d.ScoringPlay && d.PenaltyKick:
		return []string{"⚽ (P) " + player}
	case d.ScoringPlay:
		lines := []string{"⚽ " + player}
		if len(d.AthletesInvolved) > 1 {
			lines = append(lines, "🅰️ "+d.AthletesInvolved[1].ShortName)
		}
		return lines
	case d.RedCard:
		return []string{"🟥 " + player}
	case d.YellowCard:
		return []string{"🟨 " + player}
	case d.Type.Text != "":
		return []string{d.Type.Text}
	default:
		return nil
	}
}

func BuildTimeline(details []espn.Detail, homeTeamID, awayTeamID string) []TimelineRow {
	sorted := make([]espn.Detail, len(details))
	copy(sorted, details)
	sort.SliceStable(sorted, func(i, j int) bool {
		return parseClockOrder(sorted[i].Clock.DisplayValue) > parseClockOrder(sorted[j].Clock.DisplayValue)
	})

	rows := make([]TimelineRow, 0, len(sorted))
	for _, d := range sorted {
		lines := detailLines(d)
		if len(lines) == 0 {
			continue
		}
		row := TimelineRow{Minute: d.Clock.DisplayValue}
		if d.Team.ID == homeTeamID {
			row.Home = lines
		} else {
			row.Away = lines
		}
		rows = append(rows, row)
	}
	return rows
}
