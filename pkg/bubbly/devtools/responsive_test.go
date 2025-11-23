package devtools

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestLayoutManager_CalculateResponsiveLayout tests responsive layout calculation based on terminal width.
func TestLayoutManager_CalculateResponsiveLayout(t *testing.T) {
	tests := []struct {
		name          string
		width         int
		expectedMode  LayoutMode
		expectedRatio float64
	}{
		{
			name:          "narrow terminal (<80 cols) uses vertical layout",
			width:         70,
			expectedMode:  LayoutVertical,
			expectedRatio: 0.5,
		},
		{
			name:          "exactly 80 cols uses 50/50 horizontal",
			width:         80,
			expectedMode:  LayoutHorizontal,
			expectedRatio: 0.5,
		},
		{
			name:          "medium terminal (80-120 cols) uses 50/50 horizontal",
			width:         100,
			expectedMode:  LayoutHorizontal,
			expectedRatio: 0.5,
		},
		{
			name:          "exactly 120 cols uses 50/50 horizontal",
			width:         120,
			expectedMode:  LayoutHorizontal,
			expectedRatio: 0.5,
		},
		{
			name:          "wide terminal (>120 cols) uses 40/60 horizontal",
			width:         150,
			expectedMode:  LayoutHorizontal,
			expectedRatio: 0.4,
		},
		{
			name:          "very wide terminal uses 40/60 horizontal",
			width:         200,
			expectedMode:  LayoutHorizontal,
			expectedRatio: 0.4,
		},
		{
			name:          "very narrow terminal uses vertical",
			width:         40,
			expectedMode:  LayoutVertical,
			expectedRatio: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode, ratio := CalculateResponsiveLayout(tt.width)
			assert.Equal(t, tt.expectedMode, mode, "layout mode mismatch")
			assert.Equal(t, tt.expectedRatio, ratio, "split ratio mismatch")
		})
	}
}

// TestDevToolsUI_WindowSizeMsg_UpdatesDimensions tests that WindowSizeMsg updates the UI dimensions.
func TestDevToolsUI_WindowSizeMsg_UpdatesDimensions(t *testing.T) {
	store := NewDevToolsStore(100, 100, 100)
	ui := NewDevToolsUI(store)

	// Send WindowSizeMsg
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedUI, cmd := ui.Update(msg)

	assert.Nil(t, cmd, "WindowSizeMsg should not return a command")
	assert.NotNil(t, updatedUI, "Update should return updated UI")

	// Verify size was updated
	ui = updatedUI.(*UI)
	width, height := ui.layout.GetSize()
	assert.Equal(t, 120, width, "width not updated")
	assert.Equal(t, 40, height, "height not updated")
}

// TestDevToolsUI_ResponsiveLayout_Narrow tests that narrow terminals switch to vertical layout.
func TestDevToolsUI_ResponsiveLayout_Narrow(t *testing.T) {
	store := NewDevToolsStore(100, 100, 100)
	ui := NewDevToolsUI(store)

	// Send narrow WindowSizeMsg
	msg := tea.WindowSizeMsg{Width: 70, Height: 40}
	updatedUI, _ := ui.Update(msg)
	ui = updatedUI.(*UI)

	// Verify vertical layout
	mode := ui.GetLayoutMode()
	assert.Equal(t, LayoutVertical, mode, "narrow terminal should use vertical layout")

	ratio := ui.GetLayoutRatio()
	assert.Equal(t, 0.5, ratio, "narrow terminal should use 50/50 split")
}

// TestDevToolsUI_ResponsiveLayout_Medium tests that medium terminals use 50/50 horizontal layout.
func TestDevToolsUI_ResponsiveLayout_Medium(t *testing.T) {
	store := NewDevToolsStore(100, 100, 100)
	ui := NewDevToolsUI(store)

	// Send medium WindowSizeMsg
	msg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedUI, _ := ui.Update(msg)
	ui = updatedUI.(*UI)

	// Verify horizontal layout with 50/50 split
	mode := ui.GetLayoutMode()
	assert.Equal(t, LayoutHorizontal, mode, "medium terminal should use horizontal layout")

	ratio := ui.GetLayoutRatio()
	assert.Equal(t, 0.5, ratio, "medium terminal should use 50/50 split")
}

// TestDevToolsUI_ResponsiveLayout_Wide tests that wide terminals use 40/60 horizontal layout.
func TestDevToolsUI_ResponsiveLayout_Wide(t *testing.T) {
	store := NewDevToolsStore(100, 100, 100)
	ui := NewDevToolsUI(store)

	// Send wide WindowSizeMsg
	msg := tea.WindowSizeMsg{Width: 150, Height: 40}
	updatedUI, _ := ui.Update(msg)
	ui = updatedUI.(*UI)

	// Verify horizontal layout with 40/60 split
	mode := ui.GetLayoutMode()
	assert.Equal(t, LayoutHorizontal, mode, "wide terminal should use horizontal layout")

	ratio := ui.GetLayoutRatio()
	assert.Equal(t, 0.4, ratio, "wide terminal should use 40/60 split (60% tools)")
}

// TestDevToolsUI_ManualLayoutOverride tests that manual layout mode is not overridden by resize.
func TestDevToolsUI_ManualLayoutOverride(t *testing.T) {
	store := NewDevToolsStore(100, 100, 100)
	ui := NewDevToolsUI(store)

	// Set manual layout mode
	ui.SetManualLayoutMode(LayoutOverlay)

	// Send narrow WindowSizeMsg that would normally trigger vertical
	msg := tea.WindowSizeMsg{Width: 70, Height: 40}
	updatedUI, _ := ui.Update(msg)
	ui = updatedUI.(*UI)

	// Verify layout mode was NOT changed (manual override active)
	mode := ui.GetLayoutMode()
	assert.Equal(t, LayoutOverlay, mode, "manual layout override should prevent auto-layout")
}

// TestDevToolsUI_EnableAutoLayout tests re-enabling auto layout after manual override.
func TestDevToolsUI_EnableAutoLayout(t *testing.T) {
	store := NewDevToolsStore(100, 100, 100)
	ui := NewDevToolsUI(store)

	// Set manual layout mode
	ui.SetManualLayoutMode(LayoutOverlay)
	assert.Equal(t, LayoutOverlay, ui.GetLayoutMode())

	// Re-enable auto layout
	ui.EnableAutoLayout()

	// Send narrow WindowSizeMsg
	msg := tea.WindowSizeMsg{Width: 70, Height: 40}
	updatedUI, _ := ui.Update(msg)
	ui = updatedUI.(*UI)

	// Verify auto layout is working again
	mode := ui.GetLayoutMode()
	assert.Equal(t, LayoutVertical, mode, "auto layout should work after re-enabling")
}

// TestDevToolsUI_CachedSize_PreventsSameResize tests that cached size prevents redundant reflows.
func TestDevToolsUI_CachedSize_PreventsSameResize(t *testing.T) {
	store := NewDevToolsStore(100, 100, 100)
	ui := NewDevToolsUI(store)

	// Send first WindowSizeMsg
	msg1 := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedUI, _ := ui.Update(msg1)
	ui = updatedUI.(*UI)

	// Get current mode and ratio
	initialMode := ui.GetLayoutMode()
	initialRatio := ui.GetLayoutRatio()

	// Send same WindowSizeMsg again
	msg2 := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedUI, _ = ui.Update(msg2)
	ui = updatedUI.(*UI)

	// Verify mode and ratio unchanged (cached)
	finalMode := ui.GetLayoutMode()
	finalRatio := ui.GetLayoutRatio()
	assert.Equal(t, initialMode, finalMode, "mode should not change for same size")
	assert.Equal(t, initialRatio, finalRatio, "ratio should not change for same size")
}

// TestDevToolsUI_InvalidDimensions tests that invalid dimensions are ignored.
func TestDevToolsUI_InvalidDimensions(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{
			name:   "zero width",
			width:  0,
			height: 40,
		},
		{
			name:   "zero height",
			width:  100,
			height: 0,
		},
		{
			name:   "negative width",
			width:  -10,
			height: 40,
		},
		{
			name:   "negative height",
			width:  100,
			height: -20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewDevToolsStore(100, 100, 100)
			ui := NewDevToolsUI(store)

			// Get initial size
			initialWidth, initialHeight := ui.layout.GetSize()

			// Send invalid WindowSizeMsg
			msg := tea.WindowSizeMsg{Width: tt.width, Height: tt.height}
			updatedUI, _ := ui.Update(msg)
			ui = updatedUI.(*UI)

			// Verify size unchanged
			finalWidth, finalHeight := ui.layout.GetSize()
			assert.Equal(t, initialWidth, finalWidth, "width should not change for invalid dimensions")
			assert.Equal(t, initialHeight, finalHeight, "height should not change for invalid dimensions")
		})
	}
}

// TestDevToolsUI_ResizeSequence tests multiple resize events in sequence.
func TestDevToolsUI_ResizeSequence(t *testing.T) {
	store := NewDevToolsStore(100, 100, 100)
	ui := NewDevToolsUI(store)

	// Sequence: narrow → medium → wide → narrow
	resizes := []struct {
		width        int
		expectedMode LayoutMode
	}{
		{width: 70, expectedMode: LayoutVertical},
		{width: 100, expectedMode: LayoutHorizontal},
		{width: 150, expectedMode: LayoutHorizontal},
		{width: 60, expectedMode: LayoutVertical},
	}

	for _, resize := range resizes {
		msg := tea.WindowSizeMsg{Width: resize.width, Height: 40}
		updatedUI, _ := ui.Update(msg)
		ui = updatedUI.(*UI)

		mode := ui.GetLayoutMode()
		assert.Equal(t, resize.expectedMode, mode, "mode should match for width %d", resize.width)
	}
}

// TestDevToolsUI_ConcurrentResize tests thread-safe handling of resize events.
func TestDevToolsUI_ConcurrentResize(t *testing.T) {
	store := NewDevToolsStore(100, 100, 100)
	ui := NewDevToolsUI(store)

	// Launch 100 concurrent resize operations
	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func(width int) {
			msg := tea.WindowSizeMsg{Width: width, Height: 40}
			_, _ = ui.Update(msg)
			done <- true
		}(80 + i) // Widths from 80 to 179
	}

	// Wait for all to complete
	for i := 0; i < 100; i++ {
		<-done
	}

	// Just verify UI is still functional (no panic)
	_ = ui.GetLayoutMode()
	_, _ = ui.layout.GetSize()
}
