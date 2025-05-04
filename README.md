# BubblyUI

BubblyUI is an open-source Go framework for building declarative, component-based terminal UIs on top of Bubble Tea and Lip Gloss. It provides a virtual DOM diffing layer, a fluent builder API, lifecycle hooks, and hot-reload support to streamline TUI development.

## Technology Stack

* **Language:** Go (1.24+)
* **TUI Core:** [Bubble Tea](https://github.com/charmbracelet/bubble-tea)
* **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
* **Hot-Reload:** [Air](https://github.com/cosmtrek/air) (or similar file-watcher)
* **Testing:** Go’s built-in `testing` package
* **CI/CD:** GitHub Actions

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

BubblyUI wraps the Elm-style `Init/Update/View` cycle of Bubble Tea into declarative, stateful components. It introduces:

* A **virtual DOM** (`VNode`, `Element`) for diff-and-patch updates.
* A **fluent builder API** powered by Go generics for defining components.
* **Lifecycle hooks** (`OnMount`, `OnUpdate`, `OnUnmount`) as simple callbacks.
* **Per-component styling** props using Lip Gloss.
* **Hot-reload** support for rapid iteration.
* Core components: `Box`, `Text`, `List`, `Button`, and theming support.

## Architecture

BubblyUI reimagines React's architecture for terminal UIs, implementing key React concepts:

### Virtual DOM

At the core of BubblyUI is a lightweight virtual DOM implementation optimized for terminal rendering:

```go
// Simplified VNode structure
type VNode struct {
    Type      string
    Props     map[string]interface{}
    Children  []*VNode
    Component Component
    Key       string
}
```

The virtual DOM allows BubblyUI to:

1. **Declaratively define UI**: Create nested component trees that describe what the UI should look like
2. **Efficiently update**: Only re-render components that actually changed
3. **Abstract the terminal**: Work with components instead of directly manipulating terminal output

### Component System

Components in BubblyUI follow the React pattern but are adapted for Go:

```go
// Example component definition (simplified)
func Counter() *Element {
    count, setCount := useState(0)
    
    return Box().Padding(1).Border(lipgloss.RoundedBorder()).Children(
        Text(fmt.Sprintf("Count: %d", count)),
        Button("Increment").OnClick(func() { setCount(count + 1) }),
    )
}
```

Components become reusable, composable building blocks with isolated state and behavior.

### Reconciliation

When state changes, BubblyUI:

1. Creates a new virtual DOM tree
2. Diffs it against the previous tree
3. Generates minimal update instructions
4. Applies these updates to the Bubble Tea model
5. Renders only what changed

This approach provides React-like performance with minimal terminal redrawing.

### Hooks and State

BubblyUI implements React-like hooks for state management and side effects:

```go
// Example hooks (simplified API)
count, setCount := useState(0)
onMount(func() { /* component mounted */ })
onUpdate(func() { /* component updated */ })
```

These allow for functional, stateful components without class-based inheritance.

## Detailed Roadmap

| Phase | Objectives | Deliverables | Challenges | Status |
| ----- | ---------- | ------------ | ---------- | ------ |
| 1 | **Foundation & Requirements** | README, module initialization, project structure | Managing project scope, defining clear architecture vision | ✅ Completed |
| 2 | **Virtual DOM Architecture** <br>- Define core type interfaces<br>- Design component contract<br>- Create VNode structure with props | - `internal/dom/types.go`<br>- `pkg/termidom/element.go`<br>- `pkg/termidom/component.go` | - Designing a Go-idiomatic component API<br>- Balancing flexibility with type safety | 🔲 Pending |
| 3 | **Virtual DOM Implementation** <br>- Create node creation functions<br>- Implement reconciliation algorithm<br>- Design efficient property diffing | - `internal/dom/vnode.go`<br>- `internal/dom/reconciler.go`<br>- `internal/dom/diff.go` | - Adapting React's reconciliation for terminal UI<br>- Handling terminal-specific constraints<br>- Optimizing for minimal redraws | 🔲 Pending |
| 4 | **Bubble Tea Integration** <br>- Create wrapper for Bubble Tea program<br>- Design message passing system<br>- Implement render cycle | - `internal/core/program.go`<br>- `internal/core/messages.go`<br>- `internal/core/render.go` | - Converting VDOM patches to Bubble Tea model updates<br>- Managing state consistently across updates<br>- Ensuring performant rendering | 🔲 Pending |
| 5 | **Component System** <br>- Create builder API with fluent interface<br>- Implement key primitive components<br>- Design props system | - `pkg/termidom/builder.go`<br>- Core components (`Box`, `Text`, etc.)<br>- Props interfaces | - Creating an ergonomic builder API with generics<br>- Balancing component complexity<br>- Making components composable | 🔲 Pending |
| 6 | **State Management & Hooks** <br>- Implement component lifecycle hooks<br>- Create useState-like functionality<br>- Design context system | - Hook implementations<br>- State management utilities<br>- Context provider/consumer | - Managing state scopes in Go<br>- Designing lifecycles without closures<br>- Implementing React-like hooks in Go | 🔲 Pending |
| 7 | **Styling & Theming** <br>- Design theme context system<br>- Implement style inheritance<br>- Create responsive layouts | - Theme provider<br>- Style system<br>- Layout components | - Creating a consistent styling API<br>- Handling terminal size constraints<br>- Managing style inheritance | 🔲 Pending |
| 8 | **Developer Experience** <br>- Implement hot-reload<br>- Create debugging tools<br>- Design error boundaries | - Air configuration<br>- Debug utilities<br>- Hot-reload examples | - Implementing hot-reload for stateful apps<br>- Preserving state across reloads<br>- Creating useful error information | 🔲 Pending |
| 9 | **Testing & Quality** <br>- Implement unit tests for core logic<br>- Create component testing utilities<br>- Set up CI/CD pipeline | - Test suite<br>- Testing utilities<br>- GitHub Actions workflow | - Testing terminal UI components<br>- Simulating user interactions<br>- Creating reproducible tests | 🔲 Pending |
| 10 | **Documentation & Release** <br>- Create comprehensive docs<br>- Build example applications<br>- Package for distribution | - API documentation<br>- Example applications<br>- v0.1 release | - Creating clear, concise documentation<br>- Designing example apps that showcase capabilities<br>- Ensuring backward compatibility | 🔲 Pending |

## Philosophy

BubblyUI was created with several guiding principles:

- **Declarative over imperative**: Define what your UI should look like, not how to update it.
- **Component-based**: Build UIs from small, reusable pieces that manage their own state.
- **Learn once, write anywhere**: Leverage familiar React concepts in terminal applications.
- **Developer experience**: Provide helpful error messages, hot-reloading, and a cohesive API.
- **Performance-minded**: Minimize terminal redraws and optimize rendering for resource efficiency.

BubblyUI aims to bring modern frontend development patterns to the terminal, making TUI development more accessible and maintainable.

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
3. **Run an example**

   ```bash
   cd cmd/counter
   go run .
   ```
4. **Start with hot-reload**

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
