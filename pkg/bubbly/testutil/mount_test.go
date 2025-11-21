package testutil

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestComponent creates a simple component for testing
func createTestComponent(name string) bubbly.Component {
	component, err := bubbly.NewComponent(name).
		Setup(func(ctx *bubbly.Context) {
			// Create a simple ref for testing
			ctx.Ref("initial")
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test component"
		}).
		Build()

	if err != nil {
		panic(err) // Should never happen in tests
	}

	return component
}

// TestMount_ComponentMountsCorrectly tests that Mount creates a ComponentTest
func TestMount_ComponentMountsCorrectly(t *testing.T) {
	harness := NewHarness(t)
	component := createTestComponent("TestComponent")

	ct := harness.Mount(component)

	require.NotNil(t, ct, "ComponentTest should not be nil")
	assert.NotNil(t, ct.harness, "harness reference should be set")
	assert.NotNil(t, ct.component, "component should be set")
	assert.NotNil(t, ct.state, "state inspector should be set")
	assert.NotNil(t, ct.events, "event inspector should be set")
}

// TestMount_InitCalledAutomatically tests that Init() is called during mount
func TestMount_InitCalledAutomatically(t *testing.T) {
	harness := NewHarness(t)

	setupCalled := false
	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			setupCalled = true
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)

	// Before mount, setup should not have been called
	assert.False(t, setupCalled, "setup should not be called before mount")

	ct := harness.Mount(component)

	// After mount, Init() should have been called, triggering setup
	require.NotNil(t, ct)
	assert.True(t, setupCalled, "setup should be called during mount")
}

// TestMount_ComponentStoredInHarness tests that component is stored in harness
func TestMount_ComponentStoredInHarness(t *testing.T) {
	harness := NewHarness(t)
	component := createTestComponent("TestComponent")

	ct := harness.Mount(component)

	require.NotNil(t, ct)
	assert.Equal(t, component, harness.component, "component should be stored in harness")
	assert.Equal(t, component, ct.component, "component should be accessible from ComponentTest")
}

// TestMount_StateInspectorCreated tests that StateInspector is created
func TestMount_StateInspectorCreated(t *testing.T) {
	harness := NewHarness(t)
	component := createTestComponent("TestComponent")

	ct := harness.Mount(component)

	require.NotNil(t, ct)
	require.NotNil(t, ct.state, "state inspector should be created")
	// StateInspector should have access to harness refs
	assert.NotNil(t, ct.state.refs, "state inspector should have refs")
}

// TestMount_EventInspectorCreated tests that EventInspector is created
func TestMount_EventInspectorCreated(t *testing.T) {
	harness := NewHarness(t)
	component := createTestComponent("TestComponent")

	ct := harness.Mount(component)

	require.NotNil(t, ct)
	require.NotNil(t, ct.events, "event inspector should be created")
	// EventInspector should have access to harness event tracker
	assert.NotNil(t, ct.events.tracker, "event inspector should have tracker")
}

// TestMount_MultipleComponents tests mounting multiple components
func TestMount_MultipleComponents(t *testing.T) {
	harness := NewHarness(t)

	ct1 := harness.Mount(createTestComponent("Component1"))
	ct2 := harness.Mount(createTestComponent("Component2"))

	require.NotNil(t, ct1)
	require.NotNil(t, ct2)

	// Each ComponentTest should have its own component
	assert.NotEqual(t, ct1.component, ct2.component, "components should be different")

	// But they should share the same harness
	assert.Equal(t, ct1.harness, ct2.harness, "should share same harness")
}

// TestUnmount_CleanupRegistered tests that Unmount registers cleanup
func TestUnmount_CleanupRegistered(t *testing.T) {
	harness := NewHarness(t)
	component := createTestComponent("TestComponent")

	ct := harness.Mount(component)
	require.NotNil(t, ct)

	unmountCalled := false
	ct.onUnmount = func() {
		unmountCalled = true
	}

	// Call unmount
	ct.Unmount()

	// Unmount should have been called
	assert.True(t, unmountCalled, "unmount callback should be called")
}

// TestUnmount_Idempotent tests that calling Unmount multiple times is safe
func TestUnmount_Idempotent(t *testing.T) {
	harness := NewHarness(t)
	component := createTestComponent("TestComponent")

	ct := harness.Mount(component)
	require.NotNil(t, ct)

	unmountCount := 0
	ct.onUnmount = func() {
		unmountCount++
	}

	// Call unmount multiple times
	ct.Unmount()
	ct.Unmount()
	ct.Unmount()

	// Unmount should only be called once
	assert.Equal(t, 1, unmountCount, "unmount should only be called once")
}

// TestMount_WithProps tests mounting with props (future feature)
func TestMount_WithProps(t *testing.T) {
	harness := NewHarness(t)

	type TestProps struct {
		Value string
	}

	component, err := bubbly.NewComponent("TestComponent").
		Props(TestProps{Value: "test"}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(TestProps)
			return props.Value
		}).
		Build()

	require.NoError(t, err)

	ct := harness.Mount(component)

	require.NotNil(t, ct)
	// Props should be accessible through component
	props := ct.component.Props().(TestProps)
	assert.Equal(t, "test", props.Value, "props should be accessible")
}

// TestComponentTest_ComponentAccessible tests that component is accessible
func TestComponentTest_ComponentAccessible(t *testing.T) {
	harness := NewHarness(t)
	component := createTestComponent("TestComponent")

	ct := harness.Mount(component)

	require.NotNil(t, ct)
	assert.Equal(t, "TestComponent", ct.component.Name(), "component name should be accessible")
	assert.NotEmpty(t, ct.component.ID(), "component ID should be accessible")
}

// TestComponentTest_ViewAccessible tests that View() can be called
func TestComponentTest_ViewAccessible(t *testing.T) {
	harness := NewHarness(t)
	component := createTestComponent("TestComponent")

	ct := harness.Mount(component)

	require.NotNil(t, ct)
	view := ct.component.View()
	assert.Equal(t, "test component", view, "View() should return template output")
}
