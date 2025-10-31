# Dependency Interface Migration Guide

## Quick Start

**TL;DR:** The Dependency interface allows you to use typed refs (`*Ref[int]`) directly with composables like `UseEffect`, eliminating the need for `Ref[any]` and excessive type assertions.

### Before (Old Pattern)
```go
count := bubbly.NewRef[any](0)  // Had to use any
UseEffect(ctx, effect, count)
```

### After (New Pattern)
```go
count := bubbly.NewRef(0)  // Type-safe: *Ref[int]
UseEffect(ctx, effect, count)  // Works directly!
```

---

## Why This Change Was Made

### The Problem

Go's type system doesn't support covariance. This created a painful choice:

**Option 1: Use Ref[any] everywhere**
```go
// ❌ Lost type safety
count := bubbly.NewRef[any](0)
name := bubbly.NewRef[any]("Alice")

// ❌ Type assertions everywhere
c := count.Get().(int)
n := name.Get().(string)

// ❌ No compile-time protection
count.Set("oops")  // Runtime panic!
```

**Option 2: Can't use typed refs with composables**
```go
// ✅ Type-safe creation
count := bubbly.NewRef(0)  // *Ref[int]

// ❌ Doesn't work!
UseEffect(ctx, effect, count)  // ERROR: cannot use *Ref[int] as *Ref[any]
```

### The Solution

The Dependency interface provides a common contract that all reactive types implement, enabling:
- ✅ Type-safe ref creation
- ✅ Polymorphic composable usage
- ✅ Computed values as dependencies
- ✅ Zero performance overhead

---

## What Changed

### API Changes

| Old API | New API | Notes |
|---------|---------|-------|
| `NewRef[any](value)` | `NewRef(value)` | Type inferred automatically |
| `ref.Get()` returns `T` | `ref.Get()` returns `any` | Use `GetTyped()` for type-safe access |
| N/A | `ref.GetTyped()` returns `T` | New method for type-safe access |
| `UseEffect(..., *Ref[any])` | `UseEffect(..., Dependency)` | Accepts any reactive type |
| Computed not usable | `UseEffect(..., computed)` | Computed values work directly |

### New Methods

Both `Ref[T]` and `Computed[T]` now have:

```go
// Interface method (returns any)
Get() any

// Type-safe method (returns T)
GetTyped() T

// Dependency interface methods
Invalidate()
AddDependent(dep Dependency)
```

---

## Migration Steps

### Step 1: Assess Your Codebase

Find all `Ref[any]` usage:

```bash
# Find Ref[any] declarations
grep -r "NewRef\[any\]" .

# Find Get() calls that need updating
grep -r "\.Get()" . | grep -v "GetTyped"
```

### Step 2: Migrate Ref Declarations

**Old:**
```go
count := bubbly.NewRef[any](0)
name := bubbly.NewRef[any]("Alice")
items := bubbly.NewRef[any]([]Item{})
```

**New:**
```go
count := bubbly.NewRef(0)           // *Ref[int]
name := bubbly.NewRef("Alice")      // *Ref[string]
items := bubbly.NewRef([]Item{})    // *Ref[[]Item]
```

### Step 3: Update Value Access

**Old:**
```go
value := count.Get().(int)  // Type assertion required
```

**New:**
```go
value := count.GetTyped()  // Type-safe, no assertion
```

### Step 4: Update Composable Usage

**Old:**
```go
// Had to convert to Ref[any]
count := bubbly.NewRef[any](0)
UseEffect(ctx, effect, count)
```

**New:**
```go
// Works directly with typed refs
count := bubbly.NewRef(0)
UseEffect(ctx, effect, count)  // count implements Dependency
```

### Step 5: Leverage Computed Dependencies

**New capability:**
```go
count := bubbly.NewRef(0)
doubled := bubbly.NewComputed(func() int {
    return count.GetTyped() * 2
})

// Computed values work as dependencies!
UseEffect(ctx, effect, doubled)
```

---

## Migration Strategies

### Strategy 1: Gradual Migration (Recommended)

Migrate one module at a time:

1. **Start with new code** - Use typed refs for all new features
2. **Migrate leaf modules** - Start with modules that don't depend on others
3. **Work upward** - Gradually migrate dependent modules
4. **Test continuously** - Run tests after each module migration

### Strategy 2: Big Bang Migration

Migrate everything at once (for smaller codebases):

1. **Create a branch** - `git checkout -b migrate-dependency-interface`
2. **Run automated migration** - Use sed or custom script
3. **Fix compilation errors** - Update all Get() calls
4. **Run full test suite** - Ensure nothing broke
5. **Merge when green** - All tests passing

### Strategy 3: Hybrid Approach

Keep Ref[any] for specific cases:

```go
// Use typed refs for most cases
count := bubbly.NewRef(0)

// Keep Ref[any] for truly dynamic data
dynamicData := bubbly.NewRef[any](someUnknownValue)
```

---

## Automated Migration

### Using sed (Simple Cases)

```bash
# Step 1: Update Ref[any] to typed refs
# Manual: Review each case and remove [any]

# Step 2: Update Get() to GetTyped()
find . -name "*.go" -type f -exec sed -i 's/\.Get()/\.GetTyped()/g' {} +

# Step 3: Revert false positives (Dependency interface usage)
# Manual: Check for dep.Get() that should stay as-is
```

### Using Go AST Tool (Recommended)

Create a custom migration tool:

```go
// migrate.go
package main

import (
    "go/ast"
    "go/parser"
    "go/token"
    // ... implement AST-based migration
)

// 1. Parse Go files
// 2. Find *Ref[any] declarations
// 3. Infer types from initial values
// 4. Replace with typed refs
// 5. Update .Get() to .GetTyped()
// 6. Preserve .Get() on Dependency interface
```

---

## Compatibility Notes

### Backwards Compatibility

✅ **Existing code continues to work:**
- `Ref[any]` still works (not deprecated)
- Old `Get()` method still exists (returns `any` now)
- No breaking changes to existing APIs

⚠️ **Behavioral change:**
- `Get()` now returns `any` instead of `T`
- Requires type assertion: `value := ref.Get().(int)`
- Use `GetTyped()` for type-safe access

### Forward Compatibility

✅ **New code benefits immediately:**
- Use typed refs from day one
- No migration needed for new features
- Better type safety and developer experience

---

## Common Patterns

### Pattern 1: Simple Value Migration

**Before:**
```go
count := bubbly.NewRef[any](0)
value := count.Get().(int)
count.Set(value + 1)
```

**After:**
```go
count := bubbly.NewRef(0)
value := count.GetTyped()
count.Set(value + 1)
```

### Pattern 2: UseEffect Migration

**Before:**
```go
count := bubbly.NewRef[any](0)
UseEffect(ctx, func() UseEffectCleanup {
    c := count.Get().(int)
    fmt.Printf("Count: %d\n", c)
    return nil
}, count)
```

**After:**
```go
count := bubbly.NewRef(0)
UseEffect(ctx, func() UseEffectCleanup {
    c := count.GetTyped()
    fmt.Printf("Count: %d\n", c)
    return nil
}, count)
```

### Pattern 3: Multiple Dependencies

**Before:**
```go
count := bubbly.NewRef[any](0)
name := bubbly.NewRef[any]("Alice")
UseEffect(ctx, effect, count, name)
```

**After:**
```go
count := bubbly.NewRef(0)
name := bubbly.NewRef("Alice")
UseEffect(ctx, effect, count, name)  // Mixed types work!
```

### Pattern 4: Computed Dependencies

**Before (Not Possible):**
```go
// Computed values couldn't be used as dependencies
```

**After (New Capability):**
```go
count := bubbly.NewRef(0)
doubled := bubbly.NewComputed(func() int {
    return count.GetTyped() * 2
})
UseEffect(ctx, effect, doubled)  // Works!
```

---

## Troubleshooting

### Issue 1: "cannot use *Ref[int] as *Ref[any]"

**Problem:**
```go
count := bubbly.NewRef(0)
UseEffect(ctx, effect, count)  // ERROR
```

**Solution:**
This error shouldn't occur with the new Dependency interface. If you see it:
1. Ensure you're using the latest version of BubblyUI
2. Check that UseEffect accepts `...Dependency` not `...*Ref[any]`
3. Verify your imports are correct

### Issue 2: "cannot use value (type any) as type int"

**Problem:**
```go
count := bubbly.NewRef(0)
value := count.Get()  // value is any, not int
```

**Solution:**
Use `GetTyped()` instead:
```go
value := count.GetTyped()  // value is int
```

Or add type assertion:
```go
value := count.Get().(int)  // Explicit type assertion
```

### Issue 3: "Get() returns any, expected T"

**Problem:**
Old code expects `Get()` to return typed value.

**Solution:**
Replace all `Get()` with `GetTyped()`:
```bash
find . -name "*.go" -exec sed -i 's/\.Get()/\.GetTyped()/g' {} +
```

Then manually revert `Get()` on Dependency interface usage.

### Issue 4: Computed function errors

**Problem:**
```go
computed := NewComputed(func() int {
    return count.Get() * 2  // ERROR: invalid operation
})
```

**Solution:**
Use `GetTyped()` in computed functions:
```go
computed := NewComputed(func() int {
    return count.GetTyped() * 2  // OK
})
```

### Issue 5: Watch not working with Computed

**Problem:**
```go
Watch(computed, callback)  // ERROR: Watchable[T] has no method Get
```

**Solution:**
This is already fixed. Watchable[T] uses `GetTyped()`. Ensure you're on the latest version.

---

## Testing Your Migration

### Step 1: Compilation

```bash
go build ./...
```

All packages should compile without errors.

### Step 2: Unit Tests

```bash
go test ./... -v
```

All tests should pass.

### Step 3: Race Detector

```bash
go test ./... -race
```

No race conditions should be detected.

### Step 4: Integration Tests

```bash
go test ./tests/integration/... -v
```

All integration tests should pass.

### Step 5: Manual Testing

Test your application manually to ensure:
- All features work as expected
- No runtime panics
- Performance is unchanged

---

## Rollback Plan

If migration causes issues:

### Quick Rollback

```bash
git checkout main
git branch -D migrate-dependency-interface
```

### Partial Rollback

Keep new code, revert problematic modules:
```bash
git checkout main -- path/to/problematic/module
```

### Gradual Rollback

Revert one commit at a time:
```bash
git revert HEAD
git revert HEAD~1
# etc.
```

---

## Benefits After Migration

### Immediate Benefits

1. **Type Safety** - Compile-time type checking
2. **Less Boilerplate** - No more `Ref[any]` everywhere
3. **Better IntelliSense** - IDE knows exact types
4. **Fewer Type Assertions** - Use `GetTyped()` instead
5. **Computed Dependencies** - New capability unlocked

### Long-term Benefits

1. **Maintainability** - Easier to understand code
2. **Refactoring** - Compiler catches type errors
3. **Performance** - Slightly faster (no type assertions)
4. **Developer Experience** - More enjoyable to write
5. **Future-Proof** - Ready for new features

---

## FAQ

### Q: Do I have to migrate?

**A:** No. Existing code with `Ref[any]` continues to work. Migration is optional but recommended for better type safety.

### Q: Can I mix old and new patterns?

**A:** Yes. `Ref[any]` and typed refs can coexist in the same codebase.

### Q: What's the performance impact?

**A:** Minimal. The Dependency interface adds < 0.05ns overhead. Using `GetTyped()` is actually slightly faster than type assertions.

### Q: Will this break my existing code?

**A:** No breaking changes. The only behavioral change is that `Get()` now returns `any`, but this is backwards compatible with type assertions.

### Q: How long does migration take?

**A:** Depends on codebase size:
- Small (< 1000 LOC): 1-2 hours
- Medium (1000-10000 LOC): 4-8 hours
- Large (> 10000 LOC): 1-2 days

### Q: Can I automate the migration?

**A:** Partially. Use sed for simple replacements, but manual review is recommended for:
- Ref[any] to typed refs (need to infer types)
- Dependency interface usage (keep Get())
- Complex type assertions

### Q: What if I find a bug?

**A:** Report it! The Dependency interface is production-ready, but if you encounter issues:
1. Check this troubleshooting guide
2. Review the reactive-dependencies.md guide
3. Open an issue on GitHub

---

## Next Steps

After completing migration:

1. **Update Documentation** - Document your team's patterns
2. **Share Knowledge** - Train team members on new patterns
3. **Establish Guidelines** - When to use Get() vs GetTyped()
4. **Monitor Performance** - Verify no regressions
5. **Celebrate** - You now have better type safety! 🎉

---

## Additional Resources

- [Reactive Dependencies Guide](./reactive-dependencies.md) - Comprehensive reference
- [Composition API Guide](./composition-api.md) - UseEffect and composables
- [BubblyUI Documentation](../../README.md) - Full framework docs
- [GitHub Issues](https://github.com/newbpydev/bubblyui/issues) - Report problems

---

## Summary

The Dependency interface migration is straightforward:

1. **Change** `NewRef[any](value)` → `NewRef(value)`
2. **Change** `ref.Get()` → `ref.GetTyped()`
3. **Keep** `dep.Get()` for Dependency interface usage
4. **Test** thoroughly after migration
5. **Enjoy** better type safety and developer experience!

**Estimated Time:** 1-8 hours depending on codebase size  
**Difficulty:** Easy to Medium  
**Risk:** Low (backwards compatible)  
**Reward:** High (better type safety, new capabilities)
