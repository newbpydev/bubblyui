package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// DebugLogger logs all events to a file for debugging
type DebugLogger struct {
	file *os.File
	mu   sync.Mutex
}

var globalDebugLogger *DebugLogger

func InitDebugLogger(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	globalDebugLogger = &DebugLogger{file: f}
	globalDebugLogger.Log("DEBUG", "Debug logger initialized")
	return nil
}

func (d *DebugLogger) Log(level, message string, args ...interface{}) {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(message, args...)
	line := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, msg)

	d.file.WriteString(line)
	d.file.Sync() // Ensure it's written immediately
}

func (d *DebugLogger) LogWithStack(level, message string, args ...interface{}) {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(message, args...)
	stack := string(debug.Stack())

	line := fmt.Sprintf("[%s] [%s] %s\nStack:\n%s\n", timestamp, level, msg, stack)

	d.file.WriteString(line)
	d.file.Sync()
}

func (d *DebugLogger) Close() {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	d.file.WriteString(fmt.Sprintf("[%s] [DEBUG] Closing debug logger\n", time.Now().Format("2006-01-02 15:04:05.000")))
	d.file.Close()
}

// Helper functions
func DebugLog(level, message string, args ...interface{}) {
	if globalDebugLogger != nil {
		globalDebugLogger.Log(level, message, args...)
	}
}

func DebugLogWithStack(level, message string, args ...interface{}) {
	if globalDebugLogger != nil {
		globalDebugLogger.LogWithStack(level, message, args...)
	}
}

func CloseDebugLogger() {
	if globalDebugLogger != nil {
		globalDebugLogger.Close()
	}
}
