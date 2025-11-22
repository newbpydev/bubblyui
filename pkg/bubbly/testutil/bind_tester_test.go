package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
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

// TestBindTester_InvalidRefValue tests invalid ref value (not a valid reflect.Value)
func TestBindTester_InvalidRefValue(t *testing.T) {
	// Create a tester with an invalid value (not a Ref)
	tester := &BindTester{
		ref: "not a ref", // String instead of *Ref
	}

	// Should not panic, just no-op
	tester.TriggerElementChange("value")

	// GetCurrentValue should return nil for invalid ref
	value := tester.GetCurrentValue()
	assert.Nil(t, value)
}

// TestBindTester_NilPointerRef tests nil pointer ref
func TestBindTester_NilPointerRef(t *testing.T) {
	// Create a nil pointer to a Ref
	var nilRef *bubbly.Ref[string]
	tester := NewBindTester(nilRef)

	// Should not panic, just no-op
	tester.TriggerElementChange("value")

	// GetCurrentValue should return nil
	value := tester.GetCurrentValue()
	assert.Nil(t, value)
}

// TestBindTester_RefWithoutSetMethod tests ref without Set method
func TestBindTester_RefWithoutSetMethod(t *testing.T) {
	// Create a struct without Set method
	type FakeRef struct {
		value string
	}

	fakeRef := &FakeRef{value: "initial"}
	tester := NewBindTester(fakeRef)

	// Should not panic, just no-op
	tester.TriggerElementChange("new value")

	// Value should remain unchanged
	assert.Equal(t, "initial", fakeRef.value)
}

// FakeRefWithSetOnly has Set method but no Get method
type FakeRefWithSetOnly struct {
	value string
}

func (f *FakeRefWithSetOnly) Set(v string) {
	f.value = v
}

// TestBindTester_RefWithoutGetMethod tests ref without Get method
func TestBindTester_RefWithoutGetMethod(t *testing.T) {
	fakeRef := &FakeRefWithSetOnly{value: "initial"}
	tester := NewBindTester(fakeRef)

	// Should not panic, just no-op (no Get method to determine type)
	tester.TriggerElementChange("new value")

	// Value should remain unchanged
	assert.Equal(t, "initial", fakeRef.value)
}

// FakeRefWithEmptyGetResult has Get method that returns nothing
type FakeRefWithEmptyGetResult struct{}

func (f *FakeRefWithEmptyGetResult) Get() {
	// Returns nothing - no return value
}

func (f *FakeRefWithEmptyGetResult) Set(v interface{}) {
	// No-op
}

// TestBindTester_GetMethodReturnsEmpty tests Get method returning no results
func TestBindTester_GetMethodReturnsEmpty(t *testing.T) {
	fakeRef := &FakeRefWithEmptyGetResult{}
	tester := NewBindTester(fakeRef)

	// Should not panic, just no-op (Get returns nothing)
	tester.TriggerElementChange("value")
}

// TestBindTester_GetCurrentValue_InvalidRef tests GetCurrentValue with invalid ref
func TestBindTester_GetCurrentValue_InvalidRef(t *testing.T) {
	tester := &BindTester{
		ref: "not a ref",
	}

	value := tester.GetCurrentValue()
	assert.Nil(t, value)
}

// TestBindTester_GetCurrentValue_RefWithoutGetMethod tests GetCurrentValue without Get method
func TestBindTester_GetCurrentValue_RefWithoutGetMethod(t *testing.T) {
	type FakeRef struct {
		value string
	}

	tester := &BindTester{
		ref: &FakeRef{value: "test"},
	}

	value := tester.GetCurrentValue()
	assert.Nil(t, value)
}

// FakeRefWithEmptyGet is a type with Get method that returns nothing
type FakeRefWithEmptyGet struct{}

func (f *FakeRefWithEmptyGet) Get() {
	// Returns nothing - this is intentional for testing
}

// TestBindTester_GetCurrentValue_GetReturnsEmpty tests Get method returning empty results
func TestBindTester_GetCurrentValue_GetReturnsEmpty(t *testing.T) {
	tester := &BindTester{
		ref: &FakeRefWithEmptyGet{},
	}

	value := tester.GetCurrentValue()
	assert.Nil(t, value)
}

// TestBindTester_AssertRefUpdated_InvalidRef tests AssertRefUpdated with invalid ref
func TestBindTester_AssertRefUpdated_InvalidRef(t *testing.T) {
	tester := &BindTester{
		ref: "not a ref",
	}

	mockT := &mockTestingT{}
	tester.AssertRefUpdated(mockT, "expected")

	// Should fail with appropriate error
	assert.True(t, mockT.failed)
}

// TestBindTester_ConvertToType_AllBranches tests all conversion branches
func TestBindTester_ConvertToType_AllBranches(t *testing.T) {
	tests := []struct {
		name     string
		initial  interface{}
		input    interface{}
		expected interface{}
	}{
		// Int variants
		{"int8 conversion", int8(0), "42", int8(42)},
		{"int16 conversion", int16(0), "42", int16(42)},
		{"int32 conversion", int32(0), "42", int32(42)},

		// Uint variants
		{"uint8 conversion", uint8(0), "42", uint8(42)},
		{"uint16 conversion", uint16(0), "42", uint16(42)},
		{"uint32 conversion", uint32(0), "42", uint32(42)},
		{"uint64 conversion", uint64(0), "42", uint64(42)},

		// Float variants
		{"float64 conversion", float64(0), "3.14", float64(3.14)},

		// Invalid conversions
		{"invalid int8", int8(5), "invalid", int8(0)},
		{"invalid uint8", uint8(5), "invalid", uint8(0)},
		{"invalid float32", float32(5), "invalid", float32(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := bubbly.NewRef(tt.initial)
			tester := NewBindTester(ref)

			tester.TriggerElementChange(tt.input)

			actual := ref.Get()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// TestBindTester_ConvertToType_DirectConversion tests direct type conversion
func TestBindTester_ConvertToType_DirectConversion(t *testing.T) {
	// Test convertible types (int to int64, etc.)
	ref := bubbly.NewRef(int64(0))
	tester := NewBindTester(ref)

	// Pass int, should convert to int64
	tester.TriggerElementChange(int(42))
	assert.Equal(t, int64(42), ref.Get())
}

// TestBindTester_ConvertToType_UnconvertibleType tests unconvertible types
func TestBindTester_ConvertToType_UnconvertibleType(t *testing.T) {
	// Test type that can't be converted
	type CustomType struct {
		value string
	}

	ref := bubbly.NewRef(CustomType{value: "initial"})
	tester := NewBindTester(ref)

	// Try to set with incompatible type
	tester.TriggerElementChange("string value")

	// Should get zero value
	result := ref.Get().(CustomType)
	assert.Equal(t, "", result.value)
}

// TestBindTester_ConcurrentGetCurrentValue tests concurrent GetCurrentValue calls
func TestBindTester_ConcurrentGetCurrentValue(t *testing.T) {
	ref := bubbly.NewRef("test value")
	tester := NewBindTester(ref)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			value := tester.GetCurrentValue()
			assert.NotNil(t, value)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestBindTester_ConcurrentAssertRefUpdated tests concurrent assertions
func TestBindTester_ConcurrentAssertRefUpdated(t *testing.T) {
	ref := bubbly.NewRef("test")
	tester := NewBindTester(ref)

	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func() {
			tester.AssertRefUpdated(t, "test")
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}
