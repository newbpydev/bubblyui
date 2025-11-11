# Troubleshooting Guide

**Common issues and solutions for BubblyUI Dev Tools**

## Table of Contents

1. [Installation Issues](#installation-issues)
2. [Dev Tools Not Showing](#dev-tools-not-showing)
3. [Performance Problems](#performance-problems)
4. [Export/Import Errors](#exportimport-errors)
5. [Hook Registration Issues](#hook-registration-issues)
6. [Terminal Rendering Issues](#terminal-rendering-issues)
7. [Integration Problems](#integration-problems)
8. [Debug Mode](#debug-mode)
9. [FAQ](#faq)

---

## Installation Issues

### Dev Tools Package Not Found

**Error:**
```
package github.com/newbpydev/bubblyui/pkg/bubbly/devtools: cannot find package
```

**Solution:**
```bash
# Update BubblyUI to latest version
go get -u github.com/newbpydev/bubblyui

# Tidy dependencies
go mod tidy
```

### Import Cycle Detected

**Error:**
```
import cycle not allowed
```

**Cause:** Circular dependency between your code and dev tools.

**Solution:**
- Dev tools should be imported in `main` package only
- Don't import dev tools in library code
- Use framework hooks interface instead of direct dev tools access

---

## Dev Tools Not Showing

### Issue: F12 Does Nothing

**Check:**

1. **Dev tools enabled?**
   ```go
   if !devtools.IsEnabled() {
       devtools.Enable()
   }
   ```

2. **Called before component creation?**
   ```go
   // ✅ CORRECT
   devtools.Enable()
   app := NewApp()
   
   // ❌ WRONG
   app := NewApp()
   devtools.Enable()  // Too late!
   ```

3. **Alt screen mode active?**
   ```go
   // Dev tools require alt screen
   p := tea.NewProgram(app, tea.WithAltScreen())
   ```

4. **Terminal supports alt screen?**
   ```bash
   # Test with:
   echo $TERM
   # Should output: xterm-256color, screen-256color, or similar
   ```

### Issue: Blank Dev Tools Panel

**Cause:** No components registered or mounted.

**Solution:**
```go
// Verify component implements Component interface
type MyComponent struct {
    bubbly.BaseComponent  // Embed base
}

// Initialize component properly
func (c *MyComponent) Init() tea.Cmd {
    c.BaseComponent.Init()  // Important!
    return nil
}
```

### Issue: UI Corrupted After Toggle

**Cause:** Terminal size mismatch.

**Solution:**
```go
// Handle terminal resize
case tea.WindowSizeMsg:
    dt := devtools.Get()
    dt.HandleTerminalResize(msg.Width, msg.Height)
    // Update your app size too
```

---

## Performance Problems

### Issue: Application Slows Down

**Symptoms:**
- High CPU usage
- Slow rendering (>100ms)
- Dropped frames

**Solutions:**

1. **Reduce limits**
   ```go
   config := devtools.DefaultConfig()
   config.MaxComponents = 1000    // From 10000
   config.MaxEvents = 500         // From 5000
   config.MaxStateHistory = 100   // From 1000
   ```

2. **Enable sampling**
   ```go
   config.SamplingRate = 0.5  // Capture 50% of events
   ```

3. **Disable expensive features**
   ```go
   config.EnablePerformanceMonitor = false
   config.EnableStateHistory = false
   ```

4. **Use overlay mode**
   ```go
   config.LayoutMode = devtools.LayoutOverlay
   // Toggle with F12, doesn't split screen
   ```

### Issue: High Memory Usage

**Symptoms:**
- Memory growth over time
- OOM kills in long sessions

**Solutions:**

1. **Clear old data periodically**
   ```go
   ticker := time.NewTicker(1 * time.Hour)
   go func() {
       for range ticker.C {
           dt := devtools.Get()
           dt.ClearOldData(24 * time.Hour)  // Keep last 24h
       }
   }()
   ```

2. **Reduce history limits**
   ```go
   config.MaxStateHistory = 50  // Keep last 50 changes per ref
   ```

3. **Export and clear**
   ```go
   dt.Export("session-backup.json.gz", opts)
   dt.ClearAll()
   ```

### Issue: Slow Export

**Cause:** Large export size (>100MB).

**Solutions:**

1. **Use streaming mode**
   ```go
   opts := devtools.ExportOptions{
       UseStreaming: true,  // Constant memory
   }
   devtools.ExportStream("large.json", opts)
   ```

2. **Use compression**
   ```go
   opts.Compress = true
   opts.CompressionLevel = gzip.BestSpeed  // Faster
   ```

3. **Use MessagePack format**
   ```go
   opts.Format = devtools.MessagePackFormat{}  // 40% smaller
   ```

---

## Export/Import Errors

### Issue: Export File Too Large

**Error:**
```
export file exceeds 2GB
```

**Solutions:**

1. **Use incremental exports**
   ```go
   checkpoint, _ := devtools.ExportFull("full.json", opts)
   devtools.ExportIncremental("delta.json", checkpoint)
   ```

2. **Exclude performance data**
   ```go
   opts.IncludePerformance = false  // Saves ~50% size
   ```

3. **Increase sampling rate**
   ```go
   opts.SamplingRate = 0.1  // Only 10% of events
   ```

### Issue: Import Fails with "Unknown Format"

**Error:**
```
failed to detect format
```

**Solutions:**

1. **Specify format explicitly**
   ```go
   devtools.ImportWithFormat("file.dat", devtools.JSONFormat{})
   ```

2. **Check file extension**
   ```bash
   # Rename to correct extension
   mv session.dat session.json
   ```

3. **Validate file content**
   ```bash
   # Check if valid JSON
   cat session.json | jq .
   
   # Check if gzipped
   file session.json.gz
   ```

### Issue: Import Error "Version Mismatch"

**Error:**
```
export version 2.0 not compatible with current version 1.0
```

**Cause:** Export from newer dev tools version.

**Solutions:**

1. **Update BubblyUI**
   ```bash
   go get -u github.com/newbpydev/bubblyui
   ```

2. **Use migration tool** (if available)
   ```go
   migrated, err := devtools.Migrate(exportData, "2.0", "1.0")
   ```

3. **Export from newer version in compatible format**
   ```go
   // On newer version:
   opts.Version = "1.0"  // Export in older format
   ```

---

## Hook Registration Issues

### Issue: Hook Methods Not Called

**Check:**

1. **Hook registered?**
   ```go
   if !bubbly.IsHookRegistered() {
       bubbly.RegisterHook(&MyHook{})
   }
   ```

2. **Registered before component creation?**
   ```go
   // ✅ CORRECT
   bubbly.RegisterHook(hook)
   app := NewApp()
   
   // ❌ WRONG
   app := NewApp()
   bubbly.RegisterHook(hook)  // Missed mount events!
   ```

3. **All methods implemented?**
   ```go
   // Must implement ALL 11 methods
   type MyHook struct{}
   
   func (h *MyHook) OnComponentMount(id, name string) { /* ... */ }
   func (h *MyHook) OnComponentUpdate(id string, msg interface{}) { /* ... */ }
   // ... all 11 methods
   ```

### Issue: Hook Causes Panic

**Error:**
```
panic: runtime error: invalid memory address
```

**Solutions:**

1. **Add panic recovery**
   ```go
   func (h *MyHook) OnEvent(id, name string, data interface{}) {
       defer func() {
           if r := recover(); r != nil {
               log.Printf("Hook panic: %v", r)
           }
       }()
       // Your code...
   }
   ```

2. **Check for nil values**
   ```go
   func (h *MyHook) OnRefChange(id string, oldVal, newVal interface{}) {
       if oldVal == nil || newVal == nil {
           return  // Skip nil values
       }
       // Process...
   }
   ```

### Issue: Hook Slows Down Application

**Cause:** Slow hook implementation.

**Solution:**

1. **Profile hook**
   ```go
   func (h *MyHook) OnComponentUpdate(id string, msg interface{}) {
       start := time.Now()
       defer func() {
           if elapsed := time.Since(start); elapsed > 1*time.Millisecond {
               log.Printf("Slow hook: %v", elapsed)
           }
       }()
       // Your code...
   }
   ```

2. **Use buffering**
   ```go
   type BufferedHook struct {
       buffer chan Event
   }
   
   func (h *BufferedHook) OnEvent(id, name string, data interface{}) {
       select {
       case h.buffer <- Event{id, name, data}:
       default:
           // Drop if full
       }
   }
   ```

---

## Terminal Rendering Issues

### Issue: Garbled Output

**Cause:** Terminal doesn't support required features.

**Solutions:**

1. **Check terminal type**
   ```bash
   echo $TERM
   # Should be: xterm-256color, screen-256color
   ```

2. **Force terminal type**
   ```bash
   export TERM=xterm-256color
   go run main.go
   ```

3. **Use simple layout**
   ```go
   config.LayoutMode = devtools.LayoutOverlay  // Simpler rendering
   ```

### Issue: Colors Not Showing

**Check:**

1. **Terminal supports colors?**
   ```bash
   tput colors
   # Should output: 256
   ```

2. **Force color output**
   ```bash
   export COLORTERM=truecolor
   ```

### Issue: Unicode Characters Broken

**Cause:** Terminal doesn't support UTF-8.

**Solutions:**

1. **Set locale**
   ```bash
   export LC_ALL=en_US.UTF-8
   export LANG=en_US.UTF-8
   ```

2. **Use ASCII mode** (if available)
   ```go
   config.UseASCII = true  // No unicode characters
   ```

---

## Integration Problems

### Issue: Conflicts with Other TUI Libraries

**Symptoms:**
- Keyboard shortcuts not working
- Screen corruption
- Panic on startup

**Solutions:**

1. **Ensure single tea.Program**
   ```go
   // Only one program instance
   p := tea.NewProgram(app, tea.WithAltScreen())
   ```

2. **Don't mix TUI frameworks**
   - Don't use tcell directly with dev tools
   - Don't mix termbox with Bubbletea

3. **Coordinate keyboard shortcuts**
   ```go
   // Reserve F12 for dev tools
   // Use other keys for app shortcuts
   ```

### Issue: Websocket/Network Conflicts

**Cause:** Network operations in hook blocking.

**Solution:**

```go
// Use goroutines for network I/O
func (h *MyHook) OnEvent(id, name string, data interface{}) {
    go func() {
        h.sendToServer(id, name, data)  // Non-blocking
    }()
}
```

---

## Debug Mode

Enable debug mode for verbose logging:

```bash
export BUBBLY_DEVTOOLS_DEBUG=true
go run main.go 2> devtools-debug.log
```

**Debug output includes:**
- Hook call traces
- Performance measurements
- Memory allocations
- Error stack traces

**Analyze debug log:**
```bash
# Find slow operations
grep "SLOW" devtools-debug.log

# Find errors
grep "ERROR" devtools-debug.log

# Find panics
grep "PANIC" devtools-debug.log
```

---

## FAQ

**Q: Can I use dev tools in production?**  
A: Yes, but disable by default. Enable via environment variable or flag.

**Q: Do dev tools work with Docker?**  
A: Yes. Ensure TTY allocation: `docker run -it app`

**Q: Can I debug remote applications?**  
A: Export debug session, transfer file, then import locally.

**Q: Why is my terminal frozen?**  
A: Press `Ctrl+C` to quit. May be deadlock in hook.

**Q: Can I run tests with dev tools enabled?**  
A: Yes, but disable for faster tests: `devtools.Disable()`

**Q: How do I report bugs?**  
A: Include:
  - Debug log (`BUBBLY_DEVTOOLS_DEBUG=true`)
  - Export of session (`devtools.Export()`)
  - Terminal info (`echo $TERM`, `tput colors`)
  - Go version (`go version`)

---

**Still having issues?** Open an issue on GitHub with debug log and minimal reproduction.

**See also:**
- [Best Practices](./best-practices.md) for optimization tips
- [Reference](./reference.md) for configuration options
