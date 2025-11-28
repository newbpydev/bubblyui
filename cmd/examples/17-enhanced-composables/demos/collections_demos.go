// Package demos provides demo views for each composable.
package demos

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/17-enhanced-composables/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CreateUseListDemo creates the UseList demo view.
func CreateUseListDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("UseListDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			// Use shared list state
			items := state.ListDemoItems.GetTyped()
			selectedIdx := state.ListDemoIndex.GetTyped()

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `list := composables.UseList(ctx, []string{"a", "b"})
list.Append("c")           // Add to end
list.Prepend("z")          // Add to start
list.Insert(1, "x")        // Insert at index
list.Remove(0)             // Remove at index
list.Set(0, "new")         // Update at index
list.Clear()               // Remove all
items := list.Items.GetTyped()`

			// Render items with selection
			var listContent strings.Builder
			for i, item := range items {
				cursor := "  "
				style := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
				if i == selectedIdx {
					cursor = "▶ "
					style = style.Bold(true).Foreground(theme.Primary)
				}
				listContent.WriteString(style.Render(fmt.Sprintf("%s%d. %s", cursor, i+1, item)))
				listContent.WriteString("\n")
			}
			if len(items) == 0 {
				listContent.WriteString("  (empty list)\n")
			}

			stateContent := fmt.Sprintf(
				"List (%d items):\n%s\nPress j/k: move | a: add | d: delete | c: clear",
				len(items), listContent.String(),
			)

			stateCard := components.Card(components.CardProps{
				Title:   "Interactive List Demo",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "UseList provides a reactive list with CRUD operations. Supports append, prepend, insert, remove, and update operations."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("UseList Demo"),
				"",
				codeStyle.Render(usage),
				"",
				stateCard.View(),
				"",
				descCard.View(),
			)
		}).
		Build()
}

// CreateUseMapDemo creates the UseMap demo view.
func CreateUseMapDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("UseMapDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			// Use shared map state
			items := state.MapDemoData.GetTyped()
			size := len(items)

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `m := composables.UseMap(ctx, map[string]int{
    "apples": 5,
})
m.Set("bananas", 3)        // Add/update key
value := m.Get("apples")   // Get value
m.Delete("apples")         // Remove key
has := m.Has("bananas")    // Check existence
m.Clear()                  // Remove all
size := m.Size.GetTyped()  // Get count`

			// Render map entries (sorted for consistent display)
			var keys []string
			for k := range items {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			var mapContent strings.Builder
			for _, k := range keys {
				mapContent.WriteString(fmt.Sprintf("  %s: %d\n", k, items[k]))
			}
			if size == 0 {
				mapContent.WriteString("  (empty map)\n")
			}

			stateContent := fmt.Sprintf(
				"Map Contents (%d entries):\n%s\nPress a: add | d: delete | c: clear",
				size, mapContent.String(),
			)

			stateCard := components.Card(components.CardProps{
				Title:   "Interactive Map Demo",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "UseMap provides a reactive key-value store. Supports get, set, delete, has, and clear operations with reactive size tracking."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("UseMap Demo"),
				"",
				codeStyle.Render(usage),
				"",
				stateCard.View(),
				"",
				descCard.View(),
			)
		}).
		Build()
}

// CreateUseSetDemo creates the UseSet demo view.
func CreateUseSetDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("UseSetDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			// Use shared set state
			items := state.SetDemoItems.GetTyped()
			sortedItems := make([]string, len(items))
			copy(sortedItems, items)
			sort.Strings(sortedItems)
			size := len(items)

			// Check if "bubbly" is in set
			hasBubbly := false
			for _, item := range items {
				if item == "bubbly" {
					hasBubbly = true
					break
				}
			}

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `set := composables.UseSet(ctx, []string{"a", "b"})
set.Add("c")               // Add item (no duplicates)
set.Delete("a")            // Remove item
has := set.Has("b")        // Check membership
set.Toggle("c")            // Add if missing, remove if present
set.Clear()                // Remove all
slice := set.ToSlice()     // Get as slice
size := set.Size.GetTyped()`

			bubblyStatus := "not in set"
			if hasBubbly {
				bubblyStatus = "IN SET ✓"
			}

			stateContent := fmt.Sprintf(
				"Set (%d items):\n  [%s]\n\n'bubbly': %s\n\nPress a: add | d: delete | t: toggle 'bubbly' | c: clear",
				size, strings.Join(sortedItems, ", "), bubblyStatus,
			)

			stateCard := components.Card(components.CardProps{
				Title:   "Set State",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "UseSet provides a reactive set (unique values only). Supports add, delete, has, toggle, and clear operations."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("UseSet Demo"),
				"",
				codeStyle.Render(usage),
				"",
				stateCard.View(),
				"",
				descCard.View(),
			)
		}).
		Build()
}

// CreateUseQueueDemo creates the UseQueue demo view.
func CreateUseQueueDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("UseQueueDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			// Use shared queue state
			items := state.QueueDemoItems.GetTyped()
			size := len(items)
			isEmpty := size == 0

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `queue := composables.UseQueue(ctx, []string{"a", "b"})
queue.Enqueue("c")         // Add to back
item := queue.Dequeue()    // Remove from front
front := queue.Peek()      // View front without removing
queue.Clear()              // Remove all
isEmpty := queue.IsEmpty.GetTyped()
size := queue.Size.GetTyped()`

			// Visual queue representation
			var queueVisual strings.Builder
			if isEmpty {
				queueVisual.WriteString("  (empty queue)")
			} else {
				queueVisual.WriteString("  Front → ")
				for i, item := range items {
					if i > 0 {
						queueVisual.WriteString(" → ")
					}
					queueVisual.WriteString(fmt.Sprintf("[%s]", item))
				}
				queueVisual.WriteString(" ← Back")
			}

			stateContent := fmt.Sprintf(
				"Queue (%d items):\n%s\n\nIs Empty: %t\n\nPress e: enqueue | d: dequeue | c: clear",
				size, queueVisual.String(), isEmpty,
			)

			stateCard := components.Card(components.CardProps{
				Title:   "Interactive Queue Demo",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "UseQueue provides a FIFO (First In, First Out) queue. Perfect for task queues, message buffers, and ordered processing."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("UseQueue Demo"),
				"",
				codeStyle.Render(usage),
				"",
				stateCard.View(),
				"",
				descCard.View(),
			)
		}).
		Build()
}
