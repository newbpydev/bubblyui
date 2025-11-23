package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ComponentID is a unique identifier for a component instance.
type ComponentID string

// ClassName represents a CSS-like class name for styling.
type ClassName string

// CommonProps contains properties shared by all components.
// These props provide consistent styling and identification across the component library.
type CommonProps struct {
	// ID is a unique identifier for the component instance.
	// Optional - if not provided, components generate their own IDs.
	ID ComponentID

	// ClassName allows applying custom styling classes.
	// Optional - used for custom theming and styling overrides.
	ClassName ClassName

	// Style provides direct Lipgloss style overrides.
	// Optional - takes precedence over theme and className styles.
	Style *lipgloss.Style
}

// Variant represents a visual variant of a component.
// Common variants include "primary", "secondary", "danger", "success", "warning".
type Variant string

// Common variant constants used across multiple components.
const (
	// VariantPrimary represents the primary/default variant.
	VariantPrimary Variant = "primary"

	// VariantSecondary represents a secondary/alternative variant.
	VariantSecondary Variant = "secondary"

	// VariantDanger represents a destructive/dangerous action variant.
	VariantDanger Variant = "danger"

	// VariantSuccess represents a successful/positive action variant.
	VariantSuccess Variant = "success"

	// VariantWarning represents a warning/caution variant.
	VariantWarning Variant = "warning"

	// VariantInfo represents an informational variant.
	VariantInfo Variant = "info"
)

// Size represents the size of a component.
// Common sizes include "small", "medium", "large".
type Size string

// Common size constants used across multiple components.
const (
	// SizeSmall represents a small component size.
	SizeSmall Size = "small"

	// SizeMedium represents a medium/default component size.
	SizeMedium Size = "medium"

	// SizeLarge represents a large component size.
	SizeLarge Size = "large"
)

// Alignment represents text or content alignment.
type Alignment string

// Common alignment constants.
const (
	// AlignLeft aligns content to the left.
	AlignLeft Alignment = "left"

	// AlignCenter centers content.
	AlignCenter Alignment = "center"

	// AlignRight aligns content to the right.
	AlignRight Alignment = "right"
)

// Position represents the position of an element.
type Position string

// Common position constants.
const (
	// PositionTop positions element at the top.
	PositionTop Position = "top"

	// PositionBottom positions element at the bottom.
	PositionBottom Position = "bottom"

	// PositionLeft positions element on the left.
	PositionLeft Position = "left"

	// PositionRight positions element on the right.
	PositionRight Position = "right"
)

// EventHandler is a generic event handler function.
// It receives event data and performs an action.
type EventHandler func(data interface{})

// ValidateFunc is a validation function that returns an error if validation fails.
type ValidateFunc func(value string) error

// RenderFunc is a function that renders content to a string.
// Used for custom rendering in components like List and Table.
type RenderFunc func() string

// injectTheme attempts to inject a theme from context, falling back to DefaultTheme.
// This is a common helper used by all components that need theme support.
func injectTheme(ctx *bubbly.Context) Theme {
	if injected := ctx.Inject("theme", nil); injected != nil {
		if t, ok := injected.(Theme); ok {
			return t
		}
	}
	return DefaultTheme
}
