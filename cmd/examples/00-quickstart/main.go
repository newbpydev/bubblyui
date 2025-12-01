// Package main demonstrates the BubblyUI quickstart example.
//
// This example showcases the recommended patterns for building BubblyUI applications:
//
// # Clean Import Paths
//
// Using the new alias packages for cleaner imports:
//
//	import "github.com/newbpydev/bubblyui"              // Core types and functions
//	import "github.com/newbpydev/bubblyui/components"   // UI components
//	import "github.com/newbpydev/bubblyui/composables"  // Composable functions
//	import "github.com/newbpydev/bubblyui/devtools"     // Development tools
//	import "github.com/newbpydev/bubblyui/profiler"     // Performance profiling
//
// # Application Structure
//
//	00-quickstart/
//	├── main.go           # Entry point with DevTools & Profiler setup
//	├── app.go            # Root component with key bindings
//	├── composables/      # Reusable reactive logic
//	│   ├── use_tasks.go  # Task management composable
//	│   └── use_focus.go  # Focus management composable
//	└── components/       # UI components
//	    ├── task_list.go  # Task list display
//	    ├── task_input.go # New task input
//	    ├── task_stats.go # Statistics display
//	    └── help_panel.go # Keyboard shortcuts
//
// # Running the Example
//
//	cd cmd/examples/00-quickstart
//	go run .
//
// Press F12 or Ctrl+T to toggle DevTools during runtime.
package main

import (
	"flag"
	"fmt"
	"os"

	// Clean import paths using alias packages (NEW - Task 2.3)
	// NO BUBBLETEA IMPORTS - Use bubbly.Run() for zero boilerplate!
	"github.com/newbpydev/bubblyui/devtools"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/profiler"
)

func main() {
	// Command-line flags
	enableDevTools := flag.Bool("devtools", true, "Enable DevTools (press F12 to toggle)")
	enableProfiler := flag.Bool("profiler", false, "Enable performance profiler")
	profilerOutput := flag.String("profiler-output", "profile-report.html", "Profiler output file")
	flag.Parse()

	// Print banner
	printBanner()

	// =============================================================================
	// DevTools Integration (using clean import path)
	// =============================================================================
	if *enableDevTools {
		// Initialize DevTools data collector
		collector := devtools.NewDataCollector()
		devtools.SetCollector(collector)

		fmt.Println("DevTools: ENABLED (press F12 or Ctrl+T to toggle)")
	} else {
		devtools.Disable()
		fmt.Println("DevTools: Disabled")
	}

	// =============================================================================
	// Profiler Integration (using clean import path)
	// =============================================================================
	var prof *profiler.Profiler
	if *enableProfiler {
		prof = profiler.New(
			profiler.WithEnabled(true),
			profiler.WithSamplingRate(1.0), // 100% sampling for demo
		)
		if err := prof.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting profiler: %v\n", err)
		}
		fmt.Println("Profiler: ENABLED")
		defer func() {
			if err := prof.Stop(); err != nil {
				fmt.Fprintf(os.Stderr, "Error stopping profiler: %v\n", err)
			}
			// Generate report
			report := prof.GenerateReport()
			exporter := profiler.NewExporter()
			if err := exporter.ExportHTML(report, *profilerOutput); err != nil {
				fmt.Fprintf(os.Stderr, "Error exporting profiler report: %v\n", err)
			} else {
				fmt.Printf("Profiler report saved to: %s\n", *profilerOutput)
			}
		}()
	} else {
		fmt.Println("Profiler: Disabled (use --profiler to enable)")
	}

	fmt.Println()

	// =============================================================================
	// Create Application
	// =============================================================================
	app, err := CreateApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating app: %v\n", err)
		os.Exit(1)
	}

	// =============================================================================
	// Run Application - ZERO BOILERPLATE with bubbly.Run()!
	// =============================================================================
	// This is the BubblyUI way - NO tea.NewProgram, NO bubbly.Wrap()
	// Just bubbly.Run() with options!
	if err := bubbly.Run(app, bubbly.WithAltScreen()); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}

// printBanner prints the application banner with feature highlights.
func printBanner() {
	fmt.Println()
	fmt.Println("  ____        _     _     _       _   _ ___ ")
	fmt.Println(" | __ ) _   _| |__ | |__ | |_   _| | | |_ _|")
	fmt.Println(" |  _ \\| | | | '_ \\| '_ \\| | | | | | | || | ")
	fmt.Println(" | |_) | |_| | |_) | |_) | | |_| | |_| || | ")
	fmt.Println(" |____/ \\__,_|_.__/|_.__/|_|\\__, |\\___/|___|")
	fmt.Println("                            |___/           ")
	fmt.Println()
	fmt.Println(" Quickstart Example - Task Manager")
	fmt.Println(" ==================================")
	fmt.Println()
	fmt.Println(" This example demonstrates:")
	fmt.Println()
	fmt.Println("   - Clean import paths (github.com/newbpydev/bubblyui)")
	fmt.Println("   - Component architecture with props")
	fmt.Println("   - Composables for reusable logic")
	fmt.Println("   - DevTools integration (F12 to toggle)")
	fmt.Println("   - Profiler integration (--profiler flag)")
	fmt.Println("   - Built-in UI components (Card, Text)")
	fmt.Println("   - Reactive state management")
	fmt.Println("   - Lifecycle hooks")
	fmt.Println()
	fmt.Println(" Keyboard Shortcuts:")
	fmt.Println()
	fmt.Println("   Tab     - Switch focus between list and input")
	fmt.Println("   j/k     - Navigate tasks")
	fmt.Println("   Enter   - Toggle task / Submit input")
	fmt.Println("   a       - Add new task")
	fmt.Println("   d       - Delete task")
	fmt.Println("   f       - Cycle filter (all/active/done)")
	fmt.Println("   c       - Clear completed tasks")
	fmt.Println("   F12     - Toggle DevTools")
	fmt.Println("   q       - Quit")
	fmt.Println()
}
