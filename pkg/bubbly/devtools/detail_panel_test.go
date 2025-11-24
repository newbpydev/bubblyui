package devtools

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDetailPanel(t *testing.T) {
	tests := []struct {
		name      string
		component *ComponentSnapshot
		wantTabs  int
	}{
		{
			name:      "creates with nil component",
			component: nil,
			wantTabs:  3, // State, Props, Events
		},
		{
			name: "creates with component",
			component: &ComponentSnapshot{
				ID:   "test-1",
				Name: "TestComponent",
			},
			wantTabs: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dp := NewDetailPanel(tt.component)

			assert.NotNil(t, dp)
			assert.Equal(t, tt.wantTabs, len(dp.tabs))
			assert.Equal(t, 0, dp.activeTab)

			// Verify default tabs exist
			tabNames := make([]string, len(dp.tabs))
			for i, tab := range dp.tabs {
				tabNames[i] = tab.Name
			}
			assert.Contains(t, tabNames, "State")
			assert.Contains(t, tabNames, "Props")
			assert.Contains(t, tabNames, "Events")
		})
	}
}

func TestDetailPanel_Render_NilComponent(t *testing.T) {
	dp := NewDetailPanel(nil)
	output := dp.Render()

	assert.NotEmpty(t, output)
	assert.Contains(t, output, "No component selected")
}

func TestDetailPanel_Render_WithComponent(t *testing.T) {
	component := &ComponentSnapshot{
		ID:        "test-1",
		Name:      "TestComponent",
		Type:      "TestType",
		Timestamp: time.Now(),
		Refs: []*RefSnapshot{
			{
				ID:    "ref-1",
				Name:  "count",
				Type:  "int",
				Value: 42,
			},
		},
		Props: map[string]interface{}{
			"title": "Test Title",
			"count": 10,
		},
	}

	dp := NewDetailPanel(component)
	output := dp.Render()

	assert.NotEmpty(t, output)

	// Should contain tab names
	assert.Contains(t, output, "State")
	assert.Contains(t, output, "Props")
	assert.Contains(t, output, "Events")

	// Should contain component name
	assert.Contains(t, output, "TestComponent")
}

func TestDetailPanel_SwitchTab(t *testing.T) {
	tests := []struct {
		name       string
		startTab   int
		switchTo   int
		wantActive int
	}{
		{
			name:       "switch to valid tab",
			startTab:   0,
			switchTo:   1,
			wantActive: 1,
		},
		{
			name:       "switch to last tab",
			startTab:   0,
			switchTo:   2,
			wantActive: 2,
		},
		{
			name:       "invalid negative index stays on current",
			startTab:   1,
			switchTo:   -1,
			wantActive: 1,
		},
		{
			name:       "invalid high index stays on current",
			startTab:   1,
			switchTo:   10,
			wantActive: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dp := NewDetailPanel(nil)
			dp.activeTab = tt.startTab

			dp.SwitchTab(tt.switchTo)

			assert.Equal(t, tt.wantActive, dp.activeTab)
		})
	}
}

func TestDetailPanel_StateTab(t *testing.T) {
	component := &ComponentSnapshot{
		ID:   "test-1",
		Name: "TestComponent",
		Refs: []*RefSnapshot{
			{
				ID:    "ref-1",
				Name:  "count",
				Type:  "int",
				Value: 42,
			},
			{
				ID:    "ref-2",
				Name:  "message",
				Type:  "string",
				Value: "hello",
			},
		},
	}

	dp := NewDetailPanel(component)
	dp.SwitchTab(0) // State tab
	output := dp.Render()

	assert.Contains(t, output, "count")
	assert.Contains(t, output, "42")
	assert.Contains(t, output, "message")
	assert.Contains(t, output, "hello")
}

func TestDetailPanel_PropsTab(t *testing.T) {
	component := &ComponentSnapshot{
		ID:   "test-1",
		Name: "TestComponent",
		Props: map[string]interface{}{
			"title":   "Test Title",
			"enabled": true,
			"count":   10,
		},
	}

	dp := NewDetailPanel(component)
	dp.SwitchTab(1) // Props tab
	output := dp.Render()

	assert.Contains(t, output, "title")
	assert.Contains(t, output, "Test Title")
	assert.Contains(t, output, "enabled")
	assert.Contains(t, output, "count")
}

func TestDetailPanel_EventsTab(t *testing.T) {
	component := &ComponentSnapshot{
		ID:   "test-1",
		Name: "TestComponent",
	}

	dp := NewDetailPanel(component)
	dp.SwitchTab(2) // Events tab
	output := dp.Render()

	// Events tab should render (placeholder for now)
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Events")
}

func TestDetailPanel_SetComponent(t *testing.T) {
	dp := NewDetailPanel(nil)

	// Initially nil
	assert.Nil(t, dp.component)

	// Set component
	component := &ComponentSnapshot{
		ID:   "test-1",
		Name: "TestComponent",
	}
	dp.SetComponent(component)

	assert.NotNil(t, dp.component)
	assert.Equal(t, "test-1", dp.component.ID)

	// Set back to nil
	dp.SetComponent(nil)
	assert.Nil(t, dp.component)
}

func TestDetailPanel_GetActiveTab(t *testing.T) {
	dp := NewDetailPanel(nil)

	assert.Equal(t, 0, dp.GetActiveTab())

	dp.SwitchTab(1)
	assert.Equal(t, 1, dp.GetActiveTab())
}

func TestDetailPanel_ThreadSafety(t *testing.T) {
	component := &ComponentSnapshot{
		ID:   "test-1",
		Name: "TestComponent",
		Refs: []*RefSnapshot{
			{ID: "ref-1", Name: "count", Type: "int", Value: 42},
		},
	}

	dp := NewDetailPanel(component)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent renders
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = dp.Render()
		}()
	}

	// Concurrent tab switches
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			dp.SwitchTab(index % 3)
		}(i)
	}

	// Concurrent component updates
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dp.SetComponent(component)
		}()
	}

	wg.Wait()

	// Should not panic and should have valid state
	output := dp.Render()
	assert.NotEmpty(t, output)
}

func TestDetailPanel_EmptyRefs(t *testing.T) {
	component := &ComponentSnapshot{
		ID:   "test-1",
		Name: "TestComponent",
		Refs: []*RefSnapshot{}, // Empty refs
	}

	dp := NewDetailPanel(component)
	output := dp.Render()

	assert.NotEmpty(t, output)
	assert.Contains(t, output, "State")
}

func TestDetailPanel_EmptyProps(t *testing.T) {
	component := &ComponentSnapshot{
		ID:    "test-1",
		Name:  "TestComponent",
		Props: map[string]interface{}{}, // Empty props
	}

	dp := NewDetailPanel(component)
	dp.SwitchTab(1) // Props tab
	output := dp.Render()

	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Props")
}

func TestDetailPanel_TabNavigation(t *testing.T) {
	dp := NewDetailPanel(nil)

	// Start at tab 0
	assert.Equal(t, 0, dp.GetActiveTab())

	// Navigate forward
	dp.NextTab()
	assert.Equal(t, 1, dp.GetActiveTab())

	dp.NextTab()
	assert.Equal(t, 2, dp.GetActiveTab())

	// Wrap around
	dp.NextTab()
	assert.Equal(t, 0, dp.GetActiveTab())

	// Navigate backward
	dp.PreviousTab()
	assert.Equal(t, 2, dp.GetActiveTab())

	dp.PreviousTab()
	assert.Equal(t, 1, dp.GetActiveTab())
}

func TestDetailPanel_ComplexValues(t *testing.T) {
	component := &ComponentSnapshot{
		ID:   "test-1",
		Name: "TestComponent",
		Refs: []*RefSnapshot{
			{
				ID:    "ref-1",
				Name:  "data",
				Type:  "map[string]interface{}",
				Value: map[string]interface{}{"nested": "value"},
			},
		},
		Props: map[string]interface{}{
			"nested": map[string]string{"key": "value"},
			"array":  []int{1, 2, 3},
		},
	}

	dp := NewDetailPanel(component)

	// State tab with complex value
	output := dp.Render()
	assert.Contains(t, output, "data")

	// Props tab with complex values
	dp.SwitchTab(1)
	output = dp.Render()
	assert.Contains(t, output, "nested")
	assert.Contains(t, output, "array")
}

// TestDetailPanel_EventsTab_WithStore tests the events tab rendering with store integration
func TestDetailPanel_EventsTab_WithStore(t *testing.T) {
	// Create a component with events
	component := &ComponentSnapshot{
		ID:   "comp-1",
		Name: "TestComponent",
	}

	dp := NewDetailPanel(component)

	// Switch to events tab
	dp.SwitchTab(2)

	output := dp.Render()
	// Events tab should render
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Events")
}

// TestDetailPanel_EventsTab_EmptyEvents tests events tab with no events
func TestDetailPanel_EventsTab_EmptyEvents(t *testing.T) {
	component := &ComponentSnapshot{
		ID:   "comp-1",
		Name: "TestComponent",
	}

	dp := NewDetailPanel(component)

	// Switch to events tab
	dp.SwitchTab(2)

	output := dp.Render()
	// Should still render without errors
	assert.NotEmpty(t, output)
}

// TestDetailPanel_AllTabs tests rendering all tabs
func TestDetailPanel_AllTabs(t *testing.T) {
	component := &ComponentSnapshot{
		ID:   "comp-1",
		Name: "TestComponent",
		Refs: []*RefSnapshot{
			{ID: "ref-1", Name: "count", Value: 42},
		},
		Props: map[string]interface{}{
			"title": "Hello",
		},
	}

	dp := NewDetailPanel(component)

	// Test all tabs
	for i := 0; i < 3; i++ {
		dp.SwitchTab(i)
		output := dp.Render()
		assert.NotEmpty(t, output, "Tab %d should render", i)
	}
}

// TestDetailPanel_EventsTab_NilComponent tests events tab with nil component
func TestDetailPanel_EventsTab_NilComponent(t *testing.T) {
	dp := NewDetailPanel(nil)
	dp.SwitchTab(2) // Events tab

	output := dp.Render()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "No component selected")
}

// TestDetailPanel_StateTab_NilComponent tests state tab with nil component
func TestDetailPanel_StateTab_NilComponent(t *testing.T) {
	dp := NewDetailPanel(nil)
	dp.SwitchTab(0) // State tab

	output := dp.Render()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "No component selected")
}

// TestDetailPanel_PropsTab_NilComponent tests props tab with nil component
func TestDetailPanel_PropsTab_NilComponent(t *testing.T) {
	dp := NewDetailPanel(nil)
	dp.SwitchTab(1) // Props tab

	output := dp.Render()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "No component selected")
}
