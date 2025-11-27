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

// =============================================================================
// Task 4.2: Space Distribution Tests
// =============================================================================

// TestFlex_SpaceBetween_DistributesEvenly tests that space-between distributes
// remaining space evenly between items with no space on edges.
func TestFlex_SpaceBetween_DistributesEvenly(t *testing.T) {
	tests := []struct {
		name          string
		items         []string
		containerSize int
		direction     FlexDirection
		expectGapMin  int // minimum expected gap between items
	}{
		{
			name:          "two items in row",
			items:         []string{"A", "B"},
			containerSize: 20,
			direction:     FlexRow,
			expectGapMin:  18, // 20 - 2 chars = 18 space between
		},
		{
			name:          "three items in row",
			items:         []string{"A", "B", "C"},
			containerSize: 30,
			direction:     FlexRow,
			expectGapMin:  13, // (30 - 3) / 2 = 13.5, floor = 13
		},
		{
			name:          "two items in column",
			items:         []string{"Top", "Bottom"},
			containerSize: 10,
			direction:     FlexColumn,
			expectGapMin:  8, // 10 - 2 lines = 8 lines between
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := make([]bubbly.Component, len(tt.items))
			for i, content := range tt.items {
				items[i] = mockFlexComponent(content)
			}

			props := FlexProps{
				Items:     items,
				Direction: tt.direction,
				Justify:   JustifySpaceBetween,
			}
			if tt.direction == FlexRow {
				props.Width = tt.containerSize
			} else {
				props.Height = tt.containerSize
			}

			flex := Flex(props)
			flex.Init()

			result := flex.View()

			// Verify all items present
			for _, item := range tt.items {
				assert.Contains(t, result, item)
			}

			// For row direction, verify first item starts at position 0 (no leading space)
			if tt.direction == FlexRow {
				firstIdx := strings.Index(result, tt.items[0])
				assert.Equal(t, 0, firstIdx, "First item should start at position 0 (no edge space)")
			}
		})
	}
}

// TestFlex_SpaceAround_AddsHalfEdgeSpace tests that space-around adds
// half-size space on edges and full space between items.
func TestFlex_SpaceAround_AddsHalfEdgeSpace(t *testing.T) {
	tests := []struct {
		name          string
		items         []string
		containerSize int
		direction     FlexDirection
	}{
		{
			name:          "two items in row",
			items:         []string{"A", "B"},
			containerSize: 20,
			direction:     FlexRow,
		},
		{
			name:          "three items in row",
			items:         []string{"X", "Y", "Z"},
			containerSize: 30,
			direction:     FlexRow,
		},
		{
			name:          "two items in column",
			items:         []string{"Top", "Bottom"},
			containerSize: 10,
			direction:     FlexColumn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := make([]bubbly.Component, len(tt.items))
			for i, content := range tt.items {
				items[i] = mockFlexComponent(content)
			}

			props := FlexProps{
				Items:     items,
				Direction: tt.direction,
				Justify:   JustifySpaceAround,
			}
			if tt.direction == FlexRow {
				props.Width = tt.containerSize
			} else {
				props.Height = tt.containerSize
			}

			flex := Flex(props)
			flex.Init()

			result := flex.View()

			// Verify all items present
			for _, item := range tt.items {
				assert.Contains(t, result, item)
			}

			// For row direction, verify there IS leading space (edge space)
			if tt.direction == FlexRow {
				firstIdx := strings.Index(result, tt.items[0])
				assert.Greater(t, firstIdx, 0, "First item should have leading space (edge space)")
			}
		})
	}
}

// TestFlex_SpaceEvenly_DistributesAllSpaceEqually tests that space-evenly
// distributes space equally everywhere including edges.
func TestFlex_SpaceEvenly_DistributesAllSpaceEqually(t *testing.T) {
	tests := []struct {
		name          string
		items         []string
		containerSize int
		direction     FlexDirection
	}{
		{
			name:          "two items in row",
			items:         []string{"A", "B"},
			containerSize: 20,
			direction:     FlexRow,
		},
		{
			name:          "three items in row",
			items:         []string{"X", "Y", "Z"},
			containerSize: 30,
			direction:     FlexRow,
		},
		{
			name:          "two items in column",
			items:         []string{"Top", "Bottom"},
			containerSize: 12,
			direction:     FlexColumn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := make([]bubbly.Component, len(tt.items))
			for i, content := range tt.items {
				items[i] = mockFlexComponent(content)
			}

			props := FlexProps{
				Items:     items,
				Direction: tt.direction,
				Justify:   JustifySpaceEvenly,
			}
			if tt.direction == FlexRow {
				props.Width = tt.containerSize
			} else {
				props.Height = tt.containerSize
			}

			flex := Flex(props)
			flex.Init()

			result := flex.View()

			// Verify all items present
			for _, item := range tt.items {
				assert.Contains(t, result, item)
			}

			// For row direction, verify there IS leading space (edge space)
			if tt.direction == FlexRow {
				firstIdx := strings.Index(result, tt.items[0])
				assert.Greater(t, firstIdx, 0, "First item should have leading space (edge space)")
			}
		})
	}
}

// TestFlex_SpaceDistribution_SingleItem tests that space distribution
// handles single item gracefully.
func TestFlex_SpaceDistribution_SingleItem(t *testing.T) {
	tests := []struct {
		name      string
		justify   JustifyContent
		direction FlexDirection
	}{
		{name: "space-between row", justify: JustifySpaceBetween, direction: FlexRow},
		{name: "space-around row", justify: JustifySpaceAround, direction: FlexRow},
		{name: "space-evenly row", justify: JustifySpaceEvenly, direction: FlexRow},
		{name: "space-between column", justify: JustifySpaceBetween, direction: FlexColumn},
		{name: "space-around column", justify: JustifySpaceAround, direction: FlexColumn},
		{name: "space-evenly column", justify: JustifySpaceEvenly, direction: FlexColumn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := mockFlexComponent("Single")

			props := FlexProps{
				Items:     []bubbly.Component{item},
				Direction: tt.direction,
				Justify:   tt.justify,
			}
			if tt.direction == FlexRow {
				props.Width = 30
			} else {
				props.Height = 10
			}

			flex := Flex(props)
			flex.Init()

			result := flex.View()

			// Should render without panic
			assert.Contains(t, result, "Single")

			// For space-between with single item, item should be at start
			if tt.justify == JustifySpaceBetween && tt.direction == FlexRow {
				idx := strings.Index(result, "Single")
				assert.Equal(t, 0, idx, "Single item with space-between should be at start")
			}
		})
	}
}

// TestFlex_SpaceDistribution_EmptyItems tests that space distribution
// handles empty items array gracefully.
func TestFlex_SpaceDistribution_EmptyItems(t *testing.T) {
	tests := []struct {
		name    string
		justify JustifyContent
	}{
		{name: "space-between", justify: JustifySpaceBetween},
		{name: "space-around", justify: JustifySpaceAround},
		{name: "space-evenly", justify: JustifySpaceEvenly},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flex := Flex(FlexProps{
				Items:   []bubbly.Component{},
				Justify: tt.justify,
				Width:   30,
			})
			flex.Init()

			result := flex.View()

			// Should return empty string without panic
			assert.Equal(t, "", result)
		})
	}
}

// TestFlexDistributeSpaceBetween_Algorithm tests the space-between algorithm directly.
func TestFlexDistributeSpaceBetween_Algorithm(t *testing.T) {
	tests := []struct {
		name           string
		n              int
		remainingSpace int
		initialGaps    []int
		expectGaps     []int
		expectStart    int
		expectEnd      int
	}{
		{
			name:           "two items, 10 space",
			n:              2,
			remainingSpace: 10,
			initialGaps:    []int{0},
			expectGaps:     []int{10},
			expectStart:    0,
			expectEnd:      0,
		},
		{
			name:           "three items, 20 space",
			n:              3,
			remainingSpace: 20,
			initialGaps:    []int{0, 0},
			expectGaps:     []int{10, 10},
			expectStart:    0,
			expectEnd:      0,
		},
		{
			name:           "three items, 21 space (remainder)",
			n:              3,
			remainingSpace: 21,
			initialGaps:    []int{0, 0},
			expectGaps:     []int{11, 10}, // remainder distributed to first gap
			expectStart:    0,
			expectEnd:      0,
		},
		{
			name:           "single item",
			n:              1,
			remainingSpace: 10,
			initialGaps:    []int{},
			expectGaps:     []int{},
			expectStart:    0,
			expectEnd:      10, // remaining space goes to end
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gaps := make([]int, len(tt.initialGaps))
			copy(gaps, tt.initialGaps)

			result := flexDistributeSpaceBetween(tt.n, tt.remainingSpace, gaps)

			assert.Equal(t, tt.expectGaps, result.gaps, "gaps mismatch")
			assert.Equal(t, tt.expectStart, result.startPadding, "startPadding mismatch")
			assert.Equal(t, tt.expectEnd, result.endPadding, "endPadding mismatch")
		})
	}
}

// TestFlexDistributeSpaceAround_Algorithm tests the space-around algorithm directly.
func TestFlexDistributeSpaceAround_Algorithm(t *testing.T) {
	tests := []struct {
		name           string
		n              int
		remainingSpace int
		initialGaps    []int
		expectStart    int
		expectEnd      int
	}{
		{
			name:           "two items, 20 space",
			n:              2,
			remainingSpace: 20,
			initialGaps:    []int{0},
			expectStart:    5, // 20 / (2*2) = 5
			expectEnd:      5,
		},
		{
			name:           "three items, 30 space",
			n:              3,
			remainingSpace: 30,
			initialGaps:    []int{0, 0},
			expectStart:    5, // 30 / (3*2) = 5
			expectEnd:      5,
		},
		{
			name:           "single item, 10 space",
			n:              1,
			remainingSpace: 10,
			initialGaps:    []int{},
			expectStart:    5, // 10 / (1*2) = 5
			expectEnd:      5,
		},
		{
			name:           "zero items",
			n:              0,
			remainingSpace: 10,
			initialGaps:    []int{},
			expectStart:    0,
			expectEnd:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gaps := make([]int, len(tt.initialGaps))
			copy(gaps, tt.initialGaps)

			result := flexDistributeSpaceAround(tt.n, tt.remainingSpace, gaps)

			assert.Equal(t, tt.expectStart, result.startPadding, "startPadding mismatch")
			assert.Equal(t, tt.expectEnd, result.endPadding, "endPadding mismatch")
		})
	}
}

// TestFlexDistributeSpaceEvenly_Algorithm tests the space-evenly algorithm directly.
func TestFlexDistributeSpaceEvenly_Algorithm(t *testing.T) {
	tests := []struct {
		name           string
		n              int
		remainingSpace int
		initialGaps    []int
		expectStart    int
		expectEnd      int
	}{
		{
			name:           "two items, 15 space",
			n:              2,
			remainingSpace: 15,
			initialGaps:    []int{0},
			expectStart:    5, // 15 / 3 = 5
			expectEnd:      5,
		},
		{
			name:           "three items, 20 space",
			n:              3,
			remainingSpace: 20,
			initialGaps:    []int{0, 0},
			expectStart:    5, // 20 / 4 = 5
			expectEnd:      5,
		},
		{
			name:           "two items, 16 space (remainder)",
			n:              2,
			remainingSpace: 16,
			initialGaps:    []int{0},
			expectStart:    6, // 16 / 3 = 5 + 1 remainder
			expectEnd:      5,
		},
		{
			name:           "single item, 10 space",
			n:              1,
			remainingSpace: 10,
			initialGaps:    []int{},
			expectStart:    5, // 10 / 2 = 5
			expectEnd:      5,
		},
		{
			name:           "zero items",
			n:              0,
			remainingSpace: 10,
			initialGaps:    []int{},
			expectStart:    0,
			expectEnd:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gaps := make([]int, len(tt.initialGaps))
			copy(gaps, tt.initialGaps)

			result := flexDistributeSpaceEvenly(tt.n, tt.remainingSpace, gaps)

			assert.Equal(t, tt.expectStart, result.startPadding, "startPadding mismatch")
			assert.Equal(t, tt.expectEnd, result.endPadding, "endPadding mismatch")
		})
	}
}

// TestFlex_SpaceDistribution_WithGap tests space distribution combined with explicit gap.
func TestFlex_SpaceDistribution_WithGap(t *testing.T) {
	tests := []struct {
		name    string
		justify JustifyContent
		gap     int
	}{
		{name: "space-between with gap 2", justify: JustifySpaceBetween, gap: 2},
		{name: "space-around with gap 2", justify: JustifySpaceAround, gap: 2},
		{name: "space-evenly with gap 2", justify: JustifySpaceEvenly, gap: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := []bubbly.Component{
				mockFlexComponent("A"),
				mockFlexComponent("B"),
				mockFlexComponent("C"),
			}

			flex := Flex(FlexProps{
				Items:   items,
				Justify: tt.justify,
				Gap:     tt.gap,
				Width:   40,
			})
			flex.Init()

			result := flex.View()

			// All items should be present
			assert.Contains(t, result, "A")
			assert.Contains(t, result, "B")
			assert.Contains(t, result, "C")

			// Verify minimum gap is respected
			aIdx := strings.Index(result, "A")
			bIdx := strings.Index(result, "B")
			assert.GreaterOrEqual(t, bIdx-aIdx-1, tt.gap, "Gap should be at least %d", tt.gap)
		})
	}
}

// =============================================================================
// Task 4.3: Cross-Axis Alignment Tests
// =============================================================================

// TestFlex_AlignStart_PositionsAtTopLeft tests that AlignStart positions items
// at top (row) or left (column).
func TestFlex_AlignStart_PositionsAtTopLeft(t *testing.T) {
	tests := []struct {
		name      string
		direction FlexDirection
	}{
		{name: "row direction (top)", direction: FlexRow},
		{name: "column direction (left)", direction: FlexColumn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create items with different sizes
			small := mockFlexComponent("S")
			large := mockFlexComponentWithSize("L\nL\nL", 5, 3)

			flex := Flex(FlexProps{
				Items:     []bubbly.Component{small, large},
				Direction: tt.direction,
				Align:     AlignItemsStart,
			})
			flex.Init()

			result := flex.View()

			// Both items should be present
			assert.Contains(t, result, "S")
			assert.Contains(t, result, "L")

			if tt.direction == FlexRow {
				// In row with AlignStart, small item should be at top
				// First line should contain "S"
				lines := strings.Split(result, "\n")
				assert.True(t, strings.Contains(lines[0], "S"), "S should be on first line (top aligned)")
			} else {
				// In column with AlignStart, items should be left-aligned
				lines := strings.Split(result, "\n")
				for _, line := range lines {
					if strings.Contains(line, "S") {
						// S should start near the beginning
						idx := strings.Index(line, "S")
						assert.LessOrEqual(t, idx, 1, "S should be left-aligned")
					}
				}
			}
		})
	}
}

// TestFlex_AlignCenter_PositionsInMiddle tests that AlignCenter positions items
// in the middle of the cross-axis.
func TestFlex_AlignCenter_PositionsInMiddle(t *testing.T) {
	tests := []struct {
		name      string
		direction FlexDirection
	}{
		{name: "row direction (vertical center)", direction: FlexRow},
		{name: "column direction (horizontal center)", direction: FlexColumn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create items with different sizes
			small := mockFlexComponent("S")
			large := mockFlexComponentWithSize("L\nL\nL", 5, 3)

			flex := Flex(FlexProps{
				Items:     []bubbly.Component{small, large},
				Direction: tt.direction,
				Align:     AlignItemsCenter,
			})
			flex.Init()

			result := flex.View()

			// Both items should be present
			assert.Contains(t, result, "S")
			assert.Contains(t, result, "L")

			if tt.direction == FlexRow {
				// In row with AlignCenter, small item should be vertically centered
				lines := strings.Split(result, "\n")
				// With 3-line tall container, S should be on middle line (index 1)
				foundOnMiddle := false
				if len(lines) >= 2 {
					foundOnMiddle = strings.Contains(lines[1], "S")
				}
				assert.True(t, foundOnMiddle, "S should be vertically centered")
			}
		})
	}
}

// TestFlex_AlignEnd_PositionsAtBottomRight tests that AlignEnd positions items
// at bottom (row) or right (column).
func TestFlex_AlignEnd_PositionsAtBottomRight(t *testing.T) {
	tests := []struct {
		name      string
		direction FlexDirection
	}{
		{name: "row direction (bottom)", direction: FlexRow},
		{name: "column direction (right)", direction: FlexColumn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create items with different sizes
			small := mockFlexComponent("S")
			large := mockFlexComponentWithSize("L\nL\nL", 5, 3)

			flex := Flex(FlexProps{
				Items:     []bubbly.Component{small, large},
				Direction: tt.direction,
				Align:     AlignItemsEnd,
			})
			flex.Init()

			result := flex.View()

			// Both items should be present
			assert.Contains(t, result, "S")
			assert.Contains(t, result, "L")

			if tt.direction == FlexRow {
				// In row with AlignEnd, small item should be at bottom
				lines := strings.Split(result, "\n")
				// S should be on the last line
				lastLine := lines[len(lines)-1]
				assert.True(t, strings.Contains(lastLine, "S"), "S should be on last line (bottom aligned)")
			}
		})
	}
}

// TestFlex_AlignStretch_FillsAvailableSpace tests that AlignStretch fills
// the available cross-axis space.
func TestFlex_AlignStretch_FillsAvailableSpace(t *testing.T) {
	tests := []struct {
		name      string
		direction FlexDirection
	}{
		{name: "row direction (height stretch)", direction: FlexRow},
		{name: "column direction (width stretch)", direction: FlexColumn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create items with different sizes
			small := mockFlexComponent("S")
			large := mockFlexComponentWithSize("L\nL\nL", 5, 3)

			flex := Flex(FlexProps{
				Items:     []bubbly.Component{small, large},
				Direction: tt.direction,
				Align:     AlignItemsStretch,
			})
			flex.Init()

			result := flex.View()

			// Both items should be present
			assert.Contains(t, result, "S")
			assert.Contains(t, result, "L")

			if tt.direction == FlexRow {
				// In row with AlignStretch, result should have consistent height
				lines := strings.Split(result, "\n")
				assert.GreaterOrEqual(t, len(lines), 3, "Stretched items should have at least 3 lines")
			} else {
				// In column with AlignStretch, items should have same width
				lines := strings.Split(result, "\n")
				if len(lines) > 0 {
					maxWidth := 0
					for _, line := range lines {
						w := lipgloss.Width(line)
						if w > maxWidth {
							maxWidth = w
						}
					}
					assert.Greater(t, maxWidth, 1, "Stretched items should have width > 1")
				}
			}
		})
	}
}

// TestFlexAlignItemRow_Algorithm tests the row alignment algorithm directly.
func TestFlexAlignItemRow_Algorithm(t *testing.T) {
	tests := []struct {
		name      string
		item      string
		maxHeight int
		align     AlignItems
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:      "start alignment - adds bottom padding",
			item:      "X",
			maxHeight: 3,
			align:     AlignItemsStart,
			checkFunc: func(t *testing.T, result string) {
				height := lipgloss.Height(result)
				assert.Equal(t, 3, height, "Result should have maxHeight")
				lines := strings.Split(result, "\n")
				assert.True(t, strings.Contains(lines[0], "X"), "X should be on first line")
			},
		},
		{
			name:      "center alignment - adds top and bottom padding",
			item:      "X",
			maxHeight: 3,
			align:     AlignItemsCenter,
			checkFunc: func(t *testing.T, result string) {
				height := lipgloss.Height(result)
				assert.Equal(t, 3, height, "Result should have maxHeight")
				lines := strings.Split(result, "\n")
				assert.True(t, strings.Contains(lines[1], "X"), "X should be on middle line")
			},
		},
		{
			name:      "end alignment - adds top padding",
			item:      "X",
			maxHeight: 3,
			align:     AlignItemsEnd,
			checkFunc: func(t *testing.T, result string) {
				height := lipgloss.Height(result)
				assert.Equal(t, 3, height, "Result should have maxHeight")
				lines := strings.Split(result, "\n")
				assert.True(t, strings.Contains(lines[2], "X"), "X should be on last line")
			},
		},
		{
			name:      "stretch alignment - sets height",
			item:      "X",
			maxHeight: 3,
			align:     AlignItemsStretch,
			checkFunc: func(t *testing.T, result string) {
				height := lipgloss.Height(result)
				assert.Equal(t, 3, height, "Result should have maxHeight")
			},
		},
		{
			name:      "item already at max height - no change",
			item:      "X\nY\nZ",
			maxHeight: 3,
			align:     AlignItemsCenter,
			checkFunc: func(t *testing.T, result string) {
				assert.Contains(t, result, "X")
				assert.Contains(t, result, "Y")
				assert.Contains(t, result, "Z")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flexAlignItemRow(tt.item, tt.maxHeight, tt.align)
			tt.checkFunc(t, result)
		})
	}
}

// TestFlexAlignItemColumn_Algorithm tests the column alignment algorithm directly.
func TestFlexAlignItemColumn_Algorithm(t *testing.T) {
	tests := []struct {
		name      string
		item      string
		maxWidth  int
		align     AlignItems
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:     "start alignment - left aligned",
			item:     "X",
			maxWidth: 10,
			align:    AlignItemsStart,
			checkFunc: func(t *testing.T, result string) {
				width := lipgloss.Width(result)
				assert.Equal(t, 10, width, "Result should have maxWidth")
				// X should be at the start
				idx := strings.Index(result, "X")
				assert.LessOrEqual(t, idx, 1, "X should be left-aligned")
			},
		},
		{
			name:     "center alignment - centered",
			item:     "X",
			maxWidth: 10,
			align:    AlignItemsCenter,
			checkFunc: func(t *testing.T, result string) {
				width := lipgloss.Width(result)
				assert.Equal(t, 10, width, "Result should have maxWidth")
				// X should be in the middle
				idx := strings.Index(result, "X")
				assert.Greater(t, idx, 0, "X should have leading space")
				assert.Less(t, idx, 9, "X should have trailing space")
			},
		},
		{
			name:     "end alignment - right aligned",
			item:     "X",
			maxWidth: 10,
			align:    AlignItemsEnd,
			checkFunc: func(t *testing.T, result string) {
				width := lipgloss.Width(result)
				assert.Equal(t, 10, width, "Result should have maxWidth")
				// X should be at the end
				idx := strings.Index(result, "X")
				assert.GreaterOrEqual(t, idx, 8, "X should be right-aligned")
			},
		},
		{
			name:     "stretch alignment - full width",
			item:     "X",
			maxWidth: 10,
			align:    AlignItemsStretch,
			checkFunc: func(t *testing.T, result string) {
				width := lipgloss.Width(result)
				assert.Equal(t, 10, width, "Result should have maxWidth")
			},
		},
		{
			name:     "item already at max width - no change",
			item:     "XXXXXXXXXX",
			maxWidth: 10,
			align:    AlignItemsCenter,
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "XXXXXXXXXX", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flexAlignItemColumn(tt.item, tt.maxWidth, tt.align)
			tt.checkFunc(t, result)
		})
	}
}

// TestFlex_CrossAxisAlignment_WithDifferentSizedItems tests alignment with
// items of varying sizes.
func TestFlex_CrossAxisAlignment_WithDifferentSizedItems(t *testing.T) {
	tests := []struct {
		name      string
		align     AlignItems
		direction FlexDirection
	}{
		{name: "start row", align: AlignItemsStart, direction: FlexRow},
		{name: "center row", align: AlignItemsCenter, direction: FlexRow},
		{name: "end row", align: AlignItemsEnd, direction: FlexRow},
		{name: "stretch row", align: AlignItemsStretch, direction: FlexRow},
		{name: "start column", align: AlignItemsStart, direction: FlexColumn},
		{name: "center column", align: AlignItemsCenter, direction: FlexColumn},
		{name: "end column", align: AlignItemsEnd, direction: FlexColumn},
		{name: "stretch column", align: AlignItemsStretch, direction: FlexColumn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create items with significantly different sizes
			tiny := mockFlexComponent("T")
			small := mockFlexComponentWithSize("SM", 4, 2)
			large := mockFlexComponentWithSize("LG\nLG\nLG\nLG", 6, 4)

			flex := Flex(FlexProps{
				Items:     []bubbly.Component{tiny, small, large},
				Direction: tt.direction,
				Align:     tt.align,
				Gap:       1,
			})
			flex.Init()

			result := flex.View()

			// All items should be present
			assert.Contains(t, result, "T")
			assert.Contains(t, result, "SM")
			assert.Contains(t, result, "LG")

			// Result should have dimensions
			width := lipgloss.Width(result)
			height := lipgloss.Height(result)
			assert.Greater(t, width, 0, "Result should have width")
			assert.Greater(t, height, 0, "Result should have height")
		})
	}
}

// TestFlex_CrossAxisAlignment_SingleItem tests alignment with a single item.
func TestFlex_CrossAxisAlignment_SingleItem(t *testing.T) {
	tests := []struct {
		name      string
		align     AlignItems
		direction FlexDirection
	}{
		{name: "start row", align: AlignItemsStart, direction: FlexRow},
		{name: "center row", align: AlignItemsCenter, direction: FlexRow},
		{name: "end row", align: AlignItemsEnd, direction: FlexRow},
		{name: "stretch row", align: AlignItemsStretch, direction: FlexRow},
		{name: "start column", align: AlignItemsStart, direction: FlexColumn},
		{name: "center column", align: AlignItemsCenter, direction: FlexColumn},
		{name: "end column", align: AlignItemsEnd, direction: FlexColumn},
		{name: "stretch column", align: AlignItemsStretch, direction: FlexColumn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := mockFlexComponent("Single")

			flex := Flex(FlexProps{
				Items:     []bubbly.Component{item},
				Direction: tt.direction,
				Align:     tt.align,
			})
			flex.Init()

			result := flex.View()

			// Item should be present
			assert.Contains(t, result, "Single")
		})
	}
}
