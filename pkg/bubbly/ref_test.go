package bubbly

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewRef verifies that NewRef creates a ref with the initial value
func TestNewRef(t *testing.T) {
	tests := []struct {
		name     string
		initial  interface{}
		expected interface{}
	}{
		{
			name:     "integer ref",
			initial:  42,
			expected: 42,
		},
		{
			name:     "string ref",
			initial:  "hello",
			expected: "hello",
		},
		{
			name:     "zero value int",
			initial:  0,
			expected: 0,
		},
		{
			name:     "zero value string",
			initial:  "",
			expected: "",
		},
		{
			name:     "boolean ref",
			initial:  true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.initial.(type) {
			case int:
				ref := NewRef(v)
				assert.NotNil(t, ref, "NewRef should return non-nil reference")
				assert.Equal(t, tt.expected, ref.Get(), "Initial value should match")
			case string:
				ref := NewRef(v)
				assert.NotNil(t, ref, "NewRef should return non-nil reference")
				assert.Equal(t, tt.expected, ref.Get(), "Initial value should match")
			case bool:
				ref := NewRef(v)
				assert.NotNil(t, ref, "NewRef should return non-nil reference")
				assert.Equal(t, tt.expected, ref.Get(), "Initial value should match")
			}
		})
	}
}

// TestRef_Get verifies that Get returns the current value
func TestRef_Get(t *testing.T) {
	t.Run("get integer value", func(t *testing.T) {
		ref := NewRef(100)
		value := ref.Get()
		assert.Equal(t, 100, value, "Get should return current value")
	})

	t.Run("get string value", func(t *testing.T) {
		ref := NewRef("test")
		value := ref.Get()
		assert.Equal(t, "test", value, "Get should return current value")
	})

	t.Run("get struct value", func(t *testing.T) {
		type User struct {
			Name string
			Age  int
		}
		user := User{Name: "John", Age: 30}
		ref := NewRef(user)
		value := ref.Get()
		assert.Equal(t, user, value, "Get should return current struct value")
	})

	t.Run("get pointer value", func(t *testing.T) {
		val := 42
		ref := NewRef(&val)
		value := ref.Get()
		assert.Equal(t, &val, value, "Get should return current pointer value")
		assert.Equal(t, 42, *value, "Dereferenced value should be correct")
	})
}

// TestRef_Set verifies that Set updates the value
func TestRef_Set(t *testing.T) {
	t.Run("set integer value", func(t *testing.T) {
		ref := NewRef(10)
		assert.Equal(t, 10, ref.Get(), "Initial value should be 10")

		ref.Set(20)
		assert.Equal(t, 20, ref.Get(), "Value should be updated to 20")

		ref.Set(30)
		assert.Equal(t, 30, ref.Get(), "Value should be updated to 30")
	})

	t.Run("set string value", func(t *testing.T) {
		ref := NewRef("initial")
		assert.Equal(t, "initial", ref.Get(), "Initial value should be 'initial'")

		ref.Set("updated")
		assert.Equal(t, "updated", ref.Get(), "Value should be updated to 'updated'")
	})

	t.Run("set to zero value", func(t *testing.T) {
		ref := NewRef(100)
		ref.Set(0)
		assert.Equal(t, 0, ref.Get(), "Should be able to set to zero value")
	})

	t.Run("set struct value", func(t *testing.T) {
		type User struct {
			Name string
			Age  int
		}
		ref := NewRef(User{Name: "John", Age: 30})

		newUser := User{Name: "Jane", Age: 25}
		ref.Set(newUser)
		assert.Equal(t, newUser, ref.Get(), "Struct value should be updated")
	})
}

// TestRef_TypeSafety verifies compile-time type safety
func TestRef_TypeSafety(t *testing.T) {
	t.Run("int ref maintains type", func(t *testing.T) {
		ref := NewRef(42)
		// This should compile - same type
		ref.Set(100)
		assert.Equal(t, 100, ref.Get())
	})

	t.Run("string ref maintains type", func(t *testing.T) {
		ref := NewRef("hello")
		// This should compile - same type
		ref.Set("world")
		assert.Equal(t, "world", ref.Get())
	})

	// Note: Type mismatches are caught at compile time with generics
	// The following would not compile:
	// ref := NewRef(42)
	// ref.Set("string") // Compile error: cannot use "string" (type string) as type int
}

// TestRef_MultipleRefs verifies that multiple refs are independent
func TestRef_MultipleRefs(t *testing.T) {
	ref1 := NewRef(10)
	ref2 := NewRef(20)
	ref3 := NewRef("hello")

	assert.Equal(t, 10, ref1.Get(), "ref1 should have value 10")
	assert.Equal(t, 20, ref2.Get(), "ref2 should have value 20")
	assert.Equal(t, "hello", ref3.Get(), "ref3 should have value 'hello'")

	ref1.Set(100)
	assert.Equal(t, 100, ref1.Get(), "ref1 should be updated")
	assert.Equal(t, 20, ref2.Get(), "ref2 should remain unchanged")
	assert.Equal(t, "hello", ref3.Get(), "ref3 should remain unchanged")
}

// TestRef_ZeroValue verifies handling of zero values
func TestRef_ZeroValue(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "zero int",
			test: func(t *testing.T) {
				ref := NewRef(0)
				assert.Equal(t, 0, ref.Get())
			},
		},
		{
			name: "zero string",
			test: func(t *testing.T) {
				ref := NewRef("")
				assert.Equal(t, "", ref.Get())
			},
		},
		{
			name: "zero bool",
			test: func(t *testing.T) {
				ref := NewRef(false)
				assert.Equal(t, false, ref.Get())
			},
		},
		{
			name: "nil pointer",
			test: func(t *testing.T) {
				var ptr *int
				ref := NewRef(ptr)
				assert.Nil(t, ref.Get())
			},
		},
		{
			name: "nil slice",
			test: func(t *testing.T) {
				var slice []int
				ref := NewRef(slice)
				assert.Nil(t, ref.Get())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestRef_ComplexTypes verifies Ref works with complex types
func TestRef_ComplexTypes(t *testing.T) {
	t.Run("slice", func(t *testing.T) {
		ref := NewRef([]int{1, 2, 3})
		assert.Equal(t, []int{1, 2, 3}, ref.Get())

		ref.Set([]int{4, 5, 6})
		assert.Equal(t, []int{4, 5, 6}, ref.Get())
	})

	t.Run("map", func(t *testing.T) {
		ref := NewRef(map[string]int{"a": 1, "b": 2})
		value := ref.Get()
		assert.Equal(t, 1, value["a"])
		assert.Equal(t, 2, value["b"])

		ref.Set(map[string]int{"c": 3})
		value = ref.Get()
		assert.Equal(t, 3, value["c"])
	})

	t.Run("nested struct", func(t *testing.T) {
		type Address struct {
			City string
		}
		type User struct {
			Name    string
			Address Address
		}

		user := User{Name: "John", Address: Address{City: "NYC"}}
		ref := NewRef(user)
		assert.Equal(t, "NYC", ref.Get().Address.City)

		newUser := User{Name: "Jane", Address: Address{City: "LA"}}
		ref.Set(newUser)
		assert.Equal(t, "LA", ref.Get().Address.City)
	})
}
