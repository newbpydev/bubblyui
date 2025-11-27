// Package main provides the Advanced Layout System example application.
package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	localComponents "github.com/newbpydev/bubblyui/cmd/examples/14-advanced-layouts/components"
	localComposables "github.com/newbpydev/bubblyui/cmd/examples/14-advanced-layouts/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CreateApp creates the root application component for the layout showcase.
func CreateApp() (bubbly.Component, error) {
	return bubbly.NewComponent("LayoutShowcase").
		WithAutoCommands(true).
		// Navigation between demos
		WithKeyBinding("1", "demo1", "Dashboard").
		WithKeyBinding("2", "demo2", "Flex Layout").
		WithKeyBinding("3", "demo3", "Card Grid").
		WithKeyBinding("4", "demo4", "Form Layout").
		WithKeyBinding("5", "demo5", "Modal/Dialog").
		WithMultiKeyBindings("nextDemo", "Next demo", "tab", "right", "l").
		WithMultiKeyBindings("prevDemo", "Previous demo", "shift+tab", "left", "h").
		// Flex demo controls
		WithMultiKeyBindings("nextJustify", "Next justify", "j").
		WithMultiKeyBindings("prevJustify", "Prev justify", "shift+j").
		WithMultiKeyBindings("nextAlign", "Next align", "a").
		WithMultiKeyBindings("prevAlign", "Prev align", "shift+a").
		WithKeyBinding("d", "toggleDirection", "Toggle direction").
		WithKeyBinding("w", "toggleWrap", "Toggle wrap").
		WithKeyBinding("+", "increaseGap", "Increase gap").
		WithKeyBinding("-", "decreaseGap", "Decrease gap").
		// Modal demo controls
		WithKeyBinding("m", "toggleModal", "Toggle modal").
		// Quit
		WithKeyBinding("q", "quit", "Quit").
		WithKeyBinding("ctrl+c", "quit", "Quit").
		Setup(func(ctx *bubbly.Context) {
			// Provide theme for all descendants
			ctx.ProvideTheme(bubbly.DefaultTheme)
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)

			// Get shared demo state
			demoState := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("demoState", demoState)

			// Create demo components
			dashboardDemo, err := localComponents.CreateDashboardDemo()
			if err != nil {
				panic(fmt.Sprintf("Failed to create dashboard demo: %v", err))
			}

			flexDemo, err := localComponents.CreateFlexDemo()
			if err != nil {
				panic(fmt.Sprintf("Failed to create flex demo: %v", err))
			}

			cardGridDemo, err := localComponents.CreateCardGridDemo()
			if err != nil {
				panic(fmt.Sprintf("Failed to create card grid demo: %v", err))
			}

			formDemo, err := localComponents.CreateFormDemo()
			if err != nil {
				panic(fmt.Sprintf("Failed to create form demo: %v", err))
			}

			modalDemo, err := localComponents.CreateModalDemo()
			if err != nil {
				panic(fmt.Sprintf("Failed to create modal demo: %v", err))
			}

			// Expose components
			if err := ctx.ExposeComponent("dashboardDemo", dashboardDemo); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose dashboard: %v", err))
			}
			if err := ctx.ExposeComponent("flexDemo", flexDemo); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose flex: %v", err))
			}
			if err := ctx.ExposeComponent("cardGridDemo", cardGridDemo); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose card grid: %v", err))
			}
			if err := ctx.ExposeComponent("formDemo", formDemo); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose form: %v", err))
			}
			if err := ctx.ExposeComponent("modalDemo", modalDemo); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose modal: %v", err))
			}

			// Event handlers for navigation
			ctx.On("demo1", func(_ interface{}) {
				demoState.SetDemo(localComposables.DemoDashboard)
			})
			ctx.On("demo2", func(_ interface{}) {
				demoState.SetDemo(localComposables.DemoFlex)
			})
			ctx.On("demo3", func(_ interface{}) {
				demoState.SetDemo(localComposables.DemoCardGrid)
			})
			ctx.On("demo4", func(_ interface{}) {
				demoState.SetDemo(localComposables.DemoForm)
			})
			ctx.On("demo5", func(_ interface{}) {
				demoState.SetDemo(localComposables.DemoModal)
			})
			ctx.On("nextDemo", func(_ interface{}) {
				demoState.NextDemo()
			})
			ctx.On("prevDemo", func(_ interface{}) {
				demoState.PrevDemo()
			})

			// Flex demo controls
			ctx.On("nextJustify", func(_ interface{}) {
				demoState.NextJustify()
			})
			ctx.On("prevJustify", func(_ interface{}) {
				demoState.PrevJustify()
			})
			ctx.On("nextAlign", func(_ interface{}) {
				demoState.NextAlign()
			})
			ctx.On("prevAlign", func(_ interface{}) {
				demoState.PrevAlign()
			})
			ctx.On("toggleDirection", func(_ interface{}) {
				demoState.ToggleDirection()
			})
			ctx.On("toggleWrap", func(_ interface{}) {
				demoState.ToggleWrap()
			})
			ctx.On("increaseGap", func(_ interface{}) {
				demoState.IncreaseGap()
			})
			ctx.On("decreaseGap", func(_ interface{}) {
				demoState.DecreaseGap()
			})

			// Modal demo controls
			ctx.On("toggleModal", func(_ interface{}) {
				demoState.ToggleModal()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			demoState := ctx.Get("demoState").(*localComposables.DemoStateComposable)

			currentDemo := localComposables.DemoType(demoState.CurrentDemo.Get().(int))

			// === HEADER ===
			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				Background(lipgloss.Color("236")).
				Padding(0, 2)

			headerText := components.Text(components.TextProps{
				Content: headerStyle.Render("🎨 BubblyUI Advanced Layout System Showcase"),
			})
			headerText.Init()

			// === TAB BAR ===
			var tabItems []interface{}
			for i := 0; i < 5; i++ {
				demoType := localComposables.DemoType(i)
				name := localComposables.DemoNames[demoType]

				style := lipgloss.NewStyle().Padding(0, 1)
				if demoType == currentDemo {
					style = style.
						Bold(true).
						Foreground(lipgloss.Color("0")).
						Background(theme.Primary)
				} else {
					style = style.Foreground(theme.Muted)
				}

				tabText := components.Text(components.TextProps{
					Content: style.Render(fmt.Sprintf("%d. %s", i+1, name)),
				})
				tabText.Init()
				tabItems = append(tabItems, tabText)
			}

			tabBar := components.HStack(components.StackProps{
				Items:   tabItems,
				Spacing: 1,
			})
			tabBar.Init()

			// === DEMO CONTENT ===
			var demoContent bubbly.Component
			switch currentDemo {
			case localComposables.DemoDashboard:
				demoContent = ctx.Get("dashboardDemo").(bubbly.Component)
			case localComposables.DemoFlex:
				demoContent = ctx.Get("flexDemo").(bubbly.Component)
			case localComposables.DemoCardGrid:
				demoContent = ctx.Get("cardGridDemo").(bubbly.Component)
			case localComposables.DemoForm:
				demoContent = ctx.Get("formDemo").(bubbly.Component)
			case localComposables.DemoModal:
				demoContent = ctx.Get("modalDemo").(bubbly.Component)
			default:
				demoContent = ctx.Get("dashboardDemo").(bubbly.Component)
			}

			// === FOOTER ===
			footerStyle := lipgloss.NewStyle().
				Foreground(theme.Muted).
				Italic(true)

			var footerText string
			switch currentDemo {
			case localComposables.DemoFlex:
				footerText = "j/J: justify • a/A: align • d: direction • w: wrap • +/-: gap • 1-5: switch demo • q: quit"
			case localComposables.DemoModal:
				footerText = "m: toggle modal • 1-5: switch demo • tab/shift+tab: next/prev • q: quit"
			default:
				footerText = "1-5: switch demo • tab/shift+tab: next/prev • q: quit"
			}

			footer := components.Text(components.TextProps{
				Content: footerStyle.Render(footerText),
			})
			footer.Init()

			// === DIVIDERS ===
			headerDivider := components.Divider(components.DividerProps{
				Length: 80,
				Char:   "═",
			})
			headerDivider.Init()

			footerDivider := components.Divider(components.DividerProps{
				Length: 80,
			})
			footerDivider.Init()

			// === MAIN LAYOUT ===
			page := components.VStack(components.StackProps{
				Items: []interface{}{
					headerText,
					tabBar,
					headerDivider,
					demoContent,
					footerDivider,
					footer,
				},
				Spacing: 0,
			})
			page.Init()

			return page.View()
		}).
		Build()
}
