# BubblyUI

BubblyUI is an open-source Go framework for building reactive, component-based terminal UIs on top of Bubble Tea and Lip Gloss. It transforms Bubble Tea's Elm-inspired architecture into a React/Solid-like component system with fine-grained reactivity, independent component state, and elegant composition patterns.

## Technology Stack

* **Language:** Go (1.24+)
* **TUI Core:** [Bubble Tea](https://github.com/charmbracelet/bubble-tea)
* **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
* **Hot-Reload:** [Air](https://github.com/cosmtrek/air) (or similar file-watcher)
* **Testing:** Go's built-in `testing` package with comprehensive edge case coverage
* **Benchmarking:** Go's `testing/benchmark` for performance testing
* **CI/CD:** GitHub Actions with test quality gates

## Project Structure

```
bubblyui/                   # repository root
├── cmd/                    # example CLI applications
├── internal/               # private implementation
│   ├── dom/                # virtual DOM engine & diff logic
│   └── core/               # Bubble Tea integration
├── components/             # reusable UI components
├── examples/               # demos (hot-reload, theming)
├── pkg/termidom/           # public API: fluent builder & types
├── configs/                # sample config files (themes)
├── scripts/                # helper scripts (lint, test)
├── README.md               # this file
├── go.mod                  # Go module definition
└── go.sum                  # dependency checksums
```

## Project Details

BubblyUI addresses the limitations of Bubble Tea's monolithic architecture (where all state is centralized) by providing a component-based system inspired by modern frontend frameworks. It introduces:

* A **component-based architecture** where each UI element is an independent, reusable piece with its own state.
* A **reactive state model** using signals and hooks to enable fine-grained updates without re-rendering the entire UI.
* **Parent-child data flow** with props flowing down and events/callbacks flowing up.
* **Lifecycle hooks** (`OnMount`, `OnUpdate`, `OnUnmount`) to manage component lifecycles.
* **Per-component styling** with Lip Gloss for beautiful, consistent UIs.
* **Hot-reload** support for rapid iteration.
* Core components: `Box`, `Text`, `List`, `Button`, and theming support.

## Architecture

BubblyUI combines React's component model with Solid.js's reactivity system, adapting both for terminal UIs:

### Component-Based Architecture

Instead of Bubble Tea's single global model, BubblyUI uses independent components with their own state:

```go
// Simplified Component structure
type Component struct {
    Props     Props
    State     State
    Children  []Component
    Lifecycle Lifecycle
    Render    func() string
}
```

This approach allows:

1. **Modular UI development**: Build complex UIs from simple, reusable components
2. **Isolated state management**: Each component manages its own concerns
3. **Better maintainability**: Changes to one component don't affect others
4. **Easier testing**: Components can be tested in isolation

### Reactive State Model

BubblyUI implements a reactive state system inspired by Solid.js's signals:

```go
// Example signal usage (simplified)
type Counter struct {
    Count *Signal[int]
}

func NewCounter() *Counter {
    count, setCount := CreateSignal(0)
    
    // Only components that use count.Get() will update when count changes
    return &Counter{
        Count: count,
    }
}

func (c *Counter) View() string {
    // This component only re-renders when c.Count.Get() changes
    return fmt.Sprintf("Count: %d", c.Count.Get())
}
```

This fine-grained reactivity ensures:

1. **Minimal redraws**: Only affected components update when state changes
2. **Predictable data flow**: State changes trigger automatic UI updates
3. **Better performance**: Terminal doesn't flicker with unnecessary redraws

### Parent-Child Data Flow

BubblyUI implements a clean, predictable data flow pattern:

```go
// Parent component passing props down and handling events up
func Dashboard() *Element {
    notifications, setNotifications := CreateSignal(0)
    
    return Box().Children(
        Header().Title("Dashboard").Notifications(notifications.Get()),
        Sidebar().OnNotification(func(count int) {
            setNotifications(count)
        }),
    )
}
```

This pattern ensures:

1. **Props flow down**: Parents pass data to children via constructor parameters or fields
2. **Events flow up**: Children communicate with parents via callbacks or event systems
3. **Clean separation of concerns**: Components only know about their direct dependencies

### Lip Gloss Integration

BubblyUI leverages Lip Gloss for beautiful, consistent styling:

```go
// Example styling with Lip Gloss
func StyledBox(title, content string) *Element {
    titleStyle := lipgloss.NewStyle().
        Bold(true).
        Border(lipgloss.RoundedBorder()).
        Padding(0, 1)
    
    bodyStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#00ffff"))
    
    return Box().Children(
        Text(title).WithStyle(titleStyle),
        Text(content).WithStyle(bodyStyle),
    )
}
```

This integration provides:

1. **Declarative styling**: Define how components look with a fluent API
2. **Consistent design**: Apply themes and style rules across components
3. **Layout composition**: Easily stack, align, and arrange components with Lip Gloss utilities

## Key Features

BubblyUI's key features include:

1. **Independent Component Architecture**: Each UI element is a self-contained unit with its own state and rendering logic, breaking free from Bubble Tea's monolithic model structure.

2. **Reactive Signals**: Inspired by Solid.js, BubblyUI uses signals to create fine-grained reactivity, ensuring only components that depend on changed data re-render.

3. **Predictable Data Flow**: Props flow down from parents to children, while events flow up through callbacks, creating a clean, maintainable architecture.

4. **Lifecycle Hooks**: Components can respond to mounting, updating, and unmounting events with simple callback functions.

5. **Elegant Styling**: Deep integration with Lip Gloss for beautiful, consistent styling with a fluent API.

6. **Optimized Performance**: Minimal terminal redraws through intelligent tracking of component dependencies.

7. **Developer Experience**: Hot reload support, helpful error messages, and a familiar component model for developers coming from web frameworks.

8. **Comprehensive Testing**: Every component includes thorough tests for functionality, edge cases, and performance benchmarks.

9. **Test-Driven Development**: We follow a strict test-first approach where all components are tested before implementation is complete.

See the [ROADMAP.md](./ROADMAP.md) for our detailed implementation plan and current progress.

## Philosophy

BubblyUI was created to solve the limitations of traditional TUI frameworks, with these guiding principles:

- **Component-based architecture**: Replace monolithic models with independent, reusable components.
- **Fine-grained reactivity**: Update only what changed, not the entire screen.
- **Predictable data flow**: Props down, events up - creating clean, maintainable code.
- **Familiar mental model**: Apply React and Solid.js concepts to terminal UIs.
- **Developer experience**: Focus on writing business logic, not wrangling with layout or state management.
- **Performance-minded**: Optimize for minimal terminal redraws and efficient resource usage.
- **Test-driven development**: Every component is tested thoroughly with edge cases before implementation.
- **Quality first**: Comprehensive testing for functionality, edge cases, and performance benchmarks.

BubblyUI aims to bring the best of modern frontend development patterns to the terminal, making TUI development more accessible, maintainable, and enjoyable.

## Getting Started

1. **Clone the repository**

   ```bash
   git clone https://github.com/newbpydev/bubblyui.git
   cd bubblyui
   ```
2. **Install dependencies**

   ```bash
   go mod download
   ```
3. **Run tests**

   ```bash
   go test ./...
   ```
4. **Run an example**

   ```bash
   cd cmd/counter
   go run .
   ```
5. **Start with hot-reload**

   ```bash
   # Install Air if not already:
   go install github.com/cosmtrek/air@latest

   # From repo root:
   air -c .air.toml # configured for BubblyUI
   ```

### Creating Your First Component

Here's a simple example of creating a custom component with BubblyUI:

```go
package main

import (
    "fmt"
    
    "github.com/newbpydev/bubblyui/pkg/termidom"
    "github.com/charmbracelet/lipgloss"
)

// A simple counter component
func Counter() *termidom.Element {
    // Create local state
    count, setCount := termidom.useState(0)
    
    return termidom.Box().
        Border(lipgloss.RoundedBorder()).
        Padding(1).
        Children(
            termidom.Text(fmt.Sprintf("Count: %d", count)).Bold(true),
            termidom.Box().FlexDirection("row").Children(
                termidom.Button("−").OnClick(func() { 
                    if count > 0 {
                        setCount(count - 1)
                    }
                }),
                termidom.Button("+").OnClick(func() { 
                    setCount(count + 1)
                }),
            ),
        )
}

func main() {
    app := termidom.NewApp(
        // Root component
        termidom.Box().Title("BubblyUI Demo").Padding(2).Children(
            termidom.Text("Welcome to BubblyUI!").Italic(true),
            Counter(),
            termidom.Text("Press Ctrl+C to quit"),
        ),
    )
    
    // Start the application
    if err := app.Run(); err != nil {
        fmt.Println("Error running app:", err)
    }
}
```

## Contribution

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/YourFeature`)
3. Commit your changes (`git commit -m "Add feature"`)
4. Push to your branch (`git push origin feature/YourFeature`)
5. Open a Pull Request

Please ensure:

* Code compiles and is properly formatted (`go fmt`)
* Unit tests cover new functionality
* Documentation is updated where applicable

## License

This project is licensed under the [MIT License](LICENSE).
