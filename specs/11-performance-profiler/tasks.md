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

### Task 2.3: Memory Profiler (pprof Integration) ✅ COMPLETED
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
- [x] Snapshot capture
- [x] Heap profile generation
- [x] Growth calculation
- [x] pprof format valid
- [x] Integration with tools

**Estimated Effort**: 3 hours

**Implementation Notes (Completed 2024-11-29)**:
- Created `heap.go` with full `MemoryProfiler` implementation using `runtime/pprof`
- Implemented `NewMemoryProfiler()` constructor that captures baseline memory snapshot
- Implemented `TakeSnapshot()` using `runtime.ReadMemStats()` to capture memory state
- Implemented `WriteHeapProfile(filename)` using `pprof.WriteHeapProfile()` for pprof-compatible output
- Implemented `GetMemoryGrowth()` returning heap allocation difference between first and last snapshots
- Added helper methods: `GetBaseline()`, `GetSnapshots()`, `GetLatestSnapshot()`, `SnapshotCount()`, `Reset()`
- Thread-safe with `sync.RWMutex` protecting all operations
- Updated `profiler.go` to use `NewMemoryProfiler()` instead of stub
- 13 table-driven tests covering all functionality:
  - NewMemoryProfiler creation with baseline
  - TakeSnapshot captures memory stats
  - WriteHeapProfile generates valid pprof format (gzip compressed)
  - GetMemoryGrowth calculation with multiple snapshots
  - Insufficient snapshots returns zero
  - Thread-safe concurrent access (50 goroutines)
  - Integration with pprof tools (valid gzip decompression)
- **Coverage: 96.6%** (exceeds >95% requirement)
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

### Task 2.4: Memory Leak Detector ✅ COMPLETED
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
- [x] Heap leak detection
- [x] Goroutine leak detection
- [x] Severity calculation
- [x] False positive filtering
- [x] Threshold configuration

**Estimated Effort**: 3 hours

**Implementation Notes (Completed 2024-11-29)**:
- Created `leak_detector.go` with full `LeakDetector` implementation
- Implemented `LeakThresholds` struct with configurable thresholds:
  - `HeapGrowthBytes`: Minimum heap growth to report (default 1MB)
  - `GoroutineGrowth`: Minimum goroutine increase to report (default 10)
  - `HeapObjectGrowth`: Minimum object count increase (default 10,000)
  - Severity thresholds for High/Critical classification
- Implemented `LeakInfo` struct with Type, BytesLeaked, Count, Description, Severity
- Implemented `DetectLeaks()` that analyzes memory snapshots:
  - Compares first and last snapshots for heap growth
  - Detects heap object accumulation
  - Filters false positives using thresholds
  - Returns multiple leak types if detected
- Implemented `DetectGoroutineLeaks()` for goroutine leak detection
- Implemented severity calculation based on configurable thresholds:
  - Low: Below medium threshold
  - Medium: >= high/2 threshold
  - High: >= high threshold
  - Critical: >= critical threshold
- Added helper methods: `GetThresholds()`, `SetThresholds()`, `Reset()`
- Added `formatBytes()` helper for human-readable byte formatting
- Thread-safe with `sync.RWMutex` protecting all operations
- 17 table-driven tests covering all functionality
- **Coverage: 96.1%** (exceeds >95% requirement)
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

## Phase 3: Render Performance (3 tasks, 9 hours)

### Task 3.1: Render Profiler ✅ COMPLETED
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
- [x] Frame recording
- [x] FPS calculation
- [x] Dropped frame detection
- [x] Statistics accurate
- [x] Performance acceptable

**Estimated Effort**: 3 hours

**Implementation Notes (Completed 2024-11-29)**:
- Created `render.go` with full `RenderProfiler` implementation
- Implemented `RenderConfig` struct with configurable settings:
  - `TargetFPS`: Target frames per second (default 60)
  - `MaxFrames`: Maximum frames to store (default 1000)
  - `MaxFPSSamples`: Maximum FPS samples for averaging (default 60)
  - `DroppedFrameThreshold`: Duration above which frame is dropped (~16.67ms for 60fps)
- Implemented `FrameInfo` struct with Timestamp, Duration, Dropped fields
- Implemented `RecordFrame()` that:
  - Records frame with timestamp and duration
  - Detects dropped frames exceeding threshold
  - Calculates instantaneous FPS from frame intervals
  - Maintains frame and FPS sample limits
- Implemented `GetFPS()` returning average FPS from recent samples
- Implemented `GetDroppedFramePercent()` returning percentage of dropped frames
- Added helper methods: `GetFrames()`, `GetFrameCount()`, `GetAverageFrameDuration()`, `GetMinMaxFrameDuration()`
- Added configuration methods: `GetConfig()`, `SetConfig()`, `Reset()`
- Thread-safe with `sync.RWMutex` protecting all operations
- 17 table-driven tests covering all functionality including performance tests
- Removed stub `RenderProfiler` from profiler.go
- **Coverage: 96.6%** (exceeds >95% requirement)
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

### Task 3.2: FPS Calculator ✅ COMPLETED
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
- [x] Sample collection
- [x] Average calculation
- [x] Min/max tracking
- [x] Window size respected
- [x] Accuracy validation

**Estimated Effort**: 2 hours

**Implementation Notes (Completed 2024-11-29)**:
- Created `fps.go` with full `FPSCalculator` implementation
- Implemented `NewFPSCalculator()` and `NewFPSCalculatorWithWindowSize()` constructors
- Implemented core methods: `AddSample()`, `GetAverage()`, `GetMin()`, `GetMax()`, `GetMinMax()`
- Added `DefaultFPSWindowSize` constant (60 samples)
- Implemented sliding window with automatic oldest sample removal
- Added statistical methods: `GetStandardDeviation()`, `GetPercentile()`, `IsStable()`
- Added helper methods: `SampleCount()`, `GetSamples()`, `GetWindowSize()`, `SetWindowSize()`, `Reset()`
- Thread-safe with `sync.RWMutex` protecting all operations
- 18 table-driven tests covering all functionality:
  - Sample collection and window size limits
  - Average, min, max calculations
  - Standard deviation and percentile calculations
  - Stability detection with threshold
  - Thread-safe concurrent access (50 goroutines)
  - Accuracy validation with floating point precision
- **Coverage: 96.7%** (exceeds >95% requirement)
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

### Task 3.3: Component Performance Tracker ✅ COMPLETED
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
- [x] Component tracking
- [x] Metric aggregation
- [x] Statistics calculation
- [x] Multiple components
- [x] Thread-safe access

**Estimated Effort**: 4 hours

**Implementation Notes (Completed 2024-11-29)**:
- Created `component_tracker.go` with full `ComponentTracker` implementation
- Implemented `NewComponentTracker()` constructor returning empty tracker
- Implemented `RecordRender(id, name, duration)` with automatic statistics calculation:
  - Increments RenderCount
  - Updates TotalRenderTime
  - Calculates AvgRenderTime
  - Tracks MinRenderTime and MaxRenderTime
- Implemented `GetMetrics(id)` returning pointer to internal metrics (nil if not found)
- Implemented `GetMetricsSnapshot(id)` returning safe copy for external use
- Added helper methods: `GetAllMetrics()`, `GetComponentIDs()`, `ComponentCount()`, `TotalRenderCount()`
- Added `Reset()` and `ResetComponent(id)` for clearing data
- Added `RecordMemoryUsage(id, bytes)` for memory tracking (only updates existing components)
- Added `GetTopComponents(n, sortBy)` with sorting options:
  - `SortByTotalRenderTime` (default)
  - `SortByRenderCount`
  - `SortByAvgRenderTime`
  - `SortByMaxRenderTime`
- Added `MinRenderTime` field to `ComponentMetrics` in profiler.go for complete statistics
- Thread-safe with `sync.RWMutex` protecting all operations
- 27 table-driven tests covering all functionality including concurrent access (50 goroutines)
- **Coverage: 96.6%** (exceeds >95% requirement)
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

## Phase 4: Bottleneck Detection (4 tasks, 12 hours)

### Task 4.1: Bottleneck Detector Core ✅ COMPLETED
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
- [x] Detection works
- [x] Severity calculation
- [x] Impact measurement
- [x] Suggestion generation
- [x] Multiple bottlenecks

**Estimated Effort**: 3 hours

**Implementation Notes (Completed 2024-11-29)**:
- Created `bottleneck.go` with full `BottleneckDetector` implementation
- Implemented `BottleneckThresholds` struct with configurable thresholds:
  - `DefaultOperationThreshold`: 16ms (60 FPS frame budget)
  - `RenderThreshold`: 16ms
  - `UpdateThreshold`: 5ms
  - `EventThreshold`: 10ms
  - `FrequentRenderThreshold`: 1000 renders
  - `MemoryThreshold`: 10MB
- Implemented `PerformanceMetrics` struct for aggregating metrics for detection
- Implemented `NewBottleneckDetector()` and `NewBottleneckDetectorWithThresholds()` constructors
- Implemented `Check(operation, duration)` method:
  - Returns nil if duration <= threshold
  - Returns BottleneckInfo with severity, impact, description, suggestion if exceeded
  - Tracks violations per operation
- Implemented `Detect(metrics)` method:
  - Analyzes ComponentMetrics for slow renders (AvgRenderTime > threshold)
  - Detects frequent renders (RenderCount > threshold)
  - Detects memory issues (MemoryUsage > threshold)
  - Returns multiple bottlenecks if detected
- Implemented severity calculation based on ratio:
  - < 2x threshold: Low
  - 2-3x threshold: Medium
  - 3-5x threshold: High
  - > 5x threshold: Critical
- Implemented impact calculation: normalized to 0.0-1.0 range (ratio/10, capped at 1.0)
- Implemented context-aware suggestion generation for:
  - Slow operations (render, update, generic)
  - Frequent operations (memoization suggestions)
  - Memory issues (pooling, sync.Pool suggestions)
  - Pattern issues (architecture suggestions)
- Added helper methods: `SetThreshold()`, `GetThreshold()`, `GetViolations()`, `GetAllViolations()`, `GetConfig()`, `Reset()`
- Updated `profiler.go`:
  - Enhanced `BottleneckDetector` struct with `violations` and `config` fields
  - Updated `New()` to use `NewBottleneckDetector()` instead of inline struct
- Thread-safe with `sync.RWMutex` protecting all operations
- 27 table-driven tests covering all functionality including concurrent access (50 goroutines)
- **Coverage: 100%** for bottleneck.go (exceeds >95% requirement)
- **Overall profiler coverage: 96.9%**
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

### Task 4.2: Threshold Monitor ✅ COMPLETED
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
- [x] Threshold checking
- [x] Violation tracking
- [x] Configurable thresholds
- [x] Multiple operations
- [x] Alert generation

**Estimated Effort**: 2 hours

**Implementation Notes (Completed 2024-11-29)**:
- Created `threshold.go` with full `ThresholdMonitor` implementation
- Implemented `ThresholdConfig` struct with configurable settings:
  - `DefaultThreshold`: Default threshold for operations (16ms for 60 FPS)
  - `AlertCooldown`: Minimum time between alerts for same operation (1s default)
  - `MaxAlerts`: Maximum alerts to retain in history (100 default)
  - `EnableAlerts`: Toggle for alert generation
- Implemented `Alert` struct with Operation, Duration, Threshold, Severity, Timestamp, Description
- Implemented `AlertHandler` callback type for real-time alert notifications
- Implemented core methods:
  - `NewThresholdMonitor()`, `NewThresholdMonitorWithConfig()` constructors
  - `Check(operation, duration)` - returns BottleneckInfo and generates alerts
  - `SetThreshold()`, `GetThreshold()` - threshold management
  - `GetViolations()`, `GetAllViolations()` - violation tracking
  - `GetAlerts()`, `SetAlertHandler()`, `ClearAlerts()` - alert management
  - `Reset()`, `GetConfig()` - lifecycle and configuration
- Alert cooldown prevents alert storms for same operation
- Alert history limited to MaxAlerts (oldest removed when exceeded)
- Thread-safe with `sync.RWMutex` protecting all operations
- 27 table-driven tests covering all functionality including concurrent access (50 goroutines)
- **Coverage: 100%** for threshold.go (exceeds >95% requirement)
- **Overall profiler coverage: 97.2%**
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

### Task 4.3: Pattern Analyzer ✅ COMPLETED
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
- [x] Pattern detection
- [x] Multiple patterns
- [x] Custom patterns
- [x] Severity assignment
- [x] Suggestion generation

**Estimated Effort**: 4 hours

**Implementation Notes (Completed 2024-11-29)**:
- Created `pattern_analyzer.go` with full `PatternAnalyzer` implementation from design spec
- Implemented `Pattern` struct with Name, Detect function, Severity, Description, Suggestion
- Implemented `NewPatternAnalyzer()` with 5 default patterns:
  - `frequent_rerender`: RenderCount > 1000 AND AvgRenderTime < 1ms (SeverityMedium)
  - `slow_render`: AvgRenderTime > 10ms (SeverityHigh)
  - `memory_hog`: MemoryUsage > 5MB (SeverityHigh)
  - `render_spike`: MaxRenderTime > 100ms (SeverityMedium)
  - `inefficient_render`: RenderCount > 500 AND AvgRenderTime > 5ms (SeverityCritical)
- Implemented `NewPatternAnalyzerWithPatterns()` for custom pattern sets
- Implemented `Analyze()` method that checks all patterns against ComponentMetrics
- Implemented `AnalyzeAll()` for batch analysis of multiple components
- Added pattern management methods: `AddPattern()`, `RemovePattern()`, `GetPattern()`, `GetPatterns()`, `PatternCount()`
- Added `Reset()` to restore default patterns and `ClearPatterns()` to remove all
- Implemented `calculatePatternImpact()` to convert Severity to 0.0-1.0 impact score
- Thread-safe with `sync.RWMutex` protecting all operations
- 27 table-driven tests covering all functionality including concurrent access (50 goroutines)
- **Coverage: 97.3%** (exceeds >95% requirement)
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

### Task 4.4: Recommendation Engine ✅ COMPLETED
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
- [x] Rule evaluation
- [x] Priority sorting
- [x] Multiple recommendations
- [x] Custom rules
- [x] Actionable suggestions

**Estimated Effort**: 3 hours

**Implementation Notes (Completed 2024-11-29)**:
- Created `recommendations.go` with full `RecommendationEngine` implementation from design spec
- Implemented `RecommendationRule` struct with Name, Condition, Priority, Category, Title, Description, Action, Impact
- Implemented `NewRecommendationEngine()` with 5 default rules:
  - `suggest_memoization`: RenderCount > 100 AND AvgRenderTime > 5ms (PriorityHigh)
  - `reduce_memory_usage`: MemoryUsage > 5MB (PriorityHigh)
  - `optimize_slow_renders`: AvgRenderTime > 16ms frame budget (PriorityCritical)
  - `batch_state_updates`: Bottlenecks > 5 (PriorityMedium)
  - `review_architecture`: Pattern bottlenecks >= 3 (PriorityLow)
- Implemented `NewRecommendationEngineWithRules()` for custom rule sets
- Implemented `Generate()` method that:
  - Evaluates all rules against Report
  - Skips rules with nil Condition
  - Sorts recommendations by Priority descending (Critical > High > Medium > Low)
  - Returns empty slice for nil report
- Implemented rule management methods: `AddRule()`, `RemoveRule()`, `GetRule()`, `GetRules()`, `RuleCount()`
- Added `Reset()` to restore default rules and `ClearRules()` to remove all
- Thread-safe with `sync.RWMutex` protecting all operations
- 27 table-driven tests covering all functionality including concurrent access (50 goroutines)
- **Coverage: 100%** for recommendations.go (exceeds >95% requirement)
- **Overall profiler coverage: 97.5%**
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

## Phase 5: Reporting & Visualization (5 tasks, 15 hours)

### Task 5.1: Report Generator Core ✅ COMPLETED
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
- [x] Report generation
- [x] All sections included
- [x] Data aggregation correct
- [x] HTML generation
- [x] Template rendering

**Estimated Effort**: 3 hours

**Implementation Notes (Completed 2024-11-29)**:
- Created `report.go` with full `ReportGenerator` implementation using `html/template`
- Implemented `ProfileData` struct to aggregate all profiler data for report generation:
  - `ComponentTracker`, `Collector`, `CPUProfiler`, `MemoryProfiler`, `RenderProfiler`
  - `BottleneckDetector`, `RecommendationEngine`
  - Direct data fields: `Bottlenecks`, `Recommendations`, `CPUProfile`, `MemProfile`
  - `StartTime`, `EndTime` for duration calculation
- Implemented `NewReportGenerator()` with default HTML template
- Implemented `NewReportGeneratorWithTemplate()` for custom templates
- Implemented `Generate(data *ProfileData) *Report` method:
  - Handles nil ProfileData gracefully
  - Extracts components from ComponentTracker
  - Calculates duration from StartTime/EndTime
  - Aggregates bottlenecks, recommendations, CPU/memory profiles
- Implemented `GenerateHTML(report *Report) (string, error)` method:
  - Uses html/template for secure HTML generation (XSS protection)
  - Renders all sections: Summary, Components, Bottlenecks, CPU Profile, Memory Profile, Recommendations
  - Handles nil report gracefully
- Implemented `SaveHTML(report *Report, filename string) error` method
- Added template helper functions:
  - `formatDuration()`: Human-readable duration formatting
  - `formatBytesUint()`: Human-readable byte formatting (uint64)
  - `formatPercent()`: Percentage formatting
  - `severityClass()`: CSS class for severity levels
  - `priorityClass()`: CSS class for priority levels
  - `priorityString()`: Human-readable priority names
- Created comprehensive default HTML template with:
  - Modern CSS styling with CSS variables
  - Responsive grid layout
  - Color-coded severity and priority indicators
  - All report sections with proper formatting
- Thread-safe with `sync.RWMutex` protecting template access
- 27 table-driven tests covering all functionality including:
  - Report generation from various ProfileData configurations
  - HTML generation with all sections
  - XSS protection verification
  - File save operations
  - Custom template support
  - Thread-safe concurrent access (50 goroutines)
  - Full integration workflow test
- **Coverage: 97.1%** (exceeds >95% requirement)
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

### Task 5.2: Data Aggregator ✅ COMPLETED
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
- [x] Data aggregation
- [x] Summary calculation
- [x] Statistics correct
- [x] All metrics included
- [x] Performance acceptable

**Estimated Effort**: 3 hours

**Implementation Notes (Completed 2024-11-29)**:
- Created `aggregator.go` with full `DataAggregator` implementation
- Implemented `AggregatedData` struct containing:
  - `Timings`: map of operation name to `AggregatedTiming` (Count, Total, Min, Max, Mean, P50, P95, P99)
  - `Counters`: map of counter name to `AggregatedCounter` (Count, Value)
  - `Allocations`: map of location to `AggregatedAllocation` (Count, TotalSize, AvgSize)
  - `Components`: slice of `ComponentMetrics` from ComponentTracker
  - `FrameCount`, `AverageFPS`, `DroppedFramePercent` from RenderProfiler
  - `MemorySnapshot`, `GoroutineCount`, `Timestamp` for runtime info
- Implemented `NewDataAggregator()` constructor
- Implemented `Aggregate(collector)` method:
  - Extracts timings from TimingTracker via `GetOperationNames()` and `GetStats()`
  - Extracts counters from CounterTracker via `GetAllCounters()`
  - Extracts allocations from MemoryTracker via `GetAllAllocations()`
- Implemented `AggregateRenderData(profiler)` for render performance data
- Implemented `AggregateComponentData(tracker)` for component metrics
- Implemented `AggregateAll(collector, componentTracker, renderProfiler)` for full aggregation
- Implemented `CalculateSummary(data)` method:
  - Calculates TotalOperations from timing counts
  - Gets AverageFPS from counters or render data
  - Captures MemoryUsage and GoroutineCount from runtime
- Implemented helper methods on `AggregatedData`:
  - `TotalOperations()`: Sum of all timing counts
  - `TotalAllocatedMemory()`: Sum of all allocation sizes
  - `GetTiming(name)`, `GetCounter(name)`, `GetAllocation(location)`: Accessors
- Implemented `TakeMemorySnapshot()` returning `MemProfileData` with GC pauses
- Implemented `GetGoroutineCount()` using `runtime.NumGoroutine()`
- Thread-safe with `sync.RWMutex` protecting all operations
- 27 table-driven tests covering all functionality including:
  - Nil/empty collector handling
  - Timing, counter, allocation aggregation
  - Summary calculation with various data configurations
  - Thread-safe concurrent access (50 goroutines)
  - Performance test (< 100ms for 1000 metrics)
  - GC pause extraction
  - All accessor methods with nil safety
- **Coverage: 97.1%** (exceeds >95% requirement)
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

### Task 5.3: Flame Graph Generator ✅ COMPLETED
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

func (fgg *FlameGraphGenerator) Generate(profile *CPUProfileData) *CallNode
func (fgg *FlameGraphGenerator) GenerateSVG(profile *CPUProfileData) string
```

**Tests**:
- [x] Flame graph generation
- [x] SVG format valid
- [x] Nested calls shown
- [x] Percentages correct
- [x] Visual output correct

**Estimated Effort**: 4 hours

**Implementation Notes (Completed 2024-11-29)**:
- Created `flamegraph.go` with full `FlameGraphGenerator` implementation
- Implemented `FlameGraphGenerator` struct with configurable width/height dimensions
- Implemented `CallNode` struct for hierarchical call tree representation
- Implemented `NewFlameGraphGenerator()` and `NewFlameGraphGeneratorWithDimensions()` constructors
- Implemented `Generate(profile *CPUProfileData) *CallNode` method:
  - Builds hierarchical call tree from CPUProfileData
  - Identifies root functions (not called by others)
  - Recursively builds children from call graph
  - Handles circular call detection with visited map
  - Sorts children by sample count descending
- Implemented `GenerateSVG(profile *CPUProfileData) string` method:
  - Generates valid SVG with proper XML namespace
  - Renders rectangles for each function with flame-like colors
  - Adds text labels with truncation for narrow frames
  - Includes hover tooltips with function name, percentage, and sample count
  - Escapes special XML characters for security
  - Handles empty/nil profiles gracefully with "No profile data" message
- Added helper methods:
  - `buildCallTree()`: Recursive tree builder with cycle detection
  - `renderNode()`: Recursive SVG renderer
  - `getFlameColor()`: Returns flame-like colors (red → orange → yellow)
  - `truncateLabel()`: Truncates labels to fit frame width
  - `escapeXML()`: Escapes special XML characters
- Added `CallNode` methods: `TotalSamples()`, `AddChild()`
- Added dimension methods: `GetWidth()`, `GetHeight()`, `SetDimensions()`, `Reset()`
- Default dimensions: 1200x600 pixels
- Frame height: 18px with 1px padding
- Thread-safe with `sync.RWMutex` protecting all state operations
- 27 table-driven tests covering all functionality including:
  - Constructor tests with default and custom dimensions
  - Generate tests with nil, empty, single, and nested profiles
  - GenerateSVG tests for valid SVG structure, labels, colors, escaping
  - Thread-safe concurrent access (50 goroutines)
  - Helper function tests for color, truncation, XML escaping
  - Full integration workflow test
- **Coverage: 96.8%** (exceeds >95% requirement)
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

### Task 5.4: Timeline Visualizer ✅ COMPLETED
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
- [x] Timeline generation
- [x] Event ordering
- [x] Time scaling
- [x] Visual clarity
- [x] HTML output

**Estimated Effort**: 3 hours

**Implementation Notes (Completed 2024-11-29)**:
- Created `timeline.go` with full `TimelineGenerator` implementation
- Implemented `TimelineGenerator` struct with configurable width/height dimensions
- Implemented `TimedEvent` struct with Name, Type, StartTime, Duration, ComponentID, Metadata
- Implemented `TimelineData` struct for processed timeline data with time range and type counts
- Implemented `EventType` enum: Render, Update, Lifecycle, Event, Command, Custom
- Implemented `NewTimelineGenerator()` and `NewTimelineGeneratorWithDimensions()` constructors
- Implemented `Generate(events []*TimedEvent) *TimelineData` method:
  - Filters nil events
  - Sorts events by start time
  - Calculates time range (StartTime, EndTime, TotalDuration)
  - Counts events by type
  - Returns nil for empty/nil input
- Implemented `GenerateHTML(events []*TimedEvent) string` method:
  - Generates complete HTML page with embedded SVG timeline
  - Modern CSS styling with responsive layout
  - Statistics header (Total Events, Duration, Type counts)
  - Color-coded legend for event types
  - SVG timeline with time axis and markers
  - Event bars proportional to duration
  - Tooltips with event details (name, type, duration, start time)
  - XSS protection via HTML escaping
  - Handles empty events with "No Events" message
- Added helper methods:
  - `GetWidth()`, `GetHeight()`, `SetDimensions()`, `Reset()` for dimension management
  - `getEventColor()` returns color by event type (Green=Render, Blue=Update, Purple=Lifecycle, Orange=Event, Red=Command, Grey=Custom)
  - `formatTimelineDuration()` for human-readable duration (ns, μs, ms, s, m)
  - `truncateTimelineLabel()` for label truncation with ellipsis
  - `escapeHTML()` for XSS protection
- Added convenience functions:
  - `AddEvent()` creates TimedEvent with basic fields
  - `AddEventWithComponent()` creates TimedEvent with component ID
- Added `TimedEvent` methods:
  - `GetEndTime()` returns StartTime + Duration
  - `SetMetadata()` and `GetMetadata()` for metadata management
- Thread-safe with `sync.RWMutex` protecting all state operations
- Default dimensions: 1200x400 pixels
- 27 table-driven tests covering all functionality including:
  - Constructor tests with default and custom dimensions
  - Generate tests with nil, empty, single, multiple events
  - Event ordering verification
  - Time range calculation
  - HTML generation with all sections
  - XSS protection verification
  - Thread-safe concurrent access (50 goroutines)
  - Duration formatting tests
  - Label truncation tests
  - HTML escaping tests
  - Benchmark tests for performance
- **Coverage: 97.0%** (exceeds >95% requirement)
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

### Task 5.5: Export System ✅ COMPLETED
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
- [x] HTML export
- [x] JSON export
- [x] CSV export
- [x] File creation
- [x] Format validation

**Estimated Effort**: 2 hours

**Implementation Notes (Completed 2024-11-29)**:
- Created `export.go` with full `Exporter` implementation
- Implemented `ExportFormat` type with constants: `FormatHTML`, `FormatJSON`, `FormatCSV`
- Implemented `NewExporter()` constructor using `ReportGenerator` for HTML export
- Implemented `ExportHTML()` using existing `ReportGenerator.SaveHTML()` with XSS protection via html/template
- Implemented `ExportJSON()` with pretty-printed output using `json.MarshalIndent()`
- Created JSON-serializable types for all report structures:
  - `jsonReport`, `jsonSummary`, `jsonComponent`, `jsonBottleneck`
  - `jsonCPUProfile`, `jsonHotFunction`, `jsonMemProfile`, `jsonRecommendation`
- Implemented `ExportCSV()` for component metrics export with headers:
  - `component_id`, `component_name`, `render_count`, `avg_render_time_ns`, `max_render_time_ns`, `memory_usage`
- Implemented `ExportAll()` convenience method for exporting all three formats
- Implemented `ExportToString()` for in-memory export without file I/O
- Implemented `GetSupportedFormats()` returning list of supported formats
- Added `reportToJSON()` helper for converting Report to JSON-serializable representation
- Added `priorityToString()` helper for Priority enum conversion
- Proper nil handling throughout (nil reports, nil components, nil fields)
- Thread-safe with `sync.RWMutex` protecting ReportGenerator access
- 27 table-driven tests covering all functionality:
  - HTML export with XSS protection verification
  - JSON export with valid JSON validation and all fields
  - CSV export with proper escaping of special characters
  - File creation and error handling for invalid paths
  - Thread-safe concurrent access (50 goroutines)
  - All priority levels tested
  - Nil field handling
- **Coverage: 96.7%** (exceeds >95% requirement)
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

## Phase 6: Integration & Tools (4 tasks, 12 hours)

### Task 6.1: Benchmark Integration ✅ COMPLETED
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
- [x] Benchmark integration
- [x] Measurement works
- [x] Regression detection
- [x] Baseline comparison
- [x] CI/CD compatible

**Estimated Effort**: 3 hours

**Implementation Notes (Completed 2024-11-30)**:
- Created `benchmark.go` with full `BenchmarkProfiler` implementation
- Implemented `Baseline` struct with JSON serialization for CI/CD persistence:
  - Name, NsPerOp, AllocBytes, AllocsPerOp, Iterations
  - Timestamp, GoVersion, GOOS, GOARCH for environment tracking
  - Metadata map for custom key-value pairs
- Implemented `BenchmarkStats` struct with comprehensive statistics:
  - Iterations, NsPerOp, AllocBytes, AllocsPerOp
  - Min, Max, Mean, P50, P95, P99 percentiles
- Implemented `RegressionInfo` struct for detailed regression reporting:
  - HasRegression, TimeRegression, MemoryRegression, AllocRegression percentages
  - Details string, Baseline and Current stats references
- Implemented `BenchmarkProfiler` struct with:
  - `NewBenchmarkProfiler(b *testing.B)` constructor
  - `Measure(fn func())` for timing function execution
  - `StartMeasurement() func()` for manual timing control
  - `GetStats() *BenchmarkStats` with percentile calculation
  - `GetMeasurements() []time.Duration` for raw data access
  - `Reset()` to clear all measurements
- Implemented baseline management:
  - `SetBaseline()`, `GetBaseline()` for in-memory baseline
  - `NewBaseline(name string) *Baseline` from current stats
  - `SaveBaseline(filename string) error` for JSON persistence
  - `LoadBaseline(filename string) (*Baseline, error)` for loading
- Implemented regression detection:
  - `HasRegression(baseline, threshold) bool` for quick check
  - `AssertNoRegression(baseline, threshold) error` for CI/CD assertions
  - `GetRegressionInfo(baseline) *RegressionInfo` for detailed analysis
  - Threshold-based detection (0.0 to 1.0 = 0% to 100% allowed)
- Implemented reporting:
  - `ReportMetrics()` reports p50, p95, p99, min, max to testing.B
  - `String()` for human-readable summary
- Added helper methods: `MeasureCount()`, `SetAllocStats()`, `GetName()`, `SetName()`
- Added `Clone()` and `String()` methods to Baseline, RegressionInfo, BenchmarkStats
- Thread-safe with `sync.RWMutex` protecting all operations
- 27+ table-driven tests covering all functionality:
  - Constructor tests with nil/valid testing.B
  - Measure tests with fast/slow/nil functions
  - Statistics calculation with percentiles
  - Baseline save/load with JSON serialization
  - Regression detection with various thresholds
  - Thread-safe concurrent access (50 goroutines)
  - Full integration workflow test
  - Benchmark tests for overhead measurement
- **Coverage: 96.4%** (exceeds >95% requirement)
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

### Task 6.2: Component Instrumentation ✅ COMPLETED
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
- [x] Component instrumentation
- [x] Render tracking
- [x] Update tracking
- [x] Minimal overhead
- [x] No breaking changes

**Estimated Effort**: 3 hours

**Implementation Notes (Completed 2024-11-30)**:
- Created `instrumentation.go` with full `Instrumentor` implementation
- Implemented `Instrumentor` struct with:
  - `profiler *Profiler` - parent profiler instance
  - `componentTracker *ComponentTracker` - tracks per-component metrics
  - `collector *MetricCollector` - handles timing collection
  - `enabled atomic.Bool` - fast enable/disable check
- Implemented `NewInstrumentor(profiler *Profiler)` constructor
- Implemented `Enable()`, `Disable()`, `IsEnabled()` for lifecycle management
- Implemented `InstrumentRender(component Component) func()`:
  - Returns stop function that records render duration
  - Uses `ComponentTracker.RecordRender()` for metrics
  - Fast path when disabled (~3ns overhead)
- Implemented `InstrumentUpdate(component Component) func()`:
  - Returns stop function that records update duration
  - Uses `MetricCollector.GetTimings().Record()` for metrics
  - Fast path when disabled (~3ns overhead)
- Implemented `InstrumentComponent(component Component) Component`:
  - Returns `instrumentedComponent` wrapper
  - Automatically times View() and Update() calls
  - Delegates all other methods to original component
- Implemented `instrumentedComponent` wrapper:
  - Implements full `Component` interface
  - Automatically records render timing in View()
  - Automatically records update timing in Update()
  - Delegates Init(), Name(), ID(), Props(), Emit(), On(), KeyBindings(), HelpText(), IsInitialized()
- Added `KeyBinding` type to avoid circular dependency with bubbly package
- Added helper methods: `GetComponentTracker()`, `GetCollector()`, `Reset()`
- Thread-safe with `sync.RWMutex` and `atomic.Bool` for all operations
- 27 table-driven tests covering all functionality:
  - Constructor tests with nil/valid profiler
  - Enable/disable lifecycle tests
  - Nil component handling tests
  - Render and update metric recording tests
  - Disabled overhead tests (< 1μs verified)
  - No breaking changes tests
  - Thread-safe concurrent access (50 goroutines)
  - InstrumentedComponent wrapper tests
  - Multiple renders test
  - Overhead percentage test (< 3% verified at 0.36%)
- Benchmark tests:
  - `BenchmarkInstrumentor_Disabled`: ~3ns/op, 0 allocs
  - `BenchmarkInstrumentor_Enabled`: ~3.3μs/op, 1 alloc
  - `BenchmarkInstrumentedComponent_View`: ~3.2μs/op, 0 allocs
- **Coverage: 96.3%** (exceeds >95% requirement)
- All tests pass with race detector
- Zero lint warnings, proper formatting

---

### Task 6.3: Dev Tools Integration ✅ COMPLETED
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
- [x] Integration works
- [x] Metrics sent
- [x] Panel displays
- [x] Real-time updates
- [x] No performance impact

**Estimated Effort**: 3 hours

**Implementation Notes (Completed 2024-11-30)**:
- Created `devtools.go` with full `DevToolsIntegration` implementation
- Implemented `DevToolsIntegration` struct with:
  - `profiler *Profiler` - parent profiler instance
  - `metricsBuffer []*MetricsSnapshot` - stores recent metrics snapshots
  - `panels map[string]bool` - registered panel names
  - `callbacks []MetricsUpdateCallback` - update notification callbacks
  - `updateInterval time.Duration` - configurable update interval
  - `enabled atomic.Bool` - fast enable/disable check
- Implemented `NewDevToolsIntegration(profiler *Profiler)` constructor
- Implemented `Enable()`, `Disable()`, `IsEnabled()` for lifecycle management
- Implemented `SendMetrics()` method:
  - Fast path when disabled (~2.2ns/op, 0 allocs)
  - Collects timing stats, render stats, memory stats
  - Buffers snapshots (last 1000)
  - Notifies registered callbacks
- Implemented `RegisterPanel(name)`, `UnregisterPanel(name)`, `PanelExists(name)`, `GetPanelNames()`, `GetPanelCount()`
- Implemented `GetMetricsSnapshot()` returning most recent snapshot
- Implemented `ClearMetrics()`, `GetMetricsCount()` for buffer management
- Implemented `SetUpdateInterval()`, `GetUpdateInterval()` for configurable update rate
- Implemented `OnMetricsUpdate(callback)` for real-time update notifications
- Implemented `GetProfiler()`, `Reset()` helper methods
- Created `MetricsSnapshot` struct with:
  - `Timings map[string]*TimingSnapshot` - timing statistics
  - `Components []*ComponentMetrics` - per-component metrics
  - `FPS float64`, `DroppedFrames float64` - render performance
  - `MemoryUsage uint64`, `GoroutineCount int` - system metrics
  - `Timestamp time.Time` - snapshot timestamp
- Created `TimingSnapshot` struct with Count, Total, Min, Max, Mean, P50, P95, P99
- Thread-safe with `sync.RWMutex` and `atomic.Bool` for all operations
- 27 table-driven tests covering all functionality:
  - Constructor tests with nil/valid profiler
  - Enable/disable lifecycle tests
  - SendMetrics with disabled/enabled/no-metrics scenarios
  - RegisterPanel with valid/empty/custom names
  - GetMetricsSnapshot and ClearMetrics tests
  - SetUpdateInterval with valid/zero/negative intervals
  - Thread-safe concurrent access (50 goroutines × 100 iterations)
  - Performance overhead tests (< 10ms for 10000 disabled calls)
  - Real-time updates with callback verification
  - No breaking changes to profiler functionality
  - Multiple callbacks test
  - Panel management tests (exists, unregister, get names)
  - Nil callback handling
  - Default update interval verification
- Benchmark tests:
  - `BenchmarkDevToolsIntegration_SendMetrics_Disabled`: ~2.2ns/op, 0 allocs
  - `BenchmarkDevToolsIntegration_SendMetrics_Enabled`: ~29μs/op, 5 allocs
  - `BenchmarkDevToolsIntegration_GetMetricsSnapshot`: ~13.5ns/op, 0 allocs
- **Coverage: 96.3%** (exceeds >95% requirement)
- All tests pass with race detector
- Zero lint warnings, proper formatting

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
