// Package composables provides reusable reactive logic for the quickstart example.
package composables

import (
	"time"

	// Clean import using alias package
	"github.com/newbpydev/bubblyui"
)

// Task represents a single todo item.
type Task struct {
	ID        int
	Text      string
	Done      bool
	CreatedAt time.Time
}

// TasksComposable encapsulates task management logic (Vue-like composable).
// This demonstrates the composable pattern - reusable logic shared across components.
type TasksComposable struct {
	Tasks       *bubblyui.Ref[[]Task]
	NextID      *bubblyui.Ref[int]
	AddTask     func(text string)
	ToggleTask  func(id int)
	DeleteTask  func(id int)
	ClearDone   func()
	ActiveCount func() int
	DoneCount   func() int
	TotalCount  func() int
	GetFiltered func(filter string) []Task
}

// UseTasks creates a reusable task manager with reactive state.
// This is the Vue Composition API pattern adapted for Go TUI.
// Uses type-safe bubblyui.NewRef[T]() instead of ctx.Ref() (interface{}).
// IMPORTANT: Use GetTyped() for type-safe access, not Get() which returns any.
func UseTasks(initial []Task) *TasksComposable {
	// Create TYPE-SAFE reactive state for tasks (PREFERRED pattern)
	tasks := bubblyui.NewRef(initial)

	// Track next ID for new tasks
	nextID := bubblyui.NewRef(len(initial) + 1)

	// Add a new task
	addTask := func(text string) {
		if text == "" {
			return
		}
		// Use GetTyped() for type-safe access
		currentTasks := tasks.GetTyped()
		id := nextID.GetTyped()
		newTask := Task{
			ID:        id,
			Text:      text,
			Done:      false,
			CreatedAt: time.Now(),
		}
		tasks.Set(append(currentTasks, newTask))
		nextID.Set(id + 1)
	}

	// Toggle task completion
	toggleTask := func(id int) {
		currentTasks := tasks.GetTyped()
		updated := make([]Task, len(currentTasks))
		for i, t := range currentTasks {
			if t.ID == id {
				t.Done = !t.Done
			}
			updated[i] = t
		}
		tasks.Set(updated)
	}

	// Delete a task
	deleteTask := func(id int) {
		currentTasks := tasks.GetTyped()
		updated := make([]Task, 0, len(currentTasks))
		for _, t := range currentTasks {
			if t.ID != id {
				updated = append(updated, t)
			}
		}
		tasks.Set(updated)
	}

	// Clear all completed tasks
	clearDone := func() {
		currentTasks := tasks.GetTyped()
		updated := make([]Task, 0, len(currentTasks))
		for _, t := range currentTasks {
			if !t.Done {
				updated = append(updated, t)
			}
		}
		tasks.Set(updated)
	}

	// Count helpers
	activeCount := func() int {
		count := 0
		for _, t := range tasks.GetTyped() {
			if !t.Done {
				count++
			}
		}
		return count
	}

	doneCount := func() int {
		count := 0
		for _, t := range tasks.GetTyped() {
			if t.Done {
				count++
			}
		}
		return count
	}

	totalCount := func() int {
		return len(tasks.GetTyped())
	}

	// Get filtered tasks based on filter type
	getFiltered := func(filter string) []Task {
		currentTasks := tasks.GetTyped()
		switch filter {
		case "active":
			result := make([]Task, 0)
			for _, t := range currentTasks {
				if !t.Done {
					result = append(result, t)
				}
			}
			return result
		case "done":
			result := make([]Task, 0)
			for _, t := range currentTasks {
				if t.Done {
					result = append(result, t)
				}
			}
			return result
		default: // "all"
			return currentTasks
		}
	}

	return &TasksComposable{
		Tasks:       tasks,
		NextID:      nextID,
		AddTask:     addTask,
		ToggleTask:  toggleTask,
		DeleteTask:  deleteTask,
		ClearDone:   clearDone,
		ActiveCount: activeCount,
		DoneCount:   doneCount,
		TotalCount:  totalCount,
		GetFiltered: getFiltered,
	}
}
