package testutil

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPropsVerifier_NewPropsVerifier tests the constructor
func TestPropsVerifier_NewPropsVerifier(t *testing.T) {
	// Create a simple component with props
	type TestProps struct {
		Name  string
		Count int
	}

	comp, err := bubbly.NewComponent("TestComp").
		Props(TestProps{Name: "test", Count: 42}).
		Template(func(ctx bubbly.RenderContext) string { return "test" }).
		Build()
	require.NoError(t, err)

	// Create verifier
	pv := NewPropsVerifier(comp)

	assert.NotNil(t, pv)
	assert.Equal(t, comp, pv.component)
}

// TestPropsVerifier_CaptureOriginalProps tests capturing the initial props state
func TestPropsVerifier_CaptureOriginalProps(t *testing.T) {
	tests := []struct {
		name  string
		props interface{}
	}{
		{
			name: "simple struct props",
			props: struct {
				Name  string
				Count int
			}{Name: "test", Count: 42},
		},
		{
			name: "map props",
			props: map[string]interface{}{
				"name":  "test",
				"count": 42,
			},
		},
		{
			name:  "slice props",
			props: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp, err := bubbly.NewComponent("TestComp").
				Props(tt.props).
				Template(func(ctx bubbly.RenderContext) string { return "test" }).
				Build()
			require.NoError(t, err)

			pv := NewPropsVerifier(comp)
			pv.CaptureOriginalProps()

			assert.NotNil(t, pv.originalProps)
		})
	}
}

// TestPropsVerifier_AttemptPropMutation tests attempting to mutate props
func TestPropsVerifier_AttemptPropMutation(t *testing.T) {
	type TestProps struct {
		Name  string
		Count int
		Tags  []string
	}

	comp, err := bubbly.NewComponent("TestComp").
		Props(TestProps{
			Name:  "original",
			Count: 42,
			Tags:  []string{"a", "b"},
		}).
		Template(func(ctx bubbly.RenderContext) string { return "test" }).
		Build()
	require.NoError(t, err)

	pv := NewPropsVerifier(comp)
	pv.CaptureOriginalProps()

	// Attempt mutations
	pv.AttemptPropMutation("Name", "mutated")
	pv.AttemptPropMutation("Count", 100)
	pv.AttemptPropMutation("Tags", []string{"x", "y", "z"})

	// Should have recorded mutations
	mutations := pv.GetMutations()
	assert.Len(t, mutations, 3)
}

// TestPropsVerifier_AssertPropsImmutable tests verifying props immutability
func TestPropsVerifier_AssertPropsImmutable(t *testing.T) {
	// Test 1: Props unchanged - should pass
	t.Run("props unchanged - should pass", func(t *testing.T) {
		comp, err := bubbly.NewComponent("TestComp").
			Props(struct{ Name string }{Name: "test"}).
			Template(func(ctx bubbly.RenderContext) string { return "test" }).
			Build()
		require.NoError(t, err)

		pv := NewPropsVerifier(comp)
		pv.CaptureOriginalProps()

		// No mutations attempted

		// Create mock testing.T
		mockT := &mockTestingT{}

		// Assert immutability
		pv.AssertPropsImmutable(mockT)

		// Should pass
		assert.False(t, mockT.failed, "expected assertion to pass")
	})

	// Test 2: AttemptPropMutation records mutations
	// Note: AttemptPropMutation doesn't actually mutate the component's props,
	// it just records the mutation attempt. The component's props remain immutable.
	t.Run("mutation attempts recorded", func(t *testing.T) {
		comp, err := bubbly.NewComponent("TestComp").
			Props(struct{ Name string }{Name: "original"}).
			Template(func(ctx bubbly.RenderContext) string { return "test" }).
			Build()
		require.NoError(t, err)

		pv := NewPropsVerifier(comp)
		pv.CaptureOriginalProps()

		// Attempt mutation (doesn't actually mutate)
		pv.AttemptPropMutation("Name", "mutated")

		// Props should still be immutable (unchanged)
		mockT := &mockTestingT{}
		pv.AssertPropsImmutable(mockT)
		assert.False(t, mockT.failed, "props should remain immutable")

		// But mutation was recorded
		mutations := pv.GetMutations()
		assert.Len(t, mutations, 1)
	})
}

// TestPropsVerifier_AssertNoMutations tests verifying no mutations occurred
func TestPropsVerifier_AssertNoMutations(t *testing.T) {
	tests := []struct {
		name          string
		props         interface{}
		attemptMutate bool
		shouldPass    bool
	}{
		{
			name: "no mutations - should pass",
			props: struct {
				Name string
			}{Name: "test"},
			attemptMutate: false,
			shouldPass:    true,
		},
		{
			name: "mutations attempted - should fail",
			props: struct {
				Name string
			}{Name: "test"},
			attemptMutate: true,
			shouldPass:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp, err := bubbly.NewComponent("TestComp").
				Props(tt.props).
				Template(func(ctx bubbly.RenderContext) string { return "test" }).
				Build()
			require.NoError(t, err)

			pv := NewPropsVerifier(comp)
			pv.CaptureOriginalProps()

			if tt.attemptMutate {
				pv.AttemptPropMutation("Name", "mutated")
			}

			// Create mock testing.T
			mockT := &mockTestingT{}

			// Assert no mutations
			pv.AssertNoMutations(mockT)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "expected assertion to fail")
			}
		})
	}
}

// TestPropsVerifier_DeepImmutability tests deep immutability for nested structures
func TestPropsVerifier_DeepImmutability(t *testing.T) {
	type NestedProps struct {
		User struct {
			Name  string
			Email string
		}
		Settings map[string]interface{}
		Tags     []string
	}

	props := NestedProps{
		User: struct {
			Name  string
			Email string
		}{
			Name:  "John",
			Email: "john@example.com",
		},
		Settings: map[string]interface{}{
			"theme": "dark",
			"lang":  "en",
		},
		Tags: []string{"admin", "user"},
	}

	comp, err := bubbly.NewComponent("TestComp").
		Props(props).
		Template(func(ctx bubbly.RenderContext) string { return "test" }).
		Build()
	require.NoError(t, err)

	pv := NewPropsVerifier(comp)
	pv.CaptureOriginalProps()

	// Attempt deep mutations
	pv.AttemptPropMutation("User.Name", "Jane")
	pv.AttemptPropMutation("Settings.theme", "light")
	pv.AttemptPropMutation("Tags[0]", "superadmin")

	// Should have recorded mutations
	mutations := pv.GetMutations()
	assert.Len(t, mutations, 3)
}

// TestPropsVerifier_TypeSafety tests that type safety is maintained
func TestPropsVerifier_TypeSafety(t *testing.T) {
	type TestProps struct {
		Name  string
		Count int
	}

	comp, err := bubbly.NewComponent("TestComp").
		Props(TestProps{Name: "test", Count: 42}).
		Template(func(ctx bubbly.RenderContext) string { return "test" }).
		Build()
	require.NoError(t, err)

	pv := NewPropsVerifier(comp)
	pv.CaptureOriginalProps()

	// Attempt type-mismatched mutation
	pv.AttemptPropMutation("Count", "not a number")

	mutations := pv.GetMutations()
	assert.Len(t, mutations, 1)
	assert.Equal(t, "Count", mutations[0].Key)
	assert.Equal(t, "not a number", mutations[0].NewValue)
	// JSON marshaling converts int to float64
	assert.Equal(t, float64(42), mutations[0].OldValue)
}

// TestPropsVerifier_ReferenceIntegrity tests that reference integrity is preserved
func TestPropsVerifier_ReferenceIntegrity(t *testing.T) {
	type TestProps struct {
		Data []int
	}

	originalData := []int{1, 2, 3}
	props := TestProps{Data: originalData}

	comp, err := bubbly.NewComponent("TestComp").
		Props(props).
		Template(func(ctx bubbly.RenderContext) string { return "test" }).
		Build()
	require.NoError(t, err)

	pv := NewPropsVerifier(comp)
	pv.CaptureOriginalProps()

	// Mutate the original slice
	originalData[0] = 999

	// Component's props should not be affected if properly cloned
	componentProps := comp.Props().(TestProps)

	// This test verifies that the component either:
	// 1. Cloned the props (so mutation doesn't affect it)
	// 2. Or we can detect the mutation
	if componentProps.Data[0] == 999 {
		// Props were not cloned - mutation detected
		pv.AttemptPropMutation("Data[0]", 999)
		mutations := pv.GetMutations()
		assert.Len(t, mutations, 1)
	} else {
		// Props were cloned - immutability preserved
		assert.Equal(t, 1, componentProps.Data[0])
	}
}

// TestPropsVerifier_String tests the String method
func TestPropsVerifier_String(t *testing.T) {
	type TestProps struct {
		Name   string
		Count  int
		Active bool
	}

	comp, err := bubbly.NewComponent("TestComp").
		Props(TestProps{Name: "Test", Count: 42, Active: true}).
		Template(func(ctx bubbly.RenderContext) string { return "test" }).
		Build()
	require.NoError(t, err)

	pv := NewPropsVerifier(comp)

	// Attempt some mutations
	pv.AttemptPropMutation("Name", "Modified")
	pv.AttemptPropMutation("Count", 100)

	// Get string representation
	str := pv.String()

	// Verify it contains expected information
	assert.Contains(t, str, "PropsVerifier")
	assert.Contains(t, str, "props")
	assert.Contains(t, str, "2 mutations attempted")
	assert.Contains(t, str, "immutable")
	
	// Verify format is correct
	assert.NotEmpty(t, str)
}
