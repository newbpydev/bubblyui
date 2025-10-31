package bubbly

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRef_ImplementsDependency verifies Ref implements the Dependency interface
func TestRef_ImplementsDependency(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Ref implements Dependency interface",
			test: func(t *testing.T) {
				ref := NewRef(42)
				var _ Dependency = ref
			},
		},
		{
			name: "Get() any returns value",
			test: func(t *testing.T) {
				ref := NewRef(42)
				value := ref.Get()
				assert.Equal(t, 42, value)
			},
		},
		{
			name: "Get() any works with different types",
			test: func(t *testing.T) {
				intRef := NewRef(42)
				stringRef := NewRef("hello")
				boolRef := NewRef(true)
				
				assert.Equal(t, 42, intRef.Get())
				assert.Equal(t, "hello", stringRef.Get())
				assert.Equal(t, true, boolRef.Get())
			},
		},
		{
			name: "Get() any can be type asserted",
			test: func(t *testing.T) {
				ref := NewRef(42)
				value := ref.Get()
				
				// Type assertion should work
				intValue, ok := value.(int)
				assert.True(t, ok, "should be able to type assert to int")
				assert.Equal(t, 42, intValue)
			},
		},
		{
			name: "GetTyped() T returns typed value",
			test: func(t *testing.T) {
				ref := NewRef(42)
				value := ref.GetTyped()
				
				// Should be int, not any
				assert.Equal(t, 42, value)
				
				// Verify it's actually type int at compile time
				var _ int = value
			},
		},
		{
			name: "GetTyped() T preserves type safety",
			test: func(t *testing.T) {
				stringRef := NewRef("hello")
				value := stringRef.GetTyped()
				
				// Should be string, not any
				assert.Equal(t, "hello", value)
				
				// Verify it's actually type string at compile time
				var _ string = value
			},
		},
		{
			name: "Get() and GetTyped() return same value",
			test: func(t *testing.T) {
				ref := NewRef(42)
				
				anyValue := ref.Get()
				typedValue := ref.GetTyped()
				
				assert.Equal(t, typedValue, anyValue.(int))
			},
		},
		{
			name: "Ref can be used as Dependency in slice",
			test: func(t *testing.T) {
				ref1 := NewRef(1)
				ref2 := NewRef(2)
				ref3 := NewRef(3)
				
				deps := []Dependency{ref1, ref2, ref3}
				
				assert.Len(t, deps, 3)
				assert.Equal(t, 1, deps[0].Get())
				assert.Equal(t, 2, deps[1].Get())
				assert.Equal(t, 3, deps[2].Get())
			},
		},
		{
			name: "Get() any tracks dependencies",
			test: func(t *testing.T) {
				ref := NewRef(42)
				
				// Create a computed that depends on ref
				computed := NewComputed(func() int {
					return ref.Get().(int) * 2
				})
				
				// First call should track dependency
				result := computed.Get()
				assert.Equal(t, 84, result)
				
				// Change ref value
				ref.Set(10)
				
				// Computed should recompute
				result = computed.Get()
				assert.Equal(t, 20, result)
			},
		},
		{
			name: "GetTyped() T tracks dependencies",
			test: func(t *testing.T) {
				ref := NewRef(42)
				
				// Create a computed that depends on ref using GetTyped
				computed := NewComputed(func() int {
					return ref.GetTyped() * 2
				})
				
				// First call should track dependency
				result := computed.Get()
				assert.Equal(t, 84, result)
				
				// Change ref value
				ref.Set(10)
				
				// Computed should recompute
				result = computed.Get()
				assert.Equal(t, 20, result)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

// TestRef_DependencyInterfaceMethods verifies Dependency interface methods work correctly
func TestRef_DependencyInterfaceMethods(t *testing.T) {
	t.Run("Invalidate propagates to dependents", func(t *testing.T) {
		ref := NewRef(42)
		mockDep := &refTestDependency{value: 0}
		
		ref.AddDependent(mockDep)
		ref.Invalidate()
		
		assert.True(t, mockDep.invalidated, "dependent should be invalidated")
	})
	
	t.Run("AddDependent registers dependency", func(t *testing.T) {
		ref := NewRef(42)
		mockDep1 := &refTestDependency{value: 1}
		mockDep2 := &refTestDependency{value: 2}
		
		ref.AddDependent(mockDep1)
		ref.AddDependent(mockDep2)
		
		// Invalidate should propagate to both
		ref.Invalidate()
		
		assert.True(t, mockDep1.invalidated)
		assert.True(t, mockDep2.invalidated)
	})
}

// refTestDependency is a simple test implementation of Dependency for ref tests
type refTestDependency struct {
	value       any
	invalidated bool
	dependents  []Dependency
}

func (t *refTestDependency) Get() any {
	return t.value
}

func (t *refTestDependency) Invalidate() {
	t.invalidated = true
	for _, dep := range t.dependents {
		dep.Invalidate()
	}
}

func (t *refTestDependency) AddDependent(dep Dependency) {
	t.dependents = append(t.dependents, dep)
}
