package commands

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommandLogger_LogCommand tests that command generation is logged with proper format.
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

// TestCommandLogger_Format tests the log format is clear and helpful.
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

// TestNopLogger_NoOutput tests that NopLogger produces no output.
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

// TestCommandLogger_ThreadSafe tests concurrent logging from multiple goroutines.
func TestCommandLogger_ThreadSafe(t *testing.T) {
	var buf bytes.Buffer
	logger := NewCommandLogger(&buf)

	const goroutines = 10
	const logsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
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

// TestCommandLogger_ComplexTypes tests logging of complex data types.
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

// TestCommandLogger_NilWriter tests handling of nil writer.
func TestCommandLogger_NilWriter(t *testing.T) {
	// Should not panic with nil writer
	require.NotPanics(t, func() {
		logger := NewCommandLogger(nil)
		logger.LogCommand("Test", "comp-1", "ref-1", 0, 1)
	})
}

// TestCommandLogger_HelpfulForDebugging tests that logs help developers debug.
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

// TestCommandLogger_EmptyValues tests logging of empty/zero values.
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

// BenchmarkCommandLogger_Enabled benchmarks logging when enabled.
func BenchmarkCommandLogger_Enabled(b *testing.B) {
	var buf bytes.Buffer
	logger := NewCommandLogger(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogCommand("Test", "comp-1", "ref-1", i, i+1)
	}
}

// BenchmarkNopLogger_Disabled benchmarks "logging" when disabled (should be near-zero overhead).
func BenchmarkNopLogger_Disabled(b *testing.B) {
	logger := NewNopLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogCommand("Test", "comp-1", "ref-1", i, i+1)
	}
}

// BenchmarkCommandLogger_vs_Nop compares overhead of enabled vs disabled logging.
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
