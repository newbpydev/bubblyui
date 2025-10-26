package bubbly

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestWatch_NilCallback tests that Watch panics with ErrNilCallback when given a nil callback
func TestWatch_NilCallback(t *testing.T) {
	ref := NewRef(0)

	assert.PanicsWithValue(t, ErrNilCallback, func() {
		Watch(ref, nil)
	}, "Watch should panic with ErrNilCallback when callback is nil")
}

// TestNewComputed_NilFunction tests that NewComputed panics with ErrNilComputeFn when given a nil function
func TestNewComputed_NilFunction(t *testing.T) {
	assert.PanicsWithValue(t, ErrNilComputeFn, func() {
		NewComputed[int](nil)
	}, "NewComputed should panic with ErrNilComputeFn when function is nil")
}

// TestCircularDependency tests detection of circular dependencies in computed values
func TestCircularDependency(t *testing.T) {
	t.Skip("Circular dependency detection requires per-goroutine tracker - deferred to future enhancement")

	t.Run("direct circular dependency", func(t *testing.T) {
		// Create two computed values that depend on each other
		var a, b *Computed[int]

		a = NewComputed(func() int {
			if b != nil {
				return b.Get() + 1
			}
			return 0
		})

		b = NewComputed(func() int {
			return a.Get() + 1 // Circular!
		})

		// Accessing either should detect the cycle
		assert.Panics(t, func() {
			a.Get()
		}, "Should panic on circular dependency")
	})

	t.Run("indirect circular dependency", func(t *testing.T) {
		// Create a chain: A -> B -> C -> A
		var a, b, c *Computed[int]

		a = NewComputed(func() int {
			if c != nil {
				return c.Get() + 1
			}
			return 0
		})

		b = NewComputed(func() int {
			return a.Get() + 1
		})

		c = NewComputed(func() int {
			return b.Get() + 1 // Circular!
		})

		// Accessing should detect the cycle
		assert.Panics(t, func() {
			a.Get()
		}, "Should panic on indirect circular dependency")
	})

	t.Run("self-referencing computed", func(t *testing.T) {
		var self *Computed[int]

		self = NewComputed(func() int {
			if self != nil {
				return self.Get() + 1 // Self-reference!
			}
			return 0
		})

		// Should detect self-reference
		assert.Panics(t, func() {
			self.Get()
		}, "Should panic on self-referencing computed")
	})
}

// TestMaxDepthExceeded tests that deeply nested dependencies are detected
func TestMaxDepthExceeded(t *testing.T) {
	t.Run("exactly at max depth is ok", func(t *testing.T) {
		// Create a chain of exactly MaxDependencyDepth (100) computed values
		base := NewRef(1)
		current := NewComputed(func() int {
			return base.Get()
		})

		// Create 99 more levels (total 100)
		for i := 1; i < MaxDependencyDepth; i++ {
			prev := current
			current = NewComputed(func() int {
				return prev.Get() + 1
			})
		}

		// Should work fine
		assert.NotPanics(t, func() {
			result := current.Get()
			assert.Equal(t, MaxDependencyDepth, result)
		}, "Should not panic at exactly max depth")
	})

	t.Run("exceeding max depth panics", func(t *testing.T) {
		// Create a chain of MaxDependencyDepth + 1 computed values
		base := NewRef(1)
		current := NewComputed(func() int {
			return base.Get()
		})

		// Create MaxDependencyDepth more levels (total 101)
		for i := 1; i <= MaxDependencyDepth; i++ {
			prev := current
			current = NewComputed(func() int {
				return prev.Get() + 1
			})
		}

		// Should panic
		assert.Panics(t, func() {
			current.Get()
		}, "Should panic when exceeding max depth")
	})
}

// TestErrorMessages tests that error messages are clear and helpful
func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrNilCallback",
			err:      ErrNilCallback,
			expected: "callback cannot be nil",
		},
		{
			name:     "ErrNilComputeFn",
			err:      ErrNilComputeFn,
			expected: "compute function cannot be nil",
		},
		{
			name:     "ErrCircularDependency",
			err:      ErrCircularDependency,
			expected: "circular dependency detected",
		},
		{
			name:     "ErrMaxDepthExceeded",
			err:      ErrMaxDepthExceeded,
			expected: "max dependency depth exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error(), "Error message should be clear")
		})
	}
}

// TestNoFalsePositives tests that valid usage doesn't trigger errors
func TestNoFalsePositives(t *testing.T) {
	// Reset tracker before each test to ensure clean state
	defer globalTracker.Reset()

	t.Run("valid watch doesn't panic", func(t *testing.T) {
		globalTracker.Reset()
		ref := NewRef(0)
		assert.NotPanics(t, func() {
			cleanup := Watch(ref, func(n, o int) {})
			defer cleanup()
		}, "Valid Watch should not panic")
	})

	t.Run("valid computed doesn't panic", func(t *testing.T) {
		globalTracker.Reset()
		assert.NotPanics(t, func() {
			computed := NewComputed(func() int { return 42 })
			_ = computed.Get()
		}, "Valid NewComputed should not panic")
	})

	t.Run("complex valid dependency chain", func(t *testing.T) {
		globalTracker.Reset()
		// Create a valid chain without cycles
		base := NewRef(10)
		doubled := NewComputed(func() int {
			return base.Get() * 2
		})
		quadrupled := NewComputed(func() int {
			return doubled.Get() * 2
		})
		octupled := NewComputed(func() int {
			return quadrupled.Get() * 2
		})

		assert.NotPanics(t, func() {
			result := octupled.Get()
			assert.Equal(t, 80, result)
		}, "Valid dependency chain should not panic")
	})
}
