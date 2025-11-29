// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"os"
	"sync"
	"testing"

	"github.com/google/pprof/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewStackAnalyzer tests the creation of a new StackAnalyzer.
func TestNewStackAnalyzer(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates new stack analyzer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sa := NewStackAnalyzer()
			assert.NotNil(t, sa)
			assert.NotNil(t, sa.samples)
			assert.Empty(t, sa.samples)
		})
	}
}

// TestStackAnalyzer_Analyze tests the Analyze method with various profile inputs.
func TestStackAnalyzer_Analyze(t *testing.T) {
	tests := []struct {
		name            string
		profile         *profile.Profile
		wantHotFuncs    int
		wantTotalSample int64
		wantCallGraph   bool
	}{
		{
			name:            "nil profile returns empty data",
			profile:         nil,
			wantHotFuncs:    0,
			wantTotalSample: 0,
			wantCallGraph:   false,
		},
		{
			name:            "empty profile returns empty data",
			profile:         &profile.Profile{},
			wantHotFuncs:    0,
			wantTotalSample: 0,
			wantCallGraph:   false,
		},
		{
			name:            "profile with samples extracts hot functions",
			profile:         createTestProfile(),
			wantHotFuncs:    2, // main.foo and main.bar
			wantTotalSample: 150,
			wantCallGraph:   false, // Simple profile has no call graph (single location per sample)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sa := NewStackAnalyzer()
			data := sa.Analyze(tt.profile)

			assert.NotNil(t, data)
			assert.Len(t, data.HotFunctions, tt.wantHotFuncs)
			assert.Equal(t, tt.wantTotalSample, data.TotalSamples)

			if tt.wantCallGraph {
				assert.NotEmpty(t, data.CallGraph)
			}
		})
	}
}

// TestStackAnalyzer_HotFunctionDetection tests that hot functions are correctly identified.
func TestStackAnalyzer_HotFunctionDetection(t *testing.T) {
	tests := []struct {
		name         string
		profile      *profile.Profile
		wantTopFunc  string
		wantTopPct   float64
		wantMinFuncs int
	}{
		{
			name:         "identifies hottest function",
			profile:      createTestProfile(),
			wantTopFunc:  "main.foo",
			wantTopPct:   66.67, // 100 out of 150 samples
			wantMinFuncs: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sa := NewStackAnalyzer()
			data := sa.Analyze(tt.profile)

			require.GreaterOrEqual(t, len(data.HotFunctions), tt.wantMinFuncs)

			// First function should be the hottest
			topFunc := data.HotFunctions[0]
			assert.Equal(t, tt.wantTopFunc, topFunc.Name)
			assert.InDelta(t, tt.wantTopPct, topFunc.Percent, 0.1)
		})
	}
}

// TestStackAnalyzer_CallGraphBuilding tests that call graphs are correctly built.
func TestStackAnalyzer_CallGraphBuilding(t *testing.T) {
	tests := []struct {
		name        string
		profile     *profile.Profile
		wantCallers map[string][]string
	}{
		{
			name:    "builds call graph from samples",
			profile: createTestProfileWithCallGraph(),
			wantCallers: map[string][]string{
				"main.main": {"main.foo"},
				"main.foo":  {"main.bar"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sa := NewStackAnalyzer()
			data := sa.Analyze(tt.profile)

			for caller, expectedCallees := range tt.wantCallers {
				callees, ok := data.CallGraph[caller]
				assert.True(t, ok, "expected caller %s in call graph", caller)
				for _, expectedCallee := range expectedCallees {
					assert.Contains(t, callees, expectedCallee)
				}
			}
		})
	}
}

// TestStackAnalyzer_PercentageCalculation tests that percentages are calculated correctly.
func TestStackAnalyzer_PercentageCalculation(t *testing.T) {
	tests := []struct {
		name           string
		profile        *profile.Profile
		wantTotalPct   float64
		wantEachPctGt0 bool
	}{
		{
			name:           "percentages sum to approximately 100",
			profile:        createTestProfile(),
			wantTotalPct:   100.0,
			wantEachPctGt0: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sa := NewStackAnalyzer()
			data := sa.Analyze(tt.profile)

			var totalPct float64
			for _, hf := range data.HotFunctions {
				if tt.wantEachPctGt0 {
					assert.Greater(t, hf.Percent, 0.0)
				}
				totalPct += hf.Percent
			}

			assert.InDelta(t, tt.wantTotalPct, totalPct, 0.1)
		})
	}
}

// TestStackAnalyzer_SortingBySamples tests that hot functions are sorted by sample count.
func TestStackAnalyzer_SortingBySamples(t *testing.T) {
	tests := []struct {
		name    string
		profile *profile.Profile
	}{
		{
			name:    "hot functions sorted by samples descending",
			profile: createTestProfile(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sa := NewStackAnalyzer()
			data := sa.Analyze(tt.profile)

			// Verify sorted in descending order
			for i := 1; i < len(data.HotFunctions); i++ {
				assert.GreaterOrEqual(t,
					data.HotFunctions[i-1].Samples,
					data.HotFunctions[i].Samples,
					"hot functions should be sorted by samples descending")
			}
		})
	}
}

// TestStackAnalyzer_ThreadSafety tests concurrent access to the analyzer.
func TestStackAnalyzer_ThreadSafety(t *testing.T) {
	sa := NewStackAnalyzer()
	prof := createTestProfile()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data := sa.Analyze(prof)
			assert.NotNil(t, data)
		}()
	}
	wg.Wait()
}

// TestStackAnalyzer_AnalyzeFromFile tests analyzing a real pprof file.
func TestStackAnalyzer_AnalyzeFromFile(t *testing.T) {
	// Create a temporary CPU profile
	tmpFile, err := os.CreateTemp("", "cpu_profile_*.prof")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Create and write a test profile
	testProf := createTestProfile()
	err = testProf.Write(tmpFile)
	require.NoError(t, err)
	tmpFile.Close()

	// Read and analyze
	f, err := os.Open(tmpFile.Name())
	require.NoError(t, err)
	defer f.Close()

	parsedProf, err := profile.Parse(f)
	require.NoError(t, err)

	sa := NewStackAnalyzer()
	data := sa.Analyze(parsedProf)

	assert.NotNil(t, data)
	assert.NotEmpty(t, data.HotFunctions)
}

// TestStackAnalyzer_Reset tests resetting the analyzer state.
func TestStackAnalyzer_Reset(t *testing.T) {
	sa := NewStackAnalyzer()

	// Analyze a profile
	prof := createTestProfile()
	data := sa.Analyze(prof)
	assert.NotEmpty(t, data.HotFunctions)

	// Reset
	sa.Reset()

	// Verify reset
	assert.Empty(t, sa.samples)
}

// TestStackAnalyzer_GetSamples tests getting the raw sample counts.
func TestStackAnalyzer_GetSamples(t *testing.T) {
	sa := NewStackAnalyzer()
	prof := createTestProfile()
	sa.Analyze(prof)

	samples := sa.GetSamples()
	assert.NotEmpty(t, samples)
	assert.Contains(t, samples, "main.foo")
	assert.Contains(t, samples, "main.bar")
}

// Helper function to create a test profile with samples
func createTestProfile() *profile.Profile {
	// Create functions
	fooFunc := &profile.Function{
		ID:         1,
		Name:       "main.foo",
		SystemName: "main.foo",
		Filename:   "main.go",
		StartLine:  10,
	}
	barFunc := &profile.Function{
		ID:         2,
		Name:       "main.bar",
		SystemName: "main.bar",
		Filename:   "main.go",
		StartLine:  20,
	}

	// Create locations
	fooLoc := &profile.Location{
		ID:      1,
		Address: 0x1000,
		Line: []profile.Line{
			{Function: fooFunc, Line: 15},
		},
	}
	barLoc := &profile.Location{
		ID:      2,
		Address: 0x2000,
		Line: []profile.Line{
			{Function: barFunc, Line: 25},
		},
	}

	// Create samples
	// Sample 1: foo with 100 samples
	sample1 := &profile.Sample{
		Location: []*profile.Location{fooLoc},
		Value:    []int64{100}, // 100 samples
	}
	// Sample 2: bar with 50 samples
	sample2 := &profile.Sample{
		Location: []*profile.Location{barLoc},
		Value:    []int64{50}, // 50 samples
	}

	return &profile.Profile{
		SampleType: []*profile.ValueType{
			{Type: "samples", Unit: "count"},
		},
		Sample:   []*profile.Sample{sample1, sample2},
		Location: []*profile.Location{fooLoc, barLoc},
		Function: []*profile.Function{fooFunc, barFunc},
	}
}

// Helper function to create a test profile with a call graph
func createTestProfileWithCallGraph() *profile.Profile {
	// Create functions
	mainFunc := &profile.Function{
		ID:         1,
		Name:       "main.main",
		SystemName: "main.main",
		Filename:   "main.go",
		StartLine:  1,
	}
	fooFunc := &profile.Function{
		ID:         2,
		Name:       "main.foo",
		SystemName: "main.foo",
		Filename:   "main.go",
		StartLine:  10,
	}
	barFunc := &profile.Function{
		ID:         3,
		Name:       "main.bar",
		SystemName: "main.bar",
		Filename:   "main.go",
		StartLine:  20,
	}

	// Create locations
	mainLoc := &profile.Location{
		ID:      1,
		Address: 0x1000,
		Line: []profile.Line{
			{Function: mainFunc, Line: 5},
		},
	}
	fooLoc := &profile.Location{
		ID:      2,
		Address: 0x2000,
		Line: []profile.Line{
			{Function: fooFunc, Line: 15},
		},
	}
	barLoc := &profile.Location{
		ID:      3,
		Address: 0x3000,
		Line: []profile.Line{
			{Function: barFunc, Line: 25},
		},
	}

	// Create sample with call stack: main -> foo -> bar
	// In pprof, the stack is ordered from leaf to root
	sample := &profile.Sample{
		Location: []*profile.Location{barLoc, fooLoc, mainLoc},
		Value:    []int64{100},
	}

	return &profile.Profile{
		SampleType: []*profile.ValueType{
			{Type: "samples", Unit: "count"},
		},
		Sample:   []*profile.Sample{sample},
		Location: []*profile.Location{mainLoc, fooLoc, barLoc},
		Function: []*profile.Function{mainFunc, fooFunc, barFunc},
	}
}
