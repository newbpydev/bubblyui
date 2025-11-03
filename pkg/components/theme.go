package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// Theme defines the color scheme and styling properties for all components.
// It provides a consistent visual language across the component library.
//
// Themes can be customized and provided to components via the BubblyUI
// composition API's Provide/Inject mechanism:
//
//	Setup(func(ctx *bubbly.Context) {
//	    ctx.Provide("theme", CustomTheme)
//	})
//
// Components automatically inject the theme and use it for styling.
type Theme struct {
	// Primary color for main actions and emphasis.
	// Used for: primary buttons, active states, focus indicators.
	Primary lipgloss.Color

	// Secondary color for alternative actions.
	// Used for: secondary buttons, less prominent elements.
	Secondary lipgloss.Color

	// Success color for positive actions and states.
	// Used for: success messages, confirmation buttons, positive indicators.
	Success lipgloss.Color

	// Warning color for caution and attention.
	// Used for: warning messages, caution buttons, alerts.
	Warning lipgloss.Color

	// Danger color for destructive actions and errors.
	// Used for: error messages, delete buttons, critical alerts.
	Danger lipgloss.Color

	// Info color for informational content.
	// Used for: info messages, help text, neutral notifications.
	Info lipgloss.Color

	// Background color for component backgrounds.
	// Used for: panel backgrounds, card backgrounds, modal overlays.
	Background lipgloss.Color

	// Foreground color for primary text.
	// Used for: body text, labels, default content.
	Foreground lipgloss.Color

	// Muted color for secondary text and disabled states.
	// Used for: placeholders, disabled text, subtle elements.
	Muted lipgloss.Color

	// Border style for component borders.
	// Used for: input borders, card borders, dividers.
	Border lipgloss.Border

	// BorderColor for default border color.
	// Used for: neutral borders, dividers, separators.
	BorderColor lipgloss.Color

	// Padding is the default padding value for components.
	// Used for: internal spacing in buttons, cards, panels.
	Padding int

	// Margin is the default margin value for components.
	// Used for: spacing between components.
	Margin int

	// Radius for rounded corners (not directly supported in terminals,
	// but used to select appropriate border styles).
	// 0 = sharp borders, 1 = rounded borders.
	Radius int
}

// DefaultTheme provides a carefully crafted default color scheme
// optimized for terminal readability and accessibility.
//
// Color choices:
//   - Primary (75): Bright blue for main actions (>3:1 contrast)
//   - Secondary (240): Neutral gray for alternatives
//   - Success (46): Bright green for positive feedback
//   - Warning (226): Yellow for caution
//   - Danger (196): Red for errors and destructive actions
//   - Info (39): Cyan for informational content
//   - Background (235): Dark gray for backgrounds
//   - Foreground (255): White for primary text
//   - Muted (240): Gray for secondary text
//
// This theme works well in both light and dark terminal backgrounds
// with sufficient contrast for accessibility (WCAG AA compliant).
var DefaultTheme = Theme{
	Primary:     lipgloss.Color("75"),  // Bright blue (accessible contrast)
	Secondary:   lipgloss.Color("240"), // Neutral gray
	Success:     lipgloss.Color("46"),  // Bright green
	Warning:     lipgloss.Color("226"), // Yellow
	Danger:      lipgloss.Color("196"), // Red
	Info:        lipgloss.Color("39"),  // Cyan
	Background:  lipgloss.Color("235"), // Dark gray
	Foreground:  lipgloss.Color("255"), // White
	Muted:       lipgloss.Color("240"), // Gray
	Border:      lipgloss.RoundedBorder(),
	BorderColor: lipgloss.Color("240"), // Gray
	Padding:     1,
	Margin:      1,
	Radius:      1, // Rounded borders
}

// DarkTheme provides a theme optimized for dark terminal backgrounds.
// It uses higher contrast colors for better visibility on dark backgrounds.
var DarkTheme = Theme{
	Primary:     lipgloss.Color("75"),  // Lighter blue
	Secondary:   lipgloss.Color("245"), // Lighter gray
	Success:     lipgloss.Color("82"),  // Lighter green
	Warning:     lipgloss.Color("220"), // Lighter yellow
	Danger:      lipgloss.Color("203"), // Lighter red
	Info:        lipgloss.Color("51"),  // Lighter cyan
	Background:  lipgloss.Color("234"), // Very dark gray
	Foreground:  lipgloss.Color("255"), // White
	Muted:       lipgloss.Color("243"), // Medium gray
	Border:      lipgloss.RoundedBorder(),
	BorderColor: lipgloss.Color("243"), // Medium gray
	Padding:     1,
	Margin:      1,
	Radius:      1,
}

// LightTheme provides a theme optimized for light terminal backgrounds.
// It uses darker colors for better visibility on light backgrounds.
var LightTheme = Theme{
	Primary:     lipgloss.Color("27"),  // Dark blue
	Secondary:   lipgloss.Color("240"), // Dark gray
	Success:     lipgloss.Color("28"),  // Dark green
	Warning:     lipgloss.Color("136"), // Dark yellow
	Danger:      lipgloss.Color("160"), // Dark red
	Info:        lipgloss.Color("31"),  // Dark cyan
	Background:  lipgloss.Color("255"), // White
	Foreground:  lipgloss.Color("235"), // Very dark gray
	Muted:       lipgloss.Color("245"), // Light gray
	Border:      lipgloss.RoundedBorder(),
	BorderColor: lipgloss.Color("245"), // Light gray
	Padding:     1,
	Margin:      1,
	Radius:      1,
}

// HighContrastTheme provides maximum contrast for accessibility.
// Ideal for users with visual impairments or in bright environments.
var HighContrastTheme = Theme{
	Primary:     lipgloss.Color("15"),  // Bright white
	Secondary:   lipgloss.Color("250"), // Very light gray
	Success:     lipgloss.Color("10"),  // Bright green
	Warning:     lipgloss.Color("11"),  // Bright yellow
	Danger:      lipgloss.Color("9"),   // Bright red
	Info:        lipgloss.Color("14"),  // Bright cyan
	Background:  lipgloss.Color("0"),   // Black
	Foreground:  lipgloss.Color("15"),  // Bright white
	Muted:       lipgloss.Color("7"),   // Light gray
	Border:      lipgloss.NormalBorder(),
	BorderColor: lipgloss.Color("15"), // Bright white
	Padding:     1,
	Margin:      1,
	Radius:      0, // Sharp borders for clarity
}

// GetThemeFromContext retrieves the theme from component context with fallback to DefaultTheme
func GetThemeFromContext(ctx *bubbly.Context) Theme {
	theme := DefaultTheme
	if injected := ctx.Inject("theme", nil); injected != nil {
		if t, ok := injected.(Theme); ok {
			theme = t
		}
	}
	return theme
}

// Helper function for component setup to inject theme
func setupTheme(ctx *bubbly.Context) {
	theme := GetThemeFromContext(ctx)
	ctx.Expose("theme", theme)
}

// GetVariantColor returns the appropriate color for a given variant.
// This helper function maps variant types to theme colors.
func (t Theme) GetVariantColor(variant Variant) lipgloss.Color {
	switch variant {
	case VariantPrimary:
		return t.Primary
	case VariantSecondary:
		return t.Secondary
	case VariantSuccess:
		return t.Success
	case VariantWarning:
		return t.Warning
	case VariantDanger:
		return t.Danger
	case VariantInfo:
		return t.Info
	default:
		return t.Primary
	}
}

// GetBorderStyle returns the appropriate border style based on theme radius.
// Returns RoundedBorder for radius > 0, NormalBorder otherwise.
func (t Theme) GetBorderStyle() lipgloss.Border {
	if t.Radius > 0 {
		return lipgloss.RoundedBorder()
	}
	return lipgloss.NormalBorder()
}
