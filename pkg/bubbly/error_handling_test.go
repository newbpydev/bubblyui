package bubbly

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCircularReferenceDetection tests that circular component references are detected
func TestCircularReferenceDetection(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (Component, Component, error)
		wantError bool
		errorType error
	}{
		{
			name: "direct circular reference (A -> B -> A)",
			setup: func() (Component, Component, error) {
				compA, _ := NewComponent("A").
					Template(func(ctx RenderContext) string { return "A" }).
					Build()
				compB, _ := NewComponent("B").
					Template(func(ctx RenderContext) string { return "B" }).
					Build()

				// Add B as child of A
				err := compA.(*componentImpl).AddChild(compB)
				if err != nil {
					return nil, nil, err
				}

				// Try to add A as child of B (circular)
				err = compB.(*componentImpl).AddChild(compA)
				return compA, compB, err
			},
			wantError: true,
			errorType: ErrCircularRef,
		},
		{
			name: "indirect circular reference (A -> B -> C -> A)",
			setup: func() (Component, Component, error) {
				compA, _ := NewComponent("A").
					Template(func(ctx RenderContext) string { return "A" }).
					Build()
				compB, _ := NewComponent("B").
					Template(func(ctx RenderContext) string { return "B" }).
					Build()
				compC, _ := NewComponent("C").
					Template(func(ctx RenderContext) string { return "C" }).
					Build()

				// A -> B
				compA.(*componentImpl).AddChild(compB)
				// B -> C
				compB.(*componentImpl).AddChild(compC)
				// Try C -> A (circular)
				err := compC.(*componentImpl).AddChild(compA)
				return compA, compC, err
			},
			wantError: true,
			errorType: ErrCircularRef,
		},
		{
			name: "self-reference (A -> A)",
			setup: func() (Component, Component, error) {
				compA, _ := NewComponent("A").
					Template(func(ctx RenderContext) string { return "A" }).
					Build()

				// Try to add A as child of itself
				err := compA.(*componentImpl).AddChild(compA)
				return compA, compA, err
			},
			wantError: true,
			errorType: ErrCircularRef,
		},
		{
			name: "no circular reference (A -> B, C -> D)",
			setup: func() (Component, Component, error) {
				compA, _ := NewComponent("A").
					Template(func(ctx RenderContext) string { return "A" }).
					Build()
				compB, _ := NewComponent("B").
					Template(func(ctx RenderContext) string { return "B" }).
					Build()

				err := compA.(*componentImpl).AddChild(compB)
				return compA, compB, err
			},
			wantError: false,
			errorType: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := tt.setup()

			if tt.wantError {
				assert.Error(t, err, "Expected error for circular reference")
				assert.ErrorIs(t, err, tt.errorType, "Error should be ErrCircularRef")
			} else {
				assert.NoError(t, err, "Should not error for valid component tree")
			}
		})
	}
}

// TestMaxDepthEnforcement tests that maximum component depth is enforced
func TestMaxDepthEnforcement(t *testing.T) {
	tests := []struct {
		name      string
		depth     int
		wantError bool
	}{
		{
			name:      "depth within limit (5 levels)",
			depth:     5,
			wantError: false,
		},
		{
			name:      "depth at limit (51 components = depth 50)",
			depth:     MaxComponentDepth + 1,
			wantError: false,
		},
		{
			name:      "depth exceeds limit (52 components = depth 51)",
			depth:     MaxComponentDepth + 2,
			wantError: true,
		},
		{
			name:      "depth far exceeds limit (100 levels)",
			depth:     100,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a chain of components
			var root Component
			var current Component
			var err error

			for i := 0; i < tt.depth; i++ {
				comp, buildErr := NewComponent("Component").
					Template(func(ctx RenderContext) string { return "test" }).
					Build()
				assert.NoError(t, buildErr)

				if i == 0 {
					root = comp
					current = comp
				} else {
					err = current.(*componentImpl).AddChild(comp)
					if err != nil {
						break
					}
					current = comp
				}
			}

			if tt.wantError {
				assert.Error(t, err, "Expected error for exceeding max depth")
				assert.ErrorIs(t, err, ErrMaxDepth, "Error should be ErrMaxDepth")
			} else {
				assert.NoError(t, err, "Should not error within depth limit")
				assert.NotNil(t, root, "Root component should be created")
			}
		})
	}
}

// TestEventHandlerPanicRecovery tests that panics in event handlers are recovered
func TestEventHandlerPanicRecovery(t *testing.T) {
	tests := []struct {
		name          string
		handler       EventHandler
		shouldPanic   bool
		checkRecovery bool
	}{
		{
			name: "handler panics with string",
			handler: func(data interface{}) {
				panic("test panic")
			},
			shouldPanic:   true,
			checkRecovery: true,
		},
		{
			name: "handler panics with error",
			handler: func(data interface{}) {
				panic(errors.New("test error"))
			},
			shouldPanic:   true,
			checkRecovery: true,
		},
		{
			name: "handler panics with nil",
			handler: func(data interface{}) {
				panic(nil)
			},
			shouldPanic:   true,
			checkRecovery: true,
		},
		{
			name: "handler executes normally",
			handler: func(data interface{}) {
				// Normal execution
			},
			shouldPanic:   false,
			checkRecovery: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp, err := NewComponent("TestComponent").
				Setup(func(ctx *Context) {
					ctx.On("test", tt.handler)
				}).
				Template(func(ctx RenderContext) string {
					return "test"
				}).
				Build()

			assert.NoError(t, err)

			// Initialize component
			comp.Init()

			// Emit event - should not panic even if handler panics
			assert.NotPanics(t, func() {
				comp.Emit("test", "data")
			}, "Emit should not panic even if handler panics")
		})
	}
}

// TestComponentErrorMessages tests that error messages are clear and descriptive
func TestComponentErrorMessages(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantContains []string
	}{
		{
			name:         "ErrMissingTemplate",
			err:          ErrMissingTemplate,
			wantContains: []string{"template", "required"},
		},
		{
			name:         "ErrInvalidProps",
			err:          ErrInvalidProps,
			wantContains: []string{"props", "validation", "failed"},
		},
		{
			name:         "ErrCircularRef",
			err:          ErrCircularRef,
			wantContains: []string{"circular", "reference"},
		},
		{
			name:         "ErrMaxDepth",
			err:          ErrMaxDepth,
			wantContains: []string{"max", "depth"},
		},
		{
			name:         "ErrNilChild",
			err:          ErrNilChild,
			wantContains: []string{"nil", "child"},
		},
		{
			name:         "ErrChildNotFound",
			err:          ErrChildNotFound,
			wantContains: []string{"child", "not found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()
			assert.NotEmpty(t, errMsg, "Error message should not be empty")

			for _, want := range tt.wantContains {
				assert.Contains(t, strings.ToLower(errMsg), strings.ToLower(want),
					"Error message should contain '%s'", want)
			}
		})
	}
}

// TestComponentDepthCalculation tests the depth calculation helper
func TestComponentDepthCalculation(t *testing.T) {
	tests := []struct {
		name          string
		buildTree     func() Component
		expectedDepth int
	}{
		{
			name: "single component (depth 0)",
			buildTree: func() Component {
				comp, _ := NewComponent("Root").
					Template(func(ctx RenderContext) string { return "root" }).
					Build()
				return comp
			},
			expectedDepth: 0,
		},
		{
			name: "parent with one child (depth 1)",
			buildTree: func() Component {
				child, _ := NewComponent("Child").
					Template(func(ctx RenderContext) string { return "child" }).
					Build()
				parent, _ := NewComponent("Parent").
					Children(child).
					Template(func(ctx RenderContext) string { return "parent" }).
					Build()
				return parent
			},
			expectedDepth: 1,
		},
		{
			name: "three-level tree (depth 2)",
			buildTree: func() Component {
				grandchild, _ := NewComponent("Grandchild").
					Template(func(ctx RenderContext) string { return "grandchild" }).
					Build()
				child, _ := NewComponent("Child").
					Children(grandchild).
					Template(func(ctx RenderContext) string { return "child" }).
					Build()
				parent, _ := NewComponent("Parent").
					Children(child).
					Template(func(ctx RenderContext) string { return "parent" }).
					Build()
				return parent
			},
			expectedDepth: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := tt.buildTree()
			depth := calculateComponentDepth(root.(*componentImpl))
			assert.Equal(t, tt.expectedDepth, depth, "Depth should match expected value")
		})
	}
}

// TestCircularReferenceErrorDetails tests that circular reference errors include helpful details
func TestCircularReferenceErrorDetails(t *testing.T) {
	compA, _ := NewComponent("ComponentA").
		Template(func(ctx RenderContext) string { return "A" }).
		Build()
	compB, _ := NewComponent("ComponentB").
		Template(func(ctx RenderContext) string { return "B" }).
		Build()

	// A -> B
	compA.(*componentImpl).AddChild(compB)

	// Try B -> A (circular)
	err := compB.(*componentImpl).AddChild(compA)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCircularRef)

	// Error message should include component names for debugging
	errMsg := err.Error()
	assert.Contains(t, errMsg, "circular", "Error should mention circular reference")
}

// TestMaxDepthErrorDetails tests that max depth errors include helpful details
func TestMaxDepthErrorDetails(t *testing.T) {
	// Create a chain exceeding max depth
	var current Component

	for i := 0; i <= MaxComponentDepth+1; i++ {
		comp, _ := NewComponent("Component").
			Template(func(ctx RenderContext) string { return "test" }).
			Build()

		if i == 0 {
			current = comp
		} else {
			err := current.(*componentImpl).AddChild(comp)
			if err != nil {
				assert.ErrorIs(t, err, ErrMaxDepth)
				errMsg := err.Error()
				assert.Contains(t, errMsg, "depth", "Error should mention depth")
				return
			}
			current = comp
		}
	}

	t.Fatal("Expected max depth error but none occurred")
}
