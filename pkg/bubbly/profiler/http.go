// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// HTTPHandler provides HTTP handlers for pprof access.
//
// It wraps Go's runtime/pprof functionality and provides production-safe
// HTTP endpoints for CPU, memory, goroutine, and other profiles.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	h := NewHTTPHandler(profiler)
//	h.Enable()
//	mux := http.NewServeMux()
//	h.RegisterHandlers(mux, "/debug/pprof")
//
//	// Or use standalone functions:
//	RegisterHandlers(mux, profiler)
type HTTPHandler struct {
	// profiler is the parent profiler instance
	profiler *Profiler

	// enabled controls whether profiling endpoints are active
	enabled atomic.Bool

	// maxCPUProfileDuration limits CPU profile duration for safety
	maxCPUProfileDuration time.Duration

	// maxTraceDuration limits trace duration for safety
	maxTraceDuration time.Duration

	// mu protects configuration changes
	mu sync.RWMutex
}

// Common errors for HTTP handlers
var (
	// ErrInvalidDuration is returned for invalid duration parameters
	ErrInvalidDuration = errors.New("duration must be positive")

	// ErrProfilingDisabled is returned when profiling is disabled
	ErrProfilingDisabled = errors.New("profiling disabled")
)

// Default configuration values
const (
	// DefaultMaxCPUProfileDuration is the default maximum CPU profile duration
	DefaultMaxCPUProfileDuration = 30 * time.Second

	// DefaultMaxTraceDuration is the default maximum trace duration
	DefaultMaxTraceDuration = 30 * time.Second

	// DefaultCPUProfileDuration is the default CPU profile duration
	DefaultCPUProfileDuration = 30 * time.Second

	// DefaultTraceDuration is the default trace duration
	DefaultTraceDuration = 1 * time.Second
)

// NewHTTPHandler creates a new HTTP handler for pprof access.
//
// If profiler is nil, a default profiler is created.
// The handler is disabled by default for production safety.
//
// Example:
//
//	h := NewHTTPHandler(profiler)
//	h.Enable()
//	mux := http.NewServeMux()
//	h.RegisterHandlers(mux, "/debug/pprof")
func NewHTTPHandler(profiler *Profiler) *HTTPHandler {
	if profiler == nil {
		profiler = New()
	}

	return &HTTPHandler{
		profiler:              profiler,
		maxCPUProfileDuration: DefaultMaxCPUProfileDuration,
		maxTraceDuration:      DefaultMaxTraceDuration,
	}
}

// Enable activates the HTTP profiling endpoints.
//
// By default, endpoints are disabled for production safety.
// Call Enable() to allow access to profiling data.
func (h *HTTPHandler) Enable() {
	h.enabled.Store(true)
}

// Disable deactivates the HTTP profiling endpoints.
//
// When disabled, all endpoints return 503 Service Unavailable.
func (h *HTTPHandler) Disable() {
	h.enabled.Store(false)
}

// IsEnabled returns whether profiling endpoints are active.
func (h *HTTPHandler) IsEnabled() bool {
	return h.enabled.Load()
}

// Reset resets the handler to its initial state.
func (h *HTTPHandler) Reset() {
	h.enabled.Store(false)
	h.mu.Lock()
	h.maxCPUProfileDuration = DefaultMaxCPUProfileDuration
	h.maxTraceDuration = DefaultMaxTraceDuration
	h.mu.Unlock()
}

// GetProfiler returns the underlying profiler instance.
func (h *HTTPHandler) GetProfiler() *Profiler {
	return h.profiler
}

// SetMaxCPUProfileDuration sets the maximum allowed CPU profile duration.
//
// This limits how long a CPU profile can run to prevent resource exhaustion.
func (h *HTTPHandler) SetMaxCPUProfileDuration(d time.Duration) error {
	if d <= 0 {
		return ErrInvalidDuration
	}
	h.mu.Lock()
	h.maxCPUProfileDuration = d
	h.mu.Unlock()
	return nil
}

// GetMaxCPUProfileDuration returns the maximum allowed CPU profile duration.
func (h *HTTPHandler) GetMaxCPUProfileDuration() time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.maxCPUProfileDuration
}

// SetMaxTraceDuration sets the maximum allowed trace duration.
func (h *HTTPHandler) SetMaxTraceDuration(d time.Duration) error {
	if d <= 0 {
		return ErrInvalidDuration
	}
	h.mu.Lock()
	h.maxTraceDuration = d
	h.mu.Unlock()
	return nil
}

// GetMaxTraceDuration returns the maximum allowed trace duration.
func (h *HTTPHandler) GetMaxTraceDuration() time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.maxTraceDuration
}

// RegisterHandlers registers all pprof handlers on the given ServeMux.
//
// The prefix parameter specifies the URL prefix for all handlers.
// If prefix is empty, "/debug/pprof" is used.
//
// Example:
//
//	mux := http.NewServeMux()
//	h.RegisterHandlers(mux, "/debug/profiler")
func (h *HTTPHandler) RegisterHandlers(mux *http.ServeMux, prefix string) error {
	if prefix == "" {
		prefix = "/debug/pprof"
	}

	// Ensure prefix doesn't end with slash for consistent handling
	if len(prefix) > 0 && prefix[len(prefix)-1] == '/' {
		prefix = prefix[:len(prefix)-1]
	}

	mux.HandleFunc(prefix+"/", h.ServeIndex)
	mux.HandleFunc(prefix+"/profile", h.ServeCPUProfile)
	mux.HandleFunc(prefix+"/heap", h.ServeHeapProfile)
	mux.HandleFunc(prefix+"/goroutine", h.ServeGoroutineProfile)
	mux.HandleFunc(prefix+"/block", h.ServeBlockProfile)
	mux.HandleFunc(prefix+"/mutex", h.ServeMutexProfile)
	mux.HandleFunc(prefix+"/threadcreate", h.ServeThreadcreateProfile)
	mux.HandleFunc(prefix+"/allocs", h.ServeAllocsProfile)
	mux.HandleFunc(prefix+"/symbol", h.ServeSymbol)
	mux.HandleFunc(prefix+"/trace", h.ServeTrace)

	return nil
}

// checkEnabled returns an error response if profiling is disabled.
func (h *HTTPHandler) checkEnabled(w http.ResponseWriter) bool {
	if !h.IsEnabled() {
		http.Error(w, "profiling disabled", http.StatusServiceUnavailable)
		return false
	}
	return true
}

// checkMethod returns an error response if the method is not allowed.
func checkMethod(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

// servePprofLookup serves a named pprof profile.
// This helper eliminates code duplication across block, mutex, threadcreate, and allocs profiles.
func (h *HTTPHandler) servePprofLookup(w http.ResponseWriter, r *http.Request, profileName, filename string) {
	if !checkMethod(w, r) {
		return
	}
	if !h.checkEnabled(w) {
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)

	p := pprof.Lookup(profileName)
	if p == nil {
		http.Error(w, profileName+" profile not found", http.StatusInternalServerError)
		return
	}

	if err := p.WriteTo(w, 0); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ServeCPUProfile serves CPU profile data.
//
// Query parameters:
//   - seconds: duration of the profile (default: 30)
//
// The profile is written in pprof format (gzip compressed).
func (h *HTTPHandler) ServeCPUProfile(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r) {
		return
	}
	if !h.checkEnabled(w) {
		return
	}

	// Parse seconds parameter
	seconds := DefaultCPUProfileDuration
	if s := r.URL.Query().Get("seconds"); s != "" {
		sec, err := strconv.Atoi(s)
		if err != nil {
			http.Error(w, "invalid seconds parameter", http.StatusBadRequest)
			return
		}
		if sec < 0 {
			http.Error(w, "seconds must be non-negative", http.StatusBadRequest)
			return
		}
		if sec > 0 {
			seconds = time.Duration(sec) * time.Second
		}
	}

	// Limit duration
	h.mu.RLock()
	maxDuration := h.maxCPUProfileDuration
	h.mu.RUnlock()
	if seconds > maxDuration {
		seconds = maxDuration
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=profile.pb.gz")

	if err := pprof.StartCPUProfile(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	time.Sleep(seconds)
	pprof.StopCPUProfile()
}

// ServeHeapProfile serves heap profile data.
//
// Query parameters:
//   - gc: if "1", run GC before taking the profile
//
// The profile is written in pprof format (gzip compressed).
func (h *HTTPHandler) ServeHeapProfile(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r) {
		return
	}
	if !h.checkEnabled(w) {
		return
	}

	// Check if GC should be run
	if r.URL.Query().Get("gc") == "1" {
		runtime.GC()
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=heap.pb.gz")

	if err := pprof.WriteHeapProfile(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ServeGoroutineProfile serves goroutine profile data.
//
// Query parameters:
//   - debug: 0 for binary, 1 for text, 2 for text with full stack traces
//
// With debug=0, the profile is in pprof format.
// With debug=1 or debug=2, the profile is in human-readable text format.
func (h *HTTPHandler) ServeGoroutineProfile(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r) {
		return
	}
	if !h.checkEnabled(w) {
		return
	}

	debug := 0
	if d := r.URL.Query().Get("debug"); d != "" {
		var err error
		debug, err = strconv.Atoi(d)
		if err != nil {
			debug = 0
		}
	}

	if debug > 0 {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=goroutine.pb.gz")
	}

	p := pprof.Lookup("goroutine")
	if p == nil {
		http.Error(w, "goroutine profile not found", http.StatusInternalServerError)
		return
	}

	if err := p.WriteTo(w, debug); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ServeBlockProfile serves block profile data.
//
// The profile shows where goroutines block waiting on synchronization primitives.
func (h *HTTPHandler) ServeBlockProfile(w http.ResponseWriter, r *http.Request) {
	h.servePprofLookup(w, r, "block", "block.pb.gz")
}

// ServeMutexProfile serves mutex profile data.
//
// The profile shows mutex contention.
func (h *HTTPHandler) ServeMutexProfile(w http.ResponseWriter, r *http.Request) {
	h.servePprofLookup(w, r, "mutex", "mutex.pb.gz")
}

// ServeThreadcreateProfile serves threadcreate profile data.
//
// The profile shows stack traces that led to the creation of new OS threads.
func (h *HTTPHandler) ServeThreadcreateProfile(w http.ResponseWriter, r *http.Request) {
	h.servePprofLookup(w, r, "threadcreate", "threadcreate.pb.gz")
}

// ServeAllocsProfile serves allocation profile data.
//
// The profile shows a sampling of all past memory allocations.
func (h *HTTPHandler) ServeAllocsProfile(w http.ResponseWriter, r *http.Request) {
	h.servePprofLookup(w, r, "allocs", "allocs.pb.gz")
}

// indexTemplate is the HTML template for the index page
var indexTemplate = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html>
<head>
<title>BubblyUI Profiler</title>
<style>
body { font-family: sans-serif; margin: 2em; }
h1 { color: #333; }
table { border-collapse: collapse; margin-top: 1em; }
th, td { border: 1px solid #ddd; padding: 8px 12px; text-align: left; }
th { background-color: #f5f5f5; }
a { color: #0066cc; text-decoration: none; }
a:hover { text-decoration: underline; }
.description { color: #666; font-size: 0.9em; }
</style>
</head>
<body>
<h1>BubblyUI Profiler</h1>
<p>Available profiles:</p>
<table>
<tr><th>Profile</th><th>Description</th></tr>
<tr><td><a href="profile?seconds=30">profile</a></td><td class="description">CPU profile (30 second default)</td></tr>
<tr><td><a href="heap">heap</a></td><td class="description">Heap memory allocations</td></tr>
<tr><td><a href="goroutine?debug=1">goroutine</a></td><td class="description">Stack traces of all goroutines</td></tr>
<tr><td><a href="block">block</a></td><td class="description">Stack traces of blocked goroutines</td></tr>
<tr><td><a href="mutex">mutex</a></td><td class="description">Mutex contention</td></tr>
<tr><td><a href="threadcreate">threadcreate</a></td><td class="description">OS thread creation</td></tr>
<tr><td><a href="allocs">allocs</a></td><td class="description">Memory allocations sampling</td></tr>
<tr><td><a href="trace?seconds=1">trace</a></td><td class="description">Execution trace (1 second default)</td></tr>
</table>
<p style="margin-top: 2em; color: #666;">
Use <code>go tool pprof</code> to analyze profiles.<br>
Example: <code>go tool pprof http://localhost:8080/debug/pprof/heap</code>
</p>
</body>
</html>
`))

// ServeIndex serves the index page listing all available profiles.
func (h *HTTPHandler) ServeIndex(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r) {
		return
	}
	if !h.checkEnabled(w) {
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := indexTemplate.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ServeSymbol serves symbol lookup for pprof.
//
// This endpoint is used by pprof tools to look up symbol names.
func (h *HTTPHandler) ServeSymbol(w http.ResponseWriter, r *http.Request) {
	if !h.checkEnabled(w) {
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Symbol lookup is handled by reading addresses from the request body
	// and returning function names. For simplicity, we just return OK.
	// The actual symbol resolution happens in the pprof tool.
	fmt.Fprintf(w, "num_symbols: 0\n")
}

// ServeTrace serves execution trace data.
//
// Query parameters:
//   - seconds: duration of the trace (default: 1)
//
// The trace can be viewed with: go tool trace <file>
func (h *HTTPHandler) ServeTrace(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r) {
		return
	}
	if !h.checkEnabled(w) {
		return
	}

	// Parse seconds parameter
	seconds := DefaultTraceDuration
	if s := r.URL.Query().Get("seconds"); s != "" {
		sec, err := strconv.Atoi(s)
		if err != nil {
			http.Error(w, "invalid seconds parameter", http.StatusBadRequest)
			return
		}
		if sec > 0 {
			seconds = time.Duration(sec) * time.Second
		}
	}

	// Limit duration
	h.mu.RLock()
	maxDuration := h.maxTraceDuration
	h.mu.RUnlock()
	if seconds > maxDuration {
		seconds = maxDuration
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=trace.out")

	if err := trace.Start(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	time.Sleep(seconds)
	trace.Stop()
}

// ============================================================================
// Standalone functions for convenience
// ============================================================================

// defaultHandler is a package-level handler for standalone functions
var (
	defaultHandler     *HTTPHandler
	defaultHandlerOnce sync.Once
)

// getDefaultHandler returns the default HTTP handler.
func getDefaultHandler() *HTTPHandler {
	defaultHandlerOnce.Do(func() {
		defaultHandler = NewHTTPHandler(nil)
	})
	return defaultHandler
}

// RegisterHandlers registers pprof handlers on the given ServeMux using the provided profiler.
//
// This is a convenience function that creates an HTTPHandler and registers all endpoints.
// The handler is disabled by default for production safety.
//
// Example:
//
//	mux := http.NewServeMux()
//	RegisterHandlers(mux, profiler)
//	// Enable profiling:
//	// The handler needs to be enabled separately
func RegisterHandlers(mux *http.ServeMux, profiler *Profiler) {
	h := NewHTTPHandler(profiler)
	_ = h.RegisterHandlers(mux, "/debug/pprof")
}

// ServeCPUProfile is a standalone handler for CPU profiles.
//
// Uses the default handler which is disabled by default.
func ServeCPUProfile(w http.ResponseWriter, r *http.Request) {
	getDefaultHandler().ServeCPUProfile(w, r)
}

// ServeHeapProfile is a standalone handler for heap profiles.
//
// Uses the default handler which is disabled by default.
func ServeHeapProfile(w http.ResponseWriter, r *http.Request) {
	getDefaultHandler().ServeHeapProfile(w, r)
}
