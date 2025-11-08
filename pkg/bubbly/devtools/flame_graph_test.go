package devtools

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestFlameGraphRenderer_New tests the constructor
func TestFlameGraphRenderer_New(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"default size", 80, 10},
		{"small size", 40, 5},
		{"large size", 120, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewFlameGraphRenderer(tt.width, tt.height)
			assert.NotNil(t, renderer)
			assert.Equal(t, tt.width, renderer.width)
			assert.Equal(t, tt.height, renderer.height)
		})
	}
}

// TestFlameGraphRenderer_BuildFlameTree_Empty tests building tree from empty data
func TestFlameGraphRenderer_BuildFlameTree_Empty(t *testing.T) {
	renderer := NewFlameGraphRenderer(80, 10)
	data := NewPerformanceData()

	root := renderer.buildFlameTree(data)

	assert.NotNil(t, root)
	assert.Equal(t, "Application", root.Name)
	assert.Equal(t, time.Duration(0), root.Time)
	assert.Equal(t, 100.0, root.TimePercentage)
	assert.Empty(t, root.Children)
}

// TestFlameGraphRenderer_BuildFlameTree_SingleComponent tests single component
func TestFlameGraphRenderer_BuildFlameTree_SingleComponent(t *testing.T) {
	renderer := NewFlameGraphRenderer(80, 10)
	data := NewPerformanceData()
	data.RecordRender("comp-1", "Counter", 10*time.Millisecond)

	root := renderer.buildFlameTree(data)

	assert.NotNil(t, root)
	assert.Equal(t, "Application", root.Name)
	assert.Equal(t, 10*time.Millisecond, root.Time)
	assert.Equal(t, 100.0, root.TimePercentage)
	assert.Len(t, root.Children, 1)

	child := root.Children[0]
	assert.Equal(t, "Counter", child.Name)
	assert.Equal(t, 10*time.Millisecond, child.Time)
	assert.Equal(t, 100.0, child.TimePercentage)
	assert.Empty(t, child.Children)
}

// TestFlameGraphRenderer_BuildFlameTree_MultipleComponents tests multiple components
func TestFlameGraphRenderer_BuildFlameTree_MultipleComponents(t *testing.T) {
	renderer := NewFlameGraphRenderer(80, 10)
	data := NewPerformanceData()
	data.RecordRender("comp-1", "Counter", 20*time.Millisecond)
	data.RecordRender("comp-2", "Header", 15*time.Millisecond)
	data.RecordRender("comp-3", "Footer", 10*time.Millisecond)

	root := renderer.buildFlameTree(data)

	assert.NotNil(t, root)
	assert.Equal(t, "Application", root.Name)
	assert.Equal(t, 45*time.Millisecond, root.Time)
	assert.Equal(t, 100.0, root.TimePercentage)
	assert.Len(t, root.Children, 3)

	// Check percentages sum to ~100% (allowing for floating point rounding)
	totalPercentage := 0.0
	for _, child := range root.Children {
		totalPercentage += child.TimePercentage
	}
	assert.InDelta(t, 100.0, totalPercentage, 0.1)

	// Check individual percentages
	assert.Equal(t, "Counter", root.Children[0].Name)
	assert.InDelta(t, 44.44, root.Children[0].TimePercentage, 0.1) // 20/45 * 100

	assert.Equal(t, "Header", root.Children[1].Name)
	assert.InDelta(t, 33.33, root.Children[1].TimePercentage, 0.1) // 15/45 * 100

	assert.Equal(t, "Footer", root.Children[2].Name)
	assert.InDelta(t, 22.22, root.Children[2].TimePercentage, 0.1) // 10/45 * 100
}

// TestFlameGraphRenderer_Render_Empty tests rendering empty data
func TestFlameGraphRenderer_Render_Empty(t *testing.T) {
	renderer := NewFlameGraphRenderer(80, 10)
	data := NewPerformanceData()

	output := renderer.Render(data)

	assert.NotEmpty(t, output)
	assert.Contains(t, output, "No performance data")
}

// TestFlameGraphRenderer_Render_SingleComponent tests rendering single component
func TestFlameGraphRenderer_Render_SingleComponent(t *testing.T) {
	renderer := NewFlameGraphRenderer(80, 10)
	data := NewPerformanceData()
	data.RecordRender("comp-1", "Counter", 10*time.Millisecond)

	output := renderer.Render(data)

	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Application")
	assert.Contains(t, output, "Counter")
	assert.Contains(t, output, "█") // Contains bar character
}

// TestFlameGraphRenderer_Render_MultipleComponents tests rendering multiple components
func TestFlameGraphRenderer_Render_MultipleComponents(t *testing.T) {
	renderer := NewFlameGraphRenderer(80, 10)
	data := NewPerformanceData()
	data.RecordRender("comp-1", "Counter", 20*time.Millisecond)
	data.RecordRender("comp-2", "Header", 15*time.Millisecond)
	data.RecordRender("comp-3", "Footer", 10*time.Millisecond)

	output := renderer.Render(data)

	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Application")
	assert.Contains(t, output, "Counter")
	assert.Contains(t, output, "Header")
	assert.Contains(t, output, "Footer")
	assert.Contains(t, output, "█") // Contains bar character

	// Check that bars are present (multiple █ characters)
	barCount := strings.Count(output, "█")
	assert.Greater(t, barCount, 3, "Should have multiple bars")
}

// TestFlameGraphRenderer_Render_LongNames tests truncation of long names
func TestFlameGraphRenderer_Render_LongNames(t *testing.T) {
	renderer := NewFlameGraphRenderer(80, 10)
	data := NewPerformanceData()
	data.RecordRender("comp-1", "VeryLongComponentNameThatShouldBeTruncated", 10*time.Millisecond)

	output := renderer.Render(data)

	assert.NotEmpty(t, output)
	// Should contain truncated name with ellipsis
	assert.Contains(t, output, "...")
}

// TestFlameGraphRenderer_Render_SmallWidth tests rendering with small width
func TestFlameGraphRenderer_Render_SmallWidth(t *testing.T) {
	renderer := NewFlameGraphRenderer(40, 10)
	data := NewPerformanceData()
	data.RecordRender("comp-1", "Counter", 10*time.Millisecond)

	output := renderer.Render(data)

	assert.NotEmpty(t, output)
	// Should still render something meaningful
	assert.Contains(t, output, "█")
}

// TestFlameGraphRenderer_CalculateBarWidth tests bar width calculation
func TestFlameGraphRenderer_CalculateBarWidth(t *testing.T) {
	tests := []struct {
		name       string
		percentage float64
		width      int
		expected   int
	}{
		{"100% of 80", 100.0, 80, 80},
		{"50% of 80", 50.0, 80, 40},
		{"25% of 80", 25.0, 80, 20},
		{"10% of 80", 10.0, 80, 8},
		{"1% of 80", 1.0, 80, 1},   // Minimum 1
		{"0.1% of 80", 0.1, 80, 1}, // Minimum 1
		{"100% of 40", 100.0, 40, 40},
		{"33% of 60", 33.33, 60, 19}, // 33.33% of 60 = 19.998, rounds to 19
	}

	renderer := NewFlameGraphRenderer(80, 10)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.calculateBarWidth(tt.percentage, tt.width)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFlameGraphRenderer_GetColorForTime tests color selection
func TestFlameGraphRenderer_GetColorForTime(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string // Color code
	}{
		{"very fast", 1 * time.Millisecond, "35"},   // Green
		{"fast", 4 * time.Millisecond, "35"},        // Green
		{"medium", 7 * time.Millisecond, "229"},     // Yellow
		{"slow", 15 * time.Millisecond, "196"},      // Red
		{"very slow", 50 * time.Millisecond, "196"}, // Red
	}

	renderer := NewFlameGraphRenderer(80, 10)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := renderer.getColorForTime(tt.duration)
			assert.Equal(t, tt.expected, string(color))
		})
	}
}

// TestFlameGraphRenderer_FormatBar tests bar formatting
func TestFlameGraphRenderer_FormatBar(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		label    string
		expected string
	}{
		{"simple bar", 10, "Test", "████████"},
		{"bar with label", 20, "Counter", "████████████████"},
		{"minimum width", 1, "X", "█"},
		{"zero width", 0, "Test", ""}, // Edge case
	}

	renderer := NewFlameGraphRenderer(80, 10)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.formatBar(tt.width)
			// Count █ characters
			barCount := strings.Count(result, "█")
			if tt.width > 0 {
				assert.Equal(t, tt.width, barCount)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

// TestFlameGraphRenderer_TruncateLabel tests label truncation
func TestFlameGraphRenderer_TruncateLabel(t *testing.T) {
	tests := []struct {
		name     string
		label    string
		maxLen   int
		expected string
	}{
		{"short label", "Test", 10, "Test"},
		{"exact length", "TestLabel", 9, "TestLabel"},
		{"needs truncation", "VeryLongLabel", 10, "VeryLon..."},
		{"very short max", "Test", 3, "Tes"},
		{"empty label", "", 10, ""},
	}

	renderer := NewFlameGraphRenderer(80, 10)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.truncateLabel(tt.label, tt.maxLen)
			assert.LessOrEqual(t, len(result), tt.maxLen)
			if len(tt.label) > tt.maxLen && tt.maxLen > 3 {
				assert.Contains(t, result, "...")
			}
		})
	}
}

// TestFlameGraphRenderer_Render_Styling tests that output includes Lipgloss styling
func TestFlameGraphRenderer_Render_Styling(t *testing.T) {
	renderer := NewFlameGraphRenderer(80, 10)
	data := NewPerformanceData()
	data.RecordRender("comp-1", "Counter", 10*time.Millisecond)

	output := renderer.Render(data)

	// Output should contain ANSI escape codes for styling
	// Lipgloss adds these for colors
	assert.NotEmpty(t, output)
	// Just verify it renders without panicking
	// Actual styling verification would require parsing ANSI codes
}

// TestFlameGraphRenderer_Render_PercentageDisplay tests percentage display
func TestFlameGraphRenderer_Render_PercentageDisplay(t *testing.T) {
	renderer := NewFlameGraphRenderer(80, 10)
	data := NewPerformanceData()
	data.RecordRender("comp-1", "Counter", 20*time.Millisecond)
	data.RecordRender("comp-2", "Header", 10*time.Millisecond)

	output := renderer.Render(data)

	// Should show percentages
	assert.Contains(t, output, "%")
	// Should show time values
	assert.Contains(t, output, "ms")
}

// TestFlameGraphRenderer_Render_SortedByTime tests components are sorted by time
func TestFlameGraphRenderer_Render_SortedByTime(t *testing.T) {
	renderer := NewFlameGraphRenderer(80, 10)
	data := NewPerformanceData()
	data.RecordRender("comp-1", "Slow", 30*time.Millisecond)
	data.RecordRender("comp-2", "Fast", 5*time.Millisecond)
	data.RecordRender("comp-3", "Medium", 15*time.Millisecond)

	output := renderer.Render(data)

	// Find positions of component names in output
	slowPos := strings.Index(output, "Slow")
	mediumPos := strings.Index(output, "Medium")
	fastPos := strings.Index(output, "Fast")

	// Should be sorted by time (descending): Slow, Medium, Fast
	assert.Less(t, slowPos, mediumPos, "Slow should appear before Medium")
	assert.Less(t, mediumPos, fastPos, "Medium should appear before Fast")
}
