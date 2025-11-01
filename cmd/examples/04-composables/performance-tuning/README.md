# Performance Tuning Example

This example demonstrates **Phase 8: Optimization & Monitoring** features, specifically timer pooling for performance optimization.

## Features Demonstrated

- **Timer Pooling** (Task 8.1)
  - Reduces GC pressure from temporary timers
  - Configurable pool size
  - Significant performance improvement
  - Thread-safe pool management

- **Performance Benchmarking**
  - Real-time performance comparison
  - Pooling vs non-pooling measurements
  - Debounce/throttle overhead analysis

## Running the Example

```bash
# From the project root
go run ./cmd/examples/04-composables/performance-tuning/
```

## Usage

**Keyboard Controls:**
- `p` - Enable timer pooling (pool size: 100)
- `d` - Disable timer pooling
- `r` - Run benchmark (50 debounced composables)
- `c` - Clear results
- `q` - Quit

## Recommended Workflow

1. **Run Baseline** - Run benchmark with pooling OFF to establish baseline
2. **Enable Pooling** - Press `p` to enable timer pooling
3. **Run Comparison** - Press `r` to run benchmark with pooling ON
4. **Compare Results** - View the performance improvement

## Expected Performance Gains

With timer pooling enabled:
- **~12% faster** operations
- **~80% fewer GC runs**
- **More consistent** latency
- **Lower memory** pressure

Example results:
```
WITHOUT POOLING: Created 50 debounced composables in 125ms (2500 μs/op)
WITH POOLING:    Created 50 debounced composables in 110ms (2200 μs/op)
                 → 12% faster with pooling
```

## What to Observe

1. **Pooling Status** - Shows if timer pooling is enabled
2. **Last Benchmark** - Most recent benchmark timing
3. **Comparison** - Percentage improvement with pooling
4. **Benchmark Results** - Historical results log

## When to Use Timer Pooling

**Enable timer pooling when:**
- ✅ You have 100+ simultaneous debounce/throttle operations
- ✅ High-frequency timer creation (> 1000/second)
- ✅ GC pause time is a concern
- ✅ You've measured GC overhead from timers

**Don't enable if:**
- ❌ < 10 concurrent timers
- ❌ Timers created infrequently
- ❌ No GC pressure observed
- ❌ Before measuring (avoid premature optimization)

## Configuration

Adjust pool size based on your needs:

```go
// Small applications (10-50 timers)
timerpool.EnableGlobalPool(50)

// Medium applications (50-200 timers)
timerpool.EnableGlobalPool(100)

// Large applications (200+ timers)
timerpool.EnableGlobalPool(500)
```

**Rule of thumb:** Pool size = 2× peak concurrent timers

## Technical Details

Timer pooling works by:
1. Pre-allocating a pool of `time.Timer` instances
2. Reusing timers instead of creating new ones
3. Returning timers to the pool when done
4. Reducing GC pressure from short-lived timers

Memory cost: ~80 bytes per timer in pool

## Related Documentation

- [Performance Optimization Guide](../../../../docs/guides/performance-optimization.md)
- [Benchmark Guide](../../../../docs/guides/benchmark-guide.md)
