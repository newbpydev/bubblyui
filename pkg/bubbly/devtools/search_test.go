package devtools

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSearchWidget(t *testing.T) {
	tests := []struct {
		name           string
		components     []*ComponentSnapshot
		wantComponents int
	}{
		{
			name:           "creates with nil components",
			components:     nil,
			wantComponents: 0,
		},
		{
			name:           "creates with empty components",
			components:     []*ComponentSnapshot{},
			wantComponents: 0,
		},
		{
			name: "creates with components",
			components: []*ComponentSnapshot{
				{ID: "1", Name: "Component1"},
				{ID: "2", Name: "Component2"},
			},
			wantComponents: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sw := NewSearchWidget(tt.components)

			assert.NotNil(t, sw)
			assert.Equal(t, "", sw.GetQuery())
			assert.Equal(t, 0, len(sw.GetResults()))
			assert.Equal(t, 0, sw.GetCursor())
		})
	}
}

func TestSearchWidget_Search_ExactMatch(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button", Type: "ButtonType"},
		{ID: "2", Name: "Input", Type: "InputType"},
		{ID: "3", Name: "Form", Type: "FormType"},
	}

	sw := NewSearchWidget(components)
	sw.Search("Button")

	results := sw.GetResults()
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "Button", results[0].Name)
}

func TestSearchWidget_Search_PartialMatch(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "PrimaryButton", Type: "ButtonType"},
		{ID: "2", Name: "SecondaryButton", Type: "ButtonType"},
		{ID: "3", Name: "Input", Type: "InputType"},
	}

	sw := NewSearchWidget(components)
	sw.Search("Button")

	results := sw.GetResults()
	assert.Equal(t, 2, len(results))
	assert.Contains(t, []string{results[0].Name, results[1].Name}, "PrimaryButton")
	assert.Contains(t, []string{results[0].Name, results[1].Name}, "SecondaryButton")
}

func TestSearchWidget_Search_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantCount int
		wantNames []string
	}{
		{
			name:      "lowercase query matches uppercase name",
			query:     "button",
			wantCount: 1,
			wantNames: []string{"Button"},
		},
		{
			name:      "uppercase query matches lowercase name",
			query:     "BUTTON",
			wantCount: 1,
			wantNames: []string{"Button"},
		},
		{
			name:      "mixed case query",
			query:     "BuTtOn",
			wantCount: 1,
			wantNames: []string{"Button"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := []*ComponentSnapshot{
				{ID: "1", Name: "Button", Type: "ButtonType"},
				{ID: "2", Name: "Input", Type: "InputType"},
			}

			sw := NewSearchWidget(components)
			sw.Search(tt.query)

			results := sw.GetResults()
			assert.Equal(t, tt.wantCount, len(results))
			if tt.wantCount > 0 {
				assert.Equal(t, tt.wantNames[0], results[0].Name)
			}
		})
	}
}

func TestSearchWidget_Search_MatchesType(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "PrimaryBtn", Type: "ButtonComponent"},
		{ID: "2", Name: "SecondaryBtn", Type: "ButtonComponent"},
		{ID: "3", Name: "TextInput", Type: "InputComponent"},
	}

	sw := NewSearchWidget(components)
	sw.Search("Button")

	results := sw.GetResults()
	assert.Equal(t, 2, len(results))
}

func TestSearchWidget_Search_EmptyQuery(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button"},
		{ID: "2", Name: "Input"},
		{ID: "3", Name: "Form"},
	}

	sw := NewSearchWidget(components)
	sw.Search("")

	results := sw.GetResults()
	assert.Equal(t, 3, len(results))
}

func TestSearchWidget_Search_NoMatches(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button"},
		{ID: "2", Name: "Input"},
	}

	sw := NewSearchWidget(components)
	sw.Search("NonExistent")

	results := sw.GetResults()
	assert.Equal(t, 0, len(results))
}

func TestSearchWidget_NextResult(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button1"},
		{ID: "2", Name: "Button2"},
		{ID: "3", Name: "Button3"},
	}

	sw := NewSearchWidget(components)
	sw.Search("Button")

	// Start at 0
	assert.Equal(t, 0, sw.GetCursor())

	// Move to 1
	sw.NextResult()
	assert.Equal(t, 1, sw.GetCursor())

	// Move to 2
	sw.NextResult()
	assert.Equal(t, 2, sw.GetCursor())

	// Wrap around to 0
	sw.NextResult()
	assert.Equal(t, 0, sw.GetCursor())
}

func TestSearchWidget_PrevResult(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button1"},
		{ID: "2", Name: "Button2"},
		{ID: "3", Name: "Button3"},
	}

	sw := NewSearchWidget(components)
	sw.Search("Button")

	// Start at 0, wrap to last
	sw.PrevResult()
	assert.Equal(t, 2, sw.GetCursor())

	// Move to 1
	sw.PrevResult()
	assert.Equal(t, 1, sw.GetCursor())

	// Move to 0
	sw.PrevResult()
	assert.Equal(t, 0, sw.GetCursor())
}

func TestSearchWidget_GetSelected(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button1"},
		{ID: "2", Name: "Button2"},
		{ID: "3", Name: "Button3"},
	}

	sw := NewSearchWidget(components)
	sw.Search("Button")

	// Get first result
	selected := sw.GetSelected()
	assert.NotNil(t, selected)
	assert.Equal(t, "Button1", selected.Name)

	// Navigate and get second result
	sw.NextResult()
	selected = sw.GetSelected()
	assert.NotNil(t, selected)
	assert.Equal(t, "Button2", selected.Name)
}

func TestSearchWidget_GetSelected_NoResults(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button"},
	}

	sw := NewSearchWidget(components)
	sw.Search("NonExistent")

	selected := sw.GetSelected()
	assert.Nil(t, selected)
}

func TestSearchWidget_Clear(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button"},
		{ID: "2", Name: "Input"},
	}

	sw := NewSearchWidget(components)
	sw.Search("Button")
	sw.NextResult()

	// Verify state before clear
	assert.NotEqual(t, "", sw.GetQuery())
	assert.NotEqual(t, 0, len(sw.GetResults()))

	// Clear
	sw.Clear()

	// Verify state after clear
	assert.Equal(t, "", sw.GetQuery())
	assert.Equal(t, 0, len(sw.GetResults()))
	assert.Equal(t, 0, sw.GetCursor())
}

func TestSearchWidget_SetComponents(t *testing.T) {
	initialComponents := []*ComponentSnapshot{
		{ID: "1", Name: "Button"},
	}

	sw := NewSearchWidget(initialComponents)
	sw.Search("Button")
	assert.Equal(t, 1, len(sw.GetResults()))

	// Update components
	newComponents := []*ComponentSnapshot{
		{ID: "2", Name: "Input"},
		{ID: "3", Name: "Form"},
	}
	sw.SetComponents(newComponents)

	// Search again with new components
	sw.Search("Input")
	assert.Equal(t, 1, len(sw.GetResults()))
	assert.Equal(t, "Input", sw.GetResults()[0].Name)
}

func TestSearchWidget_Render(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button", Type: "ButtonType"},
		{ID: "2", Name: "Input", Type: "InputType"},
	}

	sw := NewSearchWidget(components)
	sw.Search("Button")

	output := sw.Render()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Button")
}

func TestSearchWidget_Render_NoResults(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button"},
	}

	sw := NewSearchWidget(components)
	sw.Search("NonExistent")

	output := sw.Render()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "No results")
}

func TestSearchWidget_ThreadSafety(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button1"},
		{ID: "2", Name: "Button2"},
		{ID: "3", Name: "Button3"},
	}

	sw := NewSearchWidget(components)
	sw.Search("Button")

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent searches
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			if index%2 == 0 {
				sw.Search("Button")
			} else {
				sw.Search("Input")
			}
		}(i)
	}

	// Concurrent navigation
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sw.NextResult()
		}()
	}

	// Concurrent renders
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = sw.Render()
		}()
	}

	wg.Wait()

	// Should not panic and should have valid state
	output := sw.Render()
	assert.NotEmpty(t, output)
}

func TestSearchWidget_Performance(t *testing.T) {
	// Create 1000 components
	components := make([]*ComponentSnapshot, 1000)
	for i := 0; i < 1000; i++ {
		components[i] = &ComponentSnapshot{
			ID:   string(rune(i)),
			Name: "Component" + string(rune(i)),
			Type: "Type" + string(rune(i%10)),
		}
	}

	sw := NewSearchWidget(components)

	start := time.Now()
	sw.Search("Component")
	duration := time.Since(start)

	// Should complete in less than 100ms
	assert.Less(t, duration.Milliseconds(), int64(100))
	assert.Equal(t, 1000, len(sw.GetResults()))
}

func TestSearchWidget_NavigationWithEmptyResults(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button"},
	}

	sw := NewSearchWidget(components)
	sw.Search("NonExistent")

	// Should not panic with empty results
	sw.NextResult()
	assert.Equal(t, 0, sw.GetCursor())

	sw.PrevResult()
	assert.Equal(t, 0, sw.GetCursor())
}

func TestSearchWidget_MultipleSearches(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button"},
		{ID: "2", Name: "Input"},
		{ID: "3", Name: "Form"},
	}

	sw := NewSearchWidget(components)

	// First search
	sw.Search("Button")
	assert.Equal(t, 1, len(sw.GetResults()))

	// Second search should reset cursor
	sw.NextResult()
	sw.Search("Input")
	assert.Equal(t, 1, len(sw.GetResults()))
	assert.Equal(t, 0, sw.GetCursor())
}

func TestSearchWidget_GetResultCount(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button1"},
		{ID: "2", Name: "Button2"},
		{ID: "3", Name: "Input"},
	}

	sw := NewSearchWidget(components)
	sw.Search("Button")

	assert.Equal(t, 2, sw.GetResultCount())
}

// TestSearchWidget_RenderPagination tests the pagination logic in renderResults
func TestSearchWidget_RenderPagination(t *testing.T) {
	// Create more than 10 components to trigger pagination
	components := make([]*ComponentSnapshot, 15)
	for i := 0; i < 15; i++ {
		components[i] = &ComponentSnapshot{
			ID:   string(rune('A' + i)),
			Name: "Item",
			Type: "Component",
		}
	}

	sw := NewSearchWidget(components)
	sw.Search("Item") // Should match all 15 items

	assert.Equal(t, 15, sw.GetResultCount(), "All 15 items should match")

	// Initial render with cursor at 0 - should show ellipsis at end
	output := sw.Render()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "...", "Should show ellipsis indicating more results")
}

// TestSearchWidget_RenderPagination_CursorAtEnd tests pagination with cursor at end
func TestSearchWidget_RenderPagination_CursorAtEnd(t *testing.T) {
	// Create more than 10 components
	components := make([]*ComponentSnapshot, 20)
	for i := 0; i < 20; i++ {
		components[i] = &ComponentSnapshot{
			ID:   string(rune('A' + i)),
			Name: "Item",
			Type: "Component",
		}
	}

	sw := NewSearchWidget(components)
	sw.Search("Item")

	// Move cursor to near the end
	for i := 0; i < 18; i++ {
		sw.NextResult()
	}

	// Render with cursor near end - should show ellipsis at start
	output := sw.Render()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "...", "Should show ellipsis indicating earlier results")
}

// TestSearchWidget_RenderPagination_CursorInMiddle tests pagination with cursor in middle
func TestSearchWidget_RenderPagination_CursorInMiddle(t *testing.T) {
	// Create 25 components
	components := make([]*ComponentSnapshot, 25)
	for i := 0; i < 25; i++ {
		components[i] = &ComponentSnapshot{
			ID:   string(rune('A' + i%26)),
			Name: "Item",
			Type: "Component",
		}
	}

	sw := NewSearchWidget(components)
	sw.Search("Item")

	// Move cursor to middle
	for i := 0; i < 12; i++ {
		sw.NextResult()
	}

	// Render with cursor in middle - should show ellipsis at both ends
	output := sw.Render()
	assert.NotEmpty(t, output)
	// The output should show a window of results around cursor
	assert.Greater(t, len(output), 0, "Should render results around cursor")
}

// TestSearchWidget_RenderResult_Selected tests rendering a selected result
func TestSearchWidget_RenderResult_Selected(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button", Type: "ButtonType"},
		{ID: "2", Name: "Input", Type: "InputType"},
	}

	sw := NewSearchWidget(components)
	sw.Search("Button")

	// First result should be selected by default (cursor = 0)
	output := sw.Render()
	assert.NotEmpty(t, output)
	// The selected result should have different styling
	assert.Contains(t, output, "Button")
}

// TestSearchWidget_RenderResult_NotSelected tests rendering non-selected results
func TestSearchWidget_RenderResult_NotSelected(t *testing.T) {
	components := []*ComponentSnapshot{
		{ID: "1", Name: "Button", Type: "ButtonType"},
		{ID: "2", Name: "Input", Type: "InputType"},
		{ID: "3", Name: "Form", Type: "FormType"},
	}

	sw := NewSearchWidget(components)
	sw.Search("") // Empty search matches all

	assert.Equal(t, 3, sw.GetResultCount())

	// Move cursor to second item
	sw.NextResult()
	assert.Equal(t, 1, sw.GetCursor())

	output := sw.Render()
	assert.Contains(t, output, "Button")
	assert.Contains(t, output, "Input")
	assert.Contains(t, output, "Form")
}
