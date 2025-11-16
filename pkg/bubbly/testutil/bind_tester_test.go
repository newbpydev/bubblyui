package testutil

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
)

// TestBindTester_Creation tests basic tester creation
func TestBindTester_Creation(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{"string ref", "hello"},
		{"int ref", 42},
		{"bool ref", true},
		{"float ref", 3.14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := bubbly.NewRef(tt.value)
			tester := NewBindTester(ref)

			assert.NotNil(t, tester)
			assert.Equal(t, tt.value, ref.Get())
		})
	}
}

// TestBindTester_TriggerElementChange tests simulating element changes
func TestBindTester_TriggerElementChange(t *testing.T) {
	tests := []struct {
		name     string
		initial  interface{}
		newValue interface{}
	}{
		{"string change", "hello", "world"},
		{"int change", 10, 20},
		{"bool change", false, true},
		{"float change", 1.5, 2.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := bubbly.NewRef(tt.initial)
			tester := NewBindTester(ref)

			// Trigger element change (simulates user input)
			tester.TriggerElementChange(tt.newValue)

			// Ref should be updated
			assert.Equal(t, tt.newValue, ref.Get())
		})
	}
}

// TestBindTester_AssertRefUpdated tests ref update assertions
func TestBindTester_AssertRefUpdated(t *testing.T) {
	tests := []struct {
		name     string
		initial  interface{}
		newValue interface{}
		expected interface{}
	}{
		{"string update", "old", "new", "new"},
		{"int update", 5, 10, 10},
		{"bool update", false, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := bubbly.NewRef(tt.initial)
			tester := NewBindTester(ref)

			tester.TriggerElementChange(tt.newValue)

			// This should pass
			tester.AssertRefUpdated(t, tt.expected)
		})
	}
}

// TestBindTester_AssertRefUpdated_Failure tests assertion failures
func TestBindTester_AssertRefUpdated_Failure(t *testing.T) {
	ref := bubbly.NewRef("initial")
	tester := NewBindTester(ref)

	tester.TriggerElementChange("changed")

	// Use mock testing.T to capture error
	mockT := &mockTestingT{}
	tester.AssertRefUpdated(mockT, "wrong_value")

	assert.True(t, mockT.failed, "Expected error to be called")
	assert.NotEmpty(t, mockT.errors, "Expected error message")
	assert.Contains(t, mockT.errors[0], "expected")
	assert.Contains(t, mockT.errors[0], "wrong_value")
	assert.Contains(t, mockT.errors[0], "got")
	assert.Contains(t, mockT.errors[0], "changed")
}

// TestBindTester_TwoWayBinding tests bidirectional binding
func TestBindTester_TwoWayBinding(t *testing.T) {
	t.Run("ref to element", func(t *testing.T) {
		ref := bubbly.NewRef("initial")
		tester := NewBindTester(ref)

		// Change ref value
		ref.Set("updated")

		// Element should reflect the change
		assert.Equal(t, "updated", ref.Get())
		tester.AssertRefUpdated(t, "updated")
	})

	t.Run("element to ref", func(t *testing.T) {
		ref := bubbly.NewRef("initial")
		tester := NewBindTester(ref)

		// Simulate element change
		tester.TriggerElementChange("from_element")

		// Ref should be updated
		assert.Equal(t, "from_element", ref.Get())
		tester.AssertRefUpdated(t, "from_element")
	})
}

// TestBindTester_TypeConversion tests type conversion for different types
func TestBindTester_TypeConversion(t *testing.T) {
	t.Run("string to int", func(t *testing.T) {
		ref := bubbly.NewRef(0)
		tester := NewBindTester(ref)

		// Simulate string input that should be converted to int
		tester.TriggerElementChange("42")

		assert.Equal(t, 42, ref.Get())
	})

	t.Run("string to bool", func(t *testing.T) {
		ref := bubbly.NewRef(false)
		tester := NewBindTester(ref)

		// Simulate string input that should be converted to bool
		tester.TriggerElementChange("true")

		assert.Equal(t, true, ref.Get())
	})

	t.Run("string to float", func(t *testing.T) {
		ref := bubbly.NewRef(0.0)
		tester := NewBindTester(ref)

		// Simulate string input that should be converted to float
		tester.TriggerElementChange("3.14")

		assert.Equal(t, 3.14, ref.Get())
	})
}

// TestBindTester_InvalidTypeConversion tests handling of invalid conversions
func TestBindTester_InvalidTypeConversion(t *testing.T) {
	t.Run("invalid int", func(t *testing.T) {
		ref := bubbly.NewRef(0)
		tester := NewBindTester(ref)

		// Simulate invalid int input
		tester.TriggerElementChange("not_a_number")

		// Should default to zero value
		assert.Equal(t, 0, ref.Get())
	})

	t.Run("invalid float", func(t *testing.T) {
		ref := bubbly.NewRef(0.0)
		tester := NewBindTester(ref)

		// Simulate invalid float input
		tester.TriggerElementChange("not_a_float")

		// Should default to zero value
		assert.Equal(t, 0.0, ref.Get())
	})
}

// TestBindTester_NilRef tests handling of nil ref
func TestBindTester_NilRef(t *testing.T) {
	// Creating tester with nil ref should not panic
	tester := NewBindTester(nil)
	assert.NotNil(t, tester)

	// Operations should be safe (no-ops)
	tester.TriggerElementChange("value")

	// Assertions should fail gracefully
	mockT := &mockTestingT{}
	tester.AssertRefUpdated(mockT, "value")
	assert.True(t, mockT.failed)
}

// TestBindTester_GetCurrentValue tests getting current ref value
func TestBindTester_GetCurrentValue(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{"string", "hello"},
		{"int", 42},
		{"bool", true},
		{"float", 3.14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := bubbly.NewRef(tt.value)
			tester := NewBindTester(ref)

			current := tester.GetCurrentValue()
			assert.Equal(t, tt.value, current)
		})
	}
}

// TestBindTester_MultipleChanges tests multiple sequential changes
func TestBindTester_MultipleChanges(t *testing.T) {
	ref := bubbly.NewRef("initial")
	tester := NewBindTester(ref)

	// First change
	tester.TriggerElementChange("first")
	assert.Equal(t, "first", ref.Get())

	// Second change
	tester.TriggerElementChange("second")
	assert.Equal(t, "second", ref.Get())

	// Third change
	tester.TriggerElementChange("third")
	assert.Equal(t, "third", ref.Get())

	// Final assertion
	tester.AssertRefUpdated(t, "third")
}

// TestBindTester_EmptyString tests handling of empty strings
func TestBindTester_EmptyString(t *testing.T) {
	ref := bubbly.NewRef("initial")
	tester := NewBindTester(ref)

	// Change to empty string
	tester.TriggerElementChange("")

	assert.Equal(t, "", ref.Get())
	tester.AssertRefUpdated(t, "")
}

// TestBindTester_ZeroValues tests handling of zero values
func TestBindTester_ZeroValues(t *testing.T) {
	t.Run("int zero", func(t *testing.T) {
		ref := bubbly.NewRef(42)
		tester := NewBindTester(ref)

		tester.TriggerElementChange(0)
		assert.Equal(t, 0, ref.Get())
	})

	t.Run("bool false", func(t *testing.T) {
		ref := bubbly.NewRef(true)
		tester := NewBindTester(ref)

		tester.TriggerElementChange(false)
		assert.Equal(t, false, ref.Get())
	})

	t.Run("float zero", func(t *testing.T) {
		ref := bubbly.NewRef(3.14)
		tester := NewBindTester(ref)

		tester.TriggerElementChange(0.0)
		assert.Equal(t, 0.0, ref.Get())
	})
}

// TestBindTester_GetCurrentValue_NilRef tests GetCurrentValue with nil ref
func TestBindTester_GetCurrentValue_NilRef(t *testing.T) {
	tester := NewBindTester(nil)
	value := tester.GetCurrentValue()
	assert.Nil(t, value)
}

// TestBindTester_UnsignedIntConversion tests unsigned integer conversions
func TestBindTester_UnsignedIntConversion(t *testing.T) {
	t.Run("string to uint", func(t *testing.T) {
		ref := bubbly.NewRef(uint(0))
		tester := NewBindTester(ref)

		tester.TriggerElementChange("42")
		assert.Equal(t, uint(42), ref.Get())
	})

	t.Run("invalid uint", func(t *testing.T) {
		ref := bubbly.NewRef(uint(10))
		tester := NewBindTester(ref)

		tester.TriggerElementChange("invalid")
		assert.Equal(t, uint(0), ref.Get())
	})

	t.Run("negative to uint", func(t *testing.T) {
		ref := bubbly.NewRef(uint(10))
		tester := NewBindTester(ref)

		tester.TriggerElementChange("-5")
		assert.Equal(t, uint(0), ref.Get())
	})
}

// TestBindTester_BoolStringConversions tests boolean string conversions
func TestBindTester_BoolStringConversions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"true string", "true", true},
		{"1 string", "1", true},
		{"false string", "false", false},
		{"0 string", "0", false},
		{"random string", "random", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := bubbly.NewRef(false)
			tester := NewBindTester(ref)

			tester.TriggerElementChange(tt.input)
			assert.Equal(t, tt.expected, ref.Get())
		})
	}
}

// TestBindTester_Int64Conversion tests int64 conversions
func TestBindTester_Int64Conversion(t *testing.T) {
	ref := bubbly.NewRef(int64(0))
	tester := NewBindTester(ref)

	tester.TriggerElementChange("9223372036854775807") // max int64
	assert.Equal(t, int64(9223372036854775807), ref.Get())
}

// TestBindTester_Float32Conversion tests float32 conversions
func TestBindTester_Float32Conversion(t *testing.T) {
	ref := bubbly.NewRef(float32(0.0))
	tester := NewBindTester(ref)

	tester.TriggerElementChange("3.14")
	assert.Equal(t, float32(3.14), ref.Get())
}

// TestBindTester_NilValueConversion tests nil value conversion
func TestBindTester_NilValueConversion(t *testing.T) {
	ref := bubbly.NewRef("initial")
	tester := NewBindTester(ref)

	tester.TriggerElementChange(nil)
	// Should convert to zero value (empty string)
	assert.Equal(t, "", ref.Get())
}

// TestBindTester_ThreadSafety tests concurrent access
func TestBindTester_ThreadSafety(t *testing.T) {
	ref := bubbly.NewRef(0)
	tester := NewBindTester(ref)

	// Concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(val int) {
			tester.TriggerElementChange(val)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Just verify it didn't panic and has some value
	value := tester.GetCurrentValue()
	assert.NotNil(t, value)
}
