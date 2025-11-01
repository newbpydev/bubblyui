package composables

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	defer os.Chmod(tempDir, 0755) // Restore for cleanup

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
