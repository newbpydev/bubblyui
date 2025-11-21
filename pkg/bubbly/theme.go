package bubbly

import "github.com/charmbracelet/lipgloss"

// Theme defines a standard color palette for BubblyUI components.
// It provides semantic color names that can be consistently used across
// the component hierarchy via Provide/Inject pattern.
//
// Theme is a value type (struct), not a pointer, which means it can be
// safely copied and modified without affecting other instances.
//
// Example usage:
//
//	// In parent component
//	theme := bubbly.DefaultTheme
//	theme.Primary = lipgloss.Color("99") // Override primary color
//	ctx.ProvideTheme(theme)
//
//	// In child component
//	theme := ctx.UseTheme(bubbly.DefaultTheme)
//	style := lipgloss.NewStyle().Foreground(theme.Primary)
type Theme struct {
	// Primary is the main accent color (e.g., brand color).
	// Used for important actions, active elements, and primary UI components.
	Primary lipgloss.Color

	// Secondary is the secondary accent color.
	// Used for alternative actions and secondary importance elements.
	Secondary lipgloss.Color

	// Muted is used for less prominent text and UI elements.
	// Typically used for disabled states, subtle UI, and secondary text.
	Muted lipgloss.Color

	// Warning is used for warning messages and caution states.
	// Indicates non-blocking issues and informational alerts.
	Warning lipgloss.Color

	// Error is used for error messages and danger states.
	// Indicates critical issues, blocking errors, and danger actions.
	Error lipgloss.Color

	// Success is used for success messages and positive states.
	// Indicates completed actions, positive feedback, and confirmations.
	Success lipgloss.Color

	// Background is the default background color.
	// Used for container backgrounds, card backgrounds, and main UI background.
	Background lipgloss.Color
}

// DefaultTheme provides sensible defaults for the theme colors.
// Components can use this as a fallback when no parent provides a theme.
//
// Color choices:
//   - Primary/Success: Green (35) - positive, go-ahead actions
//   - Secondary: Purple (99) - alternative actions, secondary importance
//   - Muted: Dark grey (240) - subtle, less important elements
//   - Warning: Yellow (220) - caution, non-blocking alerts
//   - Error: Red (196) - critical issues, danger
//   - Background: Dark (236) - neutral background
//
// These colors work well in most terminal environments and provide
// good contrast for readability.
var DefaultTheme = Theme{
	Primary:    lipgloss.Color("35"),  // Green
	Secondary:  lipgloss.Color("99"),  // Purple
	Muted:      lipgloss.Color("240"), // Dark grey
	Warning:    lipgloss.Color("220"), // Yellow
	Error:      lipgloss.Color("196"), // Red
	Success:    lipgloss.Color("35"),  // Green
	Background: lipgloss.Color("236"), // Dark background
}
