// Package components provides demo components for the layout showcase.
package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/14-advanced-layouts/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// FlexDemoProps defines props for the flex demo component.
type FlexDemoProps struct{}

// CreateFlexDemo creates an interactive Flex layout demonstration.
// This showcases all justify and align options with live preview.
func CreateFlexDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("FlexDemo").
		Setup(func(ctx *bubbly.Context) {
			// Provide theme
			ctx.ProvideTheme(bubbly.DefaultTheme)
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)

			// Get shared demo state
			demoState := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("demoState", demoState)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			demoState := ctx.Get("demoState").(*localComposables.DemoStateComposable)

			// Get current settings
			justifyIdx := demoState.JustifyIndex.Get().(int)
			alignIdx := demoState.AlignIndex.Get().(int)
			direction := demoState.FlexDirection.Get().(string)
			wrapEnabled := demoState.WrapEnabled.Get().(bool)
			gapSize := demoState.GapSize.Get().(int)

			currentJustify := localComposables.JustifyOptions[justifyIdx]
			currentAlign := localComposables.AlignOptions[alignIdx]

			// === TITLE ===
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary)
			title := components.Text(components.TextProps{
				Content: titleStyle.Render("🎯 Flex Layout Demo"),
			})
			title.Init()

			// === SETTINGS DISPLAY ===
			settingsStyle := lipgloss.NewStyle().Foreground(theme.Secondary)
			activeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("35"))

			settingsContent := fmt.Sprintf(
				"%s: %s  |  %s: %s  |  %s: %s  |  %s: %v  |  %s: %d",
				settingsStyle.Render("Justify"),
				activeStyle.Render(string(currentJustify)),
				settingsStyle.Render("Align"),
				activeStyle.Render(string(currentAlign)),
				settingsStyle.Render("Direction"),
				activeStyle.Render(direction),
				settingsStyle.Render("Wrap"),
				wrapEnabled,
				settingsStyle.Render("Gap"),
				gapSize,
			)
			settings := components.Text(components.TextProps{
				Content: settingsContent,
			})
			settings.Init()

			// === DEMO ITEMS ===
			// Create colored boxes of different sizes to show alignment
			colors := []lipgloss.Color{
				lipgloss.Color("35"),  // Green
				lipgloss.Color("99"),  // Purple
				lipgloss.Color("205"), // Pink
				lipgloss.Color("42"),  // Cyan
			}
			sizes := []struct{ w, h int }{
				{8, 2},
				{10, 3},
				{6, 1},
				{12, 2},
			}

			var demoItems []bubbly.Component
			for i := 0; i < 4; i++ {
				boxStyle := lipgloss.NewStyle().
					Background(colors[i]).
					Foreground(lipgloss.Color("0")).
					Bold(true).
					Padding(0, 1)

				content := fmt.Sprintf("Item %d", i+1)
				box := components.Box(components.BoxProps{
					Content: boxStyle.Render(content),
					Width:   sizes[i].w,
					Height:  sizes[i].h,
				})
				box.Init()
				demoItems = append(demoItems, box)
			}

			// === FLEX CONTAINER ===
			flexProps := components.FlexProps{
				Items:     demoItems,
				Direction: components.FlexDirection(direction),
				Justify:   currentJustify,
				Align:     currentAlign,
				Gap:       gapSize,
				Wrap:      wrapEnabled,
			}

			// Set container size based on direction
			if direction == string(components.FlexRow) {
				flexProps.Width = 70
				if wrapEnabled {
					flexProps.Width = 40 // Smaller to force wrapping
				}
			} else {
				flexProps.Height = 15
				if wrapEnabled {
					flexProps.Height = 8 // Smaller to force wrapping
				}
			}

			flex := components.Flex(flexProps)
			flex.Init()

			// Wrap in a box to show boundaries
			flexBox := components.Box(components.BoxProps{
				Child:       flex,
				Border:      true,
				BorderStyle: lipgloss.RoundedBorder(),
				Padding:     1,
				Title:       "Flex Container",
			})
			flexBox.Init()

			// === CONTROLS HELP ===
			helpStyle := lipgloss.NewStyle().Foreground(theme.Muted).Italic(true)
			helpText := helpStyle.Render("j/k: justify • a/s: align • d: direction • w: wrap • +/-: gap")
			help := components.Text(components.TextProps{
				Content: helpText,
			})
			help.Init()

			// === JUSTIFY OPTIONS DISPLAY ===
			var justifyItems []interface{}
			for i, opt := range localComposables.JustifyOptions {
				style := lipgloss.NewStyle().Foreground(theme.Muted)
				if i == justifyIdx {
					style = lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
				}
				text := components.Text(components.TextProps{
					Content: style.Render(string(opt)),
				})
				text.Init()
				justifyItems = append(justifyItems, text)
			}

			justifyRow := components.HStack(components.StackProps{
				Items:   justifyItems,
				Spacing: 2,
			})
			justifyRow.Init()

			justifyLabel := components.Text(components.TextProps{
				Content: lipgloss.NewStyle().Bold(true).Render("Justify Options:"),
			})
			justifyLabel.Init()

			// === ALIGN OPTIONS DISPLAY ===
			var alignItems []interface{}
			for i, opt := range localComposables.AlignOptions {
				style := lipgloss.NewStyle().Foreground(theme.Muted)
				if i == alignIdx {
					style = lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
				}
				text := components.Text(components.TextProps{
					Content: style.Render(string(opt)),
				})
				text.Init()
				alignItems = append(alignItems, text)
			}

			alignRow := components.HStack(components.StackProps{
				Items:   alignItems,
				Spacing: 2,
			})
			alignRow.Init()

			alignLabel := components.Text(components.TextProps{
				Content: lipgloss.NewStyle().Bold(true).Render("Align Options:"),
			})
			alignLabel.Init()

			// === LAYOUT ===
			divider := components.Divider(components.DividerProps{
				Length: 70,
			})
			divider.Init()

			page := components.VStack(components.StackProps{
				Items: []interface{}{
					title,
					settings,
					divider,
					flexBox,
					divider,
					justifyLabel,
					justifyRow,
					alignLabel,
					alignRow,
					help,
				},
				Spacing: 1,
			})
			page.Init()

			return page.View()
		}).
		Build()
}
