package tui

import (
	"fmt"

	"fulbito/internal/leagues"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type leagueItem struct {
	title   string
	targets []leagues.League
}

func (i leagueItem) Title() string { return i.title }
func (i leagueItem) Description() string {
	if len(i.targets) == 1 {
		return i.targets[0].Slug
	}
	return fmt.Sprintf("%d leagues", len(i.targets))
}
func (i leagueItem) FilterValue() string { return i.title }

type sidebarModel struct {
	list list.Model
}

func newSidebarModel() sidebarModel {
	items := make([]list.Item, 0, len(leagues.All)+1)
	items = append(items, leagueItem{title: "Partidos del día", targets: leagues.All})
	for _, l := range leagues.All {
		items = append(items, leagueItem{title: l.Name, targets: []leagues.League{l}})
	}

	sidebarList := list.New(items, list.NewDefaultDelegate(), 0, 0)
	sidebarList.Title = "Leagues"
	sidebarList.SetShowHelp(false)
	sidebarList.SetShowStatusBar(false)
	sidebarList.SetFilteringEnabled(false)
	return sidebarModel{list: sidebarList}
}

func (m sidebarModel) selectedTarget() (title string, targets []leagues.League) {
	if item, ok := m.list.SelectedItem().(leagueItem); ok {
		return item.title, item.targets
	}
	return leagues.All[0].Name, []leagues.League{leagues.All[0]}
}

func (m sidebarModel) Update(msg tea.Msg) (sidebarModel, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m sidebarModel) View() string {
	return m.list.View()
}
