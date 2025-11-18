package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

func main() {
	app, err := CreateApp()
	if err != nil {
		fmt.Printf("Error creating app: %v\n", err)
		os.Exit(1)
	}

	// Use bubbly.Wrap for automatic command generation
	p := tea.NewProgram(bubbly.Wrap(app), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
