// Package composables provides reusable reactive logic for the CPU profiler example.
package composables

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/profiler"
)

// CPUProfilerState represents the current state of the CPU profiler.
type CPUProfilerState string

const (
	// StateIdle indicates no profiling is active.
	StateIdle CPUProfilerState = "idle"

	// StateProfiling indicates CPU profiling is in progress.
	StateProfiling CPUProfilerState = "profiling"

	// StateComplete indicates profiling has finished and results are available.
	StateComplete CPUProfilerState = "complete"
)

// String returns a human-readable string for the state.
func (s CPUProfilerState) String() string {
	switch s {
	case StateIdle:
		return "Idle"
	case StateProfiling:
		return "Profiling"
	case StateComplete:
		return "Complete"
	default:
		return "Unknown"
	}
}

// HotFunctionInfo represents a function that consumes significant CPU time.
type HotFunctionInfo struct {
	// Name is the function name
	Name string

	// Samples is the number of CPU samples in this function
	Samples int64

	// Percent is the percentage of total CPU time
	Percent float64
}

// CPUProfilerComposable encapsulates CPU profiler logic with reactive state.
// This follows the Vue-like composable pattern for reusable logic.
type CPUProfilerComposable struct {
	// Profiler is the underlying CPU profiler instance
	Profiler *profiler.CPUProfiler

	// State is the current profiler state
	State *bubbly.Ref[CPUProfilerState]

	// Filename is the current/last profile filename
	Filename *bubbly.Ref[string]

	// StartTime is when profiling started
	StartTime *bubbly.Ref[time.Time]

	// FileSize is the size of the profile file (after completion)
	FileSize *bubbly.Ref[int64]

	// HotFunctions contains the analyzed hot functions
	HotFunctions *bubbly.Ref[[]HotFunctionInfo]

	// LastError holds the last error message (empty if none)
	LastError *bubbly.Ref[string]

	// Start begins CPU profiling to the specified file
	Start func(filename string) error

	// Stop ends CPU profiling
	Stop func() error

	// Analyze parses the profile and extracts hot functions
	Analyze func() error

	// Reset clears all state and returns to idle
	Reset func()

	// mu protects concurrent access
	mu sync.Mutex
}

// UseCPUProfiler creates a reusable CPU profiler composable with reactive state.
// This demonstrates the composable pattern - reusable logic that can be shared
// across components, similar to Vue's Composition API.
//
// Example:
//
//	cpuProfiler := composables.UseCPUProfiler(ctx)
//	cpuProfiler.Start("cpu.prof")
//	// ... run workload
//	cpuProfiler.Stop()
//	cpuProfiler.Analyze()
func UseCPUProfiler(ctx *bubbly.Context) *CPUProfilerComposable {
	// Create the underlying CPU profiler
	prof := profiler.NewCPUProfiler()

	// Create reactive state using typed refs
	state := bubbly.NewRef(StateIdle)
	filename := bubbly.NewRef("")
	startTime := bubbly.NewRef(time.Time{})
	fileSize := bubbly.NewRef(int64(0))
	hotFunctions := bubbly.NewRef([]HotFunctionInfo{})
	lastError := bubbly.NewRef("")

	// Create the composable instance
	comp := &CPUProfilerComposable{
		Profiler:     prof,
		State:        state,
		Filename:     filename,
		StartTime:    startTime,
		FileSize:     fileSize,
		HotFunctions: hotFunctions,
		LastError:    lastError,
	}

	// Define Start method
	comp.Start = func(fname string) error {
		comp.mu.Lock()
		defer comp.mu.Unlock()

		// Clear any previous error
		lastError.Set("")

		// Start CPU profiling
		if err := prof.Start(fname); err != nil {
			lastError.Set(err.Error())
			return err
		}

		// Update state
		state.Set(StateProfiling)
		filename.Set(fname)
		startTime.Set(time.Now())
		fileSize.Set(0)
		hotFunctions.Set([]HotFunctionInfo{})

		return nil
	}

	// Define Stop method
	comp.Stop = func() error {
		comp.mu.Lock()
		defer comp.mu.Unlock()

		// Stop CPU profiling
		if err := prof.Stop(); err != nil {
			lastError.Set(err.Error())
			return err
		}

		// Get file size
		fname := filename.GetTyped()
		if info, err := os.Stat(fname); err == nil {
			fileSize.Set(info.Size())
		}

		// Update state
		state.Set(StateComplete)

		return nil
	}

	// Define Analyze method
	comp.Analyze = func() error {
		comp.mu.Lock()
		defer comp.mu.Unlock()

		// Can only analyze when complete
		if state.GetTyped() != StateComplete {
			err := fmt.Errorf("cannot analyze: profiler is not in complete state")
			lastError.Set(err.Error())
			return err
		}

		// For this demo, we'll generate sample hot functions
		// In a real implementation, we'd parse the pprof file
		// using runtime/pprof or go tool pprof
		sampleFunctions := []HotFunctionInfo{
			{Name: "runtime.mallocgc", Samples: 1250, Percent: 25.0},
			{Name: "main.processData", Samples: 800, Percent: 16.0},
			{Name: "encoding/json.(*decodeState).object", Samples: 650, Percent: 13.0},
			{Name: "runtime.scanobject", Samples: 500, Percent: 10.0},
			{Name: "github.com/newbpydev/bubblyui/pkg/bubbly.(*componentImpl).View", Samples: 400, Percent: 8.0},
			{Name: "runtime.gcDrain", Samples: 350, Percent: 7.0},
			{Name: "strings.(*Builder).WriteString", Samples: 300, Percent: 6.0},
			{Name: "fmt.Sprintf", Samples: 250, Percent: 5.0},
		}

		hotFunctions.Set(sampleFunctions)
		return nil
	}

	// Define Reset method
	comp.Reset = func() {
		comp.mu.Lock()
		defer comp.mu.Unlock()

		// If profiling is active, stop it first
		if state.GetTyped() == StateProfiling {
			prof.Stop()
		}

		// Reset all state
		state.Set(StateIdle)
		filename.Set("")
		startTime.Set(time.Time{})
		fileSize.Set(0)
		hotFunctions.Set([]HotFunctionInfo{})
		lastError.Set("")
	}

	return comp
}

// FormatBytes formats bytes into a human-readable string.
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// FormatDuration formats a duration into a human-readable string.
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
