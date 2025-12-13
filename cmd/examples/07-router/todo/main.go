// Package main demonstrates a Todo app with BubblyUI router.
//
// This example showcases:
//   - Zero-boilerplate with bubbly.Run()
//   - Router with multiple pages (home, add, detail)
//   - Route parameters (/todo/:id)
//   - Shared state via composables (UseTodos)
//   - Mode-based input handling (navigation vs input mode)
//   - Built-in components (Card, Badge)
//   - Conditional key bindings
//   - WithMessageHandler for text input
//
// # App Structure
//
//	todo/
//	├── main.go           # Entry point with bubbly.Run()
//	├── app.go            # Root component with router setup
//	├── composables/
//	│   └── use_todos.go  # Shared todo state management
//	└── pages/
//	    ├── home.go       # Home page (todo list)
//	    ├── add.go        # Add todo page
//	    └── detail.go     # Todo detail page
//
// # Running the Example
//
//	cd cmd/examples/07-router/todo
//	go run .
//
// # Keyboard Shortcuts
//
// Navigation Mode:
//   - ↑/k: Move up
//   - ↓/j: Move down
//   - Space: Toggle todo
//   - Enter: View detail / Submit
//   - a: Add new todo
//   - d: Delete todo
//   - b: Go back
//   - q: Quit
//
// Input Mode:
//   - Tab: Next field
//   - Shift+Tab: Previous field
//   - p: Cycle priority
//   - Enter: Submit
//   - Esc: Cancel
package main

import (
	"fmt"
	"os"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

func main() {
	// Print banner
	printBanner()

	// Create the app
	app, err := CreateApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating app: %v\n", err)
		os.Exit(1)
	}

	// Run with zero boilerplate!
	if err := bubbly.Run(app, bubbly.WithAltScreen()); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}

func printBanner() {
	fmt.Println()
	fmt.Println("  ____        _     _     _       _   _ ___ ")
	fmt.Println(" | __ ) _   _| |__ | |__ | |_   _| | | |_ _|")
	fmt.Println(" |  _ \\| | | | '_ \\| '_ \\| | | | | | | || | ")
	fmt.Println(" | |_) | |_| | |_) | |_) | | |_| | |_| || | ")
	fmt.Println(" |____/ \\__,_|_.__/|_.__/|_|\\__, |\\___/|___|")
	fmt.Println("                            |___/           ")
	fmt.Println()
	fmt.Println(" Router Todo Example")
	fmt.Println(" ====================")
	fmt.Println()
	fmt.Println(" This example demonstrates:")
	fmt.Println()
	fmt.Println("   - Zero-boilerplate with bubbly.Run()")
	fmt.Println("   - Router with multiple pages")
	fmt.Println("   - Route parameters (/todo/:id)")
	fmt.Println("   - Shared state via composables")
	fmt.Println("   - Mode-based input handling")
	fmt.Println("   - Built-in components (Card, Badge)")
	fmt.Println()
	fmt.Println(" Keyboard Shortcuts:")
	fmt.Println()
	fmt.Println("   ↑/k     - Move up")
	fmt.Println("   ↓/j     - Move down")
	fmt.Println("   Space   - Toggle todo")
	fmt.Println("   Enter   - View detail / Submit")
	fmt.Println("   a       - Add new todo")
	fmt.Println("   d       - Delete todo")
	fmt.Println("   b/Esc   - Go back")
	fmt.Println("   q       - Quit")
	fmt.Println()
}
