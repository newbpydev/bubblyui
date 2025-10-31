package composables

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// UseLocalStorage creates a reactive state that persists to storage.
// Values are automatically loaded from storage on creation and saved on every change.
//
// This composable provides the same API as UseState but with automatic persistence.
// It uses JSON serialization, so the type T must be JSON-serializable.
//
// Storage behavior:
//   - On creation: Attempts to load value from storage. If successful, uses loaded value.
//     If storage doesn't exist or load fails, uses initial value.
//   - On change: Automatically saves new value to storage using JSON serialization.
//   - Errors are reported via observability system but don't crash the application.
//
// Parameters:
//   - ctx: The component context (required for all composables)
//   - key: The storage key (used as filename in FileStorage)
//   - initial: The initial/default value if storage doesn't exist
//   - storage: The storage implementation (use NewFileStorage for file-based storage)
//
// Returns:
//   - UseStateReturn[T]: Same interface as UseState, with automatic persistence
//
// Type Safety:
//   - T must be JSON-serializable (structs, primitives, slices, maps)
//   - Compile-time type checking ensures type safety
//
// Example - Simple Value:
//
//	Setup(func(ctx *Context) {
//	    storage := NewFileStorage("/home/user/.config/myapp")
//	    count := UseLocalStorage(ctx, "counter", 0, storage)
//
//	    ctx.On("increment", func(_ interface{}) {
//	        count.Set(count.Get() + 1)  // Automatically saved
//	    })
//
//	    ctx.Expose("count", count.Value)
//	})
//
// Example - Struct:
//
//	type Settings struct {
//	    Theme    string
//	    FontSize int
//	}
//
//	Setup(func(ctx *Context) {
//	    storage := NewFileStorage("/home/user/.config/myapp")
//	    settings := UseLocalStorage(ctx, "settings", Settings{
//	        Theme:    "dark",
//	        FontSize: 14,
//	    }, storage)
//
//	    ctx.On("changeTheme", func(data interface{}) {
//	        s := settings.Get()
//	        s.Theme = data.(string)
//	        settings.Set(s)  // Automatically saved to storage
//	    })
//	})
//
// Example - With Default Storage:
//
//	// Create a global storage instance
//	var appStorage = NewFileStorage(os.ExpandEnv("$HOME/.config/myapp"))
//
//	Setup(func(ctx *Context) {
//	    // Reuse storage across components
//	    username := UseLocalStorage(ctx, "username", "", appStorage)
//	    preferences := UseLocalStorage(ctx, "preferences", defaultPrefs, appStorage)
//	})
//
// Error Handling:
//   - Load errors: Uses initial value, reports error via observability
//   - Save errors: Continues execution, reports error via observability
//   - JSON errors: Uses initial value, reports error via observability
//   - Storage unavailable: Uses initial value, reports error via observability
//
// Performance:
//   - Load: One file read on creation
//   - Save: One file write per value change (not debounced)
//   - Consider using UseDebounce if values change frequently
func UseLocalStorage[T any](ctx *bubbly.Context, key string, initial T, storage Storage) UseStateReturn[T] {
	// Try to load existing value from storage
	loadedValue := initial
	data, err := storage.Load(key)

	if err == nil {
		// Storage exists, try to unmarshal
		var loaded T
		if err := json.Unmarshal(data, &loaded); err != nil {
			// JSON unmarshal failed - use initial value and report error
			reportStorageError("unmarshal_failed", err, map[string]string{
				"error_type": "json_unmarshal",
				"key":        key,
			}, map[string]interface{}{
				"data_sample": truncateData(data, 100),
				"data_size":   len(data),
			})
		} else {
			// Successfully loaded
			loadedValue = loaded
		}
	} else if !os.IsNotExist(err) {
		// Load failed for reason other than "not found" - report error
		reportStorageError("load_failed", err, map[string]string{
			"error_type": "storage_load",
			"key":        key,
		}, map[string]interface{}{})
	}
	// If err is os.ErrNotExist, that's expected - just use initial value

	// Create the underlying reactive reference with loaded/initial value
	value := bubbly.NewRef(loadedValue)

	// Watch for changes and save to storage
	// Create a watcher that monitors the value
	bubbly.Watch(value, func(newVal, _ T) {
		// Marshal to JSON
		data, err := json.Marshal(newVal)
		if err != nil {
			reportStorageError("marshal_failed", err, map[string]string{
				"error_type": "json_marshal",
				"key":        key,
			}, map[string]interface{}{
				"value_type": getTypeName(newVal),
			})
			return
		}

		// Save to storage
		if err := storage.Save(key, data); err != nil {
			reportStorageError("save_failed", err, map[string]string{
				"error_type": "storage_save",
				"key":        key,
			}, map[string]interface{}{
				"data_size": len(data),
			})
			return
		}
	})

	// Return the same interface as UseState
	return UseStateReturn[T]{
		Value: value,
		Set: func(v T) {
			value.Set(v)
		},
		Get: func() T {
			return value.GetTyped()
		},
	}
}

// reportStorageError reports storage-related errors to the observability system.
// Follows ZERO TOLERANCE policy - never silent failures.
func reportStorageError(operation string, err error, tags map[string]string, extra map[string]interface{}) {
	reporter := observability.GetErrorReporter()
	if reporter == nil {
		return
	}

	// Add common tags
	tags["component"] = "UseLocalStorage"
	tags["operation"] = operation

	// Add common extra data
	extra["error_message"] = err.Error()

	ctx := &observability.ErrorContext{
		ComponentName: "UseLocalStorage",
		ComponentID:   "composable",
		EventName:     operation,
		Timestamp:     time.Now(),
		StackTrace:    debug.Stack(),
		Tags:          tags,
		Extra:         extra,
	}

	reporter.ReportError(err, ctx)
}

// truncateData truncates byte slice to maxLen for error reporting
func truncateData(data []byte, maxLen int) string {
	if len(data) <= maxLen {
		return string(data)
	}
	return string(data[:maxLen]) + "..."
}

// getTypeName returns a string representation of the type
func getTypeName(v interface{}) string {
	return fmt.Sprintf("%T", v)
}
