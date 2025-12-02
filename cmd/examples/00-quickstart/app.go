// Package main demonstrates the BubblyUI quickstart example.
// This example showcases clean import paths, component architecture,
// composables, DevTools integration, and profiler usage.
package main

import (
	"fmt"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
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
// - WithMessageHandler for text input capture
// - Exposing state for DevTools inspection
func CreateApp() (bubbly.Component, error) {
	// =============================================================================
	// Create refs OUTSIDE Setup so they're accessible in both Setup and MessageHandler
	// This is the pattern for handling text input in BubblyUI
	// =============================================================================
	selectedIndex := bubblyui.NewRef(0)
	filter := bubblyui.NewRef("all") // "all", "active", "done"
	inputText := bubblyui.NewRef("")
	inputMode := bubblyui.NewRef(false) // Whether we're typing in input

	// Initialize composables (these don't need Context)
	taskManager := composables.UseTasks([]composables.Task{
		{ID: 1, Text: "Learn BubblyUI basics", Done: true},
		{ID: 2, Text: "Build a task manager", Done: false},
		{ID: 3, Text: "Add DevTools support", Done: false},
	})
	focusManager := composables.UseFocusManager(composables.FocusList)

	return bubbly.NewComponent("TaskManager").
		// Auto-commands for automatic UI updates on Ref.Set()
		WithAutoCommands(true).
		// =============================================================================
		// CRITICAL: Use Conditional Key Bindings to disable when in input mode
		// This prevents "j", "k", etc. from firing when typing text
		// =============================================================================
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: "tab", Event: "cycleFocus", Description: "Switch focus",
			Condition: func() bool { return !inputMode.GetTyped() },
		}).
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: "j", Event: "moveDown", Description: "Move down",
			Condition: func() bool { return !inputMode.GetTyped() },
		}).
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: "k", Event: "moveUp", Description: "Move up",
			Condition: func() bool { return !inputMode.GetTyped() },
		}).
		WithKeyBinding("down", "moveDown", "Move down").
		WithKeyBinding("up", "moveUp", "Move up").
		WithKeyBinding("enter", "submitOrToggle", "Submit/Toggle").
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: " ", Event: "toggleTask", Description: "Toggle task",
			Condition: func() bool { return !inputMode.GetTyped() },
		}).
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: "a", Event: "addMode", Description: "Add task",
			Condition: func() bool { return !inputMode.GetTyped() },
		}).
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: "d", Event: "deleteTask", Description: "Delete task",
			Condition: func() bool { return !inputMode.GetTyped() },
		}).
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: "f", Event: "cycleFilter", Description: "Cycle filter",
			Condition: func() bool { return !inputMode.GetTyped() },
		}).
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: "c", Event: "clearDone", Description: "Clear done",
			Condition: func() bool { return !inputMode.GetTyped() },
		}).
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: "q", Event: "quit", Description: "Quit",
			Condition: func() bool { return !inputMode.GetTyped() },
		}).
		WithKeyBinding("ctrl+c", "quit", "Quit").
		WithKeyBinding("esc", "cancelInput", "Cancel input").
		// =============================================================================
		// WithMessageHandler: Capture text input when in input mode
		// This is how BubblyUI handles raw keyboard input for text fields
		// =============================================================================
		WithMessageHandler(func(_ bubbly.Component, msg tea.Msg) tea.Cmd {
			// Only process key messages when in input mode
			if !inputMode.GetTyped() {
				return nil
			}

			keyMsg, ok := msg.(tea.KeyMsg)
			if !ok {
				return nil
			}

			currentText := inputText.GetTyped()

			switch keyMsg.Type {
			case tea.KeyBackspace:
				// Delete last character
				if len(currentText) > 0 {
					// Handle UTF-8 properly
					runes := []rune(currentText)
					inputText.Set(string(runes[:len(runes)-1]))
				}
				return nil

			case tea.KeyRunes:
				// Append typed characters
				for _, r := range keyMsg.Runes {
					if unicode.IsPrint(r) {
						currentText += string(r)
					}
				}
				inputText.Set(currentText)
				return nil

			case tea.KeySpace:
				// Space key in input mode adds a space
				inputText.Set(currentText + " ")
				return nil
			}

			return nil
		}).
		Setup(func(ctx *bubbly.Context) {
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
				InputMode: inputMode, // Pass inputMode so component knows when focused
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
				Filter:      filter, // Pass filter for display
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

			// =============================================================================
			// Event Handlers - These are triggered by key bindings
			// Note: Conditional key bindings already check inputMode, so handlers are simpler
			// =============================================================================

			ctx.On("cycleFocus", func(_ interface{}) {
				focusManager.Next()
			})

			ctx.On("moveDown", func(_ interface{}) {
				if focusManager.IsListFocused() {
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
				if focusManager.IsListFocused() {
					idx := selectedIndex.GetTyped()
					if idx > 0 {
						selectedIndex.Set(idx - 1)
					}
				}
			})

			// Toggle task (when space is pressed in list mode - conditional binding handles mode check)
			ctx.On("toggleTask", func(_ interface{}) {
				if focusManager.IsListFocused() {
					filteredTasks := taskManager.GetFiltered(filter.GetTyped())
					if len(filteredTasks) > 0 {
						idx := selectedIndex.GetTyped()
						if idx >= 0 && idx < len(filteredTasks) {
							taskManager.ToggleTask(filteredTasks[idx].ID)
						}
					}
				}
			})

			// Submit or toggle - Enter key works in both modes
			ctx.On("submitOrToggle", func(_ interface{}) {
				if inputMode.GetTyped() {
					// Submit input
					text := inputText.GetTyped()
					if text != "" {
						taskManager.AddTask(text)
						inputText.Set("")
						focusManager.SetFocus(composables.FocusList)
						inputMode.Set(false)
					}
				} else if focusManager.IsListFocused() {
					// Toggle task
					filteredTasks := taskManager.GetFiltered(filter.GetTyped())
					if len(filteredTasks) > 0 {
						idx := selectedIndex.GetTyped()
						if idx >= 0 && idx < len(filteredTasks) {
							taskManager.ToggleTask(filteredTasks[idx].ID)
						}
					}
				}
			})

			ctx.On("addMode", func(_ interface{}) {
				focusManager.SetFocus(composables.FocusInput)
				inputMode.Set(true)
			})

			ctx.On("deleteTask", func(_ interface{}) {
				if focusManager.IsListFocused() {
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
			})

			ctx.On("clearDone", func(_ interface{}) {
				taskManager.ClearDone()
				selectedIndex.Set(0)
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
