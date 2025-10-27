package observability

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRecordBreadcrumb tests recording single breadcrumbs
func TestRecordBreadcrumb(t *testing.T) {
	tests := []struct {
		name     string
		category string
		message  string
		data     map[string]interface{}
		wantLen  int
	}{
		{
			name:     "record simple breadcrumb",
			category: "navigation",
			message:  "User navigated to login",
			data:     nil,
			wantLen:  1,
		},
		{
			name:     "record breadcrumb with data",
			category: "user",
			message:  "User clicked button",
			data:     map[string]interface{}{"button": "submit"},
			wantLen:  1,
		},
		{
			name:     "record breadcrumb with empty message",
			category: "debug",
			message:  "",
			data:     nil,
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear breadcrumbs before test
			ClearBreadcrumbs()

			// Record breadcrumb
			RecordBreadcrumb(tt.category, tt.message, tt.data)

			// Get breadcrumbs
			breadcrumbs := GetBreadcrumbs()

			// Verify
			require.Len(t, breadcrumbs, tt.wantLen)
			if tt.wantLen > 0 {
				bc := breadcrumbs[0]
				assert.Equal(t, tt.category, bc.Category)
				assert.Equal(t, tt.message, bc.Message)
				assert.NotZero(t, bc.Timestamp)
				if tt.data != nil {
					assert.Equal(t, tt.data, bc.Data)
				}
			}
		})
	}
}

// TestGetBreadcrumbs tests retrieving breadcrumbs
func TestGetBreadcrumbs(t *testing.T) {
	tests := []struct {
		name      string
		setup     func()
		wantLen   int
		wantFirst string
		wantLast  string
	}{
		{
			name: "get empty breadcrumbs",
			setup: func() {
				ClearBreadcrumbs()
			},
			wantLen: 0,
		},
		{
			name: "get single breadcrumb",
			setup: func() {
				ClearBreadcrumbs()
				RecordBreadcrumb("navigation", "Page loaded", nil)
			},
			wantLen:   1,
			wantFirst: "Page loaded",
			wantLast:  "Page loaded",
		},
		{
			name: "get multiple breadcrumbs in order",
			setup: func() {
				ClearBreadcrumbs()
				RecordBreadcrumb("navigation", "First", nil)
				RecordBreadcrumb("user", "Second", nil)
				RecordBreadcrumb("debug", "Third", nil)
			},
			wantLen:   3,
			wantFirst: "First",
			wantLast:  "Third",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setup()

			// Get breadcrumbs
			breadcrumbs := GetBreadcrumbs()

			// Verify length
			assert.Len(t, breadcrumbs, tt.wantLen)

			// Verify order if breadcrumbs exist
			if tt.wantLen > 0 {
				assert.Equal(t, tt.wantFirst, breadcrumbs[0].Message)
				assert.Equal(t, tt.wantLast, breadcrumbs[tt.wantLen-1].Message)
			}
		})
	}
}

// TestClearBreadcrumbs tests clearing all breadcrumbs
func TestClearBreadcrumbs(t *testing.T) {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "clear empty breadcrumbs",
			setup: func() {
				ClearBreadcrumbs()
			},
		},
		{
			name: "clear single breadcrumb",
			setup: func() {
				ClearBreadcrumbs()
				RecordBreadcrumb("test", "message", nil)
			},
		},
		{
			name: "clear multiple breadcrumbs",
			setup: func() {
				ClearBreadcrumbs()
				for i := 0; i < 10; i++ {
					RecordBreadcrumb("test", "message", nil)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setup()

			// Clear
			ClearBreadcrumbs()

			// Verify empty
			breadcrumbs := GetBreadcrumbs()
			assert.Empty(t, breadcrumbs)
		})
	}
}

// TestBreadcrumbs_MaxCapacity tests that max breadcrumbs (100) is enforced
func TestBreadcrumbs_MaxCapacity(t *testing.T) {
	tests := []struct {
		name      string
		count     int
		wantLen   int
		wantFirst string
		wantLast  string
	}{
		{
			name:      "under capacity (50 breadcrumbs)",
			count:     50,
			wantLen:   50,
			wantFirst: "message-0",
			wantLast:  "message-49",
		},
		{
			name:      "at capacity (100 breadcrumbs)",
			count:     100,
			wantLen:   100,
			wantFirst: "message-0",
			wantLast:  "message-99",
		},
		{
			name:      "over capacity (150 breadcrumbs)",
			count:     150,
			wantLen:   100,
			wantFirst: "message-50", // Oldest 50 dropped
			wantLast:  "message-149",
		},
		{
			name:      "far over capacity (200 breadcrumbs)",
			count:     200,
			wantLen:   100,
			wantFirst: "message-100", // Oldest 100 dropped
			wantLast:  "message-199",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear breadcrumbs
			ClearBreadcrumbs()

			// Record breadcrumbs
			for i := 0; i < tt.count; i++ {
				RecordBreadcrumb("test", fmt.Sprintf("message-%d", i), nil)
			}

			// Get breadcrumbs
			breadcrumbs := GetBreadcrumbs()

			// Verify length
			require.Len(t, breadcrumbs, tt.wantLen)

			// Verify oldest breadcrumb (first in slice)
			assert.Equal(t, tt.wantFirst, breadcrumbs[0].Message)

			// Verify newest breadcrumb (last in slice)
			assert.Equal(t, tt.wantLast, breadcrumbs[tt.wantLen-1].Message)
		})
	}
}

// TestBreadcrumbs_Concurrent tests thread-safety with concurrent access
func TestBreadcrumbs_Concurrent(t *testing.T) {
	tests := []struct {
		name       string
		goroutines int
		operations int
	}{
		{
			name:       "10 goroutines, 10 operations each",
			goroutines: 10,
			operations: 10,
		},
		{
			name:       "30 goroutines, 30 operations each",
			goroutines: 30,
			operations: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear breadcrumbs
			ClearBreadcrumbs()

			var wg sync.WaitGroup

			// Concurrent record operations
			for i := 0; i < tt.goroutines; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < tt.operations; j++ {
						RecordBreadcrumb("concurrent", "test", map[string]interface{}{
							"goroutine": id,
							"operation": j,
						})
					}
				}(i)
			}

			// Concurrent read operations
			for i := 0; i < tt.goroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < tt.operations; j++ {
						breadcrumbs := GetBreadcrumbs()
						assert.NotNil(t, breadcrumbs)
					}
				}()
			}

			wg.Wait()

			// Verify breadcrumbs were recorded (may be capped at 100)
			breadcrumbs := GetBreadcrumbs()
			assert.NotEmpty(t, breadcrumbs)
			assert.LessOrEqual(t, len(breadcrumbs), 100)
		})
	}
}

// TestBreadcrumbs_ConcurrentClear tests concurrent clear operations
func TestBreadcrumbs_ConcurrentClear(t *testing.T) {
	tests := []struct {
		name       string
		goroutines int
	}{
		{
			name:       "10 concurrent clears",
			goroutines: 10,
		},
		{
			name:       "50 concurrent clears",
			goroutines: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Add some breadcrumbs
			ClearBreadcrumbs()
			for i := 0; i < 50; i++ {
				RecordBreadcrumb("test", "message", nil)
			}

			var wg sync.WaitGroup

			// Concurrent clear operations
			for i := 0; i < tt.goroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					ClearBreadcrumbs()
				}()
			}

			wg.Wait()

			// Verify empty
			breadcrumbs := GetBreadcrumbs()
			assert.Empty(t, breadcrumbs)
		})
	}
}

// TestBreadcrumbs_Timestamps tests that timestamps are set correctly
func TestBreadcrumbs_Timestamps(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "timestamps are chronological",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear breadcrumbs
			ClearBreadcrumbs()

			// Record breadcrumbs with small delays
			before := time.Now()
			RecordBreadcrumb("test", "first", nil)
			time.Sleep(10 * time.Millisecond)
			RecordBreadcrumb("test", "second", nil)
			time.Sleep(10 * time.Millisecond)
			RecordBreadcrumb("test", "third", nil)
			after := time.Now()

			// Get breadcrumbs
			breadcrumbs := GetBreadcrumbs()

			// Verify timestamps
			require.Len(t, breadcrumbs, 3)

			// All timestamps should be within test time range
			for _, bc := range breadcrumbs {
				assert.True(t, bc.Timestamp.After(before) || bc.Timestamp.Equal(before))
				assert.True(t, bc.Timestamp.Before(after) || bc.Timestamp.Equal(after))
			}

			// Timestamps should be in chronological order
			assert.True(t, breadcrumbs[0].Timestamp.Before(breadcrumbs[1].Timestamp) ||
				breadcrumbs[0].Timestamp.Equal(breadcrumbs[1].Timestamp))
			assert.True(t, breadcrumbs[1].Timestamp.Before(breadcrumbs[2].Timestamp) ||
				breadcrumbs[1].Timestamp.Equal(breadcrumbs[2].Timestamp))
		})
	}
}

// TestBreadcrumbs_DataIsolation tests that breadcrumb data is properly isolated
func TestBreadcrumbs_DataIsolation(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "modifying data after recording doesn't affect breadcrumb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear breadcrumbs
			ClearBreadcrumbs()

			// Create data map
			data := map[string]interface{}{
				"key": "original",
			}

			// Record breadcrumb
			RecordBreadcrumb("test", "message", data)

			// Modify original data
			data["key"] = "modified"
			data["new_key"] = "new_value"

			// Get breadcrumbs
			breadcrumbs := GetBreadcrumbs()

			// Verify original data is preserved
			require.Len(t, breadcrumbs, 1)
			assert.Equal(t, "original", breadcrumbs[0].Data["key"])
			assert.NotContains(t, breadcrumbs[0].Data, "new_key")
		})
	}
}

// TestBreadcrumbs_GetReturnsDefensiveCopy tests that GetBreadcrumbs returns a copy
func TestBreadcrumbs_GetReturnsDefensiveCopy(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "modifying returned slice doesn't affect internal state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear and add breadcrumbs
			ClearBreadcrumbs()
			RecordBreadcrumb("test", "first", nil)
			RecordBreadcrumb("test", "second", nil)

			// Get breadcrumbs
			breadcrumbs1 := GetBreadcrumbs()
			require.Len(t, breadcrumbs1, 2)

			// Modify returned slice
			breadcrumbs1[0].Message = "modified"
			_ = append(breadcrumbs1, Breadcrumb{Message: "added"})

			// Get breadcrumbs again
			breadcrumbs2 := GetBreadcrumbs()

			// Verify internal state unchanged
			require.Len(t, breadcrumbs2, 2)
			assert.Equal(t, "first", breadcrumbs2[0].Message)
			assert.Equal(t, "second", breadcrumbs2[1].Message)
		})
	}
}
