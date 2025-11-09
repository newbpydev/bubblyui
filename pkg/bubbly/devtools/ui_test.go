package devtools

import (
	"strings"
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestNewDevToolsUI tests creating a new DevTools UI.
func TestNewDevToolsUI(t *testing.T) {
	store := NewDevToolsStore(1000, 1000)
	ui := NewDevToolsUI(store)

	assert.NotNil(t, ui)
	assert.NotNil(t, ui.layout)
	assert.NotNil(t, ui.inspector)
	assert.NotNil(t, ui.state)
	assert.NotNil(t, ui.events)
	assert.NotNil(t, ui.perf)
	assert.NotNil(t, ui.timeline)
	assert.NotNil(t, ui.keyboard)
	assert.NotNil(t, ui.tabs)
	assert.Equal(t, 0, ui.GetActivePanel())
}

// TestDevToolsUI_SetAppContent tests setting app content.
func TestDevToolsUI_SetAppContent(t *testing.T) {
	store := NewDevToolsStore(1000, 1000)
	ui := NewDevToolsUI(store)

	content := "Test App Content"
	ui.SetAppContent(content)

	// Verify by rendering
	output := ui.View()
	assert.Contains(t, output, content)
}

// TestDevToolsUI_PanelSwitching tests switching between panels.
func TestDevToolsUI_PanelSwitching(t *testing.T) {
	tests := []struct {
		name        string
		panelIndex  int
		expectError bool
	}{
		{
			name:        "switch to inspector",
			panelIndex:  0,
			expectError: false,
		},
		{
			name:        "switch to state viewer",
			panelIndex:  1,
			expectError: false,
		},
		{
			name:        "switch to event tracker",
			panelIndex:  2,
			expectError: false,
		},
		{
			name:        "switch to performance monitor",
			panelIndex:  3,
			expectError: false,
		},
		{
			name:        "switch to timeline",
			panelIndex:  4,
			expectError: false,
		},
		{
			name:        "invalid panel index",
			panelIndex:  10,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewDevToolsStore(1000, 1000)
			ui := NewDevToolsUI(store)

			ui.SetActivePanel(tt.panelIndex)

			if !tt.expectError {
				assert.Equal(t, tt.panelIndex, ui.GetActivePanel())
			} else {
				// Invalid index should not change active panel
				assert.Equal(t, 0, ui.GetActivePanel())
			}
		})
	}
}

// TestDevToolsUI_Update_TabSwitching tests tab switching via keyboard.
func TestDevToolsUI_Update_TabSwitching(t *testing.T) {
	store := NewDevToolsStore(1000, 1000)
	ui := NewDevToolsUI(store)

	// Initial panel is 0
	assert.Equal(t, 0, ui.GetActivePanel())

	// Press Tab to switch to next panel
	keyMsg := tea.KeyMsg{Type: tea.KeyTab}
	updatedUI, _ := ui.Update(keyMsg)
	ui = updatedUI.(*DevToolsUI)

	assert.Equal(t, 1, ui.GetActivePanel())

	// Press Tab again
	updatedUI, _ = ui.Update(keyMsg)
	ui = updatedUI.(*DevToolsUI)

	assert.Equal(t, 2, ui.GetActivePanel())
}

// TestDevToolsUI_Update_ShiftTabSwitching tests reverse tab switching.
func TestDevToolsUI_Update_ShiftTabSwitching(t *testing.T) {
	store := NewDevToolsStore(1000, 1000)
	ui := NewDevToolsUI(store)

	// Set to panel 2
	ui.SetActivePanel(2)

	// Press Shift+Tab to go back
	keyMsg := tea.KeyMsg{Type: tea.KeyShiftTab}
	updatedUI, _ := ui.Update(keyMsg)
	ui = updatedUI.(*DevToolsUI)

	assert.Equal(t, 1, ui.GetActivePanel())
}

// TestDevToolsUI_View tests rendering the UI.
func TestDevToolsUI_View(t *testing.T) {
	store := NewDevToolsStore(1000, 1000)
	ui := NewDevToolsUI(store)

	ui.SetAppContent("My Application")

	output := ui.View()

	// Should contain app content
	assert.Contains(t, output, "My Application")

	// Should contain some panel content (inspector is default)
	assert.NotEmpty(t, output)
}

// TestDevToolsUI_LayoutMode tests changing layout modes.
func TestDevToolsUI_LayoutMode(t *testing.T) {
	tests := []struct {
		name string
		mode LayoutMode
	}{
		{
			name: "horizontal layout",
			mode: LayoutHorizontal,
		},
		{
			name: "vertical layout",
			mode: LayoutVertical,
		},
		{
			name: "overlay layout",
			mode: LayoutOverlay,
		},
		{
			name: "hidden layout",
			mode: LayoutHidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewDevToolsStore(1000, 1000)
			ui := NewDevToolsUI(store)

			ui.SetLayoutMode(tt.mode)
			assert.Equal(t, tt.mode, ui.GetLayoutMode())

			// Verify rendering doesn't crash
			output := ui.View()
			// Hidden mode may return empty string or just app content
			if tt.mode != LayoutHidden {
				assert.NotEmpty(t, output)
			}
		})
	}
}

// TestDevToolsUI_LayoutRatio tests changing layout ratio.
func TestDevToolsUI_LayoutRatio(t *testing.T) {
	tests := []struct {
		name  string
		ratio float64
	}{
		{
			name:  "50/50 split",
			ratio: 0.5,
		},
		{
			name:  "60/40 split",
			ratio: 0.6,
		},
		{
			name:  "70/30 split",
			ratio: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewDevToolsStore(1000, 1000)
			ui := NewDevToolsUI(store)

			ui.SetLayoutRatio(tt.ratio)
			assert.Equal(t, tt.ratio, ui.GetLayoutRatio())
		})
	}
}

// TestDevToolsUI_Integration tests E2E workflow.
func TestDevToolsUI_Integration(t *testing.T) {
	store := NewDevToolsStore(1000, 1000)

	// Add some test data
	store.AddComponent(&ComponentSnapshot{
		ID:   "comp-1",
		Name: "TestComponent",
		Refs: []*RefSnapshot{
			{ID: "ref-1", Name: "count", Value: 42, Type: "int"},
		},
	})

	ui := NewDevToolsUI(store)
	ui.SetAppContent("Test Application")

	// Test 1: Initial render
	output := ui.View()
	assert.Contains(t, output, "Test Application")

	// Test 2: Switch to state viewer panel
	ui.SetActivePanel(1)
	output = ui.View()
	assert.NotEmpty(t, output)

	// Test 3: Switch to event tracker panel
	ui.SetActivePanel(2)
	output = ui.View()
	assert.NotEmpty(t, output)

	// Test 4: Change layout mode
	ui.SetLayoutMode(LayoutVertical)
	output = ui.View()
	assert.NotEmpty(t, output)

	// Test 5: Change layout ratio
	ui.SetLayoutRatio(0.7)
	output = ui.View()
	assert.NotEmpty(t, output)
}

// TestDevToolsUI_Concurrent tests thread-safe concurrent access.
func TestDevToolsUI_Concurrent(t *testing.T) {
	store := NewDevToolsStore(1000, 1000)
	ui := NewDevToolsUI(store)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Concurrent operations
			switch idx % 5 {
			case 0:
				ui.SetActivePanel(idx % 5)
			case 1:
				_ = ui.GetActivePanel()
			case 2:
				ui.SetLayoutMode(LayoutMode(idx % 4))
			case 3:
				_ = ui.View()
			case 4:
				ui.SetAppContent("Test")
			}
		}(i)
	}

	wg.Wait()
}

// TestDevToolsUI_KeyboardShortcuts tests keyboard shortcut handling.
func TestDevToolsUI_KeyboardShortcuts(t *testing.T) {
	tests := []struct {
		name     string
		key      tea.KeyMsg
		validate func(*testing.T, *DevToolsUI)
	}{
		{
			name: "tab switches panel",
			key:  tea.KeyMsg{Type: tea.KeyTab},
			validate: func(t *testing.T, ui *DevToolsUI) {
				assert.Equal(t, 1, ui.GetActivePanel())
			},
		},
		{
			name: "shift+tab switches panel backward",
			key:  tea.KeyMsg{Type: tea.KeyShiftTab},
			validate: func(t *testing.T, ui *DevToolsUI) {
				// From panel 0, shift+tab wraps to last panel
				assert.Equal(t, 4, ui.GetActivePanel())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewDevToolsStore(1000, 1000)
			ui := NewDevToolsUI(store)

			updatedUI, _ := ui.Update(tt.key)
			ui = updatedUI.(*DevToolsUI)

			tt.validate(t, ui)
		})
	}
}

// TestDevToolsUI_EmptyStore tests UI with empty store.
func TestDevToolsUI_EmptyStore(t *testing.T) {
	store := NewDevToolsStore(1000, 1000)
	ui := NewDevToolsUI(store)

	// Should not crash with empty store
	output := ui.View()
	assert.NotEmpty(t, output)

	// Switch through all panels
	for i := 0; i < 5; i++ {
		ui.SetActivePanel(i)
		output = ui.View()
		assert.NotEmpty(t, output)
	}
}

// TestDevToolsUI_PanelContent tests that each panel renders its content.
func TestDevToolsUI_PanelContent(t *testing.T) {
	store := NewDevToolsStore(1000, 1000)

	// Add test data
	rootSnapshot := &ComponentSnapshot{
		ID:   "comp-1",
		Name: "TestComponent",
	}
	store.AddComponent(rootSnapshot)

	ui := NewDevToolsUI(store)
	// Set root component in inspector
	ui.inspector.SetRoot(rootSnapshot)

	tests := []struct {
		name       string
		panelIndex int
		shouldFind []string
	}{
		{
			name:       "inspector panel",
			panelIndex: 0,
			shouldFind: []string{"TestComponent"},
		},
		{
			name:       "state viewer panel",
			panelIndex: 1,
			shouldFind: []string{"Reactive State"},
		},
		{
			name:       "event tracker panel",
			panelIndex: 2,
			shouldFind: []string{"Event"},
		},
		{
			name:       "performance monitor panel",
			panelIndex: 3,
			shouldFind: []string{"Performance"},
		},
		{
			name:       "timeline panel",
			panelIndex: 4,
			shouldFind: []string{"Timeline"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui.SetActivePanel(tt.panelIndex)
			output := ui.View()

			for _, expected := range tt.shouldFind {
				assert.True(t, strings.Contains(output, expected),
					"Expected to find '%s' in panel %d output", expected, tt.panelIndex)
			}
		})
	}
}
