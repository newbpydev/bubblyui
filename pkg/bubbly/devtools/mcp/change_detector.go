package mcp

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

// StateChangeDetector hooks into DevTools to detect changes for subscriptions.
//
// It monitors ref changes, component lifecycle events, and event emissions,
// then notifies subscribed clients when changes match their subscription filters.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Lifecycle:
//
//  1. NewStateChangeDetector() - Creates the detector
//  2. Initialize() - Hooks into DevTools
//  3. ... DevTools fires hooks as changes occur ...
//  4. HandleRefChange/HandleComponentMount/HandleEventEmit - Process changes
//  5. ... Notifications sent to subscribed clients ...
//
// Example:
//
//	sm := NewSubscriptionManager(50)
//	detector := NewStateChangeDetector(sm)
//	err := detector.Initialize(devtools.Enable())
//	if err != nil {
//	    log.Printf("Failed to initialize detector: %v", err)
//	}
type StateChangeDetector struct {
	// subscriptionMgr is the subscription manager that tracks active subscriptions
	subscriptionMgr *SubscriptionManager

	// subscriptions is a direct map for testing purposes
	// In production, this is populated from subscriptionMgr
	// Key: client ID, Value: subscriptions for that client
	subscriptions map[string][]*Subscription

	// notifier sends notifications to clients (will be implemented in Task 4.4)
	// For now, this is a simple interface for testing
	notifier notificationSender

	// devtools is the DevTools instance we're hooked into
	devtools *devtools.DevTools

	// mu protects concurrent access to subscriptions map
	// Note: SubscriptionManager has its own mutex, but we need this for
	// iterating over subscriptions during change detection
	mu sync.RWMutex
}

// notificationSender is an interface for sending notifications to clients.
// This will be implemented properly in Task 4.4 (Notification Sender).
// For now, we use this interface to enable testing.
type notificationSender interface {
	QueueNotification(clientID, uri string, data map[string]interface{})
}

// NewStateChangeDetector creates a new state change detector.
//
// The detector is created but not yet hooked into DevTools.
// Call Initialize() to register the hooks.
//
// Thread Safety:
//
//	Safe to call concurrently (creates new instance each time).
//
// Example:
//
//	sm := NewSubscriptionManager(50)
//	detector := NewStateChangeDetector(sm)
//
// Parameters:
//   - subscriptionMgr: The subscription manager to use for tracking subscriptions
//
// Returns:
//   - *StateChangeDetector: A new change detector instance
func NewStateChangeDetector(subscriptionMgr *SubscriptionManager) *StateChangeDetector {
	return &StateChangeDetector{
		subscriptionMgr: subscriptionMgr,
		subscriptions:   make(map[string][]*Subscription),
	}
}

// SetNotifier sets the notification sender for this detector.
//
// This method configures the detector to use the provided notifier
// for sending notifications to clients when changes are detected.
//
// Thread Safety:
//
//	Safe to call concurrently, but should be called before any changes occur.
//
// Example:
//
//	notifier, err := NewNotificationSender(batcher)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	detector.SetNotifier(notifier)
//
// Parameters:
//   - notifier: The notification sender to use
func (d *StateChangeDetector) SetNotifier(notifier notificationSender) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.notifier = notifier
}

// Initialize hooks the detector into the DevTools system.
//
// This method registers a custom hook with DevTools that will be called
// whenever state changes, components mount/unmount, or events are emitted.
//
// Thread Safety:
//
//	Safe to call concurrently, but should only be called once per detector.
//
// Example:
//
//	dt := devtools.Enable()
//	err := detector.Initialize(dt)
//	if err != nil {
//	    log.Printf("Failed to initialize: %v", err)
//	}
//
// Parameters:
//   - dt: The DevTools instance to hook into
//
// Returns:
//   - error: Initialization error, or nil on success
func (d *StateChangeDetector) Initialize(dt *devtools.DevTools) error {
	if dt == nil {
		return fmt.Errorf("devtools cannot be nil")
	}

	d.devtools = dt

	// Register a custom hook to detect changes
	// The hook will call our Handle* methods when changes occur
	hook := &stateDetectorHook{detector: d}

	// Get the collector from the devtools package
	collector := devtools.GetCollector()
	if collector == nil {
		return fmt.Errorf("devtools collector not available")
	}

	collector.AddStateHook(hook)

	return nil
}

// HandleRefChange processes a ref value change and notifies subscribed clients.
//
// This method:
//   - Finds all subscriptions interested in ref changes
//   - Checks if the change matches each subscription's filters
//   - Queues notifications for matching subscriptions
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	detector.HandleRefChange("ref-1", 41, 42)
//
// Parameters:
//   - refID: The unique identifier of the ref that changed
//   - oldValue: The value before the change
//   - newValue: The value after the change
func (d *StateChangeDetector) HandleRefChange(refID string, oldValue, newValue interface{}) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Get all subscriptions from the manager
	// We need to check all clients' subscriptions
	allSubs := d.getAllSubscriptions()

	// Find subscriptions interested in ref changes
	for _, sub := range allSubs {
		// Check if subscription is for state/refs
		if !strings.HasPrefix(sub.ResourceURI, "bubblyui://state/refs") {
			continue
		}

		// Check if filters match
		data := map[string]interface{}{
			"ref_id": refID,
		}

		if !matchesFilter(sub.Filters, data) {
			continue
		}

		// Queue notification
		if d.notifier != nil {
			notificationData := map[string]interface{}{
				"ref_id":    refID,
				"old_value": oldValue,
				"new_value": newValue,
			}
			d.notifier.QueueNotification(sub.ClientID, sub.ResourceURI, notificationData)
		}
	}
}

// HandleComponentMount processes a component mount event and notifies subscribed clients.
//
// This method:
//   - Finds all subscriptions interested in component changes
//   - Checks if the mount event matches each subscription's filters
//   - Queues notifications for matching subscriptions
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	detector.HandleComponentMount("comp-1", "Counter")
//
// Parameters:
//   - componentID: The unique identifier of the mounted component
//   - componentName: The name of the mounted component
func (d *StateChangeDetector) HandleComponentMount(componentID, componentName string) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	allSubs := d.getAllSubscriptions()

	for _, sub := range allSubs {
		// Check if subscription is for components
		if !strings.HasPrefix(sub.ResourceURI, "bubblyui://components") {
			continue
		}

		// Check if filters match
		data := map[string]interface{}{
			"component_id":   componentID,
			"component_name": componentName,
		}

		if !matchesFilter(sub.Filters, data) {
			continue
		}

		// Queue notification
		if d.notifier != nil {
			notificationData := map[string]interface{}{
				"component_id":   componentID,
				"component_name": componentName,
				"event_type":     "mount",
			}
			d.notifier.QueueNotification(sub.ClientID, sub.ResourceURI, notificationData)
		}
	}
}

// HandleComponentUnmount processes a component unmount event and notifies subscribed clients.
//
// This method:
//   - Finds all subscriptions interested in component changes
//   - Checks if the unmount event matches each subscription's filters
//   - Queues notifications for matching subscriptions
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	detector.HandleComponentUnmount("comp-1", "Counter")
//
// Parameters:
//   - componentID: The unique identifier of the unmounted component
//   - componentName: The name of the unmounted component
func (d *StateChangeDetector) HandleComponentUnmount(componentID, componentName string) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	allSubs := d.getAllSubscriptions()

	for _, sub := range allSubs {
		// Check if subscription is for components
		if !strings.HasPrefix(sub.ResourceURI, "bubblyui://components") {
			continue
		}

		// Check if filters match
		data := map[string]interface{}{
			"component_id":   componentID,
			"component_name": componentName,
		}

		if !matchesFilter(sub.Filters, data) {
			continue
		}

		// Queue notification
		if d.notifier != nil {
			notificationData := map[string]interface{}{
				"component_id":   componentID,
				"component_name": componentName,
				"event_type":     "unmount",
			}
			d.notifier.QueueNotification(sub.ClientID, sub.ResourceURI, notificationData)
		}
	}
}

// HandleEventEmit processes an event emission and notifies subscribed clients.
//
// This method:
//   - Finds all subscriptions interested in event emissions
//   - Checks if the event matches each subscription's filters
//   - Queues notifications for matching subscriptions
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	detector.HandleEventEmit("increment", "comp-1", nil)
//
// Parameters:
//   - eventName: The name of the emitted event
//   - componentID: The ID of the component that emitted the event
//   - data: Optional event data
func (d *StateChangeDetector) HandleEventEmit(eventName, componentID string, data interface{}) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	allSubs := d.getAllSubscriptions()

	for _, sub := range allSubs {
		// Check if subscription is for events
		if !strings.HasPrefix(sub.ResourceURI, "bubblyui://events") {
			continue
		}

		// Check if filters match
		filterData := map[string]interface{}{
			"event_name":   eventName,
			"component_id": componentID,
		}

		if !matchesFilter(sub.Filters, filterData) {
			continue
		}

		// Queue notification
		if d.notifier != nil {
			notificationData := map[string]interface{}{
				"event_name":   eventName,
				"component_id": componentID,
				"data":         data,
			}
			d.notifier.QueueNotification(sub.ClientID, sub.ResourceURI, notificationData)
		}
	}
}

// getAllSubscriptions returns all subscriptions from all clients.
//
// This is a helper method that collects subscriptions from the subscription manager.
// The returned slice is safe to iterate over.
//
// Thread Safety:
//
//	Safe to call concurrently (SubscriptionManager handles its own locking).
//
// Returns:
//   - []*Subscription: All active subscriptions across all clients
func (d *StateChangeDetector) getAllSubscriptions() []*Subscription {
	// First check if we have subscriptions in the detector (for testing)
	d.mu.RLock()
	if len(d.subscriptions) > 0 {
		var allSubs []*Subscription
		for _, clientSubs := range d.subscriptions {
			allSubs = append(allSubs, clientSubs...)
		}
		d.mu.RUnlock()
		return allSubs
	}
	d.mu.RUnlock()

	// Otherwise, get from subscription manager
	if d.subscriptionMgr == nil {
		return []*Subscription{}
	}

	d.subscriptionMgr.mu.RLock()
	defer d.subscriptionMgr.mu.RUnlock()

	var allSubs []*Subscription
	for _, clientSubs := range d.subscriptionMgr.subscriptions {
		allSubs = append(allSubs, clientSubs...)
	}

	return allSubs
}

// matchesFilter checks if data matches the subscription filters.
//
// Filter matching rules:
//   - Nil or empty filters match everything
//   - All filter keys must exist in data and have matching values
//   - If any filter key doesn't match, the entire filter fails
//
// Thread Safety:
//
//	Safe to call concurrently (pure function, no shared state).
//
// Example:
//
//	filters := map[string]interface{}{"ref_id": "ref-1"}
//	data := map[string]interface{}{"ref_id": "ref-1", "value": 42}
//	matches := matchesFilter(filters, data) // true
//
// Parameters:
//   - filters: The subscription filters to check against
//   - data: The event data to match
//
// Returns:
//   - bool: True if data matches filters, false otherwise
func matchesFilter(filters, data map[string]interface{}) bool {
	// Nil or empty filters match everything
	if len(filters) == 0 {
		return true
	}

	// All filter keys must match
	for key, filterValue := range filters {
		dataValue, exists := data[key]
		if !exists {
			return false
		}

		// Simple equality check
		// For complex types, would need deep comparison
		if filterValue != dataValue {
			return false
		}
	}

	return true
}

// stateDetectorHook implements devtools.StateHook to receive change notifications.
//
// This is an internal type that bridges DevTools hooks to our change detector.
type stateDetectorHook struct {
	detector *StateChangeDetector
}

// OnRefChanged implements devtools.StateHook.
func (h *stateDetectorHook) OnRefChanged(refID string, oldValue, newValue interface{}) {
	h.detector.HandleRefChange(refID, oldValue, newValue)
}

// OnComputedEvaluated implements devtools.StateHook.
func (h *stateDetectorHook) OnComputedEvaluated(computedID string, value interface{}, duration time.Duration) {
	// Treat computed values like refs for subscription purposes
	// We don't have the old value, so we pass nil
	h.detector.HandleRefChange(computedID, nil, value)
}

// OnWatcherTriggered implements devtools.StateHook.
func (h *stateDetectorHook) OnWatcherTriggered(watcherID string, value interface{}) {
	// Watchers are triggered when watched values change
	// Treat this like a ref change for subscription purposes
	h.detector.HandleRefChange(watcherID, nil, value)
}
