package observability

import (
	"sync"
	"time"
)

// MaxBreadcrumbs is the maximum number of breadcrumbs to keep in memory.
// When this limit is reached, the oldest breadcrumbs are dropped (FIFO).
// This prevents unbounded memory growth while keeping recent context.
const MaxBreadcrumbs = 100

// breadcrumbBuffer is a thread-safe circular buffer for storing breadcrumbs.
// It automatically drops the oldest breadcrumbs when capacity is reached.
type breadcrumbBuffer struct {
	// items stores the breadcrumbs in chronological order
	items []Breadcrumb
	// mu protects concurrent access to items
	mu sync.RWMutex
}

// Global breadcrumb buffer
var globalBreadcrumbs = &breadcrumbBuffer{
	items: make([]Breadcrumb, 0, MaxBreadcrumbs),
}

// RecordBreadcrumb records a breadcrumb with the given category, message, and optional data.
// Breadcrumbs are stored in chronological order and automatically dropped when
// the maximum capacity (100) is reached.
//
// This function is thread-safe and can be called concurrently from multiple goroutines.
//
// Parameters:
//   - category: Subcategory for grouping breadcrumbs (e.g., "ui", "network", "state")
//   - message: Human-readable description of the breadcrumb
//   - data: Optional additional data about the breadcrumb (can be nil)
//
// Example:
//
//	// Record navigation breadcrumb
//	RecordBreadcrumb("navigation", "User navigated to login page", map[string]interface{}{
//	    "from": "/home",
//	    "to":   "/login",
//	})
//
//	// Record user action
//	RecordBreadcrumb("user", "User clicked submit button", map[string]interface{}{
//	    "button": "submit",
//	    "form":   "login",
//	})
//
//	// Record without data
//	RecordBreadcrumb("debug", "Component mounted", nil)
//
// Thread-safe: Safe to call concurrently from multiple goroutines.
func RecordBreadcrumb(category, message string, data map[string]interface{}) {
	globalBreadcrumbs.add(category, message, data)
}

// GetBreadcrumbs returns a copy of all recorded breadcrumbs in chronological order.
// The oldest breadcrumb is first, the newest is last.
//
// The returned slice is a defensive copy - modifying it will not affect
// the internal breadcrumb buffer.
//
// This function is thread-safe and can be called concurrently from multiple goroutines.
//
// Returns:
//   - []Breadcrumb: A copy of all breadcrumbs in chronological order
//
// Example:
//
//	breadcrumbs := GetBreadcrumbs()
//	for _, bc := range breadcrumbs {
//	    log.Printf("[%s] %s: %s", bc.Timestamp.Format(time.RFC3339), bc.Category, bc.Message)
//	}
//
// Thread-safe: Safe to call concurrently from multiple goroutines.
func GetBreadcrumbs() []Breadcrumb {
	return globalBreadcrumbs.getAll()
}

// ClearBreadcrumbs removes all recorded breadcrumbs.
// This is useful for testing or when starting a new user session.
//
// This function is thread-safe and can be called concurrently from multiple goroutines.
//
// Example:
//
//	// Clear breadcrumbs at the start of a test
//	ClearBreadcrumbs()
//
//	// Clear breadcrumbs when user logs out
//	ClearBreadcrumbs()
//
// Thread-safe: Safe to call concurrently from multiple goroutines.
func ClearBreadcrumbs() {
	globalBreadcrumbs.clear()
}

// add adds a breadcrumb to the buffer.
// If the buffer is at capacity, the oldest breadcrumb is dropped.
func (b *breadcrumbBuffer) add(category, message string, data map[string]interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Create breadcrumb with current timestamp
	breadcrumb := Breadcrumb{
		Type:      "default", // Default type, can be customized by caller
		Category:  category,
		Message:   message,
		Level:     "info", // Default level
		Timestamp: time.Now(),
		Data:      copyData(data),
	}

	// If at capacity, drop oldest breadcrumb (shift left)
	if len(b.items) >= MaxBreadcrumbs {
		// Shift all items left by 1 (drop first item)
		copy(b.items, b.items[1:])
		// Replace last item with new breadcrumb
		b.items[MaxBreadcrumbs-1] = breadcrumb
	} else {
		// Append new breadcrumb
		b.items = append(b.items, breadcrumb)
	}
}

// getAll returns a defensive copy of all breadcrumbs.
func (b *breadcrumbBuffer) getAll() []Breadcrumb {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Return defensive copy to prevent external modification
	result := make([]Breadcrumb, len(b.items))
	copy(result, b.items)
	return result
}

// clear removes all breadcrumbs from the buffer.
func (b *breadcrumbBuffer) clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Reset slice to empty (keep capacity for reuse)
	b.items = b.items[:0]
}

// copyData creates a defensive copy of the data map.
// This prevents external modifications from affecting stored breadcrumbs.
func copyData(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}

	// Create new map with same capacity
	result := make(map[string]interface{}, len(data))
	for k, v := range data {
		result[k] = v
	}
	return result
}
