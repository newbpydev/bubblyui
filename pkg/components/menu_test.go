package components

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
)

func TestMenu_Creation(t *testing.T) {
	menu := Menu(MenuProps{
		Items: []MenuItem{
			{Label: "Home", Value: "home"},
			{Label: "Settings", Value: "settings"},
		},
	})

	assert.NotNil(t, menu, "Menu should be created")
}

func TestMenu_Rendering(t *testing.T) {
	menu := Menu(MenuProps{
		Items: []MenuItem{
			{Label: "Home", Value: "home"},
			{Label: "Settings", Value: "settings"},
			{Label: "Logout", Value: "logout"},
		},
	})

	menu.Init()
	output := menu.View()

	assert.Contains(t, output, "Home", "Should render Home item")
	assert.Contains(t, output, "Settings", "Should render Settings item")
	assert.Contains(t, output, "Logout", "Should render Logout item")
}

func TestMenu_Selection(t *testing.T) {
	selected := bubbly.NewRef("")
	menu := Menu(MenuProps{
		Items: []MenuItem{
			{Label: "Home", Value: "home"},
			{Label: "Settings", Value: "settings"},
		},
		Selected: selected,
	})

	menu.Init()

	// Select an item
	selected.Set("home")
	output := menu.View()

	assert.Contains(t, output, "Home", "Should render selected item")
}

func TestMenu_OnSelect(t *testing.T) {
	var selectedValue string
	menu := Menu(MenuProps{
		Items: []MenuItem{
			{Label: "Home", Value: "home"},
			{Label: "Settings", Value: "settings"},
		},
		OnSelect: func(value string) {
			selectedValue = value
		},
	})

	menu.Init()
	menu.Emit("select", "settings")

	assert.Equal(t, "settings", selectedValue, "OnSelect should be called with value")
}

func TestMenu_DisabledItem(t *testing.T) {
	menu := Menu(MenuProps{
		Items: []MenuItem{
			{Label: "Home", Value: "home"},
			{Label: "Disabled", Value: "disabled", Disabled: true},
		},
	})

	menu.Init()
	output := menu.View()

	assert.Contains(t, output, "Disabled", "Should render disabled item")
}

func TestMenu_EmptyItems(t *testing.T) {
	menu := Menu(MenuProps{
		Items: []MenuItem{},
	})

	menu.Init()
	output := menu.View()

	assert.NotEmpty(t, output, "Should render even with empty items")
}

func TestMenu_ThemeIntegration(t *testing.T) {
	menu := Menu(MenuProps{
		Items: []MenuItem{
			{Label: "Test", Value: "test"},
		},
	})

	menu.Init()
	output := menu.View()

	assert.NotEmpty(t, output, "Should render with theme")
}
