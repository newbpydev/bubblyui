package mcp

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSubscriptionManager tests subscription manager creation
func TestNewSubscriptionManager(t *testing.T) {
	sm := NewSubscriptionManager(50)

	assert.NotNil(t, sm)
	assert.NotNil(t, sm.subscriptions)
	assert.Equal(t, 50, sm.maxPerClient)
}

// TestSubscriptionManager_Subscribe tests basic subscription functionality
func TestSubscriptionManager_Subscribe(t *testing.T) {
	tests := []struct {
		name     string
		clientID string
		uri      string
		filters  map[string]interface{}
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid subscription",
			clientID: "client-1",
			uri:      "bubblyui://state/refs",
			filters:  nil,
			wantErr:  false,
		},
		{
			name:     "valid subscription with filters",
			clientID: "client-2",
			uri:      "bubblyui://state/refs",
			filters:  map[string]interface{}{"ref_id": "count-ref"},
			wantErr:  false,
		},
		{
			name:     "empty client ID",
			clientID: "",
			uri:      "bubblyui://state/refs",
			filters:  nil,
			wantErr:  true,
			errMsg:   "clientID cannot be empty",
		},
		{
			name:     "empty URI",
			clientID: "client-3",
			uri:      "",
			filters:  nil,
			wantErr:  true,
			errMsg:   "uri cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSubscriptionManager(50)

			err := sm.Subscribe(tt.clientID, tt.uri, tt.filters)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)

				// Verify subscription was added
				subs := sm.GetSubscriptions(tt.clientID)
				assert.Len(t, subs, 1)
				assert.Equal(t, tt.clientID, subs[0].ClientID)
				assert.Equal(t, tt.uri, subs[0].ResourceURI)
				assert.NotEmpty(t, subs[0].ID)
				assert.False(t, subs[0].CreatedAt.IsZero())
			}
		})
	}
}

// TestSubscriptionManager_Subscribe_DuplicatePrevention tests duplicate detection
func TestSubscriptionManager_Subscribe_DuplicatePrevention(t *testing.T) {
	sm := NewSubscriptionManager(50)
	clientID := "client-1"
	uri := "bubblyui://state/refs"
	filters := map[string]interface{}{"ref_id": "count-ref"}

	// First subscription should succeed
	err := sm.Subscribe(clientID, uri, filters)
	require.NoError(t, err)

	// Duplicate subscription should fail
	err = sm.Subscribe(clientID, uri, filters)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate subscription")

	// Different filters should succeed
	err = sm.Subscribe(clientID, uri, map[string]interface{}{"ref_id": "other-ref"})
	require.NoError(t, err)

	// Different URI should succeed
	err = sm.Subscribe(clientID, "bubblyui://components", filters)
	require.NoError(t, err)

	// Verify count
	assert.Equal(t, 3, sm.GetSubscriptionCount(clientID))
}

// TestSubscriptionManager_Subscribe_LimitEnforcement tests subscription limits
func TestSubscriptionManager_Subscribe_LimitEnforcement(t *testing.T) {
	maxPerClient := 5
	sm := NewSubscriptionManager(maxPerClient)
	clientID := "client-1"

	// Add subscriptions up to limit
	for i := 0; i < maxPerClient; i++ {
		err := sm.Subscribe(clientID, "bubblyui://state/refs", map[string]interface{}{"index": i})
		require.NoError(t, err)
	}

	// Verify count
	assert.Equal(t, maxPerClient, sm.GetSubscriptionCount(clientID))

	// Exceeding limit should fail
	err := sm.Subscribe(clientID, "bubblyui://components", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "subscription limit exceeded")
	assert.Contains(t, err.Error(), "5 subscriptions")
}

// TestSubscriptionManager_Unsubscribe tests removing specific subscriptions
func TestSubscriptionManager_Unsubscribe(t *testing.T) {
	sm := NewSubscriptionManager(50)
	clientID := "client-1"

	// Add subscriptions
	err := sm.Subscribe(clientID, "bubblyui://state/refs", nil)
	require.NoError(t, err)
	err = sm.Subscribe(clientID, "bubblyui://components", nil)
	require.NoError(t, err)

	subs := sm.GetSubscriptions(clientID)
	require.Len(t, subs, 2)

	// Unsubscribe first subscription
	err = sm.Unsubscribe(clientID, subs[0].ID)
	require.NoError(t, err)

	// Verify one subscription remains
	subs = sm.GetSubscriptions(clientID)
	assert.Len(t, subs, 1)

	// Unsubscribe second subscription
	err = sm.Unsubscribe(clientID, subs[0].ID)
	require.NoError(t, err)

	// Verify no subscriptions remain
	subs = sm.GetSubscriptions(clientID)
	assert.Len(t, subs, 0)
}

// TestSubscriptionManager_Unsubscribe_Errors tests unsubscribe error cases
func TestSubscriptionManager_Unsubscribe_Errors(t *testing.T) {
	tests := []struct {
		name           string
		clientID       string
		subscriptionID string
		setup          func(*SubscriptionManager)
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "empty client ID",
			clientID:       "",
			subscriptionID: "sub-123",
			setup:          func(sm *SubscriptionManager) {},
			wantErr:        true,
			errMsg:         "clientID cannot be empty",
		},
		{
			name:           "empty subscription ID",
			clientID:       "client-1",
			subscriptionID: "",
			setup:          func(sm *SubscriptionManager) {},
			wantErr:        true,
			errMsg:         "subscriptionID cannot be empty",
		},
		{
			name:           "client not found",
			clientID:       "client-1",
			subscriptionID: "sub-123",
			setup:          func(sm *SubscriptionManager) {},
			wantErr:        true,
			errMsg:         "client not found",
		},
		{
			name:           "subscription not found",
			clientID:       "client-1",
			subscriptionID: "sub-nonexistent",
			setup: func(sm *SubscriptionManager) {
				sm.Subscribe("client-1", "bubblyui://state/refs", nil)
			},
			wantErr: true,
			errMsg:  "subscription not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSubscriptionManager(50)
			tt.setup(sm)

			err := sm.Unsubscribe(tt.clientID, tt.subscriptionID)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

// TestSubscriptionManager_UnsubscribeAll tests bulk unsubscribe
func TestSubscriptionManager_UnsubscribeAll(t *testing.T) {
	sm := NewSubscriptionManager(50)
	clientID := "client-1"

	// Add multiple subscriptions
	for i := 0; i < 5; i++ {
		err := sm.Subscribe(clientID, "bubblyui://state/refs", map[string]interface{}{"index": i})
		require.NoError(t, err)
	}

	// Verify subscriptions exist
	assert.Equal(t, 5, sm.GetSubscriptionCount(clientID))

	// Unsubscribe all
	err := sm.UnsubscribeAll(clientID)
	require.NoError(t, err)

	// Verify all subscriptions removed
	assert.Equal(t, 0, sm.GetSubscriptionCount(clientID))
	subs := sm.GetSubscriptions(clientID)
	assert.Len(t, subs, 0)
}

// TestSubscriptionManager_UnsubscribeAll_Errors tests unsubscribe all error cases
func TestSubscriptionManager_UnsubscribeAll_Errors(t *testing.T) {
	tests := []struct {
		name     string
		clientID string
		setup    func(*SubscriptionManager)
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "empty client ID",
			clientID: "",
			setup:    func(sm *SubscriptionManager) {},
			wantErr:  true,
			errMsg:   "clientID cannot be empty",
		},
		{
			name:     "client not found",
			clientID: "client-nonexistent",
			setup:    func(sm *SubscriptionManager) {},
			wantErr:  true,
			errMsg:   "client not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSubscriptionManager(50)
			tt.setup(sm)

			err := sm.UnsubscribeAll(tt.clientID)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

// TestSubscriptionManager_GetSubscriptions tests retrieving subscriptions
func TestSubscriptionManager_GetSubscriptions(t *testing.T) {
	sm := NewSubscriptionManager(50)

	// Empty client
	subs := sm.GetSubscriptions("client-nonexistent")
	assert.Len(t, subs, 0)

	// Add subscriptions
	clientID := "client-1"
	err := sm.Subscribe(clientID, "bubblyui://state/refs", nil)
	require.NoError(t, err)
	err = sm.Subscribe(clientID, "bubblyui://components", nil)
	require.NoError(t, err)

	// Get subscriptions
	subs = sm.GetSubscriptions(clientID)
	assert.Len(t, subs, 2)

	// Verify it's a copy (modifying returned slice doesn't affect internal state)
	subs[0] = nil
	subs = sm.GetSubscriptions(clientID)
	assert.Len(t, subs, 2)
	assert.NotNil(t, subs[0])
}

// TestSubscriptionManager_ThreadSafety tests concurrent access
func TestSubscriptionManager_ThreadSafety(t *testing.T) {
	sm := NewSubscriptionManager(100)
	numGoroutines := 10
	numOpsPerGoroutine := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent subscribe operations
	for i := 0; i < numGoroutines; i++ {
		go func(clientNum int) {
			defer wg.Done()

			clientID := "client-1" // All goroutines use same client

			for j := 0; j < numOpsPerGoroutine; j++ {
				// Subscribe
				uri := "bubblyui://state/refs"
				filters := map[string]interface{}{"index": clientNum*numOpsPerGoroutine + j}
				err := sm.Subscribe(clientID, uri, filters)
				// May hit limit, that's ok
				_ = err

				// Get subscriptions
				subs := sm.GetSubscriptions(clientID)
				_ = subs

				// Get count
				count := sm.GetSubscriptionCount(clientID)
				_ = count
			}
		}(i)
	}

	wg.Wait()

	// Verify final state is consistent
	count := sm.GetSubscriptionCount("client-1")
	subs := sm.GetSubscriptions("client-1")
	assert.Equal(t, count, len(subs))
	assert.LessOrEqual(t, count, 100) // Should not exceed limit
}

// TestSubscriptionManager_ClientDisconnect tests cleanup on disconnect
func TestSubscriptionManager_ClientDisconnect(t *testing.T) {
	sm := NewSubscriptionManager(50)

	// Simulate multiple clients
	clients := []string{"client-1", "client-2", "client-3"}

	for _, clientID := range clients {
		for i := 0; i < 3; i++ {
			err := sm.Subscribe(clientID, "bubblyui://state/refs", map[string]interface{}{"index": i})
			require.NoError(t, err)
		}
	}

	// Verify all clients have subscriptions
	for _, clientID := range clients {
		assert.Equal(t, 3, sm.GetSubscriptionCount(clientID))
	}

	// Client 2 disconnects
	err := sm.UnsubscribeAll("client-2")
	require.NoError(t, err)

	// Verify client 2 has no subscriptions
	assert.Equal(t, 0, sm.GetSubscriptionCount("client-2"))

	// Verify other clients unaffected
	assert.Equal(t, 3, sm.GetSubscriptionCount("client-1"))
	assert.Equal(t, 3, sm.GetSubscriptionCount("client-3"))
}

// TestFiltersEqual tests filter comparison
func TestFiltersEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        map[string]interface{}
		b        map[string]interface{}
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "both empty",
			a:        map[string]interface{}{},
			b:        map[string]interface{}{},
			expected: true,
		},
		{
			name:     "nil and empty",
			a:        nil,
			b:        map[string]interface{}{},
			expected: true,
		},
		{
			name:     "equal filters",
			a:        map[string]interface{}{"ref_id": "count-ref"},
			b:        map[string]interface{}{"ref_id": "count-ref"},
			expected: true,
		},
		{
			name:     "different values",
			a:        map[string]interface{}{"ref_id": "count-ref"},
			b:        map[string]interface{}{"ref_id": "other-ref"},
			expected: false,
		},
		{
			name:     "different keys",
			a:        map[string]interface{}{"ref_id": "count-ref"},
			b:        map[string]interface{}{"component_id": "count-ref"},
			expected: false,
		},
		{
			name:     "different lengths",
			a:        map[string]interface{}{"ref_id": "count-ref"},
			b:        map[string]interface{}{"ref_id": "count-ref", "extra": "value"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filtersEqual(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
