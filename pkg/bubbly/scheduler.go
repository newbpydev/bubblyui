package bubbly

import "sync"

// callbackFunc is a type-erased callback function that can be queued and executed later.
// It wraps the actual watcher callback with its arguments.
type callbackFunc func()

// CallbackScheduler manages queued watcher callbacks for post-flush mode.
// It provides batching behavior where multiple Set() calls on the same Ref
// will only trigger the watcher callback once with the final value.
//
// Thread-safe for concurrent access.
type CallbackScheduler struct {
	mu    sync.Mutex
	queue map[interface{}]callbackFunc // map[*watcher[T]]func()
}

// globalScheduler is the singleton scheduler instance used by all watchers.
var globalScheduler = &CallbackScheduler{
	queue: make(map[interface{}]callbackFunc),
}

// enqueue adds a callback to the queue. If a callback for the same watcher
// already exists, it replaces it (batching behavior).
//
// The callback function should capture the watcher and values to execute later.
// This is called when a watcher has flush mode "post".
func (s *CallbackScheduler) enqueue(watcherKey interface{}, callback callbackFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store callback, replacing any existing one for this watcher (batching)
	s.queue[watcherKey] = callback
}

// flush executes all queued callbacks and clears the queue.
// Returns the number of callbacks executed.
//
// This is called by FlushWatchers() to execute all pending callbacks.
func (s *CallbackScheduler) flush() int {
	s.mu.Lock()
	// Copy queue and clear it while holding lock
	callbacks := s.queue
	s.queue = make(map[interface{}]callbackFunc)
	s.mu.Unlock()

	// Execute callbacks outside the lock
	for _, cb := range callbacks {
		cb()
	}

	return len(callbacks)
}

// FlushWatchers executes all queued watcher callbacks.
//
// This function should be called at strategic points in your application,
// typically at the end of an Update() cycle in Bubbletea, to execute all
// callbacks that were queued during state changes.
//
// Returns the number of callbacks that were executed.
//
// Example with Bubbletea:
//
//	func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	    switch msg := msg.(type) {
//	    case someMsg:
//	        m.count.Set(m.count.GetTyped() + 1)  // Queued if using WithFlush("post")
//	    }
//
//	    // Execute all queued callbacks before returning
//	    FlushWatchers()
//
//	    return m, nil
//	}
//
// Thread-safe and can be called from any goroutine.
func FlushWatchers() int {
	return globalScheduler.flush()
}

// PendingCallbacks returns the number of callbacks currently queued.
// Useful for testing and debugging.
func PendingCallbacks() int {
	globalScheduler.mu.Lock()
	defer globalScheduler.mu.Unlock()
	return len(globalScheduler.queue)
}
