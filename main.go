package main

import (
	"fmt"
	"os"

	"fulbito/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) > 1 {
		runQuick(os.Args[1:])
		return
	}

	p := tea.NewProgram(tui.NewRootModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error running fulbito:", err)
		os.Exit(1)
	}
}
