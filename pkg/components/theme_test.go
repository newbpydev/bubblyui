package components

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestDefaultTheme(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected lipgloss.Color
	}{
		{"Primary color", "primary", lipgloss.Color("63")},
		{"Secondary color", "secondary", lipgloss.Color("240")},
		{"Success color", "success", lipgloss.Color("46")},
		{"Warning color", "warning", lipgloss.Color("226")},
		{"Danger color", "danger", lipgloss.Color("196")},
		{"Info color", "info", lipgloss.Color("39")},
		{"Background color", "background", lipgloss.Color("235")},
		{"Foreground color", "foreground", lipgloss.Color("255")},
		{"Muted color", "muted", lipgloss.Color("240")},
		{"Border color", "borderColor", lipgloss.Color("240")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual lipgloss.Color
			switch tt.field {
			case "primary":
				actual = DefaultTheme.Primary
			case "secondary":
				actual = DefaultTheme.Secondary
			case "success":
				actual = DefaultTheme.Success
			case "warning":
				actual = DefaultTheme.Warning
			case "danger":
				actual = DefaultTheme.Danger
			case "info":
				actual = DefaultTheme.Info
			case "background":
				actual = DefaultTheme.Background
			case "foreground":
				actual = DefaultTheme.Foreground
			case "muted":
				actual = DefaultTheme.Muted
			case "borderColor":
				actual = DefaultTheme.BorderColor
			}
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestDefaultTheme_Spacing(t *testing.T) {
	assert.Equal(t, 1, DefaultTheme.Padding, "Default padding should be 1")
	assert.Equal(t, 1, DefaultTheme.Margin, "Default margin should be 1")
	assert.Equal(t, 1, DefaultTheme.Radius, "Default radius should be 1")
}

func TestDefaultTheme_Border(t *testing.T) {
	assert.NotNil(t, DefaultTheme.Border, "Border should not be nil")
	// Border is RoundedBorder by default
	assert.Equal(t, lipgloss.RoundedBorder(), DefaultTheme.Border)
}

func TestDarkTheme(t *testing.T) {
	// Verify DarkTheme has lighter colors than DefaultTheme
	assert.Equal(t, lipgloss.Color("75"), DarkTheme.Primary)
	assert.Equal(t, lipgloss.Color("245"), DarkTheme.Secondary)
	assert.Equal(t, lipgloss.Color("234"), DarkTheme.Background)
}

func TestLightTheme(t *testing.T) {
	// Verify LightTheme has darker colors for light backgrounds
	assert.Equal(t, lipgloss.Color("27"), LightTheme.Primary)
	assert.Equal(t, lipgloss.Color("240"), LightTheme.Secondary)
	assert.Equal(t, lipgloss.Color("255"), LightTheme.Background)
	assert.Equal(t, lipgloss.Color("235"), LightTheme.Foreground)
}

func TestHighContrastTheme(t *testing.T) {
	// Verify HighContrastTheme uses maximum contrast
	assert.Equal(t, lipgloss.Color("15"), HighContrastTheme.Primary)
	assert.Equal(t, lipgloss.Color("0"), HighContrastTheme.Background)
	assert.Equal(t, lipgloss.Color("15"), HighContrastTheme.Foreground)
	assert.Equal(t, 0, HighContrastTheme.Radius, "High contrast should use sharp borders")
	assert.Equal(t, lipgloss.NormalBorder(), HighContrastTheme.Border)
}

func TestTheme_GetVariantColor(t *testing.T) {
	tests := []struct {
		name     string
		variant  Variant
		expected lipgloss.Color
	}{
		{"Primary variant", VariantPrimary, DefaultTheme.Primary},
		{"Secondary variant", VariantSecondary, DefaultTheme.Secondary},
		{"Success variant", VariantSuccess, DefaultTheme.Success},
		{"Warning variant", VariantWarning, DefaultTheme.Warning},
		{"Danger variant", VariantDanger, DefaultTheme.Danger},
		{"Info variant", VariantInfo, DefaultTheme.Info},
		{"Unknown variant defaults to Primary", Variant("unknown"), DefaultTheme.Primary},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := DefaultTheme.GetVariantColor(tt.variant)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestTheme_GetBorderStyle(t *testing.T) {
	tests := []struct {
		name     string
		theme    Theme
		expected lipgloss.Border
	}{
		{
			name:     "Rounded border for radius > 0",
			theme:    Theme{Radius: 1},
			expected: lipgloss.RoundedBorder(),
		},
		{
			name:     "Normal border for radius = 0",
			theme:    Theme{Radius: 0},
			expected: lipgloss.NormalBorder(),
		},
		{
			name:     "Normal border for negative radius",
			theme:    Theme{Radius: -1},
			expected: lipgloss.NormalBorder(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.theme.GetBorderStyle()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestTheme_CustomTheme(t *testing.T) {
	// Test creating a custom theme
	customTheme := Theme{
		Primary:     lipgloss.Color("99"),
		Secondary:   lipgloss.Color("100"),
		Success:     lipgloss.Color("101"),
		Warning:     lipgloss.Color("102"),
		Danger:      lipgloss.Color("103"),
		Info:        lipgloss.Color("104"),
		Background:  lipgloss.Color("105"),
		Foreground:  lipgloss.Color("106"),
		Muted:       lipgloss.Color("107"),
		Border:      lipgloss.ThickBorder(),
		BorderColor: lipgloss.Color("108"),
		Padding:     2,
		Margin:      2,
		Radius:      0,
	}

	assert.Equal(t, lipgloss.Color("99"), customTheme.Primary)
	assert.Equal(t, 2, customTheme.Padding)
	assert.Equal(t, lipgloss.ThickBorder(), customTheme.Border)
	assert.Equal(t, lipgloss.NormalBorder(), customTheme.GetBorderStyle())
}

func TestAllThemes_HaveRequiredFields(t *testing.T) {
	themes := []struct {
		name  string
		theme Theme
	}{
		{"DefaultTheme", DefaultTheme},
		{"DarkTheme", DarkTheme},
		{"LightTheme", LightTheme},
		{"HighContrastTheme", HighContrastTheme},
	}

	for _, tt := range themes {
		t.Run(tt.name, func(t *testing.T) {
			// Verify all color fields are set (not empty)
			assert.NotEmpty(t, tt.theme.Primary, "Primary color should be set")
			assert.NotEmpty(t, tt.theme.Secondary, "Secondary color should be set")
			assert.NotEmpty(t, tt.theme.Success, "Success color should be set")
			assert.NotEmpty(t, tt.theme.Warning, "Warning color should be set")
			assert.NotEmpty(t, tt.theme.Danger, "Danger color should be set")
			assert.NotEmpty(t, tt.theme.Info, "Info color should be set")
			assert.NotEmpty(t, tt.theme.Background, "Background color should be set")
			assert.NotEmpty(t, tt.theme.Foreground, "Foreground color should be set")
			assert.NotEmpty(t, tt.theme.Muted, "Muted color should be set")
			assert.NotEmpty(t, tt.theme.BorderColor, "BorderColor should be set")

			// Verify border is set
			assert.NotNil(t, tt.theme.Border, "Border should not be nil")

			// Verify spacing values are non-negative
			assert.GreaterOrEqual(t, tt.theme.Padding, 0, "Padding should be non-negative")
			assert.GreaterOrEqual(t, tt.theme.Margin, 0, "Margin should be non-negative")
			assert.GreaterOrEqual(t, tt.theme.Radius, 0, "Radius should be non-negative")
		})
	}
}
