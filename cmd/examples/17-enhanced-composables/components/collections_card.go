// Package components provides UI components for the enhanced composables demo.
package components

import (
	"fmt"
	"strings"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/17-enhanced-composables/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

func CreateCollectionsCard() (bubbly.Component, error) {
	return bubbly.NewComponent("CollectionsCard").
		Setup(func(ctx *bubbly.Context) {
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			tasks := state.TaskList.GetTyped()
			tags := state.TagsSlice.GetTyped()
			darkMode := state.DarkMode.GetTyped()

			content := fmt.Sprintf(
				"Tasks (UseList): %d items\nTags (UseSet): [%s]\nDark Mode (UseToggle): %v\n\nspace: toggle dark mode",
				len(tasks),
				strings.Join(tags, ", "),
				darkMode,
			)

			card := components.Card(components.CardProps{
				Title:   "Collections & Toggle",
				Content: content,
				Width:   42,
			})
			card.Init()
			return card.View()
		}).
		Build()
}
