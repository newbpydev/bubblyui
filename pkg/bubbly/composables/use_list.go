package composables

import (
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// ListReturn is the return value of UseList.
// It provides generic list management with CRUD operations and reactive state.
//
// The list is stored as a slice in the Items ref, which can be watched for changes.
// Length and IsEmpty are computed values that automatically update when Items changes.
//
// Thread Safety:
// All methods are thread-safe and can be called concurrently.
// The Items ref is updated atomically with the internal state.
type ListReturn[T any] struct {
	// Items is the list of items.
	// This is a reactive ref that can be watched for changes.
	Items *bubbly.Ref[[]T]

	// Length is the item count (computed).
	// This is a computed value that automatically updates when Items changes.
	Length *bubbly.Computed[int]

	// IsEmpty indicates if list is empty (computed).
	// This is a computed value that automatically updates when Items changes.
	IsEmpty *bubbly.Computed[bool]

	// mu protects internal operations
	mu sync.Mutex
}

// Push adds items to the end of the list.
// Multiple items can be added in a single call.
//
// Example:
//
//	list := UseList(ctx, []int{1, 2})
//	list.Push(3)        // [1, 2, 3]
//	list.Push(4, 5, 6)  // [1, 2, 3, 4, 5, 6]
func (l *ListReturn[T]) Push(items ...T) {
	l.mu.Lock()
	defer l.mu.Unlock()

	current := l.Items.GetTyped()
	newItems := append(current, items...)
	l.Items.Set(newItems)
}

// Pop removes and returns the last item from the list.
// Returns the item and true if successful, or zero value and false if the list is empty.
//
// Example:
//
//	list := UseList(ctx, []string{"a", "b", "c"})
//	item, ok := list.Pop()  // item="c", ok=true, list=["a", "b"]
//	item, ok = list.Pop()   // item="b", ok=true, list=["a"]
func (l *ListReturn[T]) Pop() (T, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	current := l.Items.GetTyped()
	if len(current) == 0 {
		var zero T
		return zero, false
	}

	lastIdx := len(current) - 1
	item := current[lastIdx]
	l.Items.Set(current[:lastIdx])
	return item, true
}

// Shift removes and returns the first item from the list.
// Returns the item and true if successful, or zero value and false if the list is empty.
//
// Example:
//
//	list := UseList(ctx, []int{1, 2, 3})
//	item, ok := list.Shift()  // item=1, ok=true, list=[2, 3]
func (l *ListReturn[T]) Shift() (T, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	current := l.Items.GetTyped()
	if len(current) == 0 {
		var zero T
		return zero, false
	}

	item := current[0]
	l.Items.Set(current[1:])
	return item, true
}

// Unshift adds items to the beginning of the list.
// Multiple items can be added in a single call, preserving their order.
//
// Example:
//
//	list := UseList(ctx, []string{"c"})
//	list.Unshift("b")        // ["b", "c"]
//	list.Unshift("a")        // ["a", "b", "c"]
//	list.Unshift("x", "y")   // ["x", "y", "a", "b", "c"]
func (l *ListReturn[T]) Unshift(items ...T) {
	l.mu.Lock()
	defer l.mu.Unlock()

	current := l.Items.GetTyped()
	newItems := append(items, current...)
	l.Items.Set(newItems)
}

// Insert adds an item at the specified index.
// If index is negative, it is clamped to 0.
// If index is greater than the list length, the item is appended to the end.
//
// Example:
//
//	list := UseList(ctx, []int{1, 3})
//	list.Insert(1, 2)  // [1, 2, 3]
//	list.Insert(0, 0)  // [0, 1, 2, 3]
func (l *ListReturn[T]) Insert(index int, item T) {
	l.mu.Lock()
	defer l.mu.Unlock()

	current := l.Items.GetTyped()

	// Clamp index to valid range
	if index < 0 {
		index = 0
	}
	if index > len(current) {
		index = len(current)
	}

	// Insert at index
	newItems := make([]T, 0, len(current)+1)
	newItems = append(newItems, current[:index]...)
	newItems = append(newItems, item)
	newItems = append(newItems, current[index:]...)
	l.Items.Set(newItems)
}

// RemoveAt removes and returns the item at the specified index.
// Returns the item and true if successful, or zero value and false if the index is out of bounds.
//
// Example:
//
//	list := UseList(ctx, []string{"a", "b", "c"})
//	item, ok := list.RemoveAt(1)  // item="b", ok=true, list=["a", "c"]
//	item, ok = list.RemoveAt(10)  // item="", ok=false, list unchanged
func (l *ListReturn[T]) RemoveAt(index int) (T, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	current := l.Items.GetTyped()

	// Check bounds
	if index < 0 || index >= len(current) {
		var zero T
		return zero, false
	}

	item := current[index]
	newItems := make([]T, 0, len(current)-1)
	newItems = append(newItems, current[:index]...)
	newItems = append(newItems, current[index+1:]...)
	l.Items.Set(newItems)
	return item, true
}

// UpdateAt updates the item at the specified index.
// Returns true if successful, or false if the index is out of bounds.
//
// Example:
//
//	list := UseList(ctx, []int{1, 2, 3})
//	ok := list.UpdateAt(1, 20)  // ok=true, list=[1, 20, 3]
//	ok = list.UpdateAt(10, 99)  // ok=false, list unchanged
func (l *ListReturn[T]) UpdateAt(index int, item T) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	current := l.Items.GetTyped()

	// Check bounds
	if index < 0 || index >= len(current) {
		return false
	}

	// Create a copy to avoid modifying the original slice
	newItems := make([]T, len(current))
	copy(newItems, current)
	newItems[index] = item
	l.Items.Set(newItems)
	return true
}

// Remove removes the first occurrence of an item that matches according to the equality function.
// Returns true if an item was removed, or false if no matching item was found.
//
// Example:
//
//	list := UseList(ctx, []string{"a", "b", "a", "c"})
//	eq := func(a, b string) bool { return a == b }
//	ok := list.Remove("a", eq)  // ok=true, list=["b", "a", "c"]
//	ok = list.Remove("x", eq)   // ok=false, list unchanged
func (l *ListReturn[T]) Remove(item T, eq func(a, b T) bool) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	current := l.Items.GetTyped()

	// Find first matching item
	for i, v := range current {
		if eq(v, item) {
			newItems := make([]T, 0, len(current)-1)
			newItems = append(newItems, current[:i]...)
			newItems = append(newItems, current[i+1:]...)
			l.Items.Set(newItems)
			return true
		}
	}

	return false
}

// Clear removes all items from the list.
//
// Example:
//
//	list := UseList(ctx, []int{1, 2, 3})
//	list.Clear()  // list=[]
func (l *ListReturn[T]) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.Items.Set([]T{})
}

// Get returns the item at the specified index.
// Returns the item and true if successful, or zero value and false if the index is out of bounds.
//
// Example:
//
//	list := UseList(ctx, []string{"a", "b", "c"})
//	item, ok := list.Get(1)   // item="b", ok=true
//	item, ok = list.Get(10)   // item="", ok=false
func (l *ListReturn[T]) Get(index int) (T, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	current := l.Items.GetTyped()

	// Check bounds
	if index < 0 || index >= len(current) {
		var zero T
		return zero, false
	}

	return current[index], true
}

// Set replaces the entire list with new items.
// If items is nil, the list is set to an empty slice.
//
// Example:
//
//	list := UseList(ctx, []int{1, 2, 3})
//	list.Set([]int{4, 5})  // list=[4, 5]
//	list.Set(nil)          // list=[]
func (l *ListReturn[T]) Set(items []T) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if items == nil {
		items = []T{}
	}
	l.Items.Set(items)
}

// UseList creates a list management composable.
// It provides a generic list with CRUD operations and reactive state tracking.
//
// This composable is useful for:
//   - Managing dynamic lists of items (todos, messages, etc.)
//   - Implementing list-based UI components
//   - Tracking list state with reactive updates
//   - Building data-driven interfaces
//
// Parameters:
//   - ctx: The component context (can be nil for testing)
//   - initial: The initial list of items (nil is treated as empty slice)
//
// Returns:
//   - *ListReturn[T]: A struct containing the reactive Items ref and computed Length/IsEmpty
//
// Example - Todo list:
//
//	Setup(func(ctx *bubbly.Context) {
//	    todos := composables.UseList(ctx, []Todo{})
//	    ctx.Expose("todos", todos)
//
//	    ctx.On("addTodo", func(data interface{}) {
//	        todo := data.(Todo)
//	        todos.Push(todo)
//	    })
//
//	    ctx.On("removeTodo", func(data interface{}) {
//	        idx := data.(int)
//	        todos.RemoveAt(idx)
//	    })
//	})
//
// Example - Message queue:
//
//	Setup(func(ctx *bubbly.Context) {
//	    messages := composables.UseList(ctx, []Message{})
//	    ctx.Expose("messages", messages)
//
//	    ctx.On("newMessage", func(data interface{}) {
//	        msg := data.(Message)
//	        messages.Push(msg)
//	    })
//
//	    ctx.On("processMessage", func(_ interface{}) {
//	        msg, ok := messages.Shift()
//	        if ok {
//	            handleMessage(msg)
//	        }
//	    })
//	})
//
// Template usage:
//
//	Template(func(ctx bubbly.RenderContext) string {
//	    todos := ctx.Get("todos").(*composables.ListReturn[Todo])
//	    items := todos.Items.GetTyped()
//
//	    if todos.IsEmpty.GetTyped() {
//	        return "No items"
//	    }
//
//	    var lines []string
//	    for i, item := range items {
//	        lines = append(lines, fmt.Sprintf("%d. %s", i+1, item.Title))
//	    }
//	    return strings.Join(lines, "\n")
//	})
//
// Integration with CreateShared:
//
//	var UseSharedTodos = composables.CreateShared(
//	    func(ctx *bubbly.Context) *composables.ListReturn[Todo] {
//	        return composables.UseList(ctx, []Todo{})
//	    },
//	)
//
// Thread Safety:
//
// UseList is thread-safe. All operations are synchronized with a mutex.
// The Items ref can be safely accessed from multiple goroutines.
func UseList[T any](ctx *bubbly.Context, initial []T) *ListReturn[T] {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseList", time.Since(start))
	}()

	// Normalize nil to empty slice
	if initial == nil {
		initial = []T{}
	}

	// Create Items ref
	itemsRef := bubbly.NewRef(initial)

	// Create computed values
	length := bubbly.NewComputed(func() int {
		return len(itemsRef.GetTyped())
	})

	isEmpty := bubbly.NewComputed(func() bool {
		return len(itemsRef.GetTyped()) == 0
	})

	return &ListReturn[T]{
		Items:   itemsRef,
		Length:  length,
		IsEmpty: isEmpty,
	}
}
