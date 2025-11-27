// Package main provides the Responsive Layouts example application.
package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	localComponents "github.com/newbpydev/bubblyui/cmd/examples/15-responsive-layouts/components"
	localComposables "github.com/newbpydev/bubblyui/cmd/examples/15-responsive-layouts/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// DemoType represents the current demo being displayed.
type DemoType int

const (
	// DemoDashboard shows the responsive dashboard.
	DemoDashboard DemoType = iota
	// DemoGrid shows the responsive grid.
	DemoGrid
	// DemoAdaptive shows the adaptive content layout.
	DemoAdaptive
	// DemoBreakpoint shows breakpoint information.
	DemoBreakpoint
)

// DemoNames maps demo types to display names.
var DemoNames = map[DemoType]string{
	DemoDashboard:  "Dashboard",
	DemoGrid:       "Grid",
	DemoAdaptive:   "Adaptive",
	DemoBreakpoint: "Breakpoints",
}

// CreateApp creates the root application component for the responsive layout showcase.
func CreateApp() (bubbly.Component, error) {
	return bubbly.NewComponent("ResponsiveApp").
		WithAutoCommands(true).
		// Navigation between demos
		WithKeyBinding("1", "demo1", "Dashboard").
		WithKeyBinding("2", "demo2", "Grid").
		WithKeyBinding("3", "demo3", "Adaptive").
		WithKeyBinding("4", "demo4", "Breakpoints").
		WithMultiKeyBindings("nextDemo", "Next demo", "tab", "right", "l").
		WithMultiKeyBindings("prevDemo", "Previous demo", "shift+tab", "left", "h").
		// Quit
		WithKeyBinding("q", "quit", "Quit").
		WithKeyBinding("ctrl+c", "quit", "Quit").
		// Handle window resize messages
		WithMessageHandler(func(comp bubbly.Component, msg tea.Msg) tea.Cmd {
			switch msg := msg.(type) {
			case tea.WindowSizeMsg:
				// Update shared window size state
				comp.Emit("resize", map[string]int{
					"width":  msg.Width,
					"height": msg.Height,
				})
				return nil
			}
			return nil
		}).
		Setup(func(ctx *bubbly.Context) {
			// Provide theme for all descendants
			ctx.ProvideTheme(bubbly.DefaultTheme)
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)

			// Current demo state
			currentDemo := bubbly.NewRef(int(DemoDashboard))
			ctx.Expose("currentDemo", currentDemo)

			// Get shared window size state
			windowSize := localComposables.UseSharedWindowSize(ctx)
			ctx.Expose("windowSize", windowSize)

			// Create demo components
			dashboardDemo, err := localComponents.CreateResponsiveDashboard()
			if err != nil {
				panic(fmt.Sprintf("Failed to create dashboard demo: %v", err))
			}

			gridDemo, err := localComponents.CreateResponsiveGrid()
			if err != nil {
				panic(fmt.Sprintf("Failed to create grid demo: %v", err))
			}

			adaptiveDemo, err := localComponents.CreateAdaptiveContent()
			if err != nil {
				panic(fmt.Sprintf("Failed to create adaptive demo: %v", err))
			}

			breakpointDemo, err := localComponents.CreateBreakpointDemo()
			if err != nil {
				panic(fmt.Sprintf("Failed to create breakpoint demo: %v", err))
			}

			// Expose components
			if err := ctx.ExposeComponent("dashboardDemo", dashboardDemo); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose dashboard: %v", err))
			}
			if err := ctx.ExposeComponent("gridDemo", gridDemo); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose grid: %v", err))
			}
			if err := ctx.ExposeComponent("adaptiveDemo", adaptiveDemo); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose adaptive: %v", err))
			}
			if err := ctx.ExposeComponent("breakpointDemo", breakpointDemo); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose breakpoint: %v", err))
			}

			// Event handlers for navigation
			ctx.On("demo1", func(_ interface{}) {
				currentDemo.Set(int(DemoDashboard))
			})
			ctx.On("demo2", func(_ interface{}) {
				currentDemo.Set(int(DemoGrid))
			})
			ctx.On("demo3", func(_ interface{}) {
				currentDemo.Set(int(DemoAdaptive))
			})
			ctx.On("demo4", func(_ interface{}) {
				currentDemo.Set(int(DemoBreakpoint))
			})
			ctx.On("nextDemo", func(_ interface{}) {
				current := currentDemo.GetTyped()
				next := (current + 1) % 4
				currentDemo.Set(next)
			})
			ctx.On("prevDemo", func(_ interface{}) {
				current := currentDemo.GetTyped()
				prev := (current - 1 + 4) % 4
				currentDemo.Set(prev)
			})

			// Handle window resize
			ctx.On("resize", func(data interface{}) {
				if sizeData, ok := data.(map[string]int); ok {
					windowSize.SetSize(sizeData["width"], sizeData["height"])
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			currentDemo := ctx.Get("currentDemo").(*bubbly.Ref[int]).GetTyped()
			ws := ctx.Get("windowSize").(*localComposables.WindowSizeComposable)

			width := ws.Width.GetTyped()
			breakpoint := ws.Breakpoint.GetTyped()

			// === HEADER ===
			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				Background(lipgloss.Color("236")).
				Padding(0, 2)

			headerText := components.Text(components.TextProps{
				Content: headerStyle.Render(fmt.Sprintf(
					"🎨 BubblyUI Responsive Layouts [%s]",
					breakpoint,
				)),
			})
			headerText.Init()

			// === TAB BAR ===
			var tabItems []interface{}
			for i := 0; i < 4; i++ {
				demoType := DemoType(i)
				name := DemoNames[demoType]

				style := lipgloss.NewStyle().Padding(0, 1)
				if demoType == DemoType(currentDemo) {
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
			switch DemoType(currentDemo) {
			case DemoDashboard:
				demoContent = ctx.Get("dashboardDemo").(bubbly.Component)
			case DemoGrid:
				demoContent = ctx.Get("gridDemo").(bubbly.Component)
			case DemoAdaptive:
				demoContent = ctx.Get("adaptiveDemo").(bubbly.Component)
			case DemoBreakpoint:
				demoContent = ctx.Get("breakpointDemo").(bubbly.Component)
			default:
				demoContent = ctx.Get("dashboardDemo").(bubbly.Component)
			}

			// === FOOTER ===
			footerStyle := lipgloss.NewStyle().
				Foreground(theme.Muted).
				Italic(true)

			footerText := "1-4: switch demo • tab/shift+tab: next/prev • Resize terminal to see responsive behavior • q: quit"
			footer := components.Text(components.TextProps{
				Content: footerStyle.Render(footerText),
			})
			footer.Init()

			// === DIVIDERS ===
			headerDivider := components.Divider(components.DividerProps{
				Length: width - 4,
				Char:   "═",
			})
			headerDivider.Init()

			// === MAIN LAYOUT ===
			page := components.VStack(components.StackProps{
				Items: []interface{}{
					headerText,
					tabBar,
					headerDivider,
					demoContent,
					footer,
				},
				Spacing: 0,
			})
			page.Init()

			return page.View()
		}).
		Build()
}
