# BubblyUI Observability

**Package Path:** `github.com/newbpydev/bubblyui/pkg/bubbly/observability`  
**Version:** 3.0  
**Purpose:** Error tracking, breadcrumbs, and reporting for production applications

## Overview

Observability provides error tracking with rich context, breadcrumbs, and multiple reporting backends (console, Sentry) for production BubblyUI applications.

## Quick Start

```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/observability"

// Create reporter
reporter := observability.NewSentryReporter("https://sentry.io", "your-dsn")

// Or console reporter
console := observability.NewConsoleReporter()

// Enable error tracking
observability.SetGlobalReporter(reporter)

// Use in components
Setup(func(ctx *bubbly.Context) {
    // Add breadcrumb
    observability.AddBreadcrumb("Component mounted", map[string]interface{}{
        "component": "MyComponent",
    })
    
    // Report error
    ctx.On("error", func(err interface{}) {
        observability.ReportError(err.(error))
    })
})
```

## Features

### 1. Error Reporting

```go
// Report with context
err := errors.New("database connection failed")
observability.ReportError(err, map[string]interface{}{
    "query": "SELECT * FROM users",
    "retryCount": 3,
})

// Report with tags
observability.ReportErrorWithTags(err, map[string]string{
    "severity": "high",
    "component": "database",
})
```

### 2. Breadcrumbs

```go
// Add breadcrumbs to track execution
observability.AddBreadcrumb("User clicked button", map[string]interface{}{
    "button": "submit",
})

observability.AddBreadcrumb("API called", map[string]interface{}{
    "endpoint": "/api/users",
    "method": "POST",
})

// Breadcrumbs auto-collected with component lifecycle
```

### 3. Reporters

**Console Reporter:**
```go
reporter := observability.NewConsoleReporter()
reporter.SetLevel(observability.LevelDebug)
```

**Sentry Reporter:**
```go
reporter := observability.NewSentryReporter("https://sentry.io", "your-dsn")
reporter.SetEnvironment("production")
reporter.SetRelease("v1.0.0")
```

## Integration

```go
// Global error handler
Setup(func(ctx *bubbly.Context) {
    ctx.On("panic", func(data interface{}) {
        panicked := data.(error)
        observability.ReportError(panicked)
    })
    
    // Component-specific breadcrumbs
    ctx.OnMounted(func() {
        observability.AddBreadcrumb("Component mounted", nil)
    })
})
```

## Configuration

```go
// Set global reporter
observability.SetGlobalReporter(reporter)

// Configure error levels
observability.SetLevel(observability.LevelError)  // Only errors
observability.SetLevel(observability.LevelWarning) // Errors + warnings
observability.SetLevel(observability.LevelInfo)    // Everything
```

**Package:** 2,547 LOC | Error tracking | Breadcrumbs | Multi-backend

---