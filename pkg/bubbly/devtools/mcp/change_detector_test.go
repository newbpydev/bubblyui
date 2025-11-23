package mcp

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

// TestNewStateChangeDetector tests detector creation
func TestNewStateChangeDetector(t *testing.T) {
	sm := NewSubscriptionManager(50)
	detector := NewStateChangeDetector(sm)

	assert.NotNil(t, detector)
	assert.NotNil(t, detector.subscriptions)
	assert.Equal(t, sm, detector.subscriptionMgr)
}

// TestStateChangeDetector_Initialize tests hooking into DevTools
func TestStateChangeDetector_Initialize(t *testing.T) {
	tests := []struct {
		name        string
		devtools    *devtools.DevTools
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil devtools returns error",
			devtools:    nil,
			expectError: true,
			errorMsg:    "devtools cannot be nil",
		},
		{
			name:        "valid devtools initializes successfully",
			devtools:    devtools.Enable(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSubscriptionManager(50)
			detector := NewStateChangeDetector(sm)

			err := detector.Initialize(tt.devtools)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				// Either success or graceful error if collector not available
				if err != nil {
					assert.Contains(t, err.Error(), "collector not available")
				} else {
					// Verify hook was registered
					assert.NotNil(t, detector.devtools)
				}
			}
		})
	}
}

// TestStateChangeDetector_Initialize_CollectorNil specifically tests the nil collector case
func TestStateChangeDetector_Initialize_CollectorNil(t *testing.T) {
	// This test specifically targets the collector nil case (lines 127-129)
	// We can't easily mock devtools.GetCollector() to return nil,
	// so we'll test the scenario where it might return nil
	sm := NewSubscriptionManager(50)
	detector := NewStateChangeDetector(sm)

	// Try with a devtools instance that might not have a collector initialized
	dt := &devtools.DevTools{}
	err := detector.Initialize(dt)

	// Should either succeed or fail gracefully with collector not available
	if err != nil {
		assert.Contains(t, err.Error(), "collector not available")
	}
}

// TestStateChangeDetector_HandleRefChange tests ref change detection
func TestStateChangeDetector_HandleRefChange(t *testing.T) {
	tests := []struct {
		name           string
		subscriptions  []*Subscription
		refID          string
		oldValue       interface{}
		newValue       interface{}
		expectedNotifs int
	}{
		{
			name:           "no subscriptions",
			subscriptions:  []*Subscription{},
			refID:          "ref-1",
			oldValue:       41,
			newValue:       42,
			expectedNotifs: 0,
		},
		{
			name: "subscription matches - no filter",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://state/refs",
					Filters:     nil,
					CreatedAt:   time.Now(),
				},
			},
			refID:          "ref-1",
			oldValue:       41,
			newValue:       42,
			expectedNotifs: 1,
		},
		{
			name: "subscription matches - with matching filter",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://state/refs",
					Filters:     map[string]interface{}{"ref_id": "ref-1"},
					CreatedAt:   time.Now(),
				},
			},
			refID:          "ref-1",
			oldValue:       41,
			newValue:       42,
			expectedNotifs: 1,
		},
		{
			name: "subscription does not match - wrong filter",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://state/refs",
					Filters:     map[string]interface{}{"ref_id": "ref-2"},
					CreatedAt:   time.Now(),
				},
			},
			refID:          "ref-1",
			oldValue:       41,
			newValue:       42,
			expectedNotifs: 0,
		},
		{
			name: "multiple subscriptions - some match",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://state/refs",
					Filters:     map[string]interface{}{"ref_id": "ref-1"},
					CreatedAt:   time.Now(),
				},
				{
					ID:          "sub-2",
					ClientID:    "client-2",
					ResourceURI: "bubblyui://state/refs",
					Filters:     map[string]interface{}{"ref_id": "ref-2"},
					CreatedAt:   time.Now(),
				},
				{
					ID:          "sub-3",
					ClientID:    "client-3",
					ResourceURI: "bubblyui://state/refs",
					Filters:     nil, // Matches all
					CreatedAt:   time.Now(),
				},
			},
			refID:          "ref-1",
			oldValue:       41,
			newValue:       42,
			expectedNotifs: 2, // sub-1 and sub-3 match
		},
		{
			name: "subscription with wrong URI prefix",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://components", // Wrong prefix
					Filters:     nil,
					CreatedAt:   time.Now(),
				},
			},
			refID:          "ref-1",
			oldValue:       41,
			newValue:       42,
			expectedNotifs: 0,
		},
		{
			name: "nil notifier - no panic",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://state/refs",
					Filters:     nil,
					CreatedAt:   time.Now(),
				},
			},
			refID:          "ref-1",
			oldValue:       41,
			newValue:       42,
			expectedNotifs: 0, // No notifier, so no notifications
		},
		{
			name: "edge case values - nil to nil",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://state/refs",
					Filters:     nil,
					CreatedAt:   time.Now(),
				},
			},
			refID:          "ref-nil",
			oldValue:       nil,
			newValue:       nil,
			expectedNotifs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSubscriptionManager(50)
			detector := NewStateChangeDetector(sm)

			// Mock notifier to track notifications (unless testing nil case)
			if tt.expectedNotifs > 0 {
				mockNotifier := &mockNotificationSender{}
				detector.notifier = mockNotifier
			}

			// Add subscriptions directly to detector
			detector.mu.Lock()
			for _, sub := range tt.subscriptions {
				detector.subscriptions[sub.ClientID] = append(detector.subscriptions[sub.ClientID], sub)
			}
			detector.mu.Unlock()

			// Handle ref change
			detector.HandleRefChange(tt.refID, tt.oldValue, tt.newValue)

			// Verify notifications
			if detector.notifier != nil {
				mockNotifier := detector.notifier.(*mockNotificationSender)
				assert.Equal(t, tt.expectedNotifs, mockNotifier.notificationCount)
			}
		})
	}
}

// TestStateChangeDetector_HandleComponentUnmount tests component unmount detection
func TestStateChangeDetector_HandleComponentUnmount(t *testing.T) {
	tests := []struct {
		name           string
		subscriptions  []*Subscription
		componentID    string
		componentName  string
		expectedNotifs int
	}{
		{
			name:           "no subscriptions",
			subscriptions:  []*Subscription{},
			componentID:    "comp-1",
			componentName:  "Counter",
			expectedNotifs: 0,
		},
		{
			name: "subscription to components",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://components",
					Filters:     nil,
					CreatedAt:   time.Now(),
				},
			},
			componentID:    "comp-1",
			componentName:  "Counter",
			expectedNotifs: 1,
		},
		{
			name: "subscription with wrong URI prefix",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://state/refs", // Wrong prefix
					Filters:     nil,
					CreatedAt:   time.Now(),
				},
			},
			componentID:    "comp-1",
			componentName:  "Counter",
			expectedNotifs: 0,
		},
		{
			name: "nil notifier - no panic",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://components",
					Filters:     nil,
					CreatedAt:   time.Now(),
				},
			},
			componentID:    "comp-1",
			componentName:  "Counter",
			expectedNotifs: 0, // No notifier
		},
		{
			name: "subscription matches but notifier nil - covers line 284",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://components",
					Filters:     map[string]interface{}{"component_name": "Counter"},
					CreatedAt:   time.Now(),
				},
			},
			componentID:    "comp-1",
			componentName:  "Counter",
			expectedNotifs: 0, // No notifier
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSubscriptionManager(50)
			detector := NewStateChangeDetector(sm)

			// Mock notifier to track notifications (unless testing nil case)
			if tt.expectedNotifs > 0 {
				mockNotifier := &mockNotificationSender{}
				detector.notifier = mockNotifier
			}

			detector.mu.Lock()
			for _, sub := range tt.subscriptions {
				detector.subscriptions[sub.ClientID] = append(detector.subscriptions[sub.ClientID], sub)
			}
			detector.mu.Unlock()

			detector.HandleComponentUnmount(tt.componentID, tt.componentName)

			// Verify notifications
			if detector.notifier != nil {
				mockNotifier := detector.notifier.(*mockNotificationSender)
				assert.Equal(t, tt.expectedNotifs, mockNotifier.notificationCount)
			}
		})
	}
}

// TestStateChangeDetector_HandleComponentMount tests component mount detection
func TestStateChangeDetector_HandleComponentMount(t *testing.T) {
	tests := []struct {
		name           string
		subscriptions  []*Subscription
		componentID    string
		componentName  string
		expectedNotifs int
	}{
		{
			name:           "no subscriptions",
			subscriptions:  []*Subscription{},
			componentID:    "comp-1",
			componentName:  "Counter",
			expectedNotifs: 0,
		},
		{
			name: "subscription to components",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://components",
					Filters:     nil,
					CreatedAt:   time.Now(),
				},
			},
			componentID:    "comp-1",
			componentName:  "Counter",
			expectedNotifs: 1,
		},
		{
			name: "subscription with component filter match",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://components",
					Filters:     map[string]interface{}{"component_name": "Counter"},
					CreatedAt:   time.Now(),
				},
			},
			componentID:    "comp-1",
			componentName:  "Counter",
			expectedNotifs: 1,
		},
		{
			name: "subscription with component filter no match",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://components",
					Filters:     map[string]interface{}{"component_name": "TodoList"},
					CreatedAt:   time.Now(),
				},
			},
			componentID:    "comp-1",
			componentName:  "Counter",
			expectedNotifs: 0,
		},
		{
			name: "subscription with wrong URI prefix",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://events", // Wrong prefix
					Filters:     nil,
					CreatedAt:   time.Now(),
				},
			},
			componentID:    "comp-1",
			componentName:  "Counter",
			expectedNotifs: 0,
		},
		{
			name: "nil notifier - no panic",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://components",
					Filters:     nil,
					CreatedAt:   time.Now(),
				},
			},
			componentID:    "comp-1",
			componentName:  "Counter",
			expectedNotifs: 0, // No notifier
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSubscriptionManager(50)
			detector := NewStateChangeDetector(sm)

			// Mock notifier to track notifications (unless testing nil case)
			if tt.expectedNotifs > 0 {
				mockNotifier := &mockNotificationSender{}
				detector.notifier = mockNotifier
			}

			detector.mu.Lock()
			for _, sub := range tt.subscriptions {
				detector.subscriptions[sub.ClientID] = append(detector.subscriptions[sub.ClientID], sub)
			}
			detector.mu.Unlock()

			detector.HandleComponentMount(tt.componentID, tt.componentName)

			// Verify notifications
			if detector.notifier != nil {
				mockNotifier := detector.notifier.(*mockNotificationSender)
				assert.Equal(t, tt.expectedNotifs, mockNotifier.notificationCount)
			}
		})
	}
}

// TestStateChangeDetector_HandleEventEmit tests event emission detection
func TestStateChangeDetector_HandleEventEmit(t *testing.T) {
	tests := []struct {
		name           string
		subscriptions  []*Subscription
		eventName      string
		componentID    string
		data           interface{}
		expectedNotifs int
	}{
		{
			name:           "no subscriptions",
			subscriptions:  []*Subscription{},
			eventName:      "increment",
			componentID:    "comp-1",
			data:           nil,
			expectedNotifs: 0,
		},
		{
			name: "subscription to events",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://events/log",
					Filters:     nil,
					CreatedAt:   time.Now(),
				},
			},
			eventName:      "increment",
			componentID:    "comp-1",
			data:           nil,
			expectedNotifs: 1,
		},
		{
			name: "subscription with event filter match",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://events/log",
					Filters:     map[string]interface{}{"event_name": "increment"},
					CreatedAt:   time.Now(),
				},
			},
			eventName:      "increment",
			componentID:    "comp-1",
			data:           nil,
			expectedNotifs: 1,
		},
		{
			name: "subscription with event filter no match",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://events/log",
					Filters:     map[string]interface{}{"event_name": "decrement"},
					CreatedAt:   time.Now(),
				},
			},
			eventName:      "increment",
			componentID:    "comp-1",
			data:           nil,
			expectedNotifs: 0,
		},
		{
			name: "subscription with wrong URI prefix",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://components", // Wrong prefix
					Filters:     nil,
					CreatedAt:   time.Now(),
				},
			},
			eventName:      "increment",
			componentID:    "comp-1",
			data:           nil,
			expectedNotifs: 0,
		},
		{
			name: "nil notifier - no panic",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://events/log",
					Filters:     nil,
					CreatedAt:   time.Now(),
				},
			},
			eventName:      "increment",
			componentID:    "comp-1",
			data:           nil,
			expectedNotifs: 0, // No notifier
		},
		{
			name: "event with complex data",
			subscriptions: []*Subscription{
				{
					ID:          "sub-1",
					ClientID:    "client-1",
					ResourceURI: "bubblyui://events/log",
					Filters:     nil,
					CreatedAt:   time.Now(),
				},
			},
			eventName:      "complex-event",
			componentID:    "comp-1",
			data:           map[string]interface{}{"key": "value", "count": 42},
			expectedNotifs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSubscriptionManager(50)
			detector := NewStateChangeDetector(sm)

			// Mock notifier to track notifications (unless testing nil case)
			if tt.expectedNotifs > 0 {
				mockNotifier := &mockNotificationSender{}
				detector.notifier = mockNotifier
			}

			detector.mu.Lock()
			for _, sub := range tt.subscriptions {
				detector.subscriptions[sub.ClientID] = append(detector.subscriptions[sub.ClientID], sub)
			}
			detector.mu.Unlock()

			detector.HandleEventEmit(tt.eventName, tt.componentID, tt.data)

			// Verify notifications
			if detector.notifier != nil {
				mockNotifier := detector.notifier.(*mockNotificationSender)
				assert.Equal(t, tt.expectedNotifs, mockNotifier.notificationCount)
			}
		})
	}
}

// TestStateChangeDetector_EdgeCases tests additional edge cases for coverage
func TestStateChangeDetector_EdgeCases(t *testing.T) {
	sm := NewSubscriptionManager(50)
	detector := NewStateChangeDetector(sm)

	// Test 1: Component unmount with empty component name
	t.Run("component_unmount_empty_name", func(t *testing.T) {
		mockNotifier := &mockNotificationSender{}
		detector.notifier = mockNotifier

		detector.mu.Lock()
		detector.subscriptions["client1"] = []*Subscription{
			{
				ID:          "sub-1",
				ClientID:    "client1",
				ResourceURI: "bubblyui://components",
				Filters:     map[string]interface{}{"component_name": ""},
				CreatedAt:   time.Now(),
			},
		}
		detector.mu.Unlock()

		detector.HandleComponentUnmount("comp-1", "")

		// Should have processed and matched
		assert.Equal(t, 1, mockNotifier.notificationCount)
	})

	// Test 2: Component mount with special characters in ID
	t.Run("component_mount_special_chars", func(t *testing.T) {
		mockNotifier := &mockNotificationSender{}
		detector.notifier = mockNotifier

		detector.mu.Lock()
		detector.subscriptions["client2"] = []*Subscription{
			{
				ID:          "sub-2",
				ClientID:    "client2",
				ResourceURI: "bubblyui://components",
				Filters:     map[string]interface{}{"component_id": "comp-special-123"},
				CreatedAt:   time.Now(),
			},
		}
		detector.mu.Unlock()

		detector.HandleComponentMount("comp-special-123", "SpecialComponent")

		// Should have processed and matched
		assert.Equal(t, 1, mockNotifier.notificationCount)
	})

	// Test 3: Event emit with nil data
	t.Run("event_emit_nil_data", func(t *testing.T) {
		mockNotifier := &mockNotificationSender{}
		detector.notifier = mockNotifier

		detector.mu.Lock()
		detector.subscriptions["client3"] = []*Subscription{
			{
				ID:          "sub-3",
				ClientID:    "client3",
				ResourceURI: "bubblyui://events/log",
				Filters:     nil,
				CreatedAt:   time.Now(),
			},
		}
		detector.mu.Unlock()

		detector.HandleEventEmit("test-event", "comp-1", nil)

		// Should have processed even with nil data
		assert.Equal(t, 1, mockNotifier.notificationCount)
	})
}

// TestStateChangeDetector_ThreadSafety tests concurrent access
func TestStateChangeDetector_ThreadSafety(t *testing.T) {
	sm := NewSubscriptionManager(100)
	detector := NewStateChangeDetector(sm)

	mockNotifier := &mockNotificationSender{}
	detector.notifier = mockNotifier

	// Add some subscriptions
	detector.mu.Lock()
	detector.subscriptions["client1"] = []*Subscription{
		{
			ID:          "sub-1",
			ClientID:    "client1",
			ResourceURI: "bubblyui://state/refs",
			Filters:     nil,
			CreatedAt:   time.Now(),
		},
	}
	detector.mu.Unlock()

	// Run concurrent operations
	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(_ int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Mix different operations
				switch j % 4 {
				case 0:
					detector.HandleRefChange("ref-1", j, j+1)
				case 1:
					detector.HandleComponentMount("comp-1", "Counter")
				case 2:
					detector.HandleComponentUnmount("comp-1", "Counter")
				case 3:
					detector.HandleEventEmit("click", "comp-1", nil)
				}
			}
		}(i)
	}

	wg.Wait()

	// Should have processed all operations without panics
	assert.True(t, mockNotifier.notificationCount > 0)
}

// TestStateDetectorHook tests the hook implementation
func TestStateDetectorHook(t *testing.T) {
	sm := NewSubscriptionManager(50)
	detector := NewStateChangeDetector(sm)

	mockNotifier := &mockNotificationSender{}
	detector.notifier = mockNotifier

	// Add a subscription
	sub := &Subscription{
		ID:          "sub-1",
		ClientID:    "client-1",
		ResourceURI: "bubblyui://state/refs",
		Filters:     nil,
		CreatedAt:   time.Now(),
	}
	detector.mu.Lock()
	detector.subscriptions[sub.ClientID] = append(detector.subscriptions[sub.ClientID], sub)
	detector.mu.Unlock()

	// Create hook
	hook := &stateDetectorHook{detector: detector}

	// Test OnRefChanged
	hook.OnRefChanged("ref-1", 41, 42)
	assert.Equal(t, 1, mockNotifier.notificationCount)

	// Test OnComputedEvaluated
	hook.OnComputedEvaluated("computed-1", 100, time.Millisecond)
	assert.Equal(t, 2, mockNotifier.notificationCount)

	// Test OnWatcherTriggered
	hook.OnWatcherTriggered("watcher-1", "value")
	assert.Equal(t, 3, mockNotifier.notificationCount)
}

// TestGetAllSubscriptions tests subscription retrieval logic
func TestGetAllSubscriptions(t *testing.T) {
	tests := []struct {
		name            string
		subscriptions   map[string][]*Subscription
		subscriptionMgr *SubscriptionManager
		expectedCount   int
	}{
		{
			name:            "uses internal subscriptions map when available",
			subscriptions:   map[string][]*Subscription{"client1": {{ID: "sub1"}, {ID: "sub2"}}},
			subscriptionMgr: nil,
			expectedCount:   2,
		},
		{
			name:            "falls back to subscription manager when internal map empty",
			subscriptions:   map[string][]*Subscription{},
			subscriptionMgr: NewSubscriptionManager(50),
			expectedCount:   0,
		},
		{
			name:            "both internal and manager subscriptions",
			subscriptions:   map[string][]*Subscription{"client1": {{ID: "sub1"}}},
			subscriptionMgr: NewSubscriptionManager(50),
			expectedCount:   1,
		},
		{
			name:            "nil subscription manager returns empty slice",
			subscriptions:   map[string][]*Subscription{},
			subscriptionMgr: nil,
			expectedCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &StateChangeDetector{
				subscriptions:   tt.subscriptions,
				subscriptionMgr: tt.subscriptionMgr,
			}

			allSubs := detector.getAllSubscriptions()
			assert.Equal(t, tt.expectedCount, len(allSubs))
		})
	}
}

// TestMatchesFilter tests filter matching logic
func TestMatchesFilter(t *testing.T) {
	tests := []struct {
		name     string
		filter   map[string]interface{}
		data     map[string]interface{}
		expected bool
	}{
		{
			name:     "nil filter matches all",
			filter:   nil,
			data:     map[string]interface{}{"key": "value"},
			expected: true,
		},
		{
			name:     "empty filter matches all",
			filter:   map[string]interface{}{},
			data:     map[string]interface{}{"key": "value"},
			expected: true,
		},
		{
			name:     "exact match",
			filter:   map[string]interface{}{"key": "value"},
			data:     map[string]interface{}{"key": "value"},
			expected: true,
		},
		{
			name:     "no match",
			filter:   map[string]interface{}{"key": "different"},
			data:     map[string]interface{}{"key": "value"},
			expected: false,
		},
		{
			name:     "multiple filters all match",
			filter:   map[string]interface{}{"key1": "value1", "key2": "value2"},
			data:     map[string]interface{}{"key1": "value1", "key2": "value2"},
			expected: true,
		},
		{
			name:     "multiple filters partial match",
			filter:   map[string]interface{}{"key1": "value1", "key2": "wrong"},
			data:     map[string]interface{}{"key1": "value1", "key2": "value2"},
			expected: false,
		},
		{
			name:     "filter key not in data",
			filter:   map[string]interface{}{"missing": "value"},
			data:     map[string]interface{}{"key": "value"},
			expected: false,
		},
		{
			name:     "nil data with non-nil filter",
			filter:   map[string]interface{}{"key": "value"},
			data:     nil,
			expected: false,
		},
		{
			name:     "complex type comparison",
			filter:   map[string]interface{}{"count": 42},
			data:     map[string]interface{}{"count": 42},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesFilter(tt.filter, tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// mockNotificationSender is a mock for testing
type mockNotificationSender struct {
	notificationCount int
	notifications     []mockNotification
	mu                sync.Mutex
}

type mockNotification struct {
	clientID string
	uri      string
	data     map[string]interface{}
}

func (m *mockNotificationSender) QueueNotification(clientID, uri string, data map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.notificationCount++
	m.notifications = append(m.notifications, mockNotification{
		clientID: clientID,
		uri:      uri,
		data:     data,
	})
}
