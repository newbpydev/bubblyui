package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func main() {
	// Create a basic styled text using lipgloss
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		PaddingLeft(4).
		PaddingRight(4).
		PaddingTop(2).
		PaddingBottom(2)

	// Apply the style to text
	renderedText := style.Render("BubblyUI Demo - Hot Reload Test")

	// Get current time to show hot reload is working
	currentTime := time.Now().Format("15:04:05")
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#5AF78E"))
	renderedTime := timeStyle.Render(fmt.Sprintf("Current time: %s", currentTime))

	// Print to terminal
	fmt.Println(renderedText)
	fmt.Println()
	fmt.Println(renderedTime)
	fmt.Println("\nEdit this file and save to see hot reload in action!")
	fmt.Println("Press Ctrl+C to exit.")
}
