// Package demos provides demo views for each composable.
package demos

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/17-enhanced-composables/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CreateUseWindowSizeDemo creates the UseWindowSize demo view.
func CreateUseWindowSizeDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("UseWindowSizeDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)

			// Create a local UseWindowSize instance for demo
			windowSize := composables.UseWindowSize(ctx)
			ctx.Expose("windowSize", windowSize)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)
			windowSize := ctx.Get("windowSize").(*composables.WindowSizeReturn)

			width := state.Width.GetTyped()
			height := state.Height.GetTyped()

			// Get breakpoint info
			breakpoint := windowSize.Breakpoint.GetTyped()
			sidebarVisible := windowSize.SidebarVisible.GetTyped()
			gridColumns := windowSize.GridColumns.GetTyped()

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			// Usage example
			usage := `windowSize := composables.UseWindowSize(ctx)
width := windowSize.Width.GetTyped()
height := windowSize.Height.GetTyped()
breakpoint := windowSize.Breakpoint.GetTyped()`

			// Current values
			valuesContent := fmt.Sprintf(
				"Width: %d | Height: %d\nBreakpoint: %s\n\nDerived Values:\n  Sidebar Visible: %t\n  Grid Columns: %d\n  Content Width: %d\n  Card Width: %d",
				width, height, breakpoint,
				sidebarVisible, gridColumns,
				windowSize.GetContentWidth(),
				windowSize.GetCardWidth(),
			)

			valuesCard := components.Card(components.CardProps{
				Title:   "Current Values",
				Content: valuesContent,
				Width:   40,
			})
			valuesCard.Init()

			descContent := "UseWindowSize tracks terminal dimensions and provides responsive breakpoints. Resize your terminal to see values update in real-time."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("UseWindowSize Demo"),
				"",
				codeStyle.Render(usage),
				"",
				valuesCard.View(),
				"",
				descCard.View(),
			)
		}).
		Build()
}

// CreateUseFocusDemo creates the UseFocus demo view.
func CreateUseFocusDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("UseFocusDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			// Use demo-specific focus state
			focusIndex := state.FocusDemoIndex.GetTyped()

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `focus := composables.UseFocus(ctx, FocusMain, []FocusPane{
    FocusSidebar, FocusMain,
})
focus.Next()  // Cycle to next pane
focus.Prev()  // Cycle to previous pane
focus.Set(FocusSidebar)  // Set directly`

			// Render 3 panes with focus indicator
			paneNames := []string{"Header", "Content", "Footer"}
			var panes strings.Builder
			for i, name := range paneNames {
				style := lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					Width(36).
					Padding(0, 1)
				if i == focusIndex {
					style = style.BorderForeground(theme.Primary).Bold(true)
					panes.WriteString(style.Render(fmt.Sprintf("► %s (FOCUSED)", name)))
				} else {
					style = style.BorderForeground(lipgloss.Color("240"))
					panes.WriteString(style.Render(fmt.Sprintf("  %s", name)))
				}
				panes.WriteString("\n")
			}

			stateContent := fmt.Sprintf(
				"Focused Pane: %s (index: %d)\n\n%s\nPress 1/2/3 to focus pane\nPress TAB to cycle",
				paneNames[focusIndex], focusIndex, panes.String(),
			)

			stateCard := components.Card(components.CardProps{
				Title:   "Interactive Focus Demo",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "UseFocus manages multi-pane focus with keyboard navigation. It tracks which UI section is active and provides methods to cycle through focusable areas."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("UseFocus Demo"),
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

// CreateUseScrollDemo creates the UseScroll demo view.
func CreateUseScrollDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("UseScrollDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			// Use shared scroll state
			offset := state.ScrollDemoOffset.GetTyped()
			total := state.ScrollDemoTotal
			visible := 5
			maxOffset := total - visible
			atTop := offset == 0
			atBottom := offset >= maxOffset
			progress := 0.0
			if maxOffset > 0 {
				progress = float64(offset) / float64(maxOffset)
			}

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `scroll := composables.UseScroll(ctx, 50, 10)
scroll.ScrollDown()  // Move down
scroll.ScrollUp()    // Move up
scroll.ScrollTo(25)  // Jump to position
progress := scroll.Progress.GetTyped()`

			// Render visible items
			var items strings.Builder
			for i := offset; i < offset+visible && i < total; i++ {
				items.WriteString(fmt.Sprintf("  Item %d\n", i+1))
			}

			// Visual scrollbar
			barHeight := 8
			thumbPos := int(progress * float64(barHeight-1))
			var scrollbar strings.Builder
			for i := 0; i < barHeight; i++ {
				if i == thumbPos {
					scrollbar.WriteString("█")
				} else {
					scrollbar.WriteString("░")
				}
			}

			stateContent := fmt.Sprintf(
				"Offset: %d / %d | Progress: %.0f%%\nAt Top: %t | At Bottom: %t\n\nVisible Items:\n%s\nScrollbar: %s\n\nPress j/k to scroll, g/G for top/bottom",
				offset, total, progress*100, atTop, atBottom, items.String(), scrollbar.String(),
			)

			stateCard := components.Card(components.CardProps{
				Title:   "Interactive Scroll Demo",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "UseScroll manages viewport scrolling for lists and content. It tracks scroll position, calculates progress, and detects boundary conditions."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("UseScroll Demo"),
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

// CreateUseSelectionDemo creates the UseSelection demo view.
func CreateUseSelectionDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("UseSelectionDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			// Use shared selection state
			selectedIdx := state.SelectionDemoIndex.GetTyped()
			selectedItems := state.SelectionDemoItems.GetTyped()
			items := []string{"Apple", "Banana", "Cherry", "Date", "Elderberry"}

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `items := []string{"Apple", "Banana", "Cherry"}
selection := composables.UseSelection(ctx, items,
    composables.WithWrap(true),
)
selection.Next()  // Select next item
selection.Prev()  // Select previous item
item := selection.SelectedItem.GetTyped()`

			// Render item list with selection checkboxes
			var listContent strings.Builder
			selectedCount := 0
			for i, item := range items {
				cursor := "  "
				checkbox := "[ ]"
				style := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

				if i == selectedIdx {
					cursor = "▶ "
					style = style.Bold(true).Foreground(theme.Primary)
				}
				if i < len(selectedItems) && selectedItems[i] {
					checkbox = "[✓]"
					selectedCount++
				}
				listContent.WriteString(style.Render(fmt.Sprintf("%s%s %s", cursor, checkbox, item)))
				listContent.WriteString("\n")
			}

			stateContent := fmt.Sprintf(
				"Cursor: %d (%s)\nSelected: %d items\n\nItems:\n%s\nPress j/k to move, SPACE to toggle",
				selectedIdx, items[selectedIdx], selectedCount, listContent.String(),
			)

			stateCard := components.Card(components.CardProps{
				Title:   "Interactive Selection Demo",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "UseSelection manages list/table selection with keyboard navigation. Supports wrapping, multi-select, and provides the selected item directly."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("UseSelection Demo"),
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

// CreateUseModeDemo creates the UseMode demo view.
func CreateUseModeDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("UseModeDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			// Use shared mode state
			currentMode := state.ModeDemoMode.GetTyped()
			modes := state.ModeDemoModes

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `mode := composables.UseMode(ctx, "navigation")
mode.Set("input")     // Change mode
mode.Toggle("input")  // Toggle between modes
current := mode.Current.GetTyped()
prev := mode.Previous.GetTyped()`

			// Mode colors
			modeColors := map[string]lipgloss.Color{
				"normal":  lipgloss.Color("99"), // Purple
				"insert":  theme.Success,        // Green
				"visual":  theme.Warning,        // Yellow
				"command": theme.Error,          // Red
			}

			// Render mode buttons
			var modeButtons strings.Builder
			for i, mode := range modes {
				style := lipgloss.NewStyle().
					Padding(0, 2).
					MarginRight(1)
				if mode == currentMode {
					style = style.Bold(true).
						Foreground(lipgloss.Color("0")).
						Background(modeColors[mode])
				} else {
					style = style.Foreground(lipgloss.Color("240"))
				}
				modeButtons.WriteString(style.Render(fmt.Sprintf("%d:%s", i+1, mode)))
				modeButtons.WriteString(" ")
			}

			stateContent := fmt.Sprintf(
				"Current Mode: %s\n\nModes:\n%s\n\nPress 1/2/3/4 to switch modes\n\nUse Cases:\n  • Vim-like editing\n  • Modal dialogs\n  • Input vs navigation",
				strings.ToUpper(currentMode), modeButtons.String(),
			)

			stateCard := components.Card(components.CardProps{
				Title:   "Interactive Mode Demo",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "UseMode manages application modes (navigation/input/edit). Essential for TUI apps where single keys have different meanings based on context."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("UseMode Demo"),
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
