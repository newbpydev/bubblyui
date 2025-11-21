package testutil

import "testing"

// TestSetup provides a fluent API for managing test setup and teardown functions.
// It allows registering multiple setup functions that execute before the test
// and teardown functions that execute after the test (via t.Cleanup).
//
// Setup functions execute in the order they were added (FIFO).
// Teardown functions execute in reverse order (LIFO, like defer statements).
//
// This pattern is useful for:
//   - Initializing test fixtures
//   - Setting up mock objects
//   - Configuring test environment
//   - Cleaning up resources
//   - Restoring global state
//
// Example usage:
//
//	setup := testutil.NewTestSetup().
//	    AddSetup(func(t *testing.T) {
//	        // Initialize database
//	    }).
//	    AddTeardown(func(t *testing.T) {
//	        // Close database
//	    })
//
//	setup.Run(t, func(t *testing.T) {
//	    // Your test code here
//	})
type TestSetup struct {
	// setupFuncs contains functions to run before the test.
	// They execute in the order they were added (FIFO).
	setupFuncs []func(*testing.T)

	// teardownFuncs contains functions to run after the test.
	// They execute in reverse order (LIFO, like defer).
	teardownFuncs []func(*testing.T)
}

// NewTestSetup creates a new TestSetup with empty setup and teardown lists.
// This is the entry point for using the fluent API.
//
// Example:
//
//	setup := testutil.NewTestSetup()
//	// Now chain configuration methods...
//	setup.AddSetup(...).AddTeardown(...).Run(t, testFunc)
//
// Returns:
//   - *TestSetup: A new setup instance ready for configuration
func NewTestSetup() *TestSetup {
	return &TestSetup{
		setupFuncs:    []func(*testing.T){},
		teardownFuncs: []func(*testing.T){},
	}
}

// AddSetup adds a setup function to execute before the test.
// Setup functions execute in the order they were added (FIFO).
//
// Returns the TestSetup for method chaining.
//
// Example:
//
//	setup := testutil.NewTestSetup().
//	    AddSetup(func(t *testing.T) {
//	        t.Log("First setup")
//	    }).
//	    AddSetup(func(t *testing.T) {
//	        t.Log("Second setup")
//	    })
//
// Parameters:
//   - fn: Function to execute before the test
//
// Returns:
//   - *TestSetup: Self for method chaining
func (ts *TestSetup) AddSetup(fn func(*testing.T)) *TestSetup {
	ts.setupFuncs = append(ts.setupFuncs, fn)
	return ts
}

// AddTeardown adds a teardown function to execute after the test.
// Teardown functions execute in reverse order (LIFO, like defer statements).
// They are registered with t.Cleanup() to ensure execution even if the test panics.
//
// Returns the TestSetup for method chaining.
//
// Example:
//
//	setup := testutil.NewTestSetup().
//	    AddTeardown(func(t *testing.T) {
//	        t.Log("First teardown (executes last)")
//	    }).
//	    AddTeardown(func(t *testing.T) {
//	        t.Log("Second teardown (executes first)")
//	    })
//
// Parameters:
//   - fn: Function to execute after the test
//
// Returns:
//   - *TestSetup: Self for method chaining
func (ts *TestSetup) AddTeardown(fn func(*testing.T)) *TestSetup {
	ts.teardownFuncs = append(ts.teardownFuncs, fn)
	return ts
}

// Run executes the test with setup and teardown functions.
//
// Execution order:
//  1. Execute all setup functions in order (FIFO)
//  2. Register all teardown functions with t.Cleanup (they execute in LIFO order)
//  3. Execute the test function
//  4. Teardown functions execute automatically via t.Cleanup in LIFO order
//
// The teardown functions are registered with t.Cleanup() before running the test,
// ensuring they execute even if the test panics or fails.
//
// Note: t.Cleanup() executes cleanup functions in LIFO order (last registered, first executed),
// so we register teardown functions in forward order to achieve LIFO execution relative
// to the order they were added.
//
// Example:
//
//	setup := testutil.NewTestSetup().
//	    AddSetup(func(t *testing.T) {
//	        // Initialize resources
//	    }).
//	    AddTeardown(func(t *testing.T) {
//	        // Cleanup resources
//	    })
//
//	setup.Run(t, func(t *testing.T) {
//	    // Test code here
//	    assert.True(t, true)
//	})
//
// Parameters:
//   - t: The testing.T instance for test context
//   - testFn: The test function to execute
func (ts *TestSetup) Run(t *testing.T, testFn func(*testing.T)) {
	// Execute setup functions in order (FIFO)
	for _, setupFn := range ts.setupFuncs {
		setupFn(t)
	}

	// Register teardown functions in forward order
	// t.Cleanup executes in LIFO order, so registering in forward order
	// means the last teardown added will execute first (LIFO behavior)
	for _, teardownFn := range ts.teardownFuncs {
		// Capture the function in the closure
		fn := teardownFn
		t.Cleanup(func() {
			fn(t)
		})
	}

	// Execute the test function
	testFn(t)
}
