package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

func main() {
	// ✨ ZERO-CONFIG DEV TOOLS ENABLEMENT ✨
	// This is all you need to enable dev tools!
	devtools.Enable()

	fmt.Println("🎯 Dev Tools Example 01: Basic Enablement")
	fmt.Println("══════════════════════════════════════════")
	fmt.Println("")
	fmt.Println("Press F12 to toggle dev tools visibility")
	fmt.Println("Press i/d/r to increment/decrement/reset")
	fmt.Println("Press ctrl+c to quit")
	fmt.Println("")

	// Create app component
	app, err := CreateApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating app: %v\n", err)
		os.Exit(1)
	}

	// Wrap component for Bubbletea and run
	// bubbly.Wrap() provides automatic integration - zero boilerplate!
	p := tea.NewProgram(bubbly.Wrap(app), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}
