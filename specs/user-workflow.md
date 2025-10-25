# Master User Workflow - BubblyUI

**Purpose:** Map complete user journeys across all features to ensure integration

---

## Complete User Journeys

### Journey 1: Build Simple Counter App (15 minutes)

**User:** Developer new to BubblyUI, familiar with Vue.js

1. **Feature: 01-reactivity-system**
   - Action: Import bubbly package
   - Action: Create reactive counter with `NewRef(0)`
   - Result: Type-safe reactive primitive ready
   - Leads to: 02-component-model

2. **Feature: 02-component-model**
   - Action: Create component with `NewComponent("Counter")`
   - Action: Use Setup hook to initialize state
   - Result: Component structure defined
   - Leads to: Template rendering

3. **Feature: 02-component-model (template)**
   - Action: Define template function
   - Action: Access ref.Get() in template
   - Result: UI renders counter value
   - Leads to: Event handling

4. **Feature: 05-directives (On)**
   - Action: Add On("keypress") handlers
   - Action: Increment/decrement counter on keys
   - Result: Interactive counter working
   - Leads to: Complete app

5. **End Result:**
   - Working counter application
   - 20 lines of code
   - Type-safe, tested, documented

**Data Flow:**
```
Ref[int](0) → Component.state → Template → Rendered UI
    ↑                                            ↓
    └─────────── On("keypress") ←───────────────┘
```

---

### Journey 2: Build Todo List App (45 minutes)

**User:** Developer building production app

1. **Feature: 01-reactivity-system**
   - Create `todos` Ref with slice
   - Create `filter` Ref with enum
   - Create `filteredTodos` Computed
   - Result: Reactive data model
   - Leads to: Component structure

2. **Feature: 02-component-model**
   - Create TodoList component
   - Create TodoItem component
   - Create TodoInput component
   - Result: Component hierarchy defined
   - Leads to: Composition

3. **Feature: 04-composition-api**
   - Create `useTodos` composable
   - Create `useFilter` composable
   - Share logic across components
   - Result: Reusable stateful logic
   - Leads to: Lifecycle management

4. **Feature: 03-lifecycle-hooks**
   - onMounted: Load todos from storage
   - onUpdated: Save todos to storage
   - onUnmounted: Cleanup resources
   - Result: Persistence working
   - Leads to: Enhanced UI

5. **Feature: 05-directives**
   - Use ForEach() to render list
   - Use If() for empty state
   - Use Bind() for input sync
   - Result: Dynamic UI complete
   - Leads to: Built-in components

6. **Feature: 06-built-in-components**
   - Use Input component for new todos
   - Use Checkbox for completion
   - Use Button for actions
   - Result: Polished UI
   - Leads to: Complete app

7. **End Result:**
   - Production-ready todo app
   - 150 lines of code
   - CRUD operations
   - Persistent storage
   - Type-safe throughout

**Data Flow:**
```
useTodos() composable
    ↓
todos Ref[[]Todo] → filteredTodos Computed
    ↓                      ↓
TodoList Component → ForEach directive
    ↓                      ↓
TodoItem components → Checkbox, Button
    ↓
onUpdated → Save to storage
```

---

### Journey 3: Build Form with Validation (30 minutes)

**User:** Developer creating form-heavy app

1. **Feature: 01-reactivity-system**
   - Create Ref for each field (name, email, age)
   - Create Computed for validation
   - Create Computed for form validity
   - Result: Reactive form state
   - Leads to: Form component

2. **Feature: 02-component-model**
   - Create Form component
   - Create Field subcomponents
   - Define props for validation rules
   - Result: Composable form structure
   - Leads to: Two-way binding

3. **Feature: 05-directives (Bind)**
   - Use Bind() for input synchronization
   - Automatic two-way data flow
   - Result: Less boilerplate
   - Leads to: Validation display

4. **Feature: 05-directives (If)**
   - Use If() to show/hide errors
   - Conditional field rendering
   - Result: Dynamic error display
   - Leads to: Built-in components

5. **Feature: 06-built-in-components (molecules)**
   - Use Input component
   - Use TextArea component
   - Use Select component
   - Result: Consistent UI
   - Leads to: Submit handling

6. **Feature: 04-composition-api**
   - Create `useValidation` composable
   - Create `useForm` composable
   - Reuse across forms
   - Result: DRY form logic
   - Leads to: Complete form

7. **End Result:**
   - Fully validated form
   - Real-time error feedback
   - Reusable validation logic
   - Type-safe field values

**Data Flow:**
```
Field Refs → Validation Computed → isValid Computed
    ↓              ↓                       ↓
Input (Bind) → Error (If) → Submit Button (disabled)
```

---

### Journey 4: Migrate from Pure Bubbletea (2 hours)

**User:** Existing Bubbletea developer

1. **Start: Existing Bubbletea App**
   - Has Model, Init, Update, View
   - Manual state management
   - Message passing
   - 200 lines of code

2. **Feature: 02-component-model**
   - Wrap existing model in Component
   - Keep Update logic initially
   - Result: BubblyUI structure
   - Leads to: Reactivity

3. **Feature: 01-reactivity-system**
   - Extract state to Refs
   - Replace manual updates with ref.Set()
   - Result: Reactive state
   - Leads to: Simplified Update

4. **Feature: 02-component-model (events)**
   - Replace Update branches with On() handlers
   - Simplify message handling
   - Result: Less boilerplate
   - Leads to: Template simplification

5. **Feature: 05-directives**
   - Use ForEach() instead of loops
   - Use If() instead of conditionals
   - Result: Cleaner View
   - Leads to: Final optimization

6. **End Result:**
   - Migrated app (120 lines, 40% reduction)
   - Reactive state
   - Simpler Update logic
   - Cleaner View
   - Type-safe

**Before/After:**
```go
// Before: Manual state
type model struct {
    count int
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case KeyMsg:
        if msg.String() == "+" {
            m.count++
        }
    }
    return m, nil
}

// After: Reactive
count := NewRef(0)
On("keypress", func(key string) {
    if key == "+" {
        count.Set(count.Get() + 1)
    }
})
```

---

## Feature Connection Map

```
01-reactivity-system (Foundation)
    ↓
    ├─> 02-component-model (Uses Refs)
    │       ↓
    │       ├─> 03-lifecycle-hooks (Component lifecycle)
    │       │       ↓
    │       │       └─> 04-composition-api (Composables + hooks)
    │       │
    │       └─> 05-directives (Component templates)
    │               ↓
    │               └─> 06-built-in-components (Uses directives)
    │
    └─> 04-composition-api (Uses Refs directly)
            ↓
            └─> Composables (useX functions)
```

---

## Data Flow Between Features

### Reactivity → Components
- **Exports:** Ref[T], Computed[T], Watch()
- **Imports:** None
- **Flow:** Components store Refs in state, access via Get/Set

### Components → Lifecycle
- **Exports:** Component interface, lifecycle execution
- **Imports:** None
- **Flow:** Lifecycle hooks called by component runtime

### Components → Directives
- **Exports:** Template context, render functions
- **Imports:** Directive implementations
- **Flow:** Templates call directives for enhanced rendering

### Composition API → All Features
- **Exports:** Composable pattern, Context API
- **Imports:** Reactivity, Components, Lifecycle
- **Flow:** Composables use features, return to components

### Built-in Components → Framework
- **Exports:** Reusable components
- **Imports:** All framework features
- **Flow:** Components use full framework capabilities

---

## Navigation Structure

### App Entry Point
```go
func main() {
    app := bubbly.NewApp()
    root := createRootComponent()
    app.Mount(root).Run()
}
```

### Component Tree
```
App (root)
├── Header
│   ├── Logo (atom)
│   └── Nav (molecule)
├── Main
│   ├── Sidebar (organism)
│   │   └── NavList (molecule)
│   └── Content (organism)
│       ├── Title (atom)
│       ├── Form (organism)
│       │   ├── Input (molecule)
│       │   ├── Select (molecule)
│       │   └── Button (atom)
│       └── Table (organism)
└── Footer
    └── Text (atom)
```

### State Sharing

#### Global State
- App-level Refs (user session, theme, config)
- Provided via Context API
- Injected in any component

#### Feature State
- TodoList: todos, filter
- Form: fields, validation
- Table: data, sorting, pagination

#### Component State
- Local Refs (search term, open/closed)
- Not shared with other components

---

## Integration Points

### Feature 01 → Feature 02
- Reactivity exports: Ref, Computed, Watch
- Component imports: Uses for state management
- Integration: Component.state stores Refs

### Feature 02 → Feature 03
- Component exports: Lifecycle execution points
- Lifecycle imports: Hooks API
- Integration: Hooks called at component milestones

### Feature 02 → Feature 05
- Component exports: Template context
- Directives import: Render helpers
- Integration: Directives enhance templates

### Feature 04 → All
- Composition exports: Composable pattern
- All features import: Use in composables
- Integration: Composables orchestrate features

### Feature 06 → Framework
- Components export: Reusable UI elements
- Framework imports: None (components use framework)
- Integration: Components demonstrate best practices

---

## Orphan Detection

### Components Created but Not Used
**Status:** ✅ None expected

All components in 06-built-in-components are used in:
- Example applications
- Documentation examples
- Test suites

### Types Defined but Not Used
**Status:** ✅ None expected

All types are:
- Used in public API
- Referenced in components
- Part of framework contracts

### Composables Not Used
**Status:** ✅ None expected

All composables in 04-composition-api are:
- Used in examples
- Documented with use cases
- Tested in integration tests

### API Endpoints N/A
No API endpoints in TUI framework

---

## Example Applications Flow

### Counter Example
```
01-reactivity → Create Ref(0)
                    ↓
02-component → Wrap in Component
                    ↓
05-directive → On("keypress")
                    ↓
               Render count
```

### Todo Example
```
01-reactivity → Refs + Computed
                    ↓
04-composition → useTodos()
                    ↓
02-component → TodoList, TodoItem
                    ↓
05-directive → ForEach, If, Bind
                    ↓
06-components → Input, Checkbox
                    ↓
03-lifecycle → onMounted, onUpdated
```

### Form Example
```
01-reactivity → Field Refs + Validation Computed
                    ↓
04-composition → useForm(), useValidation()
                    ↓
02-component → Form, Field components
                    ↓
05-directive → Bind, If (errors)
                    ↓
06-components → Input, Select, Button
```

### Dashboard Example
```
01-reactivity → Data Refs + Filtered/Sorted Computed
                    ↓
04-composition → useData(), useFilters()
                    ↓
02-component → Dashboard, Chart, Table components
                    ↓
06-components → Table, Card, Button organisms
                    ↓
06-components → AppLayout template
```

---

## Cross-Feature Scenarios

### Scenario: Real-time Data Updates

1. WebSocket receives data
2. Watcher updates Ref (01-reactivity)
3. Computed recalculates (01-reactivity)
4. Component detects change (02-component)
5. Template re-renders (02-component)
6. ForEach updates list (05-directive)
7. Table shows new data (06-component)

### Scenario: Form Submission

1. User types in Input (06-component)
2. Bind syncs to Ref (05-directive)
3. Validation Computed runs (01-reactivity)
4. Error state updates (01-reactivity)
5. If directive shows/hides error (05-directive)
6. Submit Button enables/disables (computed)
7. onSubmit handler called (02-component)

### Scenario: Component Unmount

1. User navigates away
2. Component onUnmounted (03-lifecycle)
3. Watchers cleaned up (01-reactivity)
4. Resources released
5. No memory leaks

---

## Validation Checklist

### Feature Integration
- [ ] Reactivity works in components ✓
- [ ] Components use lifecycle hooks ✓
- [ ] Composition API integrates all features ✓
- [ ] Directives work in templates ✓
- [ ] Built-in components use framework ✓

### User Journeys
- [ ] Counter journey complete ✓
- [ ] Todo journey complete ✓
- [ ] Form journey complete ✓
- [ ] Migration journey complete ✓

### No Orphans
- [ ] All components have parent usage ✓
- [ ] All types are referenced ✓
- [ ] All composables are used ✓
- [ ] All examples work end-to-end ✓

### Data Flow
- [ ] State flows top-down ✓
- [ ] Events bubble up ✓
- [ ] Side effects contained ✓
- [ ] No circular dependencies ✓

---

## Performance Considerations

### Journey Performance Targets

| Journey | Target Time | Actual | Status |
|---------|-------------|--------|--------|
| Counter app | 15 min | - | Not measured |
| Todo app | 45 min | - | Not measured |
| Form app | 30 min | - | Not measured |
| Migration | 2 hours | - | Not measured |

### Runtime Performance

| Operation | Target | Actual | Status |
|-----------|--------|--------|--------|
| App startup | < 100ms | - | Not measured |
| Ref update | < 100ns | - | Not measured |
| Component render | < 10ms | - | Not measured |
| List update (100 items) | < 50ms | - | Not measured |

---

## Success Criteria

✅ **All Journeys Complete:**
- [ ] Developer can build counter in 15 minutes
- [ ] Developer can build todo in 45 minutes
- [ ] Developer can build form in 30 minutes
- [ ] Developer can migrate app in 2 hours

✅ **No Orphaned Features:**
- [ ] All features used in journeys
- [ ] All components used in examples
- [ ] All types used in framework
- [ ] All composables used in apps

✅ **Smooth Integration:**
- [ ] Features compose naturally
- [ ] No awkward workarounds needed
- [ ] API feels cohesive
- [ ] Documentation clear

✅ **Performance Acceptable:**
- [ ] All targets met
- [ ] No blocking operations
- [ ] Responsive UI
- [ ] Efficient updates
