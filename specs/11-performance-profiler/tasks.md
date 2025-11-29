# Implementation Tasks: Performance Profiler

## Task Breakdown (Atomic Level)

### Prerequisites
- [x] 01-reactivity-system completed (State profiling)
- [x] 02-component-model completed (Component profiling)
- [x] 03-lifecycle-hooks completed (Hook timing)
- [ ] pprof library integrated
- [ ] Go testing/benchmarking understood

---

## Phase 1: Core Profiling Infrastructure (5 tasks, 15 hours)

### Task 1.1: Profiler Core ✅ COMPLETED
**Description**: Main profiler singleton and lifecycle management

**Prerequisites**: None

**Unlocks**: Task 1.2 (Metric Collector)

**Files**:
- `pkg/bubbly/profiler/profiler.go`
- `pkg/bubbly/profiler/profiler_test.go`

**Type Safety**:
```go
type Profiler struct {
    enabled    bool
    collector  *MetricCollector
    cpuProf    *CPUProfiler
    memProf    *MemoryProfiler
    renderProf *RenderProfiler
    detector   *BottleneckDetector
    config     *Config
    mu         sync.RWMutex
}

func New(opts ...Option) *Profiler
func (p *Profiler) Start() error
func (p *Profiler) Stop() error
func (p *Profiler) GenerateReport() *Report
```

**Tests**:
- [x] Profiler creation
- [x] Start/stop lifecycle
- [x] Enable/disable
- [x] Configuration options
- [x] Thread-safe operations

**Estimated Effort**: 3 hours

**Implementation Notes (Completed 2024-11-28)**:
- Created full `Profiler` struct with all fields from design spec
- Implemented Option pattern: `WithEnabled`, `WithSamplingRate`, `WithMaxSamples`, `WithMinimalMetrics`, `WithThreshold`
- Implemented lifecycle: `New()`, `Start()`, `Stop()`, `Enable()`, `Disable()`, `IsEnabled()`
- Implemented `GenerateReport()` returning properly initialized `Report` struct
- Added `Config` struct with `DefaultConfig()` and `Validate()` methods
- Added all type definitions from design spec: `Report`, `Summary`, `ComponentMetrics`, `BottleneckInfo`, `CPUProfileData`, `MemProfileData`, `Recommendation`
- Defined enums: `BottleneckType`, `Severity`, `Priority`, `Category`, `ImpactLevel`
- All stub types for future tasks: `MetricCollector`, `CPUProfiler`, `MemoryProfiler`, `RenderProfiler`, `BottleneckDetector`
- 10 table-driven tests with 100% coverage of test requirements
- **Coverage: 98.2%** (exceeds >80% requirement)
- All tests pass with race detector
- Thread-safe operations verified with 100 concurrent goroutines
- Zero lint warnings, proper formatting

---

### Task 1.2: Metric Collector ✅ COMPLETED
**Description**: Core metric collection system

**Prerequisites**: Task 1.1

**Unlocks**: Task 1.3 (Timing Tracker)

**Files**:
- `pkg/bubbly/profiler/collector.go`
- `pkg/bubbly/profiler/collector_test.go`

**Type Safety**:
```go
type MetricCollector struct {
    timings   *TimingTracker
    memory    *MemoryTracker
    counters  *CounterTracker
    enabled   bool
    mu        sync.RWMutex
}

func (mc *MetricCollector) Measure(name string, fn func())
func (mc *MetricCollector) StartTiming(name string) func()
func (mc *MetricCollector) RecordMetric(name string, value float64)
```

**Tests**:
- [x] Metric collection
- [x] Timing measurement
- [x] Counter tracking
- [x] Thread-safe access
- [x] Overhead < 3%

**Estimated Effort**: 3 hours

**Implementation Notes (Completed 2024-11-28)**:
- Created `MetricCollector` struct with all three trackers: `TimingTracker`, `MemoryTracker`, `CounterTracker`
- Implemented core methods: `Measure()`, `StartTiming()`, `RecordMetric()`, `IncrementCounter()`, `RecordMemory()`
- Used `atomic.Bool` for fast enabled/disabled checks (minimal overhead when disabled)
- Implemented `Enable()`, `Disable()`, `IsEnabled()`, `Reset()` for lifecycle management
- Created stub implementations for trackers with basic functionality:
  - `TimingTracker`: Records timing with Count, Total, Min, Max, Mean, and samples for percentiles
  - `MemoryTracker`: Tracks allocations with Count, TotalSize, AvgSize
  - `CounterTracker`: Tracks counters with Count and Value fields
- Thread-safe with `sync.RWMutex` protecting map operations
- 12 table-driven tests covering all functionality
- **Coverage: 82.6%** (exceeds >80% requirement)
- All tests pass with race detector
- Overhead measurements:
  - Disabled path: ~50-350ns/op (with race detector ~350ns)
  - Enabled path: ~4000-5000ns/op
- Zero lint warnings, proper formatting

---

### Task 1.3: Timing Tracker ✅ COMPLETED
**Description**: Track operation timings with statistics

**Prerequisites**: Task 1.2

**Unlocks**: Task 1.4 (Memory Tracker)

**Files**:
- `pkg/bubbly/profiler/timing.go`
- `pkg/bubbly/profiler/timing_test.go`

**Type Safety**:
```go
type TimingTracker struct {
    operations map[string]*TimingStats
    mu         sync.RWMutex
}

type TimingStats struct {
    Count   int64
    Total   time.Duration
    Min     time.Duration
    Max     time.Duration
    Mean    time.Duration
    P50     time.Duration
    P95     time.Duration
    P99     time.Duration
    samples []time.Duration
}

func (tt *TimingTracker) Record(name string, duration time.Duration)
func (tt *TimingTracker) GetStats(name string) *TimingStats
```

**Tests**:
- [x] Timing recording
- [x] Statistics calculation
- [x] Percentile calculation
- [x] Memory bounded (reservoir sampling)
- [x] Accuracy ±1ms

**Estimated Effort**: 4 hours

**Implementation Notes (Completed 2024-11-28)**:
- Created `timing.go` with full `TimingTracker` implementation (moved from collector.go stub)
- Enhanced `TimingStats` with P50, P95, P99 percentile fields
- Implemented percentile calculation using nearest-rank method with sorted samples
- Implemented reservoir sampling for memory bounding (default 10,000 samples per operation)
- Added `NewTimingTrackerWithMaxSamples()` for custom sample limits
- Added helper methods: `GetStatsSnapshot()`, `GetOperationNames()`, `Reset()`, `ResetOperation()`, `OperationCount()`, `SampleCount()`, `SampleCountForOperation()`
- Percentiles are calculated lazily on `GetStats()` call and cached until new samples arrive
- Thread-safe with `sync.RWMutex` protecting all operations
- 27 table-driven tests covering all functionality
- **Coverage: 96.6%** (exceeds >95% requirement)
- All tests pass with race detector
- Timing accuracy verified with exact duration tests
- Zero lint warnings, proper formatting

---

### Task 1.4: Memory Tracker ✅ COMPLETED
**Description**: Track memory allocations and usage

**Prerequisites**: Task 1.3

**Unlocks**: Task 1.5 (Configuration System)

**Files**:
- `pkg/bubbly/profiler/memory.go`
- `pkg/bubbly/profiler/memory_test.go`

**Type Safety**:
```go
type MemoryTracker struct {
    snapshots   []*runtime.MemStats
    allocations map[string]*AllocationStats
    mu          sync.RWMutex
}

type AllocationStats struct {
    Count     int64
    TotalSize int64
    AvgSize   int64
}

func (mt *MemoryTracker) TakeSnapshot() *runtime.MemStats
func (mt *MemoryTracker) TrackAllocation(location string, size int64)
```

**Tests**:
- [x] Snapshot collection
- [x] Allocation tracking
- [x] Memory statistics
- [x] Growth detection
- [x] Thread-safe operations

**Estimated Effort**: 3 hours

**Implementation Notes (Completed 2024-11-28)**:
- Created `memory.go` with full `MemoryTracker` implementation from design spec
- Implemented `TakeSnapshot()` using `runtime.ReadMemStats()` to capture memory state
- Implemented `TrackAllocation()` for tracking allocations by location with Count, TotalSize, AvgSize
- Added snapshot management: `GetAllSnapshots()`, `GetSnapshotAt()`, `GetFirstSnapshot()`, `GetLatestSnapshot()`
- Added growth detection: `GetMemoryGrowth()` (heap), `GetHeapObjectGrowth()`, `GetGoroutineGrowth()`
- Added helper methods: `SnapshotCount()`, `AllocationCount()`, `GetTotalAllocatedSize()`, `GetAllocationLocations()`
- Added `Reset()` for clearing all data
- Thread-safe with `sync.RWMutex` protecting all operations
- Updated `collector.go` to use `NewMemoryTracker()` from memory.go
- 28 table-driven tests covering all functionality
- **Coverage: 97.4%** (exceeds >95% requirement)
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

### Task 1.5: Configuration System ✅ COMPLETED
**Description**: Profiler configuration and options

**Prerequisites**: Task 1.4

**Unlocks**: Task 2.1 (CPU Profiler)

**Files**:
- `pkg/bubbly/profiler/config.go`
- `pkg/bubbly/profiler/config_test.go`

**Type Safety**:
```go
type Config struct {
    Enabled        bool
    SamplingRate   float64
    MaxSamples     int
    MinimalMetrics bool
    Thresholds     map[string]time.Duration
}

type Option func(*Config)

func WithSamplingRate(rate float64) Option
func WithMinimalMetrics() Option
func WithThreshold(operation string, threshold time.Duration) Option
```

**Tests**:
- [x] Default config
- [x] Options pattern
- [x] Validation
- [x] Override behavior
- [x] Env var loading

**Estimated Effort**: 2 hours

**Implementation Notes (Completed 2024-11-28)**:
- Created `config.go` with dedicated configuration management
- Implemented environment variable loading via `ConfigFromEnv()` and `LoadFromEnv()` methods
- Environment variables supported:
  - `BUBBLY_PROFILER_ENABLED` (bool: "true", "1", "false", "0")
  - `BUBBLY_PROFILER_SAMPLING_RATE` (float64: 0.0 to 1.0)
  - `BUBBLY_PROFILER_MAX_SAMPLES` (int: positive integer)
  - `BUBBLY_PROFILER_MINIMAL_METRICS` (bool)
- Invalid env var values silently use defaults (safe deployment)
- Added `Clone()` method for deep copying Config
- Added `String()` method for debugging
- Added `ApplyOptions()` helper function
- Override behavior: defaults → env vars → options (options win)
- 20 table-driven tests covering all functionality
- **Coverage: 97.9%** (exceeds >95% requirement)
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

## Phase 2: CPU & Memory Profiling (4 tasks, 12 hours)

### Task 2.1: CPU Profiler (pprof Integration) ✅ COMPLETED
**Description**: Integrate with Go's pprof for CPU profiling

**Prerequisites**: Task 1.5

**Unlocks**: Task 2.2 (Stack Analyzer)

**Files**:
- `pkg/bubbly/profiler/cpu.go`
- `pkg/bubbly/profiler/cpu_test.go`

**Type Safety**:
```go
type CPUProfiler struct {
    active bool
    file   *os.File
    mu     sync.Mutex
}

func (cp *CPUProfiler) Start(filename string) error
func (cp *CPUProfiler) Stop() error
func (cp *CPUProfiler) IsActive() bool
```

**Tests**:
- [x] Start profiling
- [x] Stop profiling
- [x] File generation
- [x] pprof format valid
- [x] Integration with pprof tools

**Estimated Effort**: 3 hours

**Implementation Notes (Completed 2024-11-29)**:
- Created `cpu.go` with full `CPUProfiler` implementation using `runtime/pprof`
- Implemented `NewCPUProfiler()` constructor returning inactive profiler
- Implemented `Start(filename)` using `pprof.StartCPUProfile()` with file creation
- Implemented `Stop()` using `pprof.StopCPUProfile()` with proper file cleanup
- Implemented `IsActive()` for thread-safe status checking
- Added `GetFilename()` helper method for retrieving current profile filename
- Added error types: `ErrCPUProfileActive`, `ErrCPUProfileNotActive`
- Thread-safe with `sync.Mutex` protecting all state operations
- Updated `profiler.go` to use `NewCPUProfiler()` instead of stub
- 9 table-driven tests covering all functionality:
  - Start/stop lifecycle with error handling
  - File generation and pprof format validation (gzip magic number check)
  - Invalid filename handling
  - Thread-safe concurrent access (50+ goroutines)
  - Multiple start/stop cycles
  - Integration with pprof tools
- **Coverage: 97.6%** (exceeds >95% requirement)
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

### Task 2.2: Stack Analyzer ✅ COMPLETED
**Description**: Analyze CPU profile data

**Prerequisites**: Task 2.1

**Unlocks**: Task 2.3 (Memory Profiler)

**Files**:
- `pkg/bubbly/profiler/stack_analyzer.go`
- `pkg/bubbly/profiler/stack_analyzer_test.go`

**Type Safety**:
```go
type StackAnalyzer struct {
    samples map[string]int64
}

type CPUProfileData struct {
    HotFunctions []*HotFunction
    CallGraph    map[string][]string
    TotalSamples int64
}

type HotFunction struct {
    Name    string
    Samples int64
    Percent float64
}

func (sa *StackAnalyzer) Analyze(profile *pprof.Profile) *CPUProfileData
```

**Tests**:
- [x] Profile parsing
- [x] Hot function detection
- [x] Call graph building
- [x] Percentage calculation
- [x] Sorting by samples

**Estimated Effort**: 3 hours

**Implementation Notes (Completed 2024-11-29)**:
- Created `stack_analyzer.go` with full `StackAnalyzer` implementation using `github.com/google/pprof/profile`
- Implemented `NewStackAnalyzer()` constructor returning analyzer with empty samples map
- Implemented `Analyze(*profile.Profile)` method that:
  - Extracts function names from sample locations
  - Counts samples per function
  - Calculates percentage of total CPU time
  - Builds call graph from stack traces (caller → callee relationships)
  - Sorts hot functions by sample count descending
- Implemented `buildCallGraph()` helper to extract caller-callee relationships from stack traces
- Implemented `Reset()` to clear internal state for fresh analysis
- Implemented `GetSamples()` to return copy of sample counts
- Added helper functions: `getFirstFunctionName()`, `containsString()`
- Thread-safe with `sync.RWMutex` protecting all operations
- 10 table-driven tests covering all functionality:
  - NewStackAnalyzer creation
  - Analyze with nil/empty/valid profiles
  - Hot function detection and sorting
  - Call graph building from stack traces
  - Percentage calculation accuracy
  - Thread-safe concurrent access (50 goroutines)
  - File-based profile parsing integration
  - Reset and GetSamples methods
- **Coverage: 96.5%** (exceeds >95% requirement)
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

### Task 2.3: Memory Profiler (pprof Integration)
**Description**: Heap profiling and analysis

**Prerequisites**: Task 2.2

**Unlocks**: Task 2.4 (Leak Detector)

**Files**:
- `pkg/bubbly/profiler/heap.go`
- `pkg/bubbly/profiler/heap_test.go`

**Type Safety**:
```go
type MemoryProfiler struct {
    baseline  *runtime.MemStats
    snapshots []*runtime.MemStats
}

func (mp *MemoryProfiler) TakeSnapshot() *runtime.MemStats
func (mp *MemoryProfiler) WriteHeapProfile(filename string) error
func (mp *MemoryProfiler) GetMemoryGrowth() int64
```

**Tests**:
- [ ] Snapshot capture
- [ ] Heap profile generation
- [ ] Growth calculation
- [ ] pprof format valid
- [ ] Integration with tools

**Estimated Effort**: 3 hours

---

### Task 2.4: Memory Leak Detector
**Description**: Detect memory leaks from snapshots

**Prerequisites**: Task 2.3

**Unlocks**: Task 3.1 (Render Profiler)

**Files**:
- `pkg/bubbly/profiler/leak_detector.go`
- `pkg/bubbly/profiler/leak_detector_test.go`

**Type Safety**:
```go
type LeakDetector struct {
    thresholds *LeakThresholds
}

type LeakInfo struct {
    Type        string
    BytesLeaked int64
    Count       int
    Description string
    Severity    Severity
}

func (ld *LeakDetector) DetectLeaks(snapshots []*runtime.MemStats) []*LeakInfo
func (ld *LeakDetector) DetectGoroutineLeaks(before, after int) *LeakInfo
```

**Tests**:
- [ ] Heap leak detection
- [ ] Goroutine leak detection
- [ ] Severity calculation
- [ ] False positive filtering
- [ ] Threshold configuration

**Estimated Effort**: 3 hours

---

## Phase 3: Render Performance (3 tasks, 9 hours)

### Task 3.1: Render Profiler
**Description**: Track render performance and FPS

**Prerequisites**: Task 2.4

**Unlocks**: Task 3.2 (FPS Calculator)

**Files**:
- `pkg/bubbly/profiler/render.go`
- `pkg/bubbly/profiler/render_test.go`

**Type Safety**:
```go
type RenderProfiler struct {
    frames      []FrameInfo
    lastFrame   time.Time
    fpsSamples  []float64
}

type FrameInfo struct {
    Timestamp time.Time
    Duration  time.Duration
    Dropped   bool
}

func (rp *RenderProfiler) RecordFrame(duration time.Duration)
func (rp *RenderProfiler) GetFPS() float64
func (rp *RenderProfiler) GetDroppedFramePercent() float64
```

**Tests**:
- [ ] Frame recording
- [ ] FPS calculation
- [ ] Dropped frame detection
- [ ] Statistics accurate
- [ ] Performance acceptable

**Estimated Effort**: 3 hours

---

### Task 3.2: FPS Calculator
**Description**: Calculate frames per second metrics

**Prerequisites**: Task 3.1

**Unlocks**: Task 3.3 (Component Tracker)

**Files**:
- `pkg/bubbly/profiler/fps.go`
- `pkg/bubbly/profiler/fps_test.go`

**Type Safety**:
```go
type FPSCalculator struct {
    samples    []float64
    windowSize int
}

func (fc *FPSCalculator) AddSample(fps float64)
func (fc *FPSCalculator) GetAverage() float64
func (fc *FPSCalculator) GetMin() float64
func (fc *FPSCalculator) GetMax() float64
```

**Tests**:
- [ ] Sample collection
- [ ] Average calculation
- [ ] Min/max tracking
- [ ] Window size respected
- [ ] Accuracy validation

**Estimated Effort**: 2 hours

---

### Task 3.3: Component Performance Tracker
**Description**: Track per-component performance metrics

**Prerequisites**: Task 3.2

**Unlocks**: Task 4.1 (Bottleneck Detector)

**Files**:
- `pkg/bubbly/profiler/component_tracker.go`
- `pkg/bubbly/profiler/component_tracker_test.go`

**Type Safety**:
```go
type ComponentTracker struct {
    components map[string]*ComponentMetrics
}

type ComponentMetrics struct {
    ComponentID     string
    ComponentName   string
    RenderCount     int64
    TotalRenderTime time.Duration
    AvgRenderTime   time.Duration
    MaxRenderTime   time.Duration
    MemoryUsage     uint64
}

func (ct *ComponentTracker) RecordRender(id, name string, duration time.Duration)
func (ct *ComponentTracker) GetMetrics(id string) *ComponentMetrics
```

**Tests**:
- [ ] Component tracking
- [ ] Metric aggregation
- [ ] Statistics calculation
- [ ] Multiple components
- [ ] Thread-safe access

**Estimated Effort**: 4 hours

---

## Phase 4: Bottleneck Detection (4 tasks, 12 hours)

### Task 4.1: Bottleneck Detector Core
**Description**: Detect performance bottlenecks

**Prerequisites**: Task 3.3

**Unlocks**: Task 4.2 (Threshold Monitor)

**Files**:
- `pkg/bubbly/profiler/bottleneck.go`
- `pkg/bubbly/profiler/bottleneck_test.go`

**Type Safety**:
```go
type BottleneckDetector struct {
    thresholds map[string]time.Duration
    violations map[string]int
}

type BottleneckInfo struct {
    Type        BottleneckType
    Location    string
    Severity    Severity
    Impact      float64
    Description string
    Suggestion  string
}

func (bd *BottleneckDetector) Detect(metrics *PerformanceMetrics) []*BottleneckInfo
```

**Tests**:
- [ ] Detection works
- [ ] Severity calculation
- [ ] Impact measurement
- [ ] Suggestion generation
- [ ] Multiple bottlenecks

**Estimated Effort**: 3 hours

---

### Task 4.2: Threshold Monitor
**Description**: Monitor operations against thresholds

**Prerequisites**: Task 4.1

**Unlocks**: Task 4.3 (Pattern Analyzer)

**Files**:
- `pkg/bubbly/profiler/threshold.go`
- `pkg/bubbly/profiler/threshold_test.go`

**Type Safety**:
```go
type ThresholdMonitor struct {
    thresholds map[string]time.Duration
    violations map[string]int
}

func (tm *ThresholdMonitor) Check(operation string, duration time.Duration) *BottleneckInfo
func (tm *ThresholdMonitor) SetThreshold(operation string, threshold time.Duration)
```

**Tests**:
- [ ] Threshold checking
- [ ] Violation tracking
- [ ] Configurable thresholds
- [ ] Multiple operations
- [ ] Alert generation

**Estimated Effort**: 2 hours

---

### Task 4.3: Pattern Analyzer
**Description**: Analyze performance patterns

**Prerequisites**: Task 4.2

**Unlocks**: Task 4.4 (Recommendation Engine)

**Files**:
- `pkg/bubbly/profiler/pattern_analyzer.go`
- `pkg/bubbly/profiler/pattern_analyzer_test.go`

**Type Safety**:
```go
type PatternAnalyzer struct {
    patterns []Pattern
}

type Pattern struct {
    Name        string
    Detect      func(*ComponentMetrics) bool
    Severity    Severity
    Description string
    Suggestion  string
}

func (pa *PatternAnalyzer) Analyze(metrics *ComponentMetrics) []*BottleneckInfo
```

**Tests**:
- [ ] Pattern detection
- [ ] Multiple patterns
- [ ] Custom patterns
- [ ] Severity assignment
- [ ] Suggestion generation

**Estimated Effort**: 4 hours

---

### Task 4.4: Recommendation Engine
**Description**: Generate optimization recommendations

**Prerequisites**: Task 4.3

**Unlocks**: Task 5.1 (Report Generator)

**Files**:
- `pkg/bubbly/profiler/recommendations.go`
- `pkg/bubbly/profiler/recommendations_test.go`

**Type Safety**:
```go
type RecommendationEngine struct {
    rules []RecommendationRule
}

type RecommendationRule struct {
    Name        string
    Condition   func(*PerformanceReport) bool
    Priority    Priority
    Category    Category
    Title       string
    Description string
    Action      string
    Impact      ImpactLevel
}

func (re *RecommendationEngine) Generate(report *PerformanceReport) []*Recommendation
```

**Tests**:
- [ ] Rule evaluation
- [ ] Priority sorting
- [ ] Multiple recommendations
- [ ] Custom rules
- [ ] Actionable suggestions

**Estimated Effort**: 3 hours

---

## Phase 5: Reporting & Visualization (5 tasks, 15 hours)

### Task 5.1: Report Generator Core
**Description**: Generate performance reports

**Prerequisites**: Task 4.4

**Unlocks**: Task 5.2 (Data Aggregator)

**Files**:
- `pkg/bubbly/profiler/report.go`
- `pkg/bubbly/profiler/report_test.go`

**Type Safety**:
```go
type ReportGenerator struct {
    templates *template.Template
}

type PerformanceReport struct {
    Summary         *Summary
    Components      []*ComponentMetrics
    Bottlenecks     []*BottleneckInfo
    CPUProfile      *CPUProfileData
    MemProfile      *MemProfileData
    Recommendations []*Recommendation
    Timestamp       time.Time
}

func (rg *ReportGenerator) Generate(data *ProfileData) *PerformanceReport
func (rg *ReportGenerator) SaveHTML(filename string) error
```

**Tests**:
- [ ] Report generation
- [ ] All sections included
- [ ] Data aggregation correct
- [ ] HTML generation
- [ ] Template rendering

**Estimated Effort**: 3 hours

---

### Task 5.2: Data Aggregator
**Description**: Aggregate profiling data for reporting

**Prerequisites**: Task 5.1

**Unlocks**: Task 5.3 (Flame Graph Generator)

**Files**:
- `pkg/bubbly/profiler/aggregator.go`
- `pkg/bubbly/profiler/aggregator_test.go`

**Type Safety**:
```go
type DataAggregator struct{}

func (da *DataAggregator) Aggregate(collector *MetricCollector) *AggregatedData
func (da *DataAggregator) CalculateSummary(data *AggregatedData) *Summary
```

**Tests**:
- [ ] Data aggregation
- [ ] Summary calculation
- [ ] Statistics correct
- [ ] All metrics included
- [ ] Performance acceptable

**Estimated Effort**: 3 hours

---

### Task 5.3: Flame Graph Generator
**Description**: Generate flame graphs from CPU profile

**Prerequisites**: Task 5.2

**Unlocks**: Task 5.4 (Timeline Visualizer)

**Files**:
- `pkg/bubbly/profiler/flamegraph.go`
- `pkg/bubbly/profiler/flamegraph_test.go`

**Type Safety**:
```go
type FlameGraphGenerator struct {
    width  int
    height int
}

type CallNode struct {
    Name     string
    Samples  int64
    Percent  float64
    Children []*CallNode
}

func (fgg *FlameGraphGenerator) Generate(profile *CPUProfileData) string
func (fgg *FlameGraphGenerator) GenerateSVG() string
```

**Tests**:
- [ ] Flame graph generation
- [ ] SVG format valid
- [ ] Nested calls shown
- [ ] Percentages correct
- [ ] Visual output correct

**Estimated Effort**: 4 hours

---

### Task 5.4: Timeline Visualizer
**Description**: Generate timeline visualization

**Prerequisites**: Task 5.3

**Unlocks**: Task 5.5 (Export System)

**Files**:
- `pkg/bubbly/profiler/timeline.go`
- `pkg/bubbly/profiler/timeline_test.go`

**Type Safety**:
```go
type TimelineGenerator struct {
    width int
}

func (tg *TimelineGenerator) Generate(events []TimedEvent) string
func (tg *TimelineGenerator) GenerateHTML() string
```

**Tests**:
- [ ] Timeline generation
- [ ] Event ordering
- [ ] Time scaling
- [ ] Visual clarity
- [ ] HTML output

**Estimated Effort**: 3 hours

---

### Task 5.5: Export System
**Description**: Export reports in multiple formats

**Prerequisites**: Task 5.4

**Unlocks**: Task 6.1 (Benchmark Integration)

**Files**:
- `pkg/bubbly/profiler/export.go`
- `pkg/bubbly/profiler/export_test.go`

**Type Safety**:
```go
type Exporter struct{}

func (e *Exporter) ExportHTML(report *PerformanceReport, filename string) error
func (e *Exporter) ExportJSON(report *PerformanceReport, filename string) error
func (e *Exporter) ExportCSV(report *PerformanceReport, filename string) error
```

**Tests**:
- [ ] HTML export
- [ ] JSON export
- [ ] CSV export
- [ ] File creation
- [ ] Format validation

**Estimated Effort**: 2 hours

---

## Phase 6: Integration & Tools (4 tasks, 12 hours)

### Task 6.1: Benchmark Integration
**Description**: Integration with Go testing benchmarks

**Prerequisites**: Task 5.5

**Unlocks**: Task 6.2 (Component Instrumentation)

**Files**:
- `pkg/bubbly/profiler/benchmark.go`
- `pkg/bubbly/profiler/benchmark_test.go`

**Type Safety**:
```go
type BenchmarkProfiler struct {
    b       *testing.B
    metrics *MetricCollector
}

func NewBenchmarkProfiler(b *testing.B) *BenchmarkProfiler
func (bp *BenchmarkProfiler) Measure(fn func())
func (bp *BenchmarkProfiler) AssertNoRegression(baseline *Baseline, threshold float64)
```

**Tests**:
- [ ] Benchmark integration
- [ ] Measurement works
- [ ] Regression detection
- [ ] Baseline comparison
- [ ] CI/CD compatible

**Estimated Effort**: 3 hours

---

### Task 6.2: Component Instrumentation
**Description**: Instrument components for profiling

**Prerequisites**: Task 6.1

**Unlocks**: Task 6.3 (Dev Tools Integration)

**Files**:
- `pkg/bubbly/profiler/instrumentation.go`
- `pkg/bubbly/profiler/instrumentation_test.go`

**Type Safety**:
```go
type Instrumentor struct {
    profiler *Profiler
}

func (i *Instrumentor) InstrumentComponent(component Component)
func (i *Instrumentor) InstrumentRender(component Component)
func (i *Instrumentor) InstrumentUpdate(component Component)
```

**Tests**:
- [ ] Component instrumentation
- [ ] Render tracking
- [ ] Update tracking
- [ ] Minimal overhead
- [ ] No breaking changes

**Estimated Effort**: 3 hours

---

### Task 6.3: Dev Tools Integration
**Description**: Integrate with dev tools for visualization

**Prerequisites**: Task 6.2

**Unlocks**: Task 6.4 (Documentation)

**Files**:
- `pkg/bubbly/profiler/devtools.go`
- `pkg/bubbly/profiler/devtools_test.go`

**Type Safety**:
```go
type DevToolsIntegration struct {
    profiler *Profiler
    devtools *devtools.DevTools
}

func (dti *DevToolsIntegration) SendMetrics()
func (dti *DevToolsIntegration) RegisterPanel()
```

**Tests**:
- [ ] Integration works
- [ ] Metrics sent
- [ ] Panel displays
- [ ] Real-time updates
- [ ] No performance impact

**Estimated Effort**: 3 hours

---

### Task 6.4: pprof HTTP Handlers
**Description**: HTTP handlers for pprof access

**Prerequisites**: Task 6.3

**Unlocks**: Task 7.1 (Documentation)

**Files**:
- `pkg/bubbly/profiler/http.go`
- `pkg/bubbly/profiler/http_test.go`

**Type Safety**:
```go
func RegisterHandlers(mux *http.ServeMux, profiler *Profiler)
func ServeCPUProfile(w http.ResponseWriter, r *http.Request)
func ServeHeapProfile(w http.ResponseWriter, r *http.Request)
```

**Tests**:
- [ ] Handlers register
- [ ] CPU profile served
- [ ] Heap profile served
- [ ] HTTP integration
- [ ] Production-safe

**Estimated Effort**: 3 hours

---

## Phase 7: Documentation & Examples (3 tasks, 9 hours)

### Task 7.1: API Documentation
**Description**: Comprehensive godoc for profiler

**Prerequisites**: Task 6.4

**Unlocks**: Task 7.2 (User Guide)

**Files**:
- All package files (add/update godoc)

**Documentation**:
- Profiler API
- Configuration options
- Report format
- Integration guide
- Best practices

**Estimated Effort**: 2 hours

---

### Task 7.2: Performance Guide
**Description**: Complete performance optimization guide

**Prerequisites**: Task 7.1

**Unlocks**: Task 7.3 (Examples)

**Files**:
- `docs/performance/README.md`
- `docs/performance/profiling.md`
- `docs/performance/optimization.md`
- `docs/performance/benchmarking.md`

**Content**:
- Getting started
- CPU profiling guide
- Memory profiling guide
- Optimization workflows
- Best practices
- Common patterns

**Estimated Effort**: 4 hours

---

### Task 7.3: Example Applications
**Description**: Complete profiling examples

**Prerequisites**: Task 7.2

**Unlocks**: Feature complete

**Files**:
- `cmd/examples/11-profiler/basic/main.go`
- `cmd/examples/11-profiler/cpu/main.go`
- `cmd/examples/11-profiler/memory/main.go`
- `cmd/examples/11-profiler/benchmark/main_test.go`

**Examples**:
- Basic profiling
- CPU profiling workflow
- Memory leak detection
- Benchmark examples
- CI integration

**Estimated Effort**: 3 hours

---

## Task Dependency Graph

```
Prerequisites (Features 01-03, pprof)
    ↓
Phase 1: Infrastructure
    1.1 Core → 1.2 Collector → 1.3 Timing → 1.4 Memory → 1.5 Config
    ↓
Phase 2: CPU & Memory
    2.1 CPU Prof → 2.2 Stack → 2.3 Heap → 2.4 Leak Detector
    ↓
Phase 3: Render
    3.1 Render Prof → 3.2 FPS → 3.3 Component Tracker
    ↓
Phase 4: Bottlenecks
    4.1 Detector → 4.2 Threshold → 4.3 Patterns → 4.4 Recommendations
    ↓
Phase 5: Reporting
    5.1 Generator → 5.2 Aggregator → 5.3 Flame → 5.4 Timeline → 5.5 Export
    ↓
Phase 6: Integration
    6.1 Benchmarks → 6.2 Instrumentation → 6.3 DevTools → 6.4 HTTP
    ↓
Phase 7: Documentation
    7.1 API Docs → 7.2 Guide → 7.3 Examples
```

---

## Validation Checklist

### Core Functionality
- [ ] Profiler starts/stops
- [ ] Metrics collected
- [ ] CPU profiling works
- [ ] Memory profiling works
- [ ] Reports generated

### Performance
- [ ] Overhead < 3% when enabled
- [ ] Overhead < 0.1% when disabled
- [ ] Timing accuracy ±1ms
- [ ] Memory tracking accurate
- [ ] Scalable to large apps

### Accuracy
- [ ] Bottlenecks detected correctly
- [ ] Recommendations actionable
- [ ] Statistics accurate
- [ ] No false positives
- [ ] Reproducible results

### Integration
- [ ] pprof tools work
- [ ] Benchmark integration
- [ ] Dev tools integration
- [ ] HTTP handlers work
- [ ] CI/CD compatible

### Usability
- [ ] Easy setup
- [ ] Clear reports
- [ ] Actionable insights
- [ ] Good documentation
- [ ] Helpful examples

---

## Estimated Total Effort

- Phase 1: 15 hours
- Phase 2: 12 hours
- Phase 3: 9 hours
- Phase 4: 12 hours
- Phase 5: 15 hours
- Phase 6: 12 hours
- Phase 7: 9 hours

**Total**: ~84 hours (approximately 2.5 weeks)

---

## Priority

**HIGH** - Critical for performance optimization and production monitoring.

**Timeline**: Implement after Features 01-08 complete. Can develop alongside Feature 09 (Dev Tools) for integration.

**Unlocks**:
- Performance optimization workflow
- Production monitoring
- Benchmark-driven development
- Performance regression detection
- Bottleneck identification
- Actionable optimization insights
