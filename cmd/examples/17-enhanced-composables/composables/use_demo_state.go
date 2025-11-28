// Package composables provides shared composables for the enhanced composables demo.
package composables

import (
	"sort"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// FocusPane represents the focusable panes in the app.
type FocusPane int

const (
	FocusSidebar FocusPane = iota
	FocusMain
)

// ComposableItem represents an item in the sidebar list.
type ComposableItem struct {
	Name     string
	Category string
}

// DemoStateComposable holds all the shared state for the demo.
type DemoStateComposable struct {
	Width          *bubbly.Ref[int]
	Height         *bubbly.Ref[int]
	SidebarWidth   int
	FocusedPane    *bubbly.Ref[FocusPane]
	SidebarIndex   *bubbly.Ref[int]
	SidebarItems   []ComposableItem
	SelectedDetail *bubbly.Ref[string]
	CounterValue   *bubbly.Ref[int]
	PreviousVal    *bubbly.Ref[*int]
	CanUndo        *bubbly.Ref[bool]
	CanRedo        *bubbly.Ref[bool]
	history        *composables.HistoryReturn[int]
	Timer          *composables.TimerReturn
	TimerIsRunning *bubbly.Ref[bool]
	TimerIsExpired *bubbly.Ref[bool]
	TimerRemaining *bubbly.Ref[time.Duration]
	TimerProgress  *bubbly.Ref[float64]
	DarkMode       *bubbly.Ref[bool]
	TaskList       *bubbly.Ref[[]string]
	TagsSlice      *bubbly.Ref[[]string]
	tags           *composables.SetReturn[string]
	Notifications  *composables.NotificationReturn
}

func (d *DemoStateComposable) SetSize(width, height int) {
	d.Width.Set(width)
	d.Height.Set(height)
}

func (d *DemoStateComposable) CycleFocus() {
	current := d.FocusedPane.GetTyped()
	if current == FocusSidebar {
		d.FocusedPane.Set(FocusMain)
	} else {
		d.FocusedPane.Set(FocusSidebar)
	}
}

func (d *DemoStateComposable) SidebarUp() {
	idx := d.SidebarIndex.GetTyped()
	if idx > 0 {
		d.SidebarIndex.Set(idx - 1)
		d.updateSelectedDetail()
	}
}

func (d *DemoStateComposable) SidebarDown() {
	idx := d.SidebarIndex.GetTyped()
	if idx < len(d.SidebarItems)-1 {
		d.SidebarIndex.Set(idx + 1)
		d.updateSelectedDetail()
	}
}

func (d *DemoStateComposable) SelectSidebarItem() {
	d.updateSelectedDetail()
}

func (d *DemoStateComposable) updateSelectedDetail() {
	idx := d.SidebarIndex.GetTyped()
	if idx >= 0 && idx < len(d.SidebarItems) {
		item := d.SidebarItems[idx]
		d.SelectedDetail.Set(item.Name)
	}
}

func (d *DemoStateComposable) Increment() {
	val := d.CounterValue.GetTyped()
	if val < 100 {
		newVal := val + 5
		d.CounterValue.Set(newVal)
		d.history.Push(newVal)
		d.updateHistoryState()
		d.PreviousVal.Set(&val)
	}
}

func (d *DemoStateComposable) Decrement() {
	val := d.CounterValue.GetTyped()
	if val > 0 {
		newVal := val - 5
		d.CounterValue.Set(newVal)
		d.history.Push(newVal)
		d.updateHistoryState()
		d.PreviousVal.Set(&val)
	}
}

func (d *DemoStateComposable) Undo() {
	d.history.Undo()
	if val, ok := d.history.Current.Get().(int); ok {
		d.CounterValue.Set(val)
	}
	d.updateHistoryState()
}

func (d *DemoStateComposable) Redo() {
	d.history.Redo()
	if val, ok := d.history.Current.Get().(int); ok {
		d.CounterValue.Set(val)
	}
	d.updateHistoryState()
}

func (d *DemoStateComposable) updateHistoryState() {
	d.CanUndo.Set(d.history.CanUndo.Get().(bool))
	d.CanRedo.Set(d.history.CanRedo.Get().(bool))
}

func (d *DemoStateComposable) ToggleTimer() {
	if d.Timer.IsRunning.GetTyped() {
		d.Timer.Stop()
	} else {
		if d.Timer.IsExpired.GetTyped() {
			d.Timer.Reset()
		}
		d.Timer.Start()
	}
	d.syncTimerState()
}

func (d *DemoStateComposable) syncTimerState() {
	d.TimerIsRunning.Set(d.Timer.IsRunning.GetTyped())
	d.TimerIsExpired.Set(d.Timer.IsExpired.GetTyped())
	d.TimerRemaining.Set(d.Timer.Remaining.GetTyped())
	d.TimerProgress.Set(d.Timer.Progress.GetTyped())
}

func (d *DemoStateComposable) UpdateTimerDisplay() {
	d.syncTimerState()
}

func (d *DemoStateComposable) ToggleDarkMode() {
	d.DarkMode.Set(!d.DarkMode.GetTyped())
}

func (d *DemoStateComposable) updateTagsSlice() {
	slice := d.tags.ToSlice()
	sort.Strings(slice)
	d.TagsSlice.Set(slice)
}

func (d *DemoStateComposable) ShowNotification(title, message string) {
	d.Notifications.Success(title, message)
}

func UseDemoState(ctx *bubbly.Context) *DemoStateComposable {
	sidebarItems := []ComposableItem{
		{Name: "UseWindowSize", Category: "TUI"},
		{Name: "UseFocus", Category: "TUI"},
		{Name: "UseScroll", Category: "TUI"},
		{Name: "UseSelection", Category: "TUI"},
		{Name: "UseMode", Category: "TUI"},
		{Name: "UseToggle", Category: "State"},
		{Name: "UseCounter", Category: "State"},
		{Name: "UsePrevious", Category: "State"},
		{Name: "UseHistory", Category: "State"},
		{Name: "UseInterval", Category: "Timing"},
		{Name: "UseTimeout", Category: "Timing"},
		{Name: "UseTimer", Category: "Timing"},
		{Name: "UseList", Category: "Collections"},
		{Name: "UseMap", Category: "Collections"},
		{Name: "UseSet", Category: "Collections"},
		{Name: "UseQueue", Category: "Collections"},
		{Name: "UseLogger", Category: "Dev"},
		{Name: "UseNotification", Category: "Dev"},
		{Name: "CreateShared", Category: "Utilities"},
		{Name: "CreateSharedWithReset", Category: "Utilities"},
	}

	history := composables.UseHistory(ctx, 50, 20)
	timer := composables.UseTimer(ctx, 30*time.Second,
		composables.WithTickInterval(100*time.Millisecond),
	)
	tags := composables.UseSet(ctx, []string{"composables", "go", "tui"})
	notifications := composables.UseNotification(ctx,
		composables.WithDefaultDuration(3*time.Second),
		composables.WithMaxNotifications(3),
	)

	tagsSlice := tags.ToSlice()
	sort.Strings(tagsSlice)

	return &DemoStateComposable{
		Width:          bubbly.NewRef(80),
		Height:         bubbly.NewRef(24),
		SidebarWidth:   26,
		FocusedPane:    bubbly.NewRef(FocusSidebar),
		SidebarIndex:   bubbly.NewRef(0),
		SidebarItems:   sidebarItems,
		SelectedDetail: bubbly.NewRef("UseWindowSize"),
		CounterValue:   bubbly.NewRef(50),
		PreviousVal:    bubbly.NewRef[*int](nil),
		CanUndo:        bubbly.NewRef(false),
		CanRedo:        bubbly.NewRef(false),
		history:        history,
		Timer:          timer,
		TimerIsRunning: bubbly.NewRef(false),
		TimerIsExpired: bubbly.NewRef(false),
		TimerRemaining: bubbly.NewRef(30 * time.Second),
		TimerProgress:  bubbly.NewRef(0.0),
		DarkMode:       bubbly.NewRef(false),
		TaskList:       bubbly.NewRef([]string{"Learn UseWindowSize", "Master UseFocus", "Implement UseScroll"}),
		TagsSlice:      bubbly.NewRef(tagsSlice),
		tags:           tags,
		Notifications:  notifications,
	}
}

var UseSharedDemoState = composables.CreateShared(UseDemoState)
