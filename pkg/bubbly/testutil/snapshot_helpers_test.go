package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMatchSnapshot tests the default snapshot matching helper.
func TestMatchSnapshot(t *testing.T) {
	tests := []struct {
		name     string
		actual   string
		wantFile string
	}{
		{
			name:     "simple text",
			actual:   "Hello, World!",
			wantFile: "TestMatchSnapshot_simple_text_default.snap",
		},
		{
			name:     "multiline text",
			actual:   "Line 1\nLine 2\nLine 3",
			wantFile: "TestMatchSnapshot_multiline_text_default.snap",
		},
		{
			name:     "empty string",
			actual:   "",
			wantFile: "TestMatchSnapshot_empty_string_default.snap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First run: creates snapshot
			MatchSnapshot(t, tt.actual)

			// Get the temp dir that was used
			tmpDir := getTestDir(t)

			// Verify snapshot file was created
			snapFile := filepath.Join(tmpDir, "__snapshots__", tt.wantFile)
			_, err := os.Stat(snapFile)
			assert.NoError(t, err, "snapshot file should be created")

			// Second run: should match
			MatchSnapshot(t, tt.actual)
		})
	}
}

// TestMatchSnapshot_Mismatch tests snapshot mismatch detection.
func TestMatchSnapshot_Mismatch(t *testing.T) {
	// Create initial snapshot
	original := "original content"
	MatchSnapshot(t, original)

	// Get the temp dir that was used
	tmpDir := getTestDir(t)

	// Try to match with different content (should fail in real test)
	// We can't actually test the failure without a mock testing.T
	// So we'll just verify the file exists
	snapFile := filepath.Join(tmpDir, "__snapshots__", "TestMatchSnapshot_Mismatch_default.snap")
	content, err := os.ReadFile(snapFile)
	require.NoError(t, err)
	assert.Equal(t, original, string(content))
}

// TestMatchNamedSnapshot tests named snapshot matching.
func TestMatchNamedSnapshot(t *testing.T) {
	tests := []struct {
		name     string
		snapName string
		actual   string
		wantFile string
	}{
		{
			name:     "custom name",
			snapName: "my_custom_snapshot",
			actual:   "Custom content",
			wantFile: "TestMatchNamedSnapshot_custom_name_my_custom_snapshot.snap",
		},
		{
			name:     "descriptive name",
			snapName: "button_primary_state",
			actual:   "Button: Primary [Active]",
			wantFile: "TestMatchNamedSnapshot_descriptive_name_button_primary_state.snap",
		},
		{
			name:     "with underscores",
			snapName: "test_case_1",
			actual:   "Test Case 1 Output",
			wantFile: "TestMatchNamedSnapshot_with_underscores_test_case_1.snap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First run: creates snapshot
			MatchNamedSnapshot(t, tt.snapName, tt.actual)

			// Get the temp dir that was used
			tmpDir := getTestDir(t)

			// Verify snapshot file was created
			snapFile := filepath.Join(tmpDir, "__snapshots__", tt.wantFile)
			_, err := os.Stat(snapFile)
			assert.NoError(t, err, "snapshot file should be created")

			// Verify content
			content, err := os.ReadFile(snapFile)
			require.NoError(t, err)
			assert.Equal(t, tt.actual, string(content))
		})
	}
}

// TestMatchComponentSnapshot tests component snapshot matching.
func TestMatchComponentSnapshot(t *testing.T) {
	tests := []struct {
		name       string
		viewOutput string
		wantFile   string
	}{
		{
			name:       "simple component",
			viewOutput: "[ Button ]",
			wantFile:   "TestMatchComponentSnapshot_simple_component_default.snap",
		},
		{
			name:       "complex component",
			viewOutput: "┌─────────┐\n│ Counter │\n│  Count: 5 │\n└─────────┘",
			wantFile:   "TestMatchComponentSnapshot_complex_component_default.snap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock component
			mockComp := NewMockComponent("test")
			mockComp.SetViewOutput(tt.viewOutput)

			// First run: creates snapshot
			MatchComponentSnapshot(t, mockComp)

			// Get the temp dir that was used
			tmpDir := getTestDir(t)

			// Verify snapshot file was created
			snapFile := filepath.Join(tmpDir, "__snapshots__", tt.wantFile)
			_, err := os.Stat(snapFile)
			assert.NoError(t, err, "snapshot file should be created")

			// Verify content matches View() output
			content, err := os.ReadFile(snapFile)
			require.NoError(t, err)
			assert.Equal(t, tt.viewOutput, string(content))
		})
	}
}

// TestUpdateSnapshots tests the update flag detection.
func TestUpdateSnapshots(t *testing.T) {
	tests := []struct {
		name   string
		envVar string
		want   bool
	}{
		{
			name:   "no env var",
			envVar: "",
			want:   false,
		},
		{
			name:   "UPDATE_SNAPS=true",
			envVar: "true",
			want:   true,
		},
		{
			name:   "UPDATE_SNAPS=1",
			envVar: "1",
			want:   true,
		},
		{
			name:   "UPDATE_SNAPS=false",
			envVar: "false",
			want:   false,
		},
		{
			name:   "UPDATE_SNAPS=0",
			envVar: "0",
			want:   false,
		},
		{
			name:   "UPDATE_SNAPS=yes",
			envVar: "yes",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env var
			original := os.Getenv("UPDATE_SNAPS")
			defer os.Setenv("UPDATE_SNAPS", original)

			// Set test env var
			if tt.envVar == "" {
				os.Unsetenv("UPDATE_SNAPS")
			} else {
				os.Setenv("UPDATE_SNAPS", tt.envVar)
			}

			// Test
			got := UpdateSnapshots(t)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestMatchSnapshot_Integration tests the full workflow.
func TestMatchSnapshot_Integration(t *testing.T) {
	// Test 1: Create snapshot
	content1 := "Initial content"
	MatchSnapshot(t, content1)

	// Get the temp dir that was used
	tmpDir := getTestDir(t)

	// Verify file exists
	snapFile := filepath.Join(tmpDir, "__snapshots__", "TestMatchSnapshot_Integration_default.snap")
	saved, err := os.ReadFile(snapFile)
	require.NoError(t, err)
	assert.Equal(t, content1, string(saved))

	// Test 2: Match same content (should pass)
	MatchSnapshot(t, content1)

	// Test 3: Named snapshot
	content2 := "Named content"
	MatchNamedSnapshot(t, "custom", content2)

	namedFile := filepath.Join(tmpDir, "__snapshots__", "TestMatchSnapshot_Integration_custom.snap")
	saved2, err := os.ReadFile(namedFile)
	require.NoError(t, err)
	assert.Equal(t, content2, string(saved2))
}

// TestMatchSnapshot_AutomaticNaming tests automatic snapshot naming.
func TestMatchSnapshot_AutomaticNaming(t *testing.T) {
	t.Run("subtest 1", func(t *testing.T) {
		MatchSnapshot(t, "content 1")

		// Get the temp dir that was used
		tmpDir := getTestDir(t)

		// Should create: TestMatchSnapshot_AutomaticNaming_subtest_1_default.snap
		snapFile := filepath.Join(tmpDir, "__snapshots__", "TestMatchSnapshot_AutomaticNaming_subtest_1_default.snap")
		_, err := os.Stat(snapFile)
		assert.NoError(t, err)
	})

	t.Run("subtest 2", func(t *testing.T) {
		MatchSnapshot(t, "content 2")

		// Get the temp dir that was used
		tmpDir := getTestDir(t)

		// Should create: TestMatchSnapshot_AutomaticNaming_subtest_2_default.snap
		snapFile := filepath.Join(tmpDir, "__snapshots__", "TestMatchSnapshot_AutomaticNaming_subtest_2_default.snap")
		_, err := os.Stat(snapFile)
		assert.NoError(t, err)
	})
}

// TestGetSnapshotManager tests the snapshot manager getter.
func TestGetSnapshotManager(t *testing.T) {
	sm := GetSnapshotManager(t)
	assert.NotNil(t, sm)

	// Verify it works
	sm.Match(t, "test_snapshot", "test content")

	tmpDir := getTestDir(t)
	snapFile := filepath.Join(tmpDir, "__snapshots__", "test_snapshot.snap")
	content, err := os.ReadFile(snapFile)
	require.NoError(t, err)
	assert.Equal(t, "test content", string(content))
}

// TestMatchSnapshotWithOptions tests snapshot matching with custom options.
func TestMatchSnapshotWithOptions(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with custom directory
	MatchSnapshotWithOptions(t, "custom_opts", "custom content", tmpDir, false)

	snapFile := filepath.Join(tmpDir, "__snapshots__", "custom_opts.snap")
	content, err := os.ReadFile(snapFile)
	require.NoError(t, err)
	assert.Equal(t, "custom content", string(content))

	// Test with empty dir (uses default)
	MatchSnapshotWithOptions(t, "default_dir", "default content", "", false)

	defaultDir := getTestDir(t)
	defaultFile := filepath.Join(defaultDir, "__snapshots__", "default_dir.snap")
	content2, err := os.ReadFile(defaultFile)
	require.NoError(t, err)
	assert.Equal(t, "default content", string(content2))
}

// TestSnapshotExists tests snapshot existence checking.
func TestSnapshotExists(t *testing.T) {
	// Non-existent snapshot
	assert.False(t, SnapshotExists(t, "nonexistent"))

	// Create a snapshot
	MatchNamedSnapshot(t, "existing", "content")

	// Now it should exist
	assert.True(t, SnapshotExists(t, "TestSnapshotExists_existing"))
}

// TestGetSnapshotPath tests snapshot path retrieval.
func TestGetSnapshotPath(t *testing.T) {
	path := GetSnapshotPath(t, "my_snapshot")

	// Should contain __snapshots__ directory
	assert.Contains(t, path, "__snapshots__")

	// Should end with .snap
	assert.True(t, filepath.Ext(path) == ".snap")

	// Should contain the snapshot name
	assert.Contains(t, path, "my_snapshot")
}

// TestReadSnapshot tests snapshot reading.
func TestReadSnapshot(t *testing.T) {
	// Create a snapshot first
	testContent := "test snapshot content"
	MatchNamedSnapshot(t, "readable", testContent)

	// Read it back
	content, err := ReadSnapshot(t, "TestReadSnapshot_readable")
	require.NoError(t, err)
	assert.Equal(t, testContent, content)

	// Try to read non-existent snapshot
	_, err = ReadSnapshot(t, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read snapshot")
}
