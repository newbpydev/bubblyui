package directives

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// mockReporter is a test implementation of ErrorReporter for directive tests
type mockReporter struct {
	panicCalls []mockPanicCall
	errorCalls []mockErrorCall
	flushCalls int
	flushError error
	mu         sync.Mutex
}

type mockPanicCall struct {
	err *observability.HandlerPanicError
	ctx *observability.ErrorContext
}

type mockErrorCall struct {
	err error
	ctx *observability.ErrorContext
}

func (m *mockReporter) ReportPanic(err *observability.HandlerPanicError, ctx *observability.ErrorContext) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.panicCalls = append(m.panicCalls, mockPanicCall{err: err, ctx: ctx})
}

func (m *mockReporter) ReportError(err error, ctx *observability.ErrorContext) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorCalls = append(m.errorCalls, mockErrorCall{err: err, ctx: ctx})
}

func (m *mockReporter) Flush(timeout time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flushCalls++
	return m.flushError
}

func (m *mockReporter) getErrorCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.errorCalls)
}

// TestForEachDirective_BasicIteration tests basic iteration over a slice
func TestForEachDirective_BasicIteration(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		render   func(string, int) string
		expected string
	}{
		{
			name:  "simple string slice",
			items: []string{"A", "B", "C"},
			render: func(item string, index int) string {
				return fmt.Sprintf("%d:%s,", index, item)
			},
			expected: "0:A,1:B,2:C,",
		},
		{
			name:  "single item",
			items: []string{"X"},
			render: func(item string, index int) string {
				return item
			},
			expected: "X",
		},
		{
			name:  "numbered list",
			items: []string{"First", "Second", "Third"},
			render: func(item string, index int) string {
				return fmt.Sprintf("%d. %s\n", index+1, item)
			},
			expected: "1. First\n2. Second\n3. Third\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ForEach(tt.items, tt.render).Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestForEachDirective_EmptySlice tests handling of empty collections
func TestForEachDirective_EmptySlice(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		render   func(string, int) string
		expected string
	}{
		{
			name:  "empty slice",
			items: []string{},
			render: func(item string, index int) string {
				return item
			},
			expected: "",
		},
		{
			name:  "nil slice",
			items: nil,
			render: func(item string, index int) string {
				return item
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ForEach(tt.items, tt.render).Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestForEachDirective_TypeSafety tests generic type safety
func TestForEachDirective_TypeSafety(t *testing.T) {
	t.Run("integer slice", func(t *testing.T) {
		items := []int{1, 2, 3}
		result := ForEach(items, func(item int, index int) string {
			return fmt.Sprintf("%d*%d=%d ", index, item, index*item)
		}).Render()
		assert.Equal(t, "0*1=0 1*2=2 2*3=6 ", result)
	})

	t.Run("struct slice", func(t *testing.T) {
		type User struct {
			Name  string
			Email string
		}
		users := []User{
			{Name: "Alice", Email: "alice@example.com"},
			{Name: "Bob", Email: "bob@example.com"},
		}
		result := ForEach(users, func(user User, index int) string {
			return fmt.Sprintf("%d. %s <%s>\n", index+1, user.Name, user.Email)
		}).Render()
		expected := "1. Alice <alice@example.com>\n2. Bob <bob@example.com>\n"
		assert.Equal(t, expected, result)
	})

	t.Run("pointer slice", func(t *testing.T) {
		items := []*string{}
		str1 := "first"
		str2 := "second"
		items = append(items, &str1, &str2)
		result := ForEach(items, func(item *string, index int) string {
			return fmt.Sprintf("%d:%s,", index, *item)
		}).Render()
		assert.Equal(t, "0:first,1:second,", result)
	})
}

// TestForEachDirective_NestedForEach tests nested ForEach directives
func TestForEachDirective_NestedForEach(t *testing.T) {
	type Category struct {
		Name  string
		Items []string
	}

	categories := []Category{
		{Name: "Fruits", Items: []string{"Apple", "Banana"}},
		{Name: "Vegetables", Items: []string{"Carrot", "Lettuce"}},
	}

	result := ForEach(categories, func(cat Category, i int) string {
		header := fmt.Sprintf("%s:\n", cat.Name)
		items := ForEach(cat.Items, func(item string, j int) string {
			return fmt.Sprintf("  - %s\n", item)
		}).Render()
		return header + items
	}).Render()

	expected := "Fruits:\n  - Apple\n  - Banana\nVegetables:\n  - Carrot\n  - Lettuce\n"
	assert.Equal(t, expected, result)
}

// TestForEachDirective_ComplexContent tests complex content rendering
func TestForEachDirective_ComplexContent(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		render   func(string, int) string
		expected string
	}{
		{
			name:  "multiline content",
			items: []string{"A", "B"},
			render: func(item string, index int) string {
				return fmt.Sprintf("Item %d:\n  Value: %s\n  Index: %d\n", index, item, index)
			},
			expected: "Item 0:\n  Value: A\n  Index: 0\nItem 1:\n  Value: B\n  Index: 1\n",
		},
		{
			name:  "special characters",
			items: []string{"hello\tworld", "foo\nbar", "baz\\qux"},
			render: func(item string, index int) string {
				return fmt.Sprintf("[%d]%s|", index, item)
			},
			expected: "[0]hello\tworld|[1]foo\nbar|[2]baz\\qux|",
		},
		{
			name:  "unicode content",
			items: []string{"🎉", "🚀", "💻"},
			render: func(item string, index int) string {
				return fmt.Sprintf("%s ", item)
			},
			expected: "🎉 🚀 💻 ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ForEach(tt.items, tt.render).Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestForEachDirective_EmptyContent tests empty return values from render function
func TestForEachDirective_EmptyContent(t *testing.T) {
	t.Run("all items return empty", func(t *testing.T) {
		items := []string{"A", "B", "C"}
		result := ForEach(items, func(item string, index int) string {
			return ""
		}).Render()
		assert.Equal(t, "", result)
	})

	t.Run("some items return empty", func(t *testing.T) {
		items := []string{"A", "B", "C"}
		result := ForEach(items, func(item string, index int) string {
			if index%2 == 0 {
				return item
			}
			return ""
		}).Render()
		assert.Equal(t, "AC", result)
	})
}

// TestForEachDirective_InterfaceCompliance tests that ForEachDirective implements Directive
func TestForEachDirective_InterfaceCompliance(t *testing.T) {
	items := []string{"test"}
	directive := ForEach(items, func(item string, index int) string {
		return item
	})

	// Verify it implements Directive interface
	var _ Directive = directive

	// Verify Render method works
	result := directive.Render()
	assert.Equal(t, "test", result)
}

// TestForEachDirective_LargeSlice tests performance with larger collections
func TestForEachDirective_LargeSlice(t *testing.T) {
	// Create a slice with 1000 items
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}

	result := ForEach(items, func(item int, index int) string {
		return fmt.Sprintf("%d,", item)
	}).Render()

	// Verify it contains all items
	assert.Contains(t, result, "0,")
	assert.Contains(t, result, "999,")
	assert.Equal(t, 1000, strings.Count(result, ","))
}

// TestForEachDirective_CompositionWithIf tests composition with If directive
func TestForEachDirective_CompositionWithIf(t *testing.T) {
	items := []string{"A", "B", "C"}
	showList := true

	result := If(showList, func() string {
		return ForEach(items, func(item string, index int) string {
			return fmt.Sprintf("%s ", item)
		}).Render()
	}).Else(func() string {
		return "No items"
	}).Render()

	assert.Equal(t, "A B C ", result)

	// Test with showList = false
	showList = false
	result = If(showList, func() string {
		return ForEach(items, func(item string, index int) string {
			return fmt.Sprintf("%s ", item)
		}).Render()
	}).Else(func() string {
		return "No items"
	}).Render()

	assert.Equal(t, "No items", result)
}

// TestForEachDirective_CompositionWithShow tests composition with Show directive
func TestForEachDirective_CompositionWithShow(t *testing.T) {
	items := []string{"X", "Y", "Z"}
	visible := true

	result := Show(visible, func() string {
		return ForEach(items, func(item string, index int) string {
			return item
		}).Render()
	}).Render()

	assert.Equal(t, "XYZ", result)

	// Test with visible = false
	visible = false
	result = Show(visible, func() string {
		return ForEach(items, func(item string, index int) string {
			return item
		}).Render()
	}).Render()

	assert.Equal(t, "", result)
}

// TestForEachDirective_IndexUsage tests proper index handling
func TestForEachDirective_IndexUsage(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}
	result := ForEach(items, func(item string, index int) string {
		return fmt.Sprintf("[%d]", index)
	}).Render()
	assert.Equal(t, "[0][1][2][3][4]", result)
}

// ============================================================================
// Benchmark Tests
// ============================================================================

// BenchmarkForEach10Items benchmarks ForEach directive with 10 items
// Target: < 100μs
func BenchmarkForEach10Items(b *testing.B) {
	items := make([]int, 10)
	for i := range items {
		items[i] = i
	}

	render := func(item int, index int) string {
		return fmt.Sprintf("%d:%d,", index, item)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ForEach(items, render).Render()
	}
}

// BenchmarkForEach100Items benchmarks ForEach directive with 100 items
// Target: < 1ms
func BenchmarkForEach100Items(b *testing.B) {
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	render := func(item int, index int) string {
		return fmt.Sprintf("%d:%d,", index, item)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ForEach(items, render).Render()
	}
}

// BenchmarkForEach1000Items benchmarks ForEach directive with 1000 items
// Target: < 10ms
func BenchmarkForEach1000Items(b *testing.B) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}

	render := func(item int, index int) string {
		return fmt.Sprintf("%d:%d,", index, item)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ForEach(items, render).Render()
	}
}

// BenchmarkForEachString benchmarks ForEach with string concatenation
func BenchmarkForEachString(b *testing.B) {
	items := make([]string, 100)
	for i := range items {
		items[i] = fmt.Sprintf("item-%d", i)
	}

	render := func(item string, index int) string {
		return fmt.Sprintf("%d. %s\n", index+1, item)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ForEach(items, render).Render()
	}
}

// BenchmarkForEachStruct benchmarks ForEach with struct iteration
func BenchmarkForEachStruct(b *testing.B) {
	type User struct {
		Name  string
		Email string
	}

	items := make([]User, 100)
	for i := range items {
		items[i] = User{
			Name:  fmt.Sprintf("User%d", i),
			Email: fmt.Sprintf("user%d@example.com", i),
		}
	}

	render := func(user User, index int) string {
		return fmt.Sprintf("%d. %s <%s>\n", index+1, user.Name, user.Email)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ForEach(items, render).Render()
	}
}

// BenchmarkForEachNested benchmarks nested ForEach directives
func BenchmarkForEachNested(b *testing.B) {
	type Category struct {
		Name  string
		Items []string
	}

	categories := make([]Category, 10)
	for i := range categories {
		categories[i] = Category{
			Name:  fmt.Sprintf("Category%d", i),
			Items: []string{"A", "B", "C", "D", "E"},
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ForEach(categories, func(cat Category, idx int) string {
			header := fmt.Sprintf("%s:\n", cat.Name)
			items := ForEach(cat.Items, func(item string, j int) string {
				return fmt.Sprintf("  - %s\n", item)
			}).Render()
			return header + items
		}).Render()
	}
}

// TestForEachDirective_PanicRecovery tests that ForEach recovers from panics in render functions
func TestForEachDirective_PanicRecovery(t *testing.T) {
	t.Run("panic in render function is recovered", func(t *testing.T) {
		items := []string{"A", "B", "C"}
		result := ForEach(items, func(item string, index int) string {
			if index == 1 {
				panic("test panic")
			}
			return item
		}).Render()

		// The directive should not crash, but return partial results
		// The panicking item returns empty string
		assert.Equal(t, "AC", result)
	})

	t.Run("panic at first index is recovered", func(t *testing.T) {
		items := []string{"A", "B", "C"}
		result := ForEach(items, func(item string, index int) string {
			if index == 0 {
				panic("panic at first")
			}
			return item
		}).Render()

		assert.Equal(t, "BC", result)
	})

	t.Run("panic at last index is recovered", func(t *testing.T) {
		items := []string{"A", "B", "C"}
		result := ForEach(items, func(item string, index int) string {
			if index == 2 {
				panic("panic at last")
			}
			return item
		}).Render()

		assert.Equal(t, "AB", result)
	})

	t.Run("all items panic returns empty strings", func(t *testing.T) {
		items := []string{"A", "B", "C"}
		result := ForEach(items, func(item string, index int) string {
			panic("always panic")
		}).Render()

		assert.Equal(t, "", result)
	})

	t.Run("nil pointer panic is recovered", func(t *testing.T) {
		items := []string{"A", "B"}
		result := ForEach(items, func(item string, index int) string {
			if index == 0 {
				var ptr *string
				return *ptr // nil pointer dereference
			}
			return item
		}).Render()

		assert.Equal(t, "B", result)
	})

	t.Run("panic with non-string value is recovered", func(t *testing.T) {
		items := []int{1, 2, 3}
		result := ForEach(items, func(item int, index int) string {
			if index == 1 {
				panic(42) // panic with int value
			}
			return fmt.Sprintf("%d", item)
		}).Render()

		assert.Equal(t, "13", result)
	})
}

// TestForEachDirective_PanicRecoveryWithReporter tests panic recovery with error reporter
func TestForEachDirective_PanicRecoveryWithReporter(t *testing.T) {
	t.Run("panic is reported when reporter is set", func(t *testing.T) {
		// Set up mock reporter
		reporter := &mockReporter{}
		observability.SetErrorReporter(reporter)
		defer observability.SetErrorReporter(nil)

		items := []string{"A", "B", "C"}
		result := ForEach(items, func(item string, index int) string {
			if index == 1 {
				panic("test panic for reporter")
			}
			return item
		}).Render()

		// Directive should still recover and continue
		assert.Equal(t, "AC", result)

		// Verify the error was reported
		assert.Equal(t, 1, reporter.getErrorCallCount(), "panic should be reported to error reporter")
	})

	t.Run("multiple panics are all reported", func(t *testing.T) {
		// Set up mock reporter
		reporter := &mockReporter{}
		observability.SetErrorReporter(reporter)
		defer observability.SetErrorReporter(nil)

		items := []string{"A", "B", "C", "D"}
		result := ForEach(items, func(item string, index int) string {
			if index == 1 || index == 3 {
				panic(fmt.Sprintf("panic at index %d", index))
			}
			return item
		}).Render()

		// Directive should still recover
		assert.Equal(t, "AC", result)

		// Verify both errors were reported
		assert.Equal(t, 2, reporter.getErrorCallCount(), "all panics should be reported")
	})

	t.Run("error context contains directive info", func(t *testing.T) {
		// Set up mock reporter
		reporter := &mockReporter{}
		observability.SetErrorReporter(reporter)
		defer observability.SetErrorReporter(nil)

		items := []string{"A", "B"}
		_ = ForEach(items, func(item string, index int) string {
			if index == 0 {
				panic("panic for context check")
			}
			return item
		}).Render()

		// Verify context was set correctly
		assert.Equal(t, 1, reporter.getErrorCallCount())
		reporter.mu.Lock()
		defer reporter.mu.Unlock()
		if len(reporter.errorCalls) > 0 {
			ctx := reporter.errorCalls[0].ctx
			assert.Equal(t, "ForEach", ctx.ComponentName)
			assert.Contains(t, ctx.Tags, "directive_type")
			assert.Equal(t, "ForEach", ctx.Tags["directive_type"])
		}
	})
}
