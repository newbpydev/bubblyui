// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"sort"
	"sync"

	"github.com/google/pprof/profile"
)

// StackAnalyzer analyzes CPU profile data to identify hot functions and build call graphs.
//
// It parses pprof Profile data and extracts:
//   - Hot functions (functions consuming the most CPU time)
//   - Call graph (caller-callee relationships)
//   - Sample counts and percentages
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	sa := NewStackAnalyzer()
//	data := sa.Analyze(profile)
//	for _, hf := range data.HotFunctions {
//	    fmt.Printf("%s: %.2f%% (%d samples)\n", hf.Name, hf.Percent, hf.Samples)
//	}
type StackAnalyzer struct {
	// samples maps function names to their sample counts
	samples map[string]int64

	// mu protects concurrent access to samples
	mu sync.RWMutex
}

// NewStackAnalyzer creates a new StackAnalyzer instance.
//
// The analyzer is created with an empty sample map, ready to analyze profiles.
//
// Example:
//
//	sa := NewStackAnalyzer()
//	data := sa.Analyze(cpuProfile)
func NewStackAnalyzer() *StackAnalyzer {
	return &StackAnalyzer{
		samples: make(map[string]int64),
	}
}

// Analyze processes a pprof Profile and returns CPU profiling data.
//
// The method extracts hot functions from the profile samples, calculates
// the percentage of CPU time for each function, builds a call graph from
// the sample stack traces, and sorts results by sample count.
//
// Returns an empty CPUProfileData if the profile is nil or has no samples.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	f, _ := os.Open("cpu.prof")
//	prof, _ := profile.Parse(f)
//	data := sa.Analyze(prof)
func (sa *StackAnalyzer) Analyze(prof *profile.Profile) *CPUProfileData {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	// Reset samples for fresh analysis
	sa.samples = make(map[string]int64)

	data := &CPUProfileData{
		HotFunctions: make([]*HotFunction, 0),
		CallGraph:    make(map[string][]string),
		TotalSamples: 0,
	}

	if prof == nil || len(prof.Sample) == 0 {
		return data
	}

	// Process all samples
	for _, sample := range prof.Sample {
		if len(sample.Value) == 0 {
			continue
		}

		// Get the sample count (first value is typically the count)
		sampleCount := sample.Value[0]
		data.TotalSamples += sampleCount

		// Process each location in the sample
		for _, loc := range sample.Location {
			for _, line := range loc.Line {
				if line.Function != nil {
					funcName := line.Function.Name
					sa.samples[funcName] += sampleCount
				}
			}
		}

		// Build call graph from the stack trace
		// In pprof, locations are ordered from leaf (index 0) to root (last index)
		sa.buildCallGraph(data, sample)
	}

	// Convert samples to hot functions
	for funcName, sampleCount := range sa.samples {
		percent := 0.0
		if data.TotalSamples > 0 {
			percent = float64(sampleCount) / float64(data.TotalSamples) * 100.0
		}

		data.HotFunctions = append(data.HotFunctions, &HotFunction{
			Name:    funcName,
			Samples: sampleCount,
			Percent: percent,
		})
	}

	// Sort hot functions by samples descending
	sort.Slice(data.HotFunctions, func(i, j int) bool {
		return data.HotFunctions[i].Samples > data.HotFunctions[j].Samples
	})

	return data
}

// buildCallGraph extracts caller-callee relationships from a sample's stack trace.
//
// In pprof profiles, the Location slice is ordered from leaf (index 0) to root.
// This method builds a map where each caller maps to its callees.
func (sa *StackAnalyzer) buildCallGraph(data *CPUProfileData, sample *profile.Sample) {
	if len(sample.Location) < 2 {
		return
	}

	// Walk the stack from leaf to root
	// Location[i] is called by Location[i+1]
	for i := 0; i < len(sample.Location)-1; i++ {
		calleeLoc := sample.Location[i]
		callerLoc := sample.Location[i+1]

		calleeName := getFirstFunctionName(calleeLoc)
		callerName := getFirstFunctionName(callerLoc)

		if callerName == "" || calleeName == "" {
			continue
		}

		// Add callee to caller's list if not already present
		callees := data.CallGraph[callerName]
		if !containsString(callees, calleeName) {
			data.CallGraph[callerName] = append(callees, calleeName)
		}
	}
}

// getFirstFunctionName extracts the function name from a location.
//
// Returns an empty string if the location has no function information.
func getFirstFunctionName(loc *profile.Location) string {
	if loc == nil || len(loc.Line) == 0 {
		return ""
	}
	if loc.Line[0].Function == nil {
		return ""
	}
	return loc.Line[0].Function.Name
}

// containsString checks if a slice contains a specific string.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// Reset clears the analyzer's internal state.
//
// Call this method to prepare the analyzer for a new analysis session.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	sa.Reset()
//	data := sa.Analyze(newProfile)
func (sa *StackAnalyzer) Reset() {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	sa.samples = make(map[string]int64)
}

// GetSamples returns a copy of the current sample counts.
//
// The returned map contains function names as keys and sample counts as values.
// Modifying the returned map does not affect the analyzer's internal state.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	samples := sa.GetSamples()
//	for funcName, count := range samples {
//	    fmt.Printf("%s: %d samples\n", funcName, count)
//	}
func (sa *StackAnalyzer) GetSamples() map[string]int64 {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]int64, len(sa.samples))
	for k, v := range sa.samples {
		result[k] = v
	}
	return result
}
