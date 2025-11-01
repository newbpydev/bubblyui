# Production Monitoring Guide

This guide covers setting up production monitoring for BubblyUI applications using Prometheus and Grafana, including metrics collection, dashboards, and alerting rules.

## 📊 Overview

Effective monitoring is essential for production applications. This guide covers:

- **Prometheus metrics** - Collecting composable and performance metrics
- **Grafana dashboards** - Visualizing application health
- **Alerting rules** - Getting notified of issues
- **Best practices** - What to monitor and why
- **Tree depth tracking** - Component hierarchy monitoring

## 🚀 Quick Start

### Enable Metrics Collection

```go
package main

import (
    "net/http"
    
    "github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    // Set up Prometheus metrics
    metrics := monitoring.NewPrometheusMetrics()
    monitoring.SetGlobalMetrics(metrics)
    
    // Expose /metrics endpoint
    http.Handle("/metrics", promhttp.Handler())
    go http.ListenAndServe(":9090", nil)
    
    // Your application
    // ...
}
```

### Verify Metrics

```bash
# Check metrics endpoint
curl http://localhost:9090/metrics

# Sample output:
# bubblyui_composable_creation_seconds_count{name="UseState"} 1523
# bubblyui_composable_creation_seconds_sum{name="UseState"} 5.234
# bubblyui_cache_hits_total{operation="GetFieldIndex"} 45231
# bubblyui_cache_misses_total{operation="GetFieldIndex"} 892
```

## 📈 Prometheus Metrics

### Available Metrics

**Composable Metrics:**

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `bubblyui_composable_creation_seconds` | Histogram | `name` | Time to create composable |
| `bubblyui_composable_allocations_bytes` | Histogram | `name` | Bytes allocated |
| `bubblyui_composable_total` | Counter | `name` | Total composables created |

**Cache Metrics:**

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `bubblyui_cache_hits_total` | Counter | `operation` | Cache hit count |
| `bubblyui_cache_misses_total` | Counter | `operation` | Cache miss count |
| `bubblyui_cache_hit_ratio` | Gauge | `operation` | Hit ratio (0-1) |

**Dependency Injection Metrics:**

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `bubblyui_inject_depth` | Histogram | - | Tree depth for inject calls |
| `bubblyui_provide_total` | Counter | - | Total provide operations |
| `bubblyui_inject_total` | Counter | - | Total inject operations |

### Metric Labels

**Composable Names:**
- `UseState`
- `UseForm`
- `UseAsync`
- `UseEffect`
- `UseDebounce`
- `UseThrottle`
- `UseLocalStorage`

**Cache Operations:**
- `GetFieldIndex`
- `GetFieldType`
- `WarmUp`

### Sample Queries

**Composable Creation Rate:**
```promql
# Creations per second
rate(bubblyui_composable_total[5m])

# By composable type
sum by (name) (rate(bubblyui_composable_total[5m]))
```

**Average Creation Time:**
```promql
# Average time in milliseconds
rate(bubblyui_composable_creation_seconds_sum[5m]) / 
rate(bubblyui_composable_creation_seconds_count[5m]) * 1000
```

**Cache Hit Rate:**
```promql
# Overall hit rate
sum(rate(bubblyui_cache_hits_total[5m])) / 
(sum(rate(bubblyui_cache_hits_total[5m])) + sum(rate(bubblyui_cache_misses_total[5m])))

# By operation
bubblyui_cache_hit_ratio
```

**P95 Creation Time:**
```promql
histogram_quantile(0.95, 
  rate(bubblyui_composable_creation_seconds_bucket[5m])
)
```

## 📊 Grafana Dashboard

### Dashboard Setup

1. **Add Prometheus Data Source**

```yaml
# datasources.yml
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    url: http://prometheus:9090
    isDefault: true
```

2. **Import Dashboard Template**

Create dashboard with these panels:

**Panel 1: Composable Creation Rate**
```json
{
  "title": "Composable Creation Rate",
  "targets": [{
    "expr": "sum by (name) (rate(bubblyui_composable_total[5m]))",
    "legendFormat": "{{name}}"
  }],
  "type": "graph"
}
```

**Panel 2: Average Creation Time**
```json
{
  "title": "Composable Creation Time (ms)",
  "targets": [{
    "expr": "rate(bubblyui_composable_creation_seconds_sum[5m]) / rate(bubblyui_composable_creation_seconds_count[5m]) * 1000",
    "legendFormat": "{{name}}"
  }],
  "type": "graph",
  "yaxes": [{
    "format": "ms"
  }]
}
```

**Panel 3: Cache Hit Rate**
```json
{
  "title": "Cache Hit Rate",
  "targets": [{
    "expr": "bubblyui_cache_hit_ratio",
    "legendFormat": "{{operation}}"
  }],
  "type": "graph",
  "yaxes": [{
    "format": "percentunit",
    "max": 1,
    "min": 0
  }]
}
```

**Panel 4: Injection Tree Depth**
```json
{
  "title": "Dependency Injection Depth (P95)",
  "targets": [{
    "expr": "histogram_quantile(0.95, rate(bubblyui_inject_depth_bucket[5m]))"
  }],
  "type": "graph"
}
```

**Panel 5: Memory Allocations**
```json
{
  "title": "Memory Allocations per Composable",
  "targets": [{
    "expr": "rate(bubblyui_composable_allocations_bytes_sum[5m]) / rate(bubblyui_composable_allocations_bytes_count[5m])",
    "legendFormat": "{{name}}"
  }],
  "type": "graph",
  "yaxes": [{
    "format": "bytes"
  }]
}
```

### Complete Dashboard JSON

```json
{
  "dashboard": {
    "title": "BubblyUI Monitoring",
    "panels": [
      {
        "id": 1,
        "title": "Composable Creation Rate",
        "gridPos": {"x": 0, "y": 0, "w": 12, "h": 8},
        "targets": [{
          "expr": "sum by (name) (rate(bubblyui_composable_total[5m]))",
          "legendFormat": "{{name}}"
        }]
      },
      {
        "id": 2,
        "title": "Cache Performance",
        "gridPos": {"x": 12, "y": 0, "w": 12, "h": 8},
        "targets": [{
          "expr": "bubblyui_cache_hit_ratio",
          "legendFormat": "{{operation}}"
        }]
      },
      {
        "id": 3,
        "title": "Creation Time P95",
        "gridPos": {"x": 0, "y": 8, "w": 12, "h": 8},
        "targets": [{
          "expr": "histogram_quantile(0.95, rate(bubblyui_composable_creation_seconds_bucket[5m]))"
        }]
      },
      {
        "id": 4,
        "title": "Component Tree Depth",
        "gridPos": {"x": 12, "y": 8, "w": 12, "h": 8},
        "targets": [{
          "expr": "histogram_quantile(0.95, rate(bubblyui_inject_depth_bucket[5m]))"
        }]
      }
    ],
    "time": {"from": "now-1h", "to": "now"},
    "refresh": "30s"
  }
}
```

## 🚨 Alerting Rules

### Prometheus Alert Rules

Create `bubblyui_alerts.yml`:

```yaml
groups:
  - name: bubblyui
    interval: 30s
    rules:
      # Alert on slow composable creation
      - alert: SlowComposableCreation
        expr: |
          rate(bubblyui_composable_creation_seconds_sum[5m]) / 
          rate(bubblyui_composable_creation_seconds_count[5m]) > 0.01
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Slow composable creation detected"
          description: "Composable {{$labels.name}} is taking >10ms to create (current: {{$value}}ms)"
      
      # Alert on low cache hit rate
      - alert: LowCacheHitRate
        expr: bubblyui_cache_hit_ratio < 0.8
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Low cache hit rate"
          description: "Cache {{$labels.operation}} hit rate is {{$value}} (target: >80%)"
      
      # Alert on deep component trees
      - alert: DeepComponentTree
        expr: |
          histogram_quantile(0.95, rate(bubblyui_inject_depth_bucket[5m])) > 15
        for: 15m
        labels:
          severity: info
        annotations:
          summary: "Deep component tree detected"
          description: "P95 injection depth is {{$value}} (recommend: <10 levels)"
      
      # Alert on high memory allocation rate
      - alert: HighMemoryAllocationRate
        expr: |
          sum(rate(bubblyui_composable_allocations_bytes_sum[5m])) > 10000000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory allocation rate"
          description: "Allocating {{$value}} bytes/sec (>10MB/s)"
```

### Alertmanager Configuration

```yaml
# alertmanager.yml
route:
  group_by: ['alertname', 'severity']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 12h
  receiver: 'slack-notifications'

receivers:
  - name: 'slack-notifications'
    slack_configs:
      - api_url: 'YOUR_SLACK_WEBHOOK_URL'
        channel: '#bubblyui-alerts'
        title: '{{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
```

## 📐 Tree Depth Tracking

### Why Track Tree Depth?

Deep component trees can impact performance:
- More provide/inject lookups
- Slower context traversal
- Memory overhead
- Difficult debugging

**Recommended:** Keep trees < 10 levels deep

### Monitoring Tree Depth

```promql
# Current P95 depth
histogram_quantile(0.95, rate(bubblyui_inject_depth_bucket[5m]))

# Alert if consistently > 15 levels
histogram_quantile(0.95, rate(bubblyui_inject_depth_bucket[30m])) > 15
```

### Optimizing Deep Trees

**Problem:**
```
App (level 0)
  ├─ Layout (level 1)
  │   ├─ Header (level 2)
  │   │   ├─ Nav (level 3)
  │   │   │   ├─ Menu (level 4)
  │   │   │   │   ├─ Item (level 5)
  │   │   │   │   │   ├─ Link (level 6)
  │   │   │   │   │   │   ├─ Icon (level 7)
  │   │   │   │   │   │   │   ├─ SVG (level 8) 😱 Too deep!
```

**Solution:**
```
App (level 0)
  ├─ Header (level 1)
  │   ├─ Navigation (level 2) ← Flattened!
  │   │   ├─ MenuItem (level 3) ← Combined Menu+Item+Link
  ├─ Content (level 1)
```

## 🔍 Monitoring Best Practices

### What to Monitor

**Essential Metrics:**
1. ✅ Composable creation rate
2. ✅ Composable creation time (P95, P99)
3. ✅ Cache hit rate
4. ✅ Memory allocation rate
5. ✅ Component tree depth

**Nice to Have:**
6. ⭐ Specific composable usage (UseState vs UseForm)
7. ⭐ Error rates (if integrated with error tracking)
8. ⭐ GC pause time
9. ⭐ Goroutine count

**Don't Over-Monitor:**
- ❌ Individual component instances
- ❌ Every single state change
- ❌ Metrics with high cardinality
- ❌ Metrics that don't drive decisions

### Metric Retention

**Short-term (hours):** High-resolution data
- Retention: 2 hours
- Interval: 15s
- Use for: Real-time debugging

**Medium-term (days):** Trend analysis
- Retention: 7 days
- Interval: 1m
- Use for: Performance trends

**Long-term (months):** Historical comparison
- Retention: 90 days
- Interval: 5m
- Use for: Capacity planning

### Dashboard Organization

**Page 1: Overview**
- Composable creation rate
- Cache hit rate
- P95 creation time
- Tree depth

**Page 2: Performance**
- Detailed timing breakdowns
- Memory allocations
- GC metrics
- CPU usage

**Page 3: Debugging**
- Individual composable metrics
- Error rates
- Slow requests
- Anomalies

## 🎯 Performance Targets

Set alerts for these thresholds:

| Metric | Warning | Critical |
|--------|---------|----------|
| Creation time (P95) | > 10ms | > 50ms |
| Cache hit rate | < 80% | < 50% |
| Tree depth (P95) | > 10 levels | > 20 levels |
| Memory allocation | > 10MB/s | > 50MB/s |
| Composable rate | > 1000/s | > 5000/s |

## 📊 Sample Dashboards

### Development Dashboard

Focus: Real-time debugging

```
┌────────────────────────────────────────┐
│  Composable Creation Rate (1m)         │
│  ▓▓▓▓░░░░▓▓▓▓░░░░▓▓▓▓                 │
└────────────────────────────────────────┘
┌──────────────┬─────────────────────────┐
│  Cache Hits  │  Recent Errors          │
│  ████████ 95%│  0 errors (1h)          │
└──────────────┴─────────────────────────┘
```

### Production Dashboard

Focus: Long-term trends

```
┌────────────────────────────────────────┐
│  P95 Creation Time (24h trend)         │
│  ▁▂▂▃▃▄▄▅▅▆▆▇▇█                       │
└────────────────────────────────────────┘
┌──────────────┬─────────────────────────┐
│  Weekly Trend│  Capacity Planning      │
│  +5% usage   │  70% capacity           │
└──────────────┴─────────────────────────┘
```

## 🔗 Integration Examples

### Kubernetes

```yaml
# deployment.yml
apiVersion: v1
kind: Service
metadata:
  name: bubblyui-metrics
  labels:
    app: bubblyui
spec:
  ports:
    - port: 9090
      name: metrics
  selector:
    app: bubblyui

---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: bubblyui-monitor
spec:
  selector:
    matchLabels:
      app: bubblyui
  endpoints:
    - port: metrics
      interval: 30s
```

### Docker Compose

```yaml
version: '3'
services:
  app:
    build: .
    ports:
      - "8080:8080"
      - "9090:9090"  # Metrics
  
  prometheus:
    image: prom/prometheus
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9091:9090"
  
  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 30s
  evaluation_interval: 30s

scrape_configs:
  - job_name: 'bubblyui'
    static_configs:
      - targets: ['app:9090']
    metric_relabel_configs:
      - source_labels: [__name__]
        regex: 'bubblyui_.*'
        action: keep
```

## 📚 Related Guides

- [Performance Optimization Guide](./performance-optimization.md) - Optimization strategies
- [Production Profiling Guide](./production-profiling.md) - CPU/memory profiling
- [Benchmark Guide](./benchmark-guide.md) - Performance testing

## 💡 Monitoring Tips

1. **Start simple** - Monitor 5 key metrics first
2. **Set meaningful thresholds** - Based on your SLAs
3. **Test alerts** - Ensure they fire correctly
4. **Review regularly** - Weekly dashboard review
5. **Correlate metrics** - Look for patterns across metrics
6. **Document runbooks** - What to do when alerts fire
7. **Track trends** - Week-over-week comparisons
8. **Capacity planning** - Use metrics to predict growth

---

**Need help?** Check the [BubblyUI documentation](../../README.md) or [open an issue](https://github.com/newbpydev/bubblyui/issues).
