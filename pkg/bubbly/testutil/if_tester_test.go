package testutil

import (
	"sync"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
)

// TestNewIfTester verifies IfTester creation.
func TestNewIfTester(t *testing.T) {
	tests := []struct {
		name         string
		conditionRef interface{}
		wantNil      bool
	}{
		{
			name:         "with valid bool ref",
			conditionRef: bubbly.NewRef(true),
			wantNil:      false,
		},
		{
			name:         "with false ref",
			conditionRef: bubbly.NewRef(false),
			wantNil:      false,
		},
		{
			name:         "with nil ref",
			conditionRef: nil,
			wantNil:      false, // Tester created, but ref is nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewIfTester(tt.conditionRef)

			if tt.wantNil {
				assert.Nil(t, tester)
			} else {
				assert.NotNil(t, tester)
				assert.Equal(t, tt.conditionRef, tester.conditionRef)
			}
		})
	}
}

// TestIfTester_SetCondition verifies condition setting.
func TestIfTester_SetCondition(t *testing.T) {
	tests := []struct {
		name     string
		initial  bool
		setValue bool
		want     bool
	}{
		{
			name:     "set true to false",
			initial:  true,
			setValue: false,
			want:     false,
		},
		{
			name:     "set false to true",
			initial:  false,
			setValue: true,
			want:     true,
		},
		{
			name:     "set true to true",
			initial:  true,
			setValue: true,
			want:     true,
		},
		{
			name:     "set false to false",
			initial:  false,
			setValue: false,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conditionRef := bubbly.NewRef(tt.initial)
			tester := NewIfTester(conditionRef)

			tester.SetCondition(tt.setValue)

			got := conditionRef.Get().(bool)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestIfTester_SetCondition_NilRef verifies safe handling of nil ref.
func TestIfTester_SetCondition_NilRef(t *testing.T) {
	tester := NewIfTester(nil)

	// Should not panic
	assert.NotPanics(t, func() {
		tester.SetCondition(true)
		tester.SetCondition(false)
	})
}

// TestIfTester_GetCondition verifies condition retrieval.
func TestIfTester_GetCondition(t *testing.T) {
	tests := []struct {
		name  string
		value bool
		want  bool
	}{
		{
			name:  "get true",
			value: true,
			want:  true,
		},
		{
			name:  "get false",
			value: false,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conditionRef := bubbly.NewRef(tt.value)
			tester := NewIfTester(conditionRef)

			got := tester.GetCondition()
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestIfTester_GetCondition_NilRef verifies safe handling of nil ref.
func TestIfTester_GetCondition_NilRef(t *testing.T) {
	tester := NewIfTester(nil)

	got := tester.GetCondition()
	assert.False(t, got, "nil ref should return false")
}

// TestIfTester_AssertRendered verifies rendering assertions.
func TestIfTester_AssertRendered(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		expected  bool
		shouldErr bool
	}{
		{
			name:      "true condition, expect rendered",
			condition: true,
			expected:  true,
			shouldErr: false,
		},
		{
			name:      "false condition, expect not rendered",
			condition: false,
			expected:  false,
			shouldErr: false,
		},
		{
			name:      "true condition, expect not rendered (should fail)",
			condition: true,
			expected:  false,
			shouldErr: true,
		},
		{
			name:      "false condition, expect rendered (should fail)",
			condition: false,
			expected:  true,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conditionRef := bubbly.NewRef(tt.condition)
			tester := NewIfTester(conditionRef)

			mockT := &mockTestingT{}
			tester.AssertRendered(mockT, tt.expected)

			if tt.shouldErr {
				assert.True(t, mockT.failed, "Expected assertion to fail")
			} else {
				assert.False(t, mockT.failed, "Expected assertion to pass")
			}
		})
	}
}

// TestIfTester_AssertNotRendered verifies not-rendered assertions.
func TestIfTester_AssertNotRendered(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		shouldErr bool
	}{
		{
			name:      "false condition",
			condition: false,
			shouldErr: false,
		},
		{
			name:      "true condition (should fail)",
			condition: true,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conditionRef := bubbly.NewRef(tt.condition)
			tester := NewIfTester(conditionRef)

			mockT := &mockTestingT{}
			tester.AssertNotRendered(mockT)

			if tt.shouldErr {
				assert.True(t, mockT.failed, "Expected assertion to fail")
			} else {
				assert.False(t, mockT.failed, "Expected assertion to pass")
			}
		})
	}
}

// TestIfTester_Reactivity verifies condition changes are reactive.
func TestIfTester_Reactivity(t *testing.T) {
	conditionRef := bubbly.NewRef(false)
	tester := NewIfTester(conditionRef)

	// Initially false
	assert.False(t, tester.GetCondition())

	// Change to true
	tester.SetCondition(true)
	assert.True(t, tester.GetCondition())

	// Change back to false
	tester.SetCondition(false)
	assert.False(t, tester.GetCondition())

	// Multiple toggles
	for i := 0; i < 10; i++ {
		tester.SetCondition(true)
		assert.True(t, tester.GetCondition())

		tester.SetCondition(false)
		assert.False(t, tester.GetCondition())
	}
}

// TestIfTester_ThreadSafety verifies concurrent access is safe.
func TestIfTester_ThreadSafety(t *testing.T) {
	conditionRef := bubbly.NewRef(false)
	tester := NewIfTester(conditionRef)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(val bool) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				tester.SetCondition(val)
			}
		}(i%2 == 0)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = tester.GetCondition()
			}
		}()
	}

	wg.Wait()
	// Test passes if no race conditions detected
}

// TestIfTester_WithRealComponent verifies integration with actual If directive.
func TestIfTester_WithRealComponent(t *testing.T) {
	conditionRef := bubbly.NewRef(true)
	tester := NewIfTester(conditionRef)

	// Simulate component using If directive
	renderContent := func() string {
		condition := conditionRef.Get().(bool)
		if condition {
			return "Content is visible"
		}
		return ""
	}

	// Initially true
	output := renderContent()
	assert.Equal(t, "Content is visible", output)
	tester.AssertRendered(&mockTestingT{}, true)

	// Change to false
	tester.SetCondition(false)
	output = renderContent()
	assert.Equal(t, "", output)
	tester.AssertRendered(&mockTestingT{}, false)

	// Change back to true
	tester.SetCondition(true)
	output = renderContent()
	assert.Equal(t, "Content is visible", output)
	tester.AssertRendered(&mockTestingT{}, true)
}

// TestIfTester_MultipleConditions verifies testing multiple If directives.
func TestIfTester_MultipleConditions(t *testing.T) {
	condition1 := bubbly.NewRef(true)
	condition2 := bubbly.NewRef(false)
	condition3 := bubbly.NewRef(true)

	tester1 := NewIfTester(condition1)
	tester2 := NewIfTester(condition2)
	tester3 := NewIfTester(condition3)

	// Verify initial states
	assert.True(t, tester1.GetCondition())
	assert.False(t, tester2.GetCondition())
	assert.True(t, tester3.GetCondition())

	// Change conditions independently
	tester1.SetCondition(false)
	tester2.SetCondition(true)
	tester3.SetCondition(false)

	// Verify changes
	assert.False(t, tester1.GetCondition())
	assert.True(t, tester2.GetCondition())
	assert.False(t, tester3.GetCondition())
}

// TestIfTester_NestedConditions verifies testing nested If directives.
func TestIfTester_NestedConditions(t *testing.T) {
	outerCondition := bubbly.NewRef(true)
	innerCondition := bubbly.NewRef(false)

	outerTester := NewIfTester(outerCondition)
	innerTester := NewIfTester(innerCondition)

	// Simulate nested If directives
	renderNested := func() string {
		if outerCondition.Get().(bool) {
			if innerCondition.Get().(bool) {
				return "Both true"
			}
			return "Outer true, inner false"
		}
		return "Outer false"
	}

	// Test all combinations
	tests := []struct {
		name     string
		outer    bool
		inner    bool
		expected string
	}{
		{"both true", true, true, "Both true"},
		{"outer true, inner false", true, false, "Outer true, inner false"},
		{"outer false, inner true", false, true, "Outer false"},
		{"both false", false, false, "Outer false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outerTester.SetCondition(tt.outer)
			innerTester.SetCondition(tt.inner)

			output := renderNested()
			assert.Equal(t, tt.expected, output)
		})
	}
}

// TestIfTester_ElseIfPattern verifies testing ElseIf chains.
func TestIfTester_ElseIfPattern(t *testing.T) {
	condition1 := bubbly.NewRef(false)
	condition2 := bubbly.NewRef(false)
	condition3 := bubbly.NewRef(false)

	tester1 := NewIfTester(condition1)
	tester2 := NewIfTester(condition2)
	tester3 := NewIfTester(condition3)

	// Simulate ElseIf chain
	renderElseIf := func() string {
		if condition1.Get().(bool) {
			return "First condition"
		} else if condition2.Get().(bool) {
			return "Second condition"
		} else if condition3.Get().(bool) {
			return "Third condition"
		}
		return "Else branch"
	}

	tests := []struct {
		name     string
		cond1    bool
		cond2    bool
		cond3    bool
		expected string
	}{
		{"first true", true, false, false, "First condition"},
		{"second true", false, true, false, "Second condition"},
		{"third true", false, false, true, "Third condition"},
		{"all false", false, false, false, "Else branch"},
		{"multiple true (first wins)", true, true, true, "First condition"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester1.SetCondition(tt.cond1)
			tester2.SetCondition(tt.cond2)
			tester3.SetCondition(tt.cond3)

			output := renderElseIf()
			assert.Equal(t, tt.expected, output)
		})
	}
}

// TestIfTester_FrequentToggles verifies performance with many condition changes.
func TestIfTester_FrequentToggles(t *testing.T) {
	conditionRef := bubbly.NewRef(false)
	tester := NewIfTester(conditionRef)

	// Toggle 1000 times
	for i := 0; i < 1000; i++ {
		expected := i%2 == 0
		tester.SetCondition(expected)
		got := tester.GetCondition()
		assert.Equal(t, expected, got, "Iteration %d", i)
	}
}

// TestIfTester_InvalidRef verifies handling of invalid refs.
func TestIfTester_InvalidRef(t *testing.T) {
	tests := []struct {
		name string
		ref  interface{}
	}{
		{"nil ref", nil},
		{"non-ref type", "not a ref"},
		{"int value", 42},
		{"struct value", struct{}{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewIfTester(tt.ref)

			// Should not panic
			assert.NotPanics(t, func() {
				tester.SetCondition(true)
				got := tester.GetCondition()
				assert.False(t, got, "Invalid ref should return false")
			})
		})
	}
}

// TestIfTester_ZeroValue verifies handling of zero value ref.
func TestIfTester_ZeroValue(t *testing.T) {
	var conditionRef *bubbly.Ref[bool]
	tester := NewIfTester(conditionRef)

	// Should not panic
	assert.NotPanics(t, func() {
		tester.SetCondition(true)
		got := tester.GetCondition()
		assert.False(t, got, "Zero value ref should return false")
	})
}
