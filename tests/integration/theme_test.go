package integration

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestTheme_ParentChildInjection verifies basic theme injection from parent to child.
// Parent provides custom theme, child uses UseTheme to retrieve it.
func TestTheme_ParentChildInjection(t *testing.T) {
	// Define custom theme
	customTheme := bubbly.DefaultTheme
	customTheme.Primary = lipgloss.Color("99")    // Purple
	customTheme.Secondary = lipgloss.Color("120") // Custom accent

	var childTheme bubbly.Theme

	// Create child component that uses theme
	child, err := bubbly.NewComponent("Child").
		Setup(func(ctx *bubbly.Context) {
			// Use theme from parent or default
			childTheme = ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", childTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			return lipgloss.NewStyle().
				Foreground(theme.Primary).
				Render("Child with theme")
		}).
		Build()

	require.NoError(t, err)

	// Create parent component that provides theme
	parent, err := bubbly.NewComponent("Parent").
		Children(child).
		Setup(func(ctx *bubbly.Context) {
			// Provide custom theme to descendants
			ctx.ProvideTheme(customTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			output := "Parent:\n"
			for _, c := range ctx.Children() {
				output += ctx.RenderChild(c)
			}
			return output
		}).
		Build()

	require.NoError(t, err)

	// Initialize parent (which initializes child)
	parent.Init()

	// Verify child received parent's theme
	assert.Equal(t, customTheme.Primary, childTheme.Primary, "Child should receive parent's Primary color")
	assert.Equal(t, customTheme.Secondary, childTheme.Secondary, "Child should receive parent's Secondary color")
	assert.Equal(t, customTheme, childTheme, "Child should receive complete parent theme")

	// Verify rendering works
	view := parent.View()
	assert.Contains(t, view, "Child with theme")
}

// TestTheme_ThreeLevelHierarchy verifies theme propagation through 3 levels:
// Grandparent → Parent → Child
func TestTheme_ThreeLevelHierarchy(t *testing.T) {
	// Custom theme at top level
	topTheme := bubbly.DefaultTheme
	topTheme.Primary = lipgloss.Color("196")   // Red
	topTheme.Secondary = lipgloss.Color("220") // Yellow

	var parentTheme, childTheme bubbly.Theme

	// Create grandchild (3rd level)
	grandchild, err := bubbly.NewComponent("Grandchild").
		Setup(func(ctx *bubbly.Context) {
			childTheme = ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", childTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Grandchild"
		}).
		Build()

	require.NoError(t, err)

	// Create child (2nd level) - passes theme through
	child, err := bubbly.NewComponent("Child").
		Children(grandchild).
		Setup(func(ctx *bubbly.Context) {
			parentTheme = ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", parentTheme)
			// Don't provide - let theme pass through
		}).
		Template(func(ctx bubbly.RenderContext) string {
			output := "Child:\n"
			for _, c := range ctx.Children() {
				output += ctx.RenderChild(c)
			}
			return output
		}).
		Build()

	require.NoError(t, err)

	// Create grandparent (1st level) - provides theme
	grandparent, err := bubbly.NewComponent("Grandparent").
		Children(child).
		Setup(func(ctx *bubbly.Context) {
			ctx.ProvideTheme(topTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			output := "Grandparent:\n"
			for _, c := range ctx.Children() {
				output += ctx.RenderChild(c)
			}
			return output
		}).
		Build()

	require.NoError(t, err)

	// Initialize hierarchy
	grandparent.Init()

	// Verify theme propagated through all 3 levels
	assert.Equal(t, topTheme.Primary, parentTheme.Primary, "Parent (2nd level) should inherit grandparent's theme")
	assert.Equal(t, topTheme.Primary, childTheme.Primary, "Grandchild (3rd level) should inherit grandparent's theme")
	assert.Equal(t, topTheme, parentTheme, "Parent should have complete grandparent theme")
	assert.Equal(t, topTheme, childTheme, "Grandchild should have complete grandparent theme")
}

// TestTheme_LocalOverride verifies that a component can override theme for its subtree
// while parent and siblings maintain the original theme. This demonstrates theme isolation:
// - App provides base theme (green)
// - RegularChild uses base theme
// - ModalWrapper provides overridden theme (purple) for its children
// - ModalContent uses the overridden theme
// This pattern allows different parts of the UI to have independent themes.
func TestTheme_LocalOverride(t *testing.T) {
	// Base theme for most components
	baseTheme := bubbly.DefaultTheme
	baseTheme.Primary = lipgloss.Color("35") // Green

	// Special theme for modal subtree
	modalTheme := bubbly.Theme{
		Primary:    lipgloss.Color("99"),  // Purple override
		Secondary:  baseTheme.Secondary,   // Keep from base
		Muted:      baseTheme.Muted,       // Keep from base
		Warning:    baseTheme.Warning,     // Keep from base
		Error:      baseTheme.Error,       // Keep from base
		Success:    baseTheme.Success,     // Keep from base
		Background: lipgloss.Color("232"), // Darker background
	}

	var regularChildTheme, modalContentTheme bubbly.Theme

	// Regular child - uses base theme
	regularChild, err := bubbly.NewComponent("RegularChild").
		Setup(func(ctx *bubbly.Context) {
			regularChildTheme = ctx.UseTheme(bubbly.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Regular child"
		}).
		Build()

	require.NoError(t, err)

	// Modal content - should get modal theme
	modalContent, err := bubbly.NewComponent("ModalContent").
		Setup(func(ctx *bubbly.Context) {
			modalContentTheme = ctx.UseTheme(bubbly.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Modal content"
		}).
		Build()

	require.NoError(t, err)

	// Modal wrapper - provides special theme for its children
	modalWrapper, err := bubbly.NewComponent("ModalWrapper").
		Children(modalContent).
		Setup(func(ctx *bubbly.Context) {
			// Provide modal-specific theme to children
			ctx.ProvideTheme(modalTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			output := "Modal:\n"
			for _, c := range ctx.Children() {
				output += ctx.RenderChild(c)
			}
			return output
		}).
		Build()

	require.NoError(t, err)

	// App root - provides base theme
	app, err := bubbly.NewComponent("App").
		Children(regularChild, modalWrapper).
		Setup(func(ctx *bubbly.Context) {
			ctx.ProvideTheme(baseTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			output := "App:\n"
			for _, c := range ctx.Children() {
				output += ctx.RenderChild(c)
			}
			return output
		}).
		Build()

	require.NoError(t, err)

	// Initialize
	app.Init()

	// Verify theme isolation
	assert.Equal(t, baseTheme.Primary, regularChildTheme.Primary, "Regular child should have base theme")
	assert.Equal(t, modalTheme.Primary, modalContentTheme.Primary, "Modal content should have modal theme")
	assert.NotEqual(t, regularChildTheme.Primary, modalContentTheme.Primary, "Themes should be different in different subtrees")
	assert.Equal(t, modalTheme.Background, modalContentTheme.Background, "Modal content should have modal's darker background")
	assert.Equal(t, lipgloss.Color("35"), regularChildTheme.Primary, "Regular child should have green")
	assert.Equal(t, lipgloss.Color("99"), modalContentTheme.Primary, "Modal content should have purple")
}

// TestTheme_MultipleSubtrees verifies that independent subtrees can have
// completely different themes without interference.
func TestTheme_MultipleSubtrees(t *testing.T) {
	// Theme for subtree A
	themeA := bubbly.DefaultTheme
	themeA.Primary = lipgloss.Color("35") // Green

	// Theme for subtree B
	themeB := bubbly.DefaultTheme
	themeB.Primary = lipgloss.Color("196") // Red

	var childATheme, childBTheme bubbly.Theme

	// Child in subtree A
	childA, err := bubbly.NewComponent("ChildA").
		Setup(func(ctx *bubbly.Context) {
			childATheme = ctx.UseTheme(bubbly.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Child A"
		}).
		Build()

	require.NoError(t, err)

	// Child in subtree B
	childB, err := bubbly.NewComponent("ChildB").
		Setup(func(ctx *bubbly.Context) {
			childBTheme = ctx.UseTheme(bubbly.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Child B"
		}).
		Build()

	require.NoError(t, err)

	// Subtree A root
	subtreeA, err := bubbly.NewComponent("SubtreeA").
		Children(childA).
		Setup(func(ctx *bubbly.Context) {
			ctx.ProvideTheme(themeA)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			output := "Subtree A:\n"
			for _, c := range ctx.Children() {
				output += ctx.RenderChild(c)
			}
			return output
		}).
		Build()

	require.NoError(t, err)

	// Subtree B root
	subtreeB, err := bubbly.NewComponent("SubtreeB").
		Children(childB).
		Setup(func(ctx *bubbly.Context) {
			ctx.ProvideTheme(themeB)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			output := "Subtree B:\n"
			for _, c := range ctx.Children() {
				output += ctx.RenderChild(c)
			}
			return output
		}).
		Build()

	require.NoError(t, err)

	// Initialize both subtrees independently
	subtreeA.Init()
	subtreeB.Init()

	// Verify complete isolation
	assert.Equal(t, themeA.Primary, childATheme.Primary, "Child A should have subtree A's theme")
	assert.Equal(t, themeB.Primary, childBTheme.Primary, "Child B should have subtree B's theme")
	assert.NotEqual(t, childATheme.Primary, childBTheme.Primary, "Subtrees should have different themes")
	assert.Equal(t, lipgloss.Color("35"), childATheme.Primary, "Child A should have green")
	assert.Equal(t, lipgloss.Color("196"), childBTheme.Primary, "Child B should have red")
}

// TestTheme_DefaultWhenNoProvider verifies graceful fallback to default theme
// when no parent provides a theme.
func TestTheme_DefaultWhenNoProvider(t *testing.T) {
	var receivedTheme bubbly.Theme

	// Create component without parent providing theme
	component, err := bubbly.NewComponent("Standalone").
		Setup(func(ctx *bubbly.Context) {
			// No parent provides theme - should use default
			receivedTheme = ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", receivedTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			return lipgloss.NewStyle().
				Foreground(theme.Primary).
				Render("Standalone component")
		}).
		Build()

	require.NoError(t, err)

	// Initialize without parent
	component.Init()

	// Verify default theme is used
	assert.Equal(t, bubbly.DefaultTheme.Primary, receivedTheme.Primary, "Should use default Primary")
	assert.Equal(t, bubbly.DefaultTheme.Secondary, receivedTheme.Secondary, "Should use default Secondary")
	assert.Equal(t, bubbly.DefaultTheme.Muted, receivedTheme.Muted, "Should use default Muted")
	assert.Equal(t, bubbly.DefaultTheme.Warning, receivedTheme.Warning, "Should use default Warning")
	assert.Equal(t, bubbly.DefaultTheme.Error, receivedTheme.Error, "Should use default Error")
	assert.Equal(t, bubbly.DefaultTheme.Success, receivedTheme.Success, "Should use default Success")
	assert.Equal(t, bubbly.DefaultTheme.Background, receivedTheme.Background, "Should use default Background")
	assert.Equal(t, bubbly.DefaultTheme, receivedTheme, "Should use complete default theme")

	// Verify rendering works
	view := component.View()
	assert.Contains(t, view, "Standalone component")
}

// TestTheme_InvalidTypeInjection verifies graceful handling when theme is
// provided as wrong type (type assertion failure).
func TestTheme_InvalidTypeInjection(t *testing.T) {
	var receivedTheme bubbly.Theme

	// Create child that expects theme
	child, err := bubbly.NewComponent("Child").
		Setup(func(ctx *bubbly.Context) {
			// Should gracefully fall back to default on type mismatch
			receivedTheme = ctx.UseTheme(bubbly.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Child"
		}).
		Build()

	require.NoError(t, err)

	// Create parent that provides WRONG TYPE (string instead of Theme)
	parent, err := bubbly.NewComponent("Parent").
		Children(child).
		Setup(func(ctx *bubbly.Context) {
			// Provide wrong type - should cause type assertion to fail
			ctx.Provide("theme", "not a theme struct") // Invalid!
		}).
		Template(func(ctx bubbly.RenderContext) string {
			output := "Parent:\n"
			for _, c := range ctx.Children() {
				output += ctx.RenderChild(c)
			}
			return output
		}).
		Build()

	require.NoError(t, err)

	// Initialize
	parent.Init()

	// Verify graceful fallback to default (no panic)
	assert.Equal(t, bubbly.DefaultTheme, receivedTheme, "Should fall back to default on type mismatch")
	assert.Equal(t, bubbly.DefaultTheme.Primary, receivedTheme.Primary, "Should use default Primary")
}

// TestTheme_MixedOldNewPatterns verifies backward compatibility:
// Old manual inject/expose patterns work alongside new UseTheme/ProvideTheme.
func TestTheme_MixedOldNewPatterns(t *testing.T) {
	customColor := lipgloss.Color("99") // Purple
	customTheme := bubbly.DefaultTheme
	customTheme.Primary = lipgloss.Color("35") // Green

	var oldPatternColor lipgloss.Color
	var newPatternTheme bubbly.Theme

	// Child using OLD pattern (manual inject/expose)
	oldChild, err := bubbly.NewComponent("OldPatternChild").
		Setup(func(ctx *bubbly.Context) {
			// Old pattern: manual inject with default
			oldPatternColor = lipgloss.Color("240") // Default grey
			if injected := ctx.Inject("customColor", nil); injected != nil {
				if color, ok := injected.(lipgloss.Color); ok {
					oldPatternColor = color
				}
			}
			ctx.Expose("color", oldPatternColor)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Old pattern child"
		}).
		Build()

	require.NoError(t, err)

	// Child using NEW pattern (UseTheme)
	newChild, err := bubbly.NewComponent("NewPatternChild").
		Setup(func(ctx *bubbly.Context) {
			// New pattern: UseTheme
			newPatternTheme = ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", newPatternTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "New pattern child"
		}).
		Build()

	require.NoError(t, err)

	// Parent provides both old and new patterns
	parent, err := bubbly.NewComponent("Parent").
		Children(oldChild, newChild).
		Setup(func(ctx *bubbly.Context) {
			// Old pattern: manual provide
			ctx.Provide("customColor", customColor)

			// New pattern: ProvideTheme
			ctx.ProvideTheme(customTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			output := "Parent:\n"
			for _, c := range ctx.Children() {
				output += ctx.RenderChild(c)
			}
			return output
		}).
		Build()

	require.NoError(t, err)

	// Initialize
	parent.Init()

	// Verify both patterns work
	assert.Equal(t, customColor, oldPatternColor, "Old pattern should receive injected color")
	assert.Equal(t, customTheme.Primary, newPatternTheme.Primary, "New pattern should receive theme")
	assert.Equal(t, customTheme, newPatternTheme, "New pattern should receive complete theme")

	// Verify rendering works for both
	view := parent.View()
	assert.Contains(t, view, "Old pattern child")
	assert.Contains(t, view, "New pattern child")
}
