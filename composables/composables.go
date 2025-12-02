// Package composables provides Vue-inspired composable functions for BubblyUI.
//
// Composables are reusable functions that encapsulate stateful logic using
// the Composition API pattern. They provide a clean way to extract and
// share reactive state management across components.
//
// This package is an alias for github.com/newbpydev/bubblyui/pkg/bubbly/composables,
// providing a cleaner import path for users.
//
// # Core Composables
//
//   - UseState: Simple reactive state with getter/setter
//   - UseEffect: Side effects with automatic cleanup
//   - UseDebounce: Debounced reactive values
//   - UseThrottle: Throttled callback execution
//   - UseEventListener: Event subscription management
//
// # Data Composables
//
//   - UseList: Reactive list operations (add, remove, filter)
//   - UseHistory: Undo/redo state management
//   - UseForm: Form state with validation
//   - UseLocalStorage: Persistent storage integration
//
// # UI Composables
//
//   - UseFocus: Focus management with tab order
//   - UseWindowSize: Responsive breakpoint tracking
//   - UseCounter: Numeric counter with min/max/step
//   - UseInterval: Periodic callback execution
//
// # Example
//
//	import "github.com/newbpydev/bubblyui/composables"
//
//	func setup(ctx *bubbly.Context) {
//	    // State management
//	    count := composables.UseState(ctx, 0)
//	    count.Set(count.Get() + 1)
//
//	    // Debounced search
//	    query := bubbly.NewRef("")
//	    debouncedQuery := composables.UseDebounce(ctx, query, 300*time.Millisecond)
//
//	    // Focus management
//	    focus := composables.UseFocus(ctx, "input1", []string{"input1", "input2", "submit"})
//	}
package composables

import (
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// =============================================================================
// Core Composables - Generic functions must be wrapped
// =============================================================================

// UseState creates a simple reactive state with getter and setter methods.
func UseState[T any](ctx *bubbly.Context, initial T) UseStateReturn[T] {
	return composables.UseState(ctx, initial)
}

// UseEffect registers a side effect that runs when dependencies change.
var UseEffect = composables.UseEffect

// UseDebounce creates a debounced version of a reactive value.
func UseDebounce[T any](ctx *bubbly.Context, value *bubbly.Ref[T], delay time.Duration) *bubbly.Ref[T] {
	return composables.UseDebounce(ctx, value, delay)
}

// UseThrottle creates a throttled version of a callback function.
var UseThrottle = composables.UseThrottle

// UseEventListener registers an event listener with automatic cleanup.
var UseEventListener = composables.UseEventListener

// =============================================================================
// Data Composables
// =============================================================================

// UseList provides reactive list operations.
func UseList[T any](ctx *bubbly.Context, initial []T) *ListReturn[T] {
	return composables.UseList(ctx, initial)
}

// UseHistory provides undo/redo state management.
func UseHistory[T any](ctx *bubbly.Context, initial T, maxSize int) *HistoryReturn[T] {
	return composables.UseHistory(ctx, initial, maxSize)
}

// UseForm provides form state management with validation.
func UseForm[T any](ctx *bubbly.Context, initial T, validate func(T) map[string]string) UseFormReturn[T] {
	return composables.UseForm(ctx, initial, validate)
}

// UseLocalStorage provides persistent storage integration.
func UseLocalStorage[T any](ctx *bubbly.Context, key string, initial T, storage Storage) UseStateReturn[T] {
	return composables.UseLocalStorage(ctx, key, initial, storage)
}

// =============================================================================
// UI Composables
// =============================================================================

// UseFocus provides focus management with tab order.
func UseFocus[T comparable](ctx *bubbly.Context, initial T, order []T) *FocusReturn[T] {
	return composables.UseFocus(ctx, initial, order)
}

// UseWindowSize provides responsive breakpoint tracking.
var UseWindowSize = composables.UseWindowSize

// UseCounter provides a numeric counter with min/max/step options.
var UseCounter = composables.UseCounter

// UseInterval provides periodic callback execution.
var UseInterval = composables.UseInterval

// UseLogger provides component logging utilities.
var UseLogger = composables.UseLogger

// UseAsync provides async operation handling.
func UseAsync[T any](ctx *bubbly.Context, fetcher func() (*T, error)) UseAsyncReturn[T] {
	return composables.UseAsync(ctx, fetcher)
}

// =============================================================================
// Utility Functions
// =============================================================================

// CreateShared creates a shared composable instance across components.
func CreateShared[T any](factory func(*bubbly.Context) T) func(*bubbly.Context) T {
	return composables.CreateShared(factory)
}

// BlinkCmd returns a Bubbletea command for cursor blinking.
var BlinkCmd = composables.BlinkCmd

// =============================================================================
// Types - Re-exported for convenience
// =============================================================================

// UseStateReturn is the return type for UseState.
type UseStateReturn[T any] = composables.UseStateReturn[T]

// UseEffectCleanup is the cleanup function type for UseEffect.
type UseEffectCleanup = composables.UseEffectCleanup

// FocusReturn is the return type for UseFocus.
type FocusReturn[T comparable] = composables.FocusReturn[T]

// HistoryReturn is the return type for UseHistory.
type HistoryReturn[T any] = composables.HistoryReturn[T]

// ListReturn is the return type for UseList.
type ListReturn[T any] = composables.ListReturn[T]

// CounterReturn is the return type for UseCounter.
type CounterReturn = composables.CounterReturn

// IntervalReturn is the return type for UseInterval.
type IntervalReturn = composables.IntervalReturn

// LoggerReturn is the return type for UseLogger.
type LoggerReturn = composables.LoggerReturn

// UseAsyncReturn is the return type for UseAsync.
type UseAsyncReturn[T any] = composables.UseAsyncReturn[T]

// UseFormReturn is the return type for UseForm.
type UseFormReturn[T any] = composables.UseFormReturn[T]

// WindowSizeReturn is the return type for UseWindowSize.
type WindowSizeReturn = composables.WindowSizeReturn

// =============================================================================
// Counter Options
// =============================================================================

// CounterOption configures UseCounter behavior.
type CounterOption = composables.CounterOption

// WithMin sets the minimum value for the counter.
var WithMin = composables.WithMin

// WithMax sets the maximum value for the counter.
var WithMax = composables.WithMax

// WithStep sets the increment/decrement step for the counter.
var WithStep = composables.WithStep

// =============================================================================
// Breakpoint Types
// =============================================================================

// Breakpoint represents a responsive breakpoint.
type Breakpoint = composables.Breakpoint

// Breakpoint constants for responsive design.
const (
	BreakpointXS = composables.BreakpointXS
	BreakpointSM = composables.BreakpointSM
	BreakpointMD = composables.BreakpointMD
	BreakpointLG = composables.BreakpointLG
	BreakpointXL = composables.BreakpointXL
)

// BreakpointConfig configures responsive breakpoints.
type BreakpointConfig = composables.BreakpointConfig

// =============================================================================
// Storage Types
// =============================================================================

// Storage defines the interface for persistent storage.
type Storage = composables.Storage

// FileStorage provides file-based storage implementation.
type FileStorage = composables.FileStorage

// NewFileStorage creates a new file-based storage.
var NewFileStorage = composables.NewFileStorage

// =============================================================================
// Logging Types
// =============================================================================

// LogLevel represents logging severity levels.
type LogLevel = composables.LogLevel

// LogEntry represents a single log entry.
type LogEntry = composables.LogEntry

// Log level constants.
const (
	LogLevelDebug = composables.LogLevelDebug
	LogLevelInfo  = composables.LogLevelInfo
	LogLevelWarn  = composables.LogLevelWarn
	LogLevelError = composables.LogLevelError
)

// =============================================================================
// Errors
// =============================================================================

// ErrComposableOutsideSetup is returned when a composable is used outside Setup.
var ErrComposableOutsideSetup = composables.ErrComposableOutsideSetup
