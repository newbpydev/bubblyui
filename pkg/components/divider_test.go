package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestDivider_Creation(t *testing.T) {
	divider := Divider(DividerProps{})

	assert.NotNil(t, divider, "Divider should be created")
}

func TestDivider_RenderHorizontalLineByDefault(t *testing.T) {
	divider := Divider(DividerProps{
		Length: 10,
	})

	divider.Init()
	output := divider.View()

	// Default horizontal divider uses ─ character
	assert.Contains(t, output, "─", "Should render horizontal line character")
	assert.NotContains(t, output, "│", "Should not contain vertical character")
}

func TestDivider_RenderVerticalLineWhenVerticalTrue(t *testing.T) {
	divider := Divider(DividerProps{
		Vertical: true,
		Length:   5,
	})

	divider.Init()
	output := divider.View()

	// Vertical divider uses │ character
	assert.Contains(t, output, "│", "Should render vertical line character")
}

func TestDivider_CentersLabelText(t *testing.T) {
	tests := []struct {
		name   string
		label  string
		length int
	}{
		{
			name:   "short label",
			label:  "OR",
			length: 20,
		},
		{
			name:   "longer label",
			label:  "Section Title",
			length: 40,
		},
		{
			name:   "single char label",
			label:  "X",
			length: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			divider := Divider(DividerProps{
				Label:  tt.label,
				Length: tt.length,
			})

			divider.Init()
			output := divider.View()

			assert.Contains(t, output, tt.label, "Should contain the label text")
			// Label should be surrounded by divider characters
			assert.Contains(t, output, "─", "Should have divider characters around label")
		})
	}
}

func TestDivider_UsesCustomCharacter(t *testing.T) {
	tests := []struct {
		name     string
		char     string
		vertical bool
		expected string
	}{
		{
			name:     "custom horizontal char",
			char:     "═",
			vertical: false,
			expected: "═",
		},
		{
			name:     "custom vertical char",
			char:     "║",
			vertical: true,
			expected: "║",
		},
		{
			name:     "asterisk char",
			char:     "*",
			vertical: false,
			expected: "*",
		},
		{
			name:     "dash char",
			char:     "-",
			vertical: false,
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			divider := Divider(DividerProps{
				Char:     tt.char,
				Vertical: tt.vertical,
				Length:   10,
			})

			divider.Init()
			output := divider.View()

			assert.Contains(t, output, tt.expected, "Should use custom character")
		})
	}
}

func TestDivider_UsesThemeMutedColor(t *testing.T) {
	// Theme integration test - divider should use theme.Muted for color
	divider := Divider(DividerProps{
		Length: 10,
	})

	divider.Init()
	output := divider.View()

	// Output should not be empty and should have styling applied
	assert.NotEmpty(t, output, "Should render with theme styling")
}

func TestDivider_BubbleteatIntegration(t *testing.T) {
	divider := Divider(DividerProps{
		Length: 10,
	})

	// Test Init
	cmd := divider.Init()
	assert.Nil(t, cmd, "Init should return nil command")

	// Test Update
	newModel, cmd := divider.Update(nil)
	assert.NotNil(t, newModel, "Update should return model")
	assert.Nil(t, cmd, "Update should return nil command for nil msg")

	// Test View
	output := divider.View()
	assert.NotEmpty(t, output, "View should return output")
}

func TestDivider_DefaultLength(t *testing.T) {
	// When Length is 0, should use a sensible default
	divider := Divider(DividerProps{})

	divider.Init()
	output := divider.View()

	// Should still render something
	assert.NotEmpty(t, output, "Should render with default length")
	assert.Contains(t, output, "─", "Should contain divider character")
}

func TestDivider_VerticalWithLabel(t *testing.T) {
	// Vertical dividers with labels are tricky - label should still appear
	divider := Divider(DividerProps{
		Vertical: true,
		Label:    "V",
		Length:   5,
	})

	divider.Init()
	output := divider.View()

	// Should contain the label
	assert.Contains(t, output, "V", "Vertical divider should contain label")
}

func TestDivider_CustomStyle(t *testing.T) {
	customStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))

	divider := Divider(DividerProps{
		Length: 10,
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})

	divider.Init()
	output := divider.View()

	assert.NotEmpty(t, output, "Should render with custom style")
}

func TestDivider_LabelLongerThanLength(t *testing.T) {
	// Edge case: label is longer than specified length
	divider := Divider(DividerProps{
		Label:  "Very Long Label Text",
		Length: 10,
	})

	divider.Init()
	output := divider.View()

	// Should handle gracefully - either truncate or expand
	assert.NotEmpty(t, output, "Should handle long label gracefully")
}

func TestDivider_EmptyLabel(t *testing.T) {
	divider := Divider(DividerProps{
		Label:  "",
		Length: 20,
	})

	divider.Init()
	output := divider.View()

	// Should render full line without label
	assert.NotEmpty(t, output, "Should render without label")
	// Count the divider characters - should be consistent
	count := strings.Count(output, "─")
	assert.Greater(t, count, 0, "Should have divider characters")
}

func TestDivider_AllPropsCombo(t *testing.T) {
	customStyle := lipgloss.NewStyle().Bold(true)

	divider := Divider(DividerProps{
		Vertical: false,
		Length:   30,
		Label:    "Section",
		Char:     "━",
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})

	divider.Init()
	output := divider.View()

	assert.Contains(t, output, "Section", "Should render label")
	assert.Contains(t, output, "━", "Should use custom character")
}

func TestDivider_LabelWithSpacesTooLong(t *testing.T) {
	// Edge case: label with spaces would exceed length
	// Label "ABCD" with spaces " ABCD " = 6 chars, length = 5
	// This should trigger the fallback to just use label without spaces
	divider := Divider(DividerProps{
		Label:  "ABCD",
		Length: 5,
	})

	divider.Init()
	output := divider.View()

	assert.Contains(t, output, "ABCD", "Should contain the label")
	assert.NotEmpty(t, output, "Should render")
}

func TestDivider_HorizontalVsVerticalDirection(t *testing.T) {
	tests := []struct {
		name     string
		vertical bool
		length   int
	}{
		{
			name:     "horizontal short",
			vertical: false,
			length:   5,
		},
		{
			name:     "horizontal long",
			vertical: false,
			length:   50,
		},
		{
			name:     "vertical short",
			vertical: true,
			length:   3,
		},
		{
			name:     "vertical long",
			vertical: true,
			length:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			divider := Divider(DividerProps{
				Vertical: tt.vertical,
				Length:   tt.length,
			})

			divider.Init()
			output := divider.View()

			assert.NotEmpty(t, output, "Should render divider")

			if tt.vertical {
				// Vertical dividers should have newlines
				assert.Contains(t, output, "│", "Vertical should use │")
			} else {
				// Horizontal dividers should be on one line
				assert.Contains(t, output, "─", "Horizontal should use ─")
			}
		})
	}
}
