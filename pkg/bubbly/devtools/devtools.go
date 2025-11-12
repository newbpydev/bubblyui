// Package devtools provides a comprehensive developer tools system for debugging
// and inspecting BubblyUI applications in real-time.
//
// The dev tools provide component tree visualization, reactive state inspection,
// event tracking, route debugging, performance monitoring, and command timeline
// analysis. Tools integrate seamlessly with running TUI applications through a
// split-pane interface or separate inspection window.
//
// # Basic Usage
//
//	// Enable dev tools globally
//	devtools.Enable()
//
//	// Toggle visibility (F12 shortcut typically)
//	dt := devtools.Enable()
//	dt.ToggleVisibility()
//
//	// Check if enabled
//	if devtools.IsEnabled() {
//	    // Dev tools are active
//	}
//
// # Thread Safety
//
// All functions and methods in this package are thread-safe and can be called
// concurrently from multiple goroutines.
//
// # Performance
//
// Dev tools overhead is < 5% when enabled and zero when disabled. The system
// uses lazy initialization and efficient data structures to minimize impact
// on application performance.
package devtools

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// DevTools is the main dev tools instance that manages the entire debugging system.
//
// It coordinates data collection, storage, and UI presentation for debugging
// BubblyUI applications. The instance is created as a singleton and accessed
// through package-level functions.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Lifecycle:
//
//  1. Enable() - Creates and initializes the singleton
//  2. SetVisible(true) - Shows the dev tools UI
//  3. ... debugging work ...
//  4. SetVisible(false) - Hides the UI (still collecting data)
//  5. Disable() - Stops data collection and cleanup
//
// Example:
//
//	dt := devtools.Enable()
//	dt.SetVisible(true)
//	defer dt.SetVisible(false)
type DevTools struct {
	// enabled indicates whether dev tools are active
	// Protected by mu for thread-safe access
	enabled bool

	// visible indicates whether dev tools UI is shown
	// Protected by mu for thread-safe access
	visible bool

	// collector hooks into application for data collection
	// Implemented in Task 1.2
	collector *DataCollector

	// store holds collected debug data in memory
	// Implemented in Task 1.3
	store *DevToolsStore

	// ui manages the dev tools user interface
	// Implemented in Task 5.4
	ui *DevToolsUI

	// config holds dev tools configuration
	// Implemented in Task 1.5
	config *Config

	// mu protects concurrent access to enabled and visible fields
	mu sync.RWMutex
}

// Global singleton state
var (
	// globalDevToolsMu protects access to globalDevTools and globalDevToolsOnce
	globalDevToolsMu sync.RWMutex

	// globalDevTools is the singleton instance
	globalDevTools *DevTools

	// globalDevToolsOnce ensures singleton is initialized only once
	globalDevToolsOnce sync.Once
)

// Enable creates and enables the dev tools singleton.
//
// This function is idempotent - calling it multiple times returns the same
// instance. The dev tools are initialized on first call and subsequent calls
// just return the existing instance.
//
// The returned DevTools instance is enabled and ready to use, but not visible
// by default. Call SetVisible(true) to show the UI.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	// Enable dev tools
//	dt := devtools.Enable()
//
//	// Show UI
//	dt.SetVisible(true)
//
//	// Check state
//	if dt.IsEnabled() {
//	    fmt.Println("Dev tools active")
//	}
//
// Returns:
//   - *DevTools: The singleton dev tools instance
func Enable() *DevTools {
	globalDevToolsMu.Lock()
	defer globalDevToolsMu.Unlock()

	// Use sync.Once to ensure initialization happens only once
	globalDevToolsOnce.Do(func() {
		// Initialize store
		store := NewDevToolsStore(1000, 1000, 1000)
		
		// Initialize UI
		ui := NewDevToolsUI(store)
		
		// Initialize data collector
		collector := NewDataCollector()
		
		// Initialize config
		config := DefaultConfig()
		
		globalDevTools = &DevTools{
			enabled:   true,
			visible:   false,
			ui:        ui,
			store:     store,
			collector: collector,
			config:    config,
		}

		// Register F12/ctrl+t global key interceptor for zero-config toggle
		// This allows F12 or ctrl+t to work without any user code changes
		// ctrl+t is provided as an alternative for Linux/terminals where F12 is intercepted
		bubbly.SetGlobalKeyInterceptor(func(key tea.KeyMsg) bool {
			// Intercept F12 or ctrl+t
			isToggleKey := key.Type == tea.KeyF12 || key.String() == "ctrl+t"
			
			if isToggleKey {
				// Toggle dev tools visibility
				if globalDevTools != nil && globalDevTools.IsEnabled() {
					globalDevTools.ToggleVisibility()
					return true // Key handled, don't forward to component
				}
			}
			return false // Not handled, forward to component
		})

		// Register global view renderer for zero-config UI integration
		// This allows dev tools UI to overlay on app view automatically
		bubbly.SetGlobalViewRenderer(RenderView)

		// Register global update hook for zero-config UI interaction
		// This allows dev tools UI to receive messages (e.g., Tab key for navigation)
		bubbly.SetGlobalUpdateHook(HandleUpdate)
	})

	// If already created but disabled, re-enable it
	if globalDevTools != nil && !globalDevTools.IsEnabled() {
		globalDevTools.mu.Lock()
		globalDevTools.enabled = true
		globalDevTools.mu.Unlock()
	}

	return globalDevTools
}

// Disable disables the dev tools system.
//
// This stops data collection and hides the UI. The singleton instance is
// preserved and can be re-enabled with Enable().
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	// Disable dev tools
//	devtools.Disable()
//
//	// Verify disabled
//	if !devtools.IsEnabled() {
//	    fmt.Println("Dev tools disabled")
//	}
func Disable() {
	globalDevToolsMu.RLock()
	dt := globalDevTools
	globalDevToolsMu.RUnlock()

	if dt != nil {
		dt.mu.Lock()
		dt.enabled = false
		dt.visible = false // Hide UI when disabling
		dt.mu.Unlock()
	}
}

// Toggle toggles the enabled state of dev tools.
//
// If dev tools are disabled, this enables them. If enabled, this disables them.
// This is useful for keyboard shortcuts (e.g., F12 to toggle dev tools).
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	// Toggle dev tools (F12 handler)
//	func handleF12() {
//	    devtools.Toggle()
//	}
func Toggle() {
	if IsEnabled() {
		Disable()
	} else {
		Enable()
	}
}

// IsEnabled returns whether dev tools are currently enabled.
//
// This is a package-level function that checks the global singleton state.
// Returns false if dev tools have never been enabled.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	if devtools.IsEnabled() {
//	    // Dev tools are active
//	    dt := devtools.Enable()
//	    dt.SetVisible(true)
//	}
//
// Returns:
//   - bool: true if dev tools are enabled, false otherwise
func IsEnabled() bool {
	globalDevToolsMu.RLock()
	dt := globalDevTools
	globalDevToolsMu.RUnlock()

	if dt == nil {
		return false
	}

	return dt.IsEnabled()
}

// IsEnabled returns whether this DevTools instance is enabled.
//
// This is a method on the DevTools instance. Use the package-level IsEnabled()
// function to check global state without getting the instance.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	dt := devtools.Enable()
//	if dt.IsEnabled() {
//	    fmt.Println("This instance is enabled")
//	}
//
// Returns:
//   - bool: true if this instance is enabled, false otherwise
func (dt *DevTools) IsEnabled() bool {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	return dt.enabled
}

// SetVisible sets the visibility of the dev tools UI.
//
// Setting visible to true shows the dev tools panel (split-pane or overlay).
// Setting visible to false hides the UI but continues data collection if enabled.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	dt := devtools.Enable()
//
//	// Show UI
//	dt.SetVisible(true)
//
//	// Hide UI (still collecting data)
//	dt.SetVisible(false)
//
// Parameters:
//   - visible: true to show UI, false to hide UI
func (dt *DevTools) SetVisible(visible bool) {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	dt.visible = visible
}

// IsVisible returns whether the dev tools UI is currently visible.
//
// The UI can be hidden while dev tools are still enabled and collecting data.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	dt := devtools.Enable()
//	if dt.IsVisible() {
//	    fmt.Println("UI is shown")
//	} else {
//	    fmt.Println("UI is hidden")
//	}
//
// Returns:
//   - bool: true if UI is visible, false otherwise
func (dt *DevTools) IsVisible() bool {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	return dt.visible
}

// ToggleVisibility toggles the visibility of the dev tools UI.
//
// If the UI is hidden, this shows it. If shown, this hides it.
// This is useful for keyboard shortcuts (e.g., F12 to toggle visibility).
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	dt := devtools.Enable()
//
//	// Toggle visibility (F12 handler)
//	func handleF12() {
//	    dt.ToggleVisibility()
//	}
func (dt *DevTools) ToggleVisibility() {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	dt.visible = !dt.visible
}

// RenderWithApp renders the application view with dev tools UI if visible.
//
// This is the main rendering method that should be called by the wrapper to
// combine the application view with the dev tools UI. If dev tools are not
// visible, it just returns the app view unchanged.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	dt := devtools.Enable()
//	appView := component.View()
//	finalView := dt.RenderWithApp(appView)
//	// finalView contains app + dev tools if visible, or just app if hidden
//
// Parameters:
//   - appView: The application's rendered view
//
// Returns:
//   - string: Combined view of app + dev tools, or just app if dev tools hidden
func (dt *DevTools) RenderWithApp(appView string) string {
	dt.mu.RLock()
	defer dt.mu.RUnlock()

	// If dev tools not visible, just return app view
	if !dt.visible {
		return appView
	}

	// Set the app content in the UI
	dt.ui.SetAppContent(appView)

	// Render combined view (app + dev tools)
	return dt.ui.View()
}

// RenderView is a package-level function to render the app view with dev tools.
//
// This is called automatically by bubbly.Wrap() to integrate dev tools rendering.
// If dev tools are not enabled or not visible, it just returns the app view.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	// In wrapper.View()
//	appView := m.component.View()
//	return devtools.RenderView(appView)
//
// Parameters:
//   - appView: The application's rendered view
//
// Returns:
//   - string: Combined view of app + dev tools, or just app if dev tools disabled/hidden
func RenderView(appView string) string {
	globalDevToolsMu.RLock()
	dt := globalDevTools
	globalDevToolsMu.RUnlock()

	// If dev tools not enabled, just return app view
	if dt == nil || !dt.IsEnabled() {
		return appView
	}

	// Render with dev tools
	return dt.RenderWithApp(appView)
}

// HandleUpdate is a package-level function to handle Bubbletea messages for dev tools.
//
// This is called automatically by bubbly.Wrap() to enable dev tools UI interaction.
// If dev tools are not enabled or not visible, it returns nil.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	// In wrapper.Update()
//	cmd := devtools.HandleUpdate(msg)
//
// Parameters:
//   - msg: The Bubbletea message
//
// Returns:
//   - tea.Cmd: Command from dev tools UI, or nil
func HandleUpdate(msg tea.Msg) tea.Cmd {
	globalDevToolsMu.RLock()
	dt := globalDevTools
	globalDevToolsMu.RUnlock()

	// If dev tools not enabled or not visible, no-op
	if dt == nil || !dt.IsEnabled() || !dt.IsVisible() {
		return nil
	}

	// Forward message to DevToolsUI
	_, cmd := dt.ui.Update(msg)
	return cmd
}
