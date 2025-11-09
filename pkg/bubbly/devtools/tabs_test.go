package devtools

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTabController_NewTabController tests creating a new tab controller.
func TestTabController_NewTabController(t *testing.T) {
	tests := []struct {
		name       string
		tabs       []TabItem
		wantLen    int
		wantActive int
	}{
		{
			name:       "empty tabs",
			tabs:       []TabItem{},
			wantLen:    0,
			wantActive: 0,
		},
		{
			name: "single tab",
			tabs: []TabItem{
				{Name: "Tab1", Content: func() string { return "Content 1" }},
			},
			wantLen:    1,
			wantActive: 0,
		},
		{
			name: "multiple tabs",
			tabs: []TabItem{
				{Name: "Tab1", Content: func() string { return "Content 1" }},
				{Name: "Tab2", Content: func() string { return "Content 2" }},
				{Name: "Tab3", Content: func() string { return "Content 3" }},
			},
			wantLen:    3,
			wantActive: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := NewTabController(tt.tabs)

			assert.NotNil(t, tc)
			assert.Equal(t, tt.wantLen, len(tc.tabs))
			assert.Equal(t, tt.wantActive, tc.activeTab)
		})
	}
}

// TestTabController_Next tests switching to next tab.
func TestTabController_Next(t *testing.T) {
	tests := []struct {
		name       string
		tabs       []TabItem
		startIndex int
		callCount  int
		wantIndex  int
	}{
		{
			name:       "empty tabs",
			tabs:       []TabItem{},
			startIndex: 0,
			callCount:  1,
			wantIndex:  0,
		},
		{
			name: "single tab - no change",
			tabs: []TabItem{
				{Name: "Tab1", Content: func() string { return "Content 1" }},
			},
			startIndex: 0,
			callCount:  1,
			wantIndex:  0,
		},
		{
			name: "three tabs - move forward",
			tabs: []TabItem{
				{Name: "Tab1", Content: func() string { return "Content 1" }},
				{Name: "Tab2", Content: func() string { return "Content 2" }},
				{Name: "Tab3", Content: func() string { return "Content 3" }},
			},
			startIndex: 0,
			callCount:  1,
			wantIndex:  1,
		},
		{
			name: "three tabs - wrap around",
			tabs: []TabItem{
				{Name: "Tab1", Content: func() string { return "Content 1" }},
				{Name: "Tab2", Content: func() string { return "Content 2" }},
				{Name: "Tab3", Content: func() string { return "Content 3" }},
			},
			startIndex: 2,
			callCount:  1,
			wantIndex:  0,
		},
		{
			name: "three tabs - multiple calls",
			tabs: []TabItem{
				{Name: "Tab1", Content: func() string { return "Content 1" }},
				{Name: "Tab2", Content: func() string { return "Content 2" }},
				{Name: "Tab3", Content: func() string { return "Content 3" }},
			},
			startIndex: 0,
			callCount:  4,
			wantIndex:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := NewTabController(tt.tabs)
			tc.activeTab = tt.startIndex

			for i := 0; i < tt.callCount; i++ {
				tc.Next()
			}

			assert.Equal(t, tt.wantIndex, tc.activeTab)
		})
	}
}

// TestTabController_Prev tests switching to previous tab.
func TestTabController_Prev(t *testing.T) {
	tests := []struct {
		name       string
		tabs       []TabItem
		startIndex int
		callCount  int
		wantIndex  int
	}{
		{
			name:       "empty tabs",
			tabs:       []TabItem{},
			startIndex: 0,
			callCount:  1,
			wantIndex:  0,
		},
		{
			name: "single tab - no change",
			tabs: []TabItem{
				{Name: "Tab1", Content: func() string { return "Content 1" }},
			},
			startIndex: 0,
			callCount:  1,
			wantIndex:  0,
		},
		{
			name: "three tabs - move backward",
			tabs: []TabItem{
				{Name: "Tab1", Content: func() string { return "Content 1" }},
				{Name: "Tab2", Content: func() string { return "Content 2" }},
				{Name: "Tab3", Content: func() string { return "Content 3" }},
			},
			startIndex: 2,
			callCount:  1,
			wantIndex:  1,
		},
		{
			name: "three tabs - wrap around",
			tabs: []TabItem{
				{Name: "Tab1", Content: func() string { return "Content 1" }},
				{Name: "Tab2", Content: func() string { return "Content 2" }},
				{Name: "Tab3", Content: func() string { return "Content 3" }},
			},
			startIndex: 0,
			callCount:  1,
			wantIndex:  2,
		},
		{
			name: "three tabs - multiple calls",
			tabs: []TabItem{
				{Name: "Tab1", Content: func() string { return "Content 1" }},
				{Name: "Tab2", Content: func() string { return "Content 2" }},
				{Name: "Tab3", Content: func() string { return "Content 3" }},
			},
			startIndex: 2,
			callCount:  4,
			wantIndex:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := NewTabController(tt.tabs)
			tc.activeTab = tt.startIndex

			for i := 0; i < tt.callCount; i++ {
				tc.Prev()
			}

			assert.Equal(t, tt.wantIndex, tc.activeTab)
		})
	}
}

// TestTabController_Select tests selecting a specific tab by index.
func TestTabController_Select(t *testing.T) {
	tabs := []TabItem{
		{Name: "Tab1", Content: func() string { return "Content 1" }},
		{Name: "Tab2", Content: func() string { return "Content 2" }},
		{Name: "Tab3", Content: func() string { return "Content 3" }},
	}

	tests := []struct {
		name        string
		selectIndex int
		wantIndex   int
	}{
		{
			name:        "select first tab",
			selectIndex: 0,
			wantIndex:   0,
		},
		{
			name:        "select middle tab",
			selectIndex: 1,
			wantIndex:   1,
		},
		{
			name:        "select last tab",
			selectIndex: 2,
			wantIndex:   2,
		},
		{
			name:        "select out of bounds (negative)",
			selectIndex: -1,
			wantIndex:   0, // Should not change
		},
		{
			name:        "select out of bounds (too high)",
			selectIndex: 10,
			wantIndex:   0, // Should not change
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := NewTabController(tabs)
			tc.Select(tt.selectIndex)

			assert.Equal(t, tt.wantIndex, tc.activeTab)
		})
	}
}

// TestTabController_GetActiveTab tests getting the active tab index.
func TestTabController_GetActiveTab(t *testing.T) {
	tabs := []TabItem{
		{Name: "Tab1", Content: func() string { return "Content 1" }},
		{Name: "Tab2", Content: func() string { return "Content 2" }},
		{Name: "Tab3", Content: func() string { return "Content 3" }},
	}

	tc := NewTabController(tabs)

	assert.Equal(t, 0, tc.GetActiveTab())

	tc.Next()
	assert.Equal(t, 1, tc.GetActiveTab())

	tc.Next()
	assert.Equal(t, 2, tc.GetActiveTab())
}

// TestTabController_Render tests rendering the tab controller.
func TestTabController_Render(t *testing.T) {
	tests := []struct {
		name         string
		tabs         []TabItem
		activeTab    int
		wantContains []string
	}{
		{
			name:         "empty tabs",
			tabs:         []TabItem{},
			activeTab:    0,
			wantContains: []string{"No tabs"},
		},
		{
			name: "single tab",
			tabs: []TabItem{
				{Name: "Tab1", Content: func() string { return "Content 1" }},
			},
			activeTab:    0,
			wantContains: []string{"Tab1", "Content 1"},
		},
		{
			name: "multiple tabs - first active",
			tabs: []TabItem{
				{Name: "Tab1", Content: func() string { return "Content 1" }},
				{Name: "Tab2", Content: func() string { return "Content 2" }},
				{Name: "Tab3", Content: func() string { return "Content 3" }},
			},
			activeTab:    0,
			wantContains: []string{"Tab1", "Tab2", "Tab3", "Content 1"},
		},
		{
			name: "multiple tabs - second active",
			tabs: []TabItem{
				{Name: "Tab1", Content: func() string { return "Content 1" }},
				{Name: "Tab2", Content: func() string { return "Content 2" }},
				{Name: "Tab3", Content: func() string { return "Content 3" }},
			},
			activeTab:    1,
			wantContains: []string{"Tab1", "Tab2", "Tab3", "Content 2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := NewTabController(tt.tabs)
			tc.activeTab = tt.activeTab

			output := tc.Render()

			for _, want := range tt.wantContains {
				assert.Contains(t, output, want, "Output should contain: %s", want)
			}
		})
	}
}

// TestTabController_Render_ActiveTabHighlighted tests that active tab is visually distinct.
func TestTabController_Render_ActiveTabHighlighted(t *testing.T) {
	tabs := []TabItem{
		{Name: "Tab1", Content: func() string { return "Content 1" }},
		{Name: "Tab2", Content: func() string { return "Content 2" }},
		{Name: "Tab3", Content: func() string { return "Content 3" }},
	}

	tc := NewTabController(tabs)

	// Render with first tab active
	output1 := tc.Render()
	assert.Contains(t, output1, "Tab1")
	assert.Contains(t, output1, "Content 1")

	// Switch to second tab
	tc.Next()
	output2 := tc.Render()
	assert.Contains(t, output2, "Tab2")
	assert.Contains(t, output2, "Content 2")
	assert.NotContains(t, output2, "Content 1")
}

// TestTabController_ThreadSafety tests concurrent access to tab controller.
func TestTabController_ThreadSafety(t *testing.T) {
	tabs := []TabItem{
		{Name: "Tab1", Content: func() string { return "Content 1" }},
		{Name: "Tab2", Content: func() string { return "Content 2" }},
		{Name: "Tab3", Content: func() string { return "Content 3" }},
	}

	tc := NewTabController(tabs)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent Next() calls
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			tc.Next()
		}
	}()

	// Concurrent Prev() calls
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			tc.Prev()
		}
	}()

	// Concurrent Select() calls
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			tc.Select(i % len(tabs))
		}
	}()

	// Concurrent GetActiveTab() calls
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = tc.GetActiveTab()
		}
	}()

	// Concurrent Render() calls
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = tc.Render()
		}
	}()

	wg.Wait()

	// Verify final state is valid
	activeTab := tc.GetActiveTab()
	assert.GreaterOrEqual(t, activeTab, 0)
	assert.Less(t, activeTab, len(tabs))
}

// TestTabController_MultipleTabGroups tests using multiple independent tab controllers.
func TestTabController_MultipleTabGroups(t *testing.T) {
	tabs1 := []TabItem{
		{Name: "Group1-Tab1", Content: func() string { return "G1 Content 1" }},
		{Name: "Group1-Tab2", Content: func() string { return "G1 Content 2" }},
	}

	tabs2 := []TabItem{
		{Name: "Group2-Tab1", Content: func() string { return "G2 Content 1" }},
		{Name: "Group2-Tab2", Content: func() string { return "G2 Content 2" }},
		{Name: "Group2-Tab3", Content: func() string { return "G2 Content 3" }},
	}

	tc1 := NewTabController(tabs1)
	tc2 := NewTabController(tabs2)

	// Navigate independently
	tc1.Next()
	tc2.Next()
	tc2.Next()

	// Verify independent state
	assert.Equal(t, 1, tc1.GetActiveTab())
	assert.Equal(t, 2, tc2.GetActiveTab())

	// Verify independent rendering
	output1 := tc1.Render()
	output2 := tc2.Render()

	assert.Contains(t, output1, "G1 Content 2")
	assert.Contains(t, output2, "G2 Content 3")
	assert.NotContains(t, output1, "G2")
	assert.NotContains(t, output2, "G1")
}
