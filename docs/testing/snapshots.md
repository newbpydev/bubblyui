# Snapshot Testing Guide

## Overview

Snapshot testing captures rendered output and compares it against saved snapshots to detect unintended changes. This guide covers creating, updating, and managing snapshots effectively.

## Table of Contents

- [What is Snapshot Testing?](#what-is-snapshot-testing)
- [Creating Snapshots](#creating-snapshots)
- [Updating Snapshots](#updating-snapshots)
- [Best Practices](#best-practices)
- [Common Patterns](#common-patterns)
- [Troubleshooting](#troubleshooting)

## What is Snapshot Testing?

Snapshot testing helps you:

- **Detect regressions** - Catch unintended UI changes
- **Document output** - Snapshots serve as documentation
- **Review changes** - See exactly what changed
- **Prevent bugs** - Catch rendering issues early

### When to Use Snapshots

✅ **Good use cases:**
- Testing rendered output
- Detecting layout changes
- Verifying component structure
- Regression testing

❌ **Avoid for:**
- Dynamic content (timestamps, random data)
- Frequently changing output
- Testing logic (use assertions instead)

## Creating Snapshots

### Basic Snapshot

```go
func TestCounterRender(t *testing.T) {
    harness := testutil.NewHarness(t)
    counter := harness.Mount(createCounter())
    
    // Take snapshot
    output := counter.Component().View()
    testutil.MatchSnapshot(t, "counter_initial", output)
}
```

**First run creates snapshot:**
```
=== RUN   TestCounterRender
    testutil.go:123: Created snapshot: __snapshots__/TestCounterRender_counter_initial.snap
--- PASS: TestCounterRender (0.00s)
```

**Snapshot file (`__snapshots__/TestCounterRender_counter_initial.snap`):**
```
Counter: 0
[+] Increment
[-] Decrement
```

### Multiple Snapshots

Test different states:

```go
func TestCounterStates(t *testing.T) {
    harness := testutil.NewHarness(t)
    counter := harness.Mount(createCounter())
    
    // Initial state
    output := counter.Component().View()
    testutil.MatchSnapshot(t, "counter_initial", output)
    
    // After increment
    counter.Component().Emit("increment", nil)
    output = counter.Component().View()
    testutil.MatchSnapshot(t, "counter_after_increment", output)
    
    // After decrement
    counter.Component().Emit("decrement", nil)
    output = counter.Component().View()
    testutil.MatchSnapshot(t, "counter_after_decrement", output)
}
```

## Updating Snapshots

### When Output Changes

**Intentional change:**
```bash
go test -update
```

**Output:**
```
=== RUN   TestCounterRender
    testutil.go:158: Updated snapshot: __snapshots__/TestCounterRender_counter_initial.snap
--- PASS: TestCounterRender (0.00s)
```

### Reviewing Changes

Before updating, review the diff:

```bash
git diff __snapshots__/
```

**Example diff:**
```diff
- Counter: 0
+ Count: 0
  [+] Increment
  [-] Decrement
```

### Selective Updates

Update specific tests:

```bash
go test -run TestCounterRender -update
```

## Best Practices

### 1. Use Descriptive Names

```go
// ✅ Good: Descriptive names
testutil.MatchSnapshot(t, "counter_initial_state", output)
testutil.MatchSnapshot(t, "counter_after_increment", output)
testutil.MatchSnapshot(t, "form_validation_error", output)

// ❌ Bad: Generic names
testutil.MatchSnapshot(t, "snapshot1", output)
testutil.MatchSnapshot(t, "test", output)
```

### 2. Test Stable Output

```go
// ✅ Good: Stable output
output := counter.Component().View()
testutil.MatchSnapshot(t, "counter", output)

// ❌ Bad: Dynamic content
output := fmt.Sprintf("Time: %s", time.Now()) // Changes every run
testutil.MatchSnapshot(t, "with_time", output)
```

### 3. Keep Snapshots Small

```go
// ✅ Good: Test specific parts
header := component.RenderHeader()
testutil.MatchSnapshot(t, "header", header)

footer := component.RenderFooter()
testutil.MatchSnapshot(t, "footer", footer)

// ❌ Bad: Entire screen
output := component.View() // Too large
testutil.MatchSnapshot(t, "full_screen", output)
```

### 4. Normalize Dynamic Content

```go
func normalizeOutput(output string) string {
    // Remove timestamps
    re := regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`)
    output = re.ReplaceAllString(output, "TIMESTAMP")
    
    // Remove IDs
    re = regexp.MustCompile(`id=\d+`)
    output = re.ReplaceAllString(output, "id=ID")
    
    return output
}

func TestWithNormalization(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createComponent())
    
    output := component.Component().View()
    normalized := normalizeOutput(output)
    
    testutil.MatchSnapshot(t, "normalized", normalized)
}
```

### 5. Review Before Committing

Always review snapshot changes:

```bash
# View changes
git diff __snapshots__/

# Review carefully
# Commit only if changes are intentional
git add __snapshots__/
git commit -m "Update snapshots for new layout"
```

## Common Patterns

### Pattern 1: State-Based Snapshots

Test different component states:

```go
func TestFormStates(t *testing.T) {
    harness := testutil.NewHarness(t)
    form := harness.Mount(createForm())
    
    states := []struct {
        name   string
        setup  func()
    }{
        {
            name: "empty",
            setup: func() {},
        },
        {
            name: "filled",
            setup: func() {
                form.State().SetRefValue("name", "John")
                form.State().SetRefValue("email", "john@example.com")
            },
        },
        {
            name: "validation_error",
            setup: func() {
                form.State().SetRefValue("email", "invalid")
                form.Component().Emit("validate", nil)
            },
        },
    }
    
    for _, state := range states {
        t.Run(state.name, func(t *testing.T) {
            state.setup()
            output := form.Component().View()
            testutil.MatchSnapshot(t, "form_"+state.name, output)
        })
    }
}
```

### Pattern 2: Component Variants

Test different component configurations:

```go
func TestButtonVariants(t *testing.T) {
    variants := []struct {
        name  string
        props ButtonProps
    }{
        {"primary", ButtonProps{Variant: "primary", Label: "Click"}},
        {"secondary", ButtonProps{Variant: "secondary", Label: "Click"}},
        {"disabled", ButtonProps{Variant: "primary", Label: "Click", Disabled: true}},
        {"loading", ButtonProps{Variant: "primary", Label: "Click", Loading: true}},
    }
    
    for _, variant := range variants {
        t.Run(variant.name, func(t *testing.T) {
            harness := testutil.NewHarness(t)
            button := harness.Mount(createButton(variant.props))
            
            output := button.Component().View()
            testutil.MatchSnapshot(t, "button_"+variant.name, output)
        })
    }
}
```

### Pattern 3: Responsive Layouts

Test different terminal sizes:

```go
func TestResponsiveLayout(t *testing.T) {
    sizes := []struct {
        name   string
        width  int
        height int
    }{
        {"small", 40, 20},
        {"medium", 80, 24},
        {"large", 120, 40},
    }
    
    for _, size := range sizes {
        t.Run(size.name, func(t *testing.T) {
            harness := testutil.NewHarness(t)
            component := harness.Mount(createResponsiveComponent())
            
            // Set terminal size
            component.Component().Emit("resize", map[string]int{
                "width":  size.width,
                "height": size.height,
            })
            
            output := component.Component().View()
            testutil.MatchSnapshot(t, "layout_"+size.name, output)
        })
    }
}
```

### Pattern 4: Error States

Test error rendering:

```go
func TestErrorStates(t *testing.T) {
    errors := []struct {
        name  string
        error error
    }{
        {"network_error", errors.New("network timeout")},
        {"validation_error", errors.New("invalid email")},
        {"not_found", errors.New("user not found")},
    }
    
    for _, errCase := range errors {
        t.Run(errCase.name, func(t *testing.T) {
            harness := testutil.NewHarness(t)
            component := harness.Mount(createComponent())
            
            component.State().SetRefValue("error", errCase.error)
            
            output := component.Component().View()
            testutil.MatchSnapshot(t, "error_"+errCase.name, output)
        })
    }
}
```

## Troubleshooting

### Snapshot Mismatch

**Problem**: Test fails with snapshot mismatch

**Output:**
```
=== RUN   TestCounterRender
    testutil.go:145: Snapshot mismatch for "counter_initial":
    
    - Counter: 0
    + Count: 0
    
    Run with -update flag to update snapshots
--- FAIL: TestCounterRender (0.00s)
```

**Solutions:**

1. **Intentional change** - Update snapshot:
```bash
go test -update
```

2. **Unintentional change** - Fix code:
```go
// Revert the change that broke the snapshot
```

3. **Review diff**:
```bash
git diff __snapshots__/
```

### Flaky Snapshots

**Problem**: Snapshots fail intermittently

**Causes:**
- Timestamps in output
- Random data
- Concurrent operations
- Non-deterministic rendering

**Solutions:**

1. **Normalize output:**
```go
func normalizeOutput(output string) string {
    // Remove timestamps
    output = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`).
        ReplaceAllString(output, "DATE")
    
    // Remove random IDs
    output = regexp.MustCompile(`id-\w+`).
        ReplaceAllString(output, "id-RANDOM")
    
    return output
}
```

2. **Use fixed values in tests:**
```go
// ✅ Good: Fixed time
fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
component := createComponentWithTime(fixedTime)

// ❌ Bad: Current time
component := createComponentWithTime(time.Now())
```

3. **Avoid async operations:**
```go
// ✅ Good: Synchronous
output := component.View()
testutil.MatchSnapshot(t, "sync", output)

// ❌ Bad: Async (timing dependent)
go updateComponent()
time.Sleep(100 * time.Millisecond)
output := component.View()
```

### Large Snapshots

**Problem**: Snapshot files are too large

**Solutions:**

1. **Test specific parts:**
```go
// Instead of full screen
header := component.RenderHeader()
testutil.MatchSnapshot(t, "header", header)
```

2. **Use multiple small snapshots:**
```go
testutil.MatchSnapshot(t, "header", component.RenderHeader())
testutil.MatchSnapshot(t, "body", component.RenderBody())
testutil.MatchSnapshot(t, "footer", component.RenderFooter())
```

3. **Use assertions for data:**
```go
// Use snapshots for layout
testutil.MatchSnapshot(t, "layout", component.View())

// Use assertions for data
data := component.State().GetRef("data")
assert.Equal(t, expectedData, data.Get())
```

## Snapshot File Organization

### Directory Structure

```
project/
├── component.go
├── component_test.go
└── __snapshots__/
    ├── TestComponent_initial.snap
    ├── TestComponent_after_update.snap
    └── TestComponent_error_state.snap
```

### Naming Convention

```
Test{TestName}_{snapshot_name}.snap
```

**Examples:**
- `TestCounter_initial.snap`
- `TestForm_validation_error.snap`
- `TestButton_primary_variant.snap`

## Summary

Snapshot testing provides:

- ✅ **Regression detection** - Catch unintended changes
- ✅ **Visual documentation** - Snapshots document output
- ✅ **Easy updates** - Single command to update
- ✅ **Git-friendly** - Track changes in version control

Best practices:
- Use descriptive names
- Test stable output
- Keep snapshots small
- Normalize dynamic content
- Review before committing

## See Also

- **[Quickstart](quickstart.md)** - Get started quickly
- **[Assertions](assertions.md)** - Assertion reference
- **[Mocking](mocking.md)** - Isolation patterns
- **[Testing Guide](../guides/testing-guide.md)** - Comprehensive guide
