package composables

import (
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// Todo represents a single todo item
type Todo struct {
	ID        int64
	Title     string
	Completed bool
	CreatedAt time.Time
}

// TodosComposable provides reactive todo list management
type TodosComposable struct {
	Todos     *bubbly.Ref[interface{}]
	Total     *bubbly.Computed[interface{}]
	Completed *bubbly.Computed[interface{}]
	Remaining *bubbly.Computed[interface{}]
	AllDone   *bubbly.Computed[interface{}]
	Add       func(title string)
	Toggle    func(id int64)
	Remove    func(id int64)
	Clear     func()
	ToggleAll func()
}

// UseTodos creates a new todos composable with reactive state management
// Returns a composable with refs, computed values, and action methods
func UseTodos(ctx *bubbly.Context, initial []Todo) *TodosComposable {
	// Initialize with provided todos or empty slice
	if initial == nil {
		initial = []Todo{}
	}

	// Create reactive state using ctx.Ref for interface{} compatibility
	todos := ctx.Ref(initial)

	// Computed: Total count
	total := ctx.Computed(func() interface{} {
		current := todos.Get().([]Todo)
		return len(current)
	})

	// Computed: Completed count
	completed := ctx.Computed(func() interface{} {
		current := todos.Get().([]Todo)
		count := 0
		for _, todo := range current {
			if todo.Completed {
				count++
			}
		}
		return count
	})

	// Computed: Remaining count
	remaining := ctx.Computed(func() interface{} {
		current := todos.Get().([]Todo)
		count := 0
		for _, todo := range current {
			if !todo.Completed {
				count++
			}
		}
		return count
	})

	// Computed: All done?
	allDone := ctx.Computed(func() interface{} {
		current := todos.Get().([]Todo)
		if len(current) == 0 {
			return false
		}
		for _, todo := range current {
			if !todo.Completed {
				return false
			}
		}
		return true
	})

	// Action: Add new todo
	add := func(title string) {
		if title == "" {
			return
		}

		current := todos.Get().([]Todo)
		newTodo := Todo{
			ID:        time.Now().UnixNano(),
			Title:     title,
			Completed: false,
			CreatedAt: time.Now(),
		}
		todos.Set(append(current, newTodo))
	}

	// Action: Toggle todo completion
	toggle := func(id int64) {
		current := todos.Get().([]Todo)
		updated := make([]Todo, len(current))
		copy(updated, current)

		for i := range updated {
			if updated[i].ID == id {
				updated[i].Completed = !updated[i].Completed
				break
			}
		}
		todos.Set(updated)
	}

	// Action: Remove todo
	remove := func(id int64) {
		current := todos.Get().([]Todo)
		filtered := []Todo{}
		for _, todo := range current {
			if todo.ID != id {
				filtered = append(filtered, todo)
			}
		}
		todos.Set(filtered)
	}

	// Action: Clear all todos
	clear := func() {
		todos.Set([]Todo{})
	}

	// Action: Toggle all todos (mark all complete/incomplete)
	toggleAll := func() {
		current := todos.Get().([]Todo)
		if len(current) == 0 {
			return
		}

		// Check if all are completed
		allCompleted := true
		for _, todo := range current {
			if !todo.Completed {
				allCompleted = false
				break
			}
		}

		// Toggle all to opposite state
		updated := make([]Todo, len(current))
		copy(updated, current)
		for i := range updated {
			updated[i].Completed = !allCompleted
		}
		todos.Set(updated)
	}

	return &TodosComposable{
		Todos:     todos,
		Total:     total,
		Completed: completed,
		Remaining: remaining,
		AllDone:   allDone,
		Add:       add,
		Toggle:    toggle,
		Remove:    remove,
		Clear:     clear,
		ToggleAll: toggleAll,
	}
}
