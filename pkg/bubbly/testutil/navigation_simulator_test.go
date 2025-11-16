package testutil

import (
	"fmt"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
	"github.com/stretchr/testify/assert"
)

// TestNewNavigationSimulator tests the constructor.
func TestNewNavigationSimulator(t *testing.T) {
	r := router.NewRouter()
	ns := NewNavigationSimulator(r)

	assert.NotNil(t, ns)
	assert.Equal(t, r, ns.router)
	assert.Empty(t, ns.history)
	assert.Equal(t, -1, ns.currentIdx)
}

// TestNavigationSimulator_Navigate tests basic navigation.
func TestNavigationSimulator_Navigate(t *testing.T) {
	tests := []struct {
		name            string
		routes          []string
		navigatePaths   []string
		expectedHistory []string
		expectedIdx     int
		expectedCurrent string
	}{
		{
			name:            "single navigation",
			routes:          []string{"/home"},
			navigatePaths:   []string{"/home"},
			expectedHistory: []string{"/home"},
			expectedIdx:     0,
			expectedCurrent: "/home",
		},
		{
			name:            "multiple navigations",
			routes:          []string{"/home", "/about", "/contact"},
			navigatePaths:   []string{"/home", "/about", "/contact"},
			expectedHistory: []string{"/home", "/about", "/contact"},
			expectedIdx:     2,
			expectedCurrent: "/contact",
		},
		{
			name:            "navigate after back truncates forward history",
			routes:          []string{"/home", "/about", "/contact", "/faq"},
			navigatePaths:   []string{"/home", "/about", "/contact"},
			expectedHistory: []string{"/home", "/about", "/contact"},
			expectedIdx:     2,
			expectedCurrent: "/contact",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build router with routes
			builder := router.NewRouterBuilder()
			for i, path := range tt.routes {
				builder.Route(path, fmt.Sprintf("route%d", i))
			}
			r, err := builder.Build()
			assert.NoError(t, err)

			ns := NewNavigationSimulator(r)

			// Perform navigations
			for _, path := range tt.navigatePaths {
				ns.Navigate(path)
			}

			// Verify history
			assert.Equal(t, tt.expectedHistory, ns.history)
			assert.Equal(t, tt.expectedIdx, ns.currentIdx)

			// Verify current route
			current := r.CurrentRoute()
			assert.NotNil(t, current)
			assert.Equal(t, tt.expectedCurrent, current.Path)
		})
	}
}

// TestNavigationSimulator_Back tests back navigation.
func TestNavigationSimulator_Back(t *testing.T) {
	tests := []struct {
		name            string
		setupNav        func(*NavigationSimulator)
		backCount       int
		expectedIdx     int
		expectedCurrent string
		shouldChange    bool
	}{
		{
			name: "back from second page",
			setupNav: func(ns *NavigationSimulator) {
				ns.Navigate("/home")
				ns.Navigate("/about")
			},
			backCount:       1,
			expectedIdx:     0,
			expectedCurrent: "/home",
			shouldChange:    true,
		},
		{
			name: "back multiple times",
			setupNav: func(ns *NavigationSimulator) {
				ns.Navigate("/home")
				ns.Navigate("/about")
				ns.Navigate("/contact")
			},
			backCount:       2,
			expectedIdx:     0,
			expectedCurrent: "/home",
			shouldChange:    true,
		},
		{
			name: "back at start does nothing",
			setupNav: func(ns *NavigationSimulator) {
				ns.Navigate("/home")
			},
			backCount:       1,
			expectedIdx:     0,
			expectedCurrent: "/home",
			shouldChange:    false,
		},
		{
			name: "back beyond start stays at start",
			setupNav: func(ns *NavigationSimulator) {
				ns.Navigate("/home")
				ns.Navigate("/about")
			},
			backCount:       5,
			expectedIdx:     0,
			expectedCurrent: "/home",
			shouldChange:    true,
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

			ns := NewNavigationSimulator(r)
			tt.setupNav(ns)

			// Perform back navigation
			for i := 0; i < tt.backCount; i++ {
				ns.Back()
			}

			// Verify index
			assert.Equal(t, tt.expectedIdx, ns.currentIdx)

			// Verify current route
			current := r.CurrentRoute()
			assert.NotNil(t, current)
			assert.Equal(t, tt.expectedCurrent, current.Path)
		})
	}
}

// TestNavigationSimulator_Forward tests forward navigation.
func TestNavigationSimulator_Forward(t *testing.T) {
	tests := []struct {
		name            string
		setupNav        func(*NavigationSimulator)
		forwardCount    int
		expectedIdx     int
		expectedCurrent string
		shouldChange    bool
	}{
		{
			name: "forward after back",
			setupNav: func(ns *NavigationSimulator) {
				ns.Navigate("/home")
				ns.Navigate("/about")
				ns.Back()
			},
			forwardCount:    1,
			expectedIdx:     1,
			expectedCurrent: "/about",
			shouldChange:    true,
		},
		{
			name: "forward multiple times",
			setupNav: func(ns *NavigationSimulator) {
				ns.Navigate("/home")
				ns.Navigate("/about")
				ns.Navigate("/contact")
				ns.Back()
				ns.Back()
			},
			forwardCount:    2,
			expectedIdx:     2,
			expectedCurrent: "/contact",
			shouldChange:    true,
		},
		{
			name: "forward at end does nothing",
			setupNav: func(ns *NavigationSimulator) {
				ns.Navigate("/home")
				ns.Navigate("/about")
			},
			forwardCount:    1,
			expectedIdx:     1,
			expectedCurrent: "/about",
			shouldChange:    false,
		},
		{
			name: "forward beyond end stays at end",
			setupNav: func(ns *NavigationSimulator) {
				ns.Navigate("/home")
				ns.Navigate("/about")
				ns.Back()
			},
			forwardCount:    5,
			expectedIdx:     1,
			expectedCurrent: "/about",
			shouldChange:    true,
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

			ns := NewNavigationSimulator(r)
			tt.setupNav(ns)

			// Perform forward navigation
			for i := 0; i < tt.forwardCount; i++ {
				ns.Forward()
			}

			// Verify index
			assert.Equal(t, tt.expectedIdx, ns.currentIdx)

			// Verify current route
			current := r.CurrentRoute()
			assert.NotNil(t, current)
			assert.Equal(t, tt.expectedCurrent, current.Path)
		})
	}
}

// TestNavigationSimulator_BackForwardIntegration tests back/forward flow.
func TestNavigationSimulator_BackForwardIntegration(t *testing.T) {
	r, err := router.NewRouterBuilder().
		Route("/home", "home").
		Route("/about", "about").
		Route("/contact", "contact").
		Route("/faq", "faq").
		Build()
	assert.NoError(t, err)

	ns := NewNavigationSimulator(r)

	// Build history: /home -> /about -> /contact
	ns.Navigate("/home")
	ns.Navigate("/about")
	ns.Navigate("/contact")

	assert.Equal(t, []string{"/home", "/about", "/contact"}, ns.history)
	assert.Equal(t, 2, ns.currentIdx)
	assert.Equal(t, "/contact", r.CurrentRoute().Path)

	// Go back twice: /contact -> /about -> /home
	ns.Back()
	assert.Equal(t, 1, ns.currentIdx)
	assert.Equal(t, "/about", r.CurrentRoute().Path)

	ns.Back()
	assert.Equal(t, 0, ns.currentIdx)
	assert.Equal(t, "/home", r.CurrentRoute().Path)

	// Go forward once: /home -> /about
	ns.Forward()
	assert.Equal(t, 1, ns.currentIdx)
	assert.Equal(t, "/about", r.CurrentRoute().Path)

	// Navigate to new page (should truncate forward history)
	ns.Navigate("/faq")
	assert.Equal(t, []string{"/home", "/about", "/faq"}, ns.history)
	assert.Equal(t, 2, ns.currentIdx)
	assert.Equal(t, "/faq", r.CurrentRoute().Path)

	// Can't go forward anymore
	ns.Forward()
	assert.Equal(t, 2, ns.currentIdx)
	assert.Equal(t, "/faq", r.CurrentRoute().Path)
}

// TestNavigationSimulator_HistoryTracking tests history is tracked correctly.
func TestNavigationSimulator_HistoryTracking(t *testing.T) {
	r, err := router.NewRouterBuilder().
		Route("/page1", "page1").
		Route("/page2", "page2").
		Route("/page3", "page3").
		Route("/page4", "page4").
		Route("/page5", "page5").
		Build()
	assert.NoError(t, err)

	ns := NewNavigationSimulator(r)

	// Navigate through multiple pages
	paths := []string{"/page1", "/page2", "/page3", "/page4", "/page5"}
	for _, path := range paths {
		ns.Navigate(path)
	}

	// Verify complete history
	assert.Equal(t, paths, ns.history)
	assert.Equal(t, 4, ns.currentIdx)
	assert.Equal(t, "/page5", r.CurrentRoute().Path)

	// Go back to middle
	ns.Back()
	ns.Back()
	assert.Equal(t, 2, ns.currentIdx)
	assert.Equal(t, "/page3", r.CurrentRoute().Path)

	// History should still be complete
	assert.Equal(t, paths, ns.history)
}

// TestNavigationSimulator_AssertionHelpers tests assertion helper methods.
func TestNavigationSimulator_AssertionHelpers(t *testing.T) {
	r, err := router.NewRouterBuilder().
		Route("/home", "home").
		Route("/about", "about").
		Build()
	assert.NoError(t, err)

	ns := NewNavigationSimulator(r)
	ns.Navigate("/home")
	ns.Navigate("/about")

	// Test AssertCurrentPath
	ns.AssertCurrentPath(t, "/about")

	// Test AssertHistoryLength
	ns.AssertHistoryLength(t, 2)

	// Test AssertCanGoBack
	ns.AssertCanGoBack(t, true)

	// Test AssertCanGoForward
	ns.AssertCanGoForward(t, false)

	// Go back and test again
	ns.Back()
	ns.AssertCurrentPath(t, "/home")
	ns.AssertCanGoBack(t, false)
	ns.AssertCanGoForward(t, true)
}
