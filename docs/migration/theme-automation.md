# Migration Guide: Advanced Internal Package Automation

**Version:** 1.0  
**Feature:** 13-adv-internal-package-automation  
**Applies to:** BubblyUI v0.x → v1.0+

This guide helps you migrate existing BubblyUI applications to use the new automation patterns that reduce boilerplate by 60-94%.

---

## Table of Contents

1. [Overview](#overview)
2. [Theme System Migration](#theme-system-migration)
3. [Multi-Key Binding Migration](#multi-key-binding-migration)
4. [Shared Composables Migration](#shared-composables-migration)
5. [Step-by-Step Migration Process](#step-by-step-migration-process)
6. [Common Pitfalls](#common-pitfalls)
7. [Backward Compatibility](#backward-compatibility)
8. [FAQ](#faq)

---

## Overview

### What's New

| Pattern | Purpose | Code Reduction |
|---------|---------|----------------|
| `UseTheme/ProvideTheme` | Consistent theming across components | **94%** (15→1 lines) |
| `WithMultiKeyBindings` | Multiple keys for same action | **67%** (6→2 lines) |
| `CreateShared` | Singleton composables across components | Enables new patterns |

### Migration Priority

1. **High**: Theme System (biggest code reduction, most common pattern)
2. **Medium**: Multi-Key Bindings (cleaner builder pattern)
3. **Optional**: Shared Composables (architectural pattern for specific use cases)

---

## Theme System Migration

### Identifying Code to Migrate

Search your codebase for these patterns:

```bash
# Find inject/expose color patterns
grep -r "ctx.Inject.*Color" --include="*.go"
grep -r "ctx.Provide.*Color" --include="*.go"
grep -r "lipgloss.Color" --include="*.go" | grep -i "primary\|secondary\|muted"
```

### Parent Component Migration

**BEFORE (5+ lines):**
```go
func CreateApp() (bubbly.Component, error) {
    return bubbly.NewComponent("App").
        Setup(func(ctx *bubbly.Context) {
            // Manual color provision - VERBOSE
            ctx.Provide("primaryColor", lipgloss.Color("35"))
            ctx.Provide("secondaryColor", lipgloss.Color("99"))
            ctx.Provide("mutedColor", lipgloss.Color("240"))
            ctx.Provide("warningColor", lipgloss.Color("220"))
            ctx.Provide("errorColor", lipgloss.Color("196"))
            ctx.Provide("successColor", lipgloss.Color("35"))
            ctx.Provide("backgroundColor", lipgloss.Color("236"))
            
            // ... rest of setup
        }).
        Build()
}
```

**AFTER (1 line):**
```go
func CreateApp() (bubbly.Component, error) {
    return bubbly.NewComponent("App").
        Setup(func(ctx *bubbly.Context) {
            // ONE LINE replaces 7 Provide calls!
            ctx.ProvideTheme(bubbly.DefaultTheme)
            
            // ... rest of setup
        }).
        Build()
}
```

**With Customization:**
```go
func CreateApp() (bubbly.Component, error) {
    return bubbly.NewComponent("App").
        Setup(func(ctx *bubbly.Context) {
            // Customize theme
            theme := bubbly.DefaultTheme
            theme.Primary = lipgloss.Color("99")    // Purple brand
            theme.Secondary = lipgloss.Color("120") // Custom accent
            ctx.ProvideTheme(theme)
            
            // ... rest of setup
        }).
        Build()
}
```

### Child Component Migration

**BEFORE (15+ lines per component):**
```go
func CreateCard(props CardProps) (bubbly.Component, error) {
    return bubbly.NewComponent("Card").
        Setup(func(ctx *bubbly.Context) {
            // Manual inject with defaults - VERBOSE
            primaryColor := lipgloss.Color("35")
            if injected := ctx.Inject("primaryColor", nil); injected != nil {
                primaryColor = injected.(lipgloss.Color)
            }
            
            secondaryColor := lipgloss.Color("99")
            if injected := ctx.Inject("secondaryColor", nil); injected != nil {
                secondaryColor = injected.(lipgloss.Color)
            }
            
            mutedColor := lipgloss.Color("240")
            if injected := ctx.Inject("mutedColor", nil); injected != nil {
                mutedColor = injected.(lipgloss.Color)
            }
            
            // Expose for template
            ctx.Expose("primaryColor", primaryColor)
            ctx.Expose("secondaryColor", secondaryColor)
            ctx.Expose("mutedColor", mutedColor)
        }).
        Template(func(ctx *bubbly.RenderContext) string {
            primary := ctx.Get("primaryColor").(lipgloss.Color)
            secondary := ctx.Get("secondaryColor").(lipgloss.Color)
            // ...
        }).
        Build()
}
```

**AFTER (2 lines):**
```go
func CreateCard(props CardProps) (bubbly.Component, error) {
    return bubbly.NewComponent("Card").
        Setup(func(ctx *bubbly.Context) {
            // ONE LINE replaces 15+ lines!
            theme := ctx.UseTheme(bubbly.DefaultTheme)
            ctx.Expose("theme", theme)
        }).
        Template(func(ctx *bubbly.RenderContext) string {
            theme := ctx.Get("theme").(bubbly.Theme)
            
            // Type-safe color access
            titleStyle := lipgloss.NewStyle().
                Foreground(theme.Primary).
                Bold(true)
            
            borderStyle := lipgloss.NewStyle().
                Border(lipgloss.RoundedBorder()).
                BorderForeground(theme.Secondary)
            
            // ...
        }).
        Build()
}
```

### Template Updates

**BEFORE:**
```go
Template(func(ctx *bubbly.RenderContext) string {
    primary := ctx.Get("primaryColor").(lipgloss.Color)
    secondary := ctx.Get("secondaryColor").(lipgloss.Color)
    muted := ctx.Get("mutedColor").(lipgloss.Color)
    
    style := lipgloss.NewStyle().Foreground(primary)
    // ...
})
```

**AFTER:**
```go
Template(func(ctx *bubbly.RenderContext) string {
    theme := ctx.Get("theme").(bubbly.Theme)
    
    style := lipgloss.NewStyle().Foreground(theme.Primary)
    // Access any color: theme.Primary, theme.Secondary, theme.Muted, etc.
})
```

---

## Multi-Key Binding Migration

### Identifying Code to Migrate

Search for repeated `WithKeyBinding` calls with the same event:

```bash
# Find repeated key bindings
grep -r "WithKeyBinding" --include="*.go" | grep -E "(increment|decrement|submit|toggle)"
```

### Migration Example

**BEFORE (6 lines):**
```go
func CreateCounter() (bubbly.Component, error) {
    return bubbly.NewComponent("Counter").
        WithKeyBinding("up", "increment", "Increment counter").
        WithKeyBinding("k", "increment", "Increment counter").
        WithKeyBinding("+", "increment", "Increment counter").
        WithKeyBinding("down", "decrement", "Decrement counter").
        WithKeyBinding("j", "decrement", "Decrement counter").
        WithKeyBinding("-", "decrement", "Decrement counter").
        Setup(func(ctx *bubbly.Context) {
            // ...
        }).
        Build()
}
```

**AFTER (2 lines):**
```go
func CreateCounter() (bubbly.Component, error) {
    return bubbly.NewComponent("Counter").
        WithMultiKeyBindings("increment", "Increment counter", "up", "k", "+").
        WithMultiKeyBindings("decrement", "Decrement counter", "down", "j", "-").
        Setup(func(ctx *bubbly.Context) {
            // ...
        }).
        Build()
}
```

### Parameter Order

Note the parameter order difference:

| Method | Order |
|--------|-------|
| `WithKeyBinding` | `(key, event, description)` |
| `WithMultiKeyBindings` | `(event, description, keys...)` |

The new method puts event first because it's the grouping concept.

---

## Shared Composables Migration

### When to Use

Use `CreateShared` when:
- Multiple components need the **same state instance**
- You want to avoid prop drilling
- State should be synchronized across components

**Do NOT use** when:
- Each component should have its own state
- State is component-specific

### Migration Example

**BEFORE (separate instances):**
```go
// Component A - has its own counter
func CreateDisplay() (bubbly.Component, error) {
    return bubbly.NewComponent("Display").
        Setup(func(ctx *bubbly.Context) {
            counter := composables.UseCounter(ctx, 0) // Instance 1
            ctx.Expose("counter", counter)
        }).
        Build()
}

// Component B - has DIFFERENT counter
func CreateControls() (bubbly.Component, error) {
    return bubbly.NewComponent("Controls").
        Setup(func(ctx *bubbly.Context) {
            counter := composables.UseCounter(ctx, 0) // Instance 2 (different!)
            // Incrementing here does NOT affect Display!
        }).
        Build()
}
```

**AFTER (shared instance):**
```go
// Define once at package level
var UseSharedCounter = composables.CreateShared(
    func(ctx *bubbly.Context) *composables.CounterComposable {
        return composables.UseCounter(ctx, 0)
    },
)

// Component A
func CreateDisplay() (bubbly.Component, error) {
    return bubbly.NewComponent("Display").
        Setup(func(ctx *bubbly.Context) {
            counter := UseSharedCounter(ctx) // Shared instance
            ctx.Expose("counter", counter)
        }).
        Build()
}

// Component B - SAME counter!
func CreateControls() (bubbly.Component, error) {
    return bubbly.NewComponent("Controls").
        Setup(func(ctx *bubbly.Context) {
            counter := UseSharedCounter(ctx) // Same instance!
            // Incrementing here DOES affect Display!
        }).
        Build()
}
```

---

## Step-by-Step Migration Process

### Phase 1: Preparation (5 minutes)

1. **Ensure tests pass:**
   ```bash
   make test-race
   ```

2. **Create a migration branch:**
   ```bash
   git checkout -b feature/automation-migration
   ```

### Phase 2: Theme Migration (30-60 minutes)

1. **Find all color inject/expose patterns:**
   ```bash
   grep -rn "ctx.Inject.*Color\|ctx.Provide.*Color" --include="*.go"
   ```

2. **Update parent components first:**
   - Replace multiple `ctx.Provide("*Color", ...)` with `ctx.ProvideTheme(theme)`
   - Customize theme if needed

3. **Update child components:**
   - Replace inject/expose blocks with `ctx.UseTheme(bubbly.DefaultTheme)`
   - Update template to use `theme.Primary`, `theme.Secondary`, etc.

4. **Run tests after each file:**
   ```bash
   go test -race ./...
   ```

### Phase 3: Key Binding Migration (15-30 minutes)

1. **Find repeated key bindings:**
   ```bash
   grep -rn "WithKeyBinding" --include="*.go" | sort
   ```

2. **Group by event and consolidate:**
   - Find all keys that emit the same event
   - Replace with single `WithMultiKeyBindings` call

3. **Verify key bindings work:**
   ```bash
   go run ./cmd/examples/your-app
   ```

### Phase 4: Shared Composables (Optional, 15-30 minutes)

1. **Identify shared state needs:**
   - Look for composables used in multiple components
   - Determine if state should be shared or isolated

2. **Create shared versions:**
   - Define `var UseSharedX = composables.CreateShared(...)` at package level
   - Update components to use shared version

3. **Test state synchronization:**
   - Verify changes in one component affect others

### Phase 5: Verification (10 minutes)

1. **Run full test suite:**
   ```bash
   make test-race
   ```

2. **Run linter:**
   ```bash
   make lint
   ```

3. **Test examples manually:**
   ```bash
   cd cmd/examples/your-app && go run .
   ```

4. **Commit and push:**
   ```bash
   git add -A
   git commit -m "feat: migrate to automation patterns (UseTheme, WithMultiKeyBindings)"
   git push origin feature/automation-migration
   ```

---

## Common Pitfalls

### Pitfall 1: Forgetting to Provide Theme

**Problem:** Child uses `UseTheme` but parent never calls `ProvideTheme`.

**Symptom:** Component uses default colors instead of custom theme.

**Solution:** This is actually **graceful degradation** - not an error. If you want custom colors, add `ctx.ProvideTheme(theme)` to a parent component.

```go
// Parent MUST provide theme for children to inherit custom colors
ctx.ProvideTheme(customTheme)
```

### Pitfall 2: Wrong Type Assertion in Template

**Problem:** Using old type assertion pattern with new theme.

**WRONG:**
```go
primary := ctx.Get("theme").(lipgloss.Color) // WRONG type!
```

**RIGHT:**
```go
theme := ctx.Get("theme").(bubbly.Theme)
primary := theme.Primary // Access color from struct
```

### Pitfall 3: Parameter Order in WithMultiKeyBindings

**Problem:** Using `WithKeyBinding` parameter order.

**WRONG:**
```go
.WithMultiKeyBindings("up", "increment", "Increment") // Wrong order!
```

**RIGHT:**
```go
.WithMultiKeyBindings("increment", "Increment", "up", "k", "+") // event, desc, keys...
```

### Pitfall 4: Shared Composable Context

**Problem:** Relying on component-specific context in shared composable factory.

**WRONG:**
```go
var UseSharedData = composables.CreateShared(
    func(ctx *bubbly.Context) *DataComposable {
        // DON'T rely on component-specific context!
        userID := ctx.Get("userID").(string) // May be nil!
        return NewDataComposable(userID)
    },
)
```

**RIGHT:**
```go
var UseSharedData = composables.CreateShared(
    func(ctx *bubbly.Context) *DataComposable {
        // Use only context-independent initialization
        return NewDataComposable()
    },
)
```

### Pitfall 5: Mixing Old and New Patterns

**Problem:** Using both old inject/expose AND UseTheme in same component.

**Solution:** Pick one pattern per component. Prefer UseTheme for new code.

---

## Backward Compatibility

### Old Code Still Works

All old patterns continue to work:

```go
// This still works - no migration required
ctx.Provide("primaryColor", lipgloss.Color("35"))
ctx.Inject("primaryColor", nil)
```

### Gradual Migration

You can migrate incrementally:
- Some components use `UseTheme`
- Others use old `Inject/Provide`
- They can coexist in the same app

### No Breaking Changes

- `WithKeyBinding` still works alongside `WithMultiKeyBindings`
- `UseCounter` still creates separate instances (use `CreateShared` for singleton)
- All existing APIs unchanged

---

## FAQ

### Q: Do I have to migrate all at once?

**A:** No. Migration is optional and incremental. Old patterns work indefinitely.

### Q: What if I don't want to use DefaultTheme colors?

**A:** Customize the theme before providing:

```go
theme := bubbly.DefaultTheme
theme.Primary = lipgloss.Color("99")
ctx.ProvideTheme(theme)
```

### Q: Can I override theme in a subtree?

**A:** Yes. Any component can call `ProvideTheme` to override for its descendants:

```go
// Modal with darker background
modalTheme := ctx.UseTheme(bubbly.DefaultTheme)
modalTheme.Background = lipgloss.Color("232")
ctx.ProvideTheme(modalTheme) // Only affects modal's children
```

### Q: Is CreateShared thread-safe?

**A:** Yes. It uses `sync.Once` internally, safe for concurrent access.

### Q: What's the performance impact?

**A:** Negligible (<5% overhead). The automation patterns are thin wrappers around existing APIs.

### Q: Can I use WithMultiKeyBindings with conditional bindings?

**A:** No. For conditional bindings, use `WithConditionalKeyBinding` separately:

```go
.WithMultiKeyBindings("increment", "Inc", "up", "k", "+").
.WithConditionalKeyBinding(bubbly.KeyBinding{
    Key: " ", Event: "toggle", Description: "Toggle",
    Condition: func() bool { return !inputMode },
})
```

### Q: How do I test shared composables?

**A:** Use the testutil harness:

```go
func TestSharedCounter(t *testing.T) {
    harness := testutil.NewHarness(t)
    defer harness.Cleanup()
    
    display := harness.Mount(CreateDisplay())
    controls := harness.Mount(CreateControls())
    
    controls.SendEvent("increment", nil)
    display.AssertRenderContains("Count: 1")
}
```

---

## Summary

| Pattern | When to Use | Migration Effort |
|---------|-------------|------------------|
| `UseTheme/ProvideTheme` | Any app with consistent colors | 30-60 min |
| `WithMultiKeyBindings` | Multiple keys → same action | 15-30 min |
| `CreateShared` | Shared state across components | 15-30 min |

**Total estimated migration time:** 1-2 hours for a medium-sized application.

**Code reduction:** 170+ lines in typical applications.

---

*For questions or issues, see the [BubblyUI documentation](../README.md) or open an issue on GitHub.*
