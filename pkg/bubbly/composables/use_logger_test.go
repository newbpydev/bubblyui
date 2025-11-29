package composables

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestUseLogger_InitialState tests that UseLogger initializes correctly
func TestUseLogger_InitialState(t *testing.T) {
	ctx := createTestContext()
	logger := UseLogger(ctx, "TestComponent")

	require.NotNil(t, logger, "UseLogger should return non-nil")
	require.NotNil(t, logger.Level, "Level should not be nil")
	require.NotNil(t, logger.Logs, "Logs should not be nil")

	// Default level should be Debug (most permissive)
	assert.Equal(t, LogLevelDebug, logger.Level.GetTyped(),
		"Default log level should be Debug")

	// Initial logs should be empty
	assert.Empty(t, logger.Logs.GetTyped(),
		"Initial logs should be empty")
}

// TestUseLogger_DebugLogsAtDebugLevel tests Debug() logs at debug level
func TestUseLogger_DebugLogsAtDebugLevel(t *testing.T) {
	ctx := createTestContext()
	logger := UseLogger(ctx, "TestComponent")

	beforeLog := time.Now()
	logger.Debug("debug message")
	afterLog := time.Now()

	logs := logger.Logs.GetTyped()
	require.Len(t, logs, 1, "Should have one log entry")

	entry := logs[0]
	assert.Equal(t, LogLevelDebug, entry.Level, "Level should be Debug")
	assert.Equal(t, "TestComponent", entry.Component, "Component should match")
	assert.Equal(t, "debug message", entry.Message, "Message should match")
	assert.True(t, entry.Time.After(beforeLog) || entry.Time.Equal(beforeLog),
		"Timestamp should be after or equal to beforeLog")
	assert.True(t, entry.Time.Before(afterLog) || entry.Time.Equal(afterLog),
		"Timestamp should be before or equal to afterLog")
}

// TestUseLogger_InfoLogsAtInfoLevel tests Info() logs at info level
func TestUseLogger_InfoLogsAtInfoLevel(t *testing.T) {
	ctx := createTestContext()
	logger := UseLogger(ctx, "TestComponent")

	logger.Info("info message")

	logs := logger.Logs.GetTyped()
	require.Len(t, logs, 1, "Should have one log entry")

	entry := logs[0]
	assert.Equal(t, LogLevelInfo, entry.Level, "Level should be Info")
	assert.Equal(t, "info message", entry.Message, "Message should match")
}

// TestUseLogger_WarnLogsAtWarnLevel tests Warn() logs at warn level
func TestUseLogger_WarnLogsAtWarnLevel(t *testing.T) {
	ctx := createTestContext()
	logger := UseLogger(ctx, "TestComponent")

	logger.Warn("warn message")

	logs := logger.Logs.GetTyped()
	require.Len(t, logs, 1, "Should have one log entry")

	entry := logs[0]
	assert.Equal(t, LogLevelWarn, entry.Level, "Level should be Warn")
	assert.Equal(t, "warn message", entry.Message, "Message should match")
}

// TestUseLogger_ErrorLogsAtErrorLevel tests Error() logs at error level
func TestUseLogger_ErrorLogsAtErrorLevel(t *testing.T) {
	ctx := createTestContext()
	logger := UseLogger(ctx, "TestComponent")

	logger.Error("error message")

	logs := logger.Logs.GetTyped()
	require.Len(t, logs, 1, "Should have one log entry")

	entry := logs[0]
	assert.Equal(t, LogLevelError, entry.Level, "Level should be Error")
	assert.Equal(t, "error message", entry.Message, "Message should match")
}

// TestUseLogger_LogEntryIncludesTimestampComponentMessage tests log entry structure
func TestUseLogger_LogEntryIncludesTimestampComponentMessage(t *testing.T) {
	tests := []struct {
		name      string
		component string
		message   string
	}{
		{"simple", "App", "hello"},
		{"with spaces", "My Component", "hello world"},
		{"empty message", "Test", ""},
		{"special chars", "Test/Sub", "message with 'quotes' and \"double quotes\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			logger := UseLogger(ctx, tt.component)

			beforeLog := time.Now()
			logger.Info(tt.message)
			afterLog := time.Now()

			logs := logger.Logs.GetTyped()
			require.Len(t, logs, 1)

			entry := logs[0]
			assert.Equal(t, tt.component, entry.Component)
			assert.Equal(t, tt.message, entry.Message)
			assert.False(t, entry.Time.IsZero(), "Timestamp should not be zero")
			assert.True(t, entry.Time.After(beforeLog) || entry.Time.Equal(beforeLog))
			assert.True(t, entry.Time.Before(afterLog) || entry.Time.Equal(afterLog))
		})
	}
}

// TestUseLogger_LevelFilteringWorks tests that level filtering works correctly
func TestUseLogger_LevelFilteringWorks(t *testing.T) {
	tests := []struct {
		name      string
		setLevel  LogLevel
		logMethod string
		shouldLog bool
	}{
		// Debug level - logs everything
		{"debug level logs debug", LogLevelDebug, "debug", true},
		{"debug level logs info", LogLevelDebug, "info", true},
		{"debug level logs warn", LogLevelDebug, "warn", true},
		{"debug level logs error", LogLevelDebug, "error", true},

		// Info level - filters debug
		{"info level filters debug", LogLevelInfo, "debug", false},
		{"info level logs info", LogLevelInfo, "info", true},
		{"info level logs warn", LogLevelInfo, "warn", true},
		{"info level logs error", LogLevelInfo, "error", true},

		// Warn level - filters debug and info
		{"warn level filters debug", LogLevelWarn, "debug", false},
		{"warn level filters info", LogLevelWarn, "info", false},
		{"warn level logs warn", LogLevelWarn, "warn", true},
		{"warn level logs error", LogLevelWarn, "error", true},

		// Error level - only logs errors
		{"error level filters debug", LogLevelError, "debug", false},
		{"error level filters info", LogLevelError, "info", false},
		{"error level filters warn", LogLevelError, "warn", false},
		{"error level logs error", LogLevelError, "error", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			logger := UseLogger(ctx, "Test")

			// Set the log level
			logger.Level.Set(tt.setLevel)

			// Log based on method
			switch tt.logMethod {
			case "debug":
				logger.Debug("test message")
			case "info":
				logger.Info("test message")
			case "warn":
				logger.Warn("test message")
			case "error":
				logger.Error("test message")
			}

			logs := logger.Logs.GetTyped()
			if tt.shouldLog {
				assert.Len(t, logs, 1, "Should have logged")
			} else {
				assert.Empty(t, logs, "Should not have logged")
			}
		})
	}
}

// TestUseLogger_ClearRemovesAllLogs tests that Clear() removes all logs
func TestUseLogger_ClearRemovesAllLogs(t *testing.T) {
	ctx := createTestContext()
	logger := UseLogger(ctx, "Test")

	// Add some logs
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	logs := logger.Logs.GetTyped()
	require.Len(t, logs, 4, "Should have 4 log entries")

	// Clear logs
	logger.Clear()

	logs = logger.Logs.GetTyped()
	assert.Empty(t, logs, "Logs should be empty after Clear()")
}

// TestUseLogger_DataAttachedToEntries tests that data is attached to log entries
func TestUseLogger_DataAttachedToEntries(t *testing.T) {
	tests := []struct {
		name     string
		data     []interface{}
		expected interface{}
	}{
		{"no data", nil, nil},
		{"single int", []interface{}{42}, 42},
		{"single string", []interface{}{"value"}, "value"},
		{"single map", []interface{}{map[string]int{"count": 5}}, map[string]int{"count": 5}},
		{"multiple values", []interface{}{"key", "value", 123}, []interface{}{"key", "value", 123}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			logger := UseLogger(ctx, "Test")

			logger.Info("message", tt.data...)

			logs := logger.Logs.GetTyped()
			require.Len(t, logs, 1)

			entry := logs[0]
			assert.Equal(t, tt.expected, entry.Data)
		})
	}
}

// TestUseLogger_MultipleLogsAccumulate tests that multiple logs accumulate
func TestUseLogger_MultipleLogsAccumulate(t *testing.T) {
	ctx := createTestContext()
	logger := UseLogger(ctx, "Test")

	logger.Debug("first")
	logger.Info("second")
	logger.Warn("third")
	logger.Error("fourth")

	logs := logger.Logs.GetTyped()
	require.Len(t, logs, 4, "Should have 4 log entries")

	// Verify order (oldest first)
	assert.Equal(t, "first", logs[0].Message)
	assert.Equal(t, "second", logs[1].Message)
	assert.Equal(t, "third", logs[2].Message)
	assert.Equal(t, "fourth", logs[3].Message)

	// Verify levels
	assert.Equal(t, LogLevelDebug, logs[0].Level)
	assert.Equal(t, LogLevelInfo, logs[1].Level)
	assert.Equal(t, LogLevelWarn, logs[2].Level)
	assert.Equal(t, LogLevelError, logs[3].Level)
}

// TestUseLogger_LevelCanBeChanged tests that level can be changed dynamically
func TestUseLogger_LevelCanBeChanged(t *testing.T) {
	ctx := createTestContext()
	logger := UseLogger(ctx, "Test")

	// Start at debug level
	assert.Equal(t, LogLevelDebug, logger.Level.GetTyped())

	// Log at debug level
	logger.Debug("debug1")
	assert.Len(t, logger.Logs.GetTyped(), 1)

	// Change to error level
	logger.Level.Set(LogLevelError)
	assert.Equal(t, LogLevelError, logger.Level.GetTyped())

	// Debug should now be filtered
	logger.Debug("debug2")
	assert.Len(t, logger.Logs.GetTyped(), 1, "Debug should be filtered at Error level")

	// Error should still log
	logger.Error("error1")
	assert.Len(t, logger.Logs.GetTyped(), 2, "Error should log at Error level")
}

// TestUseLogger_WorksWithCreateShared tests shared composable pattern
func TestUseLogger_WorksWithCreateShared(t *testing.T) {
	sharedLogger := CreateShared(func(ctx *bubbly.Context) *LoggerReturn {
		return UseLogger(ctx, "SharedLogger")
	})

	ctx := createTestContext()
	logger1 := sharedLogger(ctx)
	logger2 := sharedLogger(ctx)

	// Both should be the same instance
	logger1.Info("from logger1")

	logs := logger2.Logs.GetTyped()
	assert.Len(t, logs, 1, "Shared instance should have same logs")
	assert.Equal(t, "from logger1", logs[0].Message)
}

// TestUseLogger_LogsAreReactive tests that Logs ref is reactive
func TestUseLogger_LogsAreReactive(t *testing.T) {
	ctx := createTestContext()
	logger := UseLogger(ctx, "Test")

	// Track changes via Watch
	changeCount := 0
	bubbly.Watch(logger.Logs, func(newVal, oldVal []LogEntry) {
		changeCount++
	})

	// Each log should trigger watcher
	logger.Debug("debug")
	assert.Equal(t, 1, changeCount, "Debug should trigger watcher")

	logger.Info("info")
	assert.Equal(t, 2, changeCount, "Info should trigger watcher")

	// Clear should trigger watcher
	logger.Clear()
	assert.Equal(t, 3, changeCount, "Clear should trigger watcher")
}

// TestUseLogger_LogLevelConstants tests that log level constants are correct
func TestUseLogger_LogLevelConstants(t *testing.T) {
	// Verify ordering: Debug < Info < Warn < Error
	assert.Less(t, int(LogLevelDebug), int(LogLevelInfo), "Debug < Info")
	assert.Less(t, int(LogLevelInfo), int(LogLevelWarn), "Info < Warn")
	assert.Less(t, int(LogLevelWarn), int(LogLevelError), "Warn < Error")
}

// TestUseLogger_ComponentNamePreserved tests that component name is preserved
func TestUseLogger_ComponentNamePreserved(t *testing.T) {
	ctx := createTestContext()
	logger := UseLogger(ctx, "MyComponent")

	logger.Debug("msg1")
	logger.Info("msg2")
	logger.Warn("msg3")
	logger.Error("msg4")

	logs := logger.Logs.GetTyped()
	for _, entry := range logs {
		assert.Equal(t, "MyComponent", entry.Component,
			"All entries should have the same component name")
	}
}

// TestUseLogger_EmptyComponentName tests that empty component name works
func TestUseLogger_EmptyComponentName(t *testing.T) {
	ctx := createTestContext()
	logger := UseLogger(ctx, "")

	logger.Info("message")

	logs := logger.Logs.GetTyped()
	require.Len(t, logs, 1)
	assert.Equal(t, "", logs[0].Component)
}

// TestUseLogger_ConcurrentAccess tests thread safety
func TestUseLogger_ConcurrentAccess(t *testing.T) {
	ctx := createTestContext()
	logger := UseLogger(ctx, "Test")

	// Run concurrent logging
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 10; j++ {
				logger.Info("message")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 100 log entries
	logs := logger.Logs.GetTyped()
	assert.Len(t, logs, 100, "Should have 100 log entries from concurrent access")
}

// TestLogLevel_String tests the String() method of LogLevel
func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelDebug, "DEBUG"},
		{LogLevelInfo, "INFO"},
		{LogLevelWarn, "WARN"},
		{LogLevelError, "ERROR"},
		{LogLevel(99), "UNKNOWN"}, // Unknown level
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}

// TestUseLogger_NilContext tests that UseLogger works with nil context
func TestUseLogger_NilContext(t *testing.T) {
	// Should not panic with nil context
	assert.NotPanics(t, func() {
		logger := UseLogger(nil, "Test")
		assert.NotNil(t, logger)
		logger.Info("message")
	})
}

// TestUseLogger_LogEntryTimeOrdering tests that log entries maintain time ordering
func TestUseLogger_LogEntryTimeOrdering(t *testing.T) {
	ctx := createTestContext()
	logger := UseLogger(ctx, "Test")

	// Log multiple entries
	for i := 0; i < 5; i++ {
		logger.Info("message")
	}

	logs := logger.Logs.GetTyped()
	require.Len(t, logs, 5)

	// Verify time ordering (each entry should be >= previous)
	for i := 1; i < len(logs); i++ {
		assert.True(t, logs[i].Time.After(logs[i-1].Time) || logs[i].Time.Equal(logs[i-1].Time),
			"Log entry %d should be after or equal to entry %d", i, i-1)
	}
}

// TestUseLogger_SetLevel tests setting level via ref
func TestUseLogger_SetLevel(t *testing.T) {
	ctx := createTestContext()
	logger := UseLogger(ctx, "Test")

	// Test all levels can be set
	levels := []LogLevel{LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError}
	for _, level := range levels {
		logger.Level.Set(level)
		assert.Equal(t, level, logger.Level.GetTyped())
	}
}
