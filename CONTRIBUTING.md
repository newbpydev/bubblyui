# Contributing to BubblyUI

Thank you for your interest in contributing to BubblyUI! This document provides guidelines and instructions for contributing to this project.

## Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## How Can I Contribute?

### Reporting Bugs

This section guides you through submitting a bug report. Following these guidelines helps maintainers understand your report, reproduce the behavior, and find related reports.

- **Use the GitHub issue tracker** — Check if the bug has already been reported by searching GitHub [Issues](https://github.com/username/bubblyui/issues).
- **Use a clear and descriptive title** for the issue.
- **Describe the exact steps to reproduce the problem** in as much detail as possible.
- **Provide specific examples** to demonstrate the steps.
- **Describe the behavior you observed after following the steps** and point out what exactly is the problem with that behavior.
- **Explain which behavior you expected to see instead and why.**
- **Include screenshots and animated GIFs** where possible.
- **Include details about your environment** - Go version, OS, etc.

### Suggesting Enhancements

This section guides you through submitting an enhancement suggestion, including completely new features and minor improvements to existing functionality.

- **Use the GitHub issue tracker** — Check if the enhancement has already been suggested by searching GitHub [Issues](https://github.com/username/bubblyui/issues).
- **Use a clear and descriptive title** for the issue.
- **Provide a step-by-step description of the suggested enhancement** in as much detail as possible.
- **Provide specific examples to demonstrate the steps**. Include copy/pasteable snippets which you use in those examples, as Markdown code blocks.
- **Describe the current behavior** and **explain which behavior you expected to see instead** and why.
- **Explain why this enhancement would be useful** to most BubblyUI users.

### Pull Requests

- **Fill in the required template** (if applicable)
- **Do not include issue numbers in the PR title**
- **Follow the Go style guide**
- **Include appropriate tests** - Make sure to include tests for any bugs fixed or features added.
- **Document new code** - Add docstrings for new functions and comments for complex sections of code.
- **End all files with a newline**
- **Avoid platform-dependent code**

## Development Workflow

### Setting Up Development Environment

1. Fork the repository
2. Clone your fork: `git clone https://github.com/yourusername/bubblyui.git`
3. Add the original repository as upstream: `git remote add upstream https://github.com/username/bubblyui.git`
4. Run `go mod download` to install dependencies

### Development Process

1. Create a new branch: `git checkout -b feature/your-feature-name`
2. Make your changes
3. Run tests: `go test ./... -race -cover`
4. Run linters: `./scripts/lint.sh`
5. Commit your changes: `git commit -m "Add feature X"`
6. Push to your fork: `git push origin feature/your-feature-name`
7. Create a Pull Request

### Coding Standards

- Follow Go best practices and style guidelines.
- All code must be properly formatted using `go fmt`.
- All code must pass `go vet` without issues.
- Write clear, readable, and maintainable code.
- Ensure proper error handling.
- Keep functions and methods small and focused on a single responsibility.
- Write comprehensive tests for new features.

### Commit Messages

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line

## Project Structure

Please refer to the [README.md](README.md) for an overview of the project structure.

## License

By contributing to BubblyUI, you agree that your contributions will be licensed under the project's [MIT License](LICENSE).
