package devtools

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DevToolsUI is the main UI component that integrates all dev tools panels.
//
// It manages:
// - Layout (split-pane, overlay, etc.)
// - Tab navigation between panels
// - Keyboard shortcuts
// - Panel rendering (inspector, state, events, performance, timeline)
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	store := devtools.NewDevToolsStore(1000, 1000, 1000)
//	ui := devtools.NewDevToolsUI(store)
//	ui.SetAppContent("My Application")
//
//	// In Bubbletea Update()
//	updatedUI, cmd := ui.Update(msg)
//
//	// In Bubbletea View()
//	output := ui.View()
type DevToolsUI struct {
	mu sync.RWMutex

	// layout manages the split-pane layout
	layout *LayoutManager

	// tabs manages tab navigation between panels
	tabs *TabController

	// keyboard handles keyboard shortcuts
	keyboard *KeyboardHandler

	// inspector is the component inspector panel
	inspector *ComponentInspector

	// state is the state viewer panel
	state *StateViewer

	// events is the event tracker panel
	events *EventTracker

	// perf is the performance monitor panel
	perf *PerformanceMonitor

	// timeline is the command timeline panel
	timeline *CommandTimeline

	// router is the router debugger panel (optional, may be nil)
	router *RouterDebugger

	// activePanel is the index of the currently active panel
	activePanel int

	// appContent is the application content to display
	appContent string

	// store is the dev tools data store
	store *DevToolsStore

	// lastWidth is the cached terminal width to detect size changes
	lastWidth int

	// lastHeight is the cached terminal height to detect size changes
	lastHeight int

	// manualLayoutOverride indicates if user manually set layout mode
	// When true, automatic responsive layout adjustments are disabled
	manualLayoutOverride bool

	// focusMode indicates if DevTools has keyboard focus
	// When true, arrow keys and other inputs are routed to DevTools
	// When false, inputs go to the application
	focusMode bool
}

// NewDevToolsUI creates a new DevTools UI with all panels initialized.
//
// The UI starts with the component inspector panel active and a horizontal
// split layout with a 60/40 ratio (app 60%, tools 40%).
//
// Example:
//
//	store := devtools.NewDevToolsStore(1000, 1000, 1000)
//	ui := devtools.NewDevToolsUI(store)
//
// Parameters:
//   - store: The dev tools data store
//
// Returns:
//   - *DevToolsUI: A new DevTools UI instance
func NewDevToolsUI(store *DevToolsStore) *DevToolsUI {
	ui := &DevToolsUI{
		store:       store,
		activePanel: 0,
		appContent:  "",
	}

	// Initialize layout manager with horizontal split (60/40)
	ui.layout = NewLayoutManager(LayoutHorizontal, 0.6)
	ui.layout.SetSize(120, 40) // Default terminal size

	// Initialize keyboard handler
	ui.keyboard = NewKeyboardHandler()

	// Initialize all panels
	ui.inspector = NewComponentInspector(nil) // Will be updated with actual root
	ui.state = NewStateViewer(store)
	ui.events = NewEventTracker(1000)
	ui.perf = NewPerformanceMonitor(store.performance)
	ui.timeline = NewCommandTimeline(1000)
	// router is optional, will be nil for now

	// Create tabs for each panel
	tabs := []TabItem{
		{
			Name:    "Inspector",
			Content: func() string { return ui.inspector.View() },
		},
		{
			Name:    "State",
			Content: func() string { return ui.state.Render() },
		},
		{
			Name:    "Events",
			Content: func() string { return ui.events.Render() },
		},
		{
			Name:    "Performance",
			Content: func() string { return ui.perf.Render(SortByAvgTime) },
		},
		{
			Name:    "Timeline",
			Content: func() string { return ui.timeline.Render(80) },
		},
	}
	ui.tabs = NewTabController(tabs)

	// Register keyboard shortcuts
	ui.setupKeyboardShortcuts()

	return ui
}

// Init initializes the DevTools UI.
//
// This method is required by the tea.Model interface but currently
// does nothing as all initialization is done in NewDevToolsUI.
//
// Returns:
//   - tea.Cmd: Always returns nil
func (ui *DevToolsUI) Init() tea.Cmd {
	return nil
}

// setupKeyboardShortcuts registers keyboard shortcuts for the UI.
func (ui *DevToolsUI) setupKeyboardShortcuts() {
	// F12/ctrl+t are handled by globalKeyInterceptor in wrapper.go
	// No need to register here to avoid duplicate handling

	// Tab: Switch to next panel
	ui.keyboard.RegisterGlobal("tab", func(msg tea.KeyMsg) tea.Cmd {
		ui.mu.Lock()
		defer ui.mu.Unlock()

		ui.tabs.Next()
		ui.activePanel = ui.tabs.GetActiveTab()
		return nil
	})

	// Shift+Tab: Switch to previous panel
	ui.keyboard.RegisterGlobal("shift+tab", func(msg tea.KeyMsg) tea.Cmd {
		ui.mu.Lock()
		defer ui.mu.Unlock()

		ui.tabs.Prev()
		ui.activePanel = ui.tabs.GetActiveTab()
		return nil
	})
}

// IsFocusMode returns whether DevTools is in focus mode.
//
// When in focus mode, keyboard input is routed to DevTools for navigation.
// When not in focus mode, keyboard input goes to the application.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (ui *DevToolsUI) IsFocusMode() bool {
	ui.mu.RLock()
	defer ui.mu.RUnlock()
	return ui.focusMode
}

// SetFocusMode sets the focus mode state.
//
// This is useful for programmatically entering/exiting focus mode.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (ui *DevToolsUI) SetFocusMode(enabled bool) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.focusMode = enabled
}

// Update processes Bubbletea messages and updates the UI state.
//
// It routes keyboard messages to the keyboard handler first for global shortcuts,
// then to the active panel's Update() method if not handled.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	    updatedUI, cmd := m.ui.Update(msg)
//	    m.ui = updatedUI.(*DevToolsUI)
//	    return m, cmd
//	}
//
// Parameters:
//   - msg: The Bubbletea message to process
//
// Returns:
//   - tea.Model: The updated UI (cast to *DevToolsUI)
//   - tea.Cmd: Optional command to execute
func (ui *DevToolsUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle focus mode toggle keys FIRST (before any other processing)
		switch msg.Type {
		case tea.KeyRunes:
			// '/' enters focus mode
			if string(msg.Runes) == "/" {
				ui.mu.Lock()
				ui.focusMode = true
				ui.mu.Unlock()
				return ui, nil
			}
		case tea.KeyEsc:
			// ESC exits focus mode
			ui.mu.Lock()
			wasFocused := ui.focusMode
			ui.focusMode = false
			ui.mu.Unlock()
			
			// If we were in focus mode, ESC consumed (exit focus only)
			// If we weren't in focus mode, let it pass through
			if wasFocused {
				return ui, nil
			}
		}

		// Try keyboard handler for global shortcuts (Tab, Shift+Tab)
		cmd := ui.keyboard.Handle(msg)
		if cmd != nil {
			return ui, cmd
		}

		// Route to active panel's Update() ONLY if in focus mode
		ui.mu.RLock()
		activePanel := ui.activePanel
		inFocusMode := ui.focusMode
		ui.mu.RUnlock()

		// Only route keyboard input to panels when in focus mode
		if !inFocusMode {
			return ui, nil
		}

		switch activePanel {
		case 0: // Inspector
			ui.mu.Lock()
			cmd := ui.inspector.Update(msg)  // CRITICAL: Capture return value!
			ui.mu.Unlock()
			return ui, cmd  // Return cmd so updates trigger redraws
		case 1: // State viewer doesn't have Update()
			return ui, nil
		case 2: // Event tracker doesn't have Update()
			return ui, nil
		case 3: // Performance monitor doesn't have Update()
			return ui, nil
		case 4: // Timeline doesn't have Update()
			return ui, nil
		}

	case tea.WindowSizeMsg:
		ui.mu.Lock()
		defer ui.mu.Unlock()

		// Validate dimensions (ignore invalid sizes)
		if msg.Width <= 0 || msg.Height <= 0 {
			return ui, nil
		}

		// Check if size actually changed (use cache to avoid redundant updates)
		if msg.Width == ui.lastWidth && msg.Height == ui.lastHeight {
			return ui, nil
		}

		// Update cached size
		ui.lastWidth = msg.Width
		ui.lastHeight = msg.Height

		// Update layout manager size
		ui.layout.SetSize(msg.Width, msg.Height)

		// Apply responsive layout if auto mode is enabled
		if !ui.manualLayoutOverride {
			mode, ratio := CalculateResponsiveLayout(msg.Width)
			ui.layout.SetMode(mode)
			ui.layout.SetRatio(ratio)
		}

		return ui, nil
	}

	return ui, nil
}

// View renders the DevTools UI.
//
// It combines the application content with the active panel's content
// using the configured layout manager.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	func (m model) View() string {
//	    return m.ui.View()
//	}
//
// Returns:
//   - string: The rendered UI
func (ui *DevToolsUI) View() string {
	ui.mu.RLock()
	defer ui.mu.RUnlock()

	// Update inspector with latest component data from store
	ui.updateInspectorFromStore()

	// Render tabs + panel content
	tabsContent := ui.tabs.Render()

	// Add focus mode indicator if in focus mode
	var toolsContent string
	if ui.focusMode {
		// Show focus mode badge with keyboard shortcuts
		focusBadge := ui.renderFocusBadge()
		toolsContent = focusBadge + "\n" + tabsContent
	} else {
		// Show normal mode help text
		helpText := ui.renderNormalModeHelp()
		toolsContent = helpText + "\n" + tabsContent
	}

	// Combine app content and tools content using layout manager
	return ui.layout.Render(ui.appContent, toolsContent)
}

// updateInspectorFromStore updates the inspector with component data from the store.
// Must be called with read lock held.
func (ui *DevToolsUI) updateInspectorFromStore() {
	// Get root components (those without parents) from store
	// The store properly tracks component hierarchy via AddComponentChild
	roots := ui.store.GetRootComponents()
	
	if len(roots) > 0 {
		// Use first root as display root
		// Root components have their Children field populated by the store
		ui.inspector.SetRoot(roots[0])
	}
}

// renderFocusBadge renders the focus mode indicator badge.
// Must be called with read lock held.
func (ui *DevToolsUI) renderFocusBadge() string {
	badge := "🔧 DEVTOOLS FOCUS MODE  " +
		"↑/↓: Navigate • Enter: Expand • →/←: Tabs • ESC: Exit"
	
	// Style with green background to indicate active focus
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).    // Black text
		Background(lipgloss.Color("35")).   // Green background
		Bold(true).
		Padding(0, 1)
	
	return style.Render(badge)
}

// renderNormalModeHelp renders help text for normal mode.
// Must be called with read lock held.
func (ui *DevToolsUI) renderNormalModeHelp() string {
	helpText := "Press '/' to enter DevTools focus mode • Tab: Switch Tabs • F12: Toggle"
	
	// Style with subtle grey
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).  // Grey text
		Italic(true)
	
	return style.Render(helpText)
}

// getActivePanelContent returns the content of the currently active panel.
//
// This is an internal helper method called by View().
// Caller must hold read lock.
func (ui *DevToolsUI) getActivePanelContent() string {
	switch ui.activePanel {
	case 0: // Inspector
		return ui.inspector.View()
	case 1: // State
		return ui.state.Render()
	case 2: // Events
		return ui.events.Render()
	case 3: // Performance
		return ui.perf.Render(SortByAvgTime)
	case 4: // Timeline
		return ui.timeline.Render(80)
	default:
		return "Invalid panel"
	}
}

// SetAppContent sets the application content to display in the layout.
//
// The app content is shown alongside the dev tools panels according
// to the configured layout mode.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	ui.SetAppContent("My Application\nCounter: 42")
//
// Parameters:
//   - content: The application content to display
func (ui *DevToolsUI) SetAppContent(content string) {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	ui.appContent = content
}

// GetActivePanel returns the index of the currently active panel.
//
// Panel indices:
//   - 0: Component Inspector
//   - 1: State Viewer
//   - 2: Event Tracker
//   - 3: Performance Monitor
//   - 4: Command Timeline
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - int: The active panel index
func (ui *DevToolsUI) GetActivePanel() int {
	ui.mu.RLock()
	defer ui.mu.RUnlock()

	return ui.activePanel
}

// SetActivePanel sets the active panel by index.
//
// If the index is out of bounds, the active panel is not changed.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	ui.SetActivePanel(1) // Switch to State Viewer
//
// Parameters:
//   - index: The panel index (0-4)
func (ui *DevToolsUI) SetActivePanel(index int) {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	// Validate index
	if index < 0 || index >= 5 {
		return
	}

	ui.activePanel = index
	ui.tabs.Select(index)
}

// SetLayoutMode sets the layout mode for the UI.
//
// Available modes:
//   - LayoutHorizontal: Side-by-side split
//   - LayoutVertical: Top/bottom split
//   - LayoutOverlay: Tools overlay on app
//   - LayoutHidden: Tools hidden
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	ui.SetLayoutMode(LayoutVertical)
//
// Parameters:
//   - mode: The layout mode to set
func (ui *DevToolsUI) SetLayoutMode(mode LayoutMode) {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	ui.layout.SetMode(mode)
}

// GetLayoutMode returns the current layout mode.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - LayoutMode: The current layout mode
func (ui *DevToolsUI) GetLayoutMode() LayoutMode {
	ui.mu.RLock()
	defer ui.mu.RUnlock()

	return ui.layout.GetMode()
}

// SetLayoutRatio sets the split ratio for the layout.
//
// The ratio determines how much space the app gets vs the tools.
// Valid range is 0.0-1.0, where 0.6 means 60% app, 40% tools.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	ui.SetLayoutRatio(0.7) // 70% app, 30% tools
//
// Parameters:
//   - ratio: The split ratio (0.0-1.0)
func (ui *DevToolsUI) SetLayoutRatio(ratio float64) {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	ui.layout.SetRatio(ratio)
}

// GetLayoutRatio returns the current split ratio.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - float64: The current split ratio (0.0-1.0)
func (ui *DevToolsUI) GetLayoutRatio() float64 {
	ui.mu.RLock()
	defer ui.mu.RUnlock()

	return ui.layout.GetRatio()
}

// SetManualLayoutMode sets the layout mode manually and disables automatic responsive layout.
//
// When you manually set a layout mode, the UI will no longer automatically adjust
// the layout based on terminal size. Call EnableAutoLayout() to re-enable automatic
// responsive behavior.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	ui.SetManualLayoutMode(LayoutOverlay) // Force overlay mode
//	// Terminal resizes will NOT change layout mode
//
// Parameters:
//   - mode: The layout mode to set manually
func (ui *DevToolsUI) SetManualLayoutMode(mode LayoutMode) {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	ui.layout.SetMode(mode)
	ui.manualLayoutOverride = true
}

// EnableAutoLayout re-enables automatic responsive layout adjustments.
//
// After calling this, the UI will automatically adjust layout mode and ratio
// based on terminal size according to the responsive breakpoints:
//   - < 80 cols: Vertical layout
//   - 80-120 cols: Horizontal 50/50
//   - > 120 cols: Horizontal 40/60
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	ui.SetManualLayoutMode(LayoutOverlay) // Disable auto layout
//	// ... later ...
//	ui.EnableAutoLayout() // Re-enable auto layout
func (ui *DevToolsUI) EnableAutoLayout() {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	ui.manualLayoutOverride = false
}
