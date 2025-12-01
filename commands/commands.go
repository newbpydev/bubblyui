// Package commands provides command generation, batching, and debugging for BubblyUI.
//
// This package enables the automatic reactive bridge pattern where state changes
// automatically generate Bubbletea commands for efficient UI updates.
//
// This package is an alias for github.com/newbpydev/bubblyui/pkg/bubbly/commands,
// providing a cleaner import path for users.
//
// # Core Features
//
//   - Command generation from reactive state changes
//   - Command batching with coalescing strategies
//   - Debug logging for command inspection
//   - Loop detection and prevention
//   - Command deduplication
//
// # Example
//
//	import "github.com/newbpydev/bubblyui/commands"
//
//	// Create a command batcher
//	batcher := commands.NewCommandBatcher(commands.CoalesceAll)
//	batcher.EnableDeduplication()
//
//	// Enable debug logging
//	logger := commands.NewCommandLogger(os.Stdout)
//	commands.SetDefaultLogger(logger)
//
//	// Detect command loops
//	detector := commands.NewLoopDetector()
package commands

import (
	"io"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/commands"
)

// =============================================================================
// Core Types - Re-exported from bubbly
// =============================================================================

// CommandGenerator defines the interface for generating Bubbletea commands.
type CommandGenerator = bubbly.CommandGenerator

// StateChangedMsg represents a state change notification.
type StateChangedMsg = bubbly.StateChangedMsg

// CommandQueue manages pending commands for a component.
type CommandQueue = bubbly.CommandQueue

// NewCommandQueue creates a new command queue.
var NewCommandQueue = bubbly.NewCommandQueue

// =============================================================================
// Command Generation
// =============================================================================

// DefaultCommandGenerator is the standard implementation of CommandGenerator.
type DefaultCommandGenerator = commands.DefaultCommandGenerator

// CommandRef is a reactive reference with automatic command generation.
type CommandRef[T any] = commands.CommandRef[T]

// =============================================================================
// Command Batching
// =============================================================================

// CoalescingStrategy determines how multiple commands are batched together.
type CoalescingStrategy = commands.CoalescingStrategy

// Coalescing strategy constants.
const (
	// CoalesceAll batches all commands into a single command.
	CoalesceAll = commands.CoalesceAll
	// CoalesceByType groups commands by their message type before batching.
	CoalesceByType = commands.CoalesceByType
	// NoCoalesce executes all commands individually via tea.Batch.
	NoCoalesce = commands.NoCoalesce
)

// CommandBatcher batches multiple Bubbletea commands into a single command.
type CommandBatcher = commands.CommandBatcher

// NewCommandBatcher creates a new CommandBatcher with the specified strategy.
var NewCommandBatcher = commands.NewCommandBatcher

// StateChangedBatchMsg represents a batch of state change messages.
type StateChangedBatchMsg = commands.StateChangedBatchMsg

// =============================================================================
// Debug Logging
// =============================================================================

// CommandLogger interface for logging command generation events.
type CommandLogger = commands.CommandLogger

// NewCommandLogger creates a new logger that writes to the given writer.
func NewCommandLogger(writer io.Writer) CommandLogger {
	return commands.NewCommandLogger(writer)
}

// NewNopLogger creates a no-op logger with zero overhead.
var NewNopLogger = commands.NewNopLogger

// GetDefaultLogger returns the current default logger.
var GetDefaultLogger = commands.GetDefaultLogger

// SetDefaultLogger sets the default logger for all command generation.
var SetDefaultLogger = commands.SetDefaultLogger

// FormatValue formats a value for debug logging output.
var FormatValue = commands.FormatValue

// =============================================================================
// Command Inspection
// =============================================================================

// CommandInspector provides utilities for inspecting command queues.
type CommandInspector = commands.CommandInspector

// NewCommandInspector creates a new inspector for the given queue.
var NewCommandInspector = commands.NewCommandInspector

// CommandInfo contains detailed information about a command.
type CommandInfo = commands.CommandInfo

// =============================================================================
// Loop Detection
// =============================================================================

// LoopDetector detects and prevents infinite command loops.
type LoopDetector = commands.LoopDetector

// NewLoopDetector creates a new loop detector.
var NewLoopDetector = commands.NewLoopDetector

// CommandLoopError represents an error when a command loop is detected.
type CommandLoopError = commands.CommandLoopError
