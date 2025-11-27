package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// mockFlexComponent creates a simple component for testing.
func mockFlexComponent(content string) bubbly.Component {
	comp, _ := bubbly.NewComponent("MockFlex").
		Props(struct{ Content string }{Content: content}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(struct{ Content string })
			return p.Content
		}).
		Build()
	return comp
}

// mockFlexComponentWithSize creates a component with specific dimensions.
func mockFlexComponentWithSize(content string, width, height int) bubbly.Component {
	comp, _ := bubbly.NewComponent("MockFlexSized").
		Props(struct {
			Content string
			Width   int
			Height  int
		}{Content: content, Width: width, Height: height}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(struct {
				Content string
				Width   int
				Height  int
			})
			style := lipgloss.NewStyle()
			if p.Width > 0 {
				style = style.Width(p.Width)
			}
			if p.Height > 0 {
				style = style.Height(p.Height)
			}
			return style.Render(p.Content)
		}).
		Build()
	return comp
}

// TestFlex_EmptyItems tests that empty items array returns empty string.
func TestFlex_EmptyItems(t *testing.T) {
	flex := Flex(FlexProps{
		Items: []bubbly.Component{},
	})
	flex.Init()

	result := flex.View()
	assert.Equal(t, "", result)
}

// TestFlex_NilItems tests that nil items are skipped.
func TestFlex_NilItems(t *testing.T) {
	flex := Flex(FlexProps{
		Items: []bubbly.Component{nil, nil},
	})
	flex.Init()

	result := flex.View()
	assert.Equal(t, "", result)
}

// TestFlex_SingleItem tests rendering with a single item.
func TestFlex_SingleItem(t *testing.T) {
	item := mockFlexComponent("Hello")
	flex := Flex(FlexProps{
		Items: []bubbly.Component{item},
	})
	flex.Init()

	result := flex.View()
	assert.Contains(t, result, "Hello")
}

// TestFlex_RowDirection tests horizontal arrangement of items.
func TestFlex_RowDirection(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected []string // All items should appear in result
	}{
		{
			name:     "two items",
			items:    []string{"A", "B"},
			expected: []string{"A", "B"},
		},
		{
			name:     "three items",
			items:    []string{"One", "Two", "Three"},
			expected: []string{"One", "Two", "Three"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := make([]bubbly.Component, len(tt.items))
			for i, content := range tt.items {
				items[i] = mockFlexComponent(content)
			}

			flex := Flex(FlexProps{
				Items:     items,
				Direction: FlexRow,
			})
			flex.Init()

			result := flex.View()
			for _, exp := range tt.expected {
				assert.Contains(t, result, exp)
			}
		})
	}
}

// TestFlex_ColumnDirection tests vertical arrangement of items.
func TestFlex_ColumnDirection(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected []string
	}{
		{
			name:     "two items",
			items:    []string{"Top", "Bottom"},
			expected: []string{"Top", "Bottom"},
		},
		{
			name:     "three items",
			items:    []string{"Header", "Content", "Footer"},
			expected: []string{"Header", "Content", "Footer"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := make([]bubbly.Component, len(tt.items))
			for i, content := range tt.items {
				items[i] = mockFlexComponent(content)
			}

			flex := Flex(FlexProps{
				Items:     items,
				Direction: FlexColumn,
			})
			flex.Init()

			result := flex.View()
			for _, exp := range tt.expected {
				assert.Contains(t, result, exp)
			}

			// Verify vertical arrangement (items on separate lines)
			lines := strings.Split(result, "\n")
			assert.GreaterOrEqual(t, len(lines), len(tt.items)-1, "Column should have multiple lines")
		})
	}
}

// TestFlex_JustifyStart tests items aligned to start.
func TestFlex_JustifyStart(t *testing.T) {
	items := []bubbly.Component{
		mockFlexComponent("A"),
		mockFlexComponent("B"),
	}

	flex := Flex(FlexProps{
		Items:     items,
		Direction: FlexRow,
		Justify:   JustifyStart,
		Width:     20,
	})
	flex.Init()

	result := flex.View()
	assert.Contains(t, result, "A")
	assert.Contains(t, result, "B")

	// Items should be at the start (A comes before B in the string)
	aIdx := strings.Index(result, "A")
	bIdx := strings.Index(result, "B")
	assert.Less(t, aIdx, bIdx, "A should come before B")
}

// TestFlex_JustifyCenter tests items centered.
func TestFlex_JustifyCenter(t *testing.T) {
	items := []bubbly.Component{
		mockFlexComponent("X"),
	}

	flex := Flex(FlexProps{
		Items:     items,
		Direction: FlexRow,
		Justify:   JustifyCenter,
		Width:     20,
	})
	flex.Init()

	result := flex.View()
	assert.Contains(t, result, "X")

	// With centering, there should be leading spaces
	xIdx := strings.Index(result, "X")
	assert.Greater(t, xIdx, 0, "X should have leading space when centered")
}

// TestFlex_JustifyEnd tests items aligned to end.
func TestFlex_JustifyEnd(t *testing.T) {
	items := []bubbly.Component{
		mockFlexComponent("Z"),
	}

	flex := Flex(FlexProps{
		Items:     items,
		Direction: FlexRow,
		Justify:   JustifyEnd,
		Width:     20,
	})
	flex.Init()

	result := flex.View()
	assert.Contains(t, result, "Z")

	// With end alignment, there should be leading spaces
	zIdx := strings.Index(result, "Z")
	assert.Greater(t, zIdx, 0, "Z should have leading space when end-aligned")
}

// TestFlex_Gap tests spacing between items.
func TestFlex_Gap(t *testing.T) {
	tests := []struct {
		name      string
		gap       int
		direction FlexDirection
	}{
		{
			name:      "row with gap 2",
			gap:       2,
			direction: FlexRow,
		},
		{
			name:      "row with gap 5",
			gap:       5,
			direction: FlexRow,
		},
		{
			name:      "column with gap 1",
			gap:       1,
			direction: FlexColumn,
		},
		{
			name:      "column with gap 3",
			gap:       3,
			direction: FlexColumn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := []bubbly.Component{
				mockFlexComponent("A"),
				mockFlexComponent("B"),
			}

			flex := Flex(FlexProps{
				Items:     items,
				Direction: tt.direction,
				Gap:       tt.gap,
			})
			flex.Init()

			result := flex.View()
			assert.Contains(t, result, "A")
			assert.Contains(t, result, "B")

			if tt.direction == FlexRow {
				// For row, gap creates spaces between items
				aIdx := strings.Index(result, "A")
				bIdx := strings.Index(result, "B")
				actualGap := bIdx - aIdx - 1 // -1 for the "A" character
				assert.GreaterOrEqual(t, actualGap, tt.gap, "Gap should be at least %d", tt.gap)
			} else {
				// For column, gap creates empty lines
				lines := strings.Split(result, "\n")
				assert.GreaterOrEqual(t, len(lines), 2, "Column should have multiple lines")
			}
		})
	}
}

// TestFlex_AlignItemsRow tests cross-axis alignment in row direction.
func TestFlex_AlignItemsRow(t *testing.T) {
	tests := []struct {
		name  string
		align AlignItems
	}{
		{name: "start", align: AlignItemsStart},
		{name: "center", align: AlignItemsCenter},
		{name: "end", align: AlignItemsEnd},
		{name: "stretch", align: AlignItemsStretch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create items with different heights
			item1 := mockFlexComponentWithSize("A", 3, 1)
			item2 := mockFlexComponentWithSize("B\nB", 3, 2)

			flex := Flex(FlexProps{
				Items:     []bubbly.Component{item1, item2},
				Direction: FlexRow,
				Align:     tt.align,
			})
			flex.Init()

			result := flex.View()
			assert.Contains(t, result, "A")
			assert.Contains(t, result, "B")
		})
	}
}

// TestFlex_AlignItemsColumn tests cross-axis alignment in column direction.
func TestFlex_AlignItemsColumn(t *testing.T) {
	tests := []struct {
		name  string
		align AlignItems
	}{
		{name: "start", align: AlignItemsStart},
		{name: "center", align: AlignItemsCenter},
		{name: "end", align: AlignItemsEnd},
		{name: "stretch", align: AlignItemsStretch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create items with different widths
			item1 := mockFlexComponent("Short")
			item2 := mockFlexComponent("LongerText")

			flex := Flex(FlexProps{
				Items:     []bubbly.Component{item1, item2},
				Direction: FlexColumn,
				Align:     tt.align,
			})
			flex.Init()

			result := flex.View()
			assert.Contains(t, result, "Short")
			assert.Contains(t, result, "LongerText")
		})
	}
}

// TestFlex_DefaultValues tests that defaults are applied correctly.
func TestFlex_DefaultValues(t *testing.T) {
	props := FlexProps{}
	flexApplyDefaults(&props)

	assert.Equal(t, FlexRow, props.Direction, "Default direction should be row")
	assert.Equal(t, JustifyStart, props.Justify, "Default justify should be start")
	assert.Equal(t, AlignItemsStart, props.Align, "Default align should be start")
}

// TestFlex_CustomStyle tests that custom style is applied.
func TestFlex_CustomStyle(t *testing.T) {
	customStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("240"))

	items := []bubbly.Component{
		mockFlexComponent("Styled"),
	}

	flex := Flex(FlexProps{
		Items: items,
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})
	flex.Init()

	result := flex.View()
	assert.Contains(t, result, "Styled")
	// Style is applied (we can't easily verify ANSI codes, but no error means success)
}

// TestFlex_WidthConstraint tests fixed width container.
func TestFlex_WidthConstraint(t *testing.T) {
	items := []bubbly.Component{
		mockFlexComponent("A"),
		mockFlexComponent("B"),
	}

	flex := Flex(FlexProps{
		Items:     items,
		Direction: FlexRow,
		Justify:   JustifySpaceBetween,
		Width:     20,
	})
	flex.Init()

	result := flex.View()
	assert.Contains(t, result, "A")
	assert.Contains(t, result, "B")

	// With space-between in 20 chars, items should be spread
	width := lipgloss.Width(result)
	assert.GreaterOrEqual(t, width, 2, "Result should have width")
}

// TestFlex_HeightConstraint tests fixed height container.
func TestFlex_HeightConstraint(t *testing.T) {
	items := []bubbly.Component{
		mockFlexComponent("Top"),
		mockFlexComponent("Bottom"),
	}

	flex := Flex(FlexProps{
		Items:     items,
		Direction: FlexColumn,
		Justify:   JustifySpaceBetween,
		Height:    10,
	})
	flex.Init()

	result := flex.View()
	assert.Contains(t, result, "Top")
	assert.Contains(t, result, "Bottom")
}

// TestFlex_MixedNilItems tests that nil items among valid items are skipped.
func TestFlex_MixedNilItems(t *testing.T) {
	items := []bubbly.Component{
		mockFlexComponent("First"),
		nil,
		mockFlexComponent("Third"),
	}

	flex := Flex(FlexProps{
		Items: items,
	})
	flex.Init()

	result := flex.View()
	assert.Contains(t, result, "First")
	assert.Contains(t, result, "Third")
}

// TestFlexApplyDefaults tests the default application function.
func TestFlexApplyDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    FlexProps
		expected FlexProps
	}{
		{
			name:  "empty props get defaults",
			input: FlexProps{},
			expected: FlexProps{
				Direction: FlexRow,
				Justify:   JustifyStart,
				Align:     AlignItemsStart,
			},
		},
		{
			name: "explicit values preserved",
			input: FlexProps{
				Direction: FlexColumn,
				Justify:   JustifyCenter,
				Align:     AlignItemsEnd,
			},
			expected: FlexProps{
				Direction: FlexColumn,
				Justify:   JustifyCenter,
				Align:     AlignItemsEnd,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props := tt.input
			flexApplyDefaults(&props)

			assert.Equal(t, tt.expected.Direction, props.Direction)
			assert.Equal(t, tt.expected.Justify, props.Justify)
			assert.Equal(t, tt.expected.Align, props.Align)
		})
	}
}

// TestFlexRenderItems tests the item rendering helper.
func TestFlexRenderItems(t *testing.T) {
	tests := []struct {
		name     string
		items    []bubbly.Component
		expected int // number of rendered items
	}{
		{
			name:     "empty",
			items:    []bubbly.Component{},
			expected: 0,
		},
		{
			name:     "all nil",
			items:    []bubbly.Component{nil, nil},
			expected: 0,
		},
		{
			name: "mixed",
			items: []bubbly.Component{
				mockFlexComponent("A"),
				nil,
				mockFlexComponent("B"),
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize components
			for _, item := range tt.items {
				if item != nil {
					item.Init()
				}
			}

			rendered := flexRenderItems(tt.items)
			assert.Equal(t, tt.expected, len(rendered))
		})
	}
}

// TestFlexCalculateMaxHeight tests height calculation.
func TestFlexCalculateMaxHeight(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected int
	}{
		{
			name:     "empty",
			items:    []string{},
			expected: 0,
		},
		{
			name:     "single line items",
			items:    []string{"A", "BB", "CCC"},
			expected: 1,
		},
		{
			name:     "multi-line items",
			items:    []string{"A", "B\nB", "C\nC\nC"},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flexCalculateMaxHeight(tt.items)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFlexCalculateMaxWidth tests width calculation.
func TestFlexCalculateMaxWidth(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected int
	}{
		{
			name:     "empty",
			items:    []string{},
			expected: 0,
		},
		{
			name:     "varying widths",
			items:    []string{"A", "BB", "CCC"},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flexCalculateMaxWidth(tt.items)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFlexCalculateGaps tests gap calculation for different justify modes.
func TestFlexCalculateGaps(t *testing.T) {
	tests := []struct {
		name          string
		itemSizes     []int
		containerSize int
		justify       JustifyContent
		gap           int
		expectGaps    int // number of gaps (n-1 for n items)
	}{
		{
			name:          "start with gap",
			itemSizes:     []int{5, 5},
			containerSize: 20,
			justify:       JustifyStart,
			gap:           2,
			expectGaps:    1,
		},
		{
			name:          "space-between",
			itemSizes:     []int{5, 5},
			containerSize: 20,
			justify:       JustifySpaceBetween,
			gap:           0,
			expectGaps:    1,
		},
		{
			name:          "single item",
			itemSizes:     []int{5},
			containerSize: 20,
			justify:       JustifySpaceBetween,
			gap:           0,
			expectGaps:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gaps, _ := flexCalculateGaps(tt.itemSizes, tt.containerSize, tt.justify, tt.gap)
			assert.Equal(t, tt.expectGaps, len(gaps))
		})
	}
}

// TestFlex_ComponentInterface tests that Flex returns a valid bubbly.Component.
func TestFlex_ComponentInterface(t *testing.T) {
	flex := Flex(FlexProps{
		Items: []bubbly.Component{mockFlexComponent("Test")},
	})

	// Should implement Component interface
	require.NotNil(t, flex)

	// Init should not panic
	cmd := flex.Init()
	assert.Nil(t, cmd)

	// View should return string
	view := flex.View()
	assert.IsType(t, "", view)
}

// TestFlex_ThemeIntegration tests that theme is properly injected.
func TestFlex_ThemeIntegration(t *testing.T) {
	items := []bubbly.Component{
		mockFlexComponent("Themed"),
	}

	flex := Flex(FlexProps{
		Items: items,
	})
	flex.Init()

	// Should render without error (theme injection happens in Setup)
	result := flex.View()
	assert.Contains(t, result, "Themed")
}

// TestFlex_JustifySpaceBetween tests space-between distribution.
func TestFlex_JustifySpaceBetween(t *testing.T) {
	items := []bubbly.Component{
		mockFlexComponent("L"),
		mockFlexComponent("R"),
	}

	flex := Flex(FlexProps{
		Items:     items,
		Direction: FlexRow,
		Justify:   JustifySpaceBetween,
		Width:     20,
	})
	flex.Init()

	result := flex.View()
	assert.Contains(t, result, "L")
	assert.Contains(t, result, "R")

	// L should be at start, R should be at end
	lIdx := strings.Index(result, "L")
	rIdx := strings.Index(result, "R")
	assert.Less(t, lIdx, rIdx)
	// There should be significant space between them
	assert.Greater(t, rIdx-lIdx, 5, "Space-between should create gap")
}

// TestFlex_JustifySpaceAround tests space-around distribution.
func TestFlex_JustifySpaceAround(t *testing.T) {
	items := []bubbly.Component{
		mockFlexComponent("A"),
		mockFlexComponent("B"),
	}

	flex := Flex(FlexProps{
		Items:     items,
		Direction: FlexRow,
		Justify:   JustifySpaceAround,
		Width:     20,
	})
	flex.Init()

	result := flex.View()
	assert.Contains(t, result, "A")
	assert.Contains(t, result, "B")
}

// TestFlex_JustifySpaceEvenly tests space-evenly distribution.
func TestFlex_JustifySpaceEvenly(t *testing.T) {
	items := []bubbly.Component{
		mockFlexComponent("X"),
		mockFlexComponent("Y"),
	}

	flex := Flex(FlexProps{
		Items:     items,
		Direction: FlexRow,
		Justify:   JustifySpaceEvenly,
		Width:     20,
	})
	flex.Init()

	result := flex.View()
	assert.Contains(t, result, "X")
	assert.Contains(t, result, "Y")

	// Both items should have leading space (evenly distributed)
	xIdx := strings.Index(result, "X")
	assert.Greater(t, xIdx, 0, "X should have leading space")
}

// TestFlex_ZeroGap tests that zero gap works correctly.
func TestFlex_ZeroGap(t *testing.T) {
	items := []bubbly.Component{
		mockFlexComponent("A"),
		mockFlexComponent("B"),
	}

	flex := Flex(FlexProps{
		Items:     items,
		Direction: FlexRow,
		Gap:       0,
	})
	flex.Init()

	result := flex.View()
	// Items should be adjacent (no gap)
	assert.Contains(t, result, "AB")
}

// TestFlex_LargeGap tests large gap values.
func TestFlex_LargeGap(t *testing.T) {
	items := []bubbly.Component{
		mockFlexComponent("A"),
		mockFlexComponent("B"),
	}

	flex := Flex(FlexProps{
		Items:     items,
		Direction: FlexRow,
		Gap:       10,
	})
	flex.Init()

	result := flex.View()
	assert.Contains(t, result, "A")
	assert.Contains(t, result, "B")

	// Gap should create significant space
	aIdx := strings.Index(result, "A")
	bIdx := strings.Index(result, "B")
	assert.GreaterOrEqual(t, bIdx-aIdx, 10, "Large gap should create space")
}
