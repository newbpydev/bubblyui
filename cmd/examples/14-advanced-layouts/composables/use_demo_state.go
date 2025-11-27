// Package composables provides shared state composables for the layout demo.
package composables

import (
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// DemoType represents the type of demo being displayed.
type DemoType int

const (
	// DemoDashboard shows a complete dashboard layout.
	DemoDashboard DemoType = iota
	// DemoFlex shows interactive Flex justify/align options.
	DemoFlex
	// DemoCardGrid shows a wrapping card grid.
	DemoCardGrid
	// DemoForm shows form layout patterns.
	DemoForm
	// DemoModal shows centered modal/dialog patterns.
	DemoModal
)

// DemoNames provides human-readable names for each demo.
var DemoNames = map[DemoType]string{
	DemoDashboard: "Dashboard",
	DemoFlex:      "Flex Layout",
	DemoCardGrid:  "Card Grid",
	DemoForm:      "Form Layout",
	DemoModal:     "Modal/Dialog",
}

// JustifyOptions lists all available justify options for Flex demo.
var JustifyOptions = []components.JustifyContent{
	components.JustifyStart,
	components.JustifyCenter,
	components.JustifyEnd,
	components.JustifySpaceBetween,
	components.JustifySpaceAround,
	components.JustifySpaceEvenly,
}

// AlignOptions lists all available align options for Flex demo.
var AlignOptions = []components.AlignItems{
	components.AlignItemsStart,
	components.AlignItemsCenter,
	components.AlignItemsEnd,
	components.AlignItemsStretch,
}

// DemoStateComposable holds the shared state for the layout demo app.
type DemoStateComposable struct {
	// CurrentDemo is the currently selected demo type.
	CurrentDemo *bubbly.Ref[interface{}]

	// JustifyIndex is the current justify option index (for Flex demo).
	JustifyIndex *bubbly.Ref[interface{}]

	// AlignIndex is the current align option index (for Flex demo).
	AlignIndex *bubbly.Ref[interface{}]

	// FlexDirection is the current flex direction (row/column).
	FlexDirection *bubbly.Ref[interface{}]

	// WrapEnabled indicates if flex wrap is enabled.
	WrapEnabled *bubbly.Ref[interface{}]

	// GapSize is the current gap size.
	GapSize *bubbly.Ref[interface{}]

	// ModalVisible indicates if the modal is visible (for Modal demo).
	ModalVisible *bubbly.Ref[interface{}]

	// Methods
	NextDemo        func()
	PrevDemo        func()
	SetDemo         func(DemoType)
	NextJustify     func()
	PrevJustify     func()
	NextAlign       func()
	PrevAlign       func()
	ToggleWrap      func()
	ToggleDirection func()
	IncreaseGap     func()
	DecreaseGap     func()
	ToggleModal     func()
}

// UseDemoState creates a new demo state composable.
func UseDemoState(ctx *bubbly.Context) *DemoStateComposable {
	// Create refs for all state
	currentDemo := ctx.Ref(int(DemoDashboard))
	justifyIndex := ctx.Ref(0)
	alignIndex := ctx.Ref(0)
	flexDirection := ctx.Ref(string(components.FlexRow))
	wrapEnabled := ctx.Ref(false)
	gapSize := ctx.Ref(2)
	modalVisible := ctx.Ref(false)

	return &DemoStateComposable{
		CurrentDemo:   currentDemo,
		JustifyIndex:  justifyIndex,
		AlignIndex:    alignIndex,
		FlexDirection: flexDirection,
		WrapEnabled:   wrapEnabled,
		GapSize:       gapSize,
		ModalVisible:  modalVisible,

		NextDemo: func() {
			current := currentDemo.Get().(int)
			next := (current + 1) % 5
			currentDemo.Set(next)
		},

		PrevDemo: func() {
			current := currentDemo.Get().(int)
			prev := current - 1
			if prev < 0 {
				prev = 4
			}
			currentDemo.Set(prev)
		},

		SetDemo: func(demo DemoType) {
			currentDemo.Set(int(demo))
		},

		NextJustify: func() {
			current := justifyIndex.Get().(int)
			next := (current + 1) % len(JustifyOptions)
			justifyIndex.Set(next)
		},

		PrevJustify: func() {
			current := justifyIndex.Get().(int)
			prev := current - 1
			if prev < 0 {
				prev = len(JustifyOptions) - 1
			}
			justifyIndex.Set(prev)
		},

		NextAlign: func() {
			current := alignIndex.Get().(int)
			next := (current + 1) % len(AlignOptions)
			alignIndex.Set(next)
		},

		PrevAlign: func() {
			current := alignIndex.Get().(int)
			prev := current - 1
			if prev < 0 {
				prev = len(AlignOptions) - 1
			}
			alignIndex.Set(prev)
		},

		ToggleWrap: func() {
			current := wrapEnabled.Get().(bool)
			wrapEnabled.Set(!current)
		},

		ToggleDirection: func() {
			current := flexDirection.Get().(string)
			if current == string(components.FlexRow) {
				flexDirection.Set(string(components.FlexColumn))
			} else {
				flexDirection.Set(string(components.FlexRow))
			}
		},

		IncreaseGap: func() {
			current := gapSize.Get().(int)
			if current < 10 {
				gapSize.Set(current + 1)
			}
		},

		DecreaseGap: func() {
			current := gapSize.Get().(int)
			if current > 0 {
				gapSize.Set(current - 1)
			}
		},

		ToggleModal: func() {
			current := modalVisible.Get().(bool)
			modalVisible.Set(!current)
		},
	}
}

// UseSharedDemoState creates a shared demo state instance across all components.
var UseSharedDemoState = composables.CreateShared(
	func(ctx *bubbly.Context) *DemoStateComposable {
		return UseDemoState(ctx)
	},
)
