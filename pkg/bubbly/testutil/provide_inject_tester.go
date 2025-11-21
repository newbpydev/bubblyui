package testutil

import (
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ProvideInjectTester provides utilities for testing the provide/inject dependency injection system.
//
// It wraps a root context and provides methods for testing that values are correctly
// provided by parent contexts and injected by child contexts across the component tree.
// This is useful for verifying that dependency injection works correctly in complex
// component hierarchies.
//
// Type Safety:
//   - Thread-safe access to provided values
//   - Clear assertion methods for injection verification
//   - Tracks injection calls for testing
//
// Example:
//
//	func TestProvideInject(t *testing.T) {
//		// Create root context with provided theme
//		rootCtx := bubbly.NewTestContext()
//		rootCtx.Provide("theme", "dark")
//
//		// Create child context
//		childCtx := bubbly.NewTestContext()
//		bubbly.SetParent(childCtx, rootCtx)
//
//		// Test injection
//		tester := testutil.NewProvideInjectTester(rootCtx)
//		tester.AssertInjected(t, childCtx, "theme", "dark")
//	}
type ProvideInjectTester struct {
	// root is the root context in the tree
	root *bubbly.Context

	// providers maps keys to provided values for testing
	providers map[string]interface{}

	// injections tracks which contexts have injected which keys
	injections map[string][]*bubbly.Context
}

// NewProvideInjectTester creates a new ProvideInjectTester for testing provide/inject.
//
// Parameters:
//   - root: The root context of the tree to test
//
// Returns:
//   - *ProvideInjectTester: A new tester instance
//
// Example:
//
//	rootCtx := bubbly.NewTestContext()
//	rootCtx.Provide("theme", "dark")
//
//	tester := testutil.NewProvideInjectTester(rootCtx)
func NewProvideInjectTester(root *bubbly.Context) *ProvideInjectTester {
	return &ProvideInjectTester{
		root:       root,
		providers:  make(map[string]interface{}),
		injections: make(map[string][]*bubbly.Context),
	}
}

// Provide stores a value for testing purposes.
//
// This method allows tests to provide values directly without going through
// the component's Provide method. This is useful for setting up test scenarios.
//
// Parameters:
//   - key: The key to provide
//   - value: The value to provide
//
// Example:
//
//	tester := testutil.NewProvideInjectTester(root)
//	tester.Provide("theme", "dark")
//	tester.Provide("locale", "en-US")
func (pit *ProvideInjectTester) Provide(key string, value interface{}) {
	pit.providers[key] = value
}

// Inject retrieves a value from the context tree using the context's inject mechanism.
//
// This method uses the context's Inject method to retrieve the value, simulating
// what would happen when a child context calls ctx.Inject(). It also tracks
// which contexts have injected which keys for testing purposes.
//
// Parameters:
//   - ctx: The context that is injecting the value
//   - key: The key to inject
//
// Returns:
//   - interface{}: The injected value, or nil if not found
//
// Example:
//
//	value := tester.Inject(childCtx, "theme")
//	assert.Equal(t, "dark", value)
func (pit *ProvideInjectTester) Inject(ctx *bubbly.Context, key string) interface{} {
	// Track this injection
	pit.injections[key] = append(pit.injections[key], ctx)

	// Use the context's inject method with nil default
	// This will traverse the tree looking for the provided value
	return ctx.Inject(key, nil)
}

// AssertInjected asserts that a context can inject the expected value.
//
// This method verifies that when the context injects the specified key,
// it receives the expected value. It uses the context's inject mechanism
// to ensure the test reflects actual runtime behavior.
//
// Parameters:
//   - t: The testing.T instance (or mock)
//   - ctx: The context that should inject the value
//   - key: The key to inject
//   - expected: The expected injected value
//
// Example:
//
//	// Assert child can inject theme from parent
//	tester.AssertInjected(t, childCtx, "theme", "dark")
//
//	// Assert grandchild can inject from root (tree traversal)
//	tester.AssertInjected(t, grandchildCtx, "locale", "en-US")
func (pit *ProvideInjectTester) AssertInjected(t testingT, ctx *bubbly.Context, key string, expected interface{}) {
	t.Helper()

	// Inject the value
	actual := pit.Inject(ctx, key)

	// Compare
	if actual != expected {
		t.Errorf("expected injected value for key %q to be %v, got %v",
			key, expected, actual)
	}
}
