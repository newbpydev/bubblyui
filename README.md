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
termidom/                   # repository root
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

## Goals Roadmap

| Phase | Objectives                                                                | Deliverables                           | Status      |
| ----- | ------------------------------------------------------------------------- | -------------------------------------- | ----------- |
| 1     | Clarify requirements & setup module                                       | README, module initialization          | ✅ Completed |
| 2     | Scaffold packages & define core types (`VNode`, `Element`, `Component`)   | Folder structure, type definitions     | 🔲 Pending  |
| 3     | Implement VDOM diff/patch logic                                           | `dom/diff.go`, `dom/vnode.go`          | 🔲 Pending  |
| 4     | Integrate with Bubble Tea: wrap `Program` & dispatch patches via messages | `core/program.go`, `core/messages.go`  | 🔲 Pending  |
| 5     | Build builder API & core components (Box, Text, Button, List)             | `pkg/termidom/builder.go`, components  | 🔲 Pending  |
| 6     | Add lifecycle hooks & state management with generics                      | Hook implementations, useState example | 🔲 Pending  |
| 7     | Theming & styling per component                                           | Theme context & style props            | 🔲 Pending  |
| 8     | Hot-reload support & examples                                             | Air config & examples/hotreload        | 🔲 Pending  |
| 9     | Testing & CI (unit tests, integration tests, GH Actions)                  | Tests, GitHub Actions workflow         | 🔲 Pending  |
| 10    | Documentation, examples, first alpha release                              | Docs, examples, v0.1 tag               | 🔲 Pending  |

## Getting Started

1. **Clone the repository**

   ```bash
   git clone https://github.com/your-org/bubblyui.git
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
