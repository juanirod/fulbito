package tui

import (
	"fmt"
	"strings"

	"fulbito/internal/espn"
	"fulbito/internal/matchdata"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type detailModel struct {
	homeName string
	awayName string
	score    string
	rows     []matchdata.TimelineRow
}

func newDetailModel(event espn.Event) detailModel {
	if len(event.Competitions) == 0 {
		return detailModel{homeName: event.ShortName}
	}
	comp := event.Competitions[0]

	home, away := matchdata.HomeAway(comp)

	homeScore, awayScore := home.Score, away.Score
	if comp.Status.Type.State == "pre" {
		homeScore, awayScore = "-", "-"
	}

	return detailModel{
		homeName: home.Team.DisplayName,
		awayName: away.Team.DisplayName,
		score:    fmt.Sprintf("%s - %s", homeScore, awayScore),
		rows:     matchdata.BuildTimeline(comp.Details, home.Team.ID, away.Team.ID),
	}
}

func (m detailModel) Update(msg tea.Msg) (detailModel, tea.Cmd) {
	return m, nil
}

const detailColumnWidth = 28
const detailMinuteWidth = 8

var (
	detailHomeStyle     = lipgloss.NewStyle().Width(detailColumnWidth).Align(lipgloss.Right)
	detailMinuteStyle   = lipgloss.NewStyle().Width(detailMinuteWidth).Align(lipgloss.Center).Foreground(lipgloss.Color("240"))
	detailAwayStyle     = lipgloss.NewStyle().Width(detailColumnWidth).Align(lipgloss.Left)
	detailTeamNameStyle = lipgloss.NewStyle().Bold(true).Width(detailColumnWidth)
	detailScoreStyle    = lipgloss.NewStyle().Bold(true).Width(detailMinuteWidth).Align(lipgloss.Center)
	detailDividerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	detailBoxStyle      = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(2, 4)
)

const detailHelp = "\n\n(b/esc: back  q: quit)"

func (m detailModel) matchHeader() string {
	home := detailTeamNameStyle.Copy().Align(lipgloss.Right).Render(m.homeName)
	score := detailScoreStyle.Render(m.score)
	away := detailTeamNameStyle.Copy().Align(lipgloss.Left).Render(m.awayName)
	line := lipgloss.JoinHorizontal(lipgloss.Top, home, score, away)
	divider := detailDividerStyle.Render(strings.Repeat("─", lipgloss.Width(line)))
	return line + "\n" + divider
}

func (m detailModel) View() string {
	header := m.matchHeader()

	if len(m.rows) == 0 {
		content := header + "\n\n" + "No hay eventos detallados disponibles."
		return detailBoxStyle.Render(content) + detailHelp
	}

	var b strings.Builder
	for _, row := range m.rows {
		home := detailHomeStyle.Render(strings.Join(row.Home, "\n"))
		minute := detailMinuteStyle.Render(row.Minute)
		away := detailAwayStyle.Render(strings.Join(row.Away, "\n"))
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, home, minute, away))
		b.WriteString("\n")
	}

	content := header + "\n\n" + strings.TrimRight(b.String(), "\n")
	return detailBoxStyle.Render(content) + detailHelp
}
