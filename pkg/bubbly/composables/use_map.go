package composables

import (
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// MapReturn is the return value of UseMap.
// It provides generic key-value state management with reactive state tracking.
//
// The map is stored in the Data ref, which can be watched for changes.
// Size and IsEmpty are computed values that automatically update when Data changes.
//
// Thread Safety:
// All methods are thread-safe and can be called concurrently.
// The Data ref is updated atomically with the internal state.
type MapReturn[K comparable, V any] struct {
	// Data is the map data.
	// This is a reactive ref that can be watched for changes.
	Data *bubbly.Ref[map[K]V]

	// Size is the entry count (computed).
	// This is a computed value that automatically updates when Data changes.
	Size *bubbly.Computed[int]

	// IsEmpty indicates if map is empty (computed).
	// This is a computed value that automatically updates when Data changes.
	IsEmpty *bubbly.Computed[bool]

	// mu protects internal operations
	mu sync.Mutex
}

// Get returns the value for the given key.
// Returns the value and true if the key exists, or zero value and false if not.
//
// Example:
//
//	m := UseMap(ctx, map[string]int{"a": 1, "b": 2})
//	val, ok := m.Get("a")  // val=1, ok=true
//	val, ok = m.Get("c")   // val=0, ok=false
func (m *MapReturn[K, V]) Get(key K) (V, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data := m.Data.GetTyped()
	val, ok := data[key]
	return val, ok
}

// Set sets the value for the given key.
// If the key already exists, its value is updated.
// If the key does not exist, it is added to the map.
//
// Example:
//
//	m := UseMap(ctx, map[string]int{})
//	m.Set("a", 1)   // map={"a": 1}
//	m.Set("a", 10)  // map={"a": 10} (updated)
//	m.Set("b", 2)   // map={"a": 10, "b": 2}
func (m *MapReturn[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a copy to avoid mutating the original
	current := m.Data.GetTyped()
	newMap := make(map[K]V, len(current)+1)
	for k, v := range current {
		newMap[k] = v
	}
	newMap[key] = value
	m.Data.Set(newMap)
}

// Delete removes the key from the map.
// Returns true if the key was present and removed, false if the key did not exist.
//
// Example:
//
//	m := UseMap(ctx, map[string]int{"a": 1, "b": 2})
//	ok := m.Delete("a")  // ok=true, map={"b": 2}
//	ok = m.Delete("c")   // ok=false, map unchanged
func (m *MapReturn[K, V]) Delete(key K) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	current := m.Data.GetTyped()

	// Check if key exists
	if _, exists := current[key]; !exists {
		return false
	}

	// Create a copy without the key
	newMap := make(map[K]V, len(current)-1)
	for k, v := range current {
		if k != key {
			newMap[k] = v
		}
	}
	m.Data.Set(newMap)
	return true
}

// Has returns true if the key exists in the map.
//
// Example:
//
//	m := UseMap(ctx, map[string]int{"a": 1})
//	m.Has("a")  // true
//	m.Has("b")  // false
func (m *MapReturn[K, V]) Has(key K) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	data := m.Data.GetTyped()
	_, exists := data[key]
	return exists
}

// Keys returns all keys in the map.
// The order of keys is not guaranteed (map iteration order).
//
// Example:
//
//	m := UseMap(ctx, map[string]int{"a": 1, "b": 2})
//	keys := m.Keys()  // []string{"a", "b"} (order may vary)
func (m *MapReturn[K, V]) Keys() []K {
	m.mu.Lock()
	defer m.mu.Unlock()

	data := m.Data.GetTyped()
	keys := make([]K, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys
}

// Values returns all values in the map.
// The order of values is not guaranteed (map iteration order).
//
// Example:
//
//	m := UseMap(ctx, map[string]int{"a": 1, "b": 2})
//	values := m.Values()  // []int{1, 2} (order may vary)
func (m *MapReturn[K, V]) Values() []V {
	m.mu.Lock()
	defer m.mu.Unlock()

	data := m.Data.GetTyped()
	values := make([]V, 0, len(data))
	for _, v := range data {
		values = append(values, v)
	}
	return values
}

// Clear removes all entries from the map.
//
// Example:
//
//	m := UseMap(ctx, map[string]int{"a": 1, "b": 2})
//	m.Clear()  // map={}
func (m *MapReturn[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Data.Set(make(map[K]V))
}

// UseMap creates a map management composable.
// It provides a generic key-value store with CRUD operations and reactive state tracking.
//
// This composable is useful for:
//   - Managing configuration settings
//   - Caching computed values by key
//   - Building lookup tables
//   - Tracking entity state by ID
//
// Parameters:
//   - ctx: The component context (can be nil for testing)
//   - initial: The initial map data (nil is treated as empty map)
//
// Returns:
//   - *MapReturn[K, V]: A struct containing the reactive Data ref and computed Size/IsEmpty
//
// Example - Configuration store:
//
//	Setup(func(ctx *bubbly.Context) {
//	    config := composables.UseMap(ctx, map[string]string{
//	        "theme": "dark",
//	        "lang": "en",
//	    })
//	    ctx.Expose("config", config)
//
//	    ctx.On("setConfig", func(data interface{}) {
//	        kv := data.([]string)
//	        config.Set(kv[0], kv[1])
//	    })
//	})
//
// Example - Entity cache:
//
//	Setup(func(ctx *bubbly.Context) {
//	    users := composables.UseMap(ctx, map[int]User{})
//	    ctx.Expose("users", users)
//
//	    ctx.On("loadUser", func(data interface{}) {
//	        user := data.(User)
//	        users.Set(user.ID, user)
//	    })
//
//	    ctx.On("removeUser", func(data interface{}) {
//	        id := data.(int)
//	        users.Delete(id)
//	    })
//	})
//
// Template usage:
//
//	Template(func(ctx bubbly.RenderContext) string {
//	    config := ctx.Get("config").(*composables.MapReturn[string, string])
//
//	    theme, _ := config.Get("theme")
//	    return fmt.Sprintf("Current theme: %s", theme)
//	})
//
// Integration with CreateShared:
//
//	var UseSharedConfig = composables.CreateShared(
//	    func(ctx *bubbly.Context) *composables.MapReturn[string, string] {
//	        return composables.UseMap(ctx, map[string]string{
//	            "theme": "dark",
//	        })
//	    },
//	)
//
// Thread Safety:
//
// UseMap is thread-safe. All operations are synchronized with a mutex.
// The Data ref can be safely accessed from multiple goroutines.
func UseMap[K comparable, V any](ctx *bubbly.Context, initial map[K]V) *MapReturn[K, V] {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseMap", time.Since(start))
	}()

	// Normalize nil to empty map
	if initial == nil {
		initial = make(map[K]V)
	}

	// Create a copy to avoid external mutation
	dataCopy := make(map[K]V, len(initial))
	for k, v := range initial {
		dataCopy[k] = v
	}

	// Create Data ref
	dataRef := bubbly.NewRef(dataCopy)

	// Create computed values
	size := bubbly.NewComputed(func() int {
		return len(dataRef.GetTyped())
	})

	isEmpty := bubbly.NewComputed(func() bool {
		return len(dataRef.GetTyped()) == 0
	})

	return &MapReturn[K, V]{
		Data:    dataRef,
		Size:    size,
		IsEmpty: isEmpty,
	}
}
