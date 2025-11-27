// Package main provides the AI Chat Demo application.
//
// This example demonstrates a ChatGPT-like terminal chat interface with:
//   - Scrollable message history
//   - Simulated AI responses with typing animation
//   - Responsive layout with collapsible sidebar
//   - Vim-style navigation (j/k for scroll)
//   - Input mode switching (i to type, Esc to scroll)
//
// The app responds to terminal resize events and adapts its layout.
// On narrow terminals, the sidebar is hidden automatically.
//
// Run with: go run ./cmd/examples/16-ai-chat-demo
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

	// Use bubbly.Run for zero boilerplate - NO MANUAL BUBBLETEA!
	if err := bubbly.Run(app, bubbly.WithAltScreen()); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
