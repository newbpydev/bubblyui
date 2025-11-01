# User Workflow: Composition API

## Primary User Journey

### Journey: Developer Creates First Composable

1. **Entry Point**: Developer wants to extract counter logic for reuse
   - System response: Composition API provides composable pattern
   - UI update: N/A (development phase)

2. **Step 1**: Create composable function
   ```go
   func UseCounter(ctx *Context, initial int) (*Ref[int], func(), func()) {
       count := ctx.Ref(initial)
       
       increment := func() {
           count.Set(count.Get() + 1)
       }
       
       decrement := func() {
           count.Set(count.Get() - 1)
       }
       
       return count, increment, decrement
   }
   ```
   - System response: Composable defined with type-safe signature
   - Developer sees: Clean,

 reusable function
   - Ready for: Usage in components

3. **Step 2**: Use composable in first component
   ```go
   NewComponent("Counter").
       Setup(func(ctx *Context) {
           count, inc, dec := UseCounter(ctx, 0)
           
           ctx.Expose("count", count)
           ctx.On("increment", func(_ interface{}) { inc() })
           ctx.On("decrement", func(_ interface{}) { dec() })
       }).
       Build()
   ```
   - System response: Composable executes, returns values
   - Component state: Counter logic encapsulated
   - Ready for: Reuse in other components

4. **Step 3**: Reuse composable in second component
   ```go
   NewComponent("ScoreBoard").
       Setup(func(ctx *Context) {
           player1Score, p1Inc, _ := UseCounter(ctx, 0)
           player2Score, p2Inc, _ := UseCounter(ctx, 0)
           
           ctx.Expose("p1", player1Score)
           ctx.Expose("p2", player2Score)
       }).
       Build()
   ```
   - System response: Same logic, multiple instances
   - Result: Clean code reuse
   - Developer sees: DRY principle achieved

5. **Completion**: Reusable logic pattern established
   - Multiple components using same composable
   - Logic tested independently
   - Easy to maintain and update

---

## Alternative Paths

### Scenario A: Using Standard Composables

1. **Developer uses UseAsync for data fetching**
   ```go
   Setup(func(ctx *Context) {
       userData := UseAsync(ctx, func() (*User, error) {
           return fetchUser()
       })
       
       ctx.OnMounted(func() {
           userData.Execute()
       })
       
       ctx.Expose("user", userData.Data)
       ctx.Expose("loading", userData.Loading)
       ctx.Expose("error", userData.Error)
   })
   ```

2. **System manages loading states automatically**
   - loading=true while fetching
   - data populated on success
   - error set on failure
   - All reactive

**Use Case:** Async operations with minimal boilerplate

### Scenario B: Provide/Inject for Dependency Injection

1. **Root component provides theme**
   ```go
   Setup(func(ctx *Context) {
       theme := ctx.Ref("dark")
       ctx.Provide("theme", theme)
       
       ctx.On("toggleTheme", func(_ interface{}) {
           current := theme.Get()
           if current == "dark" {
               theme.Set("light")
           } else {
               theme.Set("dark")
           }
       })
   })
   ```

2. **Deep child component injects theme**
   ```go
   Setup(func(ctx *Context) {
       theme := ctx.Inject("theme", ctx.Ref("light")).(*Ref[string])
       
       style := ctx.Computed(func() string {
           if theme.Get() == "dark" {
               return "background: black; color: white"
           }
           return "background: white; color: black"
       })
       
       ctx.Expose("style", style)
   })
   ```

3. **Theme changes propagate automatically**
   - Root updates theme
   - All children see new value
   - No prop drilling needed

**Use Case:** Global state without prop drilling

### Scenario C: Composable Chains

1. **Create low-level composable**
   ```go
   func UseEventListener(ctx *Context, event string, handler func()) {
       ctx.OnMounted(func() {
           // Register listener
       })
       
       ctx.OnUnmounted(func() {
           // Cleanup listener
       })
   }
   ```

2. **Create mid-level composable using low-level**
   ```go
   func UseMouse(ctx *Context) (*Ref[int], *Ref[int]) {
       x := ctx.Ref(0)
       y := ctx.Ref(0)
       
       UseEventListener(ctx, "mousemove", func() {
           // Update x, y
       })
       
       return x, y
   }
   ```

3. **Use high-level composable**
   ```go
   Setup(func(ctx *Context) {
       x, y := UseMouse(ctx)
       ctx.Expose("mouseX", x)
       ctx.Expose("mouseY", y)
   })
   ```

**Use Case:** Building abstractions on abstractions

### Scenario D: Using Dependency Interface with UseEffect (Quality of Life Enhancement)

1. **Developer creates typed refs naturally**
   ```go
   Setup(func(ctx *Context) {
       // Create typed refs - no need for Ref[any]
       count := ctx.Ref(0)        // *Ref[int]
       name := ctx.Ref("Alice")   // *Ref[string]
       
       // UseEffect accepts any Dependency (Ref or Computed)
       UseEffect(ctx, func() UseEffectCleanup {
           currentCount := count.Get().(int)
           currentName := name.Get().(string)
           fmt.Printf("%s: %d\n", currentName, currentCount)
           return nil
       }, count, name)  // Works with typed refs!
   })
   ```
   - System response: No type conversion needed
   - Developer sees: Clean, ergonomic API
   - Ready for: Production use

2. **UseEffect with Computed values**
   ```go
   Setup(func(ctx *Context) {
       firstName := ctx.Ref("John")
       lastName := ctx.Ref("Doe")
       
       // Computed implements Dependency interface
       fullName := ctx.Computed(func() string {
           return firstName.Get() + " " + lastName.Get()
       })
       
       // Watch computed values directly
       UseEffect(ctx, func() UseEffectCleanup {
           fmt.Printf("Name changed: %s\n", fullName.Get())
           return nil
       }, fullName)  // Computed as dependency!
   })
   ```
   - System response: Computed values are watchable
   - Aligns with: Vue 3 behavior
   - Developer sees: Consistent, flexible API

**Use Case:** Type-safe reactive dependencies without boilerplate

### Scenario E: Form Management with UseForm

1. **Create form with validation**
   ```go
   type SignupForm struct {
       Email    string
       Password string
       Name     string
   }
   
   Setup(func(ctx *Context) {
       form := UseForm(ctx, SignupForm{}, func(f SignupForm) map[string]string {
           errors := make(map[string]string)
           if !strings.Contains(f.Email, "@") {
               errors["email"] = "Invalid email"
           }
           if len(f.Password) < 8 {
               errors["password"] = "Too short"
           }
           return errors
       })
       
       ctx.On("submit", func(_ interface{}) {
           form.Submit()
       })
       
       ctx.Expose("form", form)
   })
   ```

2. **Form handles validation automatically**
   - Real-time validation on field change
   - Error messages reactive
   - Submit only when valid

**Use Case:** Complex forms with validation

---

## Error Handling Flows

### Error 1: Composable Called Outside Setup
- **Trigger**: Developer calls composable in template or Update
- **User sees**: 
  ```
  Error: composable UseCounter called outside Setup function
  Composables must be called within Setup or other composables
  ```
- **Recovery**: Move composable call to Setup function

**Example:**
```go
// ❌ Wrong
Template(func(ctx RenderContext) string {
    count, _, _ := UseCounter(???)  // No context!
    return fmt.Sprintf("%d", count.Get())
})

// ✅ Correct
Setup(func(ctx *Context) {
    count, _, _ := UseCounter(ctx, 0)
    ctx.Expose("count", count)
})
```

### Error 2: Circular Composable Dependencies
- **Trigger**: ComposableA calls ComposableB which calls ComposableA
- **User sees**:
  ```
  Error: circular composable dependency detected
  Call stack: UseA → UseB → UseA
  ```
- **Recovery**: Restructure composables to break cycle

**Example:**
```go
// ❌ Circular
func UseA(ctx *Context) {
    UseB(ctx)  // Calls B
}

func UseB(ctx *Context) {
    UseA(ctx)  // Calls A - CIRCULAR!
}

// ✅ Fixed
func UseShared(ctx *Context) {
    // Shared logic
}

func UseA(ctx *Context) {
    UseShared(ctx)
}

func UseB(ctx *Context) {
    UseShared(ctx)
}
```

### Error 3: Inject Without Provide
- **Trigger**: Child injects key that wasn't provided
- **User sees**: Returns default value (no error if default provided)
- **Recovery**: Provide value in parent or handle default case

**Example:**
```go
// ❌ Not provided
Setup(func(ctx *Context) {
    // Nothing provided
})

// Child
Setup(func(ctx *Context) {
    theme := ctx.Inject("theme", ctx.Ref("light"))
    // Gets default "light"
})

// ✅ Provided
Setup(func(ctx *Context) {
    ctx.Provide("theme", ctx.Ref("dark"))
})
```

### Error 4: Type Mismatch in Inject
- **Trigger**: Injected value type doesn't match expected
- **User sees**: Panic on type assertion
- **Recovery**: Fix type or use type switch

**Example:**
```go
// ❌ Type mismatch
ctx.Provide("count", 42)  // Provided int

// Consumer
count := ctx.Inject("count", ctx.Ref(0)).(*Ref[int])  // Expected Ref!
// Panic: interface conversion

// ✅ Correct types
ctx.Provide("count", ctx.Ref(42))  // Provide Ref
count := ctx.Inject("count", ctx.Ref(0)).(*Ref[int])  // Works!
```

### Error 5: Composable State Leaking
- **Trigger**: Composable uses global variable instead of Context
- **User sees**: State shared between component instances
- **Recovery**: Use Context or closure for state

**Example:**
```go
// ❌ Global state (leaks between instances)
var globalCount int

func UseBadCounter(ctx *Context) (*Ref[int], func()) {
    count := ctx.Ref(globalCount)  // BAD!
    increment := func() {
        globalCount++
        count.Set(globalCount)
    }
    return count, increment
}

// ✅ Context/closure state (isolated)
func UseGoodCounter(ctx *Context, initial int) (*Ref[int], func()) {
    count := ctx.Ref(initial)
    increment := func() {
        count.Set(count.Get() + 1)
    }
    return count, increment
}
```

---

## State Transitions

### Composable Lifecycle
```
Composable defined
    ↓
Component Setup() called
    ↓
Composable called with Context
    ↓
Composable creates Refs/Computed
    ↓
Composable registers hooks
    ↓
Composable returns values
    ↓
Component uses values
    ↓
Component lifecycle runs (mounted, updated)
    ↓
Composable hooks execute
    ↓
Component unmounts
    ↓
Composable cleanup executes
```

### Provide/Inject State Flow
```
Provider Setup
    ↓
ctx.Provide(key, value) called
    ↓
Value stored in component.provides
    ↓
Consumer Setup
    ↓
ctx.Inject(key, default) called
    ↓
Walk up component tree
    ↓
Found: return provided value
Not Found: return default
    ↓
Consumer watches value
    ↓
Provider changes value
    ↓
Consumer sees new value (reactive)
```

---

## Integration Points

### Connected to: Reactivity System (Feature 01)
- **Uses:** Ref, Computed, Watch
- **Flow:** Composables create and return reactive values
- **Data:** All state management through reactivity

### Connected to: Component System (Feature 02)
- **Uses:** Context, Setup function
- **Flow:** Composables called in Setup, values exposed to template
- **Data:** Component state populated by composables

### Connected to: Lifecycle Hooks (Feature 03)
- **Uses:** onMounted, onUpdated, onUnmounted
- **Flow:** Composables register hooks for initialization and cleanup
- **Data:** Lifecycle callbacks managed by composables

---

## Performance Considerations

### Composable Call Overhead
**Target:** < 100ns per composable call

**Optimization:**
```go
// Lightweight composable
func UseState[T any](ctx *Context, initial T) UseStateReturn[T] {
    value := ctx.Ref(initial)  // Fast ref creation
    
    return UseStateReturn[T]{  // Struct return (no heap allocation)
        Value: value,
        Set:   func(v T) { value.Set(v) },
    }
}
```

### Provide/Inject Lookup
**Target:** < 500ns for inject

**Optimization:**
- Cache inject results
- Limit tree traversal depth
- Use map for O(1) lookup

---

## Common Patterns

### Pattern 1: Shared Auth State
```go
func UseAuth(ctx *Context) UseAuthReturn {
    user := ctx.Inject("currentUser", ctx.Ref[*User](nil)).(*Ref[*User])
    
    isAuthenticated := ctx.Computed(func() bool {
        return user.Get() != nil
    })
    
    login := func(email, password string) error {
        // Login logic
        user.Set(loggedInUser)
        return nil
    }
    
    logout := func() {
        user.Set(nil)
    }
    
    return UseAuthReturn{
        User:            user,
        IsAuthenticated: isAuthenticated,
        Login:           login,
        Logout:          logout,
    }
}
```

### Pattern 2: Pagination
```go
func UsePagination(ctx *Context, itemsPerPage int) UsePaginationReturn {
    currentPage := ctx.Ref(1)
    totalItems := ctx.Ref(0)
    
    totalPages := ctx.Computed(func() int {
        return (totalItems.Get() + itemsPerPage - 1) / itemsPerPage
    })
    
    nextPage := func() {
        if currentPage.Get() < totalPages.Get() {
            currentPage.Set(currentPage.Get() + 1)
        }
    }
    
    prevPage := func() {
        if currentPage.Get() > 1 {
            currentPage.Set(currentPage.Get() - 1)
        }
    }
    
    return UsePaginationReturn{
        CurrentPage: currentPage,
        TotalPages:  totalPages,
        NextPage:    nextPage,
        PrevPage:    prevPage,
        SetTotal:    func(total int) { totalItems.Set(total) },
    }
}
```

### Pattern 3: Undo/Redo
```go
func UseHistory[T any](ctx *Context, initial T) UseHistoryReturn[T] {
    history := ctx.Ref([]T{initial})
    index := ctx.Ref(0)
    
    current := ctx.Computed(func() T {
        return history.Get()[index.Get()]
    })
    
    canUndo := ctx.Computed(func() bool {
        return index.Get() > 0
    })
    
    canRedo := ctx.Computed(func() bool {
        return index.Get() < len(history.Get())-1
    })
    
    push := func(value T) {
        h := history.Get()[:index.Get()+1]
        h = append(h, value)
        history.Set(h)
        index.Set(len(h) - 1)
    }
    
    undo := func() {
        if canUndo.Get() {
            index.Set(index.Get() - 1)
        }
    }
    
    redo := func() {
        if canRedo.Get() {
            index.Set(index.Get() + 1)
        }
    }
    
    return UseHistoryReturn[T]{
        Current: current,
        CanUndo: canUndo,
        CanRedo: canRedo,
        Push:    push,
        Undo:    undo,
        Redo:    redo,
    }
}
```

---

## Testing Workflow

### Unit Test: Composable in Isolation
```go
func TestUseCounter(t *testing.T) {
    // Arrange
    ctx := createTestContext()
    
    // Act
    count, increment, decrement := UseCounter(ctx, 0)
    
    // Assert
    assert.Equal(t, 0, count.Get())
    
    increment()
    assert.Equal(t, 1, count.Get())
    
    decrement()
    assert.Equal(t, 0, count.Get())
}
```

### Integration Test: Composable in Component
```go
func TestComponentWithComposable(t *testing.T) {
    // Arrange
    component := NewComponent("Test").
        Setup(func(ctx *Context) {
            count, inc, _ := UseCounter(ctx, 0)
            ctx.Expose("count", count)
            ctx.On("click", func(_ interface{}) {
                inc()
            })
        }).
        Build()
    
    // Act
    component.Init()
    component.Update(ClickMsg{})
    
    // Assert
    count := component.Get("count").(*Ref[int])
    assert.Equal(t, 1, count.Get())
}
```

### Test: Provide/Inject
```go
func TestProvideInject(t *testing.T) {
    // Arrange
    parent := NewComponent("Parent").
        Setup(func(ctx *Context) {
            ctx.Provide("theme", ctx.Ref("dark"))
        }).
        Children(child).
        Build()
    
    child := NewComponent("Child").
        Setup(func(ctx *Context) {
            theme := ctx.Inject("theme", ctx.Ref("light"))
            ctx.Expose("theme", theme)
        }).
        Build()
    
    // Act
    parent.Init()
    child.Init()
    
    // Assert
    theme := child.Get("theme").(*Ref[string])
    assert.Equal(t, "dark", theme.Get())
}
```

---

## Documentation for Users

### Quick Start
1. Create composable function with `Use*` prefix
2. Accept `Context` as first parameter
3. Create Refs/Computed using context
4. Return stable references and functions
5. Call in component Setup

### Best Practices
- Composables return objects with named fields
- Use generics for type safety
- Register cleanup in lifecycle hooks
- Test composables independently
- Document composable contract
- Use provide/inject for cross-tree data
- Avoid global state in composables

### Troubleshooting
- **Composable not working?** Check it's called in Setup
- **State shared between instances?** Use Context not globals
- **Inject returns default?** Check parent provides value
- **Type panic?** Verify type assertion matches provided type
- **Cleanup not running?** Register with lifecycle hooks

---

## Phase 8: Performance Optimization & Monitoring Workflows

### Workflow 1: Enabling Timer Pooling (Optional Optimization)

**Scenario:** Developer wants to optimize UseDebounce/UseThrottle performance

1. **Step 1**: Import timer pool package
   ```go
   import "github.com/newbpydev/bubblyui/pkg/bubbly/composables/timerpool"
   ```
   - System response: Timer pool available for use
   - UI update: N/A (performance optimization)

2. **Step 2**: Enable global timer pool
   ```go
   func init() {
       timerpool.EnableGlobalPool()
   }
   ```
   - System response: All UseDebounce/UseThrottle use pooled timers
   - Performance: 865ns → 450ns (52% improvement)
   - Memory: Zero allocations after warmup

3. **Step 3**: Monitor pool statistics
   ```go
   stats := timerpool.GlobalPool.Stats()
   fmt.Printf("Active: %d, Hits: %d, Misses: %d\n", 
       stats.Active, stats.Hits, stats.Misses)
   ```
   - System response: Pool statistics displayed
   - Developer sees: Hit rate, active timers, allocation savings

4. **Completion**: Timer pooling active
   - UseDebounce/UseThrottle automatically optimized
   - No code changes required in composables
   - Cleanup still handled automatically

**Use Case:** Production applications with many debounce/throttle composables

---

### Workflow 2: Enabling Reflection Caching (Optional Optimization)

**Scenario:** Developer wants to optimize UseForm.SetField performance

1. **Step 1**: Import reflection cache
   ```go
   import "github.com/newbpydev/bubblyui/pkg/bubbly/composables/reflectcache"
   ```
   - System response: Reflection cache available

2. **Step 2**: Enable global cache
   ```go
   func init() {
       reflectcache.EnableGlobalCache()
   }
   ```
   - System response: All UseForm operations use cached field indices
   - Performance: 422ns → 300ns (29% improvement)
   - Memory: ~100B per cached struct type

3. **Step 3**: Monitor cache performance
   ```go
   stats := reflectcache.GlobalCache.Stats()
   fmt.Printf("Types cached: %d, Hit rate: %.2f%%\n",
       stats.TypesCached, stats.HitRate*100)
   ```
   - System response: Cache statistics displayed
   - Developer sees: Hit rate (should be > 95%)

4. **Step 4**: Pre-warm cache (optional)
   ```go
   type MyForm struct {
       Name  string
       Email string
   }
   
   reflectcache.GlobalCache.WarmUp(MyForm{})
   ```
   - System response: Form type cached before first use
   - Benefit: First SetField call is fast

**Use Case:** Production applications with heavy form usage

---

### Workflow 3: Setting Up Performance Monitoring

**Scenario:** Developer wants to monitor composable usage in production

1. **Step 1**: Choose monitoring backend
   ```go
   import (
       "github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
       "github.com/prometheus/client_golang/prometheus"
   )
   ```
   - Options: Prometheus (default), StatsD, custom

2. **Step 2**: Initialize metrics collector
   ```go
   func init() {
       metrics := monitoring.NewPrometheusMetrics(prometheus.DefaultRegisterer)
       monitoring.SetGlobalMetrics(metrics)
   }
   ```
   - System response: Composable metrics collection enabled
   - Overhead: < 5ns per composable call

3. **Step 3**: Expose metrics endpoint
   ```go
   import (
       "net/http"
       "github.com/prometheus/client_golang/prometheus/promhttp"
   )
   
   http.Handle("/metrics", promhttp.Handler())
   http.ListenAndServe(":2112", nil)
   ```
   - System response: Metrics available at /metrics
   - Format: Prometheus exposition format

4. **Step 4**: Configure Prometheus scraper
   ```yaml
   # prometheus.yml
   scrape_configs:
     - job_name: 'bubblyui'
       static_configs:
         - targets: ['localhost:2112']
   ```
   - System response: Prometheus scraping metrics
   - Frequency: Every 15s (default)

5. **Step 5**: Create Grafana dashboard
   - Import pre-built BubblyUI dashboard
   - Metrics available:
     - Composable creation rates
     - Provide/Inject tree depth distribution
     - Cache hit rates
     - Memory allocation patterns

**Metrics Exposed:**
- `bubblyui_composable_creations_total{name="UseState"}`
- `bubblyui_provide_inject_depth_seconds`
- `bubblyui_cache_hits_total{cache="reflection"}`
- `bubblyui_cache_misses_total{cache="reflection"}`
- `bubblyui_allocation_bytes{composable="UseForm"}`

**Use Case:** Production monitoring and alerting

---

### Workflow 4: Performance Regression Testing

**Scenario:** Developer wants to prevent performance regressions in CI/CD

1. **Step 1**: Add benchmark workflow to GitHub Actions
   ```yaml
   # .github/workflows/benchmark.yml
   name: Performance Benchmarks
   on: [pull_request]
   
   jobs:
     benchmark:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-go@v5
           with:
             go-version: '1.22'
         
         - name: Run benchmarks
           run: |
             go test -bench=. -benchmem -count=10 \
               ./pkg/bubbly/composables/ > new.txt
         
         - name: Download baseline
           uses: actions/download-artifact@v3
           with:
             name: benchmark-baseline
             path: .
         
         - name: Install benchstat
           run: go install golang.org/x/perf/cmd/benchstat@latest
         
         - name: Compare benchmarks
           run: |
             benchstat baseline.txt new.txt
         
         - name: Check for regressions
           run: |
             # Fail if any benchmark regresses > 10%
             benchstat -delta-test=ttest baseline.txt new.txt | \
               grep -E '\+[0-9]{2}\.' && exit 1 || exit 0
   ```
   - System response: Benchmarks run on every PR
   - Failure: PR blocked if >10% regression detected

2. **Step 2**: Update baseline (after approval)
   ```bash
   # Run locally
   go test -bench=. -benchmem -count=10 \
     ./pkg/bubbly/composables/ > benchmarks/baseline.txt
   
   # Commit baseline
   git add benchmarks/baseline.txt
   git commit -m "Update benchmark baseline"
   ```
   - System response: New baseline established
   - Next PR: Compared against updated baseline

3. **Step 3**: Analyze regression (if detected)
   ```bash
   # Local analysis
   benchstat baseline.txt new.txt
   
   # Output example:
   # name              old time/op  new time/op  delta
   # UseState-6          50.3ns ± 2%  65.4ns ± 3%  +30.02% (p=0.000 n=10+10)
   ```
   - System response: Identify which composable regressed
   - Action: Investigate code changes causing regression

4. **Completion**: Continuous performance monitoring
   - All PRs automatically tested
   - Regressions caught before merge
   - Performance trends tracked over time

**Use Case:** Maintain performance SLAs in production

---

### Workflow 5: Production Profiling

**Scenario:** Developer needs to profile composable performance in production

1. **Step 1**: Enable profiling endpoint
   ```go
   import (
       "net/http"
       _ "net/http/pprof"
   )
   
   // In main.go
   go func() {
       http.ListenAndServe("localhost:6060", nil)
   }()
   ```
   - System response: Profiling endpoints available
   - Security: Only accessible on localhost

2. **Step 2**: Capture CPU profile
   ```bash
   # 30-second CPU profile
   go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
   
   # Wait 30 seconds...
   # (pprof) top
   # Shows top CPU consumers
   ```
   - System response: CPU profile captured
   - Analysis: Identify hot code paths

3. **Step 3**: Capture memory profile
   ```bash
   # Heap profile
   go tool pprof http://localhost:6060/debug/pprof/heap
   
   # (pprof) top
   # Shows top memory allocators
   ```
   - System response: Memory profile captured
   - Analysis: Identify memory leaks or heavy allocations

4. **Step 4**: Capture goroutine profile
   ```bash
   # Goroutine profile
   go tool pprof http://localhost:6060/debug/pprof/goroutine
   
   # (pprof) list UseAsync
   # Shows goroutine creation in UseAsync
   ```
   - System response: Goroutine profile captured
   - Analysis: Verify no goroutine leaks

5. **Step 5**: Generate flame graph (optional)
   ```bash
   # CPU flame graph
   go tool pprof -http=:8080 \
     http://localhost:6060/debug/pprof/profile?seconds=30
   
   # Opens browser with interactive flame graph
   ```
   - System response: Visual flame graph displayed
   - Analysis: See call stack hierarchy

**Use Case:** Production debugging and optimization

---

### Workflow 6: Monitoring Tree Depth

**Scenario:** Developer wants to ensure Provide/Inject tree depth stays optimal

1. **Step 1**: Add tree depth monitoring
   ```go
   import "github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
   
   // Automatic if monitoring enabled
   // Tree depth recorded on every Inject call
   ```
   - System response: Tree depth tracked automatically

2. **Step 2**: Set up alerting
   ```yaml
   # Prometheus alerting rule
   groups:
     - name: bubblyui
       rules:
         - alert: DeepProvideInjectTree
           expr: histogram_quantile(0.95, bubblyui_provide_inject_depth_seconds) > 10
           for: 5m
           annotations:
             summary: "Provide/Inject tree too deep"
             description: "95th percentile tree depth is {{ $value }}, consider refactoring"
   ```
   - System response: Alert triggered if depth > 10
   - Action: Refactor component hierarchy

3. **Step 3**: Query current metrics
   ```promql
   # Average tree depth
   avg(bubblyui_provide_inject_depth_seconds)
   
   # 95th percentile
   histogram_quantile(0.95, bubblyui_provide_inject_depth_seconds)
   
   # Max depth observed
   max(bubblyui_provide_inject_depth_seconds)
   ```
   - System response: Current depth statistics
   - Target: Keep < 10 for best performance (12ns vs 56ns+ for uncached)

4. **Completion**: Proactive monitoring
   - Depth tracked continuously
   - Alerts prevent performance degradation
   - Architecture stays optimal

**Use Case:** Large applications with complex component trees

---

## Documentation for Users - Phase 8

### Optimization Quick Start
1. Install benchstat: `go install golang.org/x/perf/cmd/benchstat@latest`
2. Run benchmarks: `go test -bench=. -benchmem -count=10`
3. Enable optimizations: Import and enable pools/caches
4. Verify improvement: Re-run benchmarks and compare
5. Monitor production: Set up metrics collection

### Monitoring Best Practices
- Enable Prometheus metrics in production
- Set up Grafana dashboards for visualization
- Configure alerting for regressions
- Profile regularly with pprof
- Track tree depth to maintain performance
- Run benchmarks in CI/CD on every PR

### Performance Targets (Post-Optimization)
- Timer pool: < 50ns acquisition overhead
- Reflection cache: < 5ns hit, 95%+ hit rate
- UseDebounce: < 500ns creation (vs 865ns)
- UseThrottle: < 250ns creation (vs 473ns)
- UseForm.SetField: < 300ns (vs 422ns)

### Troubleshooting - Phase 8
- **Timer pool not helping?** Check pool warmup and hit rate
- **Reflection cache misses?** Verify struct types are stable
- **Metrics not appearing?** Check Prometheus scraper config
- **Benchmarks flaky?** Use `-count=10` for statistical significance
- **Profiling crashes?** Ensure pprof endpoint only on localhost
