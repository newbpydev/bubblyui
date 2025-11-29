// Package demos provides demo views for each composable.
package demos

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/17-enhanced-composables/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CreateUseToggleDemo creates the UseToggle demo view.
func CreateUseToggleDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("UseToggleDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			// Use shared toggle state
			toggle1 := state.ToggleDemo1.GetTyped()
			toggle2 := state.ToggleDemo2.GetTyped()
			toggle3 := state.ToggleDemo3.GetTyped()

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `toggle := composables.UseToggle(ctx, false)
toggle.Toggle()      // Flip the value
toggle.SetOn()       // Set to true
toggle.SetOff()      // Set to false
isOn := toggle.Value.GetTyped()`

			// Toggle indicators
			onStyle := lipgloss.NewStyle().Foreground(theme.Success).Bold(true)
			offStyle := lipgloss.NewStyle().Foreground(theme.Muted)

			renderToggle := func(num int, value bool) string {
				indicator := offStyle.Render("○ OFF")
				if value {
					indicator = onStyle.Render("● ON")
				}
				return fmt.Sprintf("  [%d] Toggle %d: %s", num, num, indicator)
			}

			stateContent := fmt.Sprintf(
				"Interactive Toggles:\n%s\n%s\n%s\n\nPress 1/2/3 to toggle each\n\nUse Cases:\n  • Feature flags\n  • Settings switches\n  • Checkboxes",
				renderToggle(1, toggle1),
				renderToggle(2, toggle2),
				renderToggle(3, toggle3),
			)

			stateCard := components.Card(components.CardProps{
				Title:   "Interactive Toggle Demo",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "UseToggle provides a simple boolean toggle with convenient methods. Perfect for on/off switches, checkboxes, and feature flags."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("UseToggle Demo"),
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

// CreateUseCounterDemo creates the UseCounter demo view.
func CreateUseCounterDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("UseCounterDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			// Use shared local counter
			localVal := state.LocalCounter.GetTyped()
			minVal := state.LocalCounterMin
			maxVal := state.LocalCounterMax

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `counter := composables.UseCounter(ctx, 0,
    composables.WithMin(0),
    composables.WithMax(100),
    composables.WithStep(5),
)
counter.Increment()  // Add step
counter.Decrement()  // Subtract step
counter.Set(50)      // Set directly
counter.Reset()      // Reset to initial`

			// Progress bar
			barWidth := 30
			filled := int(float64(localVal) / float64(maxVal) * float64(barWidth))
			var bar strings.Builder
			for i := 0; i < barWidth; i++ {
				if i < filled {
					bar.WriteString(lipgloss.NewStyle().Foreground(theme.Primary).Render("█"))
				} else {
					bar.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("░"))
				}
			}

			stateContent := fmt.Sprintf(
				"Counter Value: %d\n%s\n\nBounds: [%d, %d] | Step: 5\n\nPress +/- to change, r to reset",
				localVal, bar.String(), minVal, maxVal,
			)

			stateCard := components.Card(components.CardProps{
				Title:   "Interactive Counter Demo",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "UseCounter provides a bounded counter with configurable min, max, and step. Great for numeric inputs, pagination, and volume controls."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("UseCounter Demo"),
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

// CreateUsePreviousDemo creates the UsePrevious demo view.
func CreateUsePreviousDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("UsePreviousDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			counterVal := state.CounterValue.GetTyped()
			prevVal := state.PreviousVal.GetTyped()

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `count := bubbly.NewRef(0)
previous := composables.UsePrevious(ctx, count)

// After count.Set(5):
current := count.GetTyped()        // 5
prev := previous.Value.GetTyped()  // 0 (previous value)`

			prevStr := "none"
			if prevVal != nil {
				prevStr = fmt.Sprintf("%d", *prevVal)
			}

			// Show change direction
			changeIndicator := "→"
			changeStyle := lipgloss.NewStyle().Foreground(theme.Muted)
			if prevVal != nil {
				if counterVal > *prevVal {
					changeIndicator = "↑"
					changeStyle = lipgloss.NewStyle().Foreground(theme.Success)
				} else if counterVal < *prevVal {
					changeIndicator = "↓"
					changeStyle = lipgloss.NewStyle().Foreground(theme.Error)
				}
			}

			stateContent := fmt.Sprintf(
				"Current Value: %d\nPrevious Value: %s\n\nChange: %s %s\n\nUseful for:\n  • Showing deltas\n  • Detecting direction\n  • Animation transitions\n\n+/-: change counter to see previous",
				counterVal, prevStr,
				changeStyle.Render(changeIndicator),
				changeStyle.Render(fmt.Sprintf("(%s → %d)", prevStr, counterVal)),
			)

			stateCard := components.Card(components.CardProps{
				Title:   "Previous State",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "UsePrevious tracks the previous value of a ref. Essential for detecting changes, showing deltas, and implementing transitions."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("UsePrevious Demo"),
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

// CreateUseHistoryDemo creates the UseHistory demo view.
func CreateUseHistoryDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("UseHistoryDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			counterVal := state.CounterValue.GetTyped()
			canUndo := state.CanUndo.GetTyped()
			canRedo := state.CanRedo.GetTyped()

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `history := composables.UseHistory(ctx, initialValue, maxSize)
history.Push(newValue)  // Add to history
history.Undo()          // Go back
history.Redo()          // Go forward
history.Clear()         // Clear history
canUndo := history.CanUndo.Get().(bool)
canRedo := history.CanRedo.Get().(bool)`

			// Undo/Redo indicators
			undoStyle := lipgloss.NewStyle().Foreground(theme.Muted)
			redoStyle := lipgloss.NewStyle().Foreground(theme.Muted)
			if canUndo {
				undoStyle = lipgloss.NewStyle().Foreground(theme.Success).Bold(true)
			}
			if canRedo {
				redoStyle = lipgloss.NewStyle().Foreground(theme.Success).Bold(true)
			}

			stateContent := fmt.Sprintf(
				"Current Value: %d\n\nHistory Controls:\n  %s Undo (u): %t\n  %s Redo (r): %t\n\nMax History Size: 20 entries\n\nUse +/- to change counter,\nthen u/r to undo/redo",
				counterVal,
				undoStyle.Render("←"), canUndo,
				redoStyle.Render("→"), canRedo,
			)

			stateCard := components.Card(components.CardProps{
				Title:   "History State",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "UseHistory provides undo/redo functionality with configurable history size. Perfect for text editors, form changes, and any reversible actions."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("UseHistory Demo"),
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
