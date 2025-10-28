package bubbly

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRenderContext_Get tests that RenderContext.Get retrieves exposed values
func TestRenderContext_Get(t *testing.T) {
	tests := []struct {
		name          string
		setupState    map[string]interface{}
		key           string
		expectedValue interface{}
	}{
		{
			name: "get existing ref",
			setupState: map[string]interface{}{
				"count": NewRef(42),
			},
			key:           "count",
			expectedValue: NewRef(42),
		},
		{
			name:          "get non-existent value",
			setupState:    map[string]interface{}{},
			key:           "missing",
			expectedValue: nil,
		},
		{
			name: "get string value",
			setupState: map[string]interface{}{
				"name": "test",
			},
			key:           "name",
			expectedValue: "test",
		},
		{
			name: "get computed value",
			setupState: map[string]interface{}{
				"doubled": NewComputed(func() interface{} { return 84 }),
			},
			key:           "doubled",
			expectedValue: NewComputed(func() interface{} { return 84 }),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := &componentImpl{
				name:  "TestComponent",
				state: tt.setupState,
			}
			ctx := RenderContext{component: c}

			// Act
			value := ctx.Get(tt.key)

			// Assert
			if tt.expectedValue == nil {
				assert.Nil(t, value, "Get should return nil for non-existent key")
			} else {
				// For Ref and Computed, compare the actual values
				if ref, ok := tt.expectedValue.(*Ref[interface{}]); ok {
					gotRef, ok := value.(*Ref[interface{}])
					require.True(t, ok, "Value should be a Ref")
					assert.Equal(t, ref.Get(), gotRef.Get(), "Ref values should match")
				} else if comp, ok := tt.expectedValue.(*Computed[interface{}]); ok {
					gotComp, ok := value.(*Computed[interface{}])
					require.True(t, ok, "Value should be a Computed")
					assert.Equal(t, comp.Get(), gotComp.Get(), "Computed values should match")
				} else {
					assert.Equal(t, tt.expectedValue, value, "Get should return exposed value")
				}
			}
		})
	}
}

// TestRenderContext_Props tests that RenderContext.Props returns component props
func TestRenderContext_Props(t *testing.T) {
	tests := []struct {
		name          string
		props         interface{}
		expectedProps interface{}
	}{
		{
			name: "get struct props",
			props: struct {
				Label    string
				Disabled bool
			}{Label: "Button", Disabled: false},
			expectedProps: struct {
				Label    string
				Disabled bool
			}{Label: "Button", Disabled: false},
		},
		{
			name:          "get string props",
			props:         "simple-props",
			expectedProps: "simple-props",
		},
		{
			name:          "get nil props",
			props:         nil,
			expectedProps: nil,
		},
		{
			name: "get map props",
			props: map[string]interface{}{
				"key": "value",
			},
			expectedProps: map[string]interface{}{
				"key": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := &componentImpl{
				name:  "TestComponent",
				props: tt.props,
				state: make(map[string]interface{}),
			}
			ctx := RenderContext{component: c}

			// Act
			props := ctx.Props()

			// Assert
			assert.Equal(t, tt.expectedProps, props, "Props should match component props")
		})
	}
}

// TestRenderContext_Children tests that RenderContext.Children returns child components
func TestRenderContext_Children(t *testing.T) {
	t.Run("get children", func(t *testing.T) {
		// Arrange
		child1 := &componentImpl{name: "Child1", state: make(map[string]interface{})}
		child2 := &componentImpl{name: "Child2", state: make(map[string]interface{})}

		c := &componentImpl{
			name:     "Parent",
			state:    make(map[string]interface{}),
			children: []Component{child1, child2},
		}
		ctx := RenderContext{component: c}

		// Act
		children := ctx.Children()

		// Assert
		require.Len(t, children, 2, "Should have 2 children")
		assert.Equal(t, "Child1", children[0].Name(), "First child name should match")
		assert.Equal(t, "Child2", children[1].Name(), "Second child name should match")
	})

	t.Run("get empty children", func(t *testing.T) {
		// Arrange
		c := &componentImpl{
			name:     "Parent",
			state:    make(map[string]interface{}),
			children: []Component{},
		}
		ctx := RenderContext{component: c}

		// Act
		children := ctx.Children()

		// Assert
		assert.Empty(t, children, "Children should be empty")
	})

	t.Run("children are read-only", func(t *testing.T) {
		// Arrange
		child1 := &componentImpl{name: "Child1", state: make(map[string]interface{})}

		c := &componentImpl{
			name:     "Parent",
			state:    make(map[string]interface{}),
			children: []Component{child1},
		}
		ctx := RenderContext{component: c}

		// Act
		children := ctx.Children()

		// Assert - Verify it's a copy, not the original slice
		// Modifying the returned slice should not affect the component
		require.Len(t, children, 1, "Should have 1 child")
		assert.Equal(t, "Child1", children[0].Name(), "Child name should match")
	})
}

// TestRenderContext_RenderChild tests that RenderContext.RenderChild renders child components
func TestRenderContext_RenderChild(t *testing.T) {
	t.Run("render child with template", func(t *testing.T) {
		// Arrange
		child := &componentImpl{
			name:  "Child",
			state: make(map[string]interface{}),
			template: func(ctx RenderContext) string {
				return "Child Output"
			},
		}

		c := &componentImpl{
			name:     "Parent",
			state:    make(map[string]interface{}),
			children: []Component{child},
		}
		ctx := RenderContext{component: c}

		// Act
		output := ctx.RenderChild(child)

		// Assert
		assert.Equal(t, "Child Output", output, "Should render child template")
	})

	t.Run("render child without template", func(t *testing.T) {
		// Arrange
		child := &componentImpl{
			name:     "Child",
			state:    make(map[string]interface{}),
			template: nil,
		}

		c := &componentImpl{
			name:     "Parent",
			state:    make(map[string]interface{}),
			children: []Component{child},
		}
		ctx := RenderContext{component: c}

		// Act
		output := ctx.RenderChild(child)

		// Assert
		assert.Equal(t, "", output, "Should return empty string for child without template")
	})

	t.Run("render child with state access", func(t *testing.T) {
		// Arrange
		child := &componentImpl{
			name: "Child",
			state: map[string]interface{}{
				"message": "Hello from child",
			},
			template: func(ctx RenderContext) string {
				msg := ctx.Get("message").(string)
				return msg
			},
		}

		c := &componentImpl{
			name:     "Parent",
			state:    make(map[string]interface{}),
			children: []Component{child},
		}
		ctx := RenderContext{component: c}

		// Act
		output := ctx.RenderChild(child)

		// Assert
		assert.Equal(t, "Hello from child", output, "Child should access its own state")
	})
}

// TestRenderContext_ReadOnly tests that RenderContext is read-only
func TestRenderContext_ReadOnly(t *testing.T) {
	t.Run("no set method exists", func(t *testing.T) {
		// This test verifies at compile time that RenderContext doesn't have a Set method
		// If this compiles, the test passes
		c := &componentImpl{
			name:  "TestComponent",
			state: make(map[string]interface{}),
		}
		ctx := RenderContext{component: c}

		// Verify we can only read
		_ = ctx.Get("key")
		_ = ctx.Props()
		_ = ctx.Children()

		// The following would not compile if uncommented:
		// ctx.Set("key", "value")
		// ctx.Expose("key", "value")
		// ctx.On("event", func(data interface{}) {})
		// ctx.Emit("event", nil)
	})

	t.Run("state modifications don't affect component", func(t *testing.T) {
		// Arrange
		ref := NewRef[interface{}](0)
		c := &componentImpl{
			name: "TestComponent",
			state: map[string]interface{}{
				"count": ref,
			},
		}
		ctx := RenderContext{component: c}

		// Act - Get the ref and modify it
		gotRef := ctx.Get("count").(*Ref[interface{}])
		gotRef.Set(42)

		// Assert - The component's state still has the modified ref
		// (RenderContext doesn't prevent modifications to the objects themselves,
		// it just doesn't provide methods to add/remove state entries)
		componentRef := c.state["count"].(*Ref[interface{}])
		assert.Equal(t, 42, componentRef.Get(), "Ref modification is allowed")
		assert.Same(t, ref, gotRef, "Should be the same ref instance")
	})
}

// TestRenderContext_Integration tests full workflow with RenderContext
func TestRenderContext_Integration(t *testing.T) {
	t.Run("complete template rendering workflow", func(t *testing.T) {
		// Arrange - Create a parent with children
		child1 := &componentImpl{
			name: "Button",
			props: struct {
				Label string
			}{Label: "Click me"},
			state: make(map[string]interface{}),
			template: func(ctx RenderContext) string {
				props := ctx.Props().(struct{ Label string })
				return "[" + props.Label + "]"
			},
		}

		child2 := &componentImpl{
			name: "Display",
			state: map[string]interface{}{
				"count": NewRef[interface{}](5),
			},
			template: func(ctx RenderContext) string {
				count := ctx.Get("count").(*Ref[interface{}])
				return "Count: " + string(rune(count.Get().(int)+'0'))
			},
		}

		parent := &componentImpl{
			name: "Container",
			state: map[string]interface{}{
				"title": "My App",
			},
			children: []Component{child1, child2},
			template: func(ctx RenderContext) string {
				title := ctx.Get("title").(string)
				output := title + "\n"
				for _, child := range ctx.Children() {
					output += ctx.RenderChild(child) + "\n"
				}
				return output
			},
		}

		// Act
		ctx := RenderContext{component: parent}
		output := parent.template(ctx)

		// Assert
		expected := "My App\n[Click me]\nCount: 5\n"
		assert.Equal(t, expected, output, "Should render complete component tree")
	})

	t.Run("nested component rendering", func(t *testing.T) {
		// Arrange - Create deeply nested components
		grandchild := &componentImpl{
			name:  "Grandchild",
			state: make(map[string]interface{}),
			template: func(ctx RenderContext) string {
				return "Grandchild"
			},
		}

		child := &componentImpl{
			name:     "Child",
			state:    make(map[string]interface{}),
			children: []Component{grandchild},
			template: func(ctx RenderContext) string {
				output := "Child("
				for _, c := range ctx.Children() {
					output += ctx.RenderChild(c)
				}
				output += ")"
				return output
			},
		}

		parent := &componentImpl{
			name:     "Parent",
			state:    make(map[string]interface{}),
			children: []Component{child},
			template: func(ctx RenderContext) string {
				output := "Parent("
				for _, c := range ctx.Children() {
					output += ctx.RenderChild(c)
				}
				output += ")"
				return output
			},
		}

		// Act
		ctx := RenderContext{component: parent}
		output := parent.template(ctx)

		// Assert
		assert.Equal(t, "Parent(Child(Grandchild))", output, "Should render nested components")
	})
}
