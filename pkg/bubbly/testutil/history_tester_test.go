package testutil

import (
	"fmt"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
	"github.com/stretchr/testify/assert"
)

// TestNewHistoryTester tests the constructor.
func TestNewHistoryTester(t *testing.T) {
	r, err := router.NewRouterBuilder().
		Route("/home", "home").
		Build()
	assert.NoError(t, err)

	ht := NewHistoryTester(r)

	assert.NotNil(t, ht)
	assert.Equal(t, r, ht.router)
	assert.NotNil(t, ht.history)
	assert.Equal(t, 0, ht.currentIdx)
	assert.Equal(t, 0, ht.maxEntries)
}

// TestHistoryTester_AssertHistoryLength_Failure tests assertion failure path.
func TestHistoryTester_AssertHistoryLength_Failure(t *testing.T) {
	r, err := router.NewRouterBuilder().
		Route("/home", "home").
		Build()
	assert.NoError(t, err)

	ht := NewHistoryTester(r)

	// Navigate to create 1 entry
	target := &router.NavigationTarget{Path: "/home"}
	cmd := r.Push(target)
	if cmd != nil {
		_ = cmd()
	}

	// Create a mock testing.T to capture the error
	mockT := &mockHistoryTestingT{}

	// Assert wrong length (should fail)
	ht.AssertHistoryLength(mockT, 5)

	// Verify error was reported
	assert.True(t, mockT.errorCalled, "Expected Errorf to be called")
	assert.Contains(t, mockT.errorMsg, "expected history length 5, got 1")
}

// mockHistoryTestingT implements testingT for testing error paths in history_tester
type mockHistoryTestingT struct {
	errorCalled bool
	errorMsg    string
}

func (m *mockHistoryTestingT) Errorf(format string, args ...interface{}) {
	m.errorCalled = true
	m.errorMsg = fmt.Sprintf(format, args...)
}

func (m *mockHistoryTestingT) Helper() {
	// No-op for mock
}

func (m *mockHistoryTestingT) Logf(format string, args ...interface{}) {
	// No-op for mock
}

func (m *mockHistoryTestingT) Cleanup(func()) {
	// No-op for mock
}

// TestHistoryTester_AssertHistoryLength tests history length assertion.
func TestHistoryTester_AssertHistoryLength(t *testing.T) {
	tests := []struct {
		name           string
		navigations    []string
		expectedLength int
	}{
		{
			name:           "empty history",
			navigations:    []string{},
			expectedLength: 0,
		},
		{
			name:           "single navigation",
			navigations:    []string{"/home"},
			expectedLength: 1,
		},
		{
			name:           "multiple navigations",
			navigations:    []string{"/home", "/about", "/contact"},
			expectedLength: 3,
		},
		{
			name:           "navigation after back truncates forward",
			navigations:    []string{"/home", "/about", "/contact"},
			expectedLength: 3, // After back + new navigation = 3 entries total
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := router.NewRouterBuilder().
				Route("/home", "home").
				Route("/about", "about").
				Route("/contact", "contact").
				Route("/faq", "faq").
				Build()
			assert.NoError(t, err)

			ht := NewHistoryTester(r)

			// Perform navigations
			for _, path := range tt.navigations {
				target := &router.NavigationTarget{Path: path}
				cmd := r.Push(target)
				if cmd != nil {
					_ = cmd()
				}
			}

			// Special case: test truncation
			if tt.name == "navigation after back truncates forward" {
				// Back once
				cmd := r.Back()
				if cmd != nil {
					_ = cmd()
				}
				// Navigate to new path (truncates forward history)
				target := &router.NavigationTarget{Path: "/faq"}
				cmd = r.Push(target)
				if cmd != nil {
					_ = cmd()
				}
			}

			// Assert history length
			ht.AssertHistoryLength(t, tt.expectedLength)
		})
	}
}

// TestHistoryTester_AssertCanGoBack_Failure tests assertion failure path.
func TestHistoryTester_AssertCanGoBack_Failure(t *testing.T) {
	r, err := router.NewRouterBuilder().
		Route("/home", "home").
		Build()
	assert.NoError(t, err)

	ht := NewHistoryTester(r)

	// Navigate to create 1 entry (can't go back)
	target := &router.NavigationTarget{Path: "/home"}
	cmd := r.Push(target)
	if cmd != nil {
		_ = cmd()
	}

	// Create a mock testing.T to capture the error
	mockT := &mockHistoryTestingT{}

	// Assert can go back (should fail - we're at the start)
	ht.AssertCanGoBack(mockT, true)

	// Verify error was reported
	assert.True(t, mockT.errorCalled, "Expected Errorf to be called")
	assert.Contains(t, mockT.errorMsg, "expected canGoBack=true")
}

// TestHistoryTester_AssertCanGoForward_Failure tests assertion failure path.
func TestHistoryTester_AssertCanGoForward_Failure(t *testing.T) {
	r, err := router.NewRouterBuilder().
		Route("/home", "home").
		Build()
	assert.NoError(t, err)

	ht := NewHistoryTester(r)

	// Navigate to create 1 entry (can't go forward - at end)
	target := &router.NavigationTarget{Path: "/home"}
	cmd := r.Push(target)
	if cmd != nil {
		_ = cmd()
	}

	// Create a mock testing.T to capture the error
	mockT := &mockHistoryTestingT{}

	// Assert can go forward (should fail - we're at the end)
	ht.AssertCanGoForward(mockT, true)

	// Verify error was reported
	assert.True(t, mockT.errorCalled, "Expected Errorf to be called")
	assert.Contains(t, mockT.errorMsg, "expected canGoForward=true")
}

// TestHistoryTester_AssertCanGoBack tests back navigation capability assertion.
func TestHistoryTester_AssertCanGoBack(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*router.Router)
		expectedBack bool
	}{
		{
			name: "empty history cannot go back",
			setup: func(r *router.Router) {
				// No navigation
			},
			expectedBack: false,
		},
		{
			name: "single entry cannot go back",
			setup: func(r *router.Router) {
				target := &router.NavigationTarget{Path: "/home"}
				cmd := r.Push(target)
				if cmd != nil {
					_ = cmd()
				}
			},
			expectedBack: false,
		},
		{
			name: "two entries can go back",
			setup: func(r *router.Router) {
				target1 := &router.NavigationTarget{Path: "/home"}
				cmd := r.Push(target1)
				if cmd != nil {
					_ = cmd()
				}
				target2 := &router.NavigationTarget{Path: "/about"}
				cmd = r.Push(target2)
				if cmd != nil {
					_ = cmd()
				}
			},
			expectedBack: true,
		},
		{
			name: "after back at start cannot go back",
			setup: func(r *router.Router) {
				target1 := &router.NavigationTarget{Path: "/home"}
				cmd := r.Push(target1)
				if cmd != nil {
					_ = cmd()
				}
				target2 := &router.NavigationTarget{Path: "/about"}
				cmd = r.Push(target2)
				if cmd != nil {
					_ = cmd()
				}
				// Back to start
				cmd = r.Back()
				if cmd != nil {
					_ = cmd()
				}
			},
			expectedBack: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := router.NewRouterBuilder().
				Route("/home", "home").
				Route("/about", "about").
				Build()
			assert.NoError(t, err)

			ht := NewHistoryTester(r)
			tt.setup(r)

			ht.AssertCanGoBack(t, tt.expectedBack)
		})
	}
}

// TestHistoryTester_AssertCanGoForward tests forward navigation capability assertion.
func TestHistoryTester_AssertCanGoForward(t *testing.T) {
	tests := []struct {
		name            string
		setup           func(*router.Router)
		expectedForward bool
	}{
		{
			name: "empty history cannot go forward",
			setup: func(r *router.Router) {
				// No navigation
			},
			expectedForward: false,
		},
		{
			name: "at end cannot go forward",
			setup: func(r *router.Router) {
				target1 := &router.NavigationTarget{Path: "/home"}
				cmd := r.Push(target1)
				if cmd != nil {
					_ = cmd()
				}
				target2 := &router.NavigationTarget{Path: "/about"}
				cmd = r.Push(target2)
				if cmd != nil {
					_ = cmd()
				}
			},
			expectedForward: false,
		},
		{
			name: "after back can go forward",
			setup: func(r *router.Router) {
				target1 := &router.NavigationTarget{Path: "/home"}
				cmd := r.Push(target1)
				if cmd != nil {
					_ = cmd()
				}
				target2 := &router.NavigationTarget{Path: "/about"}
				cmd = r.Push(target2)
				if cmd != nil {
					_ = cmd()
				}
				// Back once
				cmd = r.Back()
				if cmd != nil {
					_ = cmd()
				}
			},
			expectedForward: true,
		},
		{
			name: "after back twice can go forward",
			setup: func(r *router.Router) {
				target1 := &router.NavigationTarget{Path: "/home"}
				cmd := r.Push(target1)
				if cmd != nil {
					_ = cmd()
				}
				target2 := &router.NavigationTarget{Path: "/about"}
				cmd = r.Push(target2)
				if cmd != nil {
					_ = cmd()
				}
				target3 := &router.NavigationTarget{Path: "/contact"}
				cmd = r.Push(target3)
				if cmd != nil {
					_ = cmd()
				}
				// Back twice
				cmd = r.Back()
				if cmd != nil {
					_ = cmd()
				}
				cmd = r.Back()
				if cmd != nil {
					_ = cmd()
				}
			},
			expectedForward: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := router.NewRouterBuilder().
				Route("/home", "home").
				Route("/about", "about").
				Route("/contact", "contact").
				Build()
			assert.NoError(t, err)

			ht := NewHistoryTester(r)
			tt.setup(r)

			ht.AssertCanGoForward(t, tt.expectedForward)
		})
	}
}

// TestHistoryTester_GetHistoryEntries tests retrieving history entries.
func TestHistoryTester_GetHistoryEntries(t *testing.T) {
	tests := []struct {
		name          string
		navigations   []string
		expectedPaths []string
	}{
		{
			name:          "empty history",
			navigations:   []string{},
			expectedPaths: []string{},
		},
		{
			name:          "single navigation",
			navigations:   []string{"/home"},
			expectedPaths: []string{"/home"},
		},
		{
			name:          "multiple navigations",
			navigations:   []string{"/home", "/about", "/contact"},
			expectedPaths: []string{"/home", "/about", "/contact"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := router.NewRouterBuilder().
				Route("/home", "home").
				Route("/about", "about").
				Route("/contact", "contact").
				Build()
			assert.NoError(t, err)

			ht := NewHistoryTester(r)

			// Perform navigations
			for _, path := range tt.navigations {
				target := &router.NavigationTarget{Path: path}
				cmd := r.Push(target)
				if cmd != nil {
					_ = cmd()
				}
			}

			// Get history entries
			entries := ht.GetHistoryEntries()

			// Verify count
			assert.Equal(t, len(tt.expectedPaths), len(entries))

			// Verify paths
			for i, expectedPath := range tt.expectedPaths {
				assert.Equal(t, expectedPath, entries[i].Route.Path)
			}
		})
	}
}

// TestHistoryTester_ReplaceNavigation tests replace navigation doesn't add entry.
func TestHistoryTester_ReplaceNavigation(t *testing.T) {
	r, err := router.NewRouterBuilder().
		Route("/home", "home").
		Route("/about", "about").
		Route("/team", "team").
		Build()
	assert.NoError(t, err)

	ht := NewHistoryTester(r)

	// Navigate to /home
	target1 := &router.NavigationTarget{Path: "/home"}
	cmd := r.Push(target1)
	if cmd != nil {
		_ = cmd()
	}

	// Navigate to /about
	target2 := &router.NavigationTarget{Path: "/about"}
	cmd = r.Push(target2)
	if cmd != nil {
		_ = cmd()
	}

	// History should have 2 entries
	ht.AssertHistoryLength(t, 2)

	// Replace current with /team (uses Replace method)
	target3 := &router.NavigationTarget{Path: "/team"}
	cmd = r.Replace(target3)
	if cmd != nil {
		_ = cmd()
	}

	// History should still have 2 entries (replace doesn't add)
	ht.AssertHistoryLength(t, 2)

	// Current route should be /team
	assert.Equal(t, "/team", r.CurrentRoute().Path)
}

// TestHistoryTester_HistoryLimit tests history limit enforcement.
func TestHistoryTester_HistoryLimit(t *testing.T) {
	// Note: History limit is configured through router options
	// For now, test basic history accumulation
	r, err := router.NewRouterBuilder().
		Route("/page1", "page1").
		Route("/page2", "page2").
		Route("/page3", "page3").
		Route("/page4", "page4").
		Route("/page5", "page5").
		Build()
	assert.NoError(t, err)

	ht := NewHistoryTester(r)

	// Navigate to 5 pages
	paths := []string{"/page1", "/page2", "/page3", "/page4", "/page5"}
	for _, path := range paths {
		target := &router.NavigationTarget{Path: path}
		cmd := r.Push(target)
		if cmd != nil {
			_ = cmd()
		}
	}

	// History should have all 5 entries (no limit set)
	ht.AssertHistoryLength(t, 5)

	// Should contain all pages in order
	entries := ht.GetHistoryEntries()
	for i, path := range paths {
		assert.Equal(t, path, entries[i].Route.Path)
	}
}

// TestHistoryTester_BackForwardFlow tests back and forward navigation flow.
func TestHistoryTester_BackForwardFlow(t *testing.T) {
	r, err := router.NewRouterBuilder().
		Route("/home", "home").
		Route("/about", "about").
		Route("/contact", "contact").
		Build()
	assert.NoError(t, err)

	ht := NewHistoryTester(r)

	// Navigate forward: /home -> /about -> /contact
	paths := []string{"/home", "/about", "/contact"}
	for _, path := range paths {
		target := &router.NavigationTarget{Path: path}
		cmd := r.Push(target)
		if cmd != nil {
			_ = cmd()
		}
	}

	// At /contact, can go back but not forward
	ht.AssertCanGoBack(t, true)
	ht.AssertCanGoForward(t, false)
	assert.Equal(t, "/contact", r.CurrentRoute().Path)

	// Go back to /about
	cmd := r.Back()
	if cmd != nil {
		_ = cmd()
	}

	// At /about, can go both back and forward
	ht.AssertCanGoBack(t, true)
	ht.AssertCanGoForward(t, true)
	assert.Equal(t, "/about", r.CurrentRoute().Path)

	// Go back to /home
	cmd = r.Back()
	if cmd != nil {
		_ = cmd()
	}

	// At /home, can only go forward
	ht.AssertCanGoBack(t, false)
	ht.AssertCanGoForward(t, true)
	assert.Equal(t, "/home", r.CurrentRoute().Path)

	// Go forward to /about
	cmd = r.Forward()
	if cmd != nil {
		_ = cmd()
	}

	// Back at /about
	assert.Equal(t, "/about", r.CurrentRoute().Path)
	ht.AssertCanGoBack(t, true)
	ht.AssertCanGoForward(t, true)
}
