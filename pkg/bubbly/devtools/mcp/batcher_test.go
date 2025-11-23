package mcp

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpdateBatcher_NewUpdateBatcher tests batcher creation
func TestUpdateBatcher_NewUpdateBatcher(t *testing.T) {
	tests := []struct {
		name          string
		flushInterval time.Duration
		maxBatchSize  int
		wantErr       bool
	}{
		{
			name:          "valid configuration",
			flushInterval: 100 * time.Millisecond,
			maxBatchSize:  10,
			wantErr:       false,
		},
		{
			name:          "minimum values",
			flushInterval: 1 * time.Millisecond,
			maxBatchSize:  1,
			wantErr:       false,
		},
		{
			name:          "zero flush interval",
			flushInterval: 0,
			maxBatchSize:  10,
			wantErr:       true,
		},
		{
			name:          "zero batch size",
			flushInterval: 100 * time.Millisecond,
			maxBatchSize:  0,
			wantErr:       true,
		},
		{
			name:          "negative flush interval",
			flushInterval: -100 * time.Millisecond,
			maxBatchSize:  10,
			wantErr:       true,
		},
		{
			name:          "negative batch size",
			flushInterval: 100 * time.Millisecond,
			maxBatchSize:  -10,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batcher, err := NewUpdateBatcher(tt.flushInterval, tt.maxBatchSize)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, batcher)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, batcher)
				assert.Equal(t, tt.flushInterval, batcher.flushInterval)
				assert.Equal(t, tt.maxBatchSize, batcher.maxBatchSize)
			}
		})
	}
}

// TestUpdateBatcher_AddUpdate tests adding updates to batch
func TestUpdateBatcher_AddUpdate(t *testing.T) {
	batcher, err := NewUpdateBatcher(100*time.Millisecond, 5)
	require.NoError(t, err)

	// Mock flush handler
	var flushedUpdates []UpdateNotification
	var mu sync.Mutex
	flushHandler := func(clientID string, updates []UpdateNotification) {
		mu.Lock()
		defer mu.Unlock()
		flushedUpdates = append(flushedUpdates, updates...)
	}

	batcher.SetFlushHandler(flushHandler)

	// Add updates
	update1 := UpdateNotification{
		ClientID: "client-1",
		URI:      "bubblyui://state/refs",
		Data:     map[string]interface{}{"ref_id": "ref-1", "value": 42},
	}

	update2 := UpdateNotification{
		ClientID: "client-1",
		URI:      "bubblyui://state/refs",
		Data:     map[string]interface{}{"ref_id": "ref-2", "value": 43},
	}

	batcher.AddUpdate(update1)
	batcher.AddUpdate(update2)

	// Verify updates are batched (not flushed yet)
	mu.Lock()
	assert.Empty(t, flushedUpdates)
	mu.Unlock()

	// Wait for flush interval
	time.Sleep(150 * time.Millisecond)

	// Verify updates were flushed
	mu.Lock()
	assert.Len(t, flushedUpdates, 2)
	mu.Unlock()
}

// TestUpdateBatcher_FlushOnBatchSize tests flushing when batch size limit is reached
func TestUpdateBatcher_FlushOnBatchSize(t *testing.T) {
	batcher, err := NewUpdateBatcher(1*time.Second, 3) // Long interval, small batch
	require.NoError(t, err)

	var flushedUpdates []UpdateNotification
	var mu sync.Mutex
	flushHandler := func(clientID string, updates []UpdateNotification) {
		mu.Lock()
		defer mu.Unlock()
		flushedUpdates = append(flushedUpdates, updates...)
	}

	batcher.SetFlushHandler(flushHandler)

	// Add 3 updates (should trigger flush immediately)
	for i := 0; i < 3; i++ {
		update := UpdateNotification{
			ClientID: "client-1",
			URI:      "bubblyui://state/refs",
			Data:     map[string]interface{}{"ref_id": "ref-1", "value": i},
		}
		batcher.AddUpdate(update)
	}

	// Small delay for flush to complete
	time.Sleep(50 * time.Millisecond)

	// Verify all 3 updates were flushed (before interval elapsed)
	mu.Lock()
	assert.Len(t, flushedUpdates, 3)
	mu.Unlock()
}

// TestUpdateBatcher_PerClientBatching tests that updates are batched per client
func TestUpdateBatcher_PerClientBatching(t *testing.T) {
	batcher, err := NewUpdateBatcher(100*time.Millisecond, 10)
	require.NoError(t, err)

	var flushedClients []string
	var mu sync.Mutex
	flushHandler := func(clientID string, updates []UpdateNotification) {
		mu.Lock()
		defer mu.Unlock()
		flushedClients = append(flushedClients, clientID)
	}

	batcher.SetFlushHandler(flushHandler)

	// Add updates for different clients
	update1 := UpdateNotification{
		ClientID: "client-1",
		URI:      "bubblyui://state/refs",
		Data:     map[string]interface{}{"ref_id": "ref-1"},
	}

	update2 := UpdateNotification{
		ClientID: "client-2",
		URI:      "bubblyui://state/refs",
		Data:     map[string]interface{}{"ref_id": "ref-2"},
	}

	batcher.AddUpdate(update1)
	batcher.AddUpdate(update2)

	// Wait for flush
	time.Sleep(150 * time.Millisecond)

	// Verify both clients were flushed separately
	mu.Lock()
	assert.Contains(t, flushedClients, "client-1")
	assert.Contains(t, flushedClients, "client-2")
	mu.Unlock()
}

// TestUpdateBatcher_ConcurrentAccess tests thread-safe concurrent access
func TestUpdateBatcher_ConcurrentAccess(t *testing.T) {
	batcher, err := NewUpdateBatcher(50*time.Millisecond, 100)
	require.NoError(t, err)

	var flushedCount int
	var mu sync.Mutex
	flushHandler := func(clientID string, updates []UpdateNotification) {
		mu.Lock()
		defer mu.Unlock()
		flushedCount += len(updates)
	}

	batcher.SetFlushHandler(flushHandler)

	// Add updates concurrently from multiple goroutines
	const numGoroutines = 10
	const updatesPerGoroutine = 10
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(_ int) {
			defer wg.Done()
			for j := 0; j < updatesPerGoroutine; j++ {
				update := UpdateNotification{
					ClientID: "client-1",
					URI:      "bubblyui://state/refs",
					Data:     map[string]interface{}{"value": j},
				}
				batcher.AddUpdate(update)
			}
		}(i)
	}

	wg.Wait()

	// Wait for all flushes to complete
	time.Sleep(200 * time.Millisecond)

	// Verify all updates were flushed
	mu.Lock()
	assert.Equal(t, numGoroutines*updatesPerGoroutine, flushedCount)
	mu.Unlock()
}

// TestUpdateBatcher_Stop tests graceful shutdown
func TestUpdateBatcher_Stop(t *testing.T) {
	batcher, err := NewUpdateBatcher(1*time.Second, 10)
	require.NoError(t, err)

	var flushedUpdates []UpdateNotification
	var mu sync.Mutex
	flushHandler := func(clientID string, updates []UpdateNotification) {
		mu.Lock()
		defer mu.Unlock()
		flushedUpdates = append(flushedUpdates, updates...)
	}

	batcher.SetFlushHandler(flushHandler)

	// Add some updates
	for i := 0; i < 3; i++ {
		update := UpdateNotification{
			ClientID: "client-1",
			URI:      "bubblyui://state/refs",
			Data:     map[string]interface{}{"value": i},
		}
		batcher.AddUpdate(update)
	}

	// Stop batcher (should flush pending updates)
	batcher.Stop()

	// Verify pending updates were flushed on stop
	mu.Lock()
	assert.Len(t, flushedUpdates, 3)
	mu.Unlock()
}

// TestThrottler_NewThrottler tests throttler creation
func TestThrottler_NewThrottler(t *testing.T) {
	tests := []struct {
		name        string
		minInterval time.Duration
		wantErr     bool
	}{
		{
			name:        "valid configuration",
			minInterval: 100 * time.Millisecond,
			wantErr:     false,
		},
		{
			name:        "minimum interval",
			minInterval: 1 * time.Millisecond,
			wantErr:     false,
		},
		{
			name:        "zero interval",
			minInterval: 0,
			wantErr:     true,
		},
		{
			name:        "negative interval",
			minInterval: -100 * time.Millisecond,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			throttler, err := NewThrottler(tt.minInterval)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, throttler)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, throttler)
				assert.Equal(t, tt.minInterval, throttler.minInterval)
			}
		})
	}
}

// TestThrottler_ShouldSend tests throttling logic
func TestThrottler_ShouldSend(t *testing.T) {
	throttler, err := NewThrottler(100 * time.Millisecond)
	require.NoError(t, err)

	// First send should always be allowed
	assert.True(t, throttler.ShouldSend("client-1", "bubblyui://state/refs"))

	// Immediate second send should be throttled
	assert.False(t, throttler.ShouldSend("client-1", "bubblyui://state/refs"))

	// Wait for interval to pass
	time.Sleep(110 * time.Millisecond)

	// Now send should be allowed again
	assert.True(t, throttler.ShouldSend("client-1", "bubblyui://state/refs"))
}

// TestThrottler_PerClientThrottling tests that throttling is per-client
func TestThrottler_PerClientThrottling(t *testing.T) {
	throttler, err := NewThrottler(100 * time.Millisecond)
	require.NoError(t, err)

	// Client 1 sends
	assert.True(t, throttler.ShouldSend("client-1", "bubblyui://state/refs"))

	// Client 2 should not be throttled (different client)
	assert.True(t, throttler.ShouldSend("client-2", "bubblyui://state/refs"))

	// Client 1 immediate second send should be throttled
	assert.False(t, throttler.ShouldSend("client-1", "bubblyui://state/refs"))

	// Client 2 immediate second send should also be throttled
	assert.False(t, throttler.ShouldSend("client-2", "bubblyui://state/refs"))
}

// TestThrottler_PerResourceThrottling tests that throttling is per-resource
func TestThrottler_PerResourceThrottling(t *testing.T) {
	throttler, err := NewThrottler(100 * time.Millisecond)
	require.NoError(t, err)

	// Send to first resource
	assert.True(t, throttler.ShouldSend("client-1", "bubblyui://state/refs"))

	// Send to different resource should not be throttled
	assert.True(t, throttler.ShouldSend("client-1", "bubblyui://components"))

	// Second send to first resource should be throttled
	assert.False(t, throttler.ShouldSend("client-1", "bubblyui://state/refs"))

	// Second send to second resource should also be throttled
	assert.False(t, throttler.ShouldSend("client-1", "bubblyui://components"))
}

// TestThrottler_ConcurrentAccess tests thread-safe concurrent access
func TestThrottler_ConcurrentAccess(t *testing.T) {
	throttler, err := NewThrottler(10 * time.Millisecond)
	require.NoError(t, err)

	const numGoroutines = 10
	var wg sync.WaitGroup
	var allowedCount int
	var mu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(_ int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				if throttler.ShouldSend("client-1", "bubblyui://state/refs") {
					mu.Lock()
					allowedCount++
					mu.Unlock()
				}
				time.Sleep(5 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Verify that throttling occurred (not all requests were allowed)
	// With 10 goroutines * 10 attempts = 100 total attempts
	// With 10ms throttle and 5ms sleep, we expect significant throttling
	mu.Lock()
	assert.Less(t, allowedCount, 100, "throttling should have occurred")
	assert.Greater(t, allowedCount, 0, "some requests should have been allowed")
	mu.Unlock()
}

// TestThrottler_Reset tests resetting throttle state for a client
func TestThrottler_Reset(t *testing.T) {
	throttler, err := NewThrottler(100 * time.Millisecond)
	require.NoError(t, err)

	// First send
	assert.True(t, throttler.ShouldSend("client-1", "bubblyui://state/refs"))

	// Immediate second send should be throttled
	assert.False(t, throttler.ShouldSend("client-1", "bubblyui://state/refs"))

	// Reset client's throttle state
	throttler.Reset("client-1")

	// Now send should be allowed immediately (without waiting for interval)
	assert.True(t, throttler.ShouldSend("client-1", "bubblyui://state/refs"))
}
