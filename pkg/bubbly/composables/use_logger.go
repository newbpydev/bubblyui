package composables

import (
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// LogLevel defines logging levels.
// Levels are ordered from most verbose (Debug) to least verbose (Error).
// When a level is set, only messages at that level or higher will be logged.
type LogLevel int

const (
	// LogLevelDebug is the most verbose level, used for debugging information.
	LogLevelDebug LogLevel = iota

	// LogLevelInfo is for informational messages about normal operation.
	LogLevelInfo

	// LogLevelWarn is for warning messages about potential issues.
	LogLevelWarn

	// LogLevelError is for error messages about failures.
	LogLevelError
)

// String returns the string representation of the log level.
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a log entry.
// Each entry contains the timestamp, level, component name, message, and optional data.
type LogEntry struct {
	// Time is the timestamp when the log entry was created.
	Time time.Time

	// Level is the log level of this entry.
	Level LogLevel

	// Component is the name of the component that created this entry.
	Component string

	// Message is the log message.
	Message string

	// Data is optional additional data attached to the entry.
	// If a single value is passed, it is stored directly.
	// If multiple values are passed, they are stored as a slice.
	Data interface{}
}

// LoggerReturn is the return value of UseLogger.
// It provides component debug logging with configurable levels and log history.
type LoggerReturn struct {
	// Level is the current log level.
	// Only messages at this level or higher will be logged.
	Level *bubbly.Ref[LogLevel]

	// Logs is the log history.
	// New entries are appended to the end.
	Logs *bubbly.Ref[[]LogEntry]

	// componentName is the name of the component for log prefixing.
	componentName string

	// mu protects concurrent access to log operations.
	mu sync.Mutex
}

// shouldLog returns true if a message at the given level should be logged.
func (l *LoggerReturn) shouldLog(level LogLevel) bool {
	return level >= l.Level.GetTyped()
}

// log adds a log entry at the specified level.
func (l *LoggerReturn) log(level LogLevel, msg string, data ...interface{}) {
	if !l.shouldLog(level) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Time:      time.Now(),
		Level:     level,
		Component: l.componentName,
		Message:   msg,
	}

	// Handle data attachment
	switch len(data) {
	case 0:
		entry.Data = nil
	case 1:
		entry.Data = data[0]
	default:
		entry.Data = data
	}

	// Append to logs
	current := l.Logs.GetTyped()
	newLogs := append(current, entry)
	l.Logs.Set(newLogs)
}

// Debug logs at debug level.
// Debug messages are the most verbose and are typically used for
// detailed debugging information during development.
//
// Example:
//
//	logger.Debug("Processing item", itemID)
//	logger.Debug("State changed", map[string]interface{}{"old": old, "new": new})
func (l *LoggerReturn) Debug(msg string, data ...interface{}) {
	l.log(LogLevelDebug, msg, data...)
}

// Info logs at info level.
// Info messages are for informational purposes about normal operation.
//
// Example:
//
//	logger.Info("User logged in", userID)
//	logger.Info("Component initialized")
func (l *LoggerReturn) Info(msg string, data ...interface{}) {
	l.log(LogLevelInfo, msg, data...)
}

// Warn logs at warn level.
// Warn messages indicate potential issues that don't prevent operation
// but should be investigated.
//
// Example:
//
//	logger.Warn("Connection slow", latency)
//	logger.Warn("Deprecated feature used", featureName)
func (l *LoggerReturn) Warn(msg string, data ...interface{}) {
	l.log(LogLevelWarn, msg, data...)
}

// Error logs at error level.
// Error messages indicate failures that may affect operation.
//
// Example:
//
//	logger.Error("Failed to save", err)
//	logger.Error("Connection lost", map[string]interface{}{"host": host, "err": err})
func (l *LoggerReturn) Error(msg string, data ...interface{}) {
	l.log(LogLevelError, msg, data...)
}

// Clear clears log history.
// This removes all log entries from the history.
//
// Example:
//
//	logger.Clear() // Remove all logs
func (l *LoggerReturn) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.Logs.Set([]LogEntry{})
}

// UseLogger creates a logging composable.
// It provides component debug logging with configurable levels and log history.
//
// This composable is useful for:
//   - Debugging component state and lifecycle
//   - Tracking state changes with context
//   - Development-time logging
//   - Integration with devtools
//
// Parameters:
//   - ctx: The component context (can be nil for testing)
//   - componentName: The name of the component for log prefixing
//
// Returns:
//   - *LoggerReturn: A struct containing the log level ref, logs ref, and logging methods
//
// Example - Basic usage:
//
//	Setup(func(ctx *bubbly.Context) {
//	    logger := composables.UseLogger(ctx, "MyComponent")
//	    ctx.Expose("logger", logger)
//
//	    // Log at different levels
//	    logger.Debug("Component setup started")
//	    logger.Info("Initializing state")
//	    logger.Warn("Using deprecated feature")
//	    logger.Error("Failed to load data", err)
//	})
//
// Example - Level filtering:
//
//	Setup(func(ctx *bubbly.Context) {
//	    logger := composables.UseLogger(ctx, "MyComponent")
//
//	    // Set to only log warnings and errors
//	    logger.Level.Set(composables.LogLevelWarn)
//
//	    logger.Debug("This won't be logged")
//	    logger.Info("This won't be logged either")
//	    logger.Warn("This will be logged")
//	    logger.Error("This will be logged")
//	})
//
// Example - Attaching data:
//
//	logger.Info("User action", map[string]interface{}{
//	    "action": "click",
//	    "target": "button",
//	    "timestamp": time.Now(),
//	})
//
// Example - Viewing logs:
//
//	Template(func(ctx bubbly.RenderContext) string {
//	    logger := ctx.Get("logger").(*composables.LoggerReturn)
//	    logs := logger.Logs.GetTyped()
//
//	    var lines []string
//	    for _, entry := range logs {
//	        lines = append(lines, fmt.Sprintf("[%s] %s: %s",
//	            entry.Level, entry.Component, entry.Message))
//	    }
//	    return strings.Join(lines, "\n")
//	})
//
// Integration with CreateShared:
//
//	var UseSharedLogger = composables.CreateShared(
//	    func(ctx *bubbly.Context) *composables.LoggerReturn {
//	        return composables.UseLogger(ctx, "SharedLogger")
//	    },
//	)
//
// Thread Safety:
//
// UseLogger is thread-safe. All logging operations are synchronized with a mutex.
// The Logs ref can be safely accessed from multiple goroutines.
func UseLogger(ctx *bubbly.Context, componentName string) *LoggerReturn {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseLogger", time.Since(start))
	}()

	// Create refs with default values
	level := bubbly.NewRef(LogLevelDebug)
	logs := bubbly.NewRef([]LogEntry{})

	return &LoggerReturn{
		Level:         level,
		Logs:          logs,
		componentName: componentName,
	}
}
