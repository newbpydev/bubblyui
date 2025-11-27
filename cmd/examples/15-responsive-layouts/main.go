// Package main provides the Responsive Layouts example application.
//
// This example demonstrates responsive layout patterns that adapt to terminal size:
//   - Automatic breakpoint detection (xs, sm, md, lg, xl)
//   - Collapsible sidebar on narrow screens
//   - Adaptive grid columns based on width
//   - Layout switching (horizontal ↔ vertical) based on breakpoint
//   - Minimum size enforcement to prevent broken layouts
//
// The app responds to terminal resize events (tea.WindowSizeMsg) and
// recalculates all layout parameters in real-time.
//
// Run with: go run ./cmd/examples/15-responsive-layouts
package main

import (
	"fmt"
	"os"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

func main() {
	app, err := CreateApp()
	if err != nil {
		fmt.Printf("Error creating app: %v\n", err)
		os.Exit(1)
	}

	// Use bubbly.Run for automatic command generation - ZERO BUBBLETEA!
	if err := bubbly.Run(app, bubbly.WithAltScreen()); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
