package bubbly

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

// TestTheme_StructInitialization tests that Theme struct can be initialized with all fields
func TestTheme_StructInitialization(t *testing.T) {
	tests := []struct {
		name  string
		theme Theme
		want  Theme
	}{
		{
			name: "initialize with all fields",
			theme: Theme{
				Primary:    lipgloss.Color("35"),
				Secondary:  lipgloss.Color("99"),
				Muted:      lipgloss.Color("240"),
				Warning:    lipgloss.Color("220"),
				Error:      lipgloss.Color("196"),
				Success:    lipgloss.Color("35"),
				Background: lipgloss.Color("236"),
			},
			want: Theme{
				Primary:    lipgloss.Color("35"),
				Secondary:  lipgloss.Color("99"),
				Muted:      lipgloss.Color("240"),
				Warning:    lipgloss.Color("220"),
				Error:      lipgloss.Color("196"),
				Success:    lipgloss.Color("35"),
				Background: lipgloss.Color("236"),
			},
		},
		{
			name: "initialize with partial fields",
			theme: Theme{
				Primary:   lipgloss.Color("99"),
				Secondary: lipgloss.Color("120"),
			},
			want: Theme{
				Primary:    lipgloss.Color("99"),
				Secondary:  lipgloss.Color("120"),
				Muted:      lipgloss.Color(""),
				Warning:    lipgloss.Color(""),
				Error:      lipgloss.Color(""),
				Success:    lipgloss.Color(""),
				Background: lipgloss.Color(""),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want.Primary, tt.theme.Primary)
			assert.Equal(t, tt.want.Secondary, tt.theme.Secondary)
			assert.Equal(t, tt.want.Muted, tt.theme.Muted)
			assert.Equal(t, tt.want.Warning, tt.theme.Warning)
			assert.Equal(t, tt.want.Error, tt.theme.Error)
			assert.Equal(t, tt.want.Success, tt.theme.Success)
			assert.Equal(t, tt.want.Background, tt.theme.Background)
		})
	}
}

// TestDefaultTheme_Values tests that DefaultTheme has expected color values
func TestDefaultTheme_Values(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		got      lipgloss.Color
		expected lipgloss.Color
	}{
		{
			name:     "Primary is green (35)",
			field:    "Primary",
			got:      DefaultTheme.Primary,
			expected: lipgloss.Color("35"),
		},
		{
			name:     "Secondary is purple (99)",
			field:    "Secondary",
			got:      DefaultTheme.Secondary,
			expected: lipgloss.Color("99"),
		},
		{
			name:     "Muted is dark grey (240)",
			field:    "Muted",
			got:      DefaultTheme.Muted,
			expected: lipgloss.Color("240"),
		},
		{
			name:     "Warning is yellow (220)",
			field:    "Warning",
			got:      DefaultTheme.Warning,
			expected: lipgloss.Color("220"),
		},
		{
			name:     "Error is red (196)",
			field:    "Error",
			got:      DefaultTheme.Error,
			expected: lipgloss.Color("196"),
		},
		{
			name:     "Success is green (35)",
			field:    "Success",
			got:      DefaultTheme.Success,
			expected: lipgloss.Color("35"),
		},
		{
			name:     "Background is dark (236)",
			field:    "Background",
			got:      DefaultTheme.Background,
			expected: lipgloss.Color("236"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.got, "DefaultTheme.%s should be %s", tt.field, tt.expected)
		})
	}
}

// TestTheme_FieldTypes tests that Theme fields are correct types (lipgloss.Color)
func TestTheme_FieldTypes(t *testing.T) {
	theme := Theme{
		Primary:    lipgloss.Color("35"),
		Secondary:  lipgloss.Color("99"),
		Muted:      lipgloss.Color("240"),
		Warning:    lipgloss.Color("220"),
		Error:      lipgloss.Color("196"),
		Success:    lipgloss.Color("35"),
		Background: lipgloss.Color("236"),
	}

	// Verify we can assign to lipgloss.Color variables
	var primary lipgloss.Color = theme.Primary
	var secondary lipgloss.Color = theme.Secondary
	var muted lipgloss.Color = theme.Muted
	var warning lipgloss.Color = theme.Warning
	var errorColor lipgloss.Color = theme.Error
	var success lipgloss.Color = theme.Success
	var background lipgloss.Color = theme.Background

	// Verify values are correct
	assert.Equal(t, lipgloss.Color("35"), primary)
	assert.Equal(t, lipgloss.Color("99"), secondary)
	assert.Equal(t, lipgloss.Color("240"), muted)
	assert.Equal(t, lipgloss.Color("220"), warning)
	assert.Equal(t, lipgloss.Color("196"), errorColor)
	assert.Equal(t, lipgloss.Color("35"), success)
	assert.Equal(t, lipgloss.Color("236"), background)
}

// TestTheme_IsValueType tests that Theme is a value type (struct, not pointer)
func TestTheme_IsValueType(t *testing.T) {
	// Create original theme
	original := Theme{
		Primary:   lipgloss.Color("35"),
		Secondary: lipgloss.Color("99"),
	}

	// Copy by assignment (value semantics)
	copied := original

	// Modify copy
	copied.Primary = lipgloss.Color("120")
	copied.Secondary = lipgloss.Color("200")

	// Verify original is unchanged (value type behavior)
	assert.Equal(t, lipgloss.Color("35"), original.Primary, "Original should be unchanged")
	assert.Equal(t, lipgloss.Color("99"), original.Secondary, "Original should be unchanged")
	assert.Equal(t, lipgloss.Color("120"), copied.Primary, "Copy should be modified")
	assert.Equal(t, lipgloss.Color("200"), copied.Secondary, "Copy should be modified")
}

// TestTheme_ZeroValue tests that zero value theme is valid (no panics)
func TestTheme_ZeroValue(t *testing.T) {
	// Zero value theme (all fields are empty lipgloss.Color)
	var theme Theme

	// Should not panic when accessing fields
	assert.NotPanics(t, func() {
		_ = theme.Primary
		_ = theme.Secondary
		_ = theme.Muted
		_ = theme.Warning
		_ = theme.Error
		_ = theme.Success
		_ = theme.Background
	})

	// Zero value should be empty colors
	assert.Equal(t, lipgloss.Color(""), theme.Primary)
	assert.Equal(t, lipgloss.Color(""), theme.Secondary)
	assert.Equal(t, lipgloss.Color(""), theme.Muted)
	assert.Equal(t, lipgloss.Color(""), theme.Warning)
	assert.Equal(t, lipgloss.Color(""), theme.Error)
	assert.Equal(t, lipgloss.Color(""), theme.Success)
	assert.Equal(t, lipgloss.Color(""), theme.Background)
}

// TestTheme_CanBeUsedWithLipgloss tests that Theme colors work with Lipgloss styles
func TestTheme_CanBeUsedWithLipgloss(t *testing.T) {
	theme := DefaultTheme

	// Should not panic when used with Lipgloss
	assert.NotPanics(t, func() {
		_ = lipgloss.NewStyle().Foreground(theme.Primary)
		_ = lipgloss.NewStyle().Background(theme.Background)
		_ = lipgloss.NewStyle().Foreground(theme.Error).Bold(true)
		_ = lipgloss.NewStyle().
			Foreground(theme.Primary).
			Background(theme.Background).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Secondary)
	})
}

// TestTheme_Modification tests that themes can be modified after creation
func TestTheme_Modification(t *testing.T) {
	// Start with default theme
	theme := DefaultTheme

	// Modify specific colors
	theme.Primary = lipgloss.Color("120")
	theme.Background = lipgloss.Color("232")

	// Verify modifications
	assert.Equal(t, lipgloss.Color("120"), theme.Primary)
	assert.Equal(t, lipgloss.Color("232"), theme.Background)

	// Verify other fields unchanged
	assert.Equal(t, DefaultTheme.Secondary, theme.Secondary)
	assert.Equal(t, DefaultTheme.Muted, theme.Muted)
	assert.Equal(t, DefaultTheme.Warning, theme.Warning)
	assert.Equal(t, DefaultTheme.Error, theme.Error)
	assert.Equal(t, DefaultTheme.Success, theme.Success)
}

// TestDefaultTheme_IsConstant tests that DefaultTheme is accessible as a constant
func TestDefaultTheme_IsConstant(t *testing.T) {
	// Should be able to access DefaultTheme directly
	theme := DefaultTheme

	// Should have all fields populated
	assert.NotEqual(t, lipgloss.Color(""), theme.Primary)
	assert.NotEqual(t, lipgloss.Color(""), theme.Secondary)
	assert.NotEqual(t, lipgloss.Color(""), theme.Muted)
	assert.NotEqual(t, lipgloss.Color(""), theme.Warning)
	assert.NotEqual(t, lipgloss.Color(""), theme.Error)
	assert.NotEqual(t, lipgloss.Color(""), theme.Success)
	assert.NotEqual(t, lipgloss.Color(""), theme.Background)
}

// TestTheme_ConcurrentAccess tests that Theme can be safely read concurrently
func TestTheme_ConcurrentAccess(t *testing.T) {
	theme := DefaultTheme
	done := make(chan bool, 10)

	// Spawn 10 goroutines reading theme concurrently
	for i := 0; i < 10; i++ {
		go func() {
			// Read all fields
			_ = theme.Primary
			_ = theme.Secondary
			_ = theme.Muted
			_ = theme.Warning
			_ = theme.Error
			_ = theme.Success
			_ = theme.Background
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test passes if no race conditions detected
}
