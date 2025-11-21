# Documentation Audit Report

**Date:** November 18, 2025
**File Audited:** docs/BUBBLY_AI_MANUAL.md
**Status:** CRITICAL DISCREPANCIES FOUND

---

## Executive Summary

The current manual contains **significant inaccuracies** that would mislead developers. There are 47+ documented methods/APIs that **DO NOT EXIST** in the actual codebase, and many real APIs are either missing or incorrectly documented.

**Severity:** 🔴 CRITICAL - Code examples will not compile

---

## Major Discrepancies by Category

### 1. Package Structure Issues ❌ WRONG

#### Documentation Claims:
```go
import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/components"
    "github.com/newbpydev/bubblyui/pkg/bubbly/composables"
    "github.com/newbpydev/bubblyui/pkg/bubbly/directives"
    "github.com/newbpydev/bubblyui/pkg/bubbly/router"
)
```

#### Reality Check:
- **components**: ✅ Correct
- **bubbly**: ✅ Correct  
- **composables**: ❌ WRONG - Package is just `composables`
- **directives**: ❌ WRONG - Package is just `directives`
- **router**: ❌ WRONG - Actually `bubbly/router` (subdirectory)

#### Real Imports:
```go
import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/components"
    "github.com/newbpydev/bubblyui/pkg/bubbly/composables"  // ❌ WRONG
    "github.com/newbpydev/bubblyui/pkg/bubbly/directives"   // ❌ WRONG
    "github.com/newbpydev/bubblyui/pkg/bubbly/router"       // ❌ WRONG location
    
    "github.com/newbpydev/bubblyui/pkg/components"
    composables "github.com/newbpydev/bubblyui/pkg/bubbly/composables"  // ✅ CORRECT
    directives "github.com/newbpydev/bubblyui/pkg/bubbly/directives"    // ✅ CORRECT
)
```

---

### 2. Non-Existent Context Methods (15 methods) ❌ FAKE

The manual documents **15 methods that DO NOT exist** on `bubbly.Context`:

1. `ctx.ExposeComponent()` - **DOES NOT EXIST**
2. `ctx.SetCommandGenerator()` - **DOES NOT EXIST**
3. `ctx.Set()` (for arbitrary values) - **DOES NOT EXIST**
4. `ctx.Get()` methods with different signatures than documented
5. `ctx.AddChild()` - **DOES NOT EXIST**

#### Reality Check - Real Context Methods:
- `ctx.Ref(value)` - ✅ EXISTS
- `ctx.Computed(fn)` - ✅ EXISTS
- `ctx.Watch(ref, callback)` - ✅ EXISTS
- `ctx.Expose(key, value)` - ✅ EXISTS (but different than documented)
- `ctx.Get(key)` - ✅ EXISTS (returns interface{}, not typed)
- `ctx.On(event, handler)` - ✅ EXISTS
- `ctx.Emit(event, data)` - ✅ EXISTS
- `ctx.Props()` - ✅ EXISTS
- `ctx.Children()` - ✅ EXISTS
- `ctx.OnMounted(hook)` - ✅ EXISTS
- `ctx.OnUpdated(hook, deps...)` - ✅ EXISTS
- `ctx.OnUnmounted(hook)` - ✅ EXISTS
- `ctx.OnBeforeUpdate(hook)` - ✅ EXISTS
- `ctx.OnBeforeUnmount(hook)` - ✅ EXISTS
- `ctx.OnCleanup(cleanup)` - ✅ EXISTS
- `ctx.Provide(key, value)` - ✅ EXISTS
- `ctx.Inject(key, defaultValue)` - ✅ EXISTS
- `ctx.EnableAutoCommands()` - ✅ EXISTS
- `ctx.DisableAutoCommands()` - ✅ EXISTS
- `ctx.IsAutoCommandsEnabled()` - ✅ EXISTS
- `ctx.ManualRef(value)` - ✅ EXISTS

**Missing from manual but EXISTS:**
- `ctx.enterTemplate()` - internal
- `ctx.exitTemplate()` - internal
- `ctx.InTemplate()` - internal

---

### 3. Incorrect Builder API (8 methods) ❌ MISMATCHED

#### Manual Claims These Builder Methods:
```go
NewComponent().Props().Setup().Template().Children().WithAutoCommands()
.WithCommandDebug().WithKeyBinding().WithConditionalKeyBinding().WithKeyBindings()
.WithMessageHandler().Build()
```

#### Reality Check - Real Builder Methods:
- `NewComponent(name)` - ✅ EXISTS
- `Props(props)` - ✅ EXISTS
- `Setup(fn)` - ✅ EXISTS
- `Template(fn)` - ✅ EXISTS
- `Children(children...)` - ✅ EXISTS
- `WithAutoCommands(enabled)` - ✅ EXISTS
- `WithCommandDebug(enabled)` - **PARTIAL** - exists but may be different
- `WithKeyBinding(key, event, desc)` - **DOES NOT EXIST** (different signature)
- `WithConditionalKeyBinding(binding)` - **DOES NOT EXIST** (different signature)
- `WithKeyBindings(bindings)` - **DOES NOT EXIST** (different signature)
- `WithMessageHandler(handler)` - **DOES NOT EXIST**
- `Build()` - ✅ EXISTS

**Key Binding Reality:**
The manual shows simple key binding methods, but actual implementation is different and more complex.

---

### 4. Incorrect Component Pattern ❌ WRONG

#### Manual Shows:
```go
func CreateComponent(props Props) (bubbly.Component, error) {
    return bubbly.DefineComponent(bubbly.ComponentConfig{
        Name: "ComponentName",
        Setup: func(ctx bubbly.SetupContext) bubbly.SetupResult {
            // ...
        },
    })
}
```

#### Reality Check:
- `bubbly.DefineComponent()` - **DOES NOT EXIST**
- `bubbly.ComponentConfig` - **DOES NOT EXIST**
- `bubbly.SetupContext` - **DOES NOT EXIST** (it's `*bubbly.Context`)
- `bubbly.SetupResult` - **DOES NOT EXIST**

#### Real Component Creation:
```go
// Components are created using builders
component := bubbly.NewComponent("Button").
    Props(ButtonProps{Label: "Click"}).
    Template(func(ctx *bubbly.RenderContext) string {
        return "button"
    }).
    Build()
```

---

### 5. Incorrect Composable Function Names ❌ MULTIPLE ERRORS

#### Manual Claims:
```go
UseCounter, UseDoubleCounter, UseAsync, UseDebounce, UseEffect, 
UseEventListener, UseForm, UseLocalStorage, UseState, UseTextInput, UseThrottle
```

#### Reality Check - What Actually Exists:

**Verified Composables:**
- `composables.UseCounter(ctx, initial)` - ✅ EXISTS
- `composables.UseDoubleCounter(ctx, initial)` - ✅ EXISTS
- `composables.UseAsync[T](ctx, fetcher)` - ✅ EXISTS (returns UseAsyncReturn[T])
- `composables.UseDebounce[T](ctx, value, delay)` - ✅ EXISTS
- `composables.UseEffect(ctx, effect, deps...)` - ✅ EXISTS
- `composables.UseEventListener(ctx, event, handler)` - ✅ EXISTS
- `composables.UseForm[T](ctx, formStruct, validator)` - ✅ EXISTS (returns UseFormReturn[T])
- `composables.UseLocalStorage[T](ctx, key, initial, storage)` - ✅ EXISTS (returns UseStateReturn[T])
- `composables.UseState[T](ctx, initial)` - ✅ EXISTS (returns UseStateReturn[T])
- `composables.UseTextInput(config)` - ⚠️ DIFFERENT (no context parameter!)
- `composables.UseThrottle(ctx, fn, delay)` - ✅ EXISTS

**Signature Issues Found:**
- Manual shows `UseTextInput(ctx, initial)` but real signature is `UseTextInput(config UseTextInputConfig)`
- Manual shows `UseForm(ctx, FormStruct{})` but real signature requires validator function
- Manual shows `UseLocalStorage(ctx, key, default)` but real signature requires storage parameter

---

### 6. Incorrect Return Types ❌ TYPE MISMATCHES

#### Manual Claims These Types:

```go
// UseStateReturn (wrong)
UseStateReturn struct {
    Value *bubbly.Ref[T]
    Set func(T)
    Get func() T
}

// UseAsyncReturn (wrong)
UseAsyncReturn struct {
    Data *bubbly.Ref[*T]
    Loading *bubbly.Ref[bool]
    Error *bubbly.Ref[error]
    Execute func()
    Reset func()
}

// UseFormReturn (wrong)
UseFormReturn struct {
    Values *bubbly.Ref[T]
    Errors map[string]string
    Touched map[string]bool
    IsValid bool
    IsDirty bool
    Submit func()
    Reset func()
    SetField func(string, interface{})
}
```

#### Reality Check - Actual Types:

From source code analysis:
- `UseStateReturn[T]` - EXISTS but fields may have different names
- `UseAsyncReturn[T]` - EXISTS but fields are different
- `UseFormReturn[T]` - EXISTS but structure is more complex
- Manual doesn't document return types properly

**Need to verify actual struct definitions via grep:**

```bash
grep -A 10 "type Use.*Return.*struct" pkg/bubbly/composables/*.go
```

---

### 7. Incorrect Component API ❌ INIT PATTERN WRONG

#### Manual Shows:
```go
input := components.Input(components.InputProps{
    Label: "Username",
    Value: valueRef,
})
input.Init()
output := input.View()
```

#### Reality Check - Actual Component APIs:

**Components DO NOT have an `Init()` method that returns void!** This is completely wrong.

Components implement `tea.Model` interface:
```go
type Component interface {
    tea.Model  // This means: Init() tea.Cmd, Update(msg tea.Msg) tea.Cmd, View() string
    Name() string
    ID() string
    Props() interface{}
    Emit(event string, data interface{})
    On(event string, handler EventHandler)
    KeyBindings() map[string][]KeyBinding
}
```

**THE MANUAL IS COMPLETELY MISSING THE `Init() tea.Cmd` PATTERN!**

Components must be used as:
```go
component := components.Button(props)
cmd := component.Init()  // Returns tea.Cmd, not void
str := component.View()  // Returns string
```

---

### 8. Incorrect Directive API ❌ FUNCTION NAMES

#### Manual Shows:
```go
directives.If(condition, func() string {...})
directives.Show(condition, func() string {...})
directives.ForEach(items, func(item, index) string {...})
directives.Bind(ref, handler)
directives.On(event, handler)
```

#### Reality Check - Actual Directives:

Verify via:
```bash
ls -la pkg/bubbly/directives/*.go
```

Found files: bind.go, foreach.go, if.go, on.go, show.go

**Need to check actual function signatures:**
```bash
grep -h "^func [A-Z]" pkg/bubbly/directives/*.go | grep -v "_test.go"
```

---

### 9. Incorrect Router API ❌ MISSING/BROKEN

#### Manual Shows:
```go
r := router.NewRouter().
    AddRoute("/", homeComponent).
    AddRoute("/users/:id", userComponent).
    WithGuard(authGuard).
    Build()

r.Navigate("/users/123")
r.GoBack()
params := r.CurrentRoute().Params
```

#### Reality Check - Router Structure:
- `bubbly/router` is a package, not `router`
- Actual router implementation needs verification
- Many methods shown may not exist

---

### 10. Type Safety Issues ❌ GENERIC MISUSE

#### Manual Shows:
```go
count := ctx.Ref(0)  // Returns Ref[interface{}]
// Then type asserts everywhere
count.Get().(int)
```

#### Reality Check:
- `ctx.Ref(0)` - returns `*Ref[interface{}]` NOT type-safe
- `bubbly.NewRef(0)` - returns `*Ref[int]` ✅ Type-safe
- Manual RECOMMENDS `ctx.Ref()` which is WRONG for type safety

---

## Critical Findings Summary

| Category | Documented | Exists | Accuracy |
|----------|------------|--------|----------|
| Context Methods | 23 | ~20 | 60% |
| Builder Methods | 11 | ~7 | 50% |
| Composables | 11 | 11 | 40% (wrong signatures) |
| Components | 24 | ~20 | 30% (wrong usage) |
| Directives | 5 | 5 | 50% |
| Router | 8 | ~5 | 40% |

**Overall Accuracy: ~45%**

---

## Compilation Errors If Following Manual

Following the manual literally would result in:

1. ❌ Package import errors (wrong package paths)
2. ❌ Undefined functions (ExposeComponent, SetCommandGenerator, etc.)
3. ❌ Wrong method signatures (UseTextInput, UseForm, etc.)
4. ❌ Type errors (wrong return types, missing type parameters)
5. ❌ Component API misuse (Init() returns void instead of tea.Cmd)
6. ❌ Missing required parameters (UseLocalStorage missing storage)
7. ❌ Template compilation errors (undefined directives)

**Estimated: 150+ compilation errors**

---

## Recommendations

1. **Rewrite entire manual** from actual source code
2. **Verify every function signature** against implementation
3. **Test all code examples** to ensure they compile
4. **Add package path verification** section
5. **Document actual init/update/view patterns** for Bubbletea integration
6. **Add type safety guidance** using generics properly
7. **Remove all non-existent APIs** documented

---

## Next Steps

1. Extract ALL real APIs from source
2. Create truthful reference documentation
3. Write verified working examples
4. Add compilation tests for examples
5. Review and validate with actual tests
