package main

import (
	"fmt"
	"os"
	"time"

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

	// Run with bubbly.Run() - async auto-detected, no tick wrapper needed!
	// The framework automatically detects WithAutoCommands(true) and enables
	// async refresh with 100ms interval. 80+ lines of boilerplate eliminated!
	if err := bubbly.Run(app, bubbly.WithAltScreen()); err != nil {
		fmt.Printf("Error running app: %v\n", err)
		os.Exit(1)
	}
}
