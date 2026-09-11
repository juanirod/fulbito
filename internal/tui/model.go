package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type focusTarget int

const (
	focusSidebar focusTarget = iota
	focusMain
)

const sidebarContentWidth = 22

// RoundedBorder (1) + Padding(1,2) on paneStyle add this much on top of content size.
const paneHorizontalOverhead = 6
const paneVerticalOverhead = 4
const helpLines = 2

var (
	focusedBorderColor   = lipgloss.Color("62")
	unfocusedBorderColor = lipgloss.Color("240")
)

func paneStyle(focused bool) lipgloss.Style {
	color := unfocusedBorderColor
	if focused {
		color = focusedBorderColor
	}
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(1, 2)
}

const rootHelp = "\nTab: switch focus  j/k: move  ←/→: change date  enter: match details  esc: close  q: quit"

type RootModel struct {
	sidebar   sidebarModel
	main      scoreboardModel
	detail    detailModel
	focus     focusTarget
	popupOpen bool
	width     int
	height    int

	paneContentWidth  int
	paneContentHeight int
}

func NewRootModel() RootModel {
	sidebar := newSidebarModel()
	title, targets := sidebar.selectedTarget()
	return RootModel{
		sidebar: sidebar,
		main:    newScoreboardModel(title, targets),
		focus:   focusSidebar,
	}
}

func (m RootModel) Init() tea.Cmd {
	return m.main.Init()
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

		innerHeight := msg.Height - helpLines - paneVerticalOverhead
		if innerHeight < 5 {
			innerHeight = 5
		}
		sidebarBoxWidth := sidebarContentWidth + paneHorizontalOverhead
		mainContentWidth := msg.Width - sidebarBoxWidth - paneHorizontalOverhead
		if mainContentWidth < 30 {
			mainContentWidth = 30
		}

		m.paneContentWidth = mainContentWidth
		m.paneContentHeight = innerHeight

		m.sidebar.list.SetSize(sidebarContentWidth, innerHeight)
		m.main.table.SetWidth(mainContentWidth)
		m.main.table.SetHeight(innerHeight)
		return m, nil
	case matchesMsg, tickMsg:
		newMain, cmd := m.main.Update(msg)
		m.main = newMain
		return m, cmd
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		if m.popupOpen {
			switch key {
			case "esc", "b":
				m.popupOpen = false
			case "q":
				return m, tea.Quit
			}
			return m, nil
		}
		switch key {
		case "q":
			return m, tea.Quit
		case "tab":
			if m.focus == focusSidebar {
				m.focus = focusMain
			} else {
				m.focus = focusSidebar
			}
			return m, nil
		}

		if m.focus == focusSidebar {
			if key == "enter" {
				m.focus = focusMain
				return m, nil
			}
			prevIndex := m.sidebar.list.Index()
			newSidebar, cmd := m.sidebar.Update(msg)
			m.sidebar = newSidebar
			if m.sidebar.list.Index() != prevIndex {
				title, targets := m.sidebar.selectedTarget()
				m.main = newScoreboardModel(title, targets)
				return m, m.main.Init()
			}
			return m, cmd
		}

		newMain, cmd := m.main.Update(msg)
		m.main = newMain
		if m.main.selectedMatch != nil {
			m.detail = newDetailModel(*m.main.selectedMatch)
			m.main.selectedMatch = nil
			m.popupOpen = true
		}
		return m, cmd
	}
	return m, nil
}

func (m RootModel) View() string {
	if m.popupOpen {
		w, h := m.width, m.height
		if w == 0 {
			w = 100
		}
		if h == 0 {
			h = 30
		}
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, m.detail.View())
	}

	sidebarStyle := paneStyle(m.focus == focusSidebar)
	mainStyle := paneStyle(m.focus == focusMain)
	if m.paneContentHeight > 0 {
		sidebarStyle = sidebarStyle.Height(m.paneContentHeight)
		mainStyle = mainStyle.Height(m.paneContentHeight)
	}
	if m.paneContentWidth > 0 {
		mainStyle = mainStyle.Width(m.paneContentWidth)
	}

	sidebar := sidebarStyle.Render(m.sidebar.View())
	main := mainStyle.Render(m.main.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, main) + rootHelp
}
