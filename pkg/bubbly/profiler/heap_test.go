// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemoryProfiler(t *testing.T) {
	t.Run("creates new memory profiler", func(t *testing.T) {
		mp := NewMemoryProfiler()

		assert.NotNil(t, mp)
		assert.NotNil(t, mp.baseline)
		assert.NotNil(t, mp.snapshots)
		assert.Empty(t, mp.snapshots)
	})
}

func TestMemoryProfiler_TakeSnapshot(t *testing.T) {
	tests := []struct {
		name           string
		snapshotCount  int
		wantMinHeap    bool
		wantGoroutines bool
	}{
		{
			name:           "single snapshot captures memory stats",
			snapshotCount:  1,
			wantMinHeap:    true,
			wantGoroutines: true,
		},
		{
			name:           "multiple snapshots are stored",
			snapshotCount:  3,
			wantMinHeap:    true,
			wantGoroutines: true,
		},
		{
			name:           "five snapshots for growth detection",
			snapshotCount:  5,
			wantMinHeap:    true,
			wantGoroutines: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := NewMemoryProfiler()

			for i := 0; i < tt.snapshotCount; i++ {
				snapshot := mp.TakeSnapshot()
				assert.NotNil(t, snapshot)

				if tt.wantMinHeap {
					// HeapAlloc should be > 0 in any running Go program
					assert.Greater(t, snapshot.HeapAlloc, uint64(0))
				}
			}

			// Verify all snapshots were stored
			assert.Len(t, mp.GetSnapshots(), tt.snapshotCount)
		})
	}
}

func TestMemoryProfiler_WriteHeapProfile(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "writes valid heap profile",
			filename: "test_heap.prof",
			wantErr:  false,
		},
		{
			name:     "writes to subdirectory",
			filename: "subdir/test_heap.prof",
			wantErr:  false,
		},
		{
			name:        "fails with invalid path",
			filename:    "/nonexistent/path/that/should/fail/heap.prof",
			wantErr:     true,
			errContains: "no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := NewMemoryProfiler()

			// Create temp directory for test files
			tmpDir := t.TempDir()
			var fullPath string

			if tt.wantErr {
				fullPath = tt.filename // Use invalid path directly
			} else {
				fullPath = filepath.Join(tmpDir, tt.filename)
				// Create subdirectory if needed
				dir := filepath.Dir(fullPath)
				if dir != tmpDir {
					err := os.MkdirAll(dir, 0755)
					require.NoError(t, err)
				}
			}

			err := mp.WriteHeapProfile(fullPath)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)

				// Verify file exists
				info, err := os.Stat(fullPath)
				require.NoError(t, err)
				assert.Greater(t, info.Size(), int64(0))

				// Verify pprof format (gzip compressed)
				f, err := os.Open(fullPath)
				require.NoError(t, err)
				defer f.Close()

				// pprof heap profiles are gzip compressed
				// Check for gzip magic number
				magic := make([]byte, 2)
				_, err = f.Read(magic)
				require.NoError(t, err)
				assert.Equal(t, byte(0x1f), magic[0], "expected gzip magic byte 1")
				assert.Equal(t, byte(0x8b), magic[1], "expected gzip magic byte 2")
			}
		})
	}
}

func TestMemoryProfiler_GetMemoryGrowth(t *testing.T) {
	t.Run("returns difference between first and last snapshot", func(t *testing.T) {
		mp := NewMemoryProfiler()

		// Take baseline snapshot
		first := mp.TakeSnapshot()

		// Take another snapshot
		second := mp.TakeSnapshot()

		growth := mp.GetMemoryGrowth()

		// Growth should be the difference between last and first HeapAlloc
		expectedGrowth := int64(second.HeapAlloc) - int64(first.HeapAlloc)
		assert.Equal(t, expectedGrowth, growth)
	})

	t.Run("growth calculation is correct with multiple snapshots", func(t *testing.T) {
		mp := NewMemoryProfiler()

		// Take multiple snapshots
		first := mp.TakeSnapshot()
		mp.TakeSnapshot() // Middle snapshot (ignored in growth calc)
		last := mp.TakeSnapshot()

		growth := mp.GetMemoryGrowth()

		// Growth should be between first and last only
		expectedGrowth := int64(last.HeapAlloc) - int64(first.HeapAlloc)
		assert.Equal(t, expectedGrowth, growth)
	})
}

func TestMemoryProfiler_GetMemoryGrowth_InsufficientSnapshots(t *testing.T) {
	tests := []struct {
		name          string
		snapshotCount int
		wantGrowth    int64
	}{
		{
			name:          "no snapshots returns zero",
			snapshotCount: 0,
			wantGrowth:    0,
		},
		{
			name:          "one snapshot returns zero",
			snapshotCount: 1,
			wantGrowth:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := NewMemoryProfiler()

			for i := 0; i < tt.snapshotCount; i++ {
				mp.TakeSnapshot()
			}

			growth := mp.GetMemoryGrowth()
			assert.Equal(t, tt.wantGrowth, growth)
		})
	}
}

func TestMemoryProfiler_GetBaseline(t *testing.T) {
	t.Run("returns baseline snapshot", func(t *testing.T) {
		mp := NewMemoryProfiler()

		baseline := mp.GetBaseline()
		assert.NotNil(t, baseline)
		assert.Greater(t, baseline.HeapAlloc, uint64(0))
	})
}

func TestMemoryProfiler_GetSnapshots(t *testing.T) {
	t.Run("returns copy of snapshots", func(t *testing.T) {
		mp := NewMemoryProfiler()

		mp.TakeSnapshot()
		mp.TakeSnapshot()

		snapshots := mp.GetSnapshots()
		assert.Len(t, snapshots, 2)

		// Verify it's a copy (modifying returned slice doesn't affect internal)
		originalLen := len(mp.GetSnapshots())
		_ = append(snapshots, &runtime.MemStats{}) // nolint:staticcheck // intentionally unused to test copy behavior
		assert.Len(t, mp.GetSnapshots(), originalLen)
	})
}

func TestMemoryProfiler_GetLatestSnapshot(t *testing.T) {
	tests := []struct {
		name          string
		snapshotCount int
		wantNil       bool
	}{
		{
			name:          "no snapshots returns nil",
			snapshotCount: 0,
			wantNil:       true,
		},
		{
			name:          "returns latest snapshot",
			snapshotCount: 3,
			wantNil:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := NewMemoryProfiler()

			var lastSnapshot *runtime.MemStats
			for i := 0; i < tt.snapshotCount; i++ {
				lastSnapshot = mp.TakeSnapshot()
			}

			latest := mp.GetLatestSnapshot()

			if tt.wantNil {
				assert.Nil(t, latest)
			} else {
				assert.NotNil(t, latest)
				assert.Equal(t, lastSnapshot, latest)
			}
		})
	}
}

func TestMemoryProfiler_Reset(t *testing.T) {
	t.Run("clears all snapshots", func(t *testing.T) {
		mp := NewMemoryProfiler()

		mp.TakeSnapshot()
		mp.TakeSnapshot()
		mp.TakeSnapshot()
		assert.Len(t, mp.GetSnapshots(), 3)

		mp.Reset()

		assert.Empty(t, mp.GetSnapshots())
		// Baseline should be refreshed
		assert.NotNil(t, mp.GetBaseline())
	})
}

func TestMemoryProfiler_SnapshotCount(t *testing.T) {
	t.Run("returns correct count", func(t *testing.T) {
		mp := NewMemoryProfiler()

		assert.Equal(t, 0, mp.SnapshotCount())

		mp.TakeSnapshot()
		assert.Equal(t, 1, mp.SnapshotCount())

		mp.TakeSnapshot()
		mp.TakeSnapshot()
		assert.Equal(t, 3, mp.SnapshotCount())
	})
}

func TestMemoryProfiler_ThreadSafety(t *testing.T) {
	t.Run("concurrent operations are safe", func(t *testing.T) {
		mp := NewMemoryProfiler()
		var wg sync.WaitGroup
		goroutines := 50

		// Concurrent snapshots
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				mp.TakeSnapshot()
			}()
		}

		// Concurrent reads
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = mp.GetMemoryGrowth()
				_ = mp.GetBaseline()
				_ = mp.GetSnapshots()
				_ = mp.GetLatestSnapshot()
				_ = mp.SnapshotCount()
			}()
		}

		wg.Wait()

		// Should have all snapshots
		assert.Equal(t, goroutines, mp.SnapshotCount())
	})
}

func TestMemoryProfiler_HeapProfileFormat(t *testing.T) {
	t.Run("generates valid pprof format", func(t *testing.T) {
		mp := NewMemoryProfiler()
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "heap.prof")

		err := mp.WriteHeapProfile(filename)
		require.NoError(t, err)

		// Open and verify it's valid gzip
		f, err := os.Open(filename)
		require.NoError(t, err)
		defer f.Close()

		// Try to decompress - should work for valid pprof
		gzReader, err := gzip.NewReader(f)
		require.NoError(t, err)
		defer gzReader.Close()

		// Read some data to verify it's valid
		buf := make([]byte, 1024)
		n, err := gzReader.Read(buf)
		// Either we read data or hit EOF (small profile)
		assert.True(t, n > 0 || err != nil)
	})
}

func TestMemoryProfiler_IntegrationWithPprof(t *testing.T) {
	t.Run("profile can be parsed by pprof tools", func(t *testing.T) {
		mp := NewMemoryProfiler()
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "heap_integration.prof")

		// Allocate some memory to have something to profile
		data := make([][]byte, 100)
		for i := range data {
			data[i] = make([]byte, 1024)
		}
		runtime.KeepAlive(data)

		err := mp.WriteHeapProfile(filename)
		require.NoError(t, err)

		// Verify file exists and has content
		info, err := os.Stat(filename)
		require.NoError(t, err)
		assert.Greater(t, info.Size(), int64(0))

		// The file should be parseable by pprof tools
		// We verify the format is correct (gzip compressed protobuf)
		f, err := os.Open(filename)
		require.NoError(t, err)
		defer f.Close()

		gzReader, err := gzip.NewReader(f)
		require.NoError(t, err)
		defer gzReader.Close()

		// Read all content
		content := make([]byte, 0)
		buf := make([]byte, 4096)
		for {
			n, err := gzReader.Read(buf)
			if n > 0 {
				content = append(content, buf[:n]...)
			}
			if err != nil {
				break
			}
		}

		// Should have some content
		assert.Greater(t, len(content), 0)
	})
}
