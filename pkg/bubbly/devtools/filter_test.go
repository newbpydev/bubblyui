package devtools

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestComponentFilter_Apply_EmptyFilter tests that an empty filter returns all components
func TestComponentFilter_Apply_EmptyFilter(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button", Type: "button", Status: "mounted"},
		{ID: "2", Name: "Input", Type: "input", Status: "mounted"},
		{ID: "3", Name: "Form", Type: "form", Status: "unmounted"},
	}

	filter := NewComponentFilter()
	result := filter.Apply(components)

	assert.Equal(t, 3, len(result), "Empty filter should return all components")
	assert.Equal(t, components, result, "Result should match input")
}

// TestComponentFilter_Apply_TypeFilter tests filtering by a single type
func TestComponentFilter_Apply_TypeFilter(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button1", Type: "button", Status: "mounted"},
		{ID: "2", Name: "Input1", Type: "input", Status: "mounted"},
		{ID: "3", Name: "Button2", Type: "button", Status: "mounted"},
	}

	filter := NewComponentFilter().WithTypes([]string{"button"})
	result := filter.Apply(components)

	assert.Equal(t, 2, len(result), "Should return 2 button components")
	assert.Equal(t, "button", result[0].Type)
	assert.Equal(t, "button", result[1].Type)
}

// TestComponentFilter_Apply_MultipleTypes tests filtering by multiple types
func TestComponentFilter_Apply_MultipleTypes(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button", Type: "button", Status: "mounted"},
		{ID: "2", Name: "Input", Type: "input", Status: "mounted"},
		{ID: "3", Name: "Form", Type: "form", Status: "mounted"},
		{ID: "4", Name: "Card", Type: "card", Status: "mounted"},
	}

	filter := NewComponentFilter().WithTypes([]string{"button", "input"})
	result := filter.Apply(components)

	assert.Equal(t, 2, len(result), "Should return button and input components")
	types := []string{result[0].Type, result[1].Type}
	assert.Contains(t, types, "button")
	assert.Contains(t, types, "input")
}

// TestComponentFilter_Apply_StatusFilter tests filtering by status
func TestComponentFilter_Apply_StatusFilter(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button", Type: "button", Status: "mounted"},
		{ID: "2", Name: "Input", Type: "input", Status: "unmounted"},
		{ID: "3", Name: "Form", Type: "form", Status: "mounted"},
	}

	filter := NewComponentFilter().WithStatuses([]string{"mounted"})
	result := filter.Apply(components)

	assert.Equal(t, 2, len(result), "Should return 2 mounted components")
	assert.Equal(t, "mounted", result[0].Status)
	assert.Equal(t, "mounted", result[1].Status)
}

// TestComponentFilter_Apply_CustomFilter tests custom predicate function
func TestComponentFilter_Apply_CustomFilter(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button", Type: "button", Status: "mounted"},
		{ID: "2", Name: "Input", Type: "input", Status: "mounted"},
		{ID: "3", Name: "Form", Type: "form", Status: "mounted"},
	}

	// Custom filter: only components with name starting with "B"
	customFunc := func(c *ComponentSnapshot) bool {
		return len(c.Name) > 0 && c.Name[0] == 'B'
	}

	filter := NewComponentFilter().WithCustom(customFunc)
	result := filter.Apply(components)

	assert.Equal(t, 1, len(result), "Should return 1 component starting with B")
	assert.Equal(t, "Button", result[0].Name)
}

// TestComponentFilter_Apply_CombinedFilters tests combining multiple filter criteria
func TestComponentFilter_Apply_CombinedFilters(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button", Type: "button", Status: "mounted"},
		{ID: "2", Name: "Input", Type: "input", Status: "mounted"},
		{ID: "3", Name: "Button2", Type: "button", Status: "unmounted"},
		{ID: "4", Name: "Form", Type: "form", Status: "mounted"},
	}

	// Filter: type=button AND status=mounted AND name contains "Button"
	customFunc := func(c *ComponentSnapshot) bool {
		return c.Name == "Button"
	}

	filter := NewComponentFilter().
		WithTypes([]string{"button"}).
		WithStatuses([]string{"mounted"}).
		WithCustom(customFunc)

	result := filter.Apply(components)

	assert.Equal(t, 1, len(result), "Should return 1 component matching all criteria")
	assert.Equal(t, "Button", result[0].Name)
	assert.Equal(t, "button", result[0].Type)
	assert.Equal(t, "mounted", result[0].Status)
}

// TestComponentFilter_Apply_NoMatches tests when no components match the filter
func TestComponentFilter_Apply_NoMatches(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button", Type: "button", Status: "mounted"},
		{ID: "2", Name: "Input", Type: "input", Status: "mounted"},
	}

	filter := NewComponentFilter().WithTypes([]string{"nonexistent"})
	result := filter.Apply(components)

	assert.Equal(t, 0, len(result), "Should return empty slice when no matches")
	assert.NotNil(t, result, "Result should not be nil")
}

// TestComponentFilter_Performance tests performance with large component trees
func TestComponentFilter_Performance(t *testing.T) {
	// Create 1000 components
	components := make([]*ComponentSnapshot, 1000)
	for i := 0; i < 1000; i++ {
		components[i] = &ComponentSnapshot{
			ID:     string(rune(i)),
			Name:   "Component",
			Type:   "button",
			Status: "mounted",
		}
	}

	filter := NewComponentFilter().WithTypes([]string{"button"})

	start := time.Now()
	result := filter.Apply(components)
	duration := time.Since(start)

	assert.Equal(t, 1000, len(result), "Should return all 1000 components")
	assert.Less(t, duration.Milliseconds(), int64(100), "Should complete in < 100ms")
}

// TestComponentFilter_ThreadSafety tests concurrent Apply calls
func TestComponentFilter_ThreadSafety(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button", Type: "button", Status: "mounted"},
		{ID: "2", Name: "Input", Type: "input", Status: "mounted"},
		{ID: "3", Name: "Form", Type: "form", Status: "mounted"},
	}

	filter := NewComponentFilter().WithTypes([]string{"button"})

	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := filter.Apply(components)
			assert.Equal(t, 1, len(result))
		}()
	}

	wg.Wait()
}

// TestComponentFilter_NilComponents tests handling of nil input
func TestComponentFilter_NilComponents(t *testing.T) {
	filter := NewComponentFilter()
	result := filter.Apply(nil)

	assert.NotNil(t, result, "Result should not be nil")
	assert.Equal(t, 0, len(result), "Result should be empty slice")
}

// TestComponentFilter_EmptyComponents tests handling of empty slice
func TestComponentFilter_EmptyComponents(t *testing.T) {
	filter := NewComponentFilter()
	result := filter.Apply([]*ComponentSnapshot{})

	assert.NotNil(t, result, "Result should not be nil")
	assert.Equal(t, 0, len(result), "Result should be empty slice")
}
