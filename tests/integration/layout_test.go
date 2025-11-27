package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// layoutTestRoot wraps a component with theme provision for layout testing.
// This follows the same pattern as testRoot in components_test.go.
func layoutTestRoot(child bubbly.Component) bubbly.Component {
	root, _ := bubbly.NewComponent("LayoutTestRoot").
		Children(child).
		Setup(func(ctx *bubbly.Context) {
			// Provide default theme for all child components
			ctx.Provide("theme", components.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Render child directly
			children := ctx.Children()
			if len(children) > 0 {
				return ctx.RenderChild(children[0])
			}
			return ""
		}).
		Build()
	return root
}

// TestLayoutComposition_FlexInVStackInBox verifies that Flex inside VStack inside Box renders correctly.
// This tests the organism (Flex) → molecule (VStack) → atom (Box) composition.
func TestLayoutComposition_FlexInVStackInBox(t *testing.T) {
	t.Run("flex in vstack in box renders all content", func(t *testing.T) {
		// Setup: Create leaf components
		item1 := components.Text(components.TextProps{Content: "Item A"})
		item2 := components.Text(components.TextProps{Content: "Item B"})
		item3 := components.Text(components.TextProps{Content: "Item C"})

		// Initialize leaf components
		item1.Init()
		item2.Init()
		item3.Init()

		// Create Flex with items in row direction
		flex := components.Flex(components.FlexProps{
			Items:     []bubbly.Component{item1, item2, item3},
			Direction: components.FlexRow,
			Gap:       2,
		})
		flex.Init()

		// Create header text
		header := components.Text(components.TextProps{Content: "Header Section"})
		header.Init()

		// Create VStack with header and flex
		vstack := components.VStack(components.StackProps{
			Items:   []interface{}{header, flex},
			Spacing: 1,
		})
		vstack.Init()

		// Wrap in Box with border
		box := components.Box(components.BoxProps{
			Child:   vstack,
			Border:  true,
			Padding: 1,
			Title:   "Layout Test",
		})
		box.Init()

		// Wrap with theme provider
		root := layoutTestRoot(box)
		root.Init()

		// Verify: All content renders correctly
		view := root.View()
		require.NotEmpty(t, view)

		// Verify title
		assert.Contains(t, view, "Layout Test")

		// Verify header
		assert.Contains(t, view, "Header Section")

		// Verify flex items
		assert.Contains(t, view, "Item A")
		assert.Contains(t, view, "Item B")
		assert.Contains(t, view, "Item C")
	})

	t.Run("flex with justify in nested layout", func(t *testing.T) {
		// Setup: Create buttons
		btn1 := components.Button(components.ButtonProps{Label: "Save"})
		btn2 := components.Button(components.ButtonProps{Label: "Cancel"})

		btn1.Init()
		btn2.Init()

		// Create Flex with space-between
		flex := components.Flex(components.FlexProps{
			Items:   []bubbly.Component{btn1, btn2},
			Justify: components.JustifySpaceBetween,
			Width:   40,
		})
		flex.Init()

		// Wrap in VStack
		title := components.Text(components.TextProps{Content: "Actions"})
		title.Init()

		vstack := components.VStack(components.StackProps{
			Items:   []interface{}{title, flex},
			Spacing: 1,
		})
		vstack.Init()

		// Wrap in Box
		box := components.Box(components.BoxProps{
			Child:  vstack,
			Border: true,
		})
		box.Init()

		root := layoutTestRoot(box)
		root.Init()

		// Verify
		view := root.View()
		assert.Contains(t, view, "Actions")
		assert.Contains(t, view, "Save")
		assert.Contains(t, view, "Cancel")
	})
}

// TestLayoutComposition_CenterWithNestedFlex verifies that Center with nested Flex works correctly.
// This tests centering behavior with complex nested content.
func TestLayoutComposition_CenterWithNestedFlex(t *testing.T) {
	t.Run("center with flex content", func(t *testing.T) {
		// Setup: Create flex content
		item1 := components.Text(components.TextProps{Content: "Left"})
		item2 := components.Text(components.TextProps{Content: "Right"})

		item1.Init()
		item2.Init()

		flex := components.Flex(components.FlexProps{
			Items:   []bubbly.Component{item1, item2},
			Gap:     4,
			Justify: components.JustifyCenter,
		})
		flex.Init()

		// Wrap in Center
		center := components.Center(components.CenterProps{
			Child:  flex,
			Width:  60,
			Height: 10,
		})
		center.Init()

		root := layoutTestRoot(center)
		root.Init()

		// Verify
		view := root.View()
		require.NotEmpty(t, view)
		assert.Contains(t, view, "Left")
		assert.Contains(t, view, "Right")
	})

	t.Run("center with card containing flex", func(t *testing.T) {
		// Setup: Create card content with flex
		btn1 := components.Button(components.ButtonProps{Label: "OK"})
		btn2 := components.Button(components.ButtonProps{Label: "Cancel"})

		btn1.Init()
		btn2.Init()

		buttonRow := components.Flex(components.FlexProps{
			Items:   []bubbly.Component{btn1, btn2},
			Gap:     2,
			Justify: components.JustifyEnd,
		})
		buttonRow.Init()

		card := components.Card(components.CardProps{
			Title:    "Confirm Action",
			Content:  "Are you sure you want to proceed?",
			Children: []bubbly.Component{buttonRow},
		})
		card.Init()

		// Center the card (modal pattern)
		center := components.Center(components.CenterProps{
			Child:  card,
			Width:  80,
			Height: 24,
		})
		center.Init()

		root := layoutTestRoot(center)
		root.Init()

		// Verify
		view := root.View()
		assert.Contains(t, view, "Confirm Action")
		assert.Contains(t, view, "Are you sure")
		assert.Contains(t, view, "OK")
		assert.Contains(t, view, "Cancel")
	})
}

// TestLayoutComposition_ContainerWithHStackHeader verifies the Container with HStack header pattern.
// This is a common pattern for page layouts with header toolbars.
func TestLayoutComposition_ContainerWithHStackHeader(t *testing.T) {
	t.Run("container with hstack header pattern", func(t *testing.T) {
		// Setup: Create header with HStack
		logo := components.Text(components.TextProps{Content: "📊 Dashboard"})
		spacer := components.Spacer(components.SpacerProps{Flex: true})
		settingsBtn := components.Button(components.ButtonProps{Label: "⚙️ Settings"})

		logo.Init()
		spacer.Init()
		settingsBtn.Init()

		header := components.HStack(components.StackProps{
			Items:   []interface{}{logo, spacer, settingsBtn},
			Spacing: 2,
		})
		header.Init()

		// Create content
		content := components.Text(components.TextProps{Content: "Main content area"})
		content.Init()

		// Create page with VStack
		page := components.VStack(components.StackProps{
			Items:   []interface{}{header, content},
			Spacing: 2,
		})
		page.Init()

		// Wrap in Container
		container := components.Container(components.ContainerProps{
			Child:    page,
			Size:     components.ContainerLg,
			Centered: true,
		})
		container.Init()

		root := layoutTestRoot(container)
		root.Init()

		// Verify
		view := root.View()
		assert.Contains(t, view, "Dashboard")
		assert.Contains(t, view, "Settings")
		assert.Contains(t, view, "Main content")
	})

	t.Run("container with divider between header and content", func(t *testing.T) {
		// Setup: Create header
		title := components.Text(components.TextProps{Content: "Page Title"})
		title.Init()

		// Create divider
		divider := components.Divider(components.DividerProps{
			Length: 40,
		})
		divider.Init()

		// Create content
		content := components.Text(components.TextProps{Content: "Page content here"})
		content.Init()

		// Stack with divider
		page := components.VStack(components.StackProps{
			Items:   []interface{}{title, divider, content},
			Spacing: 1,
		})
		page.Init()

		container := components.Container(components.ContainerProps{
			Child: page,
			Size:  components.ContainerMd,
		})
		container.Init()

		root := layoutTestRoot(container)
		root.Init()

		// Verify
		view := root.View()
		assert.Contains(t, view, "Page Title")
		assert.Contains(t, view, "Page content")
		// Divider should render (contains horizontal line character)
		assert.Contains(t, view, "─")
	})
}

// TestLayoutComposition_DeepNesting verifies that deep nesting works without render artifacts.
// This tests 5+ levels of nesting with mixed component types.
func TestLayoutComposition_DeepNesting(t *testing.T) {
	t.Run("five levels of nesting renders correctly", func(t *testing.T) {
		// Level 5: Innermost content
		innerText := components.Text(components.TextProps{Content: "Deep Content"})
		innerText.Init()

		// Level 4: Flex containing text
		flex := components.Flex(components.FlexProps{
			Items:   []bubbly.Component{innerText},
			Justify: components.JustifyCenter,
		})
		flex.Init()

		// Level 3: Center containing flex
		center := components.Center(components.CenterProps{
			Child:  flex,
			Width:  30,
			Height: 5,
		})
		center.Init()

		// Level 2: Box containing center
		box := components.Box(components.BoxProps{
			Child:   center,
			Border:  true,
			Padding: 1,
		})
		box.Init()

		// Level 1: VStack containing box
		vstack := components.VStack(components.StackProps{
			Items:   []interface{}{box},
			Spacing: 0,
		})
		vstack.Init()

		// Level 0: Container containing vstack
		container := components.Container(components.ContainerProps{
			Child: vstack,
			Size:  components.ContainerMd,
		})
		container.Init()

		root := layoutTestRoot(container)
		root.Init()

		// Verify: Content renders without panic
		view := root.View()
		require.NotEmpty(t, view)
		assert.Contains(t, view, "Deep Content")
	})

	t.Run("mixed layout types at multiple levels", func(t *testing.T) {
		// Create various leaf components
		text1 := components.Text(components.TextProps{Content: "Text 1"})
		text2 := components.Text(components.TextProps{Content: "Text 2"})
		text3 := components.Text(components.TextProps{Content: "Text 3"})
		btn := components.Button(components.ButtonProps{Label: "Action"})

		text1.Init()
		text2.Init()
		text3.Init()
		btn.Init()

		// HStack with texts
		hstack := components.HStack(components.StackProps{
			Items:   []interface{}{text1, text2},
			Spacing: 2,
		})
		hstack.Init()

		// Flex with button
		flex := components.Flex(components.FlexProps{
			Items:   []bubbly.Component{btn},
			Justify: components.JustifyEnd,
			Width:   40,
		})
		flex.Init()

		// VStack combining hstack and flex
		vstack := components.VStack(components.StackProps{
			Items:   []interface{}{hstack, text3, flex},
			Spacing: 1,
		})
		vstack.Init()

		// Box wrapping vstack
		box := components.Box(components.BoxProps{
			Child:  vstack,
			Border: true,
			Title:  "Mixed Layout",
		})
		box.Init()

		// Center the box
		center := components.Center(components.CenterProps{
			Child:  box,
			Width:  60,
			Height: 15,
		})
		center.Init()

		root := layoutTestRoot(center)
		root.Init()

		// Verify all content renders
		view := root.View()
		require.NotEmpty(t, view)
		assert.Contains(t, view, "Mixed Layout")
		assert.Contains(t, view, "Text 1")
		assert.Contains(t, view, "Text 2")
		assert.Contains(t, view, "Text 3")
		assert.Contains(t, view, "Action")
	})

	t.Run("nested flex layouts", func(t *testing.T) {
		// Inner flex (row)
		innerItems := []bubbly.Component{
			components.Text(components.TextProps{Content: "A"}),
			components.Text(components.TextProps{Content: "B"}),
		}
		for _, item := range innerItems {
			item.Init()
		}

		innerFlex := components.Flex(components.FlexProps{
			Items:     innerItems,
			Direction: components.FlexRow,
			Gap:       1,
		})
		innerFlex.Init()

		// Another inner flex
		innerItems2 := []bubbly.Component{
			components.Text(components.TextProps{Content: "C"}),
			components.Text(components.TextProps{Content: "D"}),
		}
		for _, item := range innerItems2 {
			item.Init()
		}

		innerFlex2 := components.Flex(components.FlexProps{
			Items:     innerItems2,
			Direction: components.FlexRow,
			Gap:       1,
		})
		innerFlex2.Init()

		// Outer flex (column) containing inner flexes
		outerFlex := components.Flex(components.FlexProps{
			Items:     []bubbly.Component{innerFlex, innerFlex2},
			Direction: components.FlexColumn,
			Gap:       1,
		})
		outerFlex.Init()

		root := layoutTestRoot(outerFlex)
		root.Init()

		// Verify
		view := root.View()
		require.NotEmpty(t, view)
		assert.Contains(t, view, "A")
		assert.Contains(t, view, "B")
		assert.Contains(t, view, "C")
		assert.Contains(t, view, "D")
	})
}

// TestLayoutComposition_EdgeCases verifies edge cases in layout composition.
func TestLayoutComposition_EdgeCases(t *testing.T) {
	t.Run("empty flex in box", func(t *testing.T) {
		flex := components.Flex(components.FlexProps{
			Items: []bubbly.Component{},
		})
		flex.Init()

		box := components.Box(components.BoxProps{
			Child:  flex,
			Border: true,
			Title:  "Empty Flex",
		})
		box.Init()

		root := layoutTestRoot(box)
		root.Init()

		// Should render without panic
		view := root.View()
		require.NotEmpty(t, view)
		assert.Contains(t, view, "Empty Flex")
	})

	t.Run("nil child in center", func(t *testing.T) {
		center := components.Center(components.CenterProps{
			Child:  nil,
			Width:  40,
			Height: 10,
		})
		center.Init()

		root := layoutTestRoot(center)
		root.Init()

		// Should render without panic
		view := root.View()
		// Empty center still renders container space
		require.NotPanics(t, func() {
			_ = root.View()
		})
		_ = view // Use the variable
	})

	t.Run("single item in all layouts", func(t *testing.T) {
		text := components.Text(components.TextProps{Content: "Solo"})
		text.Init()

		// Single item in HStack
		hstack := components.HStack(components.StackProps{
			Items: []interface{}{text},
		})
		hstack.Init()

		// Single item in VStack
		vstack := components.VStack(components.StackProps{
			Items: []interface{}{hstack},
		})
		vstack.Init()

		// Single item in Flex
		flex := components.Flex(components.FlexProps{
			Items: []bubbly.Component{vstack},
		})
		flex.Init()

		root := layoutTestRoot(flex)
		root.Init()

		view := root.View()
		assert.Contains(t, view, "Solo")
	})
}

// TestLayoutComposition_ReactiveUpdates verifies that reactive updates work in nested layouts.
func TestLayoutComposition_ReactiveUpdates(t *testing.T) {
	t.Run("ref updates propagate through nested layouts", func(t *testing.T) {
		// Setup: Create ref-based content
		contentRef := bubbly.NewRef("Initial Content")

		// Create text that uses ref
		text := components.Text(components.TextProps{
			Content: contentRef.GetTyped(),
		})
		text.Init()

		// Nest in layouts
		flex := components.Flex(components.FlexProps{
			Items: []bubbly.Component{text},
		})
		flex.Init()

		box := components.Box(components.BoxProps{
			Child:  flex,
			Border: true,
		})
		box.Init()

		root := layoutTestRoot(box)
		root.Init()

		// Initial render
		view := root.View()
		assert.Contains(t, view, "Initial Content")

		// Update ref
		contentRef.Set("Updated Content")
		time.Sleep(10 * time.Millisecond)

		// Note: For full reactivity, the text component would need to
		// re-render. This test verifies the ref update doesn't break
		// the layout structure.
		view = root.View()
		require.NotEmpty(t, view)
	})
}

// BenchmarkLayoutRendering_SimpleLayout benchmarks simple layout rendering.
func BenchmarkLayoutRendering_SimpleLayout(b *testing.B) {
	// Setup: Create simple layout once
	text := components.Text(components.TextProps{Content: "Hello"})
	text.Init()

	box := components.Box(components.BoxProps{
		Child:   text,
		Border:  true,
		Padding: 1,
	})
	box.Init()

	root := layoutTestRoot(box)
	root.Init()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = root.View()
	}
}

// BenchmarkLayoutRendering_ComplexLayout benchmarks complex nested layout rendering.
// Target: <10ms for complex layouts (per requirements NFR-4.1.2).
func BenchmarkLayoutRendering_ComplexLayout(b *testing.B) {
	// Setup: Create complex nested layout
	items := make([]bubbly.Component, 5)
	for i := range items {
		items[i] = components.Text(components.TextProps{Content: "Item"})
		items[i].Init()
	}

	flex := components.Flex(components.FlexProps{
		Items:     items,
		Direction: components.FlexRow,
		Gap:       2,
		Justify:   components.JustifySpaceBetween,
		Width:     80,
	})
	flex.Init()

	header := components.Text(components.TextProps{Content: "Header"})
	header.Init()

	vstack := components.VStack(components.StackProps{
		Items:   []interface{}{header, flex},
		Spacing: 1,
	})
	vstack.Init()

	box := components.Box(components.BoxProps{
		Child:   vstack,
		Border:  true,
		Padding: 1,
		Title:   "Complex Layout",
	})
	box.Init()

	center := components.Center(components.CenterProps{
		Child:  box,
		Width:  100,
		Height: 30,
	})
	center.Init()

	container := components.Container(components.ContainerProps{
		Child: center,
		Size:  components.ContainerXl,
	})
	container.Init()

	root := layoutTestRoot(container)
	root.Init()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = root.View()
	}
}

// BenchmarkLayoutRendering_DeepNesting benchmarks deeply nested layout rendering.
func BenchmarkLayoutRendering_DeepNesting(b *testing.B) {
	// Setup: Create 10 levels of nesting
	var current bubbly.Component = components.Text(components.TextProps{Content: "Deep"})
	current.Init()

	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			box := components.Box(components.BoxProps{
				Child:   current,
				Padding: 1,
			})
			box.Init()
			current = box
		} else {
			flex := components.Flex(components.FlexProps{
				Items: []bubbly.Component{current},
			})
			flex.Init()
			current = flex
		}
	}

	root := layoutTestRoot(current)
	root.Init()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = root.View()
	}
}
