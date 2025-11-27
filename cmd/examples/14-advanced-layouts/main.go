// Package main provides the Advanced Layout System example application.
//
// This example demonstrates all layout components from the BubblyUI framework:
//   - Flex: Flexbox-style layout with justify/align options
//   - HStack: Horizontal stack layout
//   - VStack: Vertical stack layout
//   - Box: Generic container with padding/border
//   - Center: Centering layout for modals/dialogs
//   - Container: Width-constrained container
//   - Divider: Horizontal/vertical separators
//   - Spacer: Flexible space filler
//
// Run with: go run ./cmd/examples/14-advanced-layouts
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
