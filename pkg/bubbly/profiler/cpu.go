// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"errors"
	"os"
	"runtime/pprof"
	"sync"
)

// CPUProfiler handles CPU profiling with pprof integration.
//
// It provides methods to start and stop CPU profiling, generating pprof-compatible
// profile files that can be analyzed with Go's pprof tools.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	cp := NewCPUProfiler()
//	if err := cp.Start("cpu.prof"); err != nil {
//	    log.Fatal(err)
//	}
//	defer cp.Stop()
//
//	// Run workload
//	runApplication()
//
//	// Analyze with: go tool pprof cpu.prof
type CPUProfiler struct {
	// active indicates whether CPU profiling is currently running
	active bool

	// file is the open file handle for the profile output
	file *os.File

	// filename stores the current profile filename
	filename string

	// mu protects concurrent access to profiler state
	mu sync.Mutex
}

// Common errors for CPU profiling
var (
	// ErrCPUProfileActive is returned when Start() is called while profiling is already active
	ErrCPUProfileActive = errors.New("CPU profiling already active")

	// ErrCPUProfileNotActive is returned when Stop() is called while profiling is not active
	ErrCPUProfileNotActive = errors.New("CPU profiling not active")
)

// NewCPUProfiler creates a new CPU profiler instance.
//
// The profiler is created in an inactive state. Call Start() to begin profiling.
//
// Example:
//
//	cp := NewCPUProfiler()
//	err := cp.Start("cpu.prof")
func NewCPUProfiler() *CPUProfiler {
	return &CPUProfiler{}
}

// Start begins CPU profiling and writes output to the specified file.
//
// The filename should be a valid path where the profile data will be written.
// The file is created if it doesn't exist, or truncated if it does.
//
// Returns ErrCPUProfileActive if profiling is already active.
// Returns an error if the file cannot be created.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	err := cp.Start("cpu.prof")
//	if err != nil {
//	    log.Fatal("could not start CPU profile:", err)
//	}
//	defer cp.Stop()
func (cp *CPUProfiler) Start(filename string) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if cp.active {
		return ErrCPUProfileActive
	}

	// Create the output file
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	// Start CPU profiling
	if err := pprof.StartCPUProfile(f); err != nil {
		// Close the file if we fail to start profiling
		f.Close()
		return err
	}

	cp.file = f
	cp.filename = filename
	cp.active = true

	return nil
}

// Stop ends CPU profiling and closes the output file.
//
// The profile data is flushed to the file before closing.
// After calling Stop(), the profile file can be analyzed with pprof tools.
//
// Returns ErrCPUProfileNotActive if profiling is not active.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	err := cp.Stop()
//	if err != nil {
//	    log.Fatal("could not stop CPU profile:", err)
//	}
//	// Now analyze with: go tool pprof cpu.prof
func (cp *CPUProfiler) Stop() error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if !cp.active {
		return ErrCPUProfileNotActive
	}

	// Stop CPU profiling (flushes data to file)
	pprof.StopCPUProfile()

	// Close the file
	err := cp.file.Close()

	// Reset state regardless of close error
	cp.file = nil
	cp.filename = ""
	cp.active = false

	return err
}

// IsActive returns whether CPU profiling is currently running.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	if cp.IsActive() {
//	    fmt.Println("CPU profiling is running")
//	}
func (cp *CPUProfiler) IsActive() bool {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	return cp.active
}

// GetFilename returns the current profile filename.
//
// Returns an empty string if profiling is not active.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	if filename := cp.GetFilename(); filename != "" {
//	    fmt.Printf("Profiling to: %s\n", filename)
//	}
func (cp *CPUProfiler) GetFilename() string {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	return cp.filename
}
