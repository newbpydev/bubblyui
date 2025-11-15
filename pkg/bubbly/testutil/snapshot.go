package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// SnapshotManager handles snapshot testing functionality.
// It provides snapshot creation, comparison, and update capabilities
// with thread-safe operations.
//
// SnapshotManager stores snapshots in a __snapshots__ directory
// relative to the test directory. Each snapshot is saved as a .snap file.
//
// Features:
//   - Snapshot creation: Creates new snapshots on first run
//   - Snapshot comparison: Compares actual output with saved snapshots
//   - Update mode: Updates snapshots when enabled
//   - Thread-safe: Safe for concurrent use
//   - Diff generation: Shows differences when snapshots don't match
//
// Example:
//
//	sm := NewSnapshotManager(t.TempDir(), false)
//	sm.Match(t, "my_snapshot", actualOutput)
type SnapshotManager struct {
	dir    string
	update bool
	mu     sync.Mutex
}

// NewSnapshotManager creates a new snapshot manager.
// The testDir parameter specifies the directory where snapshots will be stored
// in a __snapshots__ subdirectory. The update parameter controls whether
// snapshots should be updated when mismatches occur.
//
// Parameters:
//   - testDir: Directory where __snapshots__ folder will be created
//   - update: If true, updates snapshots instead of failing on mismatch
//
// Returns:
//   - *SnapshotManager: A new snapshot manager instance
//
// Example:
//
//	// Create snapshot manager in test directory
//	sm := NewSnapshotManager(t.TempDir(), false)
//
//	// Create with update mode enabled
//	sm := NewSnapshotManager(t.TempDir(), true)
func NewSnapshotManager(testDir string, update bool) *SnapshotManager {
	snapDir := filepath.Join(testDir, "__snapshots__")
	return &SnapshotManager{
		dir:    snapDir,
		update: update,
	}
}

// Match compares the actual content with a saved snapshot.
// If the snapshot doesn't exist, it creates it. If it exists and matches,
// the test passes. If it doesn't match and update mode is off, the test fails
// with a diff. If update mode is on, the snapshot is updated.
//
// Parameters:
//   - t: The testing interface
//   - name: The name of the snapshot (without extension)
//   - actual: The actual content to compare
//
// Behavior:
//   - First run: Creates snapshot file
//   - Subsequent runs: Compares with saved snapshot
//   - Update mode: Overwrites snapshot with new content
//   - Mismatch: Fails test with diff (unless update mode)
//
// Example:
//
//	sm := NewSnapshotManager(t.TempDir(), false)
//	output := component.View()
//	sm.Match(t, "component_output", output)
func (sm *SnapshotManager) Match(t testingT, name, actual string) {
	t.Helper()

	sm.mu.Lock()
	defer sm.mu.Unlock()

	filename := sm.getSnapshotFile(name)

	// Read existing snapshot
	expected, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// No snapshot exists, create it
			sm.createSnapshot(t, filename, actual)
			return
		}
		t.Errorf("failed to read snapshot: %v", err)
		return
	}

	// Compare
	if string(expected) != actual {
		if sm.update {
			// Update mode: overwrite snapshot
			sm.updateSnapshot(t, filename, actual)
		} else {
			// Show diff
			diff := sm.generateDiff(string(expected), actual)
			t.Errorf("Snapshot mismatch for %q:\n%s\n\nRun with -update flag to update snapshots",
				name, diff)
		}
	}
}

// getSnapshotFile returns the full path to a snapshot file.
func (sm *SnapshotManager) getSnapshotFile(name string) string {
	return filepath.Join(sm.dir, name+".snap")
}

// createSnapshot creates a new snapshot file.
func (sm *SnapshotManager) createSnapshot(t testingT, filename, content string) {
	// Create directory if it doesn't exist
	err := os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		t.Errorf("failed to create snapshot dir: %v", err)
		return
	}

	// Write snapshot file
	err = os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		t.Errorf("failed to write snapshot: %v", err)
		return
	}

	t.Logf("Created snapshot: %s", filename)
}

// updateSnapshot updates an existing snapshot file.
func (sm *SnapshotManager) updateSnapshot(t testingT, filename, content string) {
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		t.Errorf("failed to update snapshot: %v", err)
		return
	}

	t.Logf("Updated snapshot: %s", filename)
}

// generateDiff generates a simple diff between expected and actual content.
// This is a basic implementation that shows line-by-line differences.
func (sm *SnapshotManager) generateDiff(expected, actual string) string {
	var diff strings.Builder

	diff.WriteString("Expected:\n")
	diff.WriteString(formatContent(expected))
	diff.WriteString("\n\nActual:\n")
	diff.WriteString(formatContent(actual))

	// Add line-by-line comparison
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	diff.WriteString("\n\nDifferences:\n")

	maxLines := len(expectedLines)
	if len(actualLines) > maxLines {
		maxLines = len(actualLines)
	}

	for i := 0; i < maxLines; i++ {
		var expLine, actLine string

		if i < len(expectedLines) {
			expLine = expectedLines[i]
		}
		if i < len(actualLines) {
			actLine = actualLines[i]
		}

		if expLine != actLine {
			if expLine != "" {
				diff.WriteString(fmt.Sprintf("- Line %d: %s\n", i+1, expLine))
			}
			if actLine != "" {
				diff.WriteString(fmt.Sprintf("+ Line %d: %s\n", i+1, actLine))
			}
		}
	}

	return diff.String()
}

// formatContent formats content for display in diffs.
// Adds indentation and handles empty content.
func formatContent(content string) string {
	if content == "" {
		return "  (empty)"
	}

	lines := strings.Split(content, "\n")
	var formatted strings.Builder

	for _, line := range lines {
		formatted.WriteString("  ")
		formatted.WriteString(line)
		formatted.WriteString("\n")
	}

	return strings.TrimRight(formatted.String(), "\n")
}
