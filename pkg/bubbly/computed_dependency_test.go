package bubbly

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestComputed_ImplementsDependency verifies Computed implements the Dependency interface
func TestComputed_ImplementsDependency(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Computed implements Dependency interface",
			test: func(t *testing.T) {
				count := NewRef(5)
				computed := NewComputed(func() int {
					return count.GetTyped() * 2
				})
				var _ Dependency = computed
			},
		},
		{
			name: "Get() any returns value",
			test: func(t *testing.T) {
				count := NewRef(5)
				computed := NewComputed(func() int {
					return count.GetTyped() * 2
				})
				value := computed.Get()
				assert.Equal(t, 10, value)
			},
		},
		{
			name: "Get() any can be type asserted",
			test: func(t *testing.T) {
				count := NewRef(5)
				computed := NewComputed(func() int {
					return count.GetTyped() * 2
				})
				value := computed.Get()

				// Type assertion should work
				intValue, ok := value.(int)
				assert.True(t, ok, "should be able to type assert to int")
				assert.Equal(t, 10, intValue)
			},
		},
		{
			name: "GetTyped() T returns typed value",
			test: func(t *testing.T) {
				count := NewRef(5)
				computed := NewComputed(func() int {
					return count.GetTyped() * 2
				})
				value := computed.GetTyped()

				// Should be int, not any
				assert.Equal(t, 10, value)

				// Verify it's actually type int at compile time
				var _ int = value
			},
		},
		{
			name: "Get() and GetTyped() return same value",
			test: func(t *testing.T) {
				count := NewRef(5)
				computed := NewComputed(func() int {
					return count.GetTyped() * 2
				})

				anyValue := computed.Get()
				typedValue := computed.GetTyped()

				assert.Equal(t, typedValue, anyValue.(int))
			},
		},
		{
			name: "Computed can be used as Dependency in slice",
			test: func(t *testing.T) {
				ref1 := NewRef(1)
				ref2 := NewRef(2)
				computed1 := NewComputed(func() int { return ref1.GetTyped() * 2 })
				computed2 := NewComputed(func() int { return ref2.GetTyped() * 3 })

				deps := []Dependency{ref1, ref2, computed1, computed2}

				assert.Len(t, deps, 4)
				assert.Equal(t, 1, deps[0].Get())
				assert.Equal(t, 2, deps[1].Get())
				assert.Equal(t, 2, deps[2].Get())
				assert.Equal(t, 6, deps[3].Get())
			},
		},
		{
			name: "Get() any recomputes when dependencies change",
			test: func(t *testing.T) {
				count := NewRef(5)
				computed := NewComputed(func() int {
					return count.GetTyped() * 2
				})

				// First call
				result := computed.Get()
				assert.Equal(t, 10, result)

				// Change dependency
				count.Set(10)

				// Should recompute
				result = computed.Get()
				assert.Equal(t, 20, result)
			},
		},
		{
			name: "GetTyped() T recomputes when dependencies change",
			test: func(t *testing.T) {
				count := NewRef(5)
				computed := NewComputed(func() int {
					return count.GetTyped() * 2
				})

				// First call
				result := computed.GetTyped()
				assert.Equal(t, 10, result)

				// Change dependency
				count.Set(10)

				// Should recompute
				result = computed.GetTyped()
				assert.Equal(t, 20, result)
			},
		},
		{
			name: "Computed with different types",
			test: func(t *testing.T) {
				firstName := NewRef("John")
				lastName := NewRef("Doe")

				fullName := NewComputed(func() string {
					return firstName.GetTyped() + " " + lastName.GetTyped()
				})

				value := fullName.Get()
				assert.Equal(t, "John Doe", value)

				typedValue := fullName.GetTyped()
				assert.Equal(t, "John Doe", typedValue)
			},
		},
		{
			name: "Computed can depend on other Computed",
			test: func(t *testing.T) {
				base := NewRef(5)
				doubled := NewComputed(func() int {
					return base.GetTyped() * 2
				})
				quadrupled := NewComputed(func() int {
					return doubled.GetTyped() * 2
				})

				result := quadrupled.Get()
				assert.Equal(t, 20, result)

				base.Set(10)
				result = quadrupled.Get()
				assert.Equal(t, 40, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

// TestComputed_DependencyInterfaceMethods verifies Dependency interface methods work correctly
func TestComputed_DependencyInterfaceMethods(t *testing.T) {
	t.Run("Invalidate propagates to dependents", func(t *testing.T) {
		count := NewRef(5)
		computed := NewComputed(func() int {
			return count.GetTyped() * 2
		})
		mockDep := &refTestDependency{value: 0}

		computed.AddDependent(mockDep)
		computed.Invalidate()

		assert.True(t, mockDep.invalidated, "dependent should be invalidated")
	})

	t.Run("AddDependent registers dependency", func(t *testing.T) {
		count := NewRef(5)
		computed := NewComputed(func() int {
			return count.GetTyped() * 2
		})
		mockDep1 := &refTestDependency{value: 1}
		mockDep2 := &refTestDependency{value: 2}

		computed.AddDependent(mockDep1)
		computed.AddDependent(mockDep2)

		// Invalidate should propagate to both
		computed.Invalidate()

		assert.True(t, mockDep1.invalidated)
		assert.True(t, mockDep2.invalidated)
	})
}
