package composables

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// TestUseLocalStorage_LoadsFromStorage tests that values are loaded from storage on mount
func TestUseLocalStorage_LoadsFromStorage(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)
	key := "test-counter"

	// Pre-populate storage with value
	data, err := json.Marshal(42)
	require.NoError(t, err)
	err = storage.Save(key, data)
	require.NoError(t, err)

	// Create component context
	ctx := createTestContext()

	// Act
	state := UseLocalStorage(ctx, key, 0, storage)

	// Assert
	assert.Equal(t, 42, state.Get(), "should load value from storage")
}

// TestUseLocalStorage_SavesOnChange tests that values are saved when changed
func TestUseLocalStorage_SavesOnChange(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)
	key := "test-name"

	ctx := createTestContext()

	// Act
	state := UseLocalStorage(ctx, key, "Alice", storage)
	state.Set("Bob")

	// Give watch callback time to execute
	time.Sleep(50 * time.Millisecond)

	// Assert - verify file was written
	data, err := storage.Load(key)
	require.NoError(t, err)

	var loaded string
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)
	assert.Equal(t, "Bob", loaded, "should save new value to storage")
}

// TestUseLocalStorage_JSONSerialization tests JSON serialization of complex types
func TestUseLocalStorage_JSONSerialization(t *testing.T) {
	type User struct {
		Name  string
		Age   int
		Admin bool
	}

	tests := []struct {
		name    string
		initial interface{}
		updated interface{}
	}{
		{
			name:    "struct",
			initial: User{Name: "Alice", Age: 30, Admin: false},
			updated: User{Name: "Bob", Age: 35, Admin: true},
		},
		{
			name:    "slice",
			initial: []int{1, 2, 3},
			updated: []int{4, 5, 6, 7},
		},
		{
			name:    "map",
			initial: map[string]int{"a": 1, "b": 2},
			updated: map[string]int{"c": 3, "d": 4, "e": 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tempDir := t.TempDir()
			storage := NewFileStorage(tempDir)
			key := "test-" + tt.name

			ctx := createTestContext()

			// Act - create with initial value
			state := UseLocalStorage(ctx, key, tt.initial, storage)
			assert.Equal(t, tt.initial, state.Get())

			// Update value
			state.Value.Set(tt.updated)
			time.Sleep(50 * time.Millisecond)

			// Assert - load from storage and verify
			data, err := storage.Load(key)
			require.NoError(t, err)

			// Unmarshal into same type
			switch tt.initial.(type) {
			case User:
				var loaded User
				err = json.Unmarshal(data, &loaded)
				require.NoError(t, err)
				assert.Equal(t, tt.updated, loaded)
			case []int:
				var loaded []int
				err = json.Unmarshal(data, &loaded)
				require.NoError(t, err)
				assert.Equal(t, tt.updated, loaded)
			case map[string]int:
				var loaded map[string]int
				err = json.Unmarshal(data, &loaded)
				require.NoError(t, err)
				assert.Equal(t, tt.updated, loaded)
			}
		})
	}
}

// TestUseLocalStorage_Deserialization tests that loaded values match written values
func TestUseLocalStorage_Deserialization(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)
	key := "test-roundtrip"

	ctx := createTestContext()

	// Act - write value
	state1 := UseLocalStorage(ctx, key, 100, storage)
	state1.Set(200)
	time.Sleep(50 * time.Millisecond)

	// Create new instance - should load saved value
	ctx2 := createTestContext()
	state2 := UseLocalStorage(ctx2, key, 0, storage)

	// Assert
	assert.Equal(t, 200, state2.Get(), "should load previously saved value")
}

// TestUseLocalStorage_StorageUnavailable tests graceful handling when storage is unavailable
func TestUseLocalStorage_StorageUnavailable(t *testing.T) {
	// Arrange - use read-only directory
	tempDir := t.TempDir()
	err := os.Chmod(tempDir, 0444) // Read-only
	require.NoError(t, err)
	defer func() {
		_ = os.Chmod(tempDir, 0755) // Restore for cleanup
	}()

	storage := NewFileStorage(tempDir)
	key := "test-readonly"

	ctx := createTestContext()

	// Act - should not panic, use initial value
	state := UseLocalStorage(ctx, key, "default", storage)

	// Assert
	assert.Equal(t, "default", state.Get(), "should use initial value when storage unavailable")

	// Setting value should not panic (save will fail gracefully)
	assert.NotPanics(t, func() {
		state.Set("new value")
		time.Sleep(50 * time.Millisecond)
	})
}

// TestUseLocalStorage_TypeSafety tests that different types work independently
func TestUseLocalStorage_TypeSafety(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	ctx := createTestContext()

	// Act - create different typed states
	intState := UseLocalStorage(ctx, "int-key", 0, storage)
	stringState := UseLocalStorage(ctx, "string-key", "", storage)
	boolState := UseLocalStorage(ctx, "bool-key", false, storage)

	intState.Set(42)
	stringState.Set("hello")
	boolState.Set(true)

	// Assert
	assert.Equal(t, 42, intState.Get())
	assert.Equal(t, "hello", stringState.Get())
	assert.Equal(t, true, boolState.Get())
}

// TestUseLocalStorage_InitialValueWhenNoStorage tests using initial value when no storage exists
func TestUseLocalStorage_InitialValueWhenNoStorage(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)
	key := "nonexistent-key"

	ctx := createTestContext()

	// Act
	state := UseLocalStorage(ctx, key, "initial", storage)

	// Assert
	assert.Equal(t, "initial", state.Get(), "should use initial value when no storage exists")
}

// TestUseLocalStorage_InvalidJSON tests handling of invalid JSON in storage
func TestUseLocalStorage_InvalidJSON(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)
	key := "invalid-json"

	// Write invalid JSON to storage
	err := storage.Save(key, []byte("not valid json {{{"))
	require.NoError(t, err)

	ctx := createTestContext()

	// Act - should not panic, use initial value
	state := UseLocalStorage(ctx, key, 999, storage)

	// Assert
	assert.Equal(t, 999, state.Get(), "should use initial value when JSON is invalid")
}

// TestUseLocalStorage_MultipleInstances tests that multiple instances are independent
func TestUseLocalStorage_MultipleInstances(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	ctx1 := createTestContext()
	ctx2 := createTestContext()

	// Act - different keys
	state1 := UseLocalStorage(ctx1, "key1", 10, storage)
	state2 := UseLocalStorage(ctx2, "key2", 20, storage)

	state1.Set(100)
	state2.Set(200)

	// Assert
	assert.Equal(t, 100, state1.Get())
	assert.Equal(t, 200, state2.Get())
}

// mockFailingSaveStorage is a storage that fails on Save operations
type mockFailingSaveStorage struct {
	loadData []byte
	loadErr  error
	saveErr  error
}

func (m *mockFailingSaveStorage) Load(_ string) ([]byte, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	return m.loadData, nil
}

func (m *mockFailingSaveStorage) Save(_ string, _ []byte) error {
	return m.saveErr
}

// TestUseLocalStorage_SaveError_ReportsError tests that save errors are reported
func TestUseLocalStorage_SaveError_ReportsError(t *testing.T) {
	// Arrange - storage that fails on save
	storage := &mockFailingSaveStorage{
		loadErr: os.ErrNotExist,
		saveErr: os.ErrPermission,
	}

	ctx := createTestContext()

	// Act - should not panic
	state := UseLocalStorage(ctx, "test-key", "initial", storage)

	// Change value - save should fail but not panic
	assert.NotPanics(t, func() {
		state.Set("new value")
		// Give watch callback time to execute
		time.Sleep(50 * time.Millisecond)
	})

	// Value should still be updated locally
	assert.Equal(t, "new value", state.Get())
}

// TestUseLocalStorage_LoadError_NotFileNotExist tests error reporting for non-NotExist errors
func TestUseLocalStorage_LoadError_NotFileNotExist(t *testing.T) {
	// Arrange - storage that returns permission error on load
	storage := &mockFailingSaveStorage{
		loadErr: os.ErrPermission, // Not ErrNotExist
	}

	ctx := createTestContext()

	// Act - should not panic and use initial value
	state := UseLocalStorage(ctx, "test-key", "initial", storage)

	// Assert - should use initial value
	assert.Equal(t, "initial", state.Get())
}

// mockUnmarshalableValue is a type that fails JSON marshal
type mockUnmarshalableValue struct {
	Ch chan int // channels cannot be marshaled to JSON
}

// TestUseLocalStorage_MarshalError_ReportsError tests that marshal errors are reported
func TestUseLocalStorage_MarshalError_ReportsError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	ctx := createTestContext()

	// Create with a marshalable initial value
	state := UseLocalStorage(ctx, "test-key", mockUnmarshalableValue{}, storage)

	// Act - try to save value with unmarshable type
	assert.NotPanics(t, func() {
		state.Set(mockUnmarshalableValue{Ch: make(chan int)})
		// Give watch callback time to execute
		time.Sleep(50 * time.Millisecond)
	})
}

// TestUseLocalStorage_SaveError_WithErrorReporter tests save error with reporter
func TestUseLocalStorage_SaveError_WithErrorReporter(t *testing.T) {
	// Setup custom error reporter
	var capturedError error
	var capturedContext *observability.ErrorContext

	customReporter := &testStorageErrorReporter{
		onError: func(err error, ctx *observability.ErrorContext) {
			capturedError = err
			capturedContext = ctx
		},
	}

	observability.SetErrorReporter(customReporter)
	defer observability.SetErrorReporter(nil)

	// Arrange - storage that fails on save
	storage := &mockFailingSaveStorage{
		loadErr: os.ErrNotExist,
		saveErr: os.ErrPermission,
	}

	ctx := createTestContext()

	// Act
	state := UseLocalStorage(ctx, "test-key", "initial", storage)
	state.Set("new value")

	// Give watch callback time to execute
	time.Sleep(100 * time.Millisecond)

	// Assert - error should be reported
	assert.NotNil(t, capturedError, "Error should be reported for save failure")
	assert.NotNil(t, capturedContext, "Error context should be provided")
	assert.Equal(t, "UseLocalStorage", capturedContext.ComponentName)
	assert.Equal(t, "save_failed", capturedContext.EventName)
	assert.Equal(t, "storage_save", capturedContext.Tags["error_type"])
}

// TestUseLocalStorage_MarshalError_WithErrorReporter tests marshal error with reporter
func TestUseLocalStorage_MarshalError_WithErrorReporter(t *testing.T) {
	// Setup custom error reporter
	var capturedError error
	var capturedContext *observability.ErrorContext

	customReporter := &testStorageErrorReporter{
		onError: func(err error, ctx *observability.ErrorContext) {
			capturedError = err
			capturedContext = ctx
		},
	}

	observability.SetErrorReporter(customReporter)
	defer observability.SetErrorReporter(nil)

	// Arrange
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	ctx := createTestContext()

	// Act - create with unmarshable type
	state := UseLocalStorage(ctx, "test-key", mockUnmarshalableValue{}, storage)
	state.Set(mockUnmarshalableValue{Ch: make(chan int)})

	// Give watch callback time to execute
	time.Sleep(100 * time.Millisecond)

	// Assert - error should be reported
	assert.NotNil(t, capturedError, "Error should be reported for marshal failure")
	assert.NotNil(t, capturedContext, "Error context should be provided")
	assert.Equal(t, "UseLocalStorage", capturedContext.ComponentName)
	assert.Equal(t, "marshal_failed", capturedContext.EventName)
	assert.Equal(t, "json_marshal", capturedContext.Tags["error_type"])
}

// TestUseLocalStorage_UnmarshalError_WithErrorReporter tests unmarshal error with reporter
func TestUseLocalStorage_UnmarshalError_WithErrorReporter(t *testing.T) {
	// Setup custom error reporter
	var capturedError error
	var capturedContext *observability.ErrorContext

	customReporter := &testStorageErrorReporter{
		onError: func(err error, ctx *observability.ErrorContext) {
			capturedError = err
			capturedContext = ctx
		},
	}

	observability.SetErrorReporter(customReporter)
	defer observability.SetErrorReporter(nil)

	// Arrange - storage with invalid JSON
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)
	key := "bad-json"

	// Pre-populate with invalid JSON
	err := storage.Save(key, []byte("{invalid json}"))
	require.NoError(t, err)

	ctx := createTestContext()

	// Act
	state := UseLocalStorage(ctx, key, "default", storage)

	// Assert - error should be reported
	assert.NotNil(t, capturedError, "Error should be reported for unmarshal failure")
	assert.NotNil(t, capturedContext, "Error context should be provided")
	assert.Equal(t, "UseLocalStorage", capturedContext.ComponentName)
	assert.Equal(t, "unmarshal_failed", capturedContext.EventName)
	assert.Equal(t, "json_unmarshal", capturedContext.Tags["error_type"])

	// Should use default value
	assert.Equal(t, "default", state.Get())
}

// TestUseLocalStorage_LoadError_WithErrorReporter tests load error with reporter
func TestUseLocalStorage_LoadError_WithErrorReporter(t *testing.T) {
	// Setup custom error reporter
	var capturedError error
	var capturedContext *observability.ErrorContext

	customReporter := &testStorageErrorReporter{
		onError: func(err error, ctx *observability.ErrorContext) {
			capturedError = err
			capturedContext = ctx
		},
	}

	observability.SetErrorReporter(customReporter)
	defer observability.SetErrorReporter(nil)

	// Arrange - storage that returns error on load
	storage := &mockFailingSaveStorage{
		loadErr: os.ErrPermission, // Not ErrNotExist, so it should be reported
	}

	ctx := createTestContext()

	// Act
	state := UseLocalStorage(ctx, "test-key", "default", storage)

	// Assert - error should be reported
	assert.NotNil(t, capturedError, "Error should be reported for load failure")
	assert.NotNil(t, capturedContext, "Error context should be provided")
	assert.Equal(t, "UseLocalStorage", capturedContext.ComponentName)
	assert.Equal(t, "load_failed", capturedContext.EventName)
	assert.Equal(t, "storage_load", capturedContext.Tags["error_type"])

	// Should use default value
	assert.Equal(t, "default", state.Get())
}

// testStorageErrorReporter is defined in storage_coverage_test.go
