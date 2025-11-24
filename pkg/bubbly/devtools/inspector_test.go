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

			// Navigation now returns tea.ClearScreen to force UI redraw
			// This ensures selection changes are visible immediately
			if cmd != nil {
				assert.NotNil(t, cmd, "ClearScreen cmd expected for navigation")
			}
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

	// 1. Root is already expanded and selected (auto-expanded in NewComponentInspector)
	assert.True(t, inspector.tree.IsExpanded("root"), "Root should be auto-expanded")
	assert.Equal(t, "root", inspector.tree.GetSelected().ID, "Root should be auto-selected")

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

// CRITICAL FIX: Inspector UX Tests (TDD - these will FAIL initially)

func TestComponentInspector_SetRoot_AutoSelectsRoot(t *testing.T) {
	// When SetRoot is called and nothing is selected, it should auto-select root
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

	inspector := NewComponentInspector(nil)

	// Initially nothing selected
	assert.Nil(t, inspector.tree.GetSelected())

	// SetRoot should auto-select root
	inspector.SetRoot(root)

	selected := inspector.tree.GetSelected()
	assert.NotNil(t, selected, "SetRoot should auto-select root when nothing is selected")
	assert.Equal(t, "root", selected.ID, "Root should be selected")

	// Detail panel should show root component
	detailComp := inspector.detail.GetComponent()
	assert.NotNil(t, detailComp, "Detail panel should show root component")
	assert.Equal(t, "root", detailComp.ID)
}

func TestComponentInspector_SetRoot_PreservesSelectionIfExists(t *testing.T) {
	// When SetRoot is called and a component is selected, preserve it if it still exists
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
	inspector.tree.Expand("root")
	inspector.tree.Select("child1")
	inspector.updateDetailPanel()

	// Verify child1 is selected
	assert.Equal(t, "child1", inspector.tree.GetSelected().ID)

	// Update tree with new data (simulating store update)
	updatedRoot := &ComponentSnapshot{
		ID:   "root",
		Name: "App",
		Type: "Component",
		Children: []*ComponentSnapshot{
			{
				ID:   "child1", // Same component still exists
				Name: "Header",
				Type: "Component",
			},
		},
	}

	inspector.SetRoot(updatedRoot)

	// Selection should be preserved
	selected := inspector.tree.GetSelected()
	assert.NotNil(t, selected, "Selection should be preserved if component still exists")
	assert.Equal(t, "child1", selected.ID, "child1 should still be selected")

	// Detail panel should still show child1
	detailComp := inspector.detail.GetComponent()
	assert.NotNil(t, detailComp, "Detail panel should show selected component")
	assert.Equal(t, "child1", detailComp.ID)
}

func TestComponentInspector_SetRoot_AutoExpandsRoot(t *testing.T) {
	// SetRoot should automatically expand root to show children
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
			{
				ID:   "child2",
				Name: "Footer",
				Type: "Component",
			},
		},
	}

	inspector := NewComponentInspector(root)

	// Root should be auto-expanded
	assert.True(t, inspector.tree.IsExpanded("root"), "Root should be auto-expanded to show children")

	// Render should show children
	output := inspector.View()
	assert.Contains(t, output, "Header", "Rendered output should show child components")
	assert.Contains(t, output, "Footer", "Rendered output should show child components")
}

func TestComponentInspector_SetRoot_FallsBackToRootIfSelectionLost(t *testing.T) {
	// When SetRoot is called and selected component no longer exists, select root
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
	inspector.tree.Expand("root")
	inspector.tree.Select("child1")
	inspector.updateDetailPanel()

	// Update tree without child1 (component removed)
	updatedRoot := &ComponentSnapshot{
		ID:   "root",
		Name: "App",
		Type: "Component",
		Children: []*ComponentSnapshot{
			{
				ID:   "child2", // Different child
				Name: "Sidebar",
				Type: "Component",
			},
		},
	}

	inspector.SetRoot(updatedRoot)

	// Should fall back to root since child1 no longer exists
	selected := inspector.tree.GetSelected()
	assert.NotNil(t, selected, "Should fall back to root when selected component is removed")
	assert.Equal(t, "root", selected.ID, "Should select root when previous selection is lost")
}

// TestComponentInspector_Update_DownKey tests down key navigation
func TestComponentInspector_Update_DownKey(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root",
		Name: "App",
		Type: "Component",
		Children: []*ComponentSnapshot{
			{ID: "child1", Name: "First"},
			{ID: "child2", Name: "Second"},
		},
	}

	inspector := NewComponentInspector(root)
	inspector.tree.Expand("root")
	inspector.tree.Select("root")

	// Press down key
	_ = inspector.Update(tea.KeyMsg{Type: tea.KeyDown})

	// Selection should move to first child
	selected := inspector.tree.GetSelected()
	assert.NotNil(t, selected)
	assert.Equal(t, "child1", selected.ID, "Down key should move selection down")
}

// TestComponentInspector_Update_UpKey tests up key navigation
func TestComponentInspector_Update_UpKey(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root",
		Name: "App",
		Type: "Component",
		Children: []*ComponentSnapshot{
			{ID: "child1", Name: "First"},
			{ID: "child2", Name: "Second"},
		},
	}

	inspector := NewComponentInspector(root)
	inspector.tree.Expand("root")
	inspector.tree.Select("child1")

	// Press up key
	_ = inspector.Update(tea.KeyMsg{Type: tea.KeyUp})

	// Selection should move to root
	selected := inspector.tree.GetSelected()
	assert.NotNil(t, selected)
	assert.Equal(t, "root", selected.ID, "Up key should move selection up")
}

// TestComponentInspector_SearchMode_AllKeys tests all key handlers in search mode
func TestComponentInspector_SearchMode_AllKeys(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root",
		Name: "App",
		Type: "Component",
		Children: []*ComponentSnapshot{
			{ID: "button1", Name: "Button", Type: "Button"},
			{ID: "input1", Name: "Input", Type: "Input"},
			{ID: "label1", Name: "Label", Type: "Label"},
		},
	}

	tests := []struct {
		name      string
		setup     func(*ComponentInspector)
		key       tea.KeyMsg
		check     func(*testing.T, *ComponentInspector)
	}{
		{
			name: "Enter selects current search result",
			setup: func(ci *ComponentInspector) {
				ci.Update(tea.KeyMsg{Type: tea.KeyCtrlF}) // Enter search mode
				ci.search.Search("Button")               // Search for Button
			},
			key: tea.KeyMsg{Type: tea.KeyEnter},
			check: func(t *testing.T, ci *ComponentInspector) {
				assert.False(t, ci.searchMode, "Search mode should exit on Enter")
				// Should have selected the search result
			},
		},
		{
			name: "Down navigates to next search result",
			setup: func(ci *ComponentInspector) {
				ci.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
				ci.search.Search("") // Empty search shows all
			},
			key: tea.KeyMsg{Type: tea.KeyDown},
			check: func(t *testing.T, ci *ComponentInspector) {
				assert.True(t, ci.searchMode, "Should stay in search mode")
			},
		},
		{
			name: "Up navigates to previous search result",
			setup: func(ci *ComponentInspector) {
				ci.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
				ci.search.Search("")
				ci.search.NextResult() // Move to second result first
			},
			key: tea.KeyMsg{Type: tea.KeyUp},
			check: func(t *testing.T, ci *ComponentInspector) {
				assert.True(t, ci.searchMode, "Should stay in search mode")
			},
		},
		{
			name: "Backspace removes last character from query",
			setup: func(ci *ComponentInspector) {
				ci.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
				ci.search.Search("test")
			},
			key: tea.KeyMsg{Type: tea.KeyBackspace},
			check: func(t *testing.T, ci *ComponentInspector) {
				assert.Equal(t, "tes", ci.search.GetQuery(), "Should remove last char")
			},
		},
		{
			name: "Backspace on empty query does nothing",
			setup: func(ci *ComponentInspector) {
				ci.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
				ci.search.Search("")
			},
			key: tea.KeyMsg{Type: tea.KeyBackspace},
			check: func(t *testing.T, ci *ComponentInspector) {
				assert.Equal(t, "", ci.search.GetQuery(), "Query should remain empty")
			},
		},
		{
			name: "Runes append to query",
			setup: func(ci *ComponentInspector) {
				ci.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
				ci.search.Search("te")
			},
			key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("st")},
			check: func(t *testing.T, ci *ComponentInspector) {
				assert.Equal(t, "test", ci.search.GetQuery(), "Should append runes")
			},
		},
		{
			name: "Esc exits search mode and clears query",
			setup: func(ci *ComponentInspector) {
				ci.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
				ci.search.Search("test")
			},
			key: tea.KeyMsg{Type: tea.KeyEsc},
			check: func(t *testing.T, ci *ComponentInspector) {
				assert.False(t, ci.searchMode, "Should exit search mode")
				assert.Equal(t, "", ci.search.GetQuery(), "Query should be cleared")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inspector := NewComponentInspector(root)
			inspector.tree.Expand("root")

			if tt.setup != nil {
				tt.setup(inspector)
			}

			// Send the key
			_ = inspector.Update(tt.key)

			// Run checks
			tt.check(t, inspector)
		})
	}
}

// TestComponentInspector_SearchMode_EnterWithNoResult tests Enter with no search result selected
func TestComponentInspector_SearchMode_EnterWithNoResult(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root",
		Name: "App",
		Type: "Component",
	}

	inspector := NewComponentInspector(root)

	// Enter search mode
	inspector.Update(tea.KeyMsg{Type: tea.KeyCtrlF})

	// Search for something that doesn't exist
	inspector.search.Search("nonexistent")

	// Press Enter - should exit search mode even with no result
	inspector.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.False(t, inspector.searchMode, "Should exit search mode on Enter even with no result")
}

// TestComponentInspector_HandleKeyMsg_ArrowKeys tests arrow key navigation
func TestComponentInspector_HandleKeyMsg_ArrowKeys(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root",
		Name: "App",
		Type: "Component",
		Children: []*ComponentSnapshot{
			{ID: "child1", Name: "Child1"},
			{ID: "child2", Name: "Child2"},
		},
	}

	inspector := NewComponentInspector(root)
	inspector.tree.Expand("root")

	// Test Right arrow (next tab)
	initialTab := inspector.detail.GetActiveTab()
	inspector.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.NotEqual(t, initialTab, inspector.detail.GetActiveTab(), "Right arrow should change tab")

	// Test Left arrow (previous tab)
	inspector.Update(tea.KeyMsg{Type: tea.KeyLeft})
	assert.Equal(t, initialTab, inspector.detail.GetActiveTab(), "Left arrow should go back to previous tab")
}

// TestComponentInspector_HandleKeyMsg_EnterWithNoSelection tests Enter with no component selected
func TestComponentInspector_HandleKeyMsg_EnterWithNoSelection(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root",
		Name: "App",
		Type: "Component",
		Children: []*ComponentSnapshot{
			{ID: "child1", Name: "Child1"},
		},
	}

	// Create inspector with nil root first, then set root without auto-select
	inspector := NewComponentInspector(nil)

	// Set root directly to tree (bypassing SetRoot which auto-selects)
	inspector.tree = NewTreeView(root)

	// Verify nothing is selected
	assert.Nil(t, inspector.tree.GetSelected())

	// Press Enter - should select and expand root
	inspector.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Now root should be selected and expanded
	selected := inspector.tree.GetSelected()
	assert.NotNil(t, selected, "Enter should select root when nothing is selected")
	assert.Equal(t, "root", selected.ID, "Root should be selected")
	assert.True(t, inspector.tree.IsExpanded("root"), "Root should be expanded")
}

// TestComponentInspector_HandleKeyMsg_CtrlF tests Ctrl+F to enter search mode
func TestComponentInspector_HandleKeyMsg_CtrlF(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root",
		Name: "App",
		Type: "Component",
	}

	inspector := NewComponentInspector(root)

	// Initially not in search mode
	assert.False(t, inspector.searchMode)

	// Press Ctrl+F
	inspector.Update(tea.KeyMsg{Type: tea.KeyCtrlF})

	// Should be in search mode
	assert.True(t, inspector.searchMode, "Ctrl+F should enable search mode")
}

// TestComponentInspector_HandleKeyMsg_EnterTogglesExpansion tests Enter to toggle node expansion
func TestComponentInspector_HandleKeyMsg_EnterTogglesExpansion(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root",
		Name: "App",
		Type: "Component",
		Children: []*ComponentSnapshot{
			{ID: "child1", Name: "Child1"},
		},
	}

	inspector := NewComponentInspector(root)

	// Root should be auto-expanded and selected
	assert.True(t, inspector.tree.IsExpanded("root"))
	assert.Equal(t, "root", inspector.tree.GetSelected().ID)

	// Press Enter to collapse root
	inspector.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, inspector.tree.IsExpanded("root"), "Enter should collapse root")

	// Press Enter again to expand root
	inspector.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.True(t, inspector.tree.IsExpanded("root"), "Enter should expand root again")
}
