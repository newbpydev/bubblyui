// Package main provides the entry point for the basic profiler example.
//
// This example demonstrates:
// - Using the BubblyUI profiler package for performance monitoring
// - Composable architecture with UseProfiler
// - Multi-pane focus management
// - Live metric updates with UseInterval
// - Dynamic key bindings based on focus state
// - Component composition following BubblyUI patterns
//
// Run with:
//
//	go run ./cmd/examples/11-profiler/basic
package main

import (
	"fmt"
	"os"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

func main() {
	fmt.Println("🔬 BubblyUI Performance Profiler Example")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("")
	fmt.Println("Controls:")
	fmt.Println("  [Tab]     Switch focus between panels")
	fmt.Println("  [Space]   Start/Stop profiler (when Controls focused)")
	fmt.Println("  [r]       Reset metrics")
	fmt.Println("  [e]       Export report to HTML")
	fmt.Println("  [q]       Quit")
	fmt.Println("")
	fmt.Println("Starting application...")
	fmt.Println("")

	// Create app component
	app, err := CreateApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating app: %v\n", err)
		os.Exit(1)
	}

	// bubbly.Run() - Zero boilerplate! No manual Init/Update/View
	if err := bubbly.Run(app, bubbly.WithAltScreen()); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ Profiler example completed successfully!")
}
