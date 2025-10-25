# AI Agent Configuration

## Primary Agents

### Implementation Agent
**Role:** Feature implementation following TDD  
**Skills:** tdd-workflow, go-idioms, bubbletea-integration  
**Rules:**
- MUST read all spec files before starting
- MUST follow ultra-workflow 7 phases
- MUST write tests first (Red-Green-Refactor)
- MUST maintain >80% coverage
- MUST use type-safe generics

### Review Agent
**Role:** Code review and quality assurance  
**Skills:** code-review, go-linting  
**Rules:**
- Check Go idioms and conventions
- Verify test coverage
- Ensure type safety
- Validate Bubbletea integration
- Check performance implications

### Documentation Agent
**Role:** Update docs and specs  
**Skills:** documentation-update  
**Rules:**
- Update specs/tasks.md as tasks complete
- Add godoc comments to all exports
- Update README when features added
- Keep examples current

## Agent Behaviors

### When Reading Specs
- Read ALL files in spec directory (requirements, designs, workflows, tasks)
- Pattern: `specs/XX-feature-name/*.md` → read all 4 files
- Never skip: NO EXCEPTIONS to reading full context

### When Implementing
- Phase 1: Understand (read specs, gather context)
- Phase 2: Gather (use context7 for Go/Bubbletea patterns)
- Phase 3: Plan (create todo with sequential thinking)
- Phase 4: TDD (Red-Green-Refactor)
- Phase 5: Focus checks (verify alignment)
- Phase 6: Cleanup (test, lint, format, build)
- Phase 7: Documentation (update all docs)

### When Testing
- Table-driven tests co-located with source
- Test behavior, not implementation
- Use testify for assertions
- Run with race detector: `go test -race`
- Verify coverage: `go test -cover`

### Quality Gates
- Tests: `make test-race` (must pass)
- Lint: `make lint` (zero warnings)
- Format: `make fmt` (gofmt + goimports)
- Build: `make build` (must succeed)
- Coverage: >80% required

## Context Sources
- **Specs**: Primary truth in `specs/` directory
- **RESEARCH.md**: Framework design decisions
- **ultra-workflow.md**: Development process
- **.rules**: Core principles
- **context7**: External library docs (Bubbletea, Lipgloss, Go)

## Skill Usage
- **tdd-workflow**: For implementing features with tests
- **go-idioms**: For idiomatic Go code
- **bubbletea-integration**: For Bubbletea Model/Update/View
- **code-review**: For quality checks
- **documentation-update**: For keeping docs current

## Never
- Skip reading spec files
- Skip tests or lower coverage
- Ignore linter warnings
- Use `any` without constraints
- Create goroutines outside Bubbletea
- Commit without running quality gates
