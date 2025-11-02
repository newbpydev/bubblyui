package components

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
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
