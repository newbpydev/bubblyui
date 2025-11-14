package mcp

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNotificationSender_Creation tests NotificationSender creation and initialization.
func TestNotificationSender_Creation(t *testing.T) {
	tests := []struct {
		name        string
		batcher     *UpdateBatcher
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid batcher",
			batcher: mustCreateBatcher(t, 100*time.Millisecond, 10),
			wantErr: false,
		},
		{
			name:        "nil batcher",
			batcher:     nil,
			wantErr:     true,
			errContains: "batcher cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notifier, err := NewNotificationSender(tt.batcher)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, notifier)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, notifier)
			}
		})
	}
}

// TestNotificationSender_QueueNotification tests queuing notifications for batching.
func TestNotificationSender_QueueNotification(t *testing.T) {
	tests := []struct {
		name         string
		clientID     string
		uri          string
		data         map[string]interface{}
		expectQueued bool
	}{
		{
			name:     "valid notification",
			clientID: "client-1",
			uri:      "bubblyui://state/refs",
			data: map[string]interface{}{
				"ref_id": "count-ref",
				"value":  42,
			},
			expectQueued: true,
		},
		{
			name:         "empty data",
			clientID:     "client-2",
			uri:          "bubblyui://components",
			data:         map[string]interface{}{},
			expectQueued: true,
		},
		{
			name:         "nil data",
			clientID:     "client-3",
			uri:          "bubblyui://events/log",
			data:         nil,
			expectQueued: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create batcher with handler that tracks queued notifications
			var queuedNotifications []UpdateNotification
			var mu sync.Mutex

			batcher := mustCreateBatcher(t, 100*time.Millisecond, 10)
			batcher.SetFlushHandler(func(clientID string, updates []UpdateNotification) {
				mu.Lock()
				defer mu.Unlock()
				queuedNotifications = append(queuedNotifications, updates...)
			})

			notifier, err := NewNotificationSender(batcher)
			require.NoError(t, err)

			// Queue notification
			notifier.QueueNotification(tt.clientID, tt.uri, tt.data)

			// Wait for batch to flush
			time.Sleep(150 * time.Millisecond)

			// Verify notification was queued
			mu.Lock()
			defer mu.Unlock()

			if tt.expectQueued {
				assert.NotEmpty(t, queuedNotifications, "Expected notification to be queued")

				// Find our notification
				found := false
				for _, n := range queuedNotifications {
					if n.ClientID == tt.clientID && n.URI == tt.uri {
						found = true
						assert.Equal(t, tt.data, n.Data)
						break
					}
				}
				assert.True(t, found, "Expected to find queued notification")
			}
		})
	}
}

// TestNotificationSender_ConcurrentQueue tests concurrent notification queuing.
func TestNotificationSender_ConcurrentQueue(t *testing.T) {
	batcher := mustCreateBatcher(t, 100*time.Millisecond, 100)

	var receivedCount int
	var mu sync.Mutex

	batcher.SetFlushHandler(func(clientID string, updates []UpdateNotification) {
		mu.Lock()
		defer mu.Unlock()
		receivedCount += len(updates)
	})

	notifier, err := NewNotificationSender(batcher)
	require.NoError(t, err)

	// Queue notifications concurrently
	const numGoroutines = 10
	const notificationsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < notificationsPerGoroutine; j++ {
				notifier.QueueNotification(
					"client-1",
					"bubblyui://state/refs",
					map[string]interface{}{
						"goroutine": goroutineID,
						"index":     j,
					},
				)
			}
		}(i)
	}

	wg.Wait()

	// Wait for all batches to flush
	time.Sleep(200 * time.Millisecond)

	// Verify all notifications were received
	mu.Lock()
	defer mu.Unlock()

	expectedCount := numGoroutines * notificationsPerGoroutine
	assert.Equal(t, expectedCount, receivedCount, "All notifications should be received")
}

// TestNotificationSender_MultipleClients tests notifications to multiple clients.
func TestNotificationSender_MultipleClients(t *testing.T) {
	batcher := mustCreateBatcher(t, 100*time.Millisecond, 10)

	clientNotifications := make(map[string]int)
	var mu sync.Mutex

	batcher.SetFlushHandler(func(clientID string, updates []UpdateNotification) {
		mu.Lock()
		defer mu.Unlock()
		clientNotifications[clientID] += len(updates)
	})

	notifier, err := NewNotificationSender(batcher)
	require.NoError(t, err)

	// Queue notifications for different clients
	clients := []string{"client-1", "client-2", "client-3"}
	for _, clientID := range clients {
		for i := 0; i < 5; i++ {
			notifier.QueueNotification(
				clientID,
				"bubblyui://state/refs",
				map[string]interface{}{"index": i},
			)
		}
	}

	// Wait for batches to flush
	time.Sleep(150 * time.Millisecond)

	// Verify each client received their notifications
	mu.Lock()
	defer mu.Unlock()

	for _, clientID := range clients {
		assert.Equal(t, 5, clientNotifications[clientID],
			"Client %s should receive 5 notifications", clientID)
	}
}

// TestNotificationSender_ThreadSafety tests thread-safe concurrent access.
func TestNotificationSender_ThreadSafety(t *testing.T) {
	batcher := mustCreateBatcher(t, 50*time.Millisecond, 50)
	batcher.SetFlushHandler(func(clientID string, updates []UpdateNotification) {
		// Handler does nothing, just testing thread safety
	})

	notifier, err := NewNotificationSender(batcher)
	require.NoError(t, err)

	// Run concurrent operations
	const numGoroutines = 20
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < 50; j++ {
				notifier.QueueNotification(
					"client-1",
					"bubblyui://state/refs",
					map[string]interface{}{"id": id, "index": j},
				)
			}
		}(i)
	}

	wg.Wait()

	// If we get here without race detector errors, thread safety is verified
}

// mustCreateBatcher is a test helper that creates a batcher or fails the test.
func mustCreateBatcher(t *testing.T, flushInterval time.Duration, maxBatchSize int) *UpdateBatcher {
	t.Helper()
	batcher, err := NewUpdateBatcher(flushInterval, maxBatchSize)
	require.NoError(t, err)
	t.Cleanup(func() {
		batcher.Stop()
	})
	return batcher
}
