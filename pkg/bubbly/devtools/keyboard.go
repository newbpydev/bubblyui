package devtools

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// FocusTarget represents which part of the dev tools has focus.
//
// Focus determines which keyboard shortcuts are active and where
// keyboard input is directed.
type FocusTarget int

const (
	// FocusApp indicates the main application has focus.
	FocusApp FocusTarget = iota
	// FocusTools indicates the dev tools panel has focus.
	FocusTools
	// FocusInspector indicates the component inspector has focus.
	FocusInspector
	// FocusState indicates the state viewer has focus.
	FocusState
	// FocusEvents indicates the event tracker has focus.
	FocusEvents
	// FocusPerformance indicates the performance monitor has focus.
	FocusPerformance
)

// KeyHandler is a function that handles a keyboard message and optionally
// returns a Bubbletea command.
//
// Handlers can:
// - Process the key message
// - Update application state
// - Return commands for async operations
// - Return nil if no command needed
type KeyHandler func(tea.KeyMsg) tea.Cmd

// shortcutEntry represents a registered keyboard shortcut.
type shortcutEntry struct {
	handler KeyHandler
	focus   FocusTarget
	global  bool // If true, works regardless of focus
}

// KeyboardHandler manages keyboard shortcuts and focus for dev tools.
//
// The keyboard handler supports:
// - Global shortcuts (work with any focus)
// - Focus-specific shortcuts (only work when specific panel has focus)
// - Dynamic shortcut registration/unregistration
// - Thread-safe concurrent access
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	kh := devtools.NewKeyboardHandler()
//
//	// Register global F12 toggle
//	kh.RegisterGlobal("f12", func(msg tea.KeyMsg) tea.Cmd {
//	    devtools.Toggle()
//	    return nil
//	})
//
//	// Register inspector-specific shortcut
//	kh.RegisterWithFocus("ctrl+f", FocusInspector, func(msg tea.KeyMsg) tea.Cmd {
//	    // Open search in inspector
//	    return nil
//	})
//
//	// Handle keyboard message
//	cmd := kh.Handle(keyMsg)
type KeyboardHandler struct {
	mu        sync.RWMutex
	shortcuts map[string][]shortcutEntry
	focus     FocusTarget
}

// NewKeyboardHandler creates a new keyboard handler with default focus on the app.
func NewKeyboardHandler() *KeyboardHandler {
	return &KeyboardHandler{
		shortcuts: make(map[string][]shortcutEntry),
		focus:     FocusApp,
	}
}

// Register registers a keyboard shortcut handler.
//
// The handler will be called when the specified key is pressed,
// regardless of current focus (global shortcut).
//
// If key is empty or handler is nil, the registration is ignored.
func (kh *KeyboardHandler) Register(key string, handler KeyHandler) {
	kh.RegisterGlobal(key, handler)
}

// RegisterGlobal registers a global keyboard shortcut that works regardless of focus.
//
// Global shortcuts are useful for:
// - Toggle dev tools visibility (F12)
// - Quit application (Ctrl+C)
// - Help dialog (?)
//
// If key is empty or handler is nil, the registration is ignored.
func (kh *KeyboardHandler) RegisterGlobal(key string, handler KeyHandler) {
	if key == "" || handler == nil {
		return
	}

	kh.mu.Lock()
	defer kh.mu.Unlock()

	entry := shortcutEntry{
		handler: handler,
		global:  true,
	}

	kh.shortcuts[key] = append(kh.shortcuts[key], entry)
}

// RegisterWithFocus registers a keyboard shortcut that only works when
// the specified focus target is active.
//
// Focus-specific shortcuts are useful for:
// - Panel-specific navigation
// - Context-sensitive actions
// - Avoiding key conflicts between panels
//
// If key is empty or handler is nil, the registration is ignored.
func (kh *KeyboardHandler) RegisterWithFocus(key string, focus FocusTarget, handler KeyHandler) {
	if key == "" || handler == nil {
		return
	}

	kh.mu.Lock()
	defer kh.mu.Unlock()

	entry := shortcutEntry{
		handler: handler,
		focus:   focus,
		global:  false,
	}

	kh.shortcuts[key] = append(kh.shortcuts[key], entry)
}

// Unregister removes all handlers for the specified key.
//
// This is useful for:
// - Disabling shortcuts temporarily
// - Changing shortcut behavior dynamically
// - Cleaning up on panel close
func (kh *KeyboardHandler) Unregister(key string) {
	kh.mu.Lock()
	defer kh.mu.Unlock()

	delete(kh.shortcuts, key)
}

// Handle processes a keyboard message and calls the appropriate handler.
//
// The handler is selected based on:
// 1. Global handlers are always checked first
// 2. Focus-specific handlers are checked if focus matches
// 3. First matching handler is called
//
// Returns the command from the handler, or nil if no handler matched
// or the handler returned nil.
func (kh *KeyboardHandler) Handle(msg tea.KeyMsg) tea.Cmd {
	kh.mu.RLock()
	defer kh.mu.RUnlock()

	// Get key string for lookup
	keyStr := msg.String()

	// Look up handlers for this key
	entries, ok := kh.shortcuts[keyStr]
	if !ok {
		return nil
	}

	// Find matching handler
	for _, entry := range entries {
		// Global handlers always match
		if entry.global {
			return entry.handler(msg)
		}

		// Focus-specific handlers only match if focus matches
		if entry.focus == kh.focus {
			return entry.handler(msg)
		}
	}

	return nil
}

// SetFocus changes the current focus target.
//
// Changing focus affects which keyboard shortcuts are active.
// Only focus-specific shortcuts for the new focus will respond to keys.
// Global shortcuts continue to work regardless of focus.
func (kh *KeyboardHandler) SetFocus(focus FocusTarget) {
	kh.mu.Lock()
	defer kh.mu.Unlock()

	kh.focus = focus
}

// GetFocus returns the current focus target.
func (kh *KeyboardHandler) GetFocus() FocusTarget {
	kh.mu.RLock()
	defer kh.mu.RUnlock()

	return kh.focus
}
