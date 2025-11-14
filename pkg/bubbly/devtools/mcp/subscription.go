package mcp

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Subscription represents a client's subscription to a resource URI.
//
// Subscriptions enable real-time updates when the subscribed resource changes.
// Each subscription is identified by a unique ID and associated with a client.
//
// Thread Safety:
//
//	Subscription instances are immutable after creation.
//
// Example:
//
//	sub := &Subscription{
//	    ID:         "sub-123",
//	    ClientID:   "client-456",
//	    ResourceURI: "bubblyui://state/refs",
//	    Filters:    map[string]interface{}{"ref_id": "count-ref"},
//	    CreatedAt:  time.Now(),
//	}
type Subscription struct {
	// ID is the unique identifier for this subscription
	ID string

	// ClientID identifies the client that owns this subscription
	ClientID string

	// ResourceURI is the MCP resource URI being subscribed to
	// Examples: "bubblyui://components", "bubblyui://state/refs"
	ResourceURI string

	// Filters are optional criteria for filtering updates
	// Keys and values depend on the resource type
	// Example: {"ref_id": "count-ref"} to only receive updates for a specific ref
	Filters map[string]interface{}

	// CreatedAt is when the subscription was created
	CreatedAt time.Time
}

// SubscriptionManager manages client subscriptions to MCP resources.
//
// It maintains a registry of active subscriptions, enforces limits,
// prevents duplicates, and provides cleanup on client disconnect.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Lifecycle:
//
//  1. NewSubscriptionManager() - Creates the manager
//  2. Subscribe() - Clients add subscriptions
//  3. ... updates are sent via change detectors (Task 4.2) ...
//  4. Unsubscribe() or UnsubscribeAll() - Cleanup
//
// Example:
//
//	sm := NewSubscriptionManager(50) // Max 50 subscriptions per client
//	err := sm.Subscribe("client-1", "bubblyui://state/refs", nil)
//	if err != nil {
//	    log.Printf("Subscribe failed: %v", err)
//	}
type SubscriptionManager struct {
	// subscriptions maps client IDs to their active subscriptions
	// Key: client ID, Value: slice of subscriptions for that client
	subscriptions map[string][]*Subscription

	// maxPerClient is the maximum number of subscriptions allowed per client
	// Default: 50 (from requirements)
	maxPerClient int

	// mu protects concurrent access to subscriptions map
	mu sync.RWMutex
}

// NewSubscriptionManager creates a new subscription manager.
//
// The maxPerClient parameter sets the maximum number of subscriptions
// a single client can have. This prevents resource exhaustion from
// subscription spam.
//
// Thread Safety:
//
//	Safe to call concurrently (creates new instance each time).
//
// Example:
//
//	sm := NewSubscriptionManager(50)
//
// Parameters:
//   - maxPerClient: Maximum subscriptions per client (typically 50)
//
// Returns:
//   - *SubscriptionManager: A new subscription manager instance
func NewSubscriptionManager(maxPerClient int) *SubscriptionManager {
	return &SubscriptionManager{
		subscriptions: make(map[string][]*Subscription),
		maxPerClient:  maxPerClient,
	}
}

// Subscribe adds a new subscription for a client.
//
// This method:
//   - Validates the subscription doesn't already exist (duplicate prevention)
//   - Enforces the per-client subscription limit
//   - Generates a unique subscription ID
//   - Adds the subscription to the registry
//
// Duplicate Detection:
//
//	A subscription is considered duplicate if the same client is already
//	subscribed to the same resource URI with the same filters.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	err := sm.Subscribe("client-1", "bubblyui://state/refs", map[string]interface{}{
//	    "ref_id": "count-ref",
//	})
//	if err != nil {
//	    log.Printf("Subscribe failed: %v", err)
//	}
//
// Parameters:
//   - clientID: Unique identifier for the client
//   - uri: MCP resource URI to subscribe to
//   - filters: Optional filters for updates (can be nil)
//
// Returns:
//   - error: Validation error, limit exceeded, or nil on success
func (sm *SubscriptionManager) Subscribe(clientID, uri string, filters map[string]interface{}) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Validate inputs
	if clientID == "" {
		return fmt.Errorf("clientID cannot be empty")
	}
	if uri == "" {
		return fmt.Errorf("uri cannot be empty")
	}

	// Get existing subscriptions for this client
	clientSubs := sm.subscriptions[clientID]

	// Check subscription limit
	if len(clientSubs) >= sm.maxPerClient {
		return fmt.Errorf("subscription limit exceeded: client has %d subscriptions (max %d)", len(clientSubs), sm.maxPerClient)
	}

	// Check for duplicate subscription
	for _, sub := range clientSubs {
		if sub.ResourceURI == uri && filtersEqual(sub.Filters, filters) {
			return fmt.Errorf("duplicate subscription: client already subscribed to %s with same filters", uri)
		}
	}

	// Generate unique subscription ID
	subID := uuid.New().String()

	// Create subscription
	sub := &Subscription{
		ID:          subID,
		ClientID:    clientID,
		ResourceURI: uri,
		Filters:     filters,
		CreatedAt:   time.Now(),
	}

	// Add to registry
	sm.subscriptions[clientID] = append(clientSubs, sub)

	return nil
}

// Unsubscribe removes a specific subscription.
//
// This method:
//   - Finds the subscription by ID
//   - Removes it from the client's subscription list
//   - Cleans up empty client entries
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	err := sm.Unsubscribe("client-1", "sub-123")
//	if err != nil {
//	    log.Printf("Unsubscribe failed: %v", err)
//	}
//
// Parameters:
//   - clientID: Unique identifier for the client
//   - subscriptionID: ID of the subscription to remove
//
// Returns:
//   - error: Subscription not found, or nil on success
func (sm *SubscriptionManager) Unsubscribe(clientID, subscriptionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Validate inputs
	if clientID == "" {
		return fmt.Errorf("clientID cannot be empty")
	}
	if subscriptionID == "" {
		return fmt.Errorf("subscriptionID cannot be empty")
	}

	// Get client's subscriptions
	clientSubs, exists := sm.subscriptions[clientID]
	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	// Find and remove subscription
	for i, sub := range clientSubs {
		if sub.ID == subscriptionID {
			// Remove subscription by replacing with last element and truncating
			clientSubs[i] = clientSubs[len(clientSubs)-1]
			clientSubs = clientSubs[:len(clientSubs)-1]

			// Update map
			if len(clientSubs) == 0 {
				// Clean up empty client entry
				delete(sm.subscriptions, clientID)
			} else {
				sm.subscriptions[clientID] = clientSubs
			}

			return nil
		}
	}

	return fmt.Errorf("subscription not found: %s", subscriptionID)
}

// UnsubscribeAll removes all subscriptions for a client.
//
// This method is typically called when a client disconnects.
// It performs bulk cleanup of all the client's subscriptions.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	err := sm.UnsubscribeAll("client-1")
//	if err != nil {
//	    log.Printf("UnsubscribeAll failed: %v", err)
//	}
//
// Parameters:
//   - clientID: Unique identifier for the client
//
// Returns:
//   - error: Client not found, or nil on success
func (sm *SubscriptionManager) UnsubscribeAll(clientID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Validate input
	if clientID == "" {
		return fmt.Errorf("clientID cannot be empty")
	}

	// Check if client exists
	if _, exists := sm.subscriptions[clientID]; !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	// Remove all subscriptions for this client
	delete(sm.subscriptions, clientID)

	return nil
}

// GetSubscriptions returns all subscriptions for a client.
//
// This method returns a copy of the subscriptions slice to prevent
// external modification of the internal state.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	subs := sm.GetSubscriptions("client-1")
//	fmt.Printf("Client has %d subscriptions\n", len(subs))
//
// Parameters:
//   - clientID: Unique identifier for the client
//
// Returns:
//   - []*Subscription: Copy of client's subscriptions (empty slice if none)
func (sm *SubscriptionManager) GetSubscriptions(clientID string) []*Subscription {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	clientSubs, exists := sm.subscriptions[clientID]
	if !exists {
		return []*Subscription{}
	}

	// Return a copy to prevent external modification
	result := make([]*Subscription, len(clientSubs))
	copy(result, clientSubs)
	return result
}

// GetSubscriptionCount returns the total number of subscriptions for a client.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	count := sm.GetSubscriptionCount("client-1")
//	fmt.Printf("Client has %d subscriptions\n", count)
//
// Parameters:
//   - clientID: Unique identifier for the client
//
// Returns:
//   - int: Number of active subscriptions (0 if client not found)
func (sm *SubscriptionManager) GetSubscriptionCount(clientID string) int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	clientSubs, exists := sm.subscriptions[clientID]
	if !exists {
		return 0
	}

	return len(clientSubs)
}

// filtersEqual compares two filter maps for equality.
//
// Returns true if both maps have the same keys and values.
// Nil maps are considered equal to empty maps.
func filtersEqual(a, b map[string]interface{}) bool {
	// Nil maps are equal to empty maps
	if len(a) == 0 && len(b) == 0 {
		return true
	}

	// Different lengths means not equal
	if len(a) != len(b) {
		return false
	}

	// Compare each key-value pair
	for key, valA := range a {
		valB, exists := b[key]
		if !exists {
			return false
		}

		// Simple equality check (works for basic types)
		// For complex types, would need deep comparison
		if valA != valB {
			return false
		}
	}

	return true
}
