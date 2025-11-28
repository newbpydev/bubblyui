package composables

import (
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// QueueReturn is the return value of UseQueue.
// It provides FIFO (First-In-First-Out) queue operations with reactive state tracking.
//
// The queue is stored as a slice in the Items ref, which can be watched for changes.
// Size, IsEmpty, and Front are computed values that automatically update when Items changes.
//
// Thread Safety:
// All methods are thread-safe and can be called concurrently.
// The Items ref is updated atomically with the internal state.
type QueueReturn[T any] struct {
	// Items is the queue items (front at index 0, back at end).
	// This is a reactive ref that can be watched for changes.
	Items *bubbly.Ref[[]T]

	// Size is the item count (computed).
	// This is a computed value that automatically updates when Items changes.
	Size *bubbly.Computed[int]

	// IsEmpty indicates if queue is empty (computed).
	// This is a computed value that automatically updates when Items changes.
	IsEmpty *bubbly.Computed[bool]

	// Front is the first item in the queue (computed).
	// Returns nil if the queue is empty.
	// This is a computed value that automatically updates when Items changes.
	Front *bubbly.Computed[*T]

	// mu protects internal operations
	mu sync.Mutex
}

// Enqueue adds an item to the back of the queue.
//
// Example:
//
//	queue := UseQueue(ctx, []int{1, 2})
//	queue.Enqueue(3)  // queue=[1, 2, 3]
func (q *QueueReturn[T]) Enqueue(item T) {
	q.mu.Lock()
	defer q.mu.Unlock()

	current := q.Items.GetTyped()
	newItems := append(current, item)
	q.Items.Set(newItems)
}

// Dequeue removes and returns the front item from the queue.
// Returns the item and true if successful, or zero value and false if the queue is empty.
//
// Example:
//
//	queue := UseQueue(ctx, []string{"a", "b", "c"})
//	item, ok := queue.Dequeue()  // item="a", ok=true, queue=["b", "c"]
//	item, ok = queue.Dequeue()   // item="b", ok=true, queue=["c"]
func (q *QueueReturn[T]) Dequeue() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	current := q.Items.GetTyped()
	if len(current) == 0 {
		var zero T
		return zero, false
	}

	item := current[0]
	q.Items.Set(current[1:])
	return item, true
}

// Peek returns the front item without removing it.
// Returns the item and true if successful, or zero value and false if the queue is empty.
//
// Example:
//
//	queue := UseQueue(ctx, []int{1, 2, 3})
//	item, ok := queue.Peek()  // item=1, ok=true, queue unchanged
//	item, ok = queue.Peek()   // item=1, ok=true, queue unchanged (idempotent)
func (q *QueueReturn[T]) Peek() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	current := q.Items.GetTyped()
	if len(current) == 0 {
		var zero T
		return zero, false
	}

	return current[0], true
}

// Clear removes all items from the queue.
//
// Example:
//
//	queue := UseQueue(ctx, []int{1, 2, 3})
//	queue.Clear()  // queue=[]
func (q *QueueReturn[T]) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.Items.Set([]T{})
}

// UseQueue creates a FIFO queue composable.
// It provides queue operations with reactive state tracking.
//
// This composable is useful for:
//   - Managing task queues (job processing)
//   - Message queues (chat, notifications)
//   - Event queues (buffering events)
//   - Breadth-first traversal patterns
//
// Parameters:
//   - ctx: The component context (can be nil for testing)
//   - initial: The initial queue items (nil is treated as empty queue)
//
// Returns:
//   - *QueueReturn[T]: A struct containing the reactive Items ref and computed Size/IsEmpty/Front
//
// Example - Task queue:
//
//	Setup(func(ctx *bubbly.Context) {
//	    tasks := composables.UseQueue(ctx, []Task{})
//	    ctx.Expose("tasks", tasks)
//
//	    ctx.On("addTask", func(data interface{}) {
//	        task := data.(Task)
//	        tasks.Enqueue(task)
//	    })
//
//	    ctx.On("processNext", func(_ interface{}) {
//	        task, ok := tasks.Dequeue()
//	        if ok {
//	            processTask(task)
//	        }
//	    })
//	})
//
// Example - Message buffer:
//
//	Setup(func(ctx *bubbly.Context) {
//	    messages := composables.UseQueue(ctx, []Message{})
//	    ctx.Expose("messages", messages)
//
//	    ctx.On("newMessage", func(data interface{}) {
//	        msg := data.(Message)
//	        messages.Enqueue(msg)
//	    })
//
//	    ctx.On("displayNext", func(_ interface{}) {
//	        msg, ok := messages.Peek()
//	        if ok {
//	            displayMessage(msg)
//	            messages.Dequeue()
//	        }
//	    })
//	})
//
// Template usage:
//
//	Template(func(ctx bubbly.RenderContext) string {
//	    tasks := ctx.Get("tasks").(*composables.QueueReturn[Task])
//
//	    if tasks.IsEmpty.GetTyped() {
//	        return "No tasks in queue"
//	    }
//
//	    front := tasks.Front.GetTyped()
//	    if front != nil {
//	        return fmt.Sprintf("Next task: %s (%d in queue)", front.Name, tasks.Size.GetTyped())
//	    }
//	    return ""
//	})
//
// Integration with CreateShared:
//
//	var UseSharedTaskQueue = composables.CreateShared(
//	    func(ctx *bubbly.Context) *composables.QueueReturn[Task] {
//	        return composables.UseQueue(ctx, []Task{})
//	    },
//	)
//
// Thread Safety:
//
// UseQueue is thread-safe. All operations are synchronized with a mutex.
// The Items ref can be safely accessed from multiple goroutines.
func UseQueue[T any](ctx *bubbly.Context, initial []T) *QueueReturn[T] {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseQueue", time.Since(start))
	}()

	// Normalize nil to empty slice
	if initial == nil {
		initial = []T{}
	}

	// Create Items ref
	itemsRef := bubbly.NewRef(initial)

	// Create computed values
	size := bubbly.NewComputed(func() int {
		return len(itemsRef.GetTyped())
	})

	isEmpty := bubbly.NewComputed(func() bool {
		return len(itemsRef.GetTyped()) == 0
	})

	front := bubbly.NewComputed(func() *T {
		items := itemsRef.GetTyped()
		if len(items) == 0 {
			return nil
		}
		// Return a copy to avoid aliasing
		val := items[0]
		return &val
	})

	return &QueueReturn[T]{
		Items:   itemsRef,
		Size:    size,
		IsEmpty: isEmpty,
		Front:   front,
	}
}
