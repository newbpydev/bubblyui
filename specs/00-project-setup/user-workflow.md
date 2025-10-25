# User Workflow: Project Setup

## Primary User Journey

### Journey: Developer Sets Up BubblyUI Project from Scratch

1. **Entry Point**: Developer wants to start BubblyUI development
   - System response: Clean slate, no project exists
   - Ready for: Initial setup

2. **Step 1**: Create project directory and initialize Git
   ```bash
   mkdir bubblyui
   cd bubblyui
   git init
   ```
   - System response: Git repository initialized
   - Ready for: Go module initialization

3. **Step 2**: Initialize Go module
   ```bash
   go mod init github.com/yourusername/bubblyui
   ```
   - System creates: `go.mod` with module path
   - Go version: Will be set in next steps
   - Ready for: Adding dependencies

4. **Step 3**: Set Go version and add dependencies
   ```bash
   # Edit go.mod to set Go 1.22
   go get github.com/charmbracelet/bubbletea@latest
   go get github.com/charmbracelet/lipgloss@latest
   go get github.com/stretchr/testify@latest
   ```
   - System creates: `go.sum` with checksums
   - Dependencies: Downloaded to module cache
   - Ready for: Directory structure

5. **Step 4**: Create directory structure
   ```bash
   mkdir -p pkg/bubbly pkg/components
   mkdir -p cmd/examples
   mkdir -p docs/api docs/guides
   mkdir -p specs
   mkdir -p tests/integration
   mkdir -p .github/workflows
   mkdir -p .claude/commands
   ```
   - System creates: All directories
   - Ready for: Tool configuration

6. **Step 5**: Configure quality tools
   - Create `.gitignore` (Go patterns)
   - Create `.golangci.yml` (linter config)
   - Create `.editorconfig` (editor consistency)
   - Create `Makefile` (common tasks)
   - Result: Quality tools configured

7. **Step 6**: Set up CI/CD
   - Create `.github/workflows/ci.yml`
   - Configure test, lint, build jobs
   - Result: Automated checks on push

8. **Step 7**: Create documentation files
   - Create `README.md` (project overview)
   - Create `CONTRIBUTING.md` (guidelines)
   - Create `LICENSE` (MIT)
   - Result: Documentation structure ready

9. **Step 8**: Verify setup
   ```bash
   make test      # Should pass (no tests yet)
   make lint      # Should pass (no code yet)
   make build     # Should pass
   ```
   - All commands: Execute successfully
   - Result: Setup verified

10. **Completion**: Project ready for development
    - Time taken: ~10 minutes
    - State: All tools configured, CI/CD working
    - Ready for: Feature 01 (Reactivity System) implementation

---

## Alternative Paths

### Scenario A: Clone Existing Repository

1. **Developer clones existing BubblyUI repo**
   ```bash
   git clone https://github.com/yourusername/bubblyui.git
   cd bubblyui
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Install development tools**
   ```bash
   make install-tools
   ```

4. **Verify setup**
   ```bash
   make test
   make lint
   ```

**Time:** ~2 minutes  
**Use Case:** Contributing to existing project

---

### Scenario B: Quick Start with Script

1. **Developer uses setup script (future enhancement)**
   ```bash
   curl -sSL https://bubblyui.dev/install.sh | bash
   cd bubblyui
   ```

2. **Script automatically:**
   - Initializes Go module
   - Adds dependencies
   - Creates directory structure
   - Configures tools
   - Runs initial verification

**Time:** ~1 minute  
**Use Case:** Fastest onboarding

---

### Scenario C: Docker Development Environment (future)

1. **Developer uses Docker**
   ```bash
   docker-compose up -d
   docker-compose exec dev bash
   ```

2. **Pre-configured environment includes:**
   - Go 1.22+
   - All tools installed
   - Project set up
   - Ready to code

**Use Case:** Consistent development environment

---

## Error Handling Flows

### Error 1: Go Version Too Old

**Trigger**: Developer has Go < 1.22
```bash
$ go version
go version go1.21.0 linux/amd64
```

**User sees**:
```
Error: Go 1.22 or higher required
Current version: 1.21.0
Please upgrade: https://go.dev/doc/install
```

**Recovery**:
1. Visit https://go.dev/doc/install
2. Download Go 1.22+
3. Install
4. Verify: `go version`
5. Retry setup

---

### Error 2: Dependency Resolution Failure

**Trigger**: `go get` fails due to network/proxy issues

**User sees**:
```
go: downloading github.com/charmbracelet/bubbletea v0.25.0
go: error: connection refused
```

**Recovery**:
1. Check network connection
2. Configure proxy if needed:
   ```bash
   export GOPROXY=https://proxy.golang.org,direct
   ```
3. Retry: `go mod download`

**Alternative**:
```bash
# Use vendor directory
go mod vendor
```

---

### Error 3: golangci-lint Not Found

**Trigger**: `make lint` fails

**User sees**:
```
golangci-lint: command not found
```

**Recovery**:
1. Install tool:
   ```bash
   make install-tools
   # OR
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```
2. Verify: `golangci-lint version`
3. Retry: `make lint`

---

### Error 4: Tests Fail on Windows

**Trigger**: Line ending issues

**User sees**:
```
FAIL: TestFileContent expected "\n" got "\r\n"
```

**Recovery**:
1. Configure Git:
   ```bash
   git config core.autocrlf true
   ```
2. Re-clone repository
3. Verify `.editorconfig` being used

---

### Error 5: GitHub Actions Workflow Invalid

**Trigger**: YAML syntax error

**User sees**:
```
Error: workflow file is invalid
```

**Recovery**:
1. Validate YAML:
   ```bash
   yamllint .github/workflows/ci.yml
   ```
2. Fix syntax errors
3. Commit and push
4. Check Actions tab

---

## State Transitions

### Setup States
```
Empty Directory
    ↓
Git Initialized
    ↓
Go Module Created
    ↓
Dependencies Added
    ↓
Directory Structure Created
    ↓
Tools Configured
    ↓
CI/CD Set Up
    ↓
Documentation Created
    ↓
Setup Verified
    ↓
Ready for Development
```

---

## Integration Points

### After Setup Completion
- **Feature 01** (Reactivity): Can now create `pkg/bubbly/ref.go`
- **Feature 02** (Component): Can create component interfaces
- **Feature 03** (Lifecycle): Can implement lifecycle hooks
- **Testing**: Can write first tests with testify
- **CI/CD**: Tests will run automatically

---

## Performance Expectations

### Setup Performance Targets
- **Initial setup**: <10 minutes (manual)
- **Clone + setup**: <3 minutes
- **Dependency download**: <30 seconds
- **First test run**: <5 seconds
- **First lint run**: <30 seconds
- **First build**: <10 seconds

### Developer Experience Metrics
- **Steps to first code**: <5 steps
- **Configuration complexity**: Low (sensible defaults)
- **Tool installation**: Automated (make install-tools)
- **Error messages**: Clear and actionable

---

## Common Patterns

### Pattern 1: Verify Before Proceeding
```bash
# After each major setup step
make test      # Ensure tests work
make lint      # Ensure linting works
make build     # Ensure build works
```

### Pattern 2: Incremental Setup
```bash
# Don't do everything at once
# 1. Core structure first
go mod init
mkdir -p pkg cmd

# 2. Add dependencies
go get bubbletea lipgloss

# 3. Configure tools one by one
# ... and so on
```

### Pattern 3: Use Make for Consistency
```bash
# Always use make commands
make test       # Not: go test ./...
make lint       # Not: golangci-lint run
make fmt        # Not: gofmt -s -w .
```

---

## Testing Workflow

### Verify Setup Completeness

**Test 1: Go Module**
```bash
$ go mod verify
all modules verified

$ go list -m all
github.com/yourusername/bubblyui
github.com/charmbracelet/bubbletea v0.25.0
github.com/charmbracelet/lipgloss v0.9.1
github.com/stretchr/testify v1.8.4
```

**Test 2: Directory Structure**
```bash
$ ls -la
.github/
.claude/
cmd/
docs/
pkg/
specs/
tests/
.gitignore
.golangci.yml
go.mod
go.sum
Makefile
README.md
```

**Test 3: Tools Work**
```bash
$ make test
ok      github.com/yourusername/bubblyui/pkg/bubbly

$ make lint
# No output = success

$ make build
# Build successful
```

**Test 4: CI/CD**
```bash
$ git push origin main
# Check GitHub Actions tab
# All workflows: ✅ Passing
```

---

## Documentation for Users

### Quick Setup Guide
1. **Prerequisites**: Go 1.22+, Git
2. **Clone**: `git clone ...` or create new directory
3. **Initialize**: `go mod init ...`
4. **Dependencies**: `go get ...` (see requirements.md)
5. **Structure**: Create directories (see designs.md)
6. **Configure**: Copy configuration files
7. **Verify**: Run `make test lint build`
8. **Commit**: `git add . && git commit -m "Initial setup"`

### Best Practices
- Use `make` commands for consistency
- Run tests before committing
- Keep dependencies up to date
- Follow Go conventions
- Document as you go

### Troubleshooting
- **Build fails?** Check Go version (need 1.22+)
- **Tests fail?** Ensure dependencies downloaded
- **Lint fails?** Run `make fmt` first
- **CI fails?** Check logs in GitHub Actions
