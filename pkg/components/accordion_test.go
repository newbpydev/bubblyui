package components

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

func TestAccordion_Creation(t *testing.T) {
	accordion := Accordion(AccordionProps{
		Items: []AccordionItem{
			{Title: "Section 1", Content: "Content 1"},
			{Title: "Section 2", Content: "Content 2"},
		},
	})

	assert.NotNil(t, accordion, "Accordion should be created")
}

func TestAccordion_Rendering(t *testing.T) {
	accordion := Accordion(AccordionProps{
		Items: []AccordionItem{
			{Title: "Section 1", Content: "Content 1"},
			{Title: "Section 2", Content: "Content 2"},
		},
	})

	accordion.Init()
	output := accordion.View()

	assert.Contains(t, output, "Section 1", "Should render Section 1 title")
	assert.Contains(t, output, "Section 2", "Should render Section 2 title")
}

func TestAccordion_ExpandedPanel(t *testing.T) {
	expanded := bubbly.NewRef([]int{0})
	accordion := Accordion(AccordionProps{
		Items: []AccordionItem{
			{Title: "Section 1", Content: "Content 1"},
			{Title: "Section 2", Content: "Content 2"},
		},
		ExpandedIndexes: expanded,
	})

	accordion.Init()
	output := accordion.View()

	assert.Contains(t, output, "Content 1", "Should render expanded content")
	assert.NotContains(t, output, "Content 2", "Should not render collapsed content")
}

func TestAccordion_Toggle(t *testing.T) {
	var toggledIndex int
	var toggledState bool

	expanded := bubbly.NewRef([]int{})
	accordion := Accordion(AccordionProps{
		Items: []AccordionItem{
			{Title: "Section 1", Content: "Content 1"},
		},
		ExpandedIndexes: expanded,
		OnToggle: func(index int, state bool) {
			toggledIndex = index
			toggledState = state
		},
	})

	accordion.Init()
	accordion.Emit("toggle", 0)

	assert.Equal(t, 0, toggledIndex, "OnToggle should be called with index")
	assert.True(t, toggledState, "OnToggle should be called with expanded state")
}

func TestAccordion_MultipleExpanded(t *testing.T) {
	expanded := bubbly.NewRef([]int{0, 1})
	accordion := Accordion(AccordionProps{
		Items: []AccordionItem{
			{Title: "Section 1", Content: "Content 1"},
			{Title: "Section 2", Content: "Content 2"},
		},
		ExpandedIndexes: expanded,
		AllowMultiple:   true,
	})

	accordion.Init()
	output := accordion.View()

	assert.Contains(t, output, "Content 1", "Should render first expanded content")
	assert.Contains(t, output, "Content 2", "Should render second expanded content")
}

func TestAccordion_WithComponent(t *testing.T) {
	child := Text(TextProps{
		Content: "Component content",
	})
	child.Init()

	expanded := bubbly.NewRef([]int{0})
	accordion := Accordion(AccordionProps{
		Items: []AccordionItem{
			{Title: "Section 1", Component: child},
		},
		ExpandedIndexes: expanded,
	})

	accordion.Init()
	output := accordion.View()

	assert.Contains(t, output, "Component content", "Should render component content")
}

func TestAccordion_EmptyItems(t *testing.T) {
	accordion := Accordion(AccordionProps{
		Items: []AccordionItem{},
	})

	accordion.Init()
	output := accordion.View()

	assert.Empty(t, output, "Should render nothing with empty items")
}

func TestAccordion_ThemeIntegration(t *testing.T) {
	accordion := Accordion(AccordionProps{
		Items: []AccordionItem{
			{Title: "Test", Content: "Content"},
		},
	})

	accordion.Init()
	output := accordion.View()

	assert.NotEmpty(t, output, "Should render with theme")
}

// ============================================================================
// ACCORDION TOGGLE EXPANDED TESTS - Additional Coverage
// ============================================================================

func TestAccordion_Toggle_SingleMode_Collapse(t *testing.T) {
	// Test collapsing an expanded item in single mode (AllowMultiple = false)
	expanded := bubbly.NewRef([]int{0})
	var toggledState bool

	accordion := Accordion(AccordionProps{
		Items: []AccordionItem{
			{Title: "Section 1", Content: "Content 1"},
		},
		ExpandedIndexes: expanded,
		AllowMultiple:   false, // Single mode
		OnToggle: func(index int, state bool) {
			toggledState = state
		},
	})

	accordion.Init()

	// Toggle same index to collapse
	accordion.Emit("toggle", 0)

	// In single mode, toggling an expanded item should collapse it
	assert.False(t, toggledState, "Should collapse in single mode")
}

func TestAccordion_Toggle_MultipleMode_CollapseOne(t *testing.T) {
	// Test collapsing one item when multiple are expanded
	expanded := bubbly.NewRef([]int{0, 1})
	var toggledIndex int
	var toggledState bool

	accordion := Accordion(AccordionProps{
		Items: []AccordionItem{
			{Title: "Section 1", Content: "Content 1"},
			{Title: "Section 2", Content: "Content 2"},
			{Title: "Section 3", Content: "Content 3"},
		},
		ExpandedIndexes: expanded,
		AllowMultiple:   true,
		OnToggle: func(index int, state bool) {
			toggledIndex = index
			toggledState = state
		},
	})

	accordion.Init()

	// Toggle index 0 to collapse it
	accordion.Emit("toggle", 0)

	assert.Equal(t, 0, toggledIndex, "Should toggle index 0")
	assert.False(t, toggledState, "Should collapse index 0")

	// Verify expanded indexes updated
	currentExpanded := expanded.GetTyped()
	assert.NotContains(t, currentExpanded, 0, "Index 0 should be removed")
	assert.Contains(t, currentExpanded, 1, "Index 1 should still be expanded")
}

func TestAccordion_Toggle_MultipleMode_ExpandNew(t *testing.T) {
	// Test expanding a new item when multiple are allowed
	expanded := bubbly.NewRef([]int{0})
	var toggledIndex int
	var toggledState bool

	accordion := Accordion(AccordionProps{
		Items: []AccordionItem{
			{Title: "Section 1", Content: "Content 1"},
			{Title: "Section 2", Content: "Content 2"},
		},
		ExpandedIndexes: expanded,
		AllowMultiple:   true,
		OnToggle: func(index int, state bool) {
			toggledIndex = index
			toggledState = state
		},
	})

	accordion.Init()

	// Toggle index 1 to expand it
	accordion.Emit("toggle", 1)

	assert.Equal(t, 1, toggledIndex, "Should toggle index 1")
	assert.True(t, toggledState, "Should expand index 1")

	// Verify both are now expanded
	currentExpanded := expanded.GetTyped()
	assert.Contains(t, currentExpanded, 0, "Index 0 should still be expanded")
	assert.Contains(t, currentExpanded, 1, "Index 1 should now be expanded")
}

func TestAccordion_Toggle_NilExpandedIndexes(t *testing.T) {
	// Test toggle when ExpandedIndexes is nil
	accordion := Accordion(AccordionProps{
		Items: []AccordionItem{
			{Title: "Section 1", Content: "Content 1"},
		},
		// ExpandedIndexes is nil
	})

	accordion.Init()

	// Should not panic when toggling with nil ExpandedIndexes
	assert.NotPanics(t, func() {
		accordion.Emit("toggle", 0)
	})
}

func TestAccordion_Toggle_NoCallback(t *testing.T) {
	// Test toggle without OnToggle callback
	expanded := bubbly.NewRef([]int{})

	accordion := Accordion(AccordionProps{
		Items: []AccordionItem{
			{Title: "Section 1", Content: "Content 1"},
		},
		ExpandedIndexes: expanded,
		// OnToggle is nil
	})

	accordion.Init()

	// Should not panic when toggling without callback
	assert.NotPanics(t, func() {
		accordion.Emit("toggle", 0)
	})

	// But state should still update
	currentExpanded := expanded.GetTyped()
	assert.Contains(t, currentExpanded, 0, "Index 0 should be expanded")
}

func TestAccordion_Toggle_MultipleMode_RemoveFromMiddle(t *testing.T) {
	// Test removing an item from the middle of expanded list
	expanded := bubbly.NewRef([]int{0, 1, 2})

	accordion := Accordion(AccordionProps{
		Items: []AccordionItem{
			{Title: "Section 1", Content: "Content 1"},
			{Title: "Section 2", Content: "Content 2"},
			{Title: "Section 3", Content: "Content 3"},
		},
		ExpandedIndexes: expanded,
		AllowMultiple:   true,
	})

	accordion.Init()

	// Collapse index 1 (middle)
	accordion.Emit("toggle", 1)

	currentExpanded := expanded.GetTyped()
	assert.Contains(t, currentExpanded, 0, "Index 0 should still be expanded")
	assert.NotContains(t, currentExpanded, 1, "Index 1 should be collapsed")
	assert.Contains(t, currentExpanded, 2, "Index 2 should still be expanded")
}

func TestAccordion_SingleMode_Switch(t *testing.T) {
	// Test switching expanded item in single mode
	expanded := bubbly.NewRef([]int{0})

	accordion := Accordion(AccordionProps{
		Items: []AccordionItem{
			{Title: "Section 1", Content: "Content 1"},
			{Title: "Section 2", Content: "Content 2"},
		},
		ExpandedIndexes: expanded,
		AllowMultiple:   false, // Single mode
	})

	accordion.Init()

	// Expand index 1 (should replace index 0)
	accordion.Emit("toggle", 1)

	currentExpanded := expanded.GetTyped()
	assert.NotContains(t, currentExpanded, 0, "Index 0 should be collapsed in single mode")
	assert.Contains(t, currentExpanded, 1, "Index 1 should be expanded")
}
