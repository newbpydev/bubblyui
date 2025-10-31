package bubbly

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDependencyInterface verifies the Dependency interface is properly defined
func TestDependencyInterface(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "interface has Get method",
			test: func(t *testing.T) {
				// This test verifies that the Dependency interface has a Get() any method
				// by attempting to use it with a test implementation
				var dep Dependency = &testDependency{value: 42}
				result := dep.Get()
				assert.Equal(t, 42, result)
			},
		},
		{
			name: "interface has Invalidate method",
			test: func(t *testing.T) {
				// Verify Invalidate method exists and can be called
				mock := &testDependency{value: 10}
				var dep Dependency = mock

				dep.Invalidate()
				assert.True(t, mock.invalidated, "Invalidate should mark dependency as invalidated")
			},
		},
		{
			name: "interface has AddDependent method",
			test: func(t *testing.T) {
				// Verify AddDependent method exists and can be called
				mock := &testDependency{value: 20}
				dependent := &testDependency{value: 30}
				var dep Dependency = mock

				dep.AddDependent(dependent)
				assert.Len(t, mock.dependents, 1, "AddDependent should add dependent to list")
				assert.Equal(t, dependent, mock.dependents[0])
			},
		},
		{
			name: "interface works with multiple implementations",
			test: func(t *testing.T) {
				// Verify interface can be implemented by different types
				deps := []Dependency{
					&testDependency{value: 1},
					&testDependency{value: 2},
					&testDependency{value: 3},
				}

				for i, dep := range deps {
					assert.Equal(t, i+1, dep.Get())
				}
			},
		},
		{
			name: "Get returns any type",
			test: func(t *testing.T) {
				// Verify Get() returns any and can hold different types
				stringDep := &testDependency{value: "hello"}
				intDep := &testDependency{value: 42}
				boolDep := &testDependency{value: true}

				assert.Equal(t, "hello", stringDep.Get())
				assert.Equal(t, 42, intDep.Get())
				assert.Equal(t, true, boolDep.Get())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

// TestDependencyInterfaceCompilation verifies the interface compiles correctly
func TestDependencyInterfaceCompilation(t *testing.T) {
	// This test ensures the interface definition compiles
	// and can be used in type assertions
	var _ Dependency = (*testDependency)(nil)

	// Verify we can create a slice of dependencies
	deps := make([]Dependency, 0)
	deps = append(deps, &testDependency{value: 1})
	deps = append(deps, &testDependency{value: 2})

	assert.Len(t, deps, 2)
}

// TestDependencyChaining verifies dependencies can form chains
func TestDependencyChaining(t *testing.T) {
	// Create a chain: dep1 -> dep2 -> dep3
	dep1 := &testDependency{value: 1}
	dep2 := &testDependency{value: 2}
	dep3 := &testDependency{value: 3}

	dep1.AddDependent(dep2)
	dep2.AddDependent(dep3)

	// Invalidate dep1, which should propagate
	dep1.Invalidate()

	assert.True(t, dep1.invalidated)
	// Note: Actual propagation logic is implementation-specific
	// This test just verifies the interface supports chaining
}

// testDependency is a test implementation of the Dependency interface
type testDependency struct {
	value       any
	invalidated bool
	dependents  []Dependency
}

func (m *testDependency) Get() any {
	return m.value
}

func (m *testDependency) Invalidate() {
	m.invalidated = true
	// Propagate to dependents
	for _, dep := range m.dependents {
		dep.Invalidate()
	}
}

func (m *testDependency) AddDependent(dep Dependency) {
	m.dependents = append(m.dependents, dep)
}
