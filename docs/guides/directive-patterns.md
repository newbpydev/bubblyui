# BubblyUI Directive Patterns & Best Practices

Advanced patterns, performance optimization, and real-world examples for directive usage.

## Table of Contents

- [Composition Patterns](#composition-patterns)
- [Performance Patterns](#performance-patterns)
- [Error Handling Patterns](#error-handling-patterns)
- [Testing Patterns](#testing-patterns)
- [Real-World Examples](#real-world-examples)

---

## Composition Patterns

### Pattern 1: Conditional Lists

**Problem:** Show different content based on list state

```go
Template(func(ctx RenderContext) string {
    items := ctx.Get("items").(*Ref[[]Task])
    loading := ctx.Get("loading").(*Ref[bool])
    
    return directives.If(loading.Get(),
        func() string { return "⏳ Loading..." },
    ).ElseIf(len(items.Get()) > 0, func() string {
        return directives.ForEach(items.Get(), func(task Task, i int) string {
            return renderTask(task, i)
        }).Render()
    }).Else(func() string {
        return "✨ No tasks yet. Add one!"
    }).Render()
})
```

**When to use:** Loading states, empty states, error states

### Pattern 2: Filtered Lists with State

**Problem:** Filter list based on reactive state

```go
Setup(func(ctx *Context) {
    allTasks := ctx.Ref([]Task{})
    filter := ctx.Ref("all") // "all", "active", "completed"
    
    // Computed filtered list
    filteredTasks := ctx.Computed(func() []Task {
        tasks := allTasks.Get()
        switch filter.Get() {
        case "active":
            return filterActive(tasks)
        case "completed":
            return filterCompleted(tasks)
        default:
            return tasks
        }
    })
    
    ctx.Expose("filteredTasks", filteredTasks)
})

Template(func(ctx RenderContext) string {
    tasks := ctx.Get("filteredTasks").(*Computed[[]Task])
    return directives.ForEach(tasks.Get(), renderTask).Render()
})
```

**When to use:** Search, filters, categories

### Pattern 3: Nested Conditionals with Show

**Problem:** Multiple visibility states without removing from output

```go
Template(func(ctx RenderContext) string {
    showDetails := ctx.Get("showDetails").(*Ref[bool])
    showAdvanced := ctx.Get("showAdvanced").(*Ref[bool])
    
    basicInfo := "Basic Information\n"
    
    details := directives.Show(showDetails.Get(), func() string {
        return "Detailed Information\n"
    }).WithTransition().Render()
    
    advanced := directives.Show(showAdvanced.Get(), func() string {
        return "Advanced Options\n"
    }).WithTransition().Render()
    
    return basicInfo + details + advanced
})
```

**When to use:** Accordion menus, expandable sections, progressive disclosure

### Pattern 4: Event-Driven Lists

**Problem:** Handle events on list items

```go
Template(func(ctx RenderContext) string {
    items := ctx.Get("items").(*Ref[[]string])
    
    return directives.ForEach(items.Get(), func(item string, i int) string {
        return directives.On("click", func(data interface{}) {
            handleItemClick(i, item)
        }).On("delete", func(data interface{}) {
            deleteItem(i)
        }).Render(
            fmt.Sprintf("[%d] %s\n", i+1, item),
        )
    }).Render()
})
```

**When to use:** Interactive lists, menus, selections

### Pattern 5: Form with Multiple Bindings

**Problem:** Multiple input fields with validation

```go
type LoginForm struct {
    Username string
    Password string
    Remember bool
}

Setup(func(ctx *Context) {
    form := ctx.Ref(LoginForm{})
    
    username := ctx.Computed(func() string {
        return form.Get().Username
    })
    password := ctx.Computed(func() string {
        return form.Get().Password
    })
    remember := ctx.Computed(func() bool {
        return form.Get().Remember
    })
    
    // Create bindings
    usernameRef := ctx.Ref("")
    passwordRef := ctx.Ref("")
    rememberRef := ctx.Ref(false)
    
    ctx.Expose("usernameInput", directives.Bind(usernameRef))
    ctx.Expose("passwordInput", directives.Bind(passwordRef))
    ctx.Expose("rememberCheck", directives.BindCheckbox(rememberRef))
    
    // Validation
    isValid := ctx.Computed(func() bool {
        return len(usernameRef.Get()) >= 3 && len(passwordRef.Get()) >= 8
    })
    ctx.Expose("canSubmit", isValid)
})

Template(func(ctx RenderContext) string {
    userInput := ctx.Get("usernameInput").(*directives.BindDirective[string])
    passInput := ctx.Get("passwordInput").(*directives.BindDirective[string])
    rememberCheck := ctx.Get("rememberCheck").(*directives.BindDirective[bool])
    canSubmit := ctx.Get("canSubmit").(*Computed[bool])
    
    username := fmt.Sprintf("Username: %s\n", userInput.Render())
    password := fmt.Sprintf("Password: %s\n", passInput.Render())
    remember := fmt.Sprintf("%s Remember me\n", rememberCheck.Render())
    
    submitBtn := directives.If(canSubmit.Get(),
        func() string {
            return directives.On("submit", handleSubmit).
                PreventDefault().
                Render("[ Submit ]")
        },
    ).Else(func() string {
        return "[Submit] (disabled)"
    }).Render()
    
    return username + password + remember + "\n" + submitBtn
})
```

**When to use:** Forms, settings, configuration UIs

---

## Performance Patterns

### Pattern 6: Pre-allocate with strings.Builder

**Problem:** Building complex strings with many concatenations

```go
// ❌ Slow (multiple allocations)
func renderList(items []string) string {
    result := ""
    for i, item := range items {
        result += fmt.Sprintf("%d. %s\n", i+1, item)
    }
    return result
}

// ✅ Fast (single allocation with preallocation)
func renderList(items []string) string {
    // Pre-calculate capacity
    capacity := 0
    for _, item := range items {
        capacity += len(item) + 10 // number + ". " + "\n"
    }
    
    var builder strings.Builder
    builder.Grow(capacity)
    
    for i, item := range items {
        builder.WriteString(strconv.Itoa(i + 1))
        builder.WriteString(". ")
        builder.WriteString(item)
        builder.WriteString("\n")
    }
    
    return builder.String()
}
```

**Performance:** 2-5x faster, 60-80% fewer allocations

### Pattern 7: Type Assertions for Fast Paths

**Problem:** Generic function performance

```go
// ✅ Optimized with type assertion fast path
func renderValue[T any](value T) string {
    // Fast path for common types
    if str, ok := any(value).(string); ok {
        return str // Zero allocations
    }
    if num, ok := any(value).(int); ok {
        return strconv.Itoa(num) // Faster than fmt.Sprint
    }
    if b, ok := any(value).(bool); ok {
        if b {
            return "true"
        }
        return "false"
    }
    
    // Fallback for other types
    return fmt.Sprint(value)
}
```

**Performance:** 4-6x faster for common types

### Pattern 8: Memoize Expensive Computations

**Problem:** Re-computing same values on every render

```go
Setup(func(ctx *Context) {
    allItems := ctx.Ref([]Item{})
    searchTerm := ctx.Ref("")
    
    // ✅ Computed automatically memoizes
    searchResults := ctx.Computed(func() []Item {
        term := searchTerm.Get()
        if term == "" {
            return allItems.Get()
        }
        
        // Expensive filtering
        return filterItems(allItems.Get(), term)
    })
    
    ctx.Expose("results", searchResults)
})

Template(func(ctx RenderContext) string {
    // Only re-renders when searchResults actually changes
    results := ctx.Get("results").(*Computed[[]Item])
    return directives.ForEach(results.Get(), renderItem).Render()
})
```

**Performance:** Only recomputes when dependencies change

### Pattern 9: Avoid fmt.Sprintf in Hot Paths

**Problem:** Format overhead in frequently called code

```go
// ❌ Slow (uses reflection)
func renderTask(task Task) string {
    return fmt.Sprintf("[%s] %s", task.Status, task.Title)
}

// ✅ Fast (direct concatenation)
func renderTask(task Task) string {
    var builder strings.Builder
    builder.Grow(len(task.Status) + len(task.Title) + 4)
    builder.WriteString("[")
    builder.WriteString(task.Status)
    builder.WriteString("] ")
    builder.WriteString(task.Title)
    return builder.String()
}
```

**Performance:** 2-4x faster

### Pattern 10: Batch Updates

**Problem:** Multiple reactive updates causing renders

```go
// ❌ Triggers multiple renders
func updateForm(ctx *Context) {
    name := ctx.Get("name").(*Ref[string])
    email := ctx.Get("email").(*Ref[string])
    age := ctx.Get("age").(*Ref[int])
    
    name.Set("John")    // Render 1
    email.Set("j@e.com") // Render 2
    age.Set(30)         // Render 3
}

// ✅ Single update with struct
type FormData struct {
    Name  string
    Email string
    Age   int
}

func updateForm(ctx *Context) {
    form := ctx.Get("form").(*Ref[FormData])
    form.Set(FormData{
        Name:  "John",
        Email: "j@e.com",
        Age:   30,
    }) // Single render
}
```

**Performance:** Reduces renders by batching updates

---

## Error Handling Patterns

### Pattern 11: Safe Get with Defaults

**Problem:** Handle missing or nil values

```go
func safeGetString(ctx RenderContext, key string, defaultVal string) string {
    val := ctx.Get(key)
    if val == nil {
        return defaultVal
    }
    if ref, ok := val.(*Ref[string]); ok {
        return ref.Get()
    }
    return defaultVal
}

Template(func(ctx RenderContext) string {
    username := safeGetString(ctx, "username", "Guest")
    return fmt.Sprintf("Hello, %s!", username)
})
```

### Pattern 12: Graceful Degradation

**Problem:** Handle errors without breaking UI

```go
Template(func(ctx RenderContext) string {
    items := ctx.Get("items")
    
    // Gracefully handle missing data
    if items == nil {
        return "Loading..."
    }
    
    itemsRef, ok := items.(*Ref[[]Task])
    if !ok {
        return "Error loading tasks"
    }
    
    tasks := itemsRef.Get()
    if len(tasks) == 0 {
        return "No tasks yet"
    }
    
    return directives.ForEach(tasks, renderTask).Render()
})
```

### Pattern 13: Validation Before Binding

**Problem:** Validate input before updating Ref

```go
Setup(func(ctx *Context) {
    age := ctx.Ref(0)
    
    ctx.On("ageInput", func(data interface{}) {
        if input, ok := data.(string); ok {
            if val, err := strconv.Atoi(input); err == nil {
                // Validate range
                if val >= 0 && val <= 120 {
                    age.Set(val)
                } else {
                    // Show error
                    ctx.Get("error").(*Ref[string]).Set("Age must be 0-120")
                }
            }
        }
    })
})
```

---

## Testing Patterns

### Pattern 14: Test Directive Output

```go
func TestIfDirective_Rendering(t *testing.T) {
    tests := []struct {
        name      string
        condition bool
        expected  string
    }{
        {
            name:      "true renders content",
            condition: true,
            expected:  "Content",
        },
        {
            name:      "false renders empty",
            condition: false,
            expected:  "",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := directives.If(tt.condition,
                func() string { return "Content" },
            ).Render()
            
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Pattern 15: Test Reactive Updates

```go
func TestReactiveListUpdate(t *testing.T) {
    // Setup component
    comp := NewComponent(
        Setup(func(ctx *Context) {
            items := ctx.Ref([]string{"A", "B"})
            ctx.Expose("items", items)
        }),
        Template(func(ctx RenderContext) string {
            items := ctx.Get("items").(*Ref[[]string])
            return directives.ForEach(items.Get(), func(item string, i int) string {
                return item
            }).Render()
        }),
    )
    
    // Initial render
    output := comp.Render()
    assert.Equal(t, "AB", output)
    
    // Update and verify
    items := comp.GetExposed("items").(*Ref[[]string])
    items.Set([]string{"X", "Y", "Z"})
    
    output = comp.Render()
    assert.Equal(t, "XYZ", output)
}
```

---

## Real-World Examples

### Example 1: Todo List Application

Complete implementation with all directive types:

```go
type Task struct {
    ID        int
    Title     string
    Completed bool
}

func TodoApp() *Component {
    return NewComponent(
        Setup(func(ctx *Context) {
            // State
            tasks := ctx.Ref([]Task{})
            filter := ctx.Ref("all") // "all", "active", "completed"
            newTaskTitle := ctx.Ref("")
            
            // Computed filtered tasks
            visibleTasks := ctx.Computed(func() []Task {
                all := tasks.Get()
                switch filter.Get() {
                case "active":
                    return filterBy(all, false)
                case "completed":
                    return filterBy(all, true)
                default:
                    return all
                }
            })
            
            // Event handlers
            ctx.On("addTask", func(data interface{}) {
                title := newTaskTitle.Get()
                if title != "" {
                    current := tasks.Get()
                    tasks.Set(append(current, Task{
                        ID:    len(current) + 1,
                        Title: title,
                    }))
                    newTaskTitle.Set("")
                }
            })
            
            ctx.On("toggleTask", func(data interface{}) {
                if id, ok := data.(int); ok {
                    current := tasks.Get()
                    for i := range current {
                        if current[i].ID == id {
                            current[i].Completed = !current[i].Completed
                            break
                        }
                    }
                    tasks.Set(current)
                }
            })
            
            ctx.Expose("tasks", tasks)
            ctx.Expose("filter", filter)
            ctx.Expose("newTaskTitle", newTaskTitle)
            ctx.Expose("visibleTasks", visibleTasks)
            ctx.Expose("newTaskInput", directives.Bind(newTaskTitle))
        }),
        
        Template(func(ctx RenderContext) string {
            newTaskInput := ctx.Get("newTaskInput").(*directives.BindDirective[string])
            visibleTasks := ctx.Get("visibleTasks").(*Computed[[]Task])
            tasks := visibleTasks.Get()
            
            // Header
            header := "# Todo List\n\n"
            
            // New task input
            input := fmt.Sprintf("New task: %s\n", newTaskInput.Render())
            addBtn := directives.On("addTask", func(interface{}) {}).
                PreventDefault().
                Render("[Add Task]")
            inputSection := input + addBtn + "\n\n"
            
            // Task list
            taskList := directives.If(len(tasks) > 0,
                func() string {
                    return directives.ForEach(tasks, func(task Task, i int) string {
                        checkbox := "[" + directives.If(task.Completed,
                            func() string { return "X" },
                        ).Else(func() string {
                            return " "
                        }).Render() + "]"
                        
                        return directives.On("toggleTask", func(interface{}) {
                            // Pass task ID
                        }).Render(
                            fmt.Sprintf("%s %s\n", checkbox, task.Title),
                        )
                    }).Render()
                },
            ).Else(func() string {
                return "No tasks yet!\n"
            }).Render()
            
            // Footer with stats
            all := ctx.Get("tasks").(*Ref[[]Task])
            total := len(all.Get())
            completed := len(filterBy(all.Get(), true))
            footer := fmt.Sprintf("\n%d/%d completed", completed, total)
            
            return header + inputSection + taskList + footer
        }),
    )
}

func filterBy(tasks []Task, completed bool) []Task {
    result := make([]Task, 0, len(tasks))
    for _, task := range tasks {
        if task.Completed == completed {
            result = append(result, task)
        }
    }
    return result
}
```

### Example 2: Settings Panel

Multi-section form with validation:

```go
type Settings struct {
    Username   string
    Email      string
    Theme      string
    Notify     bool
    AutoSave   bool
}

func SettingsPanel() *Component {
    return NewComponent(
        Setup(func(ctx *Context) {
            settings := ctx.Ref(Settings{
                Username: "user",
                Email:    "user@example.com",
                Theme:    "dark",
                Notify:   true,
                AutoSave: false,
            })
            
            // Individual refs for binding
            username := ctx.Ref(settings.Get().Username)
            email := ctx.Ref(settings.Get().Email)
            theme := ctx.Ref(settings.Get().Theme)
            notify := ctx.Ref(settings.Get().Notify)
            autoSave := ctx.Ref(settings.Get().AutoSave)
            
            // Validation
            isValid := ctx.Computed(func() bool {
                return len(username.Get()) >= 3 && 
                       strings.Contains(email.Get(), "@")
            })
            
            // Expose bindings
            ctx.Expose("usernameInput", directives.Bind(username))
            ctx.Expose("emailInput", directives.Bind(email))
            ctx.Expose("themeSelect", directives.BindSelect(theme, 
                []string{"light", "dark", "auto"}))
            ctx.Expose("notifyCheck", directives.BindCheckbox(notify))
            ctx.Expose("autoSaveCheck", directives.BindCheckbox(autoSave))
            ctx.Expose("isValid", isValid)
            
            ctx.On("save", func(interface{}) {
                if isValid.Get() {
                    settings.Set(Settings{
                        Username: username.Get(),
                        Email:    email.Get(),
                        Theme:    theme.Get(),
                        Notify:   notify.Get(),
                        AutoSave: autoSave.Get(),
                    })
                }
            })
        }),
        
        Template(func(ctx RenderContext) string {
            username := ctx.Get("usernameInput").(*directives.BindDirective[string])
            email := ctx.Get("emailInput").(*directives.BindDirective[string])
            theme := ctx.Get("themeSelect").(*directives.SelectBindDirective[string])
            notify := ctx.Get("notifyCheck").(*directives.BindDirective[bool])
            autoSave := ctx.Get("autoSaveCheck").(*directives.BindDirective[bool])
            isValid := ctx.Get("isValid").(*Computed[bool])
            
            sections := []string{
                "## Settings\n",
                "\n### Account\n",
                fmt.Sprintf("Username: %s\n", username.Render()),
                fmt.Sprintf("Email: %s\n", email.Render()),
                "\n### Appearance\n",
                "Theme:\n" + theme.Render(),
                "\n### Preferences\n",
                fmt.Sprintf("%s Enable notifications\n", notify.Render()),
                fmt.Sprintf("%s Auto-save changes\n", autoSave.Render()),
                "\n",
            }
            
            saveBtn := directives.If(isValid.Get(),
                func() string {
                    return directives.On("save", func(interface{}) {}).
                        Render("[Save Settings]")
                },
            ).Else(func() string {
                return "[Save Settings] (fix errors first)"
            }).Render()
            
            var builder strings.Builder
            for _, section := range sections {
                builder.WriteString(section)
            }
            builder.WriteString(saveBtn)
            
            return builder.String()
        }),
    )
}
```

---

## Performance Optimization Summary

### Key Takeaways

1. **Use strings.Builder** with Grow() for string construction
2. **Type assertions** for fast paths on common types
3. **Memoize** with Computed for expensive operations
4. **Batch updates** to reduce render cycles
5. **Pre-filter** large lists before iteration
6. **Avoid fmt.Sprintf** in hot paths
7. **BindCheckbox** for booleans (zero allocations)
8. **Pre-allocate** slices when size is known

### Benchmark Results

Applied optimizations achieved:

- **On directive:** 5.2-6.2x faster (now meets <80ns target)
- **BindCheckbox:** 5.9x faster with zero allocations
- **BindSelect:** 3.8-6.1x faster with 68-91% fewer allocations

These optimizations used:
- strings.Builder with preallocation
- Type assertions for fast paths
- Single-pass string construction
- Elimination of redundant conversions

---

## Next Steps

- **[Directives API Reference](directives.md)**: Complete usage guide
- **[Performance Guide](performance-optimization.md)**: Deep dive into optimization
- **[Component Patterns](../components.md)**: Integrate with components
- **[Reactive System](reactive-dependencies.md)**: Use with Ref[T] and Computed[T]

---

**Remember:** BubblyUI is a TUI framework. All examples render to terminal output, not web browsers.
