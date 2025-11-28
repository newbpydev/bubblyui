// Package components provides UI components for the enhanced composables demo.
package components

import (
	localComposables "github.com/newbpydev/bubblyui/cmd/examples/17-enhanced-composables/composables"
	"github.com/newbpydev/bubblyui/cmd/examples/17-enhanced-composables/demos"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CreateContentArea creates the content area that switches between demo views.
func CreateContentArea() (bubbly.Component, error) {
	return bubbly.NewComponent("ContentArea").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)

			// Create all demo components
			homeView, _ := demos.CreateHomeView()
			useWindowSizeDemo, _ := demos.CreateUseWindowSizeDemo()
			useFocusDemo, _ := demos.CreateUseFocusDemo()
			useScrollDemo, _ := demos.CreateUseScrollDemo()
			useSelectionDemo, _ := demos.CreateUseSelectionDemo()
			useModeDemo, _ := demos.CreateUseModeDemo()
			useToggleDemo, _ := demos.CreateUseToggleDemo()
			useCounterDemo, _ := demos.CreateUseCounterDemo()
			usePreviousDemo, _ := demos.CreateUsePreviousDemo()
			useHistoryDemo, _ := demos.CreateUseHistoryDemo()
			useIntervalDemo, _ := demos.CreateUseIntervalDemo()
			useTimeoutDemo, _ := demos.CreateUseTimeoutDemo()
			useTimerDemo, _ := demos.CreateUseTimerDemo()
			useListDemo, _ := demos.CreateUseListDemo()
			useMapDemo, _ := demos.CreateUseMapDemo()
			useSetDemo, _ := demos.CreateUseSetDemo()
			useQueueDemo, _ := demos.CreateUseQueueDemo()
			useLoggerDemo, _ := demos.CreateUseLoggerDemo()
			useNotificationDemo, _ := demos.CreateUseNotificationDemo()
			createSharedDemo, _ := demos.CreateCreateSharedDemo()
			createSharedWithResetDemo, _ := demos.CreateCreateSharedWithResetDemo()

			// Expose all demos
			_ = ctx.ExposeComponent("homeView", homeView)
			_ = ctx.ExposeComponent("useWindowSizeDemo", useWindowSizeDemo)
			_ = ctx.ExposeComponent("useFocusDemo", useFocusDemo)
			_ = ctx.ExposeComponent("useScrollDemo", useScrollDemo)
			_ = ctx.ExposeComponent("useSelectionDemo", useSelectionDemo)
			_ = ctx.ExposeComponent("useModeDemo", useModeDemo)
			_ = ctx.ExposeComponent("useToggleDemo", useToggleDemo)
			_ = ctx.ExposeComponent("useCounterDemo", useCounterDemo)
			_ = ctx.ExposeComponent("usePreviousDemo", usePreviousDemo)
			_ = ctx.ExposeComponent("useHistoryDemo", useHistoryDemo)
			_ = ctx.ExposeComponent("useIntervalDemo", useIntervalDemo)
			_ = ctx.ExposeComponent("useTimeoutDemo", useTimeoutDemo)
			_ = ctx.ExposeComponent("useTimerDemo", useTimerDemo)
			_ = ctx.ExposeComponent("useListDemo", useListDemo)
			_ = ctx.ExposeComponent("useMapDemo", useMapDemo)
			_ = ctx.ExposeComponent("useSetDemo", useSetDemo)
			_ = ctx.ExposeComponent("useQueueDemo", useQueueDemo)
			_ = ctx.ExposeComponent("useLoggerDemo", useLoggerDemo)
			_ = ctx.ExposeComponent("useNotificationDemo", useNotificationDemo)
			_ = ctx.ExposeComponent("createSharedDemo", createSharedDemo)
			_ = ctx.ExposeComponent("createSharedWithResetDemo", createSharedWithResetDemo)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			state := ctx.Get("state").(*localComposables.DemoStateComposable)
			activeView := state.ActiveView.GetTyped()

			// Get the appropriate demo component based on active view
			var demoComponent bubbly.Component

			switch activeView {
			case localComposables.ViewHome:
				demoComponent = ctx.Get("homeView").(bubbly.Component)
			case localComposables.ViewUseWindowSize:
				demoComponent = ctx.Get("useWindowSizeDemo").(bubbly.Component)
			case localComposables.ViewUseFocus:
				demoComponent = ctx.Get("useFocusDemo").(bubbly.Component)
			case localComposables.ViewUseScroll:
				demoComponent = ctx.Get("useScrollDemo").(bubbly.Component)
			case localComposables.ViewUseSelection:
				demoComponent = ctx.Get("useSelectionDemo").(bubbly.Component)
			case localComposables.ViewUseMode:
				demoComponent = ctx.Get("useModeDemo").(bubbly.Component)
			case localComposables.ViewUseToggle:
				demoComponent = ctx.Get("useToggleDemo").(bubbly.Component)
			case localComposables.ViewUseCounter:
				demoComponent = ctx.Get("useCounterDemo").(bubbly.Component)
			case localComposables.ViewUsePrevious:
				demoComponent = ctx.Get("usePreviousDemo").(bubbly.Component)
			case localComposables.ViewUseHistory:
				demoComponent = ctx.Get("useHistoryDemo").(bubbly.Component)
			case localComposables.ViewUseInterval:
				demoComponent = ctx.Get("useIntervalDemo").(bubbly.Component)
			case localComposables.ViewUseTimeout:
				demoComponent = ctx.Get("useTimeoutDemo").(bubbly.Component)
			case localComposables.ViewUseTimer:
				demoComponent = ctx.Get("useTimerDemo").(bubbly.Component)
			case localComposables.ViewUseList:
				demoComponent = ctx.Get("useListDemo").(bubbly.Component)
			case localComposables.ViewUseMap:
				demoComponent = ctx.Get("useMapDemo").(bubbly.Component)
			case localComposables.ViewUseSet:
				demoComponent = ctx.Get("useSetDemo").(bubbly.Component)
			case localComposables.ViewUseQueue:
				demoComponent = ctx.Get("useQueueDemo").(bubbly.Component)
			case localComposables.ViewUseLogger:
				demoComponent = ctx.Get("useLoggerDemo").(bubbly.Component)
			case localComposables.ViewUseNotification:
				demoComponent = ctx.Get("useNotificationDemo").(bubbly.Component)
			case localComposables.ViewCreateShared:
				demoComponent = ctx.Get("createSharedDemo").(bubbly.Component)
			case localComposables.ViewCreateSharedReset:
				demoComponent = ctx.Get("createSharedWithResetDemo").(bubbly.Component)
			default:
				demoComponent = ctx.Get("homeView").(bubbly.Component)
			}

			return demoComponent.View()
		}).
		Build()
}
