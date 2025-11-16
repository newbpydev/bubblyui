package testutil

import (
	"fmt"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/directives"
	"github.com/stretchr/testify/assert"
)

// TestForEachTester_Creation tests basic tester creation
func TestForEachTester_Creation(t *testing.T) {
	tests := []struct {
		name  string
		items []string
	}{
		{"empty list", []string{}},
		{"single item", []string{"apple"}},
		{"multiple items", []string{"apple", "banana", "cherry"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemsRef := bubbly.NewRef(tt.items)
			tester := NewForEachTester(itemsRef)

			assert.NotNil(t, tester)
			assert.Equal(t, tt.items, itemsRef.Get())
		})
	}
}

// TestForEachTester_AssertItemCount tests item count assertions
func TestForEachTester_AssertItemCount(t *testing.T) {
	tests := []struct {
		name          string
		items         []string
		expectedCount int
	}{
		{"empty list", []string{}, 0},
		{"single item", []string{"apple"}, 1},
		{"three items", []string{"apple", "banana", "cherry"}, 3},
		{"ten items", []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemsRef := bubbly.NewRef(tt.items)
			tester := NewForEachTester(itemsRef)

			// This should pass
			tester.AssertItemCount(t, tt.expectedCount)
		})
	}
}

// TestForEachTester_AssertItemCount_Failure tests count assertion failures
func TestForEachTester_AssertItemCount_Failure(t *testing.T) {
	itemsRef := bubbly.NewRef([]string{"apple", "banana"})
	tester := NewForEachTester(itemsRef)

	// Use mock testing.T to capture error
	mockT := &mockTestingT{}
	tester.AssertItemCount(mockT, 5) // Wrong count

	assert.True(t, mockT.failed, "Expected error to be called")
	assert.NotEmpty(t, mockT.errors, "Expected error message")
	assert.Contains(t, mockT.errors[0], "expected 5 items")
	assert.Contains(t, mockT.errors[0], "got 2")
}

// TestForEachTester_AssertItemRendered tests item rendering assertions
func TestForEachTester_AssertItemRendered(t *testing.T) {
	items := []string{"apple", "banana", "cherry"}
	itemsRef := bubbly.NewRef(items)
	tester := NewForEachTester(itemsRef)

	// Render items
	renderFunc := func(item string, index int) string {
		return fmt.Sprintf("%d. %s\n", index+1, item)
	}
	tester.Render(renderFunc)

	// Assert each item rendered correctly
	tester.AssertItemRendered(t, 0, "1. apple\n")
	tester.AssertItemRendered(t, 1, "2. banana\n")
	tester.AssertItemRendered(t, 2, "3. cherry\n")
}

// TestForEachTester_AssertItemRendered_Failure tests rendering assertion failures
func TestForEachTester_AssertItemRendered_Failure(t *testing.T) {
	items := []string{"apple", "banana"}
	itemsRef := bubbly.NewRef(items)
	tester := NewForEachTester(itemsRef)

	// Render items
	renderFunc := func(item string, index int) string {
		return fmt.Sprintf("%d. %s\n", index+1, item)
	}
	tester.Render(renderFunc)

	// Use mock testing.T to capture error
	mockT := &mockTestingT{}
	tester.AssertItemRendered(mockT, 0, "wrong content")

	assert.True(t, mockT.failed, "Expected error to be called")
	assert.NotEmpty(t, mockT.errors, "Expected error message")
	assert.Contains(t, mockT.errors[0], "expected")
	assert.Contains(t, mockT.errors[0], "got")
}

// TestForEachTester_ItemUpdate tests updating items
func TestForEachTester_ItemUpdate(t *testing.T) {
	itemsRef := bubbly.NewRef([]string{"apple", "banana"})
	tester := NewForEachTester(itemsRef)

	// Initial count
	tester.AssertItemCount(t, 2)

	// Update items
	itemsRef.Set([]string{"apple", "banana", "cherry", "date"})

	// New count
	tester.AssertItemCount(t, 4)
}

// TestForEachTester_ItemRemoval tests removing items
func TestForEachTester_ItemRemoval(t *testing.T) {
	itemsRef := bubbly.NewRef([]string{"apple", "banana", "cherry"})
	tester := NewForEachTester(itemsRef)

	// Initial count
	tester.AssertItemCount(t, 3)

	// Remove an item
	itemsRef.Set([]string{"apple", "cherry"})

	// New count
	tester.AssertItemCount(t, 2)

	// Verify rendering reflects removal
	renderFunc := func(item string, index int) string {
		return fmt.Sprintf("%s ", item)
	}
	tester.Render(renderFunc)

	tester.AssertItemRendered(t, 0, "apple ")
	tester.AssertItemRendered(t, 1, "cherry ")
}

// TestForEachTester_ItemAddition tests adding items
func TestForEachTester_ItemAddition(t *testing.T) {
	itemsRef := bubbly.NewRef([]string{"apple"})
	tester := NewForEachTester(itemsRef)

	// Initial count
	tester.AssertItemCount(t, 1)

	// Add items
	itemsRef.Set([]string{"apple", "banana", "cherry"})

	// New count
	tester.AssertItemCount(t, 3)
}

// TestForEachTester_EmptyList tests empty list handling
func TestForEachTester_EmptyList(t *testing.T) {
	itemsRef := bubbly.NewRef([]string{})
	tester := NewForEachTester(itemsRef)

	// Empty list
	tester.AssertItemCount(t, 0)

	// Render should return empty
	renderFunc := func(item string, index int) string {
		return item
	}
	tester.Render(renderFunc)

	// No items to assert
	assert.Equal(t, 0, len(tester.GetRendered()))
}

// TestForEachTester_NilItems tests nil items handling
func TestForEachTester_NilItems(t *testing.T) {
	var nilItems []string
	itemsRef := bubbly.NewRef(nilItems)
	tester := NewForEachTester(itemsRef)

	// Nil list treated as empty
	tester.AssertItemCount(t, 0)
}

// TestForEachTester_ComplexItems tests with struct items
func TestForEachTester_ComplexItems(t *testing.T) {
	type User struct {
		Name  string
		Email string
	}

	users := []User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}

	itemsRef := bubbly.NewRef(users)
	tester := NewForEachTester(itemsRef)

	tester.AssertItemCount(t, 2)

	// Render users
	renderFunc := func(user User, index int) string {
		return fmt.Sprintf("%d. %s <%s>\n", index+1, user.Name, user.Email)
	}
	tester.Render(renderFunc)

	tester.AssertItemRendered(t, 0, "1. Alice <alice@example.com>\n")
	tester.AssertItemRendered(t, 1, "2. Bob <bob@example.com>\n")
}

// TestForEachTester_IntegrationWithDirective tests integration with ForEach directive
func TestForEachTester_IntegrationWithDirective(t *testing.T) {
	items := []string{"apple", "banana", "cherry"}
	itemsRef := bubbly.NewRef(items)

	// Create tester
	tester := NewForEachTester(itemsRef)

	// Render with directive
	renderFunc := func(item string, index int) string {
		return fmt.Sprintf("- %s\n", item)
	}

	// Use actual ForEach directive
	directive := directives.ForEach(items, renderFunc)
	output := directive.Render()

	// Verify output
	expected := "- apple\n- banana\n- cherry\n"
	assert.Equal(t, expected, output)

	// Tester should also render correctly
	tester.Render(renderFunc)
	tester.AssertItemRendered(t, 0, "- apple\n")
	tester.AssertItemRendered(t, 1, "- banana\n")
	tester.AssertItemRendered(t, 2, "- cherry\n")
}

// TestForEachTester_GetRendered tests getting all rendered items
func TestForEachTester_GetRendered(t *testing.T) {
	items := []string{"a", "b", "c"}
	itemsRef := bubbly.NewRef(items)
	tester := NewForEachTester(itemsRef)

	renderFunc := func(item string, index int) string {
		return fmt.Sprintf("[%s]", item)
	}
	tester.Render(renderFunc)

	rendered := tester.GetRendered()
	assert.Equal(t, 3, len(rendered))
	assert.Equal(t, "[a]", rendered[0])
	assert.Equal(t, "[b]", rendered[1])
	assert.Equal(t, "[c]", rendered[2])
}

// TestForEachTester_GetFullOutput tests getting full concatenated output
func TestForEachTester_GetFullOutput(t *testing.T) {
	items := []string{"a", "b", "c"}
	itemsRef := bubbly.NewRef(items)
	tester := NewForEachTester(itemsRef)

	renderFunc := func(item string, index int) string {
		return item
	}
	tester.Render(renderFunc)

	output := tester.GetFullOutput()
	assert.Equal(t, "abc", output)
}

// TestForEachTester_ThreadSafety tests concurrent access
func TestForEachTester_ThreadSafety(t *testing.T) {
	itemsRef := bubbly.NewRef([]string{"a", "b", "c"})
	tester := NewForEachTester(itemsRef)

	renderFunc := func(item string, index int) string {
		return item
	}

	// Render from multiple goroutines
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			tester.Render(renderFunc)
			tester.AssertItemCount(t, 3)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// mockTestingT is defined in assertions_state_test.go
