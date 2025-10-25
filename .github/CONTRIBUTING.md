# Contributing to BubblyUI

## Development Setup

1. **Prerequisites**: Go 1.22+, Git
2. **Clone**: `git clone https://github.com/yourusername/bubblyui.git`
3. **Dependencies**: `go mod download`
4. **Tools**: `make install-tools`
5. **Verify**: `make test lint build`

## Development Workflow

### Follow Ultra-Workflow
Use `.claude/commands/ultra-workflow.md` for all feature work:
1. **Understand** - Read ALL spec files in feature directory
2. **Gather** - Use context7 for Go/Bubbletea patterns
3. **Plan** - Create actionable task list
4. **TDD** - Red-Green-Refactor with table-driven tests
5. **Focus** - Verify alignment with specs
6. **Cleanup** - Run all quality gates
7. **Document** - Update specs and docs

### Quality Gates (Must Pass)
```bash
make test-race  # Tests with race detector
make lint       # golangci-lint (zero warnings)
make fmt        # gofmt + goimports
make build      # Verify build succeeds
```

## Code Standards

### Go Idioms
- Accept interfaces, return structs
- Error wrapping: `fmt.Errorf("context: %w", err)`
- Short variable names in limited scope
- Godoc comments on all exports

### Type Safety
- Use generics: `Ref[T]`, `Computed[T]`
- Avoid `any` without constraints
- Type assertions with safety checks

### Testing
- Table-driven tests co-located with source
- Use testify: `assert.Equal(t, expected, got)`
- Test behavior, not implementation
- Target: >80% coverage
- Always run race detector

### Bubbletea Integration
- Follow Model/Update/View pattern
- Messages for state changes
- Commands for async operations
- No direct goroutines

## Pull Request Process

1. **Branch**: `feature/your-feature` or `fix/bug-description`
2. **Commits**: Clear, atomic commits
3. **Tests**: All tests pass, coverage maintained
4. **Lint**: Zero warnings from golangci-lint
5. **Docs**: Update specs/tasks.md, godoc, README if needed
6. **PR**: Clear description, reference issues

## PR Template
- **What**: Brief description of changes
- **Why**: Problem being solved
- **How**: Implementation approach
- **Testing**: How it was tested
- **Checklist**: Tests pass, lint clean, docs updated

## Questions?
- Check `RESEARCH.md` for design decisions
- Read feature specs in `specs/` directory
- Review `.rules` and `CLAUDE` files
- Open a discussion for clarification
