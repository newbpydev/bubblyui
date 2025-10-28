# BubblyUI Gemini Assistant Context

This document provides context for the Gemini AI assistant to effectively help with development tasks in the BubblyUI project.

## Project Overview

BubblyUI is a Go framework for building terminal user interfaces (TUIs). It is inspired by the component-based architecture and reactivity system of Vue.js, and it is built on top of the Bubbletea library.

The core features of BubblyUI include:

*   **Type-Safe Reactivity:** A system of reactive references (`Ref[T]`), computed values (`Computed[T]`), and watchers (`Watch`) that automatically track dependencies and update the UI when state changes.
*   **Component System:** A Vue-inspired component model that allows developers to create reusable UI elements with lifecycle hooks (`onMounted`, `onUpdated`, `onUnmounted`).
*   **Composition API:** Advanced patterns for code organization and reuse, including composables and a provide/inject system.
*   **Template System:** A type-safe rendering system that uses Go functions instead of string templates, with built-in directives for conditional and list rendering.

## Building and Running

The project uses a `Makefile` to automate common development tasks.

*   **Build the project:**
    ```bash
    make build
    ```
*   **Run the examples:**
    The examples are located in the `cmd/examples` directory. To run an example, navigate to its directory and use `go run`. For example, to run the `reactive-counter` example:
    ```bash
    go run cmd/examples/01-reactivity-system/reactive-counter/main.go
    ```
*   **Run tests:**
    ```bash
    make test
    ```
*   **Run tests with the race detector:**
    ```bash
    make test-race
    ```
*   **Run tests with code coverage:**
    ```bash
    make test-cover
    ```

## Development Conventions

The project follows standard Go conventions and has a set of quality gates to ensure code quality.

*   **Code Style:** The project uses `gofmt` and `goimports` for code formatting. Run `make fmt` and `make imports` to format the code.
*   **Linting:** The project uses `golangci-lint` for linting. Run `make lint` to check for linting errors.
*   **Testing:** The project uses the standard Go testing library and `testify` for assertions. All new features should be accompanied by tests.
*   **Contribution:** Contributions are welcome. Please follow the guidelines in `CONTRIBUTING.md`.
