package testutil

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
)

// TestNewMockRouter tests creating a new mock router.
func TestNewMockRouter(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates empty mock router"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMockRouter()

			assert.NotNil(t, mr)
			assert.Nil(t, mr.CurrentRoute())
			assert.Equal(t, 0, mr.GetPushCallCount())
			assert.Equal(t, 0, mr.GetReplaceCallCount())
			assert.Equal(t, 0, mr.GetBackCallCount())
			assert.Empty(t, mr.GetPushCalls())
			assert.Empty(t, mr.GetReplaceCalls())
		})
	}
}

// TestMockRouter_SetCurrentRoute tests setting the current route.
func TestMockRouter_SetCurrentRoute(t *testing.T) {
	tests := []struct {
		name  string
		route *router.Route
	}{
		{
			name:  "set home route",
			route: router.NewRoute("/home", "home", nil, nil, "", nil, nil),
		},
		{
			name:  "set about route",
			route: router.NewRoute("/about", "about", nil, nil, "", nil, nil),
		},
		{
			name:  "set nil route",
			route: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMockRouter()
			mr.SetCurrentRoute(tt.route)

			assert.Equal(t, tt.route, mr.CurrentRoute())
		})
	}
}

// TestMockRouter_Push tests Push navigation tracking.
func TestMockRouter_Push(t *testing.T) {
	tests := []struct {
		name    string
		targets []*router.NavigationTarget
	}{
		{
			name: "single push",
			targets: []*router.NavigationTarget{
				{Path: "/about"},
			},
		},
		{
			name: "multiple pushes",
			targets: []*router.NavigationTarget{
				{Path: "/about"},
				{Path: "/contact"},
				{Path: "/services"},
			},
		},
		{
			name: "push with params",
			targets: []*router.NavigationTarget{
				{
					Path:   "/user/123",
					Params: map[string]string{"id": "123"},
				},
			},
		},
		{
			name: "push with query",
			targets: []*router.NavigationTarget{
				{
					Path:  "/search",
					Query: map[string]string{"q": "test"},
				},
			},
		},
		{
			name: "push by name",
			targets: []*router.NavigationTarget{
				{
					Name:   "user-detail",
					Params: map[string]string{"id": "456"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMockRouter()

			// Push all targets
			for _, target := range tt.targets {
				cmd := mr.Push(target)
				assert.NotNil(t, cmd)
				// Execute command (should return nil)
				msg := cmd()
				assert.Nil(t, msg)
			}

			// Verify count
			assert.Equal(t, len(tt.targets), mr.GetPushCallCount())

			// Verify all targets recorded
			calls := mr.GetPushCalls()
			assert.Equal(t, len(tt.targets), len(calls))
			for i, target := range tt.targets {
				assert.Equal(t, target.Path, calls[i].Path)
				assert.Equal(t, target.Name, calls[i].Name)
				assert.Equal(t, target.Params, calls[i].Params)
				assert.Equal(t, target.Query, calls[i].Query)
			}
		})
	}
}

// TestMockRouter_Replace tests Replace navigation tracking.
func TestMockRouter_Replace(t *testing.T) {
	tests := []struct {
		name    string
		targets []*router.NavigationTarget
	}{
		{
			name: "single replace",
			targets: []*router.NavigationTarget{
				{Path: "/login"},
			},
		},
		{
			name: "multiple replaces",
			targets: []*router.NavigationTarget{
				{Path: "/login"},
				{Path: "/dashboard"},
			},
		},
		{
			name: "replace with params",
			targets: []*router.NavigationTarget{
				{
					Path:   "/user/789",
					Params: map[string]string{"id": "789"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMockRouter()

			// Replace all targets
			for _, target := range tt.targets {
				cmd := mr.Replace(target)
				assert.NotNil(t, cmd)
				// Execute command (should return nil)
				msg := cmd()
				assert.Nil(t, msg)
			}

			// Verify count
			assert.Equal(t, len(tt.targets), mr.GetReplaceCallCount())

			// Verify all targets recorded
			calls := mr.GetReplaceCalls()
			assert.Equal(t, len(tt.targets), len(calls))
			for i, target := range tt.targets {
				assert.Equal(t, target.Path, calls[i].Path)
			}
		})
	}
}

// TestMockRouter_Back tests Back navigation tracking.
func TestMockRouter_Back(t *testing.T) {
	tests := []struct {
		name      string
		backCalls int
	}{
		{name: "single back", backCalls: 1},
		{name: "multiple backs", backCalls: 5},
		{name: "many backs", backCalls: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMockRouter()

			// Call Back multiple times
			for i := 0; i < tt.backCalls; i++ {
				cmd := mr.Back()
				assert.NotNil(t, cmd)
				// Execute command (should return nil)
				msg := cmd()
				assert.Nil(t, msg)
			}

			// Verify count
			assert.Equal(t, tt.backCalls, mr.GetBackCallCount())
		})
	}
}

// TestMockRouter_Reset tests resetting the mock router.
func TestMockRouter_Reset(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*MockRouter)
	}{
		{
			name: "reset after push",
			setup: func(mr *MockRouter) {
				mr.Push(&router.NavigationTarget{Path: "/about"})
				mr.Push(&router.NavigationTarget{Path: "/contact"})
			},
		},
		{
			name: "reset after replace",
			setup: func(mr *MockRouter) {
				mr.Replace(&router.NavigationTarget{Path: "/login"})
			},
		},
		{
			name: "reset after back",
			setup: func(mr *MockRouter) {
				mr.Back()
				mr.Back()
				mr.Back()
			},
		},
		{
			name: "reset after mixed calls",
			setup: func(mr *MockRouter) {
				mr.Push(&router.NavigationTarget{Path: "/about"})
				mr.Replace(&router.NavigationTarget{Path: "/login"})
				mr.Back()
				mr.SetCurrentRoute(router.NewRoute("/home", "home", nil, nil, "", nil, nil))
			},
		},
		{
			name: "reset empty router",
			setup: func(mr *MockRouter) {
				// No setup
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMockRouter()
			tt.setup(mr)

			// Reset
			mr.Reset()

			// Verify all cleared
			assert.Nil(t, mr.CurrentRoute())
			assert.Equal(t, 0, mr.GetPushCallCount())
			assert.Equal(t, 0, mr.GetReplaceCallCount())
			assert.Equal(t, 0, mr.GetBackCallCount())
			assert.Empty(t, mr.GetPushCalls())
			assert.Empty(t, mr.GetReplaceCalls())
		})
	}
}

// TestMockRouter_AssertPushed tests the AssertPushed assertion helper.
func TestMockRouter_AssertPushed(t *testing.T) {
	tests := []struct {
		name       string
		pushPaths  []string
		assertPath string
		shouldPass bool
	}{
		{
			name:       "assert existing path",
			pushPaths:  []string{"/about", "/contact"},
			assertPath: "/about",
			shouldPass: true,
		},
		{
			name:       "assert non-existing path",
			pushPaths:  []string{"/about"},
			assertPath: "/contact",
			shouldPass: false,
		},
		{
			name:       "assert in multiple paths",
			pushPaths:  []string{"/home", "/about", "/contact", "/services"},
			assertPath: "/contact",
			shouldPass: true,
		},
		{
			name:       "assert on empty",
			pushPaths:  []string{},
			assertPath: "/about",
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMockRouter()

			// Push paths
			for _, path := range tt.pushPaths {
				mr.Push(&router.NavigationTarget{Path: path})
			}

			// Create mock testing.T
			mockT := &mockTestingT{}

			// Assert
			mr.AssertPushed(mockT, tt.assertPath)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "Expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "Expected assertion to fail")
			}
		})
	}
}

// TestMockRouter_AssertReplaced tests the AssertReplaced assertion helper.
func TestMockRouter_AssertReplaced(t *testing.T) {
	tests := []struct {
		name         string
		replacePaths []string
		assertPath   string
		shouldPass   bool
	}{
		{
			name:         "assert existing path",
			replacePaths: []string{"/login"},
			assertPath:   "/login",
			shouldPass:   true,
		},
		{
			name:         "assert non-existing path",
			replacePaths: []string{"/login"},
			assertPath:   "/dashboard",
			shouldPass:   false,
		},
		{
			name:         "assert in multiple paths",
			replacePaths: []string{"/login", "/dashboard", "/profile"},
			assertPath:   "/dashboard",
			shouldPass:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMockRouter()

			// Replace paths
			for _, path := range tt.replacePaths {
				mr.Replace(&router.NavigationTarget{Path: path})
			}

			// Create mock testing.T
			mockT := &mockTestingT{}

			// Assert
			mr.AssertReplaced(mockT, tt.assertPath)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "Expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "Expected assertion to fail")
			}
		})
	}
}

// TestMockRouter_AssertBackCalled tests the AssertBackCalled assertion helper.
func TestMockRouter_AssertBackCalled(t *testing.T) {
	tests := []struct {
		name       string
		backCalls  int
		shouldPass bool
	}{
		{name: "back called once", backCalls: 1, shouldPass: true},
		{name: "back called multiple times", backCalls: 5, shouldPass: true},
		{name: "back not called", backCalls: 0, shouldPass: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMockRouter()

			// Call Back
			for i := 0; i < tt.backCalls; i++ {
				mr.Back()
			}

			// Create mock testing.T
			mockT := &mockTestingT{}

			// Assert
			mr.AssertBackCalled(mockT)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "Expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "Expected assertion to fail")
			}
		})
	}
}

// TestMockRouter_AssertBackNotCalled tests the AssertBackNotCalled assertion helper.
func TestMockRouter_AssertBackNotCalled(t *testing.T) {
	tests := []struct {
		name       string
		backCalls  int
		shouldPass bool
	}{
		{name: "back not called", backCalls: 0, shouldPass: true},
		{name: "back called once", backCalls: 1, shouldPass: false},
		{name: "back called multiple times", backCalls: 3, shouldPass: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMockRouter()

			// Call Back
			for i := 0; i < tt.backCalls; i++ {
				mr.Back()
			}

			// Create mock testing.T
			mockT := &mockTestingT{}

			// Assert
			mr.AssertBackNotCalled(mockT)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "Expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "Expected assertion to fail")
			}
		})
	}
}

// TestMockRouter_AssertPushCount tests the AssertPushCount assertion helper.
func TestMockRouter_AssertPushCount(t *testing.T) {
	tests := []struct {
		name        string
		pushCount   int
		assertCount int
		shouldPass  bool
	}{
		{name: "exact match", pushCount: 3, assertCount: 3, shouldPass: true},
		{name: "count mismatch", pushCount: 2, assertCount: 5, shouldPass: false},
		{name: "zero count", pushCount: 0, assertCount: 0, shouldPass: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMockRouter()

			// Push
			for i := 0; i < tt.pushCount; i++ {
				mr.Push(&router.NavigationTarget{Path: "/test"})
			}

			// Create mock testing.T
			mockT := &mockTestingT{}

			// Assert
			mr.AssertPushCount(mockT, tt.assertCount)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "Expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "Expected assertion to fail")
			}
		})
	}
}

// TestMockRouter_AssertReplaceCount tests the AssertReplaceCount assertion helper.
func TestMockRouter_AssertReplaceCount(t *testing.T) {
	tests := []struct {
		name         string
		replaceCount int
		assertCount  int
		shouldPass   bool
	}{
		{name: "exact match", replaceCount: 2, assertCount: 2, shouldPass: true},
		{name: "count mismatch", replaceCount: 1, assertCount: 3, shouldPass: false},
		{name: "zero count", replaceCount: 0, assertCount: 0, shouldPass: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMockRouter()

			// Replace
			for i := 0; i < tt.replaceCount; i++ {
				mr.Replace(&router.NavigationTarget{Path: "/test"})
			}

			// Create mock testing.T
			mockT := &mockTestingT{}

			// Assert
			mr.AssertReplaceCount(mockT, tt.assertCount)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "Expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "Expected assertion to fail")
			}
		})
	}
}

// TestMockRouter_AssertBackCount tests the AssertBackCount assertion helper.
func TestMockRouter_AssertBackCount(t *testing.T) {
	tests := []struct {
		name        string
		backCount   int
		assertCount int
		shouldPass  bool
	}{
		{name: "exact match", backCount: 4, assertCount: 4, shouldPass: true},
		{name: "count mismatch", backCount: 2, assertCount: 7, shouldPass: false},
		{name: "zero count", backCount: 0, assertCount: 0, shouldPass: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMockRouter()

			// Back
			for i := 0; i < tt.backCount; i++ {
				mr.Back()
			}

			// Create mock testing.T
			mockT := &mockTestingT{}

			// Assert
			mr.AssertBackCount(mockT, tt.assertCount)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "Expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "Expected assertion to fail")
			}
		})
	}
}

// TestMockRouter_ConcurrentAccess tests thread-safe concurrent access.
func TestMockRouter_ConcurrentAccess(t *testing.T) {
	mr := NewMockRouter()
	var wg sync.WaitGroup

	// Concurrent pushes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mr.Push(&router.NavigationTarget{Path: "/test"})
		}(i)
	}

	// Concurrent replaces
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mr.Replace(&router.NavigationTarget{Path: "/login"})
		}(i)
	}

	// Concurrent backs
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mr.Back()
		}()
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = mr.GetPushCallCount()
			_ = mr.GetReplaceCallCount()
			_ = mr.GetBackCallCount()
			_ = mr.CurrentRoute()
		}()
	}

	wg.Wait()

	// Verify counts
	assert.Equal(t, 10, mr.GetPushCallCount())
	assert.Equal(t, 10, mr.GetReplaceCallCount())
	assert.Equal(t, 10, mr.GetBackCallCount())
}

// TestMockRouter_GetCallsReturnsDefensiveCopy tests that Get*Calls returns copies.
func TestMockRouter_GetCallsReturnsDefensiveCopy(t *testing.T) {
	mr := NewMockRouter()
	mr.Push(&router.NavigationTarget{Path: "/about"})
	mr.Replace(&router.NavigationTarget{Path: "/login"})

	// Get calls
	pushCalls := mr.GetPushCalls()
	replaceCalls := mr.GetReplaceCalls()

	// Modify returned slices
	pushCalls[0] = &router.NavigationTarget{Path: "/modified"}
	replaceCalls[0] = &router.NavigationTarget{Path: "/modified"}

	// Verify original not modified
	originalPush := mr.GetPushCalls()
	originalReplace := mr.GetReplaceCalls()

	assert.Equal(t, "/about", originalPush[0].Path)
	assert.Equal(t, "/login", originalReplace[0].Path)
}

// TestMockRouter_IntegrationScenario tests a realistic usage scenario.
func TestMockRouter_IntegrationScenario(t *testing.T) {
	mr := NewMockRouter()

	// Set initial route
	homeRoute := router.NewRoute("/", "home", nil, nil, "", nil, nil)
	mr.SetCurrentRoute(homeRoute)
	assert.Equal(t, "/", mr.CurrentRoute().Path)

	// Navigate to about
	mr.Push(&router.NavigationTarget{Path: "/about"})
	mr.AssertPushed(t, "/about")
	mr.AssertPushCount(t, 1)

	// Navigate to contact
	mr.Push(&router.NavigationTarget{Path: "/contact"})
	mr.AssertPushed(t, "/contact")
	mr.AssertPushCount(t, 2)

	// Go back
	mr.Back()
	mr.AssertBackCalled(t)
	mr.AssertBackCount(t, 1)

	// Replace with login
	mr.Replace(&router.NavigationTarget{Path: "/login"})
	mr.AssertReplaced(t, "/login")
	mr.AssertReplaceCount(t, 1)

	// Verify all calls
	assert.Equal(t, 2, mr.GetPushCallCount())
	assert.Equal(t, 1, mr.GetReplaceCallCount())
	assert.Equal(t, 1, mr.GetBackCallCount())

	// Reset and verify clean state
	mr.Reset()
	assert.Nil(t, mr.CurrentRoute())
	assert.Equal(t, 0, mr.GetPushCallCount())
	assert.Equal(t, 0, mr.GetReplaceCallCount())
	assert.Equal(t, 0, mr.GetBackCallCount())
}
