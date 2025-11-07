package devtools

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestNewComponentInspector verifies inspector initialization.
func TestNewComponentInspector(t *testing.T) {
	// Create test component tree
	root := &ComponentSnapshot{
		ID:   "root",
		Name: "App",
		Type: "Component",
		Children: []*ComponentSnapshot{
			{
				ID:   "child1",
				Name: "Header",
				Type: "Component",
			},
		},
	}

	inspector := NewComponentInspector(root)

	assert.NotNil(t, inspector)
	assert.NotNil(t, inspector.tree)
	assert.NotNil(t, inspector.detail)
	assert.NotNil(t, inspector.search)
	assert.NotNil(t, inspector.filter)
}

// TestComponentInspector_Update_KeyboardNavigation tests keyboard navigation.
func TestComponentInspector_Update_KeyboardNavigation(t *testing.T) {
	tests := []struct {
		name           string
		initialSetup   func(*ComponentInspector)
		key            string
		expectedAction string
	}{
		{
			name: "down arrow selects next component",
			initialSetup: func(ci *ComponentInspector) {
				ci.tree.Expand("root")
			},
			key:            "down",
			expectedAction: "select_next",
		},
		{
			name: "up arrow selects previous component",
			initialSetup: func(ci *ComponentInspector) {
				ci.tree.Expand("root")
				ci.tree.SelectNext()
			},
			key:            "up",
			expectedAction: "select_previous",
		},
		{
			name: "enter toggles expansion",
			initialSetup: func(ci *ComponentInspector) {
				// Start with root selected
			},
			key:            "enter",
			expectedAction: "toggle_expand",
		},
		{
			name: "tab switches detail panel tabs",
			initialSetup: func(ci *ComponentInspector) {
				// Default tab is 0
			},
			key:            "tab",
			expectedAction: "next_tab",
		},
		{
			name: "shift+tab switches detail panel tabs backward",
			initialSetup: func(ci *ComponentInspector) {
				ci.detail.NextTab() // Move to tab 1
			},
			key:            "shift+tab",
			expectedAction: "previous_tab",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &ComponentSnapshot{
				ID:   "root",
				Name: "App",
				Type: "Component",
				Children: []*ComponentSnapshot{
					{
						ID:   "child1",
						Name: "Header",
						Type: "Component",
					},
				},
			}

			inspector := NewComponentInspector(root)
			if tt.initialSetup != nil {
				tt.initialSetup(inspector)
			}

			// Capture state before
			selectedBefore := inspector.tree.GetSelected()
			tabBefore := inspector.detail.GetActiveTab()
			expandedBefore := inspector.tree.IsExpanded("root")

			// Send key message
			keyMsg := tea.KeyMsg{Type: tea.KeyRunes}
			switch tt.key {
			case "down":
				keyMsg.Type = tea.KeyDown
			case "up":
				keyMsg.Type = tea.KeyUp
			case "enter":
				keyMsg.Type = tea.KeyEnter
			case "tab":
				keyMsg.Type = tea.KeyTab
			case "shift+tab":
				keyMsg.Type = tea.KeyShiftTab
			}

			cmd := inspector.Update(keyMsg)

			// Verify state changed appropriately
			switch tt.expectedAction {
			case "select_next":
				selectedAfter := inspector.tree.GetSelected()
				assert.NotEqual(t, selectedBefore, selectedAfter, "Selection should change")
			case "select_previous":
				selectedAfter := inspector.tree.GetSelected()
				assert.NotEqual(t, selectedBefore, selectedAfter, "Selection should change")
			case "toggle_expand":
				expandedAfter := inspector.tree.IsExpanded("root")
				assert.NotEqual(t, expandedBefore, expandedAfter, "Expansion should toggle")
			case "next_tab":
				tabAfter := inspector.detail.GetActiveTab()
				assert.NotEqual(t, tabBefore, tabAfter, "Tab should change")
			case "previous_tab":
				tabAfter := inspector.detail.GetActiveTab()
				assert.NotEqual(t, tabBefore, tabAfter, "Tab should change")
			}

			assert.Nil(t, cmd, "No commands should be returned for navigation")
		})
	}
}

// TestComponentInspector_Update_SearchMode tests search functionality.
func TestComponentInspector_Update_SearchMode(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root",
		Name: "App",
		Type: "Component",
		Children: []*ComponentSnapshot{
			{
				ID:   "button1",
				Name: "Button",
				Type: "Button",
			},
			{
				ID:   "input1",
				Name: "Input",
				Type: "Input",
			},
		},
	}

	inspector := NewComponentInspector(root)

	// Enter search mode with ctrl+f
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlF}
	cmd := inspector.Update(keyMsg)
	assert.Nil(t, cmd)

	// Verify search mode is active
	assert.True(t, inspector.searchMode, "Search mode should be active")

	// Type search query
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b', 'u', 't'}}
	cmd = inspector.Update(keyMsg)
	assert.Nil(t, cmd)

	// Verify search was performed
	assert.Equal(t, "but", inspector.search.GetQuery())
	assert.Greater(t, inspector.search.GetResultCount(), 0)

	// Exit search mode with esc
	keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
	cmd = inspector.Update(keyMsg)
	assert.Nil(t, cmd)
	assert.False(t, inspector.searchMode, "Search mode should be inactive")
}

// TestComponentInspector_View tests rendering integration.
func TestComponentInspector_View(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root",
		Name: "App",
		Type: "Component",
		Refs: []*RefSnapshot{
			{
				ID:    "ref1",
				Name:  "count",
				Type:  "int",
				Value: 42,
			},
		},
	}

	inspector := NewComponentInspector(root)

	// Get rendered output
	output := inspector.View()

	// Verify output contains expected elements
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "App", "Should contain component name")
}

// TestComponentInspector_LiveUpdates tests component updates.
func TestComponentInspector_LiveUpdates(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root",
		Name: "App",
		Type: "Component",
	}

	inspector := NewComponentInspector(root)

	// Update with new component tree
	newRoot := &ComponentSnapshot{
		ID:   "root",
		Name: "App",
		Type: "Component",
		Children: []*ComponentSnapshot{
			{
				ID:   "new-child",
				Name: "NewChild",
				Type: "Component",
			},
		},
	}

	inspector.SetRoot(newRoot)

	// Verify tree was updated
	assert.Equal(t, newRoot, inspector.tree.GetRoot())
}

// TestComponentInspector_FilterIntegration tests filter integration.
func TestComponentInspector_FilterIntegration(t *testing.T) {
	root := &ComponentSnapshot{
		ID:     "root",
		Name:   "App",
		Type:   "Component",
		Status: "mounted",
		Children: []*ComponentSnapshot{
			{
				ID:     "button1",
				Name:   "Button",
				Type:   "Button",
				Status: "mounted",
			},
			{
				ID:     "input1",
				Name:   "Input",
				Type:   "Input",
				Status: "unmounted",
			},
		},
	}

	inspector := NewComponentInspector(root)

	// Apply filter
	inspector.filter.WithStatuses([]string{"mounted"})
	inspector.ApplyFilter()

	// Perform a search to see filtered results
	inspector.search.Search("")

	// Verify filtered results in search
	results := inspector.search.GetResults()
	assert.Equal(t, 2, len(results), "Should have 2 results (root + button)")
}

// TestComponentInspector_ThreadSafety tests concurrent access.
func TestComponentInspector_ThreadSafety(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root",
		Name: "App",
		Type: "Component",
	}

	inspector := NewComponentInspector(root)

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			inspector.Update(tea.KeyMsg{Type: tea.KeyDown})
			inspector.View()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we get here without panic, thread safety is good
	assert.True(t, true)
}

// TestComponentInspector_E2E tests end-to-end workflow.
func TestComponentInspector_E2E(t *testing.T) {
	// Create component tree
	root := &ComponentSnapshot{
		ID:   "root",
		Name: "App",
		Type: "Component",
		Children: []*ComponentSnapshot{
			{
				ID:   "header",
				Name: "Header",
				Type: "Component",
				Refs: []*RefSnapshot{
					{
						ID:    "title",
						Name:  "title",
						Type:  "string",
						Value: "My App",
					},
				},
			},
			{
				ID:   "counter",
				Name: "Counter",
				Type: "Component",
				Refs: []*RefSnapshot{
					{
						ID:    "count",
						Name:  "count",
						Type:  "int",
						Value: 42,
					},
				},
			},
		},
	}

	inspector := NewComponentInspector(root)

	// 1. Expand root
	inspector.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.True(t, inspector.tree.IsExpanded("root"))

	// 2. Navigate to first child
	inspector.Update(tea.KeyMsg{Type: tea.KeyDown})
	selected := inspector.tree.GetSelected()
	assert.NotNil(t, selected)
	assert.Equal(t, "header", selected.ID)

	// 3. Verify detail panel shows selected component
	assert.Equal(t, selected, inspector.detail.GetComponent())

	// 4. Switch to State tab
	inspector.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, 1, inspector.detail.GetActiveTab())

	// 5. Search for "counter"
	inspector.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
	inspector.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("counter")})
	assert.Equal(t, "counter", inspector.search.GetQuery())
	assert.Equal(t, 1, inspector.search.GetResultCount())

	// 6. Render final state
	output := inspector.View()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Counter")
}
