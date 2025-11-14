package mcp

import (
	"fmt"
	"sync"
	"time"
)

// UpdateNotification represents a notification to be sent to a client.
//
// Notifications are batched and throttled to prevent client overload.
//
// Thread Safety:
//
//	UpdateNotification instances are immutable after creation.
//
// Example:
//
//	notification := UpdateNotification{
//	    ClientID: "client-1",
//	    URI:      "bubblyui://state/refs",
//	    Data:     map[string]interface{}{"ref_id": "ref-1", "value": 42},
//	}
type UpdateNotification struct {
	// ClientID identifies the client to send the notification to
	ClientID string

	// URI is the resource URI that changed
	URI string

	// Data contains the notification payload
	Data map[string]interface{}
}

// FlushHandler is called when a batch of updates is ready to be sent.
//
// The handler receives the client ID and a slice of updates to send.
// It should send the updates to the client via the MCP protocol.
//
// Thread Safety:
//
//	Handlers may be called concurrently for different clients.
//
// Example:
//
//	handler := func(clientID string, updates []UpdateNotification) {
//	    for _, update := range updates {
//	        // Send update to client via MCP
//	        server.SendNotification(clientID, update)
//	    }
//	}
type FlushHandler func(clientID string, updates []UpdateNotification)

// UpdateBatcher batches updates and flushes them periodically or when batch size is reached.
//
// It collects updates for each client and flushes them either:
//   - After flushInterval has elapsed since the last flush
//   - When the batch reaches maxBatchSize updates
//
// This prevents overwhelming clients with high-frequency updates.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Lifecycle:
//
//  1. NewUpdateBatcher() - Creates the batcher
//  2. SetFlushHandler() - Configures the flush handler
//  3. AddUpdate() - Add updates to batch
//  4. ... automatic flushing occurs ...
//  5. Stop() - Graceful shutdown (flushes pending updates)
//
// Example:
//
//	batcher, err := NewUpdateBatcher(100*time.Millisecond, 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	batcher.SetFlushHandler(func(clientID string, updates []UpdateNotification) {
//	    // Send updates to client
//	})
//	batcher.AddUpdate(notification)
//	defer batcher.Stop()
type UpdateBatcher struct {
	// pendingUpdates maps client IDs to their pending updates
	// Key: client ID, Value: slice of pending updates
	pendingUpdates map[string][]UpdateNotification

	// flushInterval is the maximum time to wait before flushing
	flushInterval time.Duration

	// maxBatchSize is the maximum number of updates per batch
	maxBatchSize int

	// flushHandler is called when a batch is ready to be sent
	flushHandler FlushHandler

	// ticker triggers periodic flushes
	ticker *time.Ticker

	// stopChan signals the flush goroutine to stop
	stopChan chan struct{}

	// wg tracks the flush goroutine
	wg sync.WaitGroup

	// mu protects concurrent access to pendingUpdates
	mu sync.Mutex
}

// NewUpdateBatcher creates a new update batcher.
//
// The batcher will flush updates either:
//   - After flushInterval has elapsed
//   - When a client's batch reaches maxBatchSize
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
// Parameters:
//   - flushInterval: Maximum time between flushes (must be > 0)
//   - maxBatchSize: Maximum updates per batch (must be > 0)
//
// Returns:
//   - *UpdateBatcher: A new batcher instance
//   - error: Validation error, or nil on success
func NewUpdateBatcher(flushInterval time.Duration, maxBatchSize int) (*UpdateBatcher, error) {
	// Validate parameters
	if flushInterval <= 0 {
		return nil, fmt.Errorf("flushInterval must be positive, got %v", flushInterval)
	}
	if maxBatchSize <= 0 {
		return nil, fmt.Errorf("maxBatchSize must be positive, got %d", maxBatchSize)
	}

	batcher := &UpdateBatcher{
		pendingUpdates: make(map[string][]UpdateNotification),
		flushInterval:  flushInterval,
		maxBatchSize:   maxBatchSize,
		ticker:         time.NewTicker(flushInterval),
		stopChan:       make(chan struct{}),
	}

	// Start flush goroutine
	batcher.wg.Add(1)
	go batcher.flushLoop()

	return batcher, nil
}

// SetFlushHandler sets the handler to call when flushing updates.
//
// This must be called before adding any updates.
//
// Thread Safety:
//
//	Not safe to call concurrently with AddUpdate().
//	Should be called once during initialization.
//
// Example:
//
//	batcher.SetFlushHandler(func(clientID string, updates []UpdateNotification) {
//	    for _, update := range updates {
//	        server.SendNotification(clientID, update)
//	    }
//	})
//
// Parameters:
//   - handler: The flush handler function
func (b *UpdateBatcher) SetFlushHandler(handler FlushHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flushHandler = handler
}

// AddUpdate adds an update to the batch.
//
// The update will be flushed either:
//   - After flushInterval has elapsed
//   - When the client's batch reaches maxBatchSize (immediate flush)
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	update := UpdateNotification{
//	    ClientID: "client-1",
//	    URI:      "bubblyui://state/refs",
//	    Data:     map[string]interface{}{"ref_id": "ref-1", "value": 42},
//	}
//	batcher.AddUpdate(update)
//
// Parameters:
//   - update: The notification to add to the batch
func (b *UpdateBatcher) AddUpdate(update UpdateNotification) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Add update to client's pending batch
	clientID := update.ClientID
	b.pendingUpdates[clientID] = append(b.pendingUpdates[clientID], update)

	// Check if batch size limit reached
	if len(b.pendingUpdates[clientID]) >= b.maxBatchSize {
		// Flush immediately
		b.flushClient(clientID)
	}
}

// Stop gracefully shuts down the batcher.
//
// This method:
//   - Stops the flush timer
//   - Flushes all pending updates
//   - Waits for the flush goroutine to complete
//
// Thread Safety:
//
//	Safe to call concurrently, but should only be called once.
//
// Example:
//
//	defer batcher.Stop()
//
// After calling Stop(), no more updates should be added.
func (b *UpdateBatcher) Stop() {
	// Signal stop
	close(b.stopChan)

	// Stop ticker
	b.ticker.Stop()

	// Wait for flush goroutine to complete
	b.wg.Wait()

	// Final flush of any pending updates
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flushAll()
}

// flushLoop runs in a goroutine and flushes updates periodically.
//
// It listens for:
//   - Ticker events (periodic flush)
//   - Stop signal (graceful shutdown)
//
// Thread Safety:
//
//	Runs in its own goroutine, uses mutex for shared state.
func (b *UpdateBatcher) flushLoop() {
	defer b.wg.Done()

	for {
		select {
		case <-b.ticker.C:
			// Periodic flush
			b.mu.Lock()
			b.flushAll()
			b.mu.Unlock()

		case <-b.stopChan:
			// Stop signal received
			return
		}
	}
}

// flushAll flushes all pending updates for all clients.
//
// Thread Safety:
//
//	Must be called with b.mu locked.
func (b *UpdateBatcher) flushAll() {
	for clientID := range b.pendingUpdates {
		b.flushClient(clientID)
	}
}

// flushClient flushes pending updates for a specific client.
//
// Thread Safety:
//
//	Must be called with b.mu locked.
//
// Parameters:
//   - clientID: The client to flush updates for
func (b *UpdateBatcher) flushClient(clientID string) {
	updates := b.pendingUpdates[clientID]
	if len(updates) == 0 {
		return
	}

	// Clear pending updates
	delete(b.pendingUpdates, clientID)

	// Call flush handler (if set)
	if b.flushHandler != nil {
		// Make a copy to avoid holding the lock during handler call
		updatesCopy := make([]UpdateNotification, len(updates))
		copy(updatesCopy, updates)

		// Release lock before calling handler
		b.mu.Unlock()
		b.flushHandler(clientID, updatesCopy)
		b.mu.Lock()
	}
}

// Throttler prevents sending updates too frequently to clients.
//
// It enforces a minimum interval between updates for each client+resource combination.
// This prevents overwhelming clients with high-frequency changes.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Lifecycle:
//
//  1. NewThrottler() - Creates the throttler
//  2. ShouldSend() - Check if update should be sent
//  3. ... automatic throttling occurs ...
//  4. Reset() - Optional: reset throttle state for a client
//
// Example:
//
//	throttler, err := NewThrottler(100*time.Millisecond)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if throttler.ShouldSend("client-1", "bubblyui://state/refs") {
//	    // Send update
//	}
type Throttler struct {
	// lastSent tracks the last send time for each client+resource
	// Key: "clientID:resourceURI", Value: last send time
	lastSent map[string]time.Time

	// minInterval is the minimum time between sends
	minInterval time.Duration

	// mu protects concurrent access to lastSent map
	mu sync.RWMutex
}

// NewThrottler creates a new throttler.
//
// The throttler will enforce a minimum interval between updates
// for each client+resource combination.
//
// Thread Safety:
//
//	Safe to call concurrently (creates new instance each time).
//
// Example:
//
//	throttler, err := NewThrottler(100*time.Millisecond)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - minInterval: Minimum time between updates (must be > 0)
//
// Returns:
//   - *Throttler: A new throttler instance
//   - error: Validation error, or nil on success
func NewThrottler(minInterval time.Duration) (*Throttler, error) {
	// Validate parameter
	if minInterval <= 0 {
		return nil, fmt.Errorf("minInterval must be positive, got %v", minInterval)
	}

	return &Throttler{
		lastSent:    make(map[string]time.Time),
		minInterval: minInterval,
	}, nil
}

// ShouldSend checks if an update should be sent to a client.
//
// Returns true if:
//   - This is the first update for this client+resource
//   - minInterval has elapsed since the last update
//
// Returns false if:
//   - minInterval has not yet elapsed (throttled)
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	if throttler.ShouldSend("client-1", "bubblyui://state/refs") {
//	    // Send update
//	    server.SendNotification(clientID, update)
//	}
//
// Parameters:
//   - clientID: The client to check
//   - resourceURI: The resource URI to check
//
// Returns:
//   - bool: True if update should be sent, false if throttled
func (t *Throttler) ShouldSend(clientID, resourceURI string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Create key for this client+resource combination
	key := clientID + ":" + resourceURI

	// Get last send time
	lastSendTime, exists := t.lastSent[key]

	// If never sent, allow
	if !exists {
		t.lastSent[key] = time.Now()
		return true
	}

	// Check if enough time has elapsed
	elapsed := time.Since(lastSendTime)
	if elapsed >= t.minInterval {
		// Update last send time
		t.lastSent[key] = time.Now()
		return true
	}

	// Throttled
	return false
}

// Reset clears the throttle state for a client.
//
// This allows the next update to be sent immediately,
// regardless of when the last update was sent.
//
// Typically called when a client disconnects or reconnects.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	throttler.Reset("client-1")
//
// Parameters:
//   - clientID: The client to reset throttle state for
func (t *Throttler) Reset(clientID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Remove all entries for this client
	for key := range t.lastSent {
		// Key format: "clientID:resourceURI"
		// Check if key starts with clientID
		if len(key) > len(clientID) && key[:len(clientID)] == clientID && key[len(clientID)] == ':' {
			delete(t.lastSent, key)
		}
	}
}
