// Package components provides UI components for the enhanced composables demo.
package components

import (
	"fmt"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/17-enhanced-composables/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

func CreateCounterCard() (bubbly.Component, error) {
	return bubbly.NewComponent("CounterCard").
		Setup(func(ctx *bubbly.Context) {
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			value := state.CounterValue.GetTyped()
			canUndo := state.CanUndo.GetTyped()
			canRedo := state.CanRedo.GetTyped()
			prevVal := state.PreviousVal.GetTyped()

			prevStr := "none"
			if prevVal != nil {
				prevStr = fmt.Sprintf("%d", *prevVal)
			}

			content := fmt.Sprintf(
				"Value: %d (prev: %s)\nCan Undo: %v | Can Redo: %v\n\n+/-: change | u/r: undo/redo",
				value, prevStr, canUndo, canRedo,
			)

			card := components.Card(components.CardProps{
				Title:   "Counter (UseCounter + UseHistory)",
				Content: content,
				Width:   42,
			})
			card.Init()
			return card.View()
		}).
		Build()
}
