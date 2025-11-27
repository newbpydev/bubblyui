package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestNewProvideInjectTester tests creating a new ProvideInjectTester
func TestNewProvideInjectTester(t *testing.T) {
	// Create a root context
	rootCtx := bubbly.NewTestContext()

	// Create tester
	tester := NewProvideInjectTester(rootCtx)

	// Assert tester created
	assert.NotNil(t, tester)
	assert.NotNil(t, tester.root)
	assert.NotNil(t, tester.providers)
	assert.NotNil(t, tester.injections)
	assert.Equal(t, rootCtx, tester.root)
}

// TestProvideInjectTester_Provide tests providing values
func TestProvideInjectTester_Provide(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{
			name:  "provide string",
			key:   "theme",
			value: "dark",
		},
		{
			name:  "provide int",
			key:   "count",
			value: 42,
		},
		{
			name:  "provide struct",
			key:   "config",
			value: struct{ Name string }{Name: "test"},
		},
		{
			name:  "provide ref",
			key:   "userRef",
			value: bubbly.NewRef("John"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create root context
			rootCtx := bubbly.NewTestContext()

			// Create tester
			tester := NewProvideInjectTester(rootCtx)

			// Provide value
			tester.Provide(tt.key, tt.value)

			// Assert value stored
			assert.Equal(t, tt.value, tester.providers[tt.key])
		})
	}
}

// TestProvideInjectTester_Inject tests injecting values
func TestProvideInjectTester_Inject(t *testing.T) {
	// Create root context with provided value
	rootCtx := bubbly.NewTestContext()
	rootCtx.Provide("theme", "dark")

	// Create child context
	childCtx := bubbly.NewTestContext()
	bubbly.SetParent(childCtx, rootCtx)

	// Create tester
	tester := NewProvideInjectTester(rootCtx)

	// Inject value
	value := tester.Inject(childCtx, "theme")

	// Assert value injected
	assert.Equal(t, "dark", value)
}

// TestProvideInjectTester_Inject_DefaultValue tests injection with default value
func TestProvideInjectTester_Inject_DefaultValue(t *testing.T) {
	// Create root context without providing value
	rootCtx := bubbly.NewTestContext()

	// Create child context
	childCtx := bubbly.NewTestContext()
	bubbly.SetParent(childCtx, rootCtx)

	// Create tester
	tester := NewProvideInjectTester(rootCtx)

	// Inject value with default
	value := tester.Inject(childCtx, "theme")

	// Assert default value returned (nil when not found)
	assert.Nil(t, value)
}

// TestProvideInjectTester_Inject_TreeTraversal tests injection across multiple levels
func TestProvideInjectTester_Inject_TreeTraversal(t *testing.T) {
	// Create root context with provided value
	rootCtx := bubbly.NewTestContext()
	rootCtx.Provide("theme", "dark")

	// Create middle context (no provide)
	middleCtx := bubbly.NewTestContext()
	bubbly.SetParent(middleCtx, rootCtx)

	// Create grandchild context
	grandchildCtx := bubbly.NewTestContext()
	bubbly.SetParent(grandchildCtx, middleCtx)

	// Create tester
	tester := NewProvideInjectTester(rootCtx)

	// Inject value from grandchild (should traverse to root)
	value := tester.Inject(grandchildCtx, "theme")

	// Assert value injected from root
	assert.Equal(t, "dark", value)
}

// TestProvideInjectTester_Inject_NearestProvider tests that nearest provider wins
func TestProvideInjectTester_Inject_NearestProvider(t *testing.T) {
	// Create root context with provided value
	rootCtx := bubbly.NewTestContext()
	rootCtx.Provide("theme", "dark")

	// Create middle context with override
	middleCtx := bubbly.NewTestContext()
	middleCtx.Provide("theme", "light") // Override parent
	bubbly.SetParent(middleCtx, rootCtx)

	// Create grandchild context
	grandchildCtx := bubbly.NewTestContext()
	bubbly.SetParent(grandchildCtx, middleCtx)

	// Create tester
	tester := NewProvideInjectTester(rootCtx)

	// Inject value from grandchild (should get from middle, not root)
	value := tester.Inject(grandchildCtx, "theme")

	// Assert value from nearest provider (middle)
	assert.Equal(t, "light", value)
}

// TestProvideInjectTester_AssertInjected tests assertion helper
func TestProvideInjectTester_AssertInjected(t *testing.T) {
	// Create root context with provided value
	rootCtx := bubbly.NewTestContext()
	rootCtx.Provide("theme", "dark")

	// Create child context
	childCtx := bubbly.NewTestContext()
	bubbly.SetParent(childCtx, rootCtx)

	// Create tester
	tester := NewProvideInjectTester(rootCtx)

	// Assert injected value (should pass)
	tester.AssertInjected(t, childCtx, "theme", "dark")
}

// TestProvideInjectTester_AssertInjected_Failure tests assertion failure
func TestProvideInjectTester_AssertInjected_Failure(t *testing.T) {
	// Create root context with provided value
	rootCtx := bubbly.NewTestContext()
	rootCtx.Provide("theme", "dark")

	// Create child context
	childCtx := bubbly.NewTestContext()
	bubbly.SetParent(childCtx, rootCtx)

	// Create tester
	tester := NewProvideInjectTester(rootCtx)

	// Create mock testing.T to capture error
	mockT := &mockTestingT{}

	// Assert wrong value (should fail)
	tester.AssertInjected(mockT, childCtx, "theme", "light")

	// Verify error was called
	assert.True(t, len(mockT.errors) > 0, "Expected error to be recorded")
	assert.Contains(t, mockT.errors[0], "expected injected value")
}

// TestProvideInjectTester_MultipleInjections tests tracking multiple injections
func TestProvideInjectTester_MultipleInjections(t *testing.T) {
	// Create root context with multiple provided values
	rootCtx := bubbly.NewTestContext()
	rootCtx.Provide("theme", "dark")
	rootCtx.Provide("locale", "en-US")
	rootCtx.Provide("user", "admin")

	// Create child context
	childCtx := bubbly.NewTestContext()
	bubbly.SetParent(childCtx, rootCtx)

	// Create tester
	tester := NewProvideInjectTester(rootCtx)

	// Inject multiple values
	theme := tester.Inject(childCtx, "theme")
	locale := tester.Inject(childCtx, "locale")
	user := tester.Inject(childCtx, "user")

	// Assert all values injected correctly
	assert.Equal(t, "dark", theme)
	assert.Equal(t, "en-US", locale)
	assert.Equal(t, "admin", user)

	// Verify injections tracked
	assert.Len(t, tester.injections, 3)
	assert.Contains(t, tester.injections, "theme")
	assert.Contains(t, tester.injections, "locale")
	assert.Contains(t, tester.injections, "user")
}
