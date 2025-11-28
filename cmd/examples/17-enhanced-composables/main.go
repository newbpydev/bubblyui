// Package main provides the Enhanced Composables Demo application.
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

	if err := bubbly.Run(app, bubbly.WithAltScreen()); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
