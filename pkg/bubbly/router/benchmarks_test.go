package router

import (
	"fmt"
	"runtime"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ============================================================================
// Route Matching Benchmarks (Target: < 100μs)
// ============================================================================

// BenchmarkRouteMatching_Static benchmarks static route matching
func BenchmarkRouteMatching_Static(b *testing.B) {
	matcher := NewRouteMatcher()
	_ = matcher.AddRoute("/", "home")
	_ = matcher.AddRoute("/about", "about")
	_ = matcher.AddRoute("/contact", "contact")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = matcher.Match("/about")
	}
}

// BenchmarkRouteMatching_Dynamic benchmarks dynamic route matching with params
func BenchmarkRouteMatching_Dynamic(b *testing.B) {
	matcher := NewRouteMatcher()
	_ = matcher.AddRoute("/user/:id", "user")
	_ = matcher.AddRoute("/post/:id/comment/:commentId", "comment")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = matcher.Match("/user/123")
	}
}

// BenchmarkRouteMatching_Wildcard benchmarks wildcard route matching
func BenchmarkRouteMatching_Wildcard(b *testing.B) {
	matcher := NewRouteMatcher()
	_ = matcher.AddRoute("/docs/:path*", "docs")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = matcher.Match("/docs/guide/getting-started/installation")
	}
}

// BenchmarkRouteMatching_Mixed benchmarks mixed route patterns
func BenchmarkRouteMatching_Mixed(b *testing.B) {
	matcher := NewRouteMatcher()
	_ = matcher.AddRoute("/", "home")
	_ = matcher.AddRoute("/users", "users-list")
	_ = matcher.AddRoute("/user/:id", "user-detail")
	_ = matcher.AddRoute("/user/:id/posts", "user-posts")
	_ = matcher.AddRoute("/posts/:postId/comments/:commentId", "comment")
	_ = matcher.AddRoute("/docs/:path*", "docs")
	_ = matcher.AddRoute("/profile/:id?", "profile")

	paths := []string{
		"/",
		"/users",
		"/user/123",
		"/user/456/posts",
		"/posts/789/comments/101",
		"/docs/api/reference",
		"/profile",
		"/profile/999",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := paths[i%len(paths)]
		_, _ = matcher.Match(path)
	}
}

// BenchmarkRouteMatching_LargeRouteSet benchmarks matching with many routes
func BenchmarkRouteMatching_LargeRouteSet(b *testing.B) {
	matcher := NewRouteMatcher()

	// Add 100 routes
	for i := 0; i < 100; i++ {
		path := fmt.Sprintf("/route%d/:id", i)
		name := fmt.Sprintf("route%d", i)
		_ = matcher.AddRoute(path, name)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = matcher.Match("/route50/123")
	}
}

// ============================================================================
// Navigation Benchmarks (Target: < 1ms overhead)
// ============================================================================

// BenchmarkNavigation_Push benchmarks Push navigation
func BenchmarkNavigation_Push(b *testing.B) {
	router, _ := NewRouterBuilder().
		Route("/", "home").
		Route("/about", "about").
		Route("/user/:id", "user").
		Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := router.Push(&NavigationTarget{Path: "/about"})
		if cmd != nil {
			_ = cmd()
		}
	}
}

// BenchmarkNavigation_Replace benchmarks Replace navigation
func BenchmarkNavigation_Replace(b *testing.B) {
	router, _ := NewRouterBuilder().
		Route("/", "home").
		Route("/about", "about").
		Route("/user/:id", "user").
		Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := router.Replace(&NavigationTarget{Path: "/about"})
		if cmd != nil {
			_ = cmd()
		}
	}
}

// BenchmarkNavigation_WithParams benchmarks navigation with parameters
func BenchmarkNavigation_WithParams(b *testing.B) {
	router, _ := NewRouterBuilder().
		Route("/user/:id", "user").
		Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := router.Push(&NavigationTarget{
			Path: fmt.Sprintf("/user/%d", i%1000),
		})
		if cmd != nil {
			_ = cmd()
		}
	}
}

// BenchmarkNavigation_WithQuery benchmarks navigation with query strings
func BenchmarkNavigation_WithQuery(b *testing.B) {
	router, _ := NewRouterBuilder().
		Route("/search", "search").
		Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := router.Push(&NavigationTarget{
			Path: "/search",
			Query: map[string]string{
				"q":    fmt.Sprintf("query%d", i%100),
				"page": "1",
			},
		})
		if cmd != nil {
			_ = cmd()
		}
	}
}

// BenchmarkNavigation_Concurrent benchmarks concurrent navigation
func BenchmarkNavigation_Concurrent(b *testing.B) {
	router, _ := NewRouterBuilder().
		Route("/", "home").
		Route("/about", "about").
		Route("/contact", "contact").
		Build()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			paths := []string{"/", "/about", "/contact"}
			path := paths[i%len(paths)]
			cmd := router.Push(&NavigationTarget{Path: path})
			if cmd != nil {
				_ = cmd()
			}
			i++
		}
	})
}

// ============================================================================
// History Operation Benchmarks (Target: < 50μs)
// ============================================================================

// BenchmarkHistory_Push benchmarks adding to history
func BenchmarkHistory_Push(b *testing.B) {
	history := &History{}
	route := &Route{
		Path:     "/test",
		Name:     "test",
		Params:   map[string]string{},
		Query:    map[string]string{},
		Meta:     map[string]interface{}{},
		Matched:  []*RouteRecord{},
		FullPath: "/test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		history.Push(route)
	}
}

// BenchmarkHistory_Replace benchmarks replacing history entry
func BenchmarkHistory_Replace(b *testing.B) {
	history := &History{}

	// Setup: Add initial entry
	route := &Route{
		Path:     "/initial",
		FullPath: "/initial",
	}
	history.Push(route)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newRoute := &Route{
			Path:     fmt.Sprintf("/page%d", i),
			FullPath: fmt.Sprintf("/page%d", i),
		}
		history.Replace(newRoute)
	}
}

// BenchmarkHistory_Back benchmarks router back navigation
func BenchmarkHistory_Back(b *testing.B) {
	router, _ := NewRouterBuilder().
		Route("/", "home").
		Route("/page1", "page1").
		Route("/page2", "page2").
		Route("/page3", "page3").
		Build()

	// Setup: Navigate forward to create history
	for i := 1; i <= 3; i++ {
		cmd := router.Push(&NavigationTarget{Path: fmt.Sprintf("/page%d", i)})
		if cmd != nil {
			_ = cmd()
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := router.Back()
		if cmd != nil {
			_ = cmd()
		}
		// Reset when we reach the beginning by going forward
		if !router.history.CanGoBack() {
			for router.history.CanGoForward() {
				cmd := router.Forward()
				if cmd != nil {
					_ = cmd()
				}
			}
		}
	}
}

// BenchmarkHistory_Forward benchmarks router forward navigation
func BenchmarkHistory_Forward(b *testing.B) {
	router, _ := NewRouterBuilder().
		Route("/", "home").
		Route("/page1", "page1").
		Route("/page2", "page2").
		Route("/page3", "page3").
		Build()

	// Setup: Navigate forward then back to create forward history
	for i := 1; i <= 3; i++ {
		cmd := router.Push(&NavigationTarget{Path: fmt.Sprintf("/page%d", i)})
		if cmd != nil {
			_ = cmd()
		}
	}

	// Go back to beginning
	for router.history.CanGoBack() {
		cmd := router.Back()
		if cmd != nil {
			_ = cmd()
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := router.Forward()
		if cmd != nil {
			_ = cmd()
		}
		// Reset when we reach the end by going back
		if !router.history.CanGoForward() {
			for router.history.CanGoBack() {
				cmd := router.Back()
				if cmd != nil {
					_ = cmd()
				}
			}
		}
	}
}

// BenchmarkHistory_Go benchmarks Go(n) navigation
func BenchmarkHistory_Go(b *testing.B) {
	router, _ := NewRouterBuilder().
		Route("/", "home").
		Build()

	// Setup: Create 20 history entries
	for i := 0; i < 20; i++ {
		_ = router.registry.Register(fmt.Sprintf("/page%d", i), fmt.Sprintf("page%d", i), nil)
		cmd := router.Push(&NavigationTarget{Path: fmt.Sprintf("/page%d", i)})
		if cmd != nil {
			_ = cmd()
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Alternate between backward and forward jumps
		if i%2 == 0 {
			cmd := router.Go(-5)
			if cmd != nil {
				_ = cmd()
			}
		} else {
			cmd := router.Go(5)
			if cmd != nil {
				_ = cmd()
			}
		}
	}
}

// BenchmarkHistory_CanGoBack benchmarks checking if back is possible
func BenchmarkHistory_CanGoBack(b *testing.B) {
	history := &History{}

	// Setup: Add some history entries
	for i := 0; i < 10; i++ {
		route := &Route{
			Path:     fmt.Sprintf("/page%d", i),
			FullPath: fmt.Sprintf("/page%d", i),
		}
		history.Push(route)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = history.CanGoBack()
	}
}

// BenchmarkHistory_CanGoForward benchmarks checking if forward is possible
func BenchmarkHistory_CanGoForward(b *testing.B) {
	history := &History{}

	// Setup: Add some history entries and move back
	for i := 0; i < 10; i++ {
		route := &Route{
			Path:     fmt.Sprintf("/page%d", i),
			FullPath: fmt.Sprintf("/page%d", i),
		}
		history.Push(route)
	}

	// Manually move back by updating current
	history.mu.Lock()
	history.current = 5
	history.mu.Unlock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = history.CanGoForward()
	}
}

// ============================================================================
// Guard Execution Benchmarks (Target: < 10μs)
// ============================================================================

// BenchmarkGuard_SingleBeforeEach benchmarks single global guard
func BenchmarkGuard_SingleBeforeEach(b *testing.B) {
	router, _ := NewRouterBuilder().
		Route("/", "home").
		Route("/about", "about").
		BeforeEach(func(to, from *Route, next NextFunc) {
			next(nil) // Allow navigation
		}).
		Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := router.Push(&NavigationTarget{Path: "/about"})
		if cmd != nil {
			_ = cmd()
		}
	}
}

// BenchmarkGuard_MultipleBeforeEach benchmarks multiple global guards
func BenchmarkGuard_MultipleBeforeEach(b *testing.B) {
	guardCounts := []int{1, 3, 5, 10}

	for _, count := range guardCounts {
		b.Run(fmt.Sprintf("guards_%d", count), func(b *testing.B) {
			builder := NewRouterBuilder().
				Route("/", "home").
				Route("/about", "about")

			// Add multiple guards
			for i := 0; i < count; i++ {
				builder = builder.BeforeEach(func(to, from *Route, next NextFunc) {
					next(nil) // Allow navigation
				})
			}

			router, _ := builder.Build()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cmd := router.Push(&NavigationTarget{Path: "/about"})
				if cmd != nil {
					_ = cmd()
				}
			}
		})
	}
}

// BenchmarkGuard_WithLogic benchmarks guard with actual logic
func BenchmarkGuard_WithLogic(b *testing.B) {
	isAuthenticated := true

	router, _ := NewRouterBuilder().
		Route("/", "home").
		Route("/dashboard", "dashboard").
		BeforeEach(func(to, from *Route, next NextFunc) {
			// Simulate auth check
			requiresAuth, ok := to.GetMeta("requiresAuth")
			if ok && requiresAuth == true && !isAuthenticated {
				next(&NavigationTarget{Path: "/"})
			} else {
				next(nil)
			}
		}).
		Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := router.Push(&NavigationTarget{Path: "/dashboard"})
		if cmd != nil {
			_ = cmd()
		}
	}
}

// BenchmarkGuard_AfterEach benchmarks after hooks
func BenchmarkGuard_AfterEach(b *testing.B) {
	counter := 0

	router, _ := NewRouterBuilder().
		Route("/", "home").
		Route("/about", "about").
		AfterEach(func(to, from *Route) {
			counter++ // Simulate analytics
		}).
		Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := router.Push(&NavigationTarget{Path: "/about"})
		if cmd != nil {
			_ = cmd()
		}
	}
}

// ============================================================================
// Memory Allocation Benchmarks (Target: < 1KB per route)
// ============================================================================

// BenchmarkMemory_RouteCreation benchmarks route object allocation
func BenchmarkMemory_RouteCreation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewRoute(
			fmt.Sprintf("/user/%d", i),
			"user",
			map[string]string{"id": fmt.Sprintf("%d", i)},
			map[string]string{"tab": "profile"},
			"#section",
			map[string]interface{}{"title": "User Profile"},
			[]*RouteRecord{},
		)
	}
}

// BenchmarkMemory_RouteRegistration benchmarks route record allocation
func BenchmarkMemory_RouteRegistration(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry := NewRouteRegistry()
		_ = registry.Register("/user/:id", "user", map[string]interface{}{
			"requiresAuth": true,
		})
	}
}

// BenchmarkMemory_RouterBuilder benchmarks router builder allocation
func BenchmarkMemory_RouterBuilder(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewRouterBuilder().
			Route("/", "home").
			Route("/about", "about").
			Route("/user/:id", "user").
			Build()
	}
}

// BenchmarkMemory_NavigationTarget benchmarks navigation target allocation
func BenchmarkMemory_NavigationTarget(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &NavigationTarget{
			Path: "/user/123",
			Query: map[string]string{
				"tab":  "profile",
				"edit": "true",
			},
			Hash: "#section",
		}
	}
}

// ============================================================================
// Integration Benchmarks
// ============================================================================

// BenchmarkIntegration_FullNavigation benchmarks complete navigation flow
func BenchmarkIntegration_FullNavigation(b *testing.B) {
	// Create a realistic router with guards
	router, _ := NewRouterBuilder().
		Route("/", "home").
		Route("/about", "about").
		Route("/user/:id", "user").
		Route("/dashboard", "dashboard").
		BeforeEach(func(to, from *Route, next NextFunc) {
			// Simulate auth check
			next(nil)
		}).
		AfterEach(func(to, from *Route) {
			// Simulate analytics
		}).
		Build()

	paths := []string{"/", "/about", "/user/123", "/dashboard"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := paths[i%len(paths)]
		cmd := router.Push(&NavigationTarget{Path: path})
		if cmd != nil {
			_ = cmd()
		}
	}
}

// BenchmarkIntegration_WithComponents benchmarks navigation with component updates
func BenchmarkIntegration_WithComponents(b *testing.B) {
	// Create simple components
	homeComponent, _ := bubbly.NewComponent("Home").
		Template(func(ctx bubbly.RenderContext) string {
			return "Home"
		}).
		Build()

	aboutComponent, _ := bubbly.NewComponent("About").
		Template(func(ctx bubbly.RenderContext) string {
			return "About"
		}).
		Build()

	// Create router with components
	router, _ := NewRouterBuilder().Build()
	router.registry.routes = []*RouteRecord{
		{
			Path:      "/",
			Name:      "home",
			Component: homeComponent,
		},
		{
			Path:      "/about",
			Name:      "about",
			Component: aboutComponent,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := []string{"/", "/about"}[i%2]
		cmd := router.Push(&NavigationTarget{Path: path})
		if cmd != nil {
			msg := cmd()
			// Simulate component update
			switch msg := msg.(type) {
			case RouteChangedMsg:
				if msg.To != nil && len(msg.To.Matched) > 0 {
					if component, ok := msg.To.Matched[0].Component.(bubbly.Component); ok && component != nil {
						_, _ = component.Update(tea.KeyMsg{})
					}
				}
			}
		}
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// benchName creates a consistent benchmark name
func benchName(prefix string, value int) string {
	return fmt.Sprintf("%s_%d", prefix, value)
}

// getAllocSize estimates allocation size in bytes
func getAllocSize(b *testing.B, allocBytes uint64, allocCount uint64) int {
	if allocCount == 0 {
		return 0
	}
	return int(allocBytes / allocCount)
}

// printMemStats prints memory statistics (for manual inspection)
func printMemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", m.Alloc/1024/1024)
	fmt.Printf("\tTotalAlloc = %v MiB", m.TotalAlloc/1024/1024)
	fmt.Printf("\tSys = %v MiB", m.Sys/1024/1024)
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}
