package mcp

import (
	"sync"
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewStateChangeDetector tests detector creation
func TestNewStateChangeDetector(t *testing.T) {
	sm := NewSubscriptionManager(50)
	detector := NewStateChangeDetector(sm)

	assert.NotNil(t, detector)
	assert.NotNil(t, detector.subscriptions)
	assert.Equal(t, sm, detector.subscriptionMgr)
	assert.NotNil(t, detector.subscriptions) // Internal map for testing
}

// TestStateChangeDetector_Initialize tests hooking into DevTools
func TestStateChangeDetector_Initialize(t *testing.T) {
	// Create devtools instance
	dt := devtools.Enable()
	require.NotNil(t, dt)

	// Create detector
	sm := NewSubscriptionManager(50)
	detector := NewStateChangeDetector(sm)

	// Initialize should hook into devtools
	// Note: This may fail if devtools collector is not available
	// In that case, we just verify the error is handled gracefully
	err := detector.Initialize(dt)

	// Either success or graceful error
	if err != nil {
		assert.Contains(t, err.Error(), "collector not available")
	} else {
		// Verify hook was registered
		assert.NotNil(t, detector.devtools)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSubscriptionManager(50)
			detector := NewStateChangeDetector(sm)

			// Mock notifier to track notifications
			mockNotifier := &mockNotificationSender{}
			detector.notifier = mockNotifier

			// Add subscriptions directly to detector
			detector.mu.Lock()
			for _, sub := range tt.subscriptions {
				detector.subscriptions[sub.ClientID] = append(detector.subscriptions[sub.ClientID], sub)
			}
			detector.mu.Unlock()

			// Handle ref change
			detector.HandleRefChange(tt.refID, tt.oldValue, tt.newValue)

			// Verify notifications
			assert.Equal(t, tt.expectedNotifs, mockNotifier.notificationCount)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSubscriptionManager(50)
			detector := NewStateChangeDetector(sm)

			mockNotifier := &mockNotificationSender{}
			detector.notifier = mockNotifier

			detector.mu.Lock()
			for _, sub := range tt.subscriptions {
				detector.subscriptions[sub.ClientID] = append(detector.subscriptions[sub.ClientID], sub)
			}
			detector.mu.Unlock()

			detector.HandleComponentUnmount(tt.componentID, tt.componentName)

			assert.Equal(t, tt.expectedNotifs, mockNotifier.notificationCount)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSubscriptionManager(50)
			detector := NewStateChangeDetector(sm)

			mockNotifier := &mockNotificationSender{}
			detector.notifier = mockNotifier

			detector.mu.Lock()
			for _, sub := range tt.subscriptions {
				detector.subscriptions[sub.ClientID] = append(detector.subscriptions[sub.ClientID], sub)
			}
			detector.mu.Unlock()

			detector.HandleComponentMount(tt.componentID, tt.componentName)

			assert.Equal(t, tt.expectedNotifs, mockNotifier.notificationCount)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSubscriptionManager(50)
			detector := NewStateChangeDetector(sm)

			mockNotifier := &mockNotificationSender{}
			detector.notifier = mockNotifier

			detector.mu.Lock()
			for _, sub := range tt.subscriptions {
				detector.subscriptions[sub.ClientID] = append(detector.subscriptions[sub.ClientID], sub)
			}
			detector.mu.Unlock()

			detector.HandleEventEmit(tt.eventName, tt.componentID, tt.data)

			assert.Equal(t, tt.expectedNotifs, mockNotifier.notificationCount)
		})
	}
}

// TestStateChangeDetector_ThreadSafety tests concurrent access
func TestStateChangeDetector_ThreadSafety(t *testing.T) {
	sm := NewSubscriptionManager(100)
	detector := NewStateChangeDetector(sm)

	mockNotifier := &mockNotificationSender{}
	detector.notifier = mockNotifier

	// Add some subscriptions
	for i := 0; i < 10; i++ {
		sub := &Subscription{
			ID:          "sub-" + string(rune(i)),
			ClientID:    "client-1",
			ResourceURI: "bubblyui://state/refs",
			Filters:     nil,
			CreatedAt:   time.Now(),
		}
		detector.mu.Lock()
		detector.subscriptions[sub.ClientID] = append(detector.subscriptions[sub.ClientID], sub)
		detector.mu.Unlock()
	}

	numGoroutines := 10
	numOpsPerGoroutine := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent ref changes
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineNum int) {
			defer wg.Done()

			for j := 0; j < numOpsPerGoroutine; j++ {
				detector.HandleRefChange("ref-1", j, j+1)
			}
		}(i)
	}

	wg.Wait()

	// Should have processed all changes without panic
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

// TestMatchesFilter tests filter matching logic
func TestMatchesFilter(t *testing.T) {
	tests := []struct {
		name     string
		filters  map[string]interface{}
		data     map[string]interface{}
		expected bool
	}{
		{
			name:     "nil filter matches all",
			filters:  nil,
			data:     map[string]interface{}{"ref_id": "ref-1"},
			expected: true,
		},
		{
			name:     "empty filter matches all",
			filters:  map[string]interface{}{},
			data:     map[string]interface{}{"ref_id": "ref-1"},
			expected: true,
		},
		{
			name:     "exact match",
			filters:  map[string]interface{}{"ref_id": "ref-1"},
			data:     map[string]interface{}{"ref_id": "ref-1"},
			expected: true,
		},
		{
			name:     "no match",
			filters:  map[string]interface{}{"ref_id": "ref-1"},
			data:     map[string]interface{}{"ref_id": "ref-2"},
			expected: false,
		},
		{
			name:     "multiple filters all match",
			filters:  map[string]interface{}{"ref_id": "ref-1", "component_id": "comp-1"},
			data:     map[string]interface{}{"ref_id": "ref-1", "component_id": "comp-1"},
			expected: true,
		},
		{
			name:     "multiple filters partial match",
			filters:  map[string]interface{}{"ref_id": "ref-1", "component_id": "comp-1"},
			data:     map[string]interface{}{"ref_id": "ref-1", "component_id": "comp-2"},
			expected: false,
		},
		{
			name:     "filter key not in data",
			filters:  map[string]interface{}{"ref_id": "ref-1"},
			data:     map[string]interface{}{"component_id": "comp-1"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesFilter(tt.filters, tt.data)
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
