// Package main demonstrates the BubblyUI quickstart example.
// This example showcases clean import paths, component architecture,
// composables, DevTools integration, and profiler usage.
package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	// Clean import paths using alias packages
	"github.com/newbpydev/bubblyui"
	"github.com/newbpydev/bubblyui/cmd/examples/00-quickstart/components"
	"github.com/newbpydev/bubblyui/cmd/examples/00-quickstart/composables"

	// Need pkg/bubbly for Context/RenderContext (builder callback types)
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CreateApp creates the root application component.
// This demonstrates:
// - Clean import paths with the new alias packages
// - Composable architecture pattern
// - Component composition (TaskList, TaskInput, TaskStats, HelpPanel)
// - Key bindings with multi-key support
// - Reactive state management
// - Exposing state for DevTools inspection
func CreateApp() (bubbly.Component, error) {
	return bubbly.NewComponent("TaskManager").
		// Auto-commands for automatic UI updates on Ref.Set()
		WithAutoCommands(true).
		// Navigation and interaction key bindings
		WithKeyBinding("tab", "cycleFocus", "Switch focus").
		WithKeyBinding("j", "moveDown", "Move down").
		WithKeyBinding("k", "moveUp", "Move up").
		WithKeyBinding("down", "moveDown", "Move down").
		WithKeyBinding("up", "moveUp", "Move up").
		WithKeyBinding("enter", "toggleTask", "Toggle task").
		WithKeyBinding(" ", "toggleTask", "Toggle task").
		WithKeyBinding("a", "addMode", "Add task").
		WithKeyBinding("d", "deleteTask", "Delete task").
		WithKeyBinding("f", "cycleFilter", "Cycle filter").
		WithKeyBinding("c", "clearDone", "Clear done").
		WithKeyBinding("q", "quit", "Quit").
		WithKeyBinding("ctrl+c", "quit", "Quit").
		WithKeyBinding("esc", "cancelInput", "Cancel input").
		Setup(func(ctx *bubbly.Context) {
			// Initialize composables for task management and focus
			taskManager := composables.UseTasks([]composables.Task{
				{ID: 1, Text: "Learn BubblyUI basics", Done: true},
				{ID: 2, Text: "Build a task manager", Done: false},
				{ID: 3, Text: "Add DevTools support", Done: false},
			})

			focusManager := composables.UseFocusManager(composables.FocusList)

			// Additional TYPE-SAFE reactive state (PREFERRED pattern)
			selectedIndex := bubblyui.NewRef(0)
			filter := bubblyui.NewRef("all") // "all", "active", "done"
			inputText := bubblyui.NewRef("")
			inputMode := bubblyui.NewRef(false) // Whether we're typing in input

			// Create child components
			taskList, err := components.CreateTaskList(components.TaskListProps{
				Tasks:         taskManager.Tasks,
				SelectedIndex: selectedIndex,
				Filter:        filter,
				IsFocused:     focusManager.IsListFocused,
				OnToggle:      taskManager.ToggleTask,
				OnDelete:      taskManager.DeleteTask,
				GetFiltered:   taskManager.GetFiltered,
			})
			if err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to create TaskList: %v", err))
				return
			}

			taskInput, err := components.CreateTaskInput(components.TaskInputProps{
				InputText: inputText,
				IsFocused: focusManager.IsInputFocused,
				OnSubmit: func(text string) {
					taskManager.AddTask(text)
					focusManager.SetFocus(composables.FocusList)
					inputMode.Set(false)
				},
			})
			if err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to create TaskInput: %v", err))
				return
			}

			taskStats, err := components.CreateTaskStats(components.TaskStatsProps{
				ActiveCount: taskManager.ActiveCount,
				DoneCount:   taskManager.DoneCount,
				TotalCount:  taskManager.TotalCount,
			})
			if err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to create TaskStats: %v", err))
				return
			}

			helpPanel, err := components.CreateHelpPanel(components.HelpPanelProps{})
			if err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to create HelpPanel: %v", err))
				return
			}

			// Register event handlers
			ctx.On("cycleFocus", func(_ interface{}) {
				if !inputMode.GetTyped() {
					focusManager.Next()
				}
			})

			ctx.On("moveDown", func(_ interface{}) {
				if focusManager.IsListFocused() && !inputMode.GetTyped() {
					filteredTasks := taskManager.GetFiltered(filter.GetTyped())
					if len(filteredTasks) > 0 {
						idx := selectedIndex.GetTyped()
						if idx < len(filteredTasks)-1 {
							selectedIndex.Set(idx + 1)
						}
					}
				}
			})

			ctx.On("moveUp", func(_ interface{}) {
				if focusManager.IsListFocused() && !inputMode.GetTyped() {
					idx := selectedIndex.GetTyped()
					if idx > 0 {
						selectedIndex.Set(idx - 1)
					}
				}
			})

			ctx.On("toggleTask", func(_ interface{}) {
				if focusManager.IsListFocused() && !inputMode.GetTyped() {
					filteredTasks := taskManager.GetFiltered(filter.GetTyped())
					if len(filteredTasks) > 0 {
						idx := selectedIndex.GetTyped()
						if idx >= 0 && idx < len(filteredTasks) {
							taskManager.ToggleTask(filteredTasks[idx].ID)
						}
					}
				} else if focusManager.IsInputFocused() || inputMode.GetTyped() {
					// Submit input
					text := inputText.GetTyped()
					if text != "" {
						taskManager.AddTask(text)
						inputText.Set("")
						focusManager.SetFocus(composables.FocusList)
						inputMode.Set(false)
					}
				}
			})

			ctx.On("addMode", func(_ interface{}) {
				if !inputMode.GetTyped() {
					focusManager.SetFocus(composables.FocusInput)
					inputMode.Set(true)
				}
			})

			ctx.On("deleteTask", func(_ interface{}) {
				if focusManager.IsListFocused() && !inputMode.GetTyped() {
					filteredTasks := taskManager.GetFiltered(filter.GetTyped())
					if len(filteredTasks) > 0 {
						idx := selectedIndex.GetTyped()
						if idx >= 0 && idx < len(filteredTasks) {
							taskManager.DeleteTask(filteredTasks[idx].ID)
							// Adjust selection if needed
							if idx >= len(filteredTasks)-1 && idx > 0 {
								selectedIndex.Set(idx - 1)
							}
						}
					}
				}
			})

			ctx.On("cycleFilter", func(_ interface{}) {
				if !inputMode.GetTyped() {
					currentFilter := filter.GetTyped()
					filters := []string{"all", "active", "done"}
					for i, f := range filters {
						if f == currentFilter {
							nextIdx := (i + 1) % len(filters)
							filter.Set(filters[nextIdx])
							selectedIndex.Set(0) // Reset selection
							break
						}
					}
				}
			})

			ctx.On("clearDone", func(_ interface{}) {
				if !inputMode.GetTyped() {
					taskManager.ClearDone()
					selectedIndex.Set(0)
				}
			})

			ctx.On("cancelInput", func(_ interface{}) {
				if inputMode.GetTyped() {
					inputText.Set("")
					focusManager.SetFocus(composables.FocusList)
					inputMode.Set(false)
				}
			})

			// Expose refs for DevTools inspection
			ctx.Expose("tasks", taskManager.Tasks)
			ctx.Expose("selectedIndex", selectedIndex)
			ctx.Expose("filter", filter)
			ctx.Expose("inputText", inputText)
			ctx.Expose("inputMode", inputMode)
			ctx.Expose("focusPane", focusManager.Current)

			// Expose child components (auto-initializes them AND establishes parent-child relationship)
			if err := ctx.ExposeComponent("taskList", taskList); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose taskList: %v", err))
				return
			}
			if err := ctx.ExposeComponent("taskInput", taskInput); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose taskInput: %v", err))
				return
			}
			if err := ctx.ExposeComponent("taskStats", taskStats); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose taskStats: %v", err))
				return
			}
			if err := ctx.ExposeComponent("helpPanel", helpPanel); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose helpPanel: %v", err))
				return
			}

			// Lifecycle hooks
			ctx.OnMounted(func() {
				// App mounted - DevTools will show this in component tree
			})
		}).
		// Template receives RenderContext (no pointer!)
		Template(func(ctx bubbly.RenderContext) string {
			// Get child components
			taskList := ctx.Get("taskList").(bubbly.Component)
			taskInput := ctx.Get("taskInput").(bubbly.Component)
			taskStats := ctx.Get("taskStats").(bubbly.Component)
			helpPanel := ctx.Get("helpPanel").(bubbly.Component)

			// Create title
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("99")).
				MarginBottom(1)
			title := titleStyle.Render("BubblyUI Quickstart - Task Manager")

			// Subtitle with feature highlights
			subtitleStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				MarginBottom(1)
			subtitle := subtitleStyle.Render("Demonstrating: Clean Imports | Components | Composables | DevTools | Profiler")

			// Layout: Title, Subtitle, Stats, List, Input, Help
			content := lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				subtitle,
				"",
				taskStats.View(),
				"",
				taskList.View(),
				"",
				taskInput.View(),
				"",
				strings.Repeat("-", 70),
				helpPanel.View(),
			)

			// Add padding
			containerStyle := lipgloss.NewStyle().Padding(1, 2)
			return containerStyle.Render(content)
		}).
		Build()
}
