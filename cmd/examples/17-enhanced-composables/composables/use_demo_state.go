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

// LogEntry represents a log entry for the UseLogger demo.
type LogEntry struct {
	Level   string
	Message string
}

// ViewType represents the current content view.
type ViewType string

const (
	ViewHome              ViewType = "home"
	ViewUseWindowSize     ViewType = "UseWindowSize"
	ViewUseFocus          ViewType = "UseFocus"
	ViewUseScroll         ViewType = "UseScroll"
	ViewUseSelection      ViewType = "UseSelection"
	ViewUseMode           ViewType = "UseMode"
	ViewUseToggle         ViewType = "UseToggle"
	ViewUseCounter        ViewType = "UseCounter"
	ViewUsePrevious       ViewType = "UsePrevious"
	ViewUseHistory        ViewType = "UseHistory"
	ViewUseInterval       ViewType = "UseInterval"
	ViewUseTimeout        ViewType = "UseTimeout"
	ViewUseTimer          ViewType = "UseTimer"
	ViewUseList           ViewType = "UseList"
	ViewUseMap            ViewType = "UseMap"
	ViewUseSet            ViewType = "UseSet"
	ViewUseQueue          ViewType = "UseQueue"
	ViewUseLogger         ViewType = "UseLogger"
	ViewUseNotification   ViewType = "UseNotification"
	ViewCreateShared      ViewType = "CreateShared"
	ViewCreateSharedReset ViewType = "CreateSharedWithReset"
)

// DemoStateComposable holds all the shared state for the demo.
type DemoStateComposable struct {
	Width          *bubbly.Ref[int]
	Height         *bubbly.Ref[int]
	SidebarWidth   int
	FocusedPane    *bubbly.Ref[FocusPane]
	SidebarIndex   *bubbly.Ref[int]
	SidebarItems   []ComposableItem
	SelectedDetail *bubbly.Ref[string]
	ActiveView     *bubbly.Ref[ViewType]
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

	// === TUI Demo State ===
	// UseFocus demo: which pane is focused (0-2)
	FocusDemoIndex *bubbly.Ref[int]
	// UseScroll demo: scroll offset and total items
	ScrollDemoOffset *bubbly.Ref[int]
	ScrollDemoTotal  int
	// UseSelection demo: selected items
	SelectionDemoIndex *bubbly.Ref[int]
	SelectionDemoItems *bubbly.Ref[[]bool]
	// UseMode demo: current mode
	ModeDemoMode  *bubbly.Ref[string]
	ModeDemoModes []string

	// === State Demo State ===
	// UseToggle demo: 3 toggles
	ToggleDemo1 *bubbly.Ref[bool]
	ToggleDemo2 *bubbly.Ref[bool]
	ToggleDemo3 *bubbly.Ref[bool]
	// UseCounter demo: local counter with bounds
	LocalCounter    *bubbly.Ref[int]
	LocalCounterMin int
	LocalCounterMax int

	// === Timing Demo State ===
	// UseInterval demo
	Interval        *composables.IntervalReturn
	IntervalRunning *bubbly.Ref[bool]
	IntervalCount   *bubbly.Ref[int]
	// UseTimeout demo
	Timeout          *composables.TimeoutReturn
	TimeoutPending   *bubbly.Ref[bool]
	TimeoutTriggered *bubbly.Ref[bool]

	// === Collections Demo State ===
	// UseList demo
	ListDemoItems *bubbly.Ref[[]string]
	ListDemoIndex *bubbly.Ref[int]
	// UseMap demo
	MapDemoData *bubbly.Ref[map[string]int]
	// UseSet demo
	SetDemoItems *bubbly.Ref[[]string]
	// UseQueue demo
	QueueDemoItems *bubbly.Ref[[]string]

	// === Dev Demo State ===
	// UseLogger demo
	LoggerEntries *bubbly.Ref[[]LogEntry]
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
	// Update active view based on selection
	idx := d.SidebarIndex.GetTyped()
	if idx >= 0 && idx < len(d.SidebarItems) {
		item := d.SidebarItems[idx]
		d.ActiveView.Set(ViewType(item.Name))
	}
}

// GoHome returns to the home view.
func (d *DemoStateComposable) GoHome() {
	d.ActiveView.Set(ViewHome)
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

// === TUI Demo Methods ===

// FocusDemoNext cycles to next focus pane (0-2)
func (d *DemoStateComposable) FocusDemoNext() {
	idx := d.FocusDemoIndex.GetTyped()
	d.FocusDemoIndex.Set((idx + 1) % 3)
}

// FocusDemoPrev cycles to previous focus pane
func (d *DemoStateComposable) FocusDemoPrev() {
	idx := d.FocusDemoIndex.GetTyped()
	if idx == 0 {
		d.FocusDemoIndex.Set(2)
	} else {
		d.FocusDemoIndex.Set(idx - 1)
	}
}

// FocusDemoSet sets focus to specific pane (1, 2, or 3)
func (d *DemoStateComposable) FocusDemoSet(pane int) {
	if pane >= 1 && pane <= 3 {
		d.FocusDemoIndex.Set(pane - 1)
	}
}

// ScrollDemoUp scrolls up
func (d *DemoStateComposable) ScrollDemoUp() {
	offset := d.ScrollDemoOffset.GetTyped()
	if offset > 0 {
		d.ScrollDemoOffset.Set(offset - 1)
	}
}

// ScrollDemoDown scrolls down
func (d *DemoStateComposable) ScrollDemoDown() {
	offset := d.ScrollDemoOffset.GetTyped()
	maxOffset := d.ScrollDemoTotal - 5 // visible count
	if offset < maxOffset {
		d.ScrollDemoOffset.Set(offset + 1)
	}
}

// ScrollDemoTop scrolls to top
func (d *DemoStateComposable) ScrollDemoTop() {
	d.ScrollDemoOffset.Set(0)
}

// ScrollDemoBottom scrolls to bottom
func (d *DemoStateComposable) ScrollDemoBottom() {
	d.ScrollDemoOffset.Set(d.ScrollDemoTotal - 5)
}

// SelectionDemoUp moves selection up
func (d *DemoStateComposable) SelectionDemoUp() {
	idx := d.SelectionDemoIndex.GetTyped()
	if idx > 0 {
		d.SelectionDemoIndex.Set(idx - 1)
	}
}

// SelectionDemoDown moves selection down
func (d *DemoStateComposable) SelectionDemoDown() {
	idx := d.SelectionDemoIndex.GetTyped()
	items := d.SelectionDemoItems.GetTyped()
	if idx < len(items)-1 {
		d.SelectionDemoIndex.Set(idx + 1)
	}
}

// SelectionDemoToggle toggles selection of current item
func (d *DemoStateComposable) SelectionDemoToggle() {
	idx := d.SelectionDemoIndex.GetTyped()
	items := d.SelectionDemoItems.GetTyped()
	if idx >= 0 && idx < len(items) {
		newItems := make([]bool, len(items))
		copy(newItems, items)
		newItems[idx] = !newItems[idx]
		d.SelectionDemoItems.Set(newItems)
	}
}

// ModeDemoSet sets the mode (1-4)
func (d *DemoStateComposable) ModeDemoSet(modeNum int) {
	if modeNum >= 1 && modeNum <= len(d.ModeDemoModes) {
		d.ModeDemoMode.Set(d.ModeDemoModes[modeNum-1])
	}
}

// === State Demo Methods ===

// ToggleDemoToggle toggles a specific toggle (1, 2, or 3)
func (d *DemoStateComposable) ToggleDemoToggle(num int) {
	switch num {
	case 1:
		d.ToggleDemo1.Set(!d.ToggleDemo1.GetTyped())
	case 2:
		d.ToggleDemo2.Set(!d.ToggleDemo2.GetTyped())
	case 3:
		d.ToggleDemo3.Set(!d.ToggleDemo3.GetTyped())
	}
}

// LocalCounterIncrement increments the local counter
func (d *DemoStateComposable) LocalCounterIncrement() {
	val := d.LocalCounter.GetTyped()
	if val < d.LocalCounterMax {
		d.LocalCounter.Set(val + 5)
	}
}

// LocalCounterDecrement decrements the local counter
func (d *DemoStateComposable) LocalCounterDecrement() {
	val := d.LocalCounter.GetTyped()
	if val > d.LocalCounterMin {
		d.LocalCounter.Set(val - 5)
	}
}

// LocalCounterReset resets the local counter
func (d *DemoStateComposable) LocalCounterReset() {
	d.LocalCounter.Set(0)
}

// === Timing Demo Methods ===

// IntervalToggle toggles the interval running state and starts/stops the composable
func (d *DemoStateComposable) IntervalToggle() {
	if d.IntervalRunning.GetTyped() {
		// Stop the interval
		d.Interval.Stop()
		d.IntervalRunning.Set(false)
	} else {
		// Start the interval
		d.IntervalRunning.Set(true)
		d.Interval.Start()
	}
}

// IntervalReset resets the interval count and stops the interval
func (d *DemoStateComposable) IntervalReset() {
	d.Interval.Stop()
	d.IntervalRunning.Set(false)
	d.IntervalCount.Set(0)
}

// TimeoutStart starts the timeout composable
func (d *DemoStateComposable) TimeoutStart() {
	if !d.TimeoutPending.GetTyped() && !d.TimeoutTriggered.GetTyped() {
		d.TimeoutPending.Set(true)
		d.Timeout.Start()
	}
}

// TimeoutReset resets the timeout state and composable
func (d *DemoStateComposable) TimeoutReset() {
	d.Timeout.Cancel()
	d.TimeoutPending.Set(false)
	d.TimeoutTriggered.Set(false)
}

// TimerReset resets the timer
func (d *DemoStateComposable) TimerReset() {
	d.Timer.Reset()
	d.syncTimerState()
}

// === Collections Demo Methods ===

// ListDemoAdd adds an item to the list
func (d *DemoStateComposable) ListDemoAdd() {
	items := d.ListDemoItems.GetTyped()
	newItem := "Item " + string(rune('A'+len(items)%26))
	d.ListDemoItems.Set(append(items, newItem))
}

// ListDemoDelete deletes the selected item
func (d *DemoStateComposable) ListDemoDelete() {
	items := d.ListDemoItems.GetTyped()
	idx := d.ListDemoIndex.GetTyped()
	if len(items) > 0 && idx >= 0 && idx < len(items) {
		newItems := append(items[:idx], items[idx+1:]...)
		d.ListDemoItems.Set(newItems)
		if idx >= len(newItems) && len(newItems) > 0 {
			d.ListDemoIndex.Set(len(newItems) - 1)
		}
	}
}

// ListDemoClear clears the list
func (d *DemoStateComposable) ListDemoClear() {
	d.ListDemoItems.Set([]string{})
	d.ListDemoIndex.Set(0)
}

// ListDemoUp moves selection up
func (d *DemoStateComposable) ListDemoUp() {
	idx := d.ListDemoIndex.GetTyped()
	if idx > 0 {
		d.ListDemoIndex.Set(idx - 1)
	}
}

// ListDemoDown moves selection down
func (d *DemoStateComposable) ListDemoDown() {
	idx := d.ListDemoIndex.GetTyped()
	items := d.ListDemoItems.GetTyped()
	if idx < len(items)-1 {
		d.ListDemoIndex.Set(idx + 1)
	}
}

// MapDemoAdd adds a random item to the map
func (d *DemoStateComposable) MapDemoAdd() {
	data := d.MapDemoData.GetTyped()
	newData := make(map[string]int)
	for k, v := range data {
		newData[k] = v
	}
	keys := []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta"}
	for _, k := range keys {
		if _, exists := newData[k]; !exists {
			newData[k] = len(newData) + 1
			break
		}
	}
	d.MapDemoData.Set(newData)
}

// MapDemoDelete deletes the last added item
func (d *DemoStateComposable) MapDemoDelete() {
	data := d.MapDemoData.GetTyped()
	if len(data) > 0 {
		newData := make(map[string]int)
		var lastKey string
		for k, v := range data {
			newData[k] = v
			lastKey = k
		}
		delete(newData, lastKey)
		d.MapDemoData.Set(newData)
	}
}

// MapDemoClear clears the map
func (d *DemoStateComposable) MapDemoClear() {
	d.MapDemoData.Set(make(map[string]int))
}

// SetDemoAdd adds an item to the set
func (d *DemoStateComposable) SetDemoAdd() {
	items := d.SetDemoItems.GetTyped()
	candidates := []string{"react", "vue", "angular", "svelte", "solid", "preact", "alpine", "htmx"}
	for _, c := range candidates {
		found := false
		for _, item := range items {
			if item == c {
				found = true
				break
			}
		}
		if !found {
			d.SetDemoItems.Set(append(items, c))
			return
		}
	}
}

// SetDemoDelete deletes the last item from the set
func (d *DemoStateComposable) SetDemoDelete() {
	items := d.SetDemoItems.GetTyped()
	if len(items) > 0 {
		d.SetDemoItems.Set(items[:len(items)-1])
	}
}

// SetDemoToggle toggles an item in the set
func (d *DemoStateComposable) SetDemoToggle() {
	items := d.SetDemoItems.GetTyped()
	// Toggle "bubbly" - add if not present, remove if present
	toggleItem := "bubbly"
	for i, item := range items {
		if item == toggleItem {
			d.SetDemoItems.Set(append(items[:i], items[i+1:]...))
			return
		}
	}
	d.SetDemoItems.Set(append(items, toggleItem))
}

// SetDemoClear clears the set
func (d *DemoStateComposable) SetDemoClear() {
	d.SetDemoItems.Set([]string{})
}

// QueueDemoEnqueue adds an item to the queue
func (d *DemoStateComposable) QueueDemoEnqueue() {
	items := d.QueueDemoItems.GetTyped()
	newItem := "Task " + string(rune('A'+len(items)%26))
	d.QueueDemoItems.Set(append(items, newItem))
}

// QueueDemoDequeue removes the first item from the queue
func (d *DemoStateComposable) QueueDemoDequeue() {
	items := d.QueueDemoItems.GetTyped()
	if len(items) > 0 {
		d.QueueDemoItems.Set(items[1:])
	}
}

// QueueDemoClear clears the queue
func (d *DemoStateComposable) QueueDemoClear() {
	d.QueueDemoItems.Set([]string{})
}

// === Dev Demo Methods ===

// LoggerDemoLog adds a log entry
func (d *DemoStateComposable) LoggerDemoLog(level, message string) {
	entries := d.LoggerEntries.GetTyped()
	newEntry := LogEntry{Level: level, Message: message}
	d.LoggerEntries.Set(append(entries, newEntry))
}

// LoggerDemoClear clears all log entries
func (d *DemoStateComposable) LoggerDemoClear() {
	d.LoggerEntries.Set([]LogEntry{})
}

// NotificationDemoShow shows a notification of the given type
func (d *DemoStateComposable) NotificationDemoShow(notifType string) {
	switch notifType {
	case "success":
		d.Notifications.Success("Success", "Operation completed!")
	case "error":
		d.Notifications.Error("Error", "Something went wrong!")
	case "warning":
		d.Notifications.Warning("Warning", "Please be careful!")
	case "info":
		d.Notifications.Info("Info", "Here's some information")
	}
}

// NotificationDemoClear clears all notifications
func (d *DemoStateComposable) NotificationDemoClear() {
	d.Notifications.DismissAll()
}

// === Utilities Demo Methods ===

// SharedReset resets the shared counter (for CreateSharedWithReset demo)
func (d *DemoStateComposable) SharedReset() {
	d.CounterValue.Set(50)
	d.PreviousVal.Set(nil)
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

	// Create refs for interval demo that will be updated by the interval callback
	intervalCount := bubbly.NewRef(0)
	intervalRunning := bubbly.NewRef(false)

	// Create the actual interval composable
	interval := composables.UseInterval(ctx, func() {
		// Only increment if running
		if intervalRunning.GetTyped() {
			intervalCount.Set(intervalCount.GetTyped() + 1)
		}
	}, 500*time.Millisecond)

	// Create refs for timeout demo
	timeoutPending := bubbly.NewRef(false)
	timeoutTriggered := bubbly.NewRef(false)

	// Create the actual timeout composable
	timeout := composables.UseTimeout(ctx, func() {
		timeoutPending.Set(false)
		timeoutTriggered.Set(true)
	}, 3*time.Second)

	tagsSlice := tags.ToSlice()
	sort.Strings(tagsSlice)

	// Initialize mode demo modes
	modeModes := []string{"normal", "insert", "visual", "command"}

	// Create refs for window size
	width := bubbly.NewRef(80)
	height := bubbly.NewRef(24)

	// Task 6.1/6.2: Auto-subscribe to framework's "windowResize" event
	// No manual WithMessageHandler needed - the framework emits this automatically
	if ctx != nil {
		ctx.On("windowResize", func(data interface{}) {
			if sizeData, ok := data.(map[string]int); ok {
				width.Set(sizeData["width"])
				height.Set(sizeData["height"])
			}
		})
	}

	return &DemoStateComposable{
		Width:          width,
		Height:         height,
		SidebarWidth:   26,
		FocusedPane:    bubbly.NewRef(FocusSidebar),
		SidebarIndex:   bubbly.NewRef(0),
		SidebarItems:   sidebarItems,
		SelectedDetail: bubbly.NewRef("Home"),
		ActiveView:     bubbly.NewRef(ViewHome),
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

		// === TUI Demo State ===
		FocusDemoIndex:     bubbly.NewRef(0),
		ScrollDemoOffset:   bubbly.NewRef(0),
		ScrollDemoTotal:    20,
		SelectionDemoIndex: bubbly.NewRef(0),
		SelectionDemoItems: bubbly.NewRef([]bool{false, false, false, false, false}),
		ModeDemoMode:       bubbly.NewRef("normal"),
		ModeDemoModes:      modeModes,

		// === State Demo State ===
		ToggleDemo1:     bubbly.NewRef(false),
		ToggleDemo2:     bubbly.NewRef(true),
		ToggleDemo3:     bubbly.NewRef(false),
		LocalCounter:    bubbly.NewRef(0),
		LocalCounterMin: 0,
		LocalCounterMax: 100,

		// === Timing Demo State ===
		Interval:         interval,
		IntervalRunning:  intervalRunning,
		IntervalCount:    intervalCount,
		Timeout:          timeout,
		TimeoutPending:   timeoutPending,
		TimeoutTriggered: timeoutTriggered,

		// === Collections Demo State ===
		ListDemoItems:  bubbly.NewRef([]string{"Item A", "Item B", "Item C"}),
		ListDemoIndex:  bubbly.NewRef(0),
		MapDemoData:    bubbly.NewRef(map[string]int{"alpha": 1, "beta": 2}),
		SetDemoItems:   bubbly.NewRef([]string{"react", "vue", "angular"}),
		QueueDemoItems: bubbly.NewRef([]string{"Task A", "Task B"}),

		// === Dev Demo State ===
		LoggerEntries: bubbly.NewRef([]LogEntry{}),
	}
}

var UseSharedDemoState = composables.CreateShared(UseDemoState)
