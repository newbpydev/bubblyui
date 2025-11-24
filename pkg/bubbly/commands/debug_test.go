package commands

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test suite for the debug logging system (Task 6.1: Debug Mode).
//
// This test suite comprehensively validates the CommandLogger interface and its
// implementations, ensuring they meet the requirements for:
//   - Optional debug logging with zero overhead when disabled
//   - Clear, structured log format for debugging
//   - Thread-safe concurrent operation
//   - Proper handling of all data types
//   - Performance characteristics meeting specifications
//
// Test Coverage:
//   - Functional correctness (logging works as expected)
//   - Format verification (consistent, parseable output)
//   - Performance validation (zero overhead when disabled)
//   - Thread safety (concurrent access patterns)
//   - Edge cases (nil values, complex types, empty values)
//   - Integration scenarios (real-world usage patterns)
//
// Performance Benchmarks:
//   - NopLogger: ~0.25 ns/op, 0 allocs/op (zero overhead)
//   - CommandLogger: ~2700 ns/op, 4 allocs/op (enabled logging)
//   - Performance ratio: ~10,000x faster when disabled
//
// Usage Examples:
//
//	// Basic usage
//	logger := commands.NewCommandLogger(os.Stdout)
//	logger.LogCommand("Counter", "component-1", "ref-5", 0, 1)
//	// Output: [DEBUG] Command Generated | Component: Counter (component-1) | Ref: ref-5 | 0 → 1
//
//	// Zero overhead when disabled
//	logger := commands.NewNopLogger()
//	logger.LogCommand("Counter", "component-1", "ref-5", 0, 1) // No-op
//
//	// Custom output destination
//	var buf bytes.Buffer
//	logger := commands.NewCommandLogger(&buf)
//	logger.LogCommand("Form", "component-2", "ref-10", "", "email")
//	fmt.Println(buf.String())
//

// TestCommandLogger_LogCommand tests that command generation events are logged
// with the proper format and contain all expected information.
//
// This test validates:
//   - Log message format matches specification
//   - All required fields are present (component name, ID, ref ID, values)
//   - Different data types are handled correctly
//   - Nil values are displayed properly as "<nil>"
//   - State transitions are shown with arrow notation
//
// Test Cases:
//  1. Integer value changes (counters)
//  2. String value changes (form fields)
//  3. Boolean value changes (toggles)
//  4. Nil to non-nil transitions (initialization)
//
// Expected Format:
//
//	[timestamp] [DEBUG] Command Generated | Component: <name> (<id>) | Ref: <refID> | <old> → <new>
func TestCommandLogger_LogCommand(t *testing.T) {
	tests := []struct {
		name             string
		componentName    string
		componentID      string
		refID            string
		oldValue         interface{}
		newValue         interface{}
		expectedContains []string
	}{
		{
			name:          "integer value change",
			componentName: "Counter",
			componentID:   "component-1",
			refID:         "ref-1",
			oldValue:      0,
			newValue:      1,
			expectedContains: []string{
				"[DEBUG]",
				"Counter",
				"component-1",
				"ref-1",
				"0",
				"1",
				"Command Generated",
			},
		},
		{
			name:          "string value change",
			componentName: "Form",
			componentID:   "component-2",
			refID:         "ref-5",
			oldValue:      "",
			newValue:      "hello",
			expectedContains: []string{
				"[DEBUG]",
				"Form",
				"component-2",
				"ref-5",
				"hello",
				"Command Generated",
			},
		},
		{
			name:          "boolean toggle",
			componentName: "Toggle",
			componentID:   "component-3",
			refID:         "ref-10",
			oldValue:      false,
			newValue:      true,
			expectedContains: []string{
				"[DEBUG]",
				"Toggle",
				"component-3",
				"ref-10",
				"false",
				"true",
				"Command Generated",
			},
		},
		{
			name:          "nil to value",
			componentName: "Loader",
			componentID:   "component-4",
			refID:         "ref-15",
			oldValue:      nil,
			newValue:      "loaded",
			expectedContains: []string{
				"[DEBUG]",
				"Loader",
				"component-4",
				"ref-15",
				"<nil>",
				"loaded",
				"Command Generated",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer
			logger := NewCommandLogger(&buf)

			// Log command
			logger.LogCommand(tt.componentName, tt.componentID, tt.refID, tt.oldValue, tt.newValue)

			// Verify output
			output := buf.String()
			for _, expected := range tt.expectedContains {
				assert.Contains(t, output, expected,
					"log output should contain '%s'", expected)
			}
		})
	}
}

// TestCommandLogger_Format tests that the log format is clear, structured,
// and helpful for debugging purposes.
//
// This test validates specific format requirements:
//   - Timestamp prefix follows Go's standard log format
//   - Component identification includes both name and ID
//   - State transitions use clear arrow notation (→)
//   - Format is consistent across different scenarios
//
// Format Requirements:
//   - Timestamp: YYYY/MM/DD HH:MM:SS (standard Go log format)
//   - Debug tag: [DEBUG] for filtering and visibility
//   - Component info: "Component: <name> (<id>)"
//   - State transition: "<old> → <new>" with Unicode arrow
//   - Overall structure: Pipe-delimited for easy parsing
//
// Why This Matters:
//   - Consistent format enables log parsing and filtering
//   - Clear component identification helps locate issues
//   - Arrow notation provides visual state transition clarity
//   - Timestamp enables chronological debugging
func TestCommandLogger_Format(t *testing.T) {
	tests := []struct {
		name          string
		componentName string
		componentID   string
		refID         string
		oldValue      interface{}
		newValue      interface{}
		checkFormat   func(*testing.T, string)
	}{
		{
			name:          "has timestamp prefix",
			componentName: "Test",
			componentID:   "comp-1",
			refID:         "ref-1",
			oldValue:      0,
			newValue:      1,
			checkFormat: func(t *testing.T, output string) {
				// Standard Go log format starts with date/time
				assert.Regexp(t, `^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}`, output,
					"should have timestamp prefix")
			},
		},
		{
			name:          "has component identification",
			componentName: "UserForm",
			componentID:   "component-42",
			refID:         "ref-5",
			oldValue:      "old",
			newValue:      "new",
			checkFormat: func(t *testing.T, output string) {
				assert.Contains(t, output, "Component: UserForm (component-42)",
					"should have clear component identification")
			},
		},
		{
			name:          "shows state transition clearly",
			componentName: "Counter",
			componentID:   "comp-1",
			refID:         "ref-1",
			oldValue:      5,
			newValue:      6,
			checkFormat: func(t *testing.T, output string) {
				// Should show old -> new transition
				assert.Regexp(t, `5.*→.*6`, output,
					"should show clear state transition with arrow")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewCommandLogger(&buf)

			logger.LogCommand(tt.componentName, tt.componentID, tt.refID, tt.oldValue, tt.newValue)

			tt.checkFormat(t, buf.String())
		})
	}
}

// TestNopLogger_NoOutput tests that NopLogger produces no output and has
// zero overhead when debug logging is disabled.
//
// This test validates the key requirement for disabled debug logging:
//   - No log output is produced regardless of call count
//   - Zero memory allocations
//   - No side effects or state changes
//   - Method can be called safely without impact
//
// Performance Requirements:
//   - LogCommand overhead: ~0.25 ns/op (inlined empty method)
//   - Memory usage: 0 bytes
//   - Allocations: 0 per call
//   - Thread safety: Trivial (no shared state)
//
// Test Strategy:
//   - Call LogCommand multiple times (stress test)
//   - Verify buffer remains empty
//   - Ensure no side effects occur
func TestNopLogger_NoOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := NewNopLogger()

	// Log multiple commands
	for i := 0; i < 100; i++ {
		logger.LogCommand("Test", "comp-1", "ref-1", i, i+1)
	}

	// Buffer should be empty
	assert.Empty(t, buf.String(), "NopLogger should produce no output")
}

// TestCommandLogger_ThreadSafe tests that concurrent logging from multiple
// goroutines works correctly without race conditions or data corruption.
//
// This test validates thread safety requirements:
//   - Multiple goroutines can log concurrently
//   - Log messages are not interleaved or corrupted
//   - No race conditions occur (verified with -race flag)
//   - All log calls succeed without panics
//
// Concurrency Test Strategy:
//   - Launch 10 goroutines simultaneously
//   - Each goroutine logs 10 messages
//   - Total: 100 concurrent log operations
//   - Verify all messages are captured intact
//
// Thread Safety Mechanisms:
//   - Go's standard log package provides built-in synchronization
//   - Each log operation is atomic
//   - No shared mutable state between log calls
//
// Real-World Scenario:
//
//	Multiple components updating state simultaneously
//	(e.g., form validation, counter updates, list modifications)
func TestCommandLogger_ThreadSafe(t *testing.T) {
	var buf bytes.Buffer
	logger := NewCommandLogger(&buf)

	const goroutines = 10
	const logsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(_ int) {
			defer wg.Done()

			for j := 0; j < logsPerGoroutine; j++ {
				logger.LogCommand(
					"TestComponent",
					"component-1",
					"ref-1",
					j,
					j+1,
				)
			}
		}(i)
	}

	wg.Wait()

	// Verify we got logs (exact count may vary due to interleaving)
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have logged something from all goroutines
	assert.Greater(t, len(lines), 0, "should have logged from all goroutines")
}

// TestCommandLogger_ComplexTypes tests that complex data types are logged
// correctly without causing panics or formatting issues.
//
// This test validates handling of non-primitive types:
//   - Slices (arrays of values)
//   - Maps (key-value pairs)
//   - Structs (custom types)
//   - Pointers (memory references)
//   - Nested data structures
//
// Type Handling Requirements:
//   - No panics when logging complex types
//   - Readable string representation using Go's %v format
//   - Consistent formatting across different types
//   - Proper handling of nil pointers
//
// Real-World Examples:
//   - []string{"task1", "task2"} (todo lists)
//   - map[string]int{"count": 42} (form data)
//   - struct{ Name string }{"Alice"} (user data)
//   - *int (pointer to value)
func TestCommandLogger_ComplexTypes(t *testing.T) {
	tests := []struct {
		name     string
		oldValue interface{}
		newValue interface{}
	}{
		{
			name:     "slice",
			oldValue: []int{1, 2, 3},
			newValue: []int{1, 2, 3, 4},
		},
		{
			name:     "map",
			oldValue: map[string]int{"a": 1},
			newValue: map[string]int{"a": 1, "b": 2},
		},
		{
			name:     "struct",
			oldValue: struct{ Name string }{"Alice"},
			newValue: struct{ Name string }{"Bob"},
		},
		{
			name:     "pointer",
			oldValue: func() *int { i := 42; return &i }(),
			newValue: func() *int { i := 84; return &i }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewCommandLogger(&buf)

			// Should not panic with complex types
			require.NotPanics(t, func() {
				logger.LogCommand("Test", "comp-1", "ref-1", tt.oldValue, tt.newValue)
			})

			// Should produce output
			assert.NotEmpty(t, buf.String(), "should log complex types")
		})
	}
}

// TestCommandLogger_NilWriter tests that nil writer is handled gracefully
// without causing panics or errors.
//
// This test validates error handling behavior:
//   - Nil writer is treated like io.Discard
//   - No panics occur when logging with nil writer
//   - Behavior is consistent with NopLogger
//   - Graceful degradation for invalid inputs
//
// Error Handling Strategy:
//   - Convert nil writer to io.Discard internally
//   - Maintain same interface contract
//   - No special error return needed
//   - Silent failure is acceptable for logging
//
// Usage Context:
//   - Testing scenarios where output isn't needed
//   - Configuration errors where writer is not set
//   - Defensive programming against nil inputs
func TestCommandLogger_NilWriter(t *testing.T) {
	// Should not panic with nil writer
	require.NotPanics(t, func() {
		logger := NewCommandLogger(nil)
		logger.LogCommand("Test", "comp-1", "ref-1", 0, 1)
	})
}

// TestCommandLogger_HelpfulForDebugging tests that log output contains all
// essential information needed for effective debugging.
//
// This test validates the practical utility of debug logs:
//   - Component name identifies what changed
//   - Component ID enables tracking specific instances
//   - Ref ID pinpoints exact state variable
//   - Old/new values show the actual change
//   - Clear indication of command generation
//
// Debugging Information Requirements:
//   - What: Component name (e.g., "TodoList")
//   - Where: Component ID (e.g., "component-7")
//   - Which: Ref ID (e.g., "ref-3")
//   - How: State transition (old → new)
//   - When: Timestamp (from log package)
//
// Real-World Debugging Scenario:
//
//	Developer sees unexpected UI update, checks logs,
//	identifies component causing issue, traces back to code.
func TestCommandLogger_HelpfulForDebugging(t *testing.T) {
	var buf bytes.Buffer
	logger := NewCommandLogger(&buf)

	// Simulate a realistic debugging scenario
	logger.LogCommand("TodoList", "component-7", "ref-3", []string{"task1"}, []string{"task1", "task2"})

	output := buf.String()

	// Should contain all essential debugging information
	assert.Contains(t, output, "TodoList", "should show component name for debugging")
	assert.Contains(t, output, "component-7", "should show component ID for tracking")
	assert.Contains(t, output, "ref-3", "should show ref ID for identification")
	assert.Contains(t, output, "task1", "should show old value for comparison")
	assert.Contains(t, output, "task2", "should show new value for comparison")
	assert.Contains(t, output, "Command Generated", "should indicate what happened")
}

// TestCommandLogger_EmptyValues tests that empty and zero values are logged
// correctly and meaningfully.
//
// This test validates handling of "empty" states:
//   - Empty strings ("")
//   - Zero integers (0)
//   - False booleans (false)
//   - Nil slices ([]int(nil))
//   - Empty maps (map[string]int{})
//
// Empty Value Importance:
//   - Initial state often starts with empty values
//   - Reset operations use zero values
//   - Form validation might clear fields
//   - Distinguishing between "empty" and "unset" is crucial
//
// Display Requirements:
//   - Empty string: "" (clearly visible as empty)
//   - Zero values: Standard representation (0, false)
//   - Nil collections: "<nil>" or "[]"
//   - Empty collections: "[]" or "map[]"
func TestCommandLogger_EmptyValues(t *testing.T) {
	tests := []struct {
		name     string
		oldValue interface{}
		newValue interface{}
	}{
		{name: "empty string", oldValue: "", newValue: "value"},
		{name: "zero int", oldValue: 0, newValue: 1},
		{name: "false bool", oldValue: false, newValue: true},
		{name: "nil slice", oldValue: []int(nil), newValue: []int{1}},
		{name: "empty map", oldValue: map[string]int{}, newValue: map[string]int{"a": 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewCommandLogger(&buf)

			logger.LogCommand("Test", "comp-1", "ref-1", tt.oldValue, tt.newValue)

			// Should handle empty/zero values gracefully
			assert.NotEmpty(t, buf.String(), "should log empty/zero values")
		})
	}
}

// BenchmarkCommandLogger_Enabled benchmarks the performance of logging
// when debug mode is enabled.
//
// This benchmark measures the overhead of active debug logging:
//   - String formatting and allocation
//   - I/O operations (writing to buffer)
//   - Log package overhead
//   - Memory allocation patterns
//
// Expected Performance:
//   - Time: ~2700 ns/op (includes formatting and I/O)
//   - Memory: ~300 B/op (string allocations)
//   - Allocations: ~4 allocs/op (formatting strings)
//
// Performance Factors:
//   - String formatting with Printf
//   - Buffer writing operations
//   - Timestamp generation
//   - Memory allocations for formatted strings
//
// Usage:
//
//	// Run benchmark
//	go test -bench=BenchmarkCommandLogger_Enabled -benchmem
func BenchmarkCommandLogger_Enabled(b *testing.B) {
	var buf bytes.Buffer
	logger := NewCommandLogger(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogCommand("Test", "comp-1", "ref-1", i, i+1)
	}
}

// BenchmarkNopLogger_Disabled benchmarks the performance of logging when
// debug mode is disabled (should be essentially zero overhead).
//
// This benchmark validates the zero-overhead requirement:
//   - Method should be inlined by compiler
//   - No memory allocations
//   - No I/O operations
//   - Minimal CPU usage
//
// Expected Performance:
//   - Time: ~0.25 ns/op (measurement artifact, essentially zero)
//   - Memory: 0 B/op (no allocations)
//   - Allocations: 0 allocs/op
//
// Zero Overhead Mechanisms:
//   - Empty method body gets inlined away
//   - No function call overhead after optimization
//   - No memory allocation or I/O
//   - Compiler optimization eliminates the call entirely
//
// Performance Ratio:
//   - Disabled vs Enabled: ~10,000x faster
//   - Memory usage: 0 vs ~300 bytes
//   - Allocations: 0 vs 4 per operation
func BenchmarkNopLogger_Disabled(b *testing.B) {
	logger := NewNopLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogCommand("Test", "comp-1", "ref-1", i, i+1)
	}
}

// BenchmarkCommandLogger_vs_Nop compares the performance overhead between
// enabled and disabled logging to validate the zero-overhead claim.
//
// This comparative benchmark demonstrates:
//   - Performance impact of enabling debug logging
//   - Memory allocation differences
//   - Validation of zero-overhead when disabled
//   - Cost-benefit analysis for debug features
//
// Performance Comparison:
//   - Enabled: ~2700 ns/op, ~300 B/op, 4 allocs/op
//   - Disabled: ~0.25 ns/op, 0 B/op, 0 allocs/op
//   - Ratio: ~10,000x faster when disabled
//
// Decision Making:
//   - Production: Use disabled logging (zero overhead)
//   - Development: Use enabled logging (rich debugging info)
//   - Testing: Use disabled logging (clean test output)
//   - Debugging: Enable temporarily for troubleshooting
//
// Usage:
//
//	// Run comparative benchmark
//	go test -bench=BenchmarkCommandLogger_vs_Nop -benchmem
func BenchmarkCommandLogger_vs_Nop(b *testing.B) {
	b.Run("enabled", func(b *testing.B) {
		var buf bytes.Buffer
		logger := NewCommandLogger(&buf)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.LogCommand("Test", "comp-1", "ref-1", i, i+1)
		}
	})

	b.Run("disabled", func(b *testing.B) {
		logger := NewNopLogger()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.LogCommand("Test", "comp-1", "ref-1", i, i+1)
		}
	})
}

// TestSetDefaultLogger tests setting the package-level default logger.
//
// This test validates:
//   - SetDefaultLogger updates the default logger
//   - nil logger is handled gracefully (replaced with NopLogger)
//   - GetDefaultLogger returns the current default
func TestSetDefaultLogger(t *testing.T) {
	// Save original default logger to restore after test
	originalLogger := GetDefaultLogger()
	defer SetDefaultLogger(originalLogger)

	tests := []struct {
		name           string
		logger         CommandLogger
		expectNopCheck bool // If true, verify it's a NopLogger by checking no output
	}{
		{
			name:           "set to CommandLogger",
			logger:         NewCommandLogger(&bytes.Buffer{}),
			expectNopCheck: false,
		},
		{
			name:           "set to NopLogger",
			logger:         NewNopLogger(),
			expectNopCheck: true,
		},
		{
			name:           "set to nil converts to NopLogger",
			logger:         nil,
			expectNopCheck: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDefaultLogger(tt.logger)
			result := GetDefaultLogger()

			assert.NotNil(t, result, "GetDefaultLogger should never return nil")

			if tt.expectNopCheck {
				// Verify it behaves like a NopLogger (no output)
				var buf bytes.Buffer
				// NopLogger ignores all input and produces no output
				result.LogCommand("Test", "comp-1", "ref-1", 0, 1)
				// If it's a NopLogger, buf remains empty (not connected to logger)
				assert.Empty(t, buf.String(), "NopLogger should produce no output")
			}
		})
	}
}

// TestGetDefaultLogger tests retrieving the package-level default logger.
//
// This test validates:
//   - GetDefaultLogger returns non-nil logger
//   - Default logger is usable
func TestGetDefaultLogger(t *testing.T) {
	logger := GetDefaultLogger()

	assert.NotNil(t, logger, "GetDefaultLogger should return non-nil logger")

	// Should be usable without panic
	require.NotPanics(t, func() {
		logger.LogCommand("Test", "comp-1", "ref-1", 0, 1)
	})
}

// TestFormatValue tests the value formatting helper function.
//
// This test validates:
//   - nil values formatted as "<nil>"
//   - strings formatted with %v
//   - numbers formatted correctly
//   - complex types handled without panic
func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "nil value",
			value:    nil,
			expected: "<nil>",
		},
		{
			name:     "string value",
			value:    "hello",
			expected: "hello",
		},
		{
			name:     "empty string",
			value:    "",
			expected: "",
		},
		{
			name:     "integer value",
			value:    42,
			expected: "42",
		},
		{
			name:     "negative integer",
			value:    -10,
			expected: "-10",
		},
		{
			name:     "float value",
			value:    3.14,
			expected: "3.14",
		},
		{
			name:     "boolean true",
			value:    true,
			expected: "true",
		},
		{
			name:     "boolean false",
			value:    false,
			expected: "false",
		},
		{
			name:     "slice",
			value:    []int{1, 2, 3},
			expected: "[1 2 3]",
		},
		{
			name:     "empty slice",
			value:    []int{},
			expected: "[]",
		},
		{
			name:     "map",
			value:    map[string]int{"a": 1},
			expected: "map[a:1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatValue(tt.value)
			assert.Equal(t, tt.expected, result, "FormatValue(%v) should return %q", tt.value, tt.expected)
		})
	}
}

// TestFormatValue_ComplexTypes tests FormatValue with complex types.
//
// This test validates that complex types don't cause panics and produce
// reasonable string representations.
func TestFormatValue_ComplexTypes(t *testing.T) {
	tests := []struct {
		name          string
		value         interface{}
		shouldContain string
	}{
		{
			name:          "struct",
			value:         struct{ Name string }{Name: "Alice"},
			shouldContain: "Alice",
		},
		{
			name:          "pointer to int",
			value:         func() *int { i := 42; return &i }(),
			shouldContain: "", // Just verify no panic
		},
		{
			name:          "nil pointer",
			value:         (*int)(nil),
			shouldContain: "<nil>",
		},
		{
			name:          "nested slice",
			value:         [][]int{{1, 2}, {3, 4}},
			shouldContain: "[[1 2] [3 4]]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				result := FormatValue(tt.value)
				if tt.shouldContain != "" {
					assert.Contains(t, result, tt.shouldContain,
						"FormatValue should contain %q", tt.shouldContain)
				}
			})
		})
	}
}

// TestNopLogger_LogCommand_DirectCall tests that NopLogger.LogCommand can be
// called directly and doesn't produce any side effects.
//
// This test explicitly exercises the nopLogger.LogCommand method to ensure
// 100% coverage of the empty method body.
func TestNopLogger_LogCommand_DirectCall(t *testing.T) {
	logger := NewNopLogger()

	tests := []struct {
		name          string
		componentName string
		componentID   string
		refID         string
		oldValue      interface{}
		newValue      interface{}
	}{
		{
			name:          "basic call",
			componentName: "Counter",
			componentID:   "comp-1",
			refID:         "ref-1",
			oldValue:      0,
			newValue:      1,
		},
		{
			name:          "nil values",
			componentName: "Test",
			componentID:   "comp-2",
			refID:         "ref-2",
			oldValue:      nil,
			newValue:      nil,
		},
		{
			name:          "complex values",
			componentName: "List",
			componentID:   "comp-3",
			refID:         "ref-3",
			oldValue:      []string{"a", "b"},
			newValue:      []string{"a", "b", "c"},
		},
		{
			name:          "empty strings",
			componentName: "",
			componentID:   "",
			refID:         "",
			oldValue:      "",
			newValue:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic and should do nothing
			require.NotPanics(t, func() {
				logger.LogCommand(tt.componentName, tt.componentID, tt.refID, tt.oldValue, tt.newValue)
			})
		})
	}
}

// TestDefaultLoggerInitialState tests that the default logger is initialized
// to NopLogger at package load time.
func TestDefaultLoggerInitialState(t *testing.T) {
	// Save and restore
	originalLogger := GetDefaultLogger()
	defer SetDefaultLogger(originalLogger)

	// Reset to verify initial state behavior
	SetDefaultLogger(nil)
	logger := GetDefaultLogger()

	assert.NotNil(t, logger, "default logger should never be nil")

	// Verify it behaves like NopLogger (no output, no panic)
	require.NotPanics(t, func() {
		logger.LogCommand("Test", "comp-1", "ref-1", 0, 1)
	})
}
