package core

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// State represents mutable state within a component.
// Unlike Props, State is designed to be updated by the component that owns it.
type State[T any] struct {
	signal        *Signal[T]
	mutex         sync.RWMutex
	updateQueue   []StateUpdate[T]
	batchDepth    int32
	changeHistory []T
	historySize   int
	onChange      []func(old, new T)
}

// StateUpdate represents a state update function
type StateUpdate[T any] func(current T) T

// NewState creates a new State instance with the provided initial value.
func NewState[T any](initialValue T) *State[T] {
	return &State[T]{
		signal:        NewSignal(initialValue),
		updateQueue:   make([]StateUpdate[T], 0),
		changeHistory: make([]T, 0),
		historySize:   10, // Default history size
		onChange:      make([]func(old, new T), 0),
	}
}

// NewStateWithHistory creates a new State instance with a specific history size.
func NewStateWithHistory[T any](initialValue T, historySize int) *State[T] {
	return &State[T]{
		signal:        NewSignal(initialValue),
		updateQueue:   make([]StateUpdate[T], 0),
		changeHistory: make([]T, 0, historySize),
		historySize:   historySize,
		onChange:      make([]func(old, new T), 0),
	}
}

// NewStateWithEquals creates a new State instance with a custom equality function.
func NewStateWithEquals[T any](initialValue T, equals func(a, b T) bool) *State[T] {
	return &State[T]{
		signal:        NewSignalWithEquals(initialValue, equals),
		updateQueue:   make([]StateUpdate[T], 0),
		changeHistory: make([]T, 0),
		historySize:   10, // Default history size
		onChange:      make([]func(old, new T), 0),
	}
}

// Get returns the current state value.
func (s *State[T]) Get() T {
	return s.signal.Value()
}

// Set updates the state with a new value.
func (s *State[T]) Set(newValue T) {
	s.mutex.Lock()
	oldValue := s.signal.Value()

	// Check if the value is actually changing using the signal's equality function
	// This respects custom equality functions passed to NewStateWithEquals
	equals := false
	if s.signal.equalsFn != nil {
		// Use the custom equality function
		equals = s.signal.equalsFn(oldValue, newValue)
	} else {
		// Use the default string representation comparison from the signal
		oldStr := fmt.Sprintf("%v", oldValue)
		newStr := fmt.Sprintf("%v", newValue)
		equals = (oldStr == newStr)
	}

	// If values are equal according to the equality function, skip update
	if equals {
		s.mutex.Unlock()
		return
	}

	// Record in history if needed
	if s.historySize > 0 {
		if len(s.changeHistory) >= s.historySize {
			// Remove oldest entry if we've reached capacity
			s.changeHistory = s.changeHistory[1:]
		}
		s.changeHistory = append(s.changeHistory, oldValue)
	}

	// If we're batching updates, queue this update
	if atomic.LoadInt32(&s.batchDepth) > 0 {
		s.updateQueue = append(s.updateQueue, func(current T) T {
			return newValue
		})
		s.mutex.Unlock()
		return
	}

	// Apply the update immediately
	s.signal.Set(newValue)
	s.mutex.Unlock()

	// Notify listeners
	for _, fn := range s.onChange {
		fn(oldValue, newValue)
	}
}

// Update applies a function to update the state.
func (s *State[T]) Update(updateFn StateUpdate[T]) {
	s.mutex.Lock()
	oldValue := s.signal.Value()

	// Record in history if needed
	if s.historySize > 0 {
		if len(s.changeHistory) >= s.historySize {
			// Remove oldest entry if we've reached capacity
			s.changeHistory = s.changeHistory[1:]
		}
		s.changeHistory = append(s.changeHistory, oldValue)
	}

	// If we're batching updates, queue this update
	if atomic.LoadInt32(&s.batchDepth) > 0 {
		s.updateQueue = append(s.updateQueue, updateFn)
		s.mutex.Unlock()
		return
	}

	// Apply the update immediately
	newValue := updateFn(oldValue)
	s.signal.Set(newValue)
	s.mutex.Unlock()

	// Notify listeners
	for _, fn := range s.onChange {
		fn(oldValue, newValue)
	}
}

// Batch applies multiple state updates as a single operation.
func (s *State[T]) Batch(fn func()) {
	// Start batching
	wasBatching := atomic.LoadInt32(&s.batchDepth) > 0
	atomic.AddInt32(&s.batchDepth, 1)

	// Run the batched operations
	fn()

	// End batching
	newDepth := atomic.AddInt32(&s.batchDepth, -1)

	// If this was the outermost batch, apply all queued updates
	if newDepth == 0 && !wasBatching {
		s.mutex.Lock()
		if len(s.updateQueue) > 0 {
			// Get the current value
			oldValue := s.signal.Value()
			currentValue := oldValue

			// Record in history if needed
			if s.historySize > 0 {
				if len(s.changeHistory) >= s.historySize {
					// Remove oldest entry if we've reached capacity
					s.changeHistory = s.changeHistory[1:]
				}
				s.changeHistory = append(s.changeHistory, oldValue)
			}

			// Apply all updates sequentially
			for _, update := range s.updateQueue {
				currentValue = update(currentValue)
			}

			// Set the final value
			s.signal.Set(currentValue)

			// Clear the queue
			s.updateQueue = s.updateQueue[:0]

			// Store final value for callbacks
			newValue := currentValue
			s.mutex.Unlock()

			// Notify listeners
			for _, fn := range s.onChange {
				fn(oldValue, newValue)
			}
		} else {
			s.mutex.Unlock()
		}
	}
}

// OnChange registers a callback that will be called whenever the state changes.
func (s *State[T]) OnChange(callback func(old, new T)) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.onChange = append(s.onChange, callback)
}

// RemoveOnChange removes a previously registered callback.
func (s *State[T]) RemoveOnChange(callback func(old, new T)) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i, cb := range s.onChange {
		// This is a simplistic approach to compare functions
		// In Go, function values are not comparable with ==
		// We're using fmt.Sprintf as a workaround, but it's not 100% reliable
		if fmt.Sprintf("%p", cb) == fmt.Sprintf("%p", callback) {
			// Remove the callback by replacing it with the last one and truncating
			s.onChange[i] = s.onChange[len(s.onChange)-1]
			s.onChange = s.onChange[:len(s.onChange)-1]
			return true
		}
	}
	return false
}

// GetHistory returns the state change history.
func (s *State[T]) GetHistory() []T {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy to prevent external modifications
	history := make([]T, len(s.changeHistory))
	copy(history, s.changeHistory)
	return history
}

// ClearHistory clears the state change history.
func (s *State[T]) ClearHistory() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.changeHistory = s.changeHistory[:0]
}

// GetSignal returns the underlying signal for this state.
// This can be used for creating derived values and subscriptions.
func (s *State[T]) GetSignal() *Signal[T] {
	return s.signal
}
