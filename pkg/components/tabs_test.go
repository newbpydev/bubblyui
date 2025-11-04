package components

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

func TestTabs_Creation(t *testing.T) {
	tabs := Tabs(TabsProps{
		Tabs: []Tab{
			{Label: "Tab 1", Content: "Content 1"},
			{Label: "Tab 2", Content: "Content 2"},
		},
	})

	assert.NotNil(t, tabs, "Tabs should be created")
}

func TestTabs_Rendering(t *testing.T) {
	tabs := Tabs(TabsProps{
		Tabs: []Tab{
			{Label: "Profile", Content: "Profile content"},
			{Label: "Settings", Content: "Settings content"},
		},
	})

	tabs.Init()
	output := tabs.View()

	assert.Contains(t, output, "Profile", "Should render Profile tab")
	assert.Contains(t, output, "Settings", "Should render Settings tab")
	assert.Contains(t, output, "Profile content", "Should render active tab content")
}

func TestTabs_ActiveTab(t *testing.T) {
	activeIndex := bubbly.NewRef(1)
	tabs := Tabs(TabsProps{
		Tabs: []Tab{
			{Label: "Tab 1", Content: "Content 1"},
			{Label: "Tab 2", Content: "Content 2"},
		},
		ActiveIndex: activeIndex,
	})

	tabs.Init()
	output := tabs.View()

	assert.Contains(t, output, "Content 2", "Should render second tab content")
}

func TestTabs_OnTabChange(t *testing.T) {
	var changedIndex int
	tabs := Tabs(TabsProps{
		Tabs: []Tab{
			{Label: "Tab 1", Content: "Content 1"},
			{Label: "Tab 2", Content: "Content 2"},
		},
		OnTabChange: func(index int) {
			changedIndex = index
		},
	})

	tabs.Init()
	tabs.Emit("changeTab", 1)

	assert.Equal(t, 1, changedIndex, "OnTabChange should be called with index")
}

func TestTabs_WithComponent(t *testing.T) {
	child := Text(TextProps{
		Content: "Component content",
	})
	child.Init()

	tabs := Tabs(TabsProps{
		Tabs: []Tab{
			{Label: "Tab 1", Component: child},
		},
	})

	tabs.Init()
	output := tabs.View()

	assert.Contains(t, output, "Component content", "Should render component content")
}

func TestTabs_EmptyTabs(t *testing.T) {
	tabs := Tabs(TabsProps{
		Tabs: []Tab{},
	})

	tabs.Init()
	output := tabs.View()

	assert.Empty(t, output, "Should render nothing with empty tabs")
}

func TestTabs_ThemeIntegration(t *testing.T) {
	tabs := Tabs(TabsProps{
		Tabs: []Tab{
			{Label: "Test", Content: "Content"},
		},
	})

	tabs.Init()
	output := tabs.View()

	assert.NotEmpty(t, output, "Should render with theme")
}
