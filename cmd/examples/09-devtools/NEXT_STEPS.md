# Next Steps: Examples 03-10

**Implementation roadmap for remaining dev tools examples**

## Status Summary

✅ **Completed:**
- Architecture guide (`docs/architecture/composable-apps.md`)
- Examples overview README
- Example 01: Basic Enablement
- Example 02: Component Inspection

🚧 **Remaining:**
- Example 03: State Debugging
- Example 04: Event Monitoring
- Example 05: Performance Profiling
- Example 06: Reactive Cascade
- Example 07: Export & Import
- Example 08: Custom Sanitization
- Example 09: Custom Hooks
- Example 10: Production Ready

---

## Example 03: State Debugging

**Purpose:** Ref and Computed tracking with history

### Components Needed
```
03-state-debugging/
├── main.go
├── app.go
├── components/
│   ├── form_field.go          # Input with validation
│   ├── validation_display.go  # Shows validation state
│   └── history_viewer.go      # Shows state history
└── composables/
    └── use_form_validation.go # Form with Ref + Computed
```

### Key Features to Show
- Ref state changes with timestamps
- Computed value derivation (validation)
- State history timeline
- Time-travel debugging (restore previous values)
- State edit functionality in dev tools

### Implementation Focus
- Use `Input` component for text fields
- Create Computed for validation (isValid, errors)
- Show how to view history in dev tools (h key)
- Demonstrate state restoration

### Estimated Effort: 4 hours

---

## Example 04: Event Monitoring

**Purpose:** Event emission and capture

### Components Needed
```
04-event-monitoring/
├── main.go
├── app.go
├── components/
│   ├── event_emitter.go    # Emits custom events
│   ├── event_logger.go     # Displays event log
│   └── event_stats.go      # Event statistics
└── README.md
```

### Key Features to Show
- Custom event emission (`ctx.Emit`)
- Event bubbling through component tree
- Event log in dev tools
- Event filtering by name/source
- Event replay (if implemented)

### Implementation Focus
- Create components that emit various events
- Show event flow from child → parent
- Use dev tools Event tab
- Filter events by type

### Estimated Effort: 3 hours

---

## Example 05: Performance Profiling

**Purpose:** Render performance analysis

### Components Needed
```
05-performance-profiling/
├── main.go
├── app.go
├── components/
│   ├── slow_component.go      # Intentionally slow (10ms+)
│   ├── fast_component.go      # Fast rendering
│   ├── data_table.go          # Large list
│   └── perf_summary.go        # Performance stats
└── README.md
```

### Key Features to Show
- Slow rendering detection (>50ms threshold)
- Flame graph visualization
- Timeline analysis
- Performance metrics (avg, min, max render time)

### Implementation Focus
- Use `time.Sleep()` to simulate slow operations
- Show flame graph in dev tools
- Identify bottlenecks
- Compare slow vs fast components

### Estimated Effort: 4 hours

---

## Example 06: Reactive Cascade

**Purpose:** Visualize complete reactive flow

### Components Needed
```
06-reactive-cascade/
├── main.go
├── app.go
├── components/
│   ├── cascade_visualizer.go  # Shows flow diagram
│   └── reactive_demo.go       # Triggers cascades
├── composables/
│   └── use_reactive_cascade.go # Complex reactive setup
└── README.md
```

### Key Features to Show
- Ref changes trigger Computed updates
- Computed changes trigger Watch callbacks
- Watch callbacks trigger side effects
- WatchEffect automatic re-runs
- Component tree mutations (add/remove children)
- Full cascade visibility through framework hooks

### Implementation Focus
- Create complex reactive dependencies
- Use `bubbly.Watch()` to observe changes
- Use `bubbly.WatchEffect()` for auto-tracking
- Show cascade in dev tools hooks view
- Demonstrate `OnChildAdded`/`OnChildRemoved` hooks

### Estimated Effort: 6 hours

---

## Example 07: Export & Import

**Purpose:** Debug session export/import workflow

### Components Needed
```
07-export-import/
├── main.go
├── app.go
├── components/
│   ├── export_controls.go     # Export buttons
│   ├── import_controls.go     # Import selector
│   └── format_selector.go     # Choose format
└── README.md
```

### Key Features to Show
- Export with compression (gzip)
- Multiple formats (JSON, YAML, MessagePack)
- Format auto-detection on import
- Versioned exports
- Sharing debug sessions workflow

### Implementation Focus
- Use `devtools.Export()` with different options
- Show compression levels (BestSpeed, Default, BestCompression)
- Demonstrate format selection
- Import and verify session

### Estimated Effort: 4 hours

---

## Example 08: Custom Sanitization

**Purpose:** PII removal and custom patterns

### Components Needed
```
08-custom-sanitization/
├── main.go
├── app.go
├── components/
│   ├── data_form.go           # Form with sensitive data
│   ├── sanitizer_config.go    # Configure sanitization
│   ├── preview_pane.go        # Dry-run preview
│   └── export_sanitized.go    # Export with sanitization
└── README.md
```

### Key Features to Show
- Built-in compliance templates (PII, PCI, HIPAA, GDPR)
- Custom sanitization patterns
- Priority-based rule system
- Dry-run preview before export
- Streaming sanitization for large exports
- Sanitization metrics

### Implementation Focus
- Create form with email, phone, SSN fields
- Load compliance templates
- Add custom patterns (API keys)
- Show preview with `sanitizer.Preview()`
- Export sanitized data

### Estimated Effort: 5 hours

---

## Example 09: Custom Hooks

**Purpose:** Framework hook implementation

### Components Needed
```
09-custom-hooks/
├── main.go
├── app.go
├── hooks/
│   ├── perf_monitor_hook.go   # Performance monitoring
│   ├── audit_hook.go          # State change auditing
│   └── telemetry_hook.go      # External integration
└── README.md
```

### Key Features to Show
- Implement FrameworkHook interface (all 11 methods)
- Custom performance monitoring hook
- State change auditing
- Integration with external tools (console logging)
- Hook lifecycle management

### Implementation Focus
- Create custom hook implementing all methods
- Register with `bubbly.RegisterHook()`
- Show hook output in console
- Demonstrate hook use cases

### Estimated Effort: 5 hours

---

## Example 10: Production Ready

**Purpose:** Production-ready integration

### Components Needed
```
10-production-ready/
├── main.go
├── app.go
├── config/
│   └── devtools.yaml          # Configuration file
└── README.md
```

### Key Features to Show
- Environment-based enablement
- Configuration from files
- Resource limits (MaxComponents, MaxEvents)
- Export sanitization
- Error handling best practices
- Performance optimization

### Implementation Focus
- Load config from environment variables
- Load config from YAML file
- Set appropriate limits
- Handle errors gracefully
- Show production deployment pattern

### Estimated Effort: 4 hours

---

## Implementation Strategy

### Phase 1: Core Debugging (Examples 03-04)
**Timeline:** Week 1
- 03-state-debugging (4h)
- 04-event-monitoring (3h)
- **Total:** 7 hours

### Phase 2: Performance & Reactivity (Examples 05-06)
**Timeline:** Week 2
- 05-performance-profiling (4h)
- 06-reactive-cascade (6h)
- **Total:** 10 hours

### Phase 3: Data Management (Examples 07-08)
**Timeline:** Week 3
- 07-export-import (4h)
- 08-custom-sanitization (5h)
- **Total:** 9 hours

### Phase 4: Advanced & Production (Examples 09-10)
**Timeline:** Week 4
- 09-custom-hooks (5h)
- 10-production-ready (4h)
- **Total:** 9 hours

**Grand Total:** 35 hours (~5 working days)

---

## Consistent Patterns

All examples should follow:

### Directory Structure
```
example/
├── main.go              # Entry point with devtools.Enable()
├── app.go               # Root component
├── components/          # UI components
│   └── *.go
├── composables/         # Shared logic (optional)
│   └── use_*.go
└── README.md            # Documentation
```

### Component Pattern
- Factory functions: `CreateComponent(props)`
- Props structs for configuration
- Setup function for logic
- Template function for rendering
- Use BubblyUI components (not raw Lipgloss)

### README Pattern
- What This Demonstrates
- Architecture (hierarchy diagram)
- Key Features (numbered list)
- Code Highlights (snippets)
- Run the Example
- Using Dev Tools (step-by-step)
- Troubleshooting
- Next Steps link

---

## Quality Standards

Each example must:

1. ✅ **Build without errors** - `go build ./...`
2. ✅ **Follow composable architecture** - Per guide
3. ✅ **Use BubblyUI components** - No manual Lipgloss for components
4. ✅ **Expose state properly** - `ctx.Expose()` for dev tools
5. ✅ **Include README** - Complete documentation
6. ✅ **Comment key concepts** - Explain "why" not just "what"
7. ✅ **Dev tools integration** - Show specific features
8. ✅ **Runnable** - Can execute and interact immediately

---

## Testing Checklist

Before marking example complete:

- [ ] App runs without errors
- [ ] Dev tools toggle with F12 works
- [ ] Component tree shows correctly
- [ ] State is visible in dev tools
- [ ] Keyboard shortcuts work
- [ ] README is accurate
- [ ] Code is well-commented
- [ ] Follows architecture patterns

---

## Documentation Updates

After completing all examples:

1. Update main examples README with completion status
2. Update dev tools documentation with example links
3. Add screenshots (ASCII art) to READMEs
4. Create video/GIF walkthrough (optional)
5. Update CHANGELOG.md with new examples

---

## Priority Order

If time-constrained, implement in this order:

1. **Example 03** (State Debugging) - Most requested feature
2. **Example 06** (Reactive Cascade) - Unique to BubblyUI
3. **Example 05** (Performance) - Practical optimization
4. **Example 07** (Export/Import) - Sharing capability
5. **Example 08** (Sanitization) - Production necessity
6. **Example 09** (Custom Hooks) - Advanced users
7. **Example 04** (Events) - Can infer from others
8. **Example 10** (Production) - Wrap-up best practices

---

**Current Status:** Examples 01-02 complete. Ready to begin Example 03.

**Next Action:** Implement Example 03: State Debugging following the pattern established in 01-02.
