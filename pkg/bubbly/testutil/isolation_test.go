package testutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// TestNewTestIsolation tests creating a new TestIsolation instance
func TestNewTestIsolation(t *testing.T) {
	isolation := NewTestIsolation()

	assert.NotNil(t, isolation)
	assert.NotNil(t, isolation.savedGlobals)
	assert.Equal(t, 0, len(isolation.savedGlobals))
}

// TestTestIsolation_Isolate_SavesGlobalHook tests that Isolate saves the current global hook
func TestTestIsolation_Isolate_SavesGlobalHook(t *testing.T) {
	// Setup: Register a hook
	testHook := &mockFrameworkHook{}
	_ = bubbly.RegisterHook(testHook)
	defer func() { _ = bubbly.UnregisterHook() }()

	// Create isolation and isolate
	isolation := NewTestIsolation()
	isolation.Isolate(t)

	// Verify hook was saved
	assert.Contains(t, isolation.savedGlobals, "frameworkHook")
	assert.Equal(t, testHook, isolation.savedGlobals["frameworkHook"])
}

// TestTestIsolation_Isolate_ClearsGlobalHook tests that Isolate clears the global hook
func TestTestIsolation_Isolate_ClearsGlobalHook(t *testing.T) {
	// Setup: Register a hook
	testHook := &mockFrameworkHook{}
	_ = bubbly.RegisterHook(testHook)
	defer func() { _ = bubbly.UnregisterHook() }()

	// Verify hook is registered
	assert.True(t, bubbly.IsHookRegistered())

	// Create isolation and isolate
	isolation := NewTestIsolation()
	isolation.Isolate(t)

	// Verify hook was cleared
	assert.False(t, bubbly.IsHookRegistered())
}

// TestTestIsolation_Restore_RestoresGlobalHook tests that Restore restores the saved hook
func TestTestIsolation_Restore_RestoresGlobalHook(t *testing.T) {
	// Setup: Register a hook
	testHook := &mockFrameworkHook{}
	_ = bubbly.RegisterHook(testHook)
	defer func() { _ = bubbly.UnregisterHook() }()

	// Create isolation, isolate, then restore
	isolation := NewTestIsolation()
	isolation.Isolate(t)

	// Verify hook was cleared
	assert.False(t, bubbly.IsHookRegistered())

	// Restore
	isolation.Restore()

	// Verify hook was restored
	assert.True(t, bubbly.IsHookRegistered())
}

// TestTestIsolation_Restore_WithNoHook tests restoring when no hook was registered
func TestTestIsolation_Restore_WithNoHook(t *testing.T) {
	// Ensure no hook is registered
	_ = bubbly.UnregisterHook()

	// Create isolation and isolate
	isolation := NewTestIsolation()
	isolation.Isolate(t)

	// Restore (should not panic)
	isolation.Restore()

	// Verify still no hook
	assert.False(t, bubbly.IsHookRegistered())
}

// TestTestIsolation_Isolate_SavesErrorReporter tests that Isolate saves the error reporter
func TestTestIsolation_Isolate_SavesErrorReporter(t *testing.T) {
	// Setup: Set an error reporter
	reporter := observability.NewConsoleReporter(false)
	observability.SetErrorReporter(reporter)
	defer observability.SetErrorReporter(nil)

	// Create isolation and isolate
	isolation := NewTestIsolation()
	isolation.Isolate(t)

	// Verify reporter was saved
	assert.Contains(t, isolation.savedGlobals, "errorReporter")
	assert.Equal(t, reporter, isolation.savedGlobals["errorReporter"])
}

// TestTestIsolation_Isolate_ClearsErrorReporter tests that Isolate clears the error reporter
func TestTestIsolation_Isolate_ClearsErrorReporter(t *testing.T) {
	// Setup: Set an error reporter
	reporter := observability.NewConsoleReporter(false)
	observability.SetErrorReporter(reporter)
	defer observability.SetErrorReporter(nil)

	// Verify reporter is set
	assert.NotNil(t, observability.GetErrorReporter())

	// Create isolation and isolate
	isolation := NewTestIsolation()
	isolation.Isolate(t)

	// Verify reporter was cleared
	assert.Nil(t, observability.GetErrorReporter())
}

// TestTestIsolation_Restore_RestoresErrorReporter tests that Restore restores the error reporter
func TestTestIsolation_Restore_RestoresErrorReporter(t *testing.T) {
	// Setup: Set an error reporter
	reporter := observability.NewConsoleReporter(false)
	observability.SetErrorReporter(reporter)
	defer observability.SetErrorReporter(nil)

	// Create isolation, isolate, then restore
	isolation := NewTestIsolation()
	isolation.Isolate(t)

	// Verify reporter was cleared
	assert.Nil(t, observability.GetErrorReporter())

	// Restore
	isolation.Restore()

	// Verify reporter was restored
	assert.NotNil(t, observability.GetErrorReporter())
	assert.Equal(t, reporter, observability.GetErrorReporter())
}

// TestTestIsolation_AutomaticCleanup tests that cleanup is automatic with t.Cleanup
func TestTestIsolation_AutomaticCleanup(t *testing.T) {
	// Setup: Register a hook
	testHook := &mockFrameworkHook{}
	_ = bubbly.RegisterHook(testHook)
	defer func() { _ = bubbly.UnregisterHook() }()

	// Run in subtest to test cleanup
	t.Run("subtest", func(t *testing.T) {
		isolation := NewTestIsolation()
		isolation.Isolate(t)

		// Verify hook was cleared
		assert.False(t, bubbly.IsHookRegistered())

		// When subtest ends, t.Cleanup should restore the hook
	})

	// After subtest completes, hook should be restored
	assert.True(t, bubbly.IsHookRegistered())
}

// TestTestIsolation_ParallelTests tests that isolation works with parallel tests
func TestTestIsolation_ParallelTests(t *testing.T) {
	// Setup: Register a hook
	testHook := &mockFrameworkHook{}
	_ = bubbly.RegisterHook(testHook)
	defer func() { _ = bubbly.UnregisterHook() }()

	// Run parallel subtests
	t.Run("test1", func(t *testing.T) {
		t.Parallel()

		isolation := NewTestIsolation()
		isolation.Isolate(t)

		// Each test has isolated state
		assert.False(t, bubbly.IsHookRegistered())
	})

	t.Run("test2", func(t *testing.T) {
		t.Parallel()

		isolation := NewTestIsolation()
		isolation.Isolate(t)

		// Each test has isolated state
		assert.False(t, bubbly.IsHookRegistered())
	})
}

// TestTestIsolation_MultipleIsolations tests multiple isolations don't interfere
func TestTestIsolation_MultipleIsolations(t *testing.T) {
	// Setup: Register a hook
	testHook := &mockFrameworkHook{}
	_ = bubbly.RegisterHook(testHook)
	defer func() { _ = bubbly.UnregisterHook() }()

	// First isolation
	isolation1 := NewTestIsolation()
	isolation1.Isolate(t)

	// Register a different hook
	testHook2 := &mockFrameworkHook{}
	_ = bubbly.RegisterHook(testHook2)

	// Second isolation
	isolation2 := NewTestIsolation()
	isolation2.Isolate(t)

	// Restore second (should restore testHook2)
	isolation2.Restore()
	assert.True(t, bubbly.IsHookRegistered())

	// Restore first (should restore testHook)
	isolation1.Restore()
	assert.True(t, bubbly.IsHookRegistered())
}

// TestTestIsolation_EmptyRestore tests restoring without isolating first
func TestTestIsolation_EmptyRestore(t *testing.T) {
	isolation := NewTestIsolation()

	// Restore without isolating (should not panic)
	isolation.Restore()

	// No assertions needed, just verify no panic
}

// mockFrameworkHook is a simple mock for testing
type mockFrameworkHook struct{}

func (m *mockFrameworkHook) OnComponentMount(id, name string)                      {}
func (m *mockFrameworkHook) OnComponentUpdate(id string, msg interface{})          {}
func (m *mockFrameworkHook) OnComponentUnmount(id string)                          {}
func (m *mockFrameworkHook) OnRefChange(id string, oldValue, newValue interface{}) {}
func (m *mockFrameworkHook) OnRefExposed(componentID, refID, refName string)       {}
func (m *mockFrameworkHook) OnRenderComplete(componentID string, duration time.Duration) {
}
func (m *mockFrameworkHook) OnComputedChange(id string, oldValue, newValue interface{})       {}
func (m *mockFrameworkHook) OnWatchCallback(watcherID string, oldValue, newValue interface{}) {}
func (m *mockFrameworkHook) OnEffectRun(effectID string)                                      {}
func (m *mockFrameworkHook) OnChildAdded(parentID, childID string)                            {}
func (m *mockFrameworkHook) OnChildRemoved(parentID, childID string)                          {}
func (m *mockFrameworkHook) OnEvent(componentID, eventName string, data interface{})          {}
