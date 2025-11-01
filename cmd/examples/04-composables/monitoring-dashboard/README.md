# Monitoring Dashboard Example

This example demonstrates **Phase 8: Optimization & Monitoring** features, specifically metrics collection and Prometheus integration.

## Features Demonstrated

- **Metrics Collection** (Task 8.3)
  - Automatic composable creation tracking
  - Cache hit/miss tracking
  - Real-time metrics updates

- **Prometheus Integration** (Task 8.4)
  - Prometheus metrics exposed at `:9090/metrics`
  - Standard metric types (Counter, Gauge, Histogram)
  - Production-ready metric names

- **Real-time Monitoring**
  - Live dashboard with metrics visualization
  - Activity log tracking
  - Cache performance monitoring

## Running the Example

```bash
# From the project root
go run ./cmd/examples/04-composables/monitoring-dashboard/
```

## Usage

**Keyboard Controls:**
- `r` - Generate random activity
- `c` - Create composables (generates real metrics)
- `m` - Toggle metrics display
- `q` - Quit

## Prometheus Metrics

While the example is running, access Prometheus metrics at:
```
http://localhost:9090/metrics
```

Example metrics exposed:
```
bubblyui_composable_total{name="UseState"} 42
bubblyui_composable_creation_seconds_sum{name="UseState"} 0.123
bubblyui_cache_hits_total{operation="GetFieldIndex"} 156
bubblyui_cache_misses_total{operation="GetFieldIndex"} 12
bubblyui_cache_hit_ratio{operation="GetFieldIndex"} 0.928
```

## What to Observe

1. **Total Composables Counter** - Increases as composables are created
2. **Cache Hit Rate** - Shows cache effectiveness (target: >80%)
3. **Activity Log** - Real-time event tracking
4. **Metrics Endpoint** - View raw Prometheus metrics

## Integration with Production

This example shows how to:
- Enable metrics collection in your application
- Expose Prometheus metrics endpoint
- Monitor composable performance
- Track cache effectiveness

See `docs/guides/production-monitoring.md` for full production setup including Grafana dashboards and alerting.

## Related Documentation

- [Production Monitoring Guide](../../../../docs/guides/production-monitoring.md)
- [Performance Optimization Guide](../../../../docs/guides/performance-optimization.md)
