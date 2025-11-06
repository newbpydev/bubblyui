// Package commands provides command generator implementations for the automatic reactive bridge.
//
// This package contains implementations of the CommandGenerator interface defined in the
// bubbly package. The core types (CommandGenerator, StateChangedMsg, CommandQueue) are
// defined in the bubbly package to avoid import cycles.
//
// This package provides:
//   - DefaultCommandGenerator: Standard implementation for most use cases
//
// Example:
//
//	import "github.com/newbpydev/bubblyui/pkg/bubbly"
//	import "github.com/newbpydev/bubblyui/pkg/bubbly/commands"
//
//	// Create a command generator
//	gen := &commands.DefaultCommandGenerator{}
//
//	// Generate command from state change
//	cmd := gen.Generate("counter-1", "count", 0, 1)
//
//	// Command returns bubbly.StateChangedMsg when executed
//	msg := cmd() // bubbly.StateChangedMsg{ComponentID: "counter-1", RefID: "count", ...}
package commands

import (
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CommandGenerator is re-exported from bubbly package for convenience.
// See bubbly.CommandGenerator for full documentation.
type CommandGenerator = bubbly.CommandGenerator

// StateChangedMsg is re-exported from bubbly package for convenience.
// See bubbly.StateChangedMsg for full documentation.
type StateChangedMsg = bubbly.StateChangedMsg

// CommandQueue is re-exported from bubbly package for convenience.
// See bubbly.CommandQueue for full documentation.
type CommandQueue = bubbly.CommandQueue

// NewCommandQueue is re-exported from bubbly package for convenience.
// See bubbly.NewCommandQueue for full documentation.
var NewCommandQueue = bubbly.NewCommandQueue
