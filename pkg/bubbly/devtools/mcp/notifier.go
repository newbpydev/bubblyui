package mcp

import (
	"fmt"
	"sync"
)

// NotificationSender sends resource update notifications to MCP clients.
//
// It integrates with the UpdateBatcher to queue notifications for batching
// and throttling, preventing client overload from high-frequency updates.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Lifecycle:
//
//  1. NewNotificationSender() - Creates the sender with a batcher
//  2. QueueNotification() - Queue notifications for batching
//  3. ... Batcher handles flushing and sending ...
//
// Example:
//
//	batcher, err := NewUpdateBatcher(100*time.Millisecond, 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	notifier, err := NewNotificationSender(batcher)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Queue a notification
//	notifier.QueueNotification("client-1", "bubblyui://state/refs", map[string]interface{}{
//	    "ref_id": "count-ref",
//	    "value":  42,
//	})
type NotificationSender struct {
	// batcher handles batching and throttling of notifications
	batcher *UpdateBatcher

	// mu protects concurrent access to the sender
	// Currently minimal state, but prepared for future additions
	mu sync.RWMutex
}

// NewNotificationSender creates a new notification sender.
//
// The sender uses the provided batcher to queue notifications for
// batching and throttling. The batcher must have a flush handler
// configured to actually send the notifications.
//
// Thread Safety:
//
//	Safe to call concurrently (creates new instance each time).
//
// Example:
//
//	batcher, err := NewUpdateBatcher(100*time.Millisecond, 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	notifier, err := NewNotificationSender(batcher)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - batcher: The update batcher to use for queuing notifications
//
// Returns:
//   - *NotificationSender: A new notification sender instance
//   - error: Validation error, or nil on success
func NewNotificationSender(batcher *UpdateBatcher) (*NotificationSender, error) {
	// Validate input
	if batcher == nil {
		return nil, fmt.Errorf("batcher cannot be nil")
	}

	return &NotificationSender{
		batcher: batcher,
	}, nil
}

// QueueNotification queues a notification for batching and sending.
//
// The notification will be added to the batcher, which will flush it
// either after the flush interval or when the batch size is reached.
//
// This method does not block - it queues the notification and returns
// immediately. The actual sending is handled asynchronously by the batcher.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	notifier.QueueNotification("client-1", "bubblyui://state/refs", map[string]interface{}{
//	    "ref_id": "count-ref",
//	    "value":  42,
//	})
//
// Parameters:
//   - clientID: The client to send the notification to
//   - uri: The resource URI that changed
//   - data: The notification payload (can be nil)
func (n *NotificationSender) QueueNotification(clientID, uri string, data map[string]interface{}) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Create update notification
	update := UpdateNotification{
		ClientID: clientID,
		URI:      uri,
		Data:     data,
	}

	// Queue in batcher
	n.batcher.AddUpdate(update)
}
