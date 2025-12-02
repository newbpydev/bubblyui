package composables

import (
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// SetReturn is the return value of UseSet.
// It provides unique value set management with reactive state tracking.
//
// The set is stored internally as a map[T]struct{} for O(1) operations.
// Size and IsEmpty are computed values that automatically update when Values changes.
//
// Thread Safety:
// All methods are thread-safe and can be called concurrently.
// The Values ref is updated atomically with the internal state.
type SetReturn[T comparable] struct {
	// Values is the set values stored as a map for O(1) operations.
	// This is a reactive ref that can be watched for changes.
	Values *bubbly.Ref[map[T]struct{}]

	// Size is the value count (computed).
	// This is a computed value that automatically updates when Values changes.
	Size *bubbly.Computed[int]

	// IsEmpty indicates if set is empty (computed).
	// This is a computed value that automatically updates when Values changes.
	IsEmpty *bubbly.Computed[bool]

	// mu protects internal operations
	mu sync.Mutex
}

// Add adds a value to the set.
// If the value already exists, this is a no-op (set semantics).
//
// Example:
//
//	s := UseSet(ctx, []string{"a", "b"})
//	s.Add("c")  // set={"a", "b", "c"}
//	s.Add("a")  // set={"a", "b", "c"} (no change, "a" already exists)
func (s *SetReturn[T]) Add(value T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a copy to avoid mutating the original
	current := s.Values.GetTyped()
	newSet := make(map[T]struct{}, len(current)+1)
	for k := range current {
		newSet[k] = struct{}{}
	}
	newSet[value] = struct{}{}
	s.Values.Set(newSet)
}

// Delete removes a value from the set.
// Returns true if the value was present and removed, false if the value did not exist.
//
// Example:
//
//	s := UseSet(ctx, []string{"a", "b", "c"})
//	ok := s.Delete("b")  // ok=true, set={"a", "c"}
//	ok = s.Delete("x")   // ok=false, set unchanged
func (s *SetReturn[T]) Delete(value T) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	current := s.Values.GetTyped()

	// Check if value exists
	if _, exists := current[value]; !exists {
		return false
	}

	// Create a copy without the value
	newSet := make(map[T]struct{}, len(current)-1)
	for k := range current {
		if k != value {
			newSet[k] = struct{}{}
		}
	}
	s.Values.Set(newSet)
	return true
}

// Has returns true if the value exists in the set.
//
// Example:
//
//	s := UseSet(ctx, []string{"a", "b"})
//	s.Has("a")  // true
//	s.Has("c")  // false
func (s *SetReturn[T]) Has(value T) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	data := s.Values.GetTyped()
	_, exists := data[value]
	return exists
}

// Toggle adds the value if not present, removes it if present.
// This is useful for implementing checkbox-like behavior.
//
// Example:
//
//	s := UseSet(ctx, []string{"a", "b"})
//	s.Toggle("c")  // set={"a", "b", "c"} (added)
//	s.Toggle("b")  // set={"a", "c"} (removed)
func (s *SetReturn[T]) Toggle(value T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current := s.Values.GetTyped()

	if _, exists := current[value]; exists {
		// Remove value
		newSet := make(map[T]struct{}, len(current)-1)
		for k := range current {
			if k != value {
				newSet[k] = struct{}{}
			}
		}
		s.Values.Set(newSet)
	} else {
		// Add value
		newSet := make(map[T]struct{}, len(current)+1)
		for k := range current {
			newSet[k] = struct{}{}
		}
		newSet[value] = struct{}{}
		s.Values.Set(newSet)
	}
}

// Clear removes all values from the set.
//
// Example:
//
//	s := UseSet(ctx, []string{"a", "b", "c"})
//	s.Clear()  // set={}
func (s *SetReturn[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Values.Set(make(map[T]struct{}))
}

// ToSlice returns all values in the set as a slice.
// The order of values is not guaranteed (map iteration order).
//
// Example:
//
//	s := UseSet(ctx, []string{"a", "b", "c"})
//	slice := s.ToSlice()  // []string{"a", "b", "c"} (order may vary)
func (s *SetReturn[T]) ToSlice() []T {
	s.mu.Lock()
	defer s.mu.Unlock()

	data := s.Values.GetTyped()
	slice := make([]T, 0, len(data))
	for k := range data {
		slice = append(slice, k)
	}
	return slice
}

// UseSet creates a set management composable.
// It provides unique value management with O(1) add, delete, and has operations.
//
// This composable is useful for:
//   - Managing selected items in a multi-select list
//   - Tracking active tags or filters
//   - Implementing checkbox groups
//   - Deduplicating values
//
// Parameters:
//   - ctx: The component context (can be nil for testing)
//   - initial: The initial values (nil is treated as empty set, duplicates are ignored)
//
// Returns:
//   - *SetReturn[T]: A struct containing the reactive Values ref and computed Size/IsEmpty
//
// Example - Tag management:
//
//	Setup(func(ctx *bubbly.Context) {
//	    tags := composables.UseSet(ctx, []string{"urgent", "todo"})
//	    ctx.Expose("tags", tags)
//
//	    ctx.On("toggleTag", func(data interface{}) {
//	        tag := data.(string)
//	        tags.Toggle(tag)
//	    })
//
//	    ctx.On("clearTags", func(_ interface{}) {
//	        tags.Clear()
//	    })
//	})
//
// Example - Multi-select:
//
//	Setup(func(ctx *bubbly.Context) {
//	    selected := composables.UseSet(ctx, []int{})
//	    ctx.Expose("selected", selected)
//
//	    ctx.On("toggleSelection", func(data interface{}) {
//	        id := data.(int)
//	        selected.Toggle(id)
//	    })
//
//	    ctx.On("selectAll", func(data interface{}) {
//	        ids := data.([]int)
//	        for _, id := range ids {
//	            selected.Add(id)
//	        }
//	    })
//	})
//
// Template usage:
//
//	Template(func(ctx bubbly.RenderContext) string {
//	    tags := ctx.Get("tags").(*composables.SetReturn[string])
//
//	    if tags.IsEmpty.GetTyped() {
//	        return "No tags selected"
//	    }
//
//	    return fmt.Sprintf("Selected: %v", tags.ToSlice())
//	})
//
// Integration with CreateShared:
//
//	var UseSharedSelectedTags = composables.CreateShared(
//	    func(ctx *bubbly.Context) *composables.SetReturn[string] {
//	        return composables.UseSet(ctx, []string{})
//	    },
//	)
//
// Thread Safety:
//
// UseSet is thread-safe. All operations are synchronized with a mutex.
// The Values ref can be safely accessed from multiple goroutines.
func UseSet[T comparable](ctx *bubbly.Context, initial []T) *SetReturn[T] {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseSet", time.Since(start))
	}()

	// Convert slice to map (set), handling duplicates
	// Note: ranging over nil slice is safe in Go (no iteration occurs)
	setData := make(map[T]struct{})
	for _, v := range initial {
		setData[v] = struct{}{}
	}

	// Create Values ref
	valuesRef := bubbly.NewRef(setData)

	// Create computed values
	size := bubbly.NewComputed(func() int {
		return len(valuesRef.GetTyped())
	})

	isEmpty := bubbly.NewComputed(func() bool {
		return len(valuesRef.GetTyped()) == 0
	})

	return &SetReturn[T]{
		Values:  valuesRef,
		Size:    size,
		IsEmpty: isEmpty,
	}
}
