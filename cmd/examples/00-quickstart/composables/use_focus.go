// Package composables provides reusable reactive logic for the quickstart example.
package composables

import (
	// Clean import using alias package
	"github.com/newbpydev/bubblyui"
)

// FocusPane represents different UI panes that can be focused.
type FocusPane int

const (
	FocusList FocusPane = iota
	FocusInput
)

// FocusComposable encapsulates focus management logic.
type FocusComposable struct {
	Current        *bubblyui.Ref[FocusPane]
	Next           func()
	Prev           func()
	SetFocus       func(pane FocusPane)
	IsListFocused  func() bool
	IsInputFocused func() bool
}

// UseFocusManager creates a focus management composable.
// Uses type-safe bubblyui.NewRef[T]() (PREFERRED pattern).
// IMPORTANT: Use GetTyped() for type-safe access, not Get() which returns any.
func UseFocusManager(initial FocusPane) *FocusComposable {
	current := bubblyui.NewRef(initial)
	panes := []FocusPane{FocusList, FocusInput}

	next := func() {
		curr := current.GetTyped()
		for i, p := range panes {
			if p == curr {
				nextIdx := (i + 1) % len(panes)
				current.Set(panes[nextIdx])
				return
			}
		}
	}

	prev := func() {
		curr := current.GetTyped()
		for i, p := range panes {
			if p == curr {
				prevIdx := (i - 1 + len(panes)) % len(panes)
				current.Set(panes[prevIdx])
				return
			}
		}
	}

	setFocus := func(pane FocusPane) {
		current.Set(pane)
	}

	isListFocused := func() bool {
		return current.GetTyped() == FocusList
	}

	isInputFocused := func() bool {
		return current.GetTyped() == FocusInput
	}

	return &FocusComposable{
		Current:        current,
		Next:           next,
		Prev:           prev,
		SetFocus:       setFocus,
		IsListFocused:  isListFocused,
		IsInputFocused: isInputFocused,
	}
}
