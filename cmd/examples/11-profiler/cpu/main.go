// Package main provides the entry point for the CPU profiler example.
//
// This example demonstrates:
// - Using the BubblyUI profiler package for CPU profiling
// - Composable architecture with UseCPUProfiler
// - Multi-pane focus management (Profile, Controls, Results)
// - State machine workflow (Idle → Profiling → Complete → Analyzed)
// - Dynamic key bindings based on state and focus
// - pprof integration for CPU analysis
//
// Run with:
//
//	go run ./cmd/examples/11-profiler/cpu
package main

import (
	"fmt"
	"os"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

func main() {
	fmt.Println("🔬 BubblyUI CPU Profiler Example")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("")
	fmt.Println("Controls:")
	fmt.Println("  [Tab]     Switch focus between panels")
	fmt.Println("  [Space]   Start/Stop profiling (when Controls focused)")
	fmt.Println("  [a]       Analyze results (when complete)")
	fmt.Println("  [r]       Reset profiler")
	fmt.Println("  [q]       Quit")
	fmt.Println("")
	fmt.Println("Workflow:")
	fmt.Println("  1. Press [Space] to start CPU profiling")
	fmt.Println("  2. Wait for desired duration")
	fmt.Println("  3. Press [Space] to stop profiling")
	fmt.Println("  4. Press [a] to analyze results")
	fmt.Println("  5. View hot functions in Results panel")
	fmt.Println("  6. Use 'go tool pprof <file>' for detailed analysis")
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

	fmt.Println("\n✅ CPU Profiler example completed successfully!")
}
