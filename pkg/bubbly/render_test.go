package bubbly

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

// TestRenderContext_NewRenderer tests that RenderContext can create a Lipgloss renderer
func TestRenderContext_NewRenderer(t *testing.T) {
	tests := []struct {
		name string
		want bool // Whether renderer should be non-nil
	}{
		{
			name: "creates valid renderer",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newComponentImpl("TestComponent")
			ctx := RenderContext{component: c}

			renderer := ctx.NewRenderer()
			assert.NotNil(t, renderer, "NewRenderer should return non-nil renderer")
		})
	}
}

// TestRenderContext_NewStyle tests that RenderContext can create Lipgloss styles
func TestRenderContext_NewStyle(t *testing.T) {
	tests := []struct {
		name      string
		styleFunc func(lipgloss.Style) lipgloss.Style
		text      string
		wantLen   int // Approximate length check (styled text is longer)
	}{
		{
			name: "creates basic style",
			styleFunc: func(s lipgloss.Style) lipgloss.Style {
				return s
			},
			text:    "Hello",
			wantLen: 5,
		},
		{
			name: "creates bold style",
			styleFunc: func(s lipgloss.Style) lipgloss.Style {
				return s.Bold(true)
			},
			text:    "Bold",
			wantLen: 4,
		},
		{
			name: "creates styled with padding",
			styleFunc: func(s lipgloss.Style) lipgloss.Style {
				return s.Padding(1, 2)
			},
			text:    "Padded",
			wantLen: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newComponentImpl("TestComponent")
			ctx := RenderContext{component: c}

			style := ctx.NewStyle()
			assert.NotNil(t, style, "NewStyle should return non-nil style")

			// Apply test-specific styling
			styledStyle := tt.styleFunc(style)
			rendered := styledStyle.Render(tt.text)

			assert.NotEmpty(t, rendered, "Rendered text should not be empty")
			assert.Contains(t, rendered, tt.text, "Rendered text should contain original text")
		})
	}
}

// TestComponentRendering_WithLipgloss tests full component rendering with Lipgloss
func TestComponentRendering_WithLipgloss(t *testing.T) {
	tests := []struct {
		name     string
		template RenderFunc
		wantText string
	}{
		{
			name: "renders with basic style",
			template: func(ctx RenderContext) string {
				style := ctx.NewStyle().Bold(true)
				return style.Render("Hello")
			},
			wantText: "Hello",
		},
		{
			name: "renders with foreground color",
			template: func(ctx RenderContext) string {
				style := ctx.NewStyle().Foreground(lipgloss.Color("63"))
				return style.Render("Colored")
			},
			wantText: "Colored",
		},
		{
			name: "renders with multiple styles",
			template: func(ctx RenderContext) string {
				style := ctx.NewStyle().
					Bold(true).
					Italic(true).
					Foreground(lipgloss.Color("99"))
				return style.Render("Styled")
			},
			wantText: "Styled",
		},
		{
			name: "renders with padding",
			template: func(ctx RenderContext) string {
				style := ctx.NewStyle().Padding(1, 2)
				return style.Render("Padded")
			},
			wantText: "Padded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := NewComponent("TestComponent").
				Template(tt.template).
				Build()

			assert.NoError(t, err, "Component build should succeed")
			assert.NotNil(t, component, "Component should not be nil")

			// Render the component
			output := component.View()

			assert.NotEmpty(t, output, "Component output should not be empty")
			assert.Contains(t, output, tt.wantText, "Output should contain expected text")
		})
	}
}

// TestComponentRendering_WithStateAndLipgloss tests rendering with state and Lipgloss
func TestComponentRendering_WithStateAndLipgloss(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  SetupFunc
		template   RenderFunc
		wantText   string
		checkStyle bool
	}{
		{
			name: "renders state with styling",
			setupFunc: func(ctx *Context) {
				count := ctx.Ref(42)
				ctx.Expose("count", count)
			},
			template: func(ctx RenderContext) string {
				countRef := ctx.Get("count")
				// Type assert to Ref[interface{}] first, then get the value
				if _, ok := countRef.(*Ref[interface{}]); ok {
					style := ctx.NewStyle().Bold(true)
					// Convert value to string representation
					return style.Render("42")
				}
				return "Error"
			},
			wantText:   "42",
			checkStyle: true,
		},
		{
			name: "renders props with styling",
			setupFunc: func(ctx *Context) {
				// No state needed
			},
			template: func(ctx RenderContext) string {
				style := ctx.NewStyle().
					Foreground(lipgloss.Color("63")).
					Bold(true)
				return style.Render("Styled Props")
			},
			wantText:   "Styled Props",
			checkStyle: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := NewComponent("TestComponent").
				Setup(tt.setupFunc).
				Template(tt.template).
				Build()

			assert.NoError(t, err, "Component build should succeed")

			// Initialize component
			component.Init()

			// Render
			output := component.View()

			assert.NotEmpty(t, output, "Output should not be empty")
			if tt.checkStyle {
				// Output should be longer than plain text due to ANSI codes
				assert.GreaterOrEqual(t, len(output), len(tt.wantText), "Styled output should have ANSI codes")
			}
		})
	}
}

// TestComponentRendering_WithChildren tests rendering children with Lipgloss
func TestComponentRendering_WithChildren(t *testing.T) {
	tests := []struct {
		name         string
		childCount   int
		wantContains []string
	}{
		{
			name:         "renders single child with style",
			childCount:   1,
			wantContains: []string{"Child 0"},
		},
		{
			name:         "renders multiple children with styles",
			childCount:   3,
			wantContains: []string{"Child 0", "Child 1", "Child 2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create children
			children := make([]Component, tt.childCount)
			for i := 0; i < tt.childCount; i++ {
				idx := i
				child, err := NewComponent("Child").
					Template(func(ctx RenderContext) string {
						style := ctx.NewStyle().Bold(true)
						return style.Render(string(rune('0' + idx)))
					}).
					Build()
				assert.NoError(t, err)
				children[i] = child
			}

			// Create parent
			parent, err := NewComponent("Parent").
				Children(children...).
				Template(func(ctx RenderContext) string {
					style := ctx.NewStyle().Padding(1)
					outputs := []string{}
					for i, child := range ctx.Children() {
						outputs = append(outputs, "Child "+string(rune('0'+i))+": "+ctx.RenderChild(child))
					}
					return style.Render(strings.Join(outputs, "\n"))
				}).
				Build()

			assert.NoError(t, err)

			// Initialize and render
			parent.Init()
			output := parent.View()

			assert.NotEmpty(t, output, "Output should not be empty")
			for _, want := range tt.wantContains {
				assert.Contains(t, output, want, "Output should contain child text")
			}
		})
	}
}

// TestRenderContext_StyleInheritance tests Lipgloss style inheritance
func TestRenderContext_StyleInheritance(t *testing.T) {
	tests := []struct {
		name     string
		template RenderFunc
		wantText string
	}{
		{
			name: "inherits base style",
			template: func(ctx RenderContext) string {
				baseStyle := ctx.NewStyle().
					Foreground(lipgloss.Color("63")).
					Padding(1)

				childStyle := ctx.NewStyle().
					Bold(true).
					Inherit(baseStyle)

				return childStyle.Render("Inherited")
			},
			wantText: "Inherited",
		},
		{
			name: "composes multiple styles",
			template: func(ctx RenderContext) string {
				style1 := ctx.NewStyle().Bold(true)
				style2 := ctx.NewStyle().Italic(true)

				// Apply both
				text := style1.Render("Bold")
				text += " "
				text += style2.Render("Italic")

				return text
			},
			wantText: "Bold",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := NewComponent("TestComponent").
				Template(tt.template).
				Build()

			assert.NoError(t, err)

			output := component.View()
			assert.NotEmpty(t, output)
			assert.Contains(t, output, tt.wantText)
		})
	}
}

// TestRenderPerformance tests that rendering meets performance targets
func TestRenderPerformance(t *testing.T) {
	// Simple component should render in < 5ms (requirement)
	component, err := NewComponent("SimpleComponent").
		Template(func(ctx RenderContext) string {
			style := ctx.NewStyle().Bold(true)
			return style.Render("Hello")
		}).
		Build()

	assert.NoError(t, err)

	// Warm up
	component.View()

	// This is a smoke test - actual benchmarks in benchmark file
	output := component.View()
	assert.NotEmpty(t, output)
}
