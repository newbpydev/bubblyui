package reflectcache

import (
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test struct types for caching
type TestForm struct {
	Name  string
	Email string
	Age   int
}

type SimpleStruct struct {
	Value int
}

type NestedStruct struct {
	Inner SimpleStruct
	Data  string
}

// TestFieldCache_NewFieldCache tests cache creation
func TestFieldCache_NewFieldCache(t *testing.T) {
	cache := NewFieldCache()

	require.NotNil(t, cache, "NewFieldCache should return non-nil cache")
	require.NotNil(t, cache.cache, "cache.cache should be initialized")
}

// TestFieldCache_CacheType tests caching a struct type
func TestFieldCache_CacheType(t *testing.T) {
	cache := NewFieldCache()

	formType := reflect.TypeOf(TestForm{})
	entry := cache.CacheType(formType)

	require.NotNil(t, entry, "CacheType should return non-nil entry")
	require.NotNil(t, entry.Indices, "entry.Indices should be initialized")
	require.NotNil(t, entry.Types, "entry.Types should be initialized")

	// Verify all fields are cached
	assert.Equal(t, 3, len(entry.Indices), "Should cache all 3 fields")
	assert.Equal(t, 3, len(entry.Types), "Should cache all 3 field types")

	// Verify field indices
	nameIdx, ok := entry.Indices["Name"]
	require.True(t, ok, "Should have Name field")
	assert.Equal(t, 0, nameIdx, "Name should be at index 0")

	emailIdx, ok := entry.Indices["Email"]
	require.True(t, ok, "Should have Email field")
	assert.Equal(t, 1, emailIdx, "Email should be at index 1")

	ageIdx, ok := entry.Indices["Age"]
	require.True(t, ok, "Should have Age field")
	assert.Equal(t, 2, ageIdx, "Age should be at index 2")

	// Verify field types
	nameType, ok := entry.Types["Name"]
	require.True(t, ok, "Should have Name type")
	assert.Equal(t, reflect.TypeOf(""), nameType, "Name should be string type")

	emailType, ok := entry.Types["Email"]
	require.True(t, ok, "Should have Email type")
	assert.Equal(t, reflect.TypeOf(""), emailType, "Email should be string type")

	ageType, ok := entry.Types["Age"]
	require.True(t, ok, "Should have Age type")
	assert.Equal(t, reflect.TypeOf(0), ageType, "Age should be int type")
}

// TestFieldCache_GetFieldIndex tests retrieving cached field indices
func TestFieldCache_GetFieldIndex(t *testing.T) {
	cache := NewFieldCache()
	formType := reflect.TypeOf(TestForm{})

	// First call - cache miss, should populate cache
	idx, ok := cache.GetFieldIndex(formType, "Name")
	assert.True(t, ok, "Should find Name field")
	assert.Equal(t, 0, idx, "Name should be at index 0")

	// Second call - cache hit
	idx, ok = cache.GetFieldIndex(formType, "Email")
	assert.True(t, ok, "Should find Email field")
	assert.Equal(t, 1, idx, "Email should be at index 1")

	// Invalid field
	idx, ok = cache.GetFieldIndex(formType, "Invalid")
	assert.False(t, ok, "Should not find invalid field")
	assert.Equal(t, -1, idx, "Should return -1 for invalid field")
}

// TestFieldCache_GetFieldType tests retrieving cached field types
func TestFieldCache_GetFieldType(t *testing.T) {
	cache := NewFieldCache()
	formType := reflect.TypeOf(TestForm{})

	// Get field type
	fieldType, ok := cache.GetFieldType(formType, "Name")
	assert.True(t, ok, "Should find Name field type")
	assert.Equal(t, reflect.TypeOf(""), fieldType, "Name should be string type")

	// Get another field type
	fieldType, ok = cache.GetFieldType(formType, "Age")
	assert.True(t, ok, "Should find Age field type")
	assert.Equal(t, reflect.TypeOf(0), fieldType, "Age should be int type")

	// Invalid field
	fieldType, ok = cache.GetFieldType(formType, "Invalid")
	assert.False(t, ok, "Should not find invalid field type")
	assert.Nil(t, fieldType, "Should return nil for invalid field")
}

// TestFieldCache_CachePersistence tests that cached types persist
func TestFieldCache_CachePersistence(t *testing.T) {
	cache := NewFieldCache()
	formType := reflect.TypeOf(TestForm{})

	// Cache the type
	entry1 := cache.CacheType(formType)

	// Cache again - should return same entry
	entry2 := cache.CacheType(formType)

	// Should be the exact same entry (pointer equality)
	assert.Equal(t, entry1, entry2, "Should return cached entry, not create new one")
}

// TestFieldCache_MultipleTypes tests caching multiple struct types
func TestFieldCache_MultipleTypes(t *testing.T) {
	cache := NewFieldCache()

	// Cache multiple types
	testFormType := reflect.TypeOf(TestForm{})
	simpleType := reflect.TypeOf(SimpleStruct{})
	nestedType := reflect.TypeOf(NestedStruct{})

	entry1 := cache.CacheType(testFormType)
	entry2 := cache.CacheType(simpleType)
	entry3 := cache.CacheType(nestedType)

	assert.NotNil(t, entry1, "TestForm should be cached")
	assert.NotNil(t, entry2, "SimpleStruct should be cached")
	assert.NotNil(t, entry3, "NestedStruct should be cached")

	// Each type should have different entries
	assert.NotEqual(t, entry1, entry2, "Different types should have different entries")
	assert.NotEqual(t, entry2, entry3, "Different types should have different entries")

	// Verify counts
	assert.Equal(t, 3, len(entry1.Indices), "TestForm has 3 fields")
	assert.Equal(t, 1, len(entry2.Indices), "SimpleStruct has 1 field")
	assert.Equal(t, 2, len(entry3.Indices), "NestedStruct has 2 fields")
}

// TestFieldCache_WarmUp tests pre-caching with WarmUp
func TestFieldCache_WarmUp(t *testing.T) {
	cache := NewFieldCache()

	// WarmUp with a struct instance
	cache.WarmUp(TestForm{})

	// Type should already be cached
	formType := reflect.TypeOf(TestForm{})
	idx, ok := cache.GetFieldIndex(formType, "Name")
	assert.True(t, ok, "Type should be pre-cached by WarmUp")
	assert.Equal(t, 0, idx, "Name should be at correct index")
}

// TestFieldCache_WarmUpPointer tests WarmUp with pointer
func TestFieldCache_WarmUpPointer(t *testing.T) {
	cache := NewFieldCache()

	// WarmUp with a pointer to struct
	cache.WarmUp(&TestForm{})

	// Should cache the underlying struct type, not the pointer
	formType := reflect.TypeOf(TestForm{})
	idx, ok := cache.GetFieldIndex(formType, "Name")
	assert.True(t, ok, "Should cache underlying struct type")
	assert.Equal(t, 0, idx, "Name should be at correct index")
}

// TestFieldCache_Stats tests statistics tracking
func TestFieldCache_Stats(t *testing.T) {
	cache := NewFieldCache()

	// Initially zero stats
	stats := cache.Stats()
	assert.Equal(t, 0, stats.TypesCached, "Initially no types cached")
	assert.Equal(t, int64(0), stats.Hits, "Initially no hits")
	assert.Equal(t, int64(0), stats.Misses, "Initially no misses")
	assert.Equal(t, 0.0, stats.HitRate, "Initially 0% hit rate")

	// Cache a type (miss)
	formType := reflect.TypeOf(TestForm{})
	cache.GetFieldIndex(formType, "Name")

	stats = cache.Stats()
	assert.Equal(t, 1, stats.TypesCached, "Should have 1 type cached")
	assert.Equal(t, int64(0), stats.Hits, "First access is a miss")
	assert.Equal(t, int64(1), stats.Misses, "Should record 1 miss")

	// Access again (hit)
	cache.GetFieldIndex(formType, "Email")

	stats = cache.Stats()
	assert.Equal(t, int64(1), stats.Hits, "Should record 1 hit")
	assert.Equal(t, int64(1), stats.Misses, "Misses should stay at 1")
	assert.Equal(t, 0.5, stats.HitRate, "Hit rate should be 50%")

	// More hits
	cache.GetFieldIndex(formType, "Age")
	cache.GetFieldIndex(formType, "Name")

	stats = cache.Stats()
	assert.Equal(t, int64(3), stats.Hits, "Should record 3 hits")
	assert.Equal(t, int64(1), stats.Misses, "Misses should stay at 1")
	assert.Equal(t, 0.75, stats.HitRate, "Hit rate should be 75%")
}

// TestFieldCache_ConcurrentAccess tests thread-safe concurrent access
func TestFieldCache_ConcurrentAccess(t *testing.T) {
	cache := NewFieldCache()

	var wg sync.WaitGroup
	numGoroutines := 100

	formType := reflect.TypeOf(TestForm{})

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Alternate between different fields
			fieldName := "Name"
			if id%3 == 1 {
				fieldName = "Email"
			} else if id%3 == 2 {
				fieldName = "Age"
			}

			idx, ok := cache.GetFieldIndex(formType, fieldName)
			require.True(t, ok, "Should find field %s", fieldName)
			require.GreaterOrEqual(t, idx, 0, "Index should be valid")
		}(i)
	}

	wg.Wait()

	// Verify stats are consistent
	stats := cache.Stats()
	assert.Equal(t, 1, stats.TypesCached, "Should have 1 type cached")
	assert.Greater(t, stats.Hits+stats.Misses, int64(0), "Should have recorded accesses")
}

// TestFieldCache_InvalidType tests handling of non-struct types
func TestFieldCache_InvalidType(t *testing.T) {
	cache := NewFieldCache()

	// Try with non-struct types
	intType := reflect.TypeOf(42)
	entry := cache.CacheType(intType)

	// Should return empty entry for non-struct
	assert.NotNil(t, entry, "Should return entry even for non-struct")
	assert.Equal(t, 0, len(entry.Indices), "Non-struct should have no fields")
	assert.Equal(t, 0, len(entry.Types), "Non-struct should have no field types")
}

// TestFieldCache_UnexportedFields tests handling unexported fields
func TestFieldCache_UnexportedFields(t *testing.T) {
	type WithUnexported struct {
		Public  string
		private int //nolint:unused // Intentionally unused for testing unexported field handling
		Another bool
	}

	cache := NewFieldCache()
	structType := reflect.TypeOf(WithUnexported{})
	entry := cache.CacheType(structType)

	// Should cache all fields (including unexported)
	// This matches reflect.Value.FieldByName() behavior
	assert.Equal(t, 3, len(entry.Indices), "Should cache all fields including unexported")

	// Verify public field
	idx, ok := entry.Indices["Public"]
	assert.True(t, ok, "Should find Public field")
	assert.Equal(t, 0, idx, "Public at index 0")

	// Verify unexported field
	idx, ok = entry.Indices["private"]
	assert.True(t, ok, "Should find private field")
	assert.Equal(t, 1, idx, "private at index 1")
}

// TestFieldCache_EmptyStruct tests caching empty struct
func TestFieldCache_EmptyStruct(t *testing.T) {
	type Empty struct{}

	cache := NewFieldCache()
	emptyType := reflect.TypeOf(Empty{})
	entry := cache.CacheType(emptyType)

	assert.NotNil(t, entry, "Should return entry for empty struct")
	assert.Equal(t, 0, len(entry.Indices), "Empty struct has no fields")
	assert.Equal(t, 0, len(entry.Types), "Empty struct has no field types")
}

// TestGlobalCache_EnableGlobalCache tests global cache initialization
func TestGlobalCache_EnableGlobalCache(t *testing.T) {
	// Store original state
	originalCache := GlobalCache
	defer func() { GlobalCache = originalCache }()

	// Reset global cache
	GlobalCache = nil

	// Enable global cache
	EnableGlobalCache()

	assert.NotNil(t, GlobalCache, "GlobalCache should be initialized")

	// Enable again (idempotent)
	cache1 := GlobalCache
	EnableGlobalCache()
	cache2 := GlobalCache

	assert.Equal(t, cache1, cache2, "Should be idempotent, returning same instance")
}
