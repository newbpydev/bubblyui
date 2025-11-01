# Profiling Demo Example

This example demonstrates **Phase 8: Optimization & Monitoring** features, specifically profiling utilities and pprof integration.

## Features Demonstrated

- **Profiling Utilities** (Task 8.7)
  - `EnableProfiling` / `StopProfiling`
  - Standard pprof endpoints
  - `ProfileComposables` for custom profiling
  - Thread-safe operations

- **Production Profiling**
  - Localhost binding for security
  - Opt-in activation
  - Graceful shutdown
  - Real-time profile data collection

## Running the Example

```bash
# From the project root
go run ./cmd/examples/04-composables/profiling-demo/
```

## Usage

**Keyboard Controls:**
- `s` - Start profiling (enables pprof endpoints)
- `p` - Stop profiling
- `c` - Create workload (150 composables)
- `r` - Run composable profile (1 second)
- `q` - Quit

## pprof Endpoints

Once profiling is started, access pprof at:
```
http://localhost:6060/debug/pprof/
```

### Available Profiles

**CPU Profile:**
```bash
curl -o cpu.prof http://localhost:6060/debug/pprof/profile?seconds=10
go tool pprof cpu.prof
```

**Heap Profile:**
```bash
curl -o heap.prof http://localhost:6060/debug/pprof/heap
go tool pprof heap.prof
```

**Goroutine Profile:**
```bash
curl -o goroutine.prof http://localhost:6060/debug/pprof/goroutine
go tool pprof goroutine.prof
```

**All Profiles:**
Open in browser: `http://localhost:6060/debug/pprof/`

## Composable Profiling

The example includes custom composable profiling that shows:
- Number of composable calls
- Average execution time per composable
- Memory allocations
- Total duration

Example output:
```
Composable Profile (1s):

UseState: 150 calls, avg 3.5µs, 19200 bytes allocated
UseAsync: 75 calls, avg 4.2µs, 26400 bytes allocated
```

## What to Observe

1. **Profiling Status** - Shows when profiling is active
2. **Workload Size** - Tracks number of composables created
3. **pprof Endpoints** - Lists available profiling endpoints
4. **Profile Summary** - Shows composable performance stats

## Security Note

⚠️ **IMPORTANT:** This example binds to `localhost:6060` for security. In production:
- Always bind to localhost only
- Use SSH tunneling for remote access
- Never expose profiling endpoints publicly
- Consider adding authentication

## Production Usage

```bash
# On production server
ssh -L 6060:localhost:6060 user@production-server

# On local machine
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof
```

## Related Documentation

- [Production Profiling Guide](../../../../docs/guides/production-profiling.md)
- [Performance Optimization Guide](../../../../docs/guides/performance-optimization.md)
