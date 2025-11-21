package testutil

import (
	"sync"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
)

// TestNewShowTester tests ShowTester creation
func TestNewShowTester(t *testing.T) {
	tests := []struct {
		name       string
		visibleRef interface{}
		wantNil    bool
	}{
		{
			name:       "creates tester with valid ref",
			visibleRef: bubbly.NewRef(true),
			wantNil:    false,
		},
		{
			name:       "creates tester with false ref",
			visibleRef: bubbly.NewRef(false),
			wantNil:    false,
		},
		{
			name:       "creates tester with nil ref",
			visibleRef: nil,
			wantNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewShowTester(tt.visibleRef)
			if tt.wantNil {
				assert.Nil(t, tester)
			} else {
				assert.NotNil(t, tester)
			}
		})
	}
}

// TestShowTester_SetVisible tests setting visibility
func TestShowTester_SetVisible(t *testing.T) {
	tests := []struct {
		name     string
		initial  bool
		setValue bool
		expected bool
	}{
		{
			name:     "set true to false",
			initial:  true,
			setValue: false,
			expected: false,
		},
		{
			name:     "set false to true",
			initial:  false,
			setValue: true,
			expected: true,
		},
		{
			name:     "set true to true",
			initial:  true,
			setValue: true,
			expected: true,
		},
		{
			name:     "set false to false",
			initial:  false,
			setValue: false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visibleRef := bubbly.NewRef(tt.initial)
			tester := NewShowTester(visibleRef)

			tester.SetVisible(tt.setValue)

			actual := visibleRef.Get().(bool)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// TestShowTester_GetVisible tests getting visibility
func TestShowTester_GetVisible(t *testing.T) {
	tests := []struct {
		name     string
		initial  bool
		expected bool
	}{
		{
			name:     "get true",
			initial:  true,
			expected: true,
		},
		{
			name:     "get false",
			initial:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visibleRef := bubbly.NewRef(tt.initial)
			tester := NewShowTester(visibleRef)

			actual := tester.GetVisible()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// TestShowTester_AssertVisible tests visibility assertions
func TestShowTester_AssertVisible(t *testing.T) {
	tests := []struct {
		name      string
		visible   bool
		expected  bool
		shouldErr bool
	}{
		{
			name:      "assert true when true",
			visible:   true,
			expected:  true,
			shouldErr: false,
		},
		{
			name:      "assert false when false",
			visible:   false,
			expected:  false,
			shouldErr: false,
		},
		{
			name:      "assert true when false fails",
			visible:   false,
			expected:  true,
			shouldErr: true,
		},
		{
			name:      "assert false when true fails",
			visible:   true,
			expected:  false,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visibleRef := bubbly.NewRef(tt.visible)
			tester := NewShowTester(visibleRef)

			mockT := &mockTestingT{}
			tester.AssertVisible(mockT, tt.expected)

			if tt.shouldErr {
				assert.True(t, mockT.failed, "expected error to be called")
			} else {
				assert.False(t, mockT.failed, "expected no error")
			}
		})
	}
}

// TestShowTester_AssertHidden tests hidden assertion convenience method
func TestShowTester_AssertHidden(t *testing.T) {
	tests := []struct {
		name      string
		visible   bool
		shouldErr bool
	}{
		{
			name:      "assert hidden when false",
			visible:   false,
			shouldErr: false,
		},
		{
			name:      "assert hidden when true fails",
			visible:   true,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visibleRef := bubbly.NewRef(tt.visible)
			tester := NewShowTester(visibleRef)

			mockT := &mockTestingT{}
			tester.AssertHidden(mockT)

			if tt.shouldErr {
				assert.True(t, mockT.failed, "expected error to be called")
			} else {
				assert.False(t, mockT.failed, "expected no error")
			}
		})
	}
}

// TestShowTester_NilRef tests behavior with nil ref
func TestShowTester_NilRef(t *testing.T) {
	t.Run("SetVisible with nil ref is safe", func(t *testing.T) {
		tester := NewShowTester(nil)
		assert.NotPanics(t, func() {
			tester.SetVisible(true)
		})
	})

	t.Run("GetVisible with nil ref returns false", func(t *testing.T) {
		tester := NewShowTester(nil)
		assert.False(t, tester.GetVisible())
	})

	t.Run("AssertVisible with nil ref", func(t *testing.T) {
		tester := NewShowTester(nil)
		mockT := &mockTestingT{}
		tester.AssertVisible(mockT, false)
		assert.False(t, mockT.failed)
	})
}

// TestShowTester_ThreadSafety tests concurrent access
func TestShowTester_ThreadSafety(t *testing.T) {
	t.Run("concurrent SetVisible and GetVisible", func(t *testing.T) {
		visibleRef := bubbly.NewRef(false)
		tester := NewShowTester(visibleRef)

		var wg sync.WaitGroup
		iterations := 100

		// Concurrent writes
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func(val bool) {
				defer wg.Done()
				tester.SetVisible(val)
			}(i%2 == 0)
		}

		// Concurrent reads
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = tester.GetVisible()
			}()
		}

		wg.Wait()
		// Test passes if no race conditions detected
	})
}

// TestShowTester_Reactivity tests reactivity with Show directive
func TestShowTester_Reactivity(t *testing.T) {
	t.Run("changing visibility affects Show directive output", func(t *testing.T) {
		visibleRef := bubbly.NewRef(true)
		tester := NewShowTester(visibleRef)

		// Initially visible
		assert.True(t, tester.GetVisible())

		// Hide content
		tester.SetVisible(false)
		assert.False(t, tester.GetVisible())

		// Show content again
		tester.SetVisible(true)
		assert.True(t, tester.GetVisible())
	})
}

// TestShowTester_DifferenceFromIf tests Show vs If behavior
func TestShowTester_DifferenceFromIf(t *testing.T) {
	t.Run("Show keeps element in output with transition", func(t *testing.T) {
		// This test documents the difference between Show and If directives
		// Show: Toggles visibility (with optional [Hidden] marker)
		// If: Removes element from DOM completely

		visibleRef := bubbly.NewRef(false)
		tester := NewShowTester(visibleRef)

		// Show directive with transition would output "[Hidden]content"
		// If directive would output ""
		assert.False(t, tester.GetVisible())
		// The actual rendering difference is tested in the directive tests
	})
}
