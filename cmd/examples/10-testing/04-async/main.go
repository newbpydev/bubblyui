package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

func main() {
	// Create mock API with realistic delays
	mockAPI := NewMockGitHubAPI()
	mockAPI.SetDelay(500*time.Millisecond, 500*time.Millisecond)

	// Create app component
	app, err := CreateApp(mockAPI)
	if err != nil {
		fmt.Printf("Error creating app: %v\n", err)
		os.Exit(1)
	}

	// Wrap and run with Bubbletea
	p := tea.NewProgram(
		bubbly.Wrap(app),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running app: %v\n", err)
		os.Exit(1)
	}
}
