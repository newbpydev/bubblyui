# Export & Import Guide

**Complete guide to debug session persistence and sharing**

Save, compress, sanitize, and share debug sessions with your team or for offline analysis.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Export Options](#export-options)
3. [Compression](#compression)
4. [Format Selection](#format-selection)
5. [Sanitization Integration](#sanitization-integration)
6. [Incremental Exports](#incremental-exports)
7. [Streaming Mode](#streaming-mode)
8. [Import](#import)
9. [Best Practices](#best-practices)

---

## Quick Start

### Basic Export

```go
dt := devtools.Get()
err := dt.Export("debug-session.json", devtools.ExportOptions{
    IncludeComponents:  true,
    IncludeState:       true,
    IncludeEvents:      true,
    IncludePerformance: true,
})
```

### Basic Import

```go
err := devtools.Import("debug-session.json")
```

That's it! Format and compression are auto-detected.

---

## Export Options

Control what gets included in exports:

```go
opts := devtools.ExportOptions{
    // Content
    IncludeComponents:  true,   // Component tree snapshots
    IncludeState:       true,   // Ref/Computed values and history
    IncludeEvents:      true,   // Event log
    IncludePerformance: true,   // Render timing data
    IncludeTimestamps:  true,   // Precise timestamps
    
    // Compression
    Compress:         true,
    CompressionLevel: gzip.BestCompression,
    
    // Format
    Format: devtools.JSONFormat{},  // Or YAMLFormat, MessagePackFormat
    
    // Sanitization
    Sanitize: sanitizer,  // Optional data sanitizer
    
    // Streaming
    UseStreaming:      false,
    ProgressCallback:  nil,
}
```

### Content Options Explained

**IncludeComponents:**
- Component tree hierarchy
- Component state (refs, computed)
- Component props and metadata
- ~5KB per component

**IncludeState:**
- All ref values with types
- Computed value cache
- State change history
- ~1KB per ref + history

**IncludeEvents:**
- Complete event log
- Event payloads
- Timing information
- ~500 bytes per event

**IncludePerformance:**
- Render timing per component
- Flame graph data
- Memory usage snapshots
- ~2KB per component

**Typical sizes (100 components, 1000 events, 5min session):**
- All included: ~2MB uncompressed
- With gzip: ~400KB (80% reduction)
- With MessagePack: ~1.2MB uncompressed
- MessagePack + gzip: ~250KB

---

## Compression

Reduce file sizes by 60-70% with gzip compression.

### Compression Levels

```go
import "compress/gzip"

opts := devtools.ExportOptions{
    Compress:         true,
    CompressionLevel: gzip.BestCompression,
}
```

**Available levels:**
- `gzip.BestSpeed` (1) - 50% reduction, fastest
- `gzip.DefaultCompression` (-1) - 60% reduction, balanced
- `gzip.BestCompression` (9) - 70% reduction, maximum

**Benchmark (2MB export):**
```
Level              Size    Time     Read Time
─────────────────────────────────────────────
No compression     2.0MB   0ms      10ms
BestSpeed          1.0MB   50ms     15ms
Default            800KB   100ms    20ms
BestCompression    600KB   200ms    25ms
```

**Recommendation:** Use `gzip.DefaultCompression` for best balance.

### Auto-Detection

Compressed files are detected by `.gz` extension:

```go
// Exports with gzip
devtools.Export("session.json.gz", opts)

// Imports automatically decompress
devtools.Import("session.json.gz")
```

---

## Format Selection

Choose the best format for your use case.

### JSON Format

**Best for:** Universal compatibility, human readability

```go
opts.Format = devtools.JSONFormat{}
devtools.Export("session.json", opts)
```

**Pros:**
- Universal tool support
- Human-readable
- Easy to diff/merge

**Cons:**
- Largest file size (baseline)
- Slower parsing

**Size:** 100% (2.0MB example)

### YAML Format

**Best for:** Configuration management, readability

```go
opts.Format = devtools.YAMLFormat{}
devtools.Export("session.yaml", opts)
```

**Pros:**
- Most human-readable
- Comments supported
- Configuration tool integration

**Cons:**
- Slightly larger than JSON
- Slower parsing than JSON

**Size:** 110% (2.2MB example)

### MessagePack Format

**Best for:** Performance, minimal size

```go
opts.Format = devtools.MessagePackFormat{}
devtools.Export("session.msgpack", opts)
```

**Pros:**
- Smallest size (binary)
- Fastest parsing
- Efficient encoding

**Cons:**
- Not human-readable
- Fewer tool support

**Size:** 60% (1.2MB example)

### Format Comparison Table

| Format      | Size   | Speed  | Readable | Tools |
|-------------|--------|--------|----------|-------|
| JSON        | 100%   | Medium | ✓        | ✓✓✓   |
| YAML        | 110%   | Slow   | ✓✓       | ✓✓    |
| MessagePack | 60%    | Fast   | ✗        | ✓     |

**Combined with compression:**
- JSON + gzip: 400KB (20% of uncompressed)
- YAML + gzip: 440KB (22% of uncompressed)
- MessagePack + gzip: 250KB (12.5% of uncompressed)

**Recommendation:** 
- Development: JSON (universal)
- Production: MessagePack + gzip (smallest)
- Documentation: YAML (readable)

---

## Sanitization Integration

Remove sensitive data before exporting.

### Basic Sanitization

```go
// Create sanitizer with templates
sanitizer := devtools.NewSanitizer()
sanitizer.LoadTemplates("pii", "pci")

// Export with sanitization
opts := devtools.ExportOptions{
    IncludeState: true,
    Sanitize:     sanitizer,
}

devtools.Export("safe-session.json", opts)
```

### Custom Patterns

```go
// Add API key pattern
sanitizer.AddPatternWithPriority(
    `(?i)(api[_-]?key)(["'\s:=]+)([^\s"']+)`,
    "${1}${2}[REDACTED]",
    80,  // High priority
    "api_key",
)

// Add custom database connection strings
sanitizer.AddPattern(
    `postgres://[^@]+@[^/]+/\w+`,
    `postgres://[REDACTED]:[REDACTED]@[REDACTED]/[REDACTED]`,
)
```

### Preview Sanitization

See what would be redacted before exporting:

```go
// Get export data (without saving)
data := devtools.GetExportData(opts)

// Preview sanitization
result := sanitizer.Preview(data)

fmt.Printf("Would redact %d values:\n", result.WouldRedactCount)
for _, match := range result.Matches[:10] {  // First 10
    fmt.Printf("  - %s at %s\n", match.Pattern, match.Location)
}

// Confirm and export
if userConfirms() {
    devtools.Export("session.json", opts)
}
```

### Sanitization Metrics

```go
// Export with sanitization
devtools.Export("session.json", opts)

// Get stats
stats := sanitizer.GetLastStats()
fmt.Printf("Patterns applied: %d\n", stats.PatternsApplied)
fmt.Printf("Values redacted: %d\n", stats.ValuesRedacted)
fmt.Printf("Processing time: %v\n", stats.Duration)
```

---

## Incremental Exports

For long-running applications, save only changes since last export.

### How It Works

```
Day 1: Full export (125MB)
  ↓
Day 2: Delta export (8MB)  ← Only changes
  ↓
Day 3: Delta export (7MB)  ← Only changes
```

**Storage savings:** 93% (15MB vs 375MB for 3 days)

### Usage

```go
// Day 1: Full snapshot
checkpoint, err := devtools.ExportFull("day-1.json", opts)
if err != nil {
    log.Fatal(err)
}

// Day 2: Only changes since checkpoint
checkpoint, err = devtools.ExportIncremental("day-2-delta.json", checkpoint)
if err != nil {
    log.Fatal(err)
}

// Day 3: Only changes since last checkpoint
checkpoint, err = devtools.ExportIncremental("day-3-delta.json", checkpoint)
if err != nil {
    log.Fatal(err)
}
```

### Reconstructing Timeline

```go
// Import full snapshot
err := devtools.Import("day-1.json")

// Apply deltas in order
err = devtools.ImportDelta("day-2-delta.json")
err = devtools.ImportDelta("day-3-delta.json")

// Now have complete 3-day history
```

### Checkpoint Structure

```go
type ExportCheckpoint struct {
    Timestamp        time.Time
    ComponentCount   int
    EventCount       int
    StateVersion     int
    LastComponentID  string
    LastEventID      string
}
```

**Best practices:**
- Full export daily or weekly
- Delta exports hourly or after significant events
- Keep at least one full export for recovery

---

## Streaming Mode

Handle exports > 100MB without memory issues.

### When to Use Streaming

- Exports > 100MB
- Long-running applications (days/weeks)
- Memory-constrained environments
- CI/CD debug logs

### Basic Streaming

```go
opts := devtools.ExportOptions{
    IncludeState:  true,
    IncludeEvents: true,
    UseStreaming:  true,
}

err := devtools.ExportStream("large-session.json", opts)
```

### With Progress Callback

```go
opts.ProgressCallback = func(bytesProcessed int64) {
    mb := bytesProcessed / 1024 / 1024
    fmt.Printf("\rProcessed: %d MB", mb)
}

err := devtools.ExportStream("session.json", opts)
```

### Memory Guarantees

**Non-streaming export:**
```
Memory usage = Full export size in RAM
2GB export = 2GB RAM usage
```

**Streaming export:**
```
Memory usage = Buffer size (default: 64KB)
2GB export = 64KB RAM usage
```

### Performance

**Non-streaming:**
- Faster (all in memory)
- Memory = export size
- Fails on large exports (OOM)

**Streaming:**
- Slightly slower (I/O overhead)
- Constant memory (bounded buffer)
- Never fails due to size

**Recommendation:** Use streaming for exports > 100MB or in memory-constrained environments.

---

## Import

Load debug sessions for analysis.

### Auto-Detection

Format and compression are auto-detected:

```go
// Works for all formats
devtools.Import("session.json")
devtools.Import("session.json.gz")
devtools.Import("session.yaml")
devtools.Import("session.msgpack.gz")
```

### Manual Format

```go
// Override auto-detection
err := devtools.ImportWithFormat("session.dat", devtools.MessagePackFormat{})
```

### Validation

Validate before importing:

```go
// Read file
data, err := os.ReadFile("session.json")
if err != nil {
    log.Fatal(err)
}

// Parse
exportData, err := devtools.ParseExportData(data, devtools.JSONFormat{})
if err != nil {
    log.Fatal(err)
}

// Validate
err = devtools.ValidateImport(exportData)
if err != nil {
    log.Printf("Invalid export: %v", err)
    return
}

// Import
err = devtools.ImportData(exportData)
```

### Merge vs Replace

```go
// Replace: Clear existing data
devtools.ClearAll()
devtools.Import("session.json")

// Merge: Add to existing data
devtools.Import("session-1.json")
devtools.Import("session-2.json")  // Merged
```

---

## Best Practices

### Development

```go
// Quick exports during development
devtools.Export("debug.json", devtools.ExportOptions{
    IncludeState:  true,
    IncludeEvents: true,
    Compress:      false,  // Faster, human-readable
})
```

### Production

```go
// Sanitized, compressed, minimal
sanitizer := devtools.NewSanitizer()
sanitizer.LoadTemplates("pii", "pci", "hipaa")

devtools.Export("prod-session.msgpack.gz", devtools.ExportOptions{
    IncludeComponents:  false,  // Reduce size
    IncludeState:       true,
    IncludeEvents:      true,
    IncludePerformance: true,
    Compress:           true,
    CompressionLevel:   gzip.BestCompression,
    Format:             devtools.MessagePackFormat{},
    Sanitize:           sanitizer,
})
```

### Sharing with Team

```go
// Readable format with sanitization
devtools.Export("team-debug.yaml", devtools.ExportOptions{
    IncludeState:  true,
    IncludeEvents: true,
    Compress:      true,
    Format:        devtools.YAMLFormat{},
    Sanitize:      sanitizer,
})
```

### Long-Running Applications

```go
// Daily full export + hourly deltas
var checkpoint *devtools.ExportCheckpoint

// Once per day
checkpoint, _ = devtools.ExportFull("full-2024-11-11.json.gz", opts)

// Every hour
checkpoint, _ = devtools.ExportIncremental(
    fmt.Sprintf("delta-%s.json.gz", time.Now().Format("2006-01-02-15")),
    checkpoint,
)
```

### CI/CD Debug Logs

```go
// Streaming export to artifact
opts := devtools.ExportOptions{
    IncludeState:     true,
    IncludeEvents:    true,
    UseStreaming:     true,
    Compress:         true,
    Format:           devtools.MessagePackFormat{},
    ProgressCallback: func(bytes int64) {
        log.Printf("Exported %d MB", bytes/1024/1024)
    },
}

devtools.ExportStream("ci-debug-log.msgpack.gz", opts)
```

---

## FAQ

**Q: Can I export only specific components?**  
A: Not directly. Export all, then filter during import or use a custom exporter.

**Q: How big can exports get?**  
A: No hard limit. Use streaming for exports > 100MB. Tested up to 10GB.

**Q: Can I edit exports?**  
A: Yes for JSON/YAML (text editors). No for MessagePack (binary). Use `jq` or `yq` for JSON/YAML manipulation.

**Q: Are exports portable across versions?**  
A: Yes. Versioned exports include schema version. Auto-migration on import (future feature).

**Q: Can I export to cloud storage?**  
A: Yes. Export to file, then upload. Or implement custom exporter writing to S3/GCS directly.

---

**Next:** [Best Practices Guide](./best-practices.md) for optimization tips →
