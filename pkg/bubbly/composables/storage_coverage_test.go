package composables

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
	"github.com/stretchr/testify/assert"
)

// testErrorReporter is a simple error reporter for testing
type testStorageErrorReporter struct {
	onError func(error, *observability.ErrorContext)
}

func (r *testStorageErrorReporter) ReportError(err error, ctx *observability.ErrorContext) {
	if r.onError != nil {
		r.onError(err, ctx)
	}
}

func (r *testStorageErrorReporter) ReportPanic(panicErr *observability.HandlerPanicError, ctx *observability.ErrorContext) {
	// Not needed for these tests
}

func (r *testStorageErrorReporter) Flush(timeout time.Duration) error {
	return nil
}

// TestFileStorage_SaveMkdirError tests Save handling when directory creation fails
func TestFileStorage_SaveMkdirError(t *testing.T) {
	// Arrange - create a file where directory should be (will cause mkdir to fail)
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "blockingfile")
	err := os.WriteFile(filePath, []byte("test"), 0644)
	assert.NoError(t, err)

	reportedError := false
	reporter := &testStorageErrorReporter{
		onError: func(err error, ctx *observability.ErrorContext) {
			reportedError = true
			assert.Equal(t, "FileStorage", ctx.Tags["component"])
			assert.Equal(t, "mkdir_failed", ctx.Tags["operation"])
			assert.Equal(t, "directory_creation", ctx.Tags["error_type"])
		},
	}
	observability.SetErrorReporter(reporter)
	defer observability.SetErrorReporter(nil)

	// Use the file path as a directory (will fail)
	storage := NewFileStorage(filePath)

	// Act
	err = storage.Save("test.json", []byte(`{"test": true}`))

	// Assert
	assert.Error(t, err, "Save should fail when mkdir fails")
	assert.True(t, reportedError, "Error should be reported")
}

// TestFileStorage_SaveWriteError tests Save handling when file write fails
func TestFileStorage_SaveWriteError(t *testing.T) {
	// Arrange - create a read-only directory
	tempDir := t.TempDir()
	readOnlyDir := filepath.Join(tempDir, "readonly")
	err := os.Mkdir(readOnlyDir, 0755)
	assert.NoError(t, err)

	reportedError := false
	reporter := &testStorageErrorReporter{
		onError: func(err error, ctx *observability.ErrorContext) {
			reportedError = true
			assert.Equal(t, "FileStorage", ctx.Tags["component"])
			assert.Equal(t, "save_failed", ctx.Tags["operation"])
			assert.Equal(t, "file_write", ctx.Tags["error_type"])
		},
	}
	observability.SetErrorReporter(reporter)
	defer observability.SetErrorReporter(nil)

	// Make directory read-only
	err = os.Chmod(readOnlyDir, 0444)
	assert.NoError(t, err)
	defer os.Chmod(readOnlyDir, 0755) // Restore for cleanup

	storage := NewFileStorage(readOnlyDir)

	// Act
	err = storage.Save("test.json", []byte(`{"test": true}`))

	// Assert
	assert.Error(t, err, "Save should fail in read-only directory")
	assert.True(t, reportedError, "Error should be reported")
}

// TestFileStorage_ReportErrorWithReporter tests error reporting with configured reporter
func TestFileStorage_ReportErrorWithReporter(t *testing.T) {
	// Arrange
	reportedError := false
	reporter := &testStorageErrorReporter{
		onError: func(err error, ctx *observability.ErrorContext) {
			reportedError = true
			assert.Equal(t, "FileStorage", ctx.ComponentName)
			assert.Equal(t, "test_operation", ctx.EventName)
			assert.Equal(t, "FileStorage", ctx.Tags["component"])
			assert.Equal(t, "test_operation", ctx.Tags["operation"])
			assert.Equal(t, "test_value", ctx.Tags["custom_tag"])
			assert.Equal(t, "test_data", ctx.Extra["custom_extra"])
		},
	}
	observability.SetErrorReporter(reporter)
	defer observability.SetErrorReporter(nil)

	storage := NewFileStorage(t.TempDir())

	// Act
	err := errors.New("test error")
	tags := map[string]string{"custom_tag": "test_value"}
	extra := map[string]interface{}{"custom_extra": "test_data"}
	storage.reportError("test_operation", err, tags, extra)

	// Assert
	assert.True(t, reportedError, "Error should have been reported")
}

// TestFileStorage_ReportErrorWithoutReporter tests error reporting without configured reporter
func TestFileStorage_ReportErrorWithoutReporter(t *testing.T) {
	// Arrange
	observability.SetErrorReporter(nil)
	storage := NewFileStorage(t.TempDir())

	// Act & Assert - should not panic
	err := errors.New("test error")
	tags := map[string]string{}
	extra := map[string]interface{}{}
	storage.reportError("test_operation", err, tags, extra)
}

// TestFileStorage_SaveSuccessAfterMkdir tests successful save after creating directory
func TestFileStorage_SaveSuccessAfterMkdir(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "nested", "path")
	storage := NewFileStorage(nestedDir)

	// Act
	err := storage.Save("test.json", []byte(`{"success": true}`))

	// Assert
	assert.NoError(t, err, "Save should succeed and create nested directories")
	
	// Verify file was created
	data, err := os.ReadFile(filepath.Join(nestedDir, "test.json"))
	assert.NoError(t, err)
	assert.Equal(t, `{"success": true}`, string(data))
}

// TestFileStorage_SaveOverwritesExisting tests that Save overwrites existing files
func TestFileStorage_SaveOverwritesExisting(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	// Save initial data
	err := storage.Save("test.json", []byte(`{"version": 1}`))
	assert.NoError(t, err)

	// Act - overwrite with new data
	err = storage.Save("test.json", []byte(`{"version": 2}`))

	// Assert
	assert.NoError(t, err, "Save should succeed when overwriting")
	
	// Verify new data
	data, err := os.ReadFile(filepath.Join(tempDir, "test.json"))
	assert.NoError(t, err)
	assert.Equal(t, `{"version": 2}`, string(data))
}
