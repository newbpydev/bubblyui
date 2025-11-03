package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ListProps defines the configuration properties for a List component.
//
// List is a generic component that works with any slice type []T.
// It displays a scrollable list of items with optional virtual scrolling for performance,
// keyboard navigation, and item selection.
//
// Example usage:
//
//	type Todo struct {
//	    Title     string
//	    Completed bool
//	}
//
//	todosData := bubbly.NewRef([]Todo{
//	    {Title: "Buy groceries", Completed: false},
//	    {Title: "Write code", Completed: true},
//	})
//
//	list := components.List(components.ListProps[Todo]{
//	    Items: todosData,
//	    RenderItem: func(todo Todo, index int) string {
//	        checkbox := "☐"
//	        if todo.Completed {
//	            checkbox = "☑"
//	        }
//	        return fmt.Sprintf("%s %s", checkbox, todo.Title)
//	    },
//	    Height: 10,
//	    Virtual: true,
//	})
type ListProps[T any] struct {
	// Items is a reactive reference to the list data.
	// Must be a slice of type T.
	// Required - updates trigger re-renders.
	Items *bubbly.Ref[[]T]

	// RenderItem is a function that renders each item.
	// Receives the item and its index as parameters.
	// Required - defines how each item is displayed.
	RenderItem func(T, int) string

	// Height is the visible height of the list in lines.
	// Items beyond this height require scrolling.
	// Optional - defaults to 10 if not specified.
	Height int

	// Virtual enables virtual scrolling for large lists.
	// When true, only visible items are rendered for performance.
	// Recommended for lists with 100+ items.
	// Optional - defaults to false.
	Virtual bool

	// OnSelect is a callback function executed when an item is selected.
	// Receives the selected item and its index as parameters.
	// Optional - if nil, no callback is executed.
	OnSelect func(T, int)

	// Common props for all components
	CommonProps
}

// List creates a new List organism component with generic type support.
//
// The List component provides:
//   - Generic type support for any item type T
//   - Reactive data binding with automatic re-rendering
//   - Keyboard navigation (up/down arrows, j/k vim keys, enter to select)
//   - Item selection with visual highlighting
//   - Virtual scrolling for performance with large datasets
//   - Custom item rendering via RenderItem function
//   - Theme integration for consistent styling
//
// Keyboard controls:
//   - ↑/k: Move selection up
//   - ↓/j: Move selection down
//   - Enter/Space: Select current item
//   - Home: Jump to first item
//   - End: Jump to last item
//
// The component integrates with the framework's reactivity system,
// automatically updating when the Items ref changes.
//
// Example:
//
//	list := components.List(components.ListProps[string]{
//	    Items: itemsRef,
//	    RenderItem: func(item string, i int) string {
//	        return fmt.Sprintf("%d. %s", i+1, item)
//	    },
//	    Height: 10,
//	    Virtual: true,
//	    OnSelect: func(item string, index int) {
//	        handleSelection(item)
//	    },
//	})
func List[T any](props ListProps[T]) bubbly.Component {
	comp, err := bubbly.NewComponent("List").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			// Inject theme for styling
			theme := ctx.Inject("theme", DefaultTheme).(Theme)

			// Internal state
			selectedIndex := bubbly.NewRef(-1) // Currently selected item index (-1 = none)
			scrollOffset := bubbly.NewRef(0)   // Scroll position for virtual scrolling

			// Expose state for testing
			ctx.Expose("selectedIndex", selectedIndex)
			ctx.Expose("scrollOffset", scrollOffset)
			ctx.Expose("theme", theme)

			// Helper function to get visible height
			getHeight := func() int {
				if props.Height > 0 {
					return props.Height
				}
				return 10 // Default height
			}

			// Helper function to select an item
			selectItem := func(index int) {
				items := props.Items.Get().([]T)
				if index < 0 || index >= len(items) {
					return
				}

				selectedIndex.Set(index)

				// Adjust scroll offset for virtual scrolling
				height := getHeight()
				offset := scrollOffset.Get().(int)

				// Scroll down if selected item is below visible area
				if index >= offset+height {
					scrollOffset.Set(index - height + 1)
				}

				// Scroll up if selected item is above visible area
				if index < offset {
					scrollOffset.Set(index)
				}

				// Trigger callback
				if props.OnSelect != nil {
					props.OnSelect(items[index], index)
				}
			}

			// Keyboard navigation: Move down
			ctx.On("keyDown", func(_ interface{}) {
				items := props.Items.Get().([]T)
				if len(items) == 0 {
					return
				}

				current := selectedIndex.Get().(int)
				if current == -1 {
					// No selection, select first item
					selectItem(0)
				} else if current < len(items)-1 {
					// Move down
					selectItem(current + 1)
				}
			})

			// Keyboard navigation: Move up
			ctx.On("keyUp", func(_ interface{}) {
				items := props.Items.Get().([]T)
				if len(items) == 0 {
					return
				}

				current := selectedIndex.Get().(int)
				if current == -1 {
					// No selection, select last item
					selectItem(len(items) - 1)
				} else if current > 0 {
					// Move up
					selectItem(current - 1)
				}
			})

			// Keyboard navigation: Select current item
			ctx.On("keyEnter", func(_ interface{}) {
				items := props.Items.Get().([]T)
				current := selectedIndex.Get().(int)

				if current >= 0 && current < len(items) && props.OnSelect != nil {
					props.OnSelect(items[current], current)
				}
			})

			// Keyboard navigation: Jump to first
			ctx.On("keyHome", func(_ interface{}) {
				items := props.Items.Get().([]T)
				if len(items) > 0 {
					selectItem(0)
				}
			})

			// Keyboard navigation: Jump to last
			ctx.On("keyEnd", func(_ interface{}) {
				items := props.Items.Get().([]T)
				if len(items) > 0 {
					selectItem(len(items) - 1)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(ListProps[T])

			// Type assert the exposed values
			selectedIndex := ctx.Get("selectedIndex").(*bubbly.Ref[int]).Get().(int)
			scrollOffset := ctx.Get("scrollOffset").(*bubbly.Ref[int]).Get().(int)
			theme := ctx.Get("theme").(Theme)

			items := p.Items.Get().([]T)

			// Handle empty list
			if len(items) == 0 {
				emptyStyle := lipgloss.NewStyle().
					Foreground(theme.Muted).
					Italic(true).
					Padding(1, 2)
				return emptyStyle.Render("No items to display")
			}

			// Determine visible range
			height := p.Height
			if height <= 0 {
				height = 10
			}

			var visibleItems []T
			var startIndex int

			if p.Virtual {
				// Virtual scrolling: only render visible items
				offset := scrollOffset
				if offset < 0 {
					offset = 0
				}
				if offset > len(items)-1 {
					offset = len(items) - 1
				}

				endIndex := offset + height
				if endIndex > len(items) {
					endIndex = len(items)
				}

				visibleItems = items[offset:endIndex]
				startIndex = offset
			} else {
				// No virtual scrolling: render all items (may be slow for large lists)
				visibleItems = items
				startIndex = 0
			}

			// Render items
			var output strings.Builder

			for i, item := range visibleItems {
				actualIndex := startIndex + i
				itemText := p.RenderItem(item, actualIndex)

				// Style based on selection
				var itemStyle lipgloss.Style
				if actualIndex == selectedIndex {
					// Selected item: highlighted with primary color
					itemStyle = lipgloss.NewStyle().
						Foreground(lipgloss.Color("230")).
						Background(theme.Primary).
						Bold(true).
						Padding(0, 1)
				} else {
					// Normal item
					itemStyle = lipgloss.NewStyle().
						Foreground(theme.Foreground).
						Padding(0, 1)
				}

				output.WriteString(itemStyle.Render(itemText))
				output.WriteString("\n")
			}

			// Add scroll indicators if using virtual scrolling
			if p.Virtual && len(items) > height {
				indicatorStyle := lipgloss.NewStyle().
					Foreground(theme.Muted).
					Italic(true)

				if scrollOffset > 0 {
					output.WriteString(indicatorStyle.Render("↑ More items above"))
					output.WriteString("\n")
				}

				if scrollOffset+height < len(items) {
					output.WriteString(indicatorStyle.Render("↓ More items below"))
					output.WriteString("\n")
				}
			}

			// Apply custom style if provided
			result := output.String()
			if p.Style != nil {
				result = p.Style.Render(result)
			}

			return result
		}).
		Build()

	if err != nil {
		panic(err) // Should never happen with valid setup
	}

	return comp
}
