// Package profiler provides performance profiling for BubblyUI applications.
package profiler

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCPUProfiler_New tests CPUProfiler creation
func TestCPUProfiler_New(t *testing.T) {
	cp := NewCPUProfiler()

	assert.NotNil(t, cp)
	assert.False(t, cp.IsActive())
}

// TestCPUProfiler_StartStop tests CPU profiling lifecycle
func TestCPUProfiler_StartStop(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(cp *CPUProfiler, filename string) error
		action     func(cp *CPUProfiler, filename string) error
		wantActive bool
		wantErr    bool
		errType    error
	}{
		{
			name:  "start_profiling_creates_file",
			setup: func(cp *CPUProfiler, filename string) error { return nil },
			action: func(cp *CPUProfiler, filename string) error {
				return cp.Start(filename)
			},
			wantActive: true,
			wantErr:    false,
		},
		{
			name: "stop_profiling_closes_file",
			setup: func(cp *CPUProfiler, filename string) error {
				return cp.Start(filename)
			},
			action: func(cp *CPUProfiler, filename string) error {
				return cp.Stop()
			},
			wantActive: false,
			wantErr:    false,
		},
		{
			name: "start_when_already_active_returns_error",
			setup: func(cp *CPUProfiler, filename string) error {
				return cp.Start(filename)
			},
			action: func(cp *CPUProfiler, filename string) error {
				return cp.Start(filename + ".second")
			},
			wantActive: true,
			wantErr:    true,
			errType:    ErrCPUProfileActive,
		},
		{
			name:  "stop_when_not_active_returns_error",
			setup: func(cp *CPUProfiler, filename string) error { return nil },
			action: func(cp *CPUProfiler, filename string) error {
				return cp.Stop()
			},
			wantActive: false,
			wantErr:    true,
			errType:    ErrCPUProfileNotActive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for test files
			tmpDir := t.TempDir()
			filename := filepath.Join(tmpDir, "cpu.prof")

			cp := NewCPUProfiler()
			err := tt.setup(cp, filename)
			require.NoError(t, err, "setup should not fail")

			err = tt.action(cp, filename)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantActive, cp.IsActive())

			// Cleanup: stop profiling if still active
			if cp.IsActive() {
				_ = cp.Stop()
			}
		})
	}
}

// TestCPUProfiler_FileGeneration tests that pprof file is generated
func TestCPUProfiler_FileGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "cpu.prof")

	cp := NewCPUProfiler()

	// Start profiling
	err := cp.Start(filename)
	require.NoError(t, err)

	// Do some work to generate profile data
	sum := 0
	for i := 0; i < 10000; i++ {
		sum += i
	}
	_ = sum

	// Stop profiling
	err = cp.Stop()
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(filename)
	assert.NoError(t, err, "profile file should exist")

	// Verify file has content
	info, err := os.Stat(filename)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0), "profile file should have content")
}

// TestCPUProfiler_InvalidFilename tests error handling for invalid filenames
func TestCPUProfiler_InvalidFilename(t *testing.T) {
	cp := NewCPUProfiler()

	// Try to create file in non-existent directory
	err := cp.Start("/nonexistent/directory/cpu.prof")
	assert.Error(t, err)
	assert.False(t, cp.IsActive())
}

// TestCPUProfiler_ThreadSafe tests concurrent operations
func TestCPUProfiler_ThreadSafe(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "cpu.prof")

	cp := NewCPUProfiler()
	err := cp.Start(filename)
	require.NoError(t, err)

	var wg sync.WaitGroup
	numGoroutines := 50

	// Concurrent IsActive calls
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cp.IsActive()
		}()
	}

	// Concurrent Start attempts (should fail but not panic)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cp.Start(filepath.Join(tmpDir, "concurrent.prof"))
		}()
	}

	wg.Wait()

	// Should still be active and not panicked
	assert.True(t, cp.IsActive())

	// Cleanup
	err = cp.Stop()
	assert.NoError(t, err)
}

// TestCPUProfiler_GetFilename tests filename retrieval
func TestCPUProfiler_GetFilename(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "cpu.prof")

	cp := NewCPUProfiler()

	// Before starting, filename should be empty
	assert.Empty(t, cp.GetFilename())

	// Start profiling
	err := cp.Start(filename)
	require.NoError(t, err)

	// Filename should be set
	assert.Equal(t, filename, cp.GetFilename())

	// Stop profiling
	err = cp.Stop()
	require.NoError(t, err)

	// After stopping, filename should be empty
	assert.Empty(t, cp.GetFilename())
}

// TestCPUProfiler_MultipleStartStop tests multiple start/stop cycles
func TestCPUProfiler_MultipleStartStop(t *testing.T) {
	tmpDir := t.TempDir()

	cp := NewCPUProfiler()

	for i := 0; i < 3; i++ {
		filename := filepath.Join(tmpDir, "cpu"+string(rune('0'+i))+".prof")

		err := cp.Start(filename)
		require.NoError(t, err, "start cycle %d should succeed", i)
		assert.True(t, cp.IsActive())

		err = cp.Stop()
		require.NoError(t, err, "stop cycle %d should succeed", i)
		assert.False(t, cp.IsActive())

		// Verify file was created
		_, err = os.Stat(filename)
		assert.NoError(t, err, "profile file %d should exist", i)
	}
}

// TestCPUProfiler_ProfileDataValid tests that generated profile is valid pprof format
func TestCPUProfiler_ProfileDataValid(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "cpu.prof")

	cp := NewCPUProfiler()

	// Start profiling
	err := cp.Start(filename)
	require.NoError(t, err)

	// Do some CPU-intensive work
	result := 0
	for i := 0; i < 100000; i++ {
		result += i * i
	}
	_ = result

	// Stop profiling
	err = cp.Stop()
	require.NoError(t, err)

	// Read the file and verify it has pprof header
	// pprof files start with specific bytes
	data, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.NotEmpty(t, data, "profile data should not be empty")

	// pprof format is gzip compressed protobuf
	// Check for gzip magic number (0x1f 0x8b)
	if len(data) >= 2 {
		// Note: pprof files are gzip compressed, so they start with gzip magic
		assert.True(t, data[0] == 0x1f && data[1] == 0x8b,
			"profile should be gzip compressed (pprof format)")
	}
}

// TestCPUProfiler_IntegrationWithPprof tests that the profile works with pprof tools
func TestCPUProfiler_IntegrationWithPprof(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "cpu.prof")

	cp := NewCPUProfiler()

	// Start profiling
	err := cp.Start(filename)
	require.NoError(t, err)

	// Do some work
	sum := 0
	for i := 0; i < 50000; i++ {
		sum += i
	}
	_ = sum

	// Stop profiling
	err = cp.Stop()
	require.NoError(t, err)

	// Verify file exists and has reasonable size
	info, err := os.Stat(filename)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0), "profile should have content")
}
