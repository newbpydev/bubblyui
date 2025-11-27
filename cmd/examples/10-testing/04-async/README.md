# GitHub Dashboard - Async Operations Example

A real-world example demonstrating BubblyUI's async operations, composables, and the new **UseTheme/ProvideTheme** automation pattern.

## Overview

This example showcases:
- **Async data fetching** with `bubbly.Run()` for GitHub repositories and activity
- **Composable pattern** with `UseGitHubDashboard` for testable business logic
- **Theme system automation** with `UseTheme`/`ProvideTheme` (NEW!)
- **Component composition** with RepoList and ActivityFeed child components
- **Auto-refresh** functionality with periodic data updates
- **Comprehensive testing** with testutil harness

## Running the Example

```bash
go run ./cmd/examples/10-testing/04-async
```

**Controls:**
- `r` - Refresh data manually
- `ctrl+c` - Quit

## Architecture

```
GitHubDashboard (App)
├── ProvideTheme(DefaultTheme)  ← Parent provides theme
├── UseGitHubDashboard          ← Composable for async logic
├── RepoList Component
│   └── UseTheme(DefaultTheme)  ← Child uses theme (2 lines!)
└── ActivityFeed Component
    └── UseTheme(DefaultTheme)  ← Child uses theme (2 lines!)
```

## Theme System Migration (Feature 13)

This example was migrated from manual inject/expose to the new UseTheme/ProvideTheme pattern:

### Before (39 lines of boilerplate):

**app.go** - Parent provides 5 separate colors:
```go
ctx.Provide("primaryColor", lipgloss.Color("35"))
ctx.Provide("secondaryColor", lipgloss.Color("99"))
ctx.Provide("mutedColor", lipgloss.Color("240"))
ctx.Provide("warningColor", lipgloss.Color("220"))
ctx.Provide("errorColor", lipgloss.Color("196"))
```

**repo_list.go** - Child injects 3 colors (15 lines):
```go
primaryColor := lipgloss.Color("35")
if injected := ctx.Inject("primaryColor", nil); injected != nil {
    primaryColor = injected.(lipgloss.Color)
}
// ... repeat for secondaryColor, mutedColor ...
ctx.Expose("primaryColor", primaryColor)
ctx.Expose("secondaryColor", secondaryColor)
ctx.Expose("mutedColor", mutedColor)
```

**activity_feed.go** - Child injects 4 colors (19 lines):
```go
// Same pattern for primaryColor, secondaryColor, mutedColor, warningColor
```

### After (5 lines total - 94% reduction!):

**app.go** - Parent provides theme (1 line):
```go
ctx.ProvideTheme(bubbly.DefaultTheme)
```

**repo_list.go** - Child uses theme (2 lines):
```go
theme := ctx.UseTheme(bubbly.DefaultTheme)
ctx.Expose("theme", theme)
```

**activity_feed.go** - Child uses theme (2 lines):
```go
theme := ctx.UseTheme(bubbly.DefaultTheme)
ctx.Expose("theme", theme)
```

### Template Usage

**Before:**
```go
primaryColor := ctx.Get("primaryColor").(lipgloss.Color)
titleStyle := lipgloss.NewStyle().Foreground(primaryColor)
```

**After:**
```go
theme := ctx.Get("theme").(bubbly.Theme)
titleStyle := lipgloss.NewStyle().Foreground(theme.Primary)
```

## Benefits of UseTheme/ProvideTheme

✅ **94% code reduction** in theme injection boilerplate  
✅ **Type-safe** - `theme.Primary` instead of type assertions  
✅ **Semantic colors** - `Primary`, `Secondary`, `Muted`, `Warning`, `Error`, `Success`  
✅ **Zero behavior change** - same colors, same output  
✅ **Backward compatible** - old inject/expose still works  
✅ **Clear intent** - "use theme" vs manual inject/expose  

## Code Metrics

- **Lines eliminated**: 34 lines (39 → 5)
- **Components migrated**: 3 (app, repo_list, activity_feed)
- **Tests**: All 24 tests pass with race detector
- **Visual output**: Identical before/after

## Testing

Run tests with race detector:
```bash
go test -race -v ./cmd/examples/10-testing/04-async/...
```

All tests use the testutil harness for proper component lifecycle management.

## Key Files

- `main.go` - Entry point with bubbly.Run()
- `app.go` - Root component with theme provision
- `components/repo_list.go` - Repository list with theme usage
- `components/activity_feed.go` - Activity feed with theme usage
- `composables/use_github_dashboard.go` - Async business logic
- `mock_github_api.go` - Mock API for testing

## Learn More

- **Theme System**: See `specs/13-adv-internal-package-automation/`
- **Composables**: See `specs/04-composition-api/`
- **Async Operations**: See `specs/08-automatic-reactive-bridge/`
- **Testing**: See `docs/api/testutil-reference.md`
