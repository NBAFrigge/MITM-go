package main

import (
	"fmt"
	"os"

	tui "httpDebugger/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	model := tui.NewModel()

	p := tea.NewProgram(&model, tea.WithAltScreen())
	model.SetProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
