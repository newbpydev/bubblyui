package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

func main() {
	// Enable dev tools
	devtools.Enable()

	fmt.Println("🔍 Dev Tools Example 02: Component Inspection")
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("")
	fmt.Println("This example demonstrates:")
	fmt.Println("  • Multi-level component hierarchy")
	fmt.Println("  • Component tree navigation")
	fmt.Println("  • State inspection across components")
	fmt.Println("  • Reactive state updates")
	fmt.Println("")
	fmt.Println("Press F12 to toggle dev tools")
	fmt.Println("Use ↑/↓ to navigate todos")
	fmt.Println("Press Space to toggle completion")
	fmt.Println("Press ctrl+c to quit")
	fmt.Println("")

	// Create app component
	app, err := CreateApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating app: %v\n", err)
		os.Exit(1)
	}

	// Wrap and run with Bubbletea
	p := tea.NewProgram(bubbly.Wrap(app), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}
