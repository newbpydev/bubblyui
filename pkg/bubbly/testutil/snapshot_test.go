package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSnapshotManager_Creation tests SnapshotManager creation
func TestSnapshotManager_Creation(t *testing.T) {
	tests := []struct {
		name      string
		testDir   string
		update    bool
		wantDir   string
		wantPanic bool
	}{
		{
			name:    "creates with valid directory",
			testDir: t.TempDir(),
			update:  false,
			wantDir: "__snapshots__",
		},
		{
			name:    "creates with update mode enabled",
			testDir: t.TempDir(),
			update:  true,
			wantDir: "__snapshots__",
		},
		{
			name:    "creates with empty directory",
			testDir: "",
			update:  false,
			wantDir: "__snapshots__",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSnapshotManager(tt.testDir, tt.update)

			assert.NotNil(t, sm)
			assert.Contains(t, sm.dir, tt.wantDir)
		})
	}
}

// TestSnapshotManager_Match_CreateSnapshot tests snapshot creation
func TestSnapshotManager_Match_CreateSnapshot(t *testing.T) {
	tests := []struct {
		name     string
		snapName string
		content  string
	}{
		{
			name:     "creates simple text snapshot",
			snapName: "simple",
			content:  "Hello World",
		},
		{
			name:     "creates multiline snapshot",
			snapName: "multiline",
			content:  "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "creates empty snapshot",
			snapName: "empty",
			content:  "",
		},
		{
			name:     "creates snapshot with special characters",
			snapName: "special",
			content:  "Special: !@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			sm := NewSnapshotManager(tempDir, false)

			// First call should create snapshot
			sm.Match(t, tt.snapName, tt.content)

			// Verify snapshot file exists
			snapFile := filepath.Join(tempDir, "__snapshots__", tt.snapName+".snap")
			assert.FileExists(t, snapFile)

			// Verify content
			data, err := os.ReadFile(snapFile)
			require.NoError(t, err)
			assert.Equal(t, tt.content, string(data))
		})
	}
}

// TestSnapshotManager_Match_CompareSnapshot tests snapshot comparison
func TestSnapshotManager_Match_CompareSnapshot(t *testing.T) {
	tests := []struct {
		name        string
		snapName    string
		initial     string
		actual      string
		shouldMatch bool
	}{
		{
			name:        "matches identical content",
			snapName:    "identical",
			initial:     "Hello World",
			actual:      "Hello World",
			shouldMatch: true,
		},
		{
			name:        "detects content mismatch",
			snapName:    "mismatch",
			initial:     "Hello World",
			actual:      "Hello Universe",
			shouldMatch: false,
		},
		{
			name:        "detects multiline mismatch",
			snapName:    "multiline_mismatch",
			initial:     "Line 1\nLine 2",
			actual:      "Line 1\nLine 3",
			shouldMatch: false,
		},
		{
			name:        "matches empty content",
			snapName:    "empty_match",
			initial:     "",
			actual:      "",
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			sm := NewSnapshotManager(tempDir, false)

			// Create initial snapshot
			sm.Match(t, tt.snapName, tt.initial)

			// Create mock testing.T to capture failures
			mockT := &mockTestingT{}

			// Compare with actual content
			sm.Match(mockT, tt.snapName, tt.actual)

			if tt.shouldMatch {
				assert.False(t, mockT.failed, "Expected match but test failed")
			} else {
				assert.True(t, mockT.failed, "Expected mismatch but test passed")
			}
		})
	}
}

// TestSnapshotManager_Match_UpdateMode tests update mode
func TestSnapshotManager_Match_UpdateMode(t *testing.T) {
	tests := []struct {
		name     string
		snapName string
		initial  string
		updated  string
	}{
		{
			name:     "updates snapshot in update mode",
			snapName: "update_test",
			initial:  "Old Content",
			updated:  "New Content",
		},
		{
			name:     "updates multiline snapshot",
			snapName: "multiline_update",
			initial:  "Line 1\nLine 2",
			updated:  "Line A\nLine B\nLine C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create initial snapshot (update mode off)
			sm1 := NewSnapshotManager(tempDir, false)
			sm1.Match(t, tt.snapName, tt.initial)

			// Update snapshot (update mode on)
			sm2 := NewSnapshotManager(tempDir, true)
			sm2.Match(t, tt.snapName, tt.updated)

			// Verify snapshot was updated
			snapFile := filepath.Join(tempDir, "__snapshots__", tt.snapName+".snap")
			data, err := os.ReadFile(snapFile)
			require.NoError(t, err)
			assert.Equal(t, tt.updated, string(data))
		})
	}
}

// TestSnapshotManager_Match_ThreadSafe tests thread-safe operations
func TestSnapshotManager_Match_ThreadSafe(t *testing.T) {
	tempDir := t.TempDir()
	sm := NewSnapshotManager(tempDir, false)

	// Run concurrent snapshot operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			snapName := "concurrent"
			content := "Concurrent Test"
			sm.Match(t, snapName, content)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify snapshot exists and is valid
	snapFile := filepath.Join(tempDir, "__snapshots__", "concurrent.snap")
	assert.FileExists(t, snapFile)
}

// TestSnapshotManager_Match_FileFormat tests snapshot file format
func TestSnapshotManager_Match_FileFormat(t *testing.T) {
	tempDir := t.TempDir()
	sm := NewSnapshotManager(tempDir, false)

	content := "Test Content"
	sm.Match(t, "format_test", content)

	// Verify file has .snap extension
	snapFile := filepath.Join(tempDir, "__snapshots__", "format_test.snap")
	assert.FileExists(t, snapFile)

	// Verify content is stored correctly
	data, err := os.ReadFile(snapFile)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

// TestSnapshotManager_Match_DirectoryCreation tests directory creation
func TestSnapshotManager_Match_DirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()
	sm := NewSnapshotManager(tempDir, false)

	sm.Match(t, "dir_test", "Content")

	// Verify __snapshots__ directory was created
	snapDir := filepath.Join(tempDir, "__snapshots__")
	info, err := os.Stat(snapDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

// TestSnapshotManager_createSnapshot_EdgeCases tests error paths for createSnapshot
func TestSnapshotManager_createSnapshot_EdgeCases(t *testing.T) {
	sm := NewSnapshotManager(t.TempDir(), false)

	// Test createSnapshot with invalid directory path (permission error)
	mockT := &mockTestingT{}
	// Use a directory that likely doesn't exist or isn't writable
	invalidPath := "/root/nonexistent/path/snapshot.txt"
	sm.createSnapshot(mockT, invalidPath, "test content")
	assert.True(t, mockT.failed, "should fail when directory creation fails")
	assert.Contains(t, mockT.errors[0], "failed to create snapshot dir")

	// Test createSnapshot with valid directory but invalid file path
	mockT = &mockTestingT{}
	// Use a directory that exists but try to write to an invalid location
	validDir := t.TempDir()
	invalidFile := filepath.Join(validDir, "subdir", "snapshot.txt")
	sm.createSnapshot(mockT, invalidFile, "test content")
	// This should succeed because MkdirAll creates the directory
	assert.False(t, mockT.failed, "should succeed when directory can be created")
}

// TestSnapshotManager_updateSnapshot_EdgeCases tests error paths for updateSnapshot
func TestSnapshotManager_updateSnapshot_EdgeCases(t *testing.T) {
	sm := NewSnapshotManager(t.TempDir(), false)

	// Test updateSnapshot with invalid file path (permission error)
	mockT := &mockTestingT{}
	invalidPath := "/root/nonexistent/path/snapshot.txt"
	sm.updateSnapshot(mockT, invalidPath, "test content")
	assert.True(t, mockT.failed, "should fail when file write fails")
	assert.Contains(t, mockT.errors[0], "failed to update snapshot")
}

// mockTestingT is already defined in assertions_state_test.go
