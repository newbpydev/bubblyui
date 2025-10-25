# Contributing to BubblyUI

A comprehensive guide for contributors following professional open source practices.

## 🌟 Quick Start

### 1. Fork and Clone
```bash
git clone https://github.com/yourusername/bubblyui.git
cd bubblyui
git remote add upstream https://github.com/originalowner/bubblyui.git
```

### 2. Set Up Development Environment
```bash
# Install Go 1.22+
go version  # Must be 1.22.0 or later

# Install development tools
make install-tools

# Run tests to verify setup
make test lint build
```

### 3. Create Feature Branch
```bash
git checkout main
git pull upstream main
git checkout -b feature/your-feature-name
```

### 4. Make Changes
- Read relevant specs in `specs/XX-feature/` directory
- Follow TDD: write failing tests first
- Implement minimal code to pass tests
- Run quality gates: `make test-race lint fmt build`

### 5. Submit Pull Request
- Push branch: `git push origin feature/your-feature-name`
- Open PR with clear description
- Reference any related issues
- Ensure all CI checks pass

---

## 📋 Development Workflow

### Follow the Ultra-Workflow (7 Phases)

All feature development must follow this systematic approach:

1. **🎯 Understand** - Read ALL specification files in feature directory
2. **🔍 Gather** - Research Go/Bubbletea patterns using available tools
3. **📝 Plan** - Create actionable task breakdown with sequential thinking
4. **🧪 TDD** - Red-Green-Refactor with table-driven tests
5. **🎯 Focus** - Verify alignment with specifications and integration
6. **🧹 Cleanup** - Run all quality gates and validation
7. **📚 Document** - Update specs, godoc, README, and CHANGELOG

### Quality Gates (Mandatory)
All contributions must pass these automated checks:

```bash
make test-race    # Tests with race detector (must pass)
make lint         # golangci-lint (zero warnings allowed)
make fmt          # gofmt + goimports (must be clean)
make build        # Compilation (must succeed)
go test -cover    # >80% coverage (must maintain)
go test ./...     # Integration tests (must pass)
```

**No pull request will be merged without passing all quality gates.**

---

## 🏗️ Code Standards

### Go Style Guide Compliance
Follow [Google's Go Style Guide](https://google.github.io/styleguide/go/guide) principles:

#### Type Safety
- ✅ Use generics: `Ref[T]`, `Computed[T]`, `func Map[T, U any](...)`
- ❌ Never use `any` without constraints: `func Process(data any)` → `func Process[T fmt.Stringer](data T)`
- ✅ Type assertions with safety checks: `if val, ok := i.(string); ok { ... }`

#### Error Handling
- ✅ Wrap errors with context: `return fmt.Errorf("failed to process: %w", err)`
- ❌ Never ignore errors: `result := doSomething()` → `result, err := doSomething(); if err != nil { ... }`
- ✅ Use idiomatic patterns: `if err := doSomething(); err != nil { ... }`

#### Testing
- ✅ Table-driven tests: Use `[]struct{ name, input, expected }` pattern
- ✅ Co-locate tests: `feature.go` + `feature_test.go`
- ✅ Test behavior, not implementation details
- ✅ Use testify assertions: `assert.Equal(t, expected, got)`
- ✅ Include edge cases: empty inputs, nil pointers, concurrent access

#### Documentation
- ✅ Godoc comments on all exported functions/types
- ✅ Runnable examples in godoc where appropriate
- ✅ Clear parameter and return value descriptions

### Bubbletea Integration Standards
- ✅ Implement `tea.Model` interface: `Init() tea.Cmd`, `Update(tea.Msg) (tea.Model, tea.Cmd)`, `View() string`
- ✅ Use messages for state changes: `type DataLoadedMsg struct { data interface{} }`
- ✅ Use commands for async: `return tea.Cmd(func() tea.Msg { return fetchData() })`
- ❌ Never create goroutines directly in components
- ❌ Never block in `Update()` method

---

## 🔄 Pull Request Process

### Before Opening PR

1. **Run Quality Gates**: Ensure all automated checks pass
2. **Test Thoroughly**: Include unit, integration, and edge case tests
3. **Update Documentation**: Specs, godoc, examples, README if needed
4. **Clean History**: Use `git rebase -i` for clean commit history
5. **Self-Review**: Use the systematic code review checklist

### PR Template Requirements

**Title Format:**
- Features: `feat: implement reactive references system`
- Fixes: `fix: resolve race condition in component updates`
- Documentation: `docs: update API reference with examples`
- Refactoring: `refactor: optimize rendering performance`

**Description Must Include:**
- **What**: Brief description of changes
- **Why**: Problem being solved or improvement made
- **How**: Implementation approach and key decisions
- **Testing**: How changes were tested and validated
- **Breaking Changes**: Any breaking changes and migration path

### During Code Review

1. **Automated Checks**: CI must pass all quality gates
2. **Manual Review**: Follow 5-phase systematic review process
3. **Integration Testing**: Verify changes work with existing features
4. **Documentation Review**: Check all docs updated appropriately

### After Approval

1. **Merge**: Maintainers will merge after all checks pass
2. **Cleanup**: Delete feature branch after merge
3. **Follow-up**: Address any post-merge issues promptly

---

## 🐛 Issue Reporting

### Bug Reports
Use the [Bug Report Template](.github/ISSUE_TEMPLATE/bug_report.md) with:

- **Clear Description**: What you expected vs. what happened
- **Reproduction Steps**: Minimal code to reproduce the issue
- **Environment Details**: Go version, OS, BubblyUI version
- **Error Output**: Complete error messages and stack traces

### Feature Requests
Use the [Feature Request Template](.github/ISSUE_TEMPLATE/feature_request.md) with:

- **Problem Statement**: What problem does this solve?
- **Proposed Solution**: How should it work?
- **Example Usage**: Code examples showing the feature
- **Alternatives**: Other solutions considered
- **Framework Alignment**: How it fits with BubblyUI's goals

---

## 📚 Documentation Standards

### When to Update Documentation

**Always update when:**
- Adding new public APIs or features
- Changing existing API behavior
- Modifying framework architecture
- Adding examples or tutorials

**Files to Update:**
- `specs/XX-feature/tasks.md` - Mark tasks complete with notes
- `README.md` - Update if new features exposed to users
- `CHANGELOG.md` - Add entries for changes
- `pkg/bubbly/` - Add godoc comments to new exports
- `examples/` - Add examples for new features

### Documentation Quality
- **Clarity**: Use clear, concise language
- **Examples**: Include runnable code examples
- **Context**: Explain why decisions were made
- **Integration**: Show how features work together

---

## 🔒 Security Considerations

### Reporting Security Issues
- **Do not** open public issues for security vulnerabilities
- **Email**: security@bubblyui.org with details
- **Response**: Expect acknowledgment within 48 hours
- **Process**: Follow responsible disclosure practices

### Secure Development Practices
- **Input Validation**: Validate all user inputs
- **Error Handling**: Don't expose internal system details
- **Dependencies**: Audit third-party packages
- **Testing**: Include security test cases

---

## 🤝 Community Guidelines

### Communication
- **Respectful**: Treat all contributors with respect
- **Inclusive**: Welcome diverse perspectives and experiences
- **Constructive**: Focus on solutions, not problems
- **Clear**: Use clear language and provide context

### Collaboration
- **Issues**: Use issues for discussions and questions
- **Discussions**: Use GitHub Discussions for broader topics
- **Reviews**: Provide specific, actionable feedback
- **Help**: Help others learn and contribute

### Recognition
- **Attribution**: Credit contributors appropriately
- **Thanks**: Acknowledge good contributions
- **Learning**: Help others learn from the codebase

---

## 📞 Getting Help

### Where to Ask Questions

1. **Issues**: For bugs, feature requests, or questions
2. **Discussions**: For broader topics and community discussions
3. **Documentation**: Check specs/, README, and examples first
4. **Community**: Engage with other contributors respectfully

### What to Include in Questions

- **Context**: What you're trying to achieve
- **Problem**: Specific issue or blocker
- **Attempts**: What you've already tried
- **Environment**: Go version, OS, relevant setup details

---

## 🎯 Project Goals

BubblyUI aims to be:
- **Type-Safe**: Compile-time safety with generics
- **Reactive**: Functional reactive programming patterns
- **Intuitive**: Vue-inspired API for Go developers
- **Performant**: Efficient rendering and state management
- **Well-Tested**: Comprehensive test coverage
- **Well-Documented**: Clear specs and examples

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

Thank you for contributing to BubblyUI! Your contributions help make this a better framework for the Go community.
