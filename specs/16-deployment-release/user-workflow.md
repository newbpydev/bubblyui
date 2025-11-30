# User Workflow: Deployment & Release Preparation

## Overview

This document describes workflows for two primary user types:
1. **Framework Users** - Developers using BubblyUI to build applications
2. **Framework Maintainers** - Developers releasing new versions of BubblyUI

---

## Part 1: Framework User Workflows

### Workflow A: Initial Installation

**Goal:** Install BubblyUI and start building a TUI application

#### Journey
1. **Discovery**
   - User finds BubblyUI via pkg.go.dev, GitHub, or community recommendation
   - User reads README.md for quick overview
   - User checks features match their needs

2. **Installation**
   - User runs: `go get github.com/newbpydev/bubblyui`
   - Go module system fetches latest stable version
   - Dependencies (Bubbletea, Lipgloss) automatically resolved

3. **First Application**
   ```go
   package main

   import (
       "fmt"
       "github.com/newbpydev/bubblyui"
   )

   func main() {
       counter, _ := bubblyui.NewComponent("Counter").
           Setup(func(ctx *bubblyui.Context) {
               count := ctx.Ref(0)
               ctx.Expose("count", count)
           }).
           Template(func(ctx bubblyui.RenderContext) string {
               return fmt.Sprintf("Count: %v", ctx.Get("count"))
           }).
           Build()

       bubblyui.Run(counter)
   }
   ```

4. **Verification**
   - User runs: `go run main.go`
   - Application starts in terminal
   - User sees TUI interface

#### State Transitions
```
No Module → go get → Module Added → Import → Build → Run
```

---

### Workflow B: Using Subpackages

**Goal:** Access additional features like router, components, composables

#### Journey
1. **Need Identification**
   - User needs routing for multi-screen app
   - User checks documentation for router usage

2. **Import Subpackage**
   ```go
   import (
       "github.com/newbpydev/bubblyui"
       "github.com/newbpydev/bubblyui/pkg/bubbly/router"
       "github.com/newbpydev/bubblyui/pkg/components"
   )
   ```

3. **Use Features**
   ```go
   // Use router
   r := router.New()
   r.Route("/home", HomeComponent)
   r.Route("/settings", SettingsComponent)

   // Use built-in components
   btn := components.NewButton("Click Me")
   input := components.NewInput("Enter name")
   ```

4. **Build Application**
   - Combine core features with subpackages
   - Application compiles and runs

---

### Workflow C: Version Upgrade

**Goal:** Upgrade BubblyUI to newer version

#### Journey
1. **Check for Updates**
   - User visits GitHub releases or pkg.go.dev
   - User reads CHANGELOG.md for changes since current version

2. **Review Changes**
   ```markdown
   ## [0.12.0] - 2025-11-30
   ### Added
   - Root package exports for cleaner imports
   - GoReleaser configuration
   ### Changed
   - Recommended import path is now `github.com/newbpydev/bubblyui`
   ```

3. **Upgrade**
   - User runs: `go get github.com/newbpydev/bubblyui@v0.12.0`
   - Or updates go.mod manually

4. **Verify**
   - User runs: `go test ./...`
   - User runs application to verify functionality
   - User updates imports if desired (optional for v0.x)

#### Error Handling
| Error | Cause | Recovery |
|-------|-------|----------|
| `module not found` | Network issue | Retry or check GOPROXY |
| `version not found` | Invalid version | Check available versions |
| `import cycle` | Code issue | Review import structure |

---

### Workflow D: Checking Version History

**Goal:** Understand what changed between versions

#### Journey
1. **Access CHANGELOG**
   - GitHub: `CHANGELOG.md` in repository root
   - Direct: `https://github.com/newbpydev/bubblyui/blob/main/CHANGELOG.md`

2. **Navigate to Version**
   - Find section for version of interest
   - Read Added/Changed/Fixed/Removed sections

3. **Check Breaking Changes**
   - Look for "Breaking" or "BREAKING" annotations
   - Review migration notes if present

4. **Plan Upgrade**
   - Assess impact on current code
   - Schedule upgrade based on risk level

---

## Part 2: Framework Maintainer Workflows

### Workflow E: Preparing a Release

**Goal:** Prepare and publish a new version of BubblyUI

#### Journey

##### Phase 1: Pre-Release Validation
1. **Ensure Tests Pass**
   ```bash
   make test-race
   # Expected: All tests pass, 0 failures

   make lint
   # Expected: 0 warnings, 0 errors

   make build
   # Expected: Build succeeds
   ```

2. **Verify Coverage**
   ```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -func=coverage.out | grep total
   # Expected: >80% coverage
   ```

3. **Review Changes**
   ```bash
   git log --oneline $(git describe --tags --abbrev=0)..HEAD
   # Review all commits since last release
   ```

##### Phase 2: Documentation Update
4. **Update CHANGELOG**
   - Add new version section with date
   - Categorize changes: Added, Changed, Fixed, etc.
   - Note any breaking changes prominently

5. **Update README** (if needed)
   - Update feature list
   - Update examples
   - Verify badges are correct

6. **Commit Documentation**
   ```bash
   git add CHANGELOG.md README.md
   git commit -m "docs: prepare v0.12.0 release"
   ```

##### Phase 3: Tagging
7. **Create Tag**
   ```bash
   # Verify you're on main branch
   git checkout main
   git pull origin main

   # Create annotated tag
   git tag -a v0.12.0 -m "Release v0.12.0: Deployment & Release"

   # Push tag
   git push origin v0.12.0
   ```

##### Phase 4: Automated Release
8. **GitHub Actions Triggered**
   - Workflow runs automatically on tag push
   - Tests run with race detector
   - Coverage verified
   - GoReleaser creates GitHub Release
   - Release notes generated from commits

9. **Verify Release**
   - Check GitHub Releases page
   - Verify release notes are correct
   - Check pkg.go.dev shows new version (may take ~1 hour)

#### State Transitions
```
Development → Tests Pass → Docs Updated → Tagged → CI Runs → Released
     ↓             ↓            ↓           ↓          ↓
  Fix Issues   Fix Tests   Edit Docs    Retag    Fix CI
```

---

### Workflow F: Hotfix Release

**Goal:** Quickly fix a critical bug in production

#### Journey
1. **Identify Issue**
   - Bug report received
   - Impact assessed as critical

2. **Create Fix**
   ```bash
   git checkout -b hotfix/fix-critical-bug
   # Make minimal fix
   # Write regression test
   git commit -m "fix: resolve critical bug in component lifecycle"
   ```

3. **Fast-Track Review**
   - PR created with "hotfix" label
   - Expedited review process

4. **Merge and Release**
   ```bash
   git checkout main
   git merge hotfix/fix-critical-bug
   git tag -a v0.11.1 -m "Hotfix: Critical bug fix"
   git push origin main --tags
   ```

5. **Communicate**
   - Update CHANGELOG with fix details
   - Notify users who reported issue
   - Consider GitHub Security Advisory if security-related

---

### Workflow G: Creating Retroactive Tags

**Goal:** Add version tags for previously released features

#### Journey
1. **Identify Commits**
   ```bash
   # Find commit where feature was completed
   git log --oneline --all | grep "feat: complete feature 01"
   ```

2. **Create Historical Tags**
   ```bash
   # Tag specific commits
   git tag -a v0.2.0 <commit-sha> -m "Release v0.2.0: Reactivity System"
   git tag -a v0.3.0 <commit-sha> -m "Release v0.3.0: Component Model"
   # ... continue for all features
   ```

3. **Push All Tags**
   ```bash
   git push origin --tags
   ```

4. **Verify on pkg.go.dev**
   - Each version should appear
   - Documentation should be indexed

---

## Error Handling Flows

### Error 1: CI Fails on Release

**Trigger:** Tag pushed but CI workflow fails

**User Sees:**
- GitHub Actions shows failed workflow
- No GitHub Release created

**Recovery:**
1. Identify failure cause from CI logs
2. Fix issue in codebase
3. Delete failed tag: `git push origin --delete v0.12.0`
4. Delete local tag: `git tag -d v0.12.0`
5. Fix and re-commit
6. Create new tag and push

---

### Error 2: pkg.go.dev Not Updating

**Trigger:** Release created but pkg.go.dev doesn't show new version

**User Sees:**
- Old version listed on pkg.go.dev
- `go get @latest` doesn't fetch new version

**Recovery:**
1. Wait up to 1 hour (indexing delay)
2. Manually trigger: `curl https://proxy.golang.org/github.com/newbpydev/bubblyui/@v/v0.12.0.info`
3. If still not working, check Go proxy status
4. Verify tag exists on GitHub

---

### Error 3: Breaking Change in Minor Version

**Trigger:** Users report breaking changes in v0.x.x upgrade

**User Sees:**
- Compilation errors after upgrade
- Unexpected behavior changes

**Recovery:**
1. Document breaking changes in CHANGELOG immediately
2. Create migration guide
3. Consider releasing patch with backward compatibility
4. For v0.x, this is acceptable per semver

---

## Integration Points

### Connected Systems

| System | Integration Point |
|--------|-------------------|
| GitHub | Repository, Releases, Actions |
| pkg.go.dev | Module discovery, documentation |
| proxy.golang.org | Module distribution |
| sum.golang.org | Checksum verification |
| Codecov (optional) | Coverage reporting |

### Data Flow Between Features

```
Feature Development
        ↓
Merge to main
        ↓
CHANGELOG Updated
        ↓
Version Tagged ───────────────────────┐
        ↓                             │
GitHub Actions ←──────────────────────┘
        ↓
GoReleaser
        ↓
GitHub Release ──→ Users (GitHub)
        ↓
proxy.golang.org ──→ Users (go get)
        ↓
pkg.go.dev ──→ Users (documentation)
```

---

## State Transitions Summary

### Version Lifecycle
```
Unreleased → Tagged → CI Passed → Released → Indexed
                ↓
            CI Failed → Fixed → Retagged
```

### User Adoption
```
Discovery → Installation → Development → Upgrade → Feedback
                                ↓
                         Report Issue → Fix → New Release
```

---

## Validation Checklist

### Pre-Release
- [ ] All tests pass with race detector
- [ ] Coverage >80%
- [ ] Lint warnings resolved
- [ ] CHANGELOG updated
- [ ] README accurate
- [ ] Examples compile and run
- [ ] Breaking changes documented

### Post-Release
- [ ] GitHub Release created
- [ ] Release notes correct
- [ ] pkg.go.dev indexed (within 1 hour)
- [ ] `go get @v0.12.0` works
- [ ] Example applications build
- [ ] No urgent bug reports
