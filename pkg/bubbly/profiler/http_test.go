// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewHTTPHandler tests the constructor.
func TestNewHTTPHandler(t *testing.T) {
	tests := []struct {
		name     string
		profiler *Profiler
		wantNil  bool
	}{
		{
			name:     "with nil profiler creates default",
			profiler: nil,
			wantNil:  false,
		},
		{
			name:     "with valid profiler",
			profiler: New(),
			wantNil:  false,
		},
		{
			name:     "with enabled profiler",
			profiler: New(WithEnabled(true)),
			wantNil:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHTTPHandler(tt.profiler)
			if tt.wantNil {
				assert.Nil(t, h)
			} else {
				assert.NotNil(t, h)
				assert.NotNil(t, h.profiler)
			}
		})
	}
}

// TestHTTPHandler_RegisterHandlers tests handler registration.
func TestHTTPHandler_RegisterHandlers(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		wantErr bool
	}{
		{
			name:    "default prefix",
			prefix:  "",
			wantErr: false,
		},
		{
			name:    "custom prefix",
			prefix:  "/debug/profiler",
			wantErr: false,
		},
		{
			name:    "trailing slash prefix",
			prefix:  "/debug/pprof/",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHTTPHandler(nil)
			mux := http.NewServeMux()

			err := h.RegisterHandlers(mux, tt.prefix)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestHTTPHandler_ServeCPUProfile tests CPU profile serving.
func TestHTTPHandler_ServeCPUProfile(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		querySeconds   string
		enabled        bool
		wantStatusCode int
		wantGzip       bool
	}{
		{
			name:           "disabled returns 503",
			method:         http.MethodGet,
			querySeconds:   "",
			enabled:        false,
			wantStatusCode: http.StatusServiceUnavailable,
			wantGzip:       false,
		},
		{
			name:           "POST request not allowed",
			method:         http.MethodPost,
			querySeconds:   "",
			enabled:        true,
			wantStatusCode: http.StatusMethodNotAllowed,
			wantGzip:       false,
		},
		{
			name:           "invalid seconds parameter",
			method:         http.MethodGet,
			querySeconds:   "invalid",
			enabled:        true,
			wantStatusCode: http.StatusBadRequest,
			wantGzip:       false,
		},
		{
			name:           "negative seconds parameter",
			method:         http.MethodGet,
			querySeconds:   "-1",
			enabled:        true,
			wantStatusCode: http.StatusBadRequest,
			wantGzip:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHTTPHandler(nil)
			if tt.enabled {
				h.Enable()
			}

			url := "/debug/pprof/profile"
			if tt.querySeconds != "" {
				url += "?seconds=" + tt.querySeconds
			}

			req := httptest.NewRequest(tt.method, url, nil)
			rec := httptest.NewRecorder()

			h.ServeCPUProfile(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)
			if tt.wantGzip && rec.Code == http.StatusOK {
				// CPU profiles are gzip compressed
				assert.Equal(t, "application/octet-stream", rec.Header().Get("Content-Type"))
				// Verify it's valid gzip
				_, err := gzip.NewReader(rec.Body)
				assert.NoError(t, err)
			}
		})
	}
}

// TestHTTPHandler_ServeHeapProfile tests heap profile serving.
func TestHTTPHandler_ServeHeapProfile(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		queryGC        string
		enabled        bool
		wantStatusCode int
		wantGzip       bool
	}{
		{
			name:           "disabled returns 503",
			method:         http.MethodGet,
			queryGC:        "",
			enabled:        false,
			wantStatusCode: http.StatusServiceUnavailable,
			wantGzip:       false,
		},
		{
			name:           "GET request",
			method:         http.MethodGet,
			queryGC:        "",
			enabled:        true,
			wantStatusCode: http.StatusOK,
			wantGzip:       true,
		},
		{
			name:           "GET request with gc=1",
			method:         http.MethodGet,
			queryGC:        "1",
			enabled:        true,
			wantStatusCode: http.StatusOK,
			wantGzip:       true,
		},
		{
			name:           "POST request not allowed",
			method:         http.MethodPost,
			queryGC:        "",
			enabled:        true,
			wantStatusCode: http.StatusMethodNotAllowed,
			wantGzip:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHTTPHandler(nil)
			if tt.enabled {
				h.Enable()
			}

			url := "/debug/pprof/heap"
			if tt.queryGC != "" {
				url += "?gc=" + tt.queryGC
			}

			req := httptest.NewRequest(tt.method, url, nil)
			rec := httptest.NewRecorder()

			h.ServeHeapProfile(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)
			if tt.wantGzip && rec.Code == http.StatusOK {
				assert.Equal(t, "application/octet-stream", rec.Header().Get("Content-Type"))
				// Verify it's valid gzip
				_, err := gzip.NewReader(rec.Body)
				assert.NoError(t, err)
			}
		})
	}
}

// TestHTTPHandler_ServeGoroutineProfile tests goroutine profile serving.
func TestHTTPHandler_ServeGoroutineProfile(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		queryDebug     string
		enabled        bool
		wantStatusCode int
		wantTextOutput bool
	}{
		{
			name:           "disabled returns 503",
			method:         http.MethodGet,
			queryDebug:     "",
			enabled:        false,
			wantStatusCode: http.StatusServiceUnavailable,
			wantTextOutput: false,
		},
		{
			name:           "GET request default",
			method:         http.MethodGet,
			queryDebug:     "",
			enabled:        true,
			wantStatusCode: http.StatusOK,
			wantTextOutput: false,
		},
		{
			name:           "GET request with debug=1",
			method:         http.MethodGet,
			queryDebug:     "1",
			enabled:        true,
			wantStatusCode: http.StatusOK,
			wantTextOutput: true,
		},
		{
			name:           "GET request with debug=2",
			method:         http.MethodGet,
			queryDebug:     "2",
			enabled:        true,
			wantStatusCode: http.StatusOK,
			wantTextOutput: true,
		},
		{
			name:           "POST request not allowed",
			method:         http.MethodPost,
			queryDebug:     "",
			enabled:        true,
			wantStatusCode: http.StatusMethodNotAllowed,
			wantTextOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHTTPHandler(nil)
			if tt.enabled {
				h.Enable()
			}

			url := "/debug/pprof/goroutine"
			if tt.queryDebug != "" {
				url += "?debug=" + tt.queryDebug
			}

			req := httptest.NewRequest(tt.method, url, nil)
			rec := httptest.NewRecorder()

			h.ServeGoroutineProfile(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)
			if rec.Code == http.StatusOK {
				if tt.wantTextOutput {
					assert.Equal(t, "text/plain; charset=utf-8", rec.Header().Get("Content-Type"))
					// Text output should contain "goroutine"
					assert.Contains(t, rec.Body.String(), "goroutine")
				} else {
					assert.Equal(t, "application/octet-stream", rec.Header().Get("Content-Type"))
				}
			}
		})
	}
}

// TestHTTPHandler_ServeBlockProfile tests block profile serving.
func TestHTTPHandler_ServeBlockProfile(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		enabled        bool
		wantStatusCode int
	}{
		{
			name:           "disabled returns 503",
			method:         http.MethodGet,
			enabled:        false,
			wantStatusCode: http.StatusServiceUnavailable,
		},
		{
			name:           "GET request",
			method:         http.MethodGet,
			enabled:        true,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "POST request not allowed",
			method:         http.MethodPost,
			enabled:        true,
			wantStatusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHTTPHandler(nil)
			if tt.enabled {
				h.Enable()
			}

			req := httptest.NewRequest(tt.method, "/debug/pprof/block", nil)
			rec := httptest.NewRecorder()

			h.ServeBlockProfile(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)
		})
	}
}

// TestHTTPHandler_ServeMutexProfile tests mutex profile serving.
func TestHTTPHandler_ServeMutexProfile(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		enabled        bool
		wantStatusCode int
	}{
		{
			name:           "disabled returns 503",
			method:         http.MethodGet,
			enabled:        false,
			wantStatusCode: http.StatusServiceUnavailable,
		},
		{
			name:           "GET request",
			method:         http.MethodGet,
			enabled:        true,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "POST request not allowed",
			method:         http.MethodPost,
			enabled:        true,
			wantStatusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHTTPHandler(nil)
			if tt.enabled {
				h.Enable()
			}

			req := httptest.NewRequest(tt.method, "/debug/pprof/mutex", nil)
			rec := httptest.NewRecorder()

			h.ServeMutexProfile(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)
		})
	}
}

// TestHTTPHandler_ServeThreadcreateProfile tests threadcreate profile serving.
func TestHTTPHandler_ServeThreadcreateProfile(t *testing.T) {
	h := NewHTTPHandler(nil)
	h.Enable()

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/threadcreate", nil)
	rec := httptest.NewRecorder()

	h.ServeThreadcreateProfile(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/octet-stream", rec.Header().Get("Content-Type"))
}

// TestHTTPHandler_ServeAllocsProfile tests allocs profile serving.
func TestHTTPHandler_ServeAllocsProfile(t *testing.T) {
	h := NewHTTPHandler(nil)
	h.Enable()

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/allocs", nil)
	rec := httptest.NewRecorder()

	h.ServeAllocsProfile(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/octet-stream", rec.Header().Get("Content-Type"))
}

// TestHTTPHandler_ServeIndex tests the index page.
func TestHTTPHandler_ServeIndex(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		enabled        bool
		wantStatusCode int
		wantContains   []string
	}{
		{
			name:           "disabled returns 503",
			method:         http.MethodGet,
			enabled:        false,
			wantStatusCode: http.StatusServiceUnavailable,
			wantContains:   nil,
		},
		{
			name:           "GET request",
			method:         http.MethodGet,
			enabled:        true,
			wantStatusCode: http.StatusOK,
			wantContains:   []string{"profile", "heap", "goroutine", "block"},
		},
		{
			name:           "POST request not allowed",
			method:         http.MethodPost,
			enabled:        true,
			wantStatusCode: http.StatusMethodNotAllowed,
			wantContains:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHTTPHandler(nil)
			if tt.enabled {
				h.Enable()
			}

			req := httptest.NewRequest(tt.method, "/debug/pprof/", nil)
			rec := httptest.NewRecorder()

			h.ServeIndex(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)
			if tt.wantContains != nil {
				body := rec.Body.String()
				for _, s := range tt.wantContains {
					assert.Contains(t, body, s)
				}
			}
		})
	}
}

// TestHTTPHandler_ServeSymbol tests symbol lookup.
func TestHTTPHandler_ServeSymbol(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		enabled        bool
		wantStatusCode int
	}{
		{
			name:           "disabled returns 503",
			method:         http.MethodGet,
			body:           "",
			enabled:        false,
			wantStatusCode: http.StatusServiceUnavailable,
		},
		{
			name:           "GET request",
			method:         http.MethodGet,
			body:           "",
			enabled:        true,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "POST request with addresses",
			method:         http.MethodPost,
			body:           "0x12345\n0x67890",
			enabled:        true,
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHTTPHandler(nil)
			if tt.enabled {
				h.Enable()
			}

			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}

			req := httptest.NewRequest(tt.method, "/debug/pprof/symbol", body)
			rec := httptest.NewRecorder()

			h.ServeSymbol(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)
		})
	}
}

// TestHTTPHandler_ServeTrace tests trace profile serving.
func TestHTTPHandler_ServeTrace(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		querySeconds   string
		enabled        bool
		wantStatusCode int
	}{
		{
			name:           "disabled returns 503",
			method:         http.MethodGet,
			querySeconds:   "",
			enabled:        false,
			wantStatusCode: http.StatusServiceUnavailable,
		},
		{
			name:           "POST request not allowed",
			method:         http.MethodPost,
			querySeconds:   "",
			enabled:        true,
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid seconds parameter",
			method:         http.MethodGet,
			querySeconds:   "invalid",
			enabled:        true,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHTTPHandler(nil)
			if tt.enabled {
				h.Enable()
			}

			url := "/debug/pprof/trace"
			if tt.querySeconds != "" {
				url += "?seconds=" + tt.querySeconds
			}

			req := httptest.NewRequest(tt.method, url, nil)
			rec := httptest.NewRecorder()

			h.ServeTrace(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)
		})
	}
}

// TestHTTPHandler_ThreadSafety tests concurrent access.
func TestHTTPHandler_ThreadSafety(t *testing.T) {
	h := NewHTTPHandler(nil)

	var wg sync.WaitGroup
	const goroutines = 50
	const iterations = 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// Mix of operations
				switch j % 5 {
				case 0:
					req := httptest.NewRequest(http.MethodGet, "/debug/pprof/heap", nil)
					rec := httptest.NewRecorder()
					h.ServeHeapProfile(rec, req)
				case 1:
					req := httptest.NewRequest(http.MethodGet, "/debug/pprof/goroutine?debug=1", nil)
					rec := httptest.NewRecorder()
					h.ServeGoroutineProfile(rec, req)
				case 2:
					req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
					rec := httptest.NewRecorder()
					h.ServeIndex(rec, req)
				case 3:
					req := httptest.NewRequest(http.MethodGet, "/debug/pprof/allocs", nil)
					rec := httptest.NewRecorder()
					h.ServeAllocsProfile(rec, req)
				case 4:
					_ = h.GetProfiler()
				}
			}
		}()
	}

	wg.Wait()
	// If we get here without deadlock or panic, test passes
}

// TestHTTPHandler_ProductionSafe tests production safety features.
func TestHTTPHandler_ProductionSafe(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*HTTPHandler)
		wantEnabled bool
	}{
		{
			name:        "disabled by default",
			setupFunc:   nil,
			wantEnabled: false,
		},
		{
			name: "can be enabled",
			setupFunc: func(h *HTTPHandler) {
				h.Enable()
			},
			wantEnabled: true,
		},
		{
			name: "can be disabled after enabling",
			setupFunc: func(h *HTTPHandler) {
				h.Enable()
				h.Disable()
			},
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHTTPHandler(nil)

			if tt.setupFunc != nil {
				tt.setupFunc(h)
			}

			assert.Equal(t, tt.wantEnabled, h.IsEnabled())
		})
	}
}

// TestHTTPHandler_DisabledReturns503 tests that disabled handler returns 503.
func TestHTTPHandler_DisabledReturns503(t *testing.T) {
	h := NewHTTPHandler(nil)
	// Handler is disabled by default

	endpoints := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{"heap", h.ServeHeapProfile},
		{"goroutine", h.ServeGoroutineProfile},
		{"block", h.ServeBlockProfile},
		{"mutex", h.ServeMutexProfile},
		{"allocs", h.ServeAllocsProfile},
		{"threadcreate", h.ServeThreadcreateProfile},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/debug/pprof/"+ep.name, nil)
			rec := httptest.NewRecorder()

			ep.handler(rec, req)

			assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
			assert.Contains(t, rec.Body.String(), "profiling disabled")
		})
	}
}

// TestHTTPHandler_EnabledReturnsProfile tests that enabled handler returns profile.
func TestHTTPHandler_EnabledReturnsProfile(t *testing.T) {
	h := NewHTTPHandler(nil)
	h.Enable()

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/heap", nil)
	rec := httptest.NewRecorder()

	h.ServeHeapProfile(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestHTTPHandler_GetProfiler tests getting the profiler reference.
func TestHTTPHandler_GetProfiler(t *testing.T) {
	prof := New()
	h := NewHTTPHandler(prof)

	assert.Equal(t, prof, h.GetProfiler())
}

// TestHTTPHandler_Reset tests resetting the handler.
func TestHTTPHandler_Reset(t *testing.T) {
	h := NewHTTPHandler(nil)
	h.Enable()
	assert.True(t, h.IsEnabled())

	h.Reset()

	assert.False(t, h.IsEnabled())
}

// TestHTTPHandler_SetMaxCPUProfileDuration tests setting max CPU profile duration.
func TestHTTPHandler_SetMaxCPUProfileDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		wantErr  bool
	}{
		{
			name:     "valid duration 30s",
			duration: 30 * time.Second,
			wantErr:  false,
		},
		{
			name:     "valid duration 1m",
			duration: time.Minute,
			wantErr:  false,
		},
		{
			name:     "zero duration error",
			duration: 0,
			wantErr:  true,
		},
		{
			name:     "negative duration error",
			duration: -1 * time.Second,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHTTPHandler(nil)

			err := h.SetMaxCPUProfileDuration(tt.duration)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.duration, h.GetMaxCPUProfileDuration())
			}
		})
	}
}

// TestHTTPHandler_SetMaxTraceDuration tests setting max trace duration.
func TestHTTPHandler_SetMaxTraceDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		wantErr  bool
	}{
		{
			name:     "valid duration 30s",
			duration: 30 * time.Second,
			wantErr:  false,
		},
		{
			name:     "zero duration error",
			duration: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHTTPHandler(nil)

			err := h.SetMaxTraceDuration(tt.duration)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.duration, h.GetMaxTraceDuration())
			}
		})
	}
}

// TestRegisterHandlers tests the standalone RegisterHandlers function.
func TestRegisterHandlers(t *testing.T) {
	prof := New()
	mux := http.NewServeMux()

	RegisterHandlers(mux, prof)

	// Verify handlers are registered by making requests
	server := httptest.NewServer(mux)
	defer server.Close()

	// Test index endpoint
	resp, err := http.Get(server.URL + "/debug/pprof/")
	require.NoError(t, err)
	defer resp.Body.Close()
	// Should return 503 because profiling is disabled by default
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

// TestServeCPUProfile tests the standalone ServeCPUProfile function.
func TestServeCPUProfile(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/profile?seconds=1", nil)
	rec := httptest.NewRecorder()

	ServeCPUProfile(rec, req)

	// Should return 503 because default handler is disabled
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

// TestServeHeapProfile tests the standalone ServeHeapProfile function.
func TestServeHeapProfile(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/heap", nil)
	rec := httptest.NewRecorder()

	ServeHeapProfile(rec, req)

	// Should return 503 because default handler is disabled
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

// TestHTTPHandler_CPUProfileDurationLimit tests CPU profile duration limiting.
func TestHTTPHandler_CPUProfileDurationLimit(t *testing.T) {
	h := NewHTTPHandler(nil)
	h.Enable()
	err := h.SetMaxCPUProfileDuration(5 * time.Second)
	require.NoError(t, err)

	// Verify the max duration is set correctly
	assert.Equal(t, 5*time.Second, h.GetMaxCPUProfileDuration())

	// We can't easily test the actual limiting without a long-running test
	// The implementation limits the duration in ServeCPUProfile
}

// TestHTTPHandler_ServeProfiles_ContentType tests content types for all profiles.
func TestHTTPHandler_ServeProfiles_ContentType(t *testing.T) {
	h := NewHTTPHandler(nil)
	h.Enable()

	profiles := []struct {
		name        string
		path        string
		handler     func(http.ResponseWriter, *http.Request)
		contentType string
	}{
		{"heap", "/debug/pprof/heap", h.ServeHeapProfile, "application/octet-stream"},
		{"goroutine", "/debug/pprof/goroutine", h.ServeGoroutineProfile, "application/octet-stream"},
		{"block", "/debug/pprof/block", h.ServeBlockProfile, "application/octet-stream"},
		{"mutex", "/debug/pprof/mutex", h.ServeMutexProfile, "application/octet-stream"},
		{"threadcreate", "/debug/pprof/threadcreate", h.ServeThreadcreateProfile, "application/octet-stream"},
		{"allocs", "/debug/pprof/allocs", h.ServeAllocsProfile, "application/octet-stream"},
	}

	for _, p := range profiles {
		t.Run(p.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, p.path, nil)
			rec := httptest.NewRecorder()

			p.handler(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, p.contentType, rec.Header().Get("Content-Type"))
			// Verify body is not empty
			assert.Greater(t, rec.Body.Len(), 0)
		})
	}
}

// TestHTTPHandler_ServeProfiles_ContentDisposition tests content disposition headers.
func TestHTTPHandler_ServeProfiles_ContentDisposition(t *testing.T) {
	h := NewHTTPHandler(nil)
	h.Enable()

	profiles := []struct {
		name     string
		path     string
		handler  func(http.ResponseWriter, *http.Request)
		filename string
	}{
		{"heap", "/debug/pprof/heap", h.ServeHeapProfile, "heap.pb.gz"},
		{"goroutine", "/debug/pprof/goroutine", h.ServeGoroutineProfile, "goroutine.pb.gz"},
		{"block", "/debug/pprof/block", h.ServeBlockProfile, "block.pb.gz"},
		{"mutex", "/debug/pprof/mutex", h.ServeMutexProfile, "mutex.pb.gz"},
		{"threadcreate", "/debug/pprof/threadcreate", h.ServeThreadcreateProfile, "threadcreate.pb.gz"},
		{"allocs", "/debug/pprof/allocs", h.ServeAllocsProfile, "allocs.pb.gz"},
	}

	for _, p := range profiles {
		t.Run(p.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, p.path, nil)
			rec := httptest.NewRecorder()

			p.handler(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Header().Get("Content-Disposition"), p.filename)
		})
	}
}

// TestHTTPHandler_ServeGoroutine_DebugModes tests all debug modes for goroutine profile.
func TestHTTPHandler_ServeGoroutine_DebugModes(t *testing.T) {
	h := NewHTTPHandler(nil)
	h.Enable()

	tests := []struct {
		debug       string
		contentType string
	}{
		{"0", "application/octet-stream"},
		{"1", "text/plain; charset=utf-8"},
		{"2", "text/plain; charset=utf-8"},
	}

	for _, tt := range tests {
		t.Run("debug="+tt.debug, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/debug/pprof/goroutine?debug="+tt.debug, nil)
			rec := httptest.NewRecorder()

			h.ServeGoroutineProfile(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, tt.contentType, rec.Header().Get("Content-Type"))
		})
	}
}

// TestHTTPHandler_ServeGoroutine_InvalidDebug tests invalid debug parameter.
func TestHTTPHandler_ServeGoroutine_InvalidDebug(t *testing.T) {
	h := NewHTTPHandler(nil)
	h.Enable()

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/goroutine?debug=invalid", nil)
	rec := httptest.NewRecorder()

	h.ServeGoroutineProfile(rec, req)

	// Invalid debug should default to 0 (binary format)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/octet-stream", rec.Header().Get("Content-Type"))
}

// TestHTTPHandler_ServeHeap_GC tests GC parameter for heap profile.
func TestHTTPHandler_ServeHeap_GC(t *testing.T) {
	h := NewHTTPHandler(nil)
	h.Enable()

	tests := []struct {
		gc string
	}{
		{""},
		{"0"},
		{"1"},
	}

	for _, tt := range tests {
		name := "gc=" + tt.gc
		if tt.gc == "" {
			name = "no gc param"
		}
		t.Run(name, func(t *testing.T) {
			url := "/debug/pprof/heap"
			if tt.gc != "" {
				url += "?gc=" + tt.gc
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()

			h.ServeHeapProfile(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

// TestHTTPHandler_ServeIndex_HTMLContent tests index page HTML content.
func TestHTTPHandler_ServeIndex_HTMLContent(t *testing.T) {
	h := NewHTTPHandler(nil)
	h.Enable()

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	rec := httptest.NewRecorder()

	h.ServeIndex(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))

	body := rec.Body.String()
	// Verify HTML structure
	assert.Contains(t, body, "<!DOCTYPE html>")
	assert.Contains(t, body, "<html>")
	assert.Contains(t, body, "</html>")
	assert.Contains(t, body, "BubblyUI Profiler")
	// Verify all profile links are present
	assert.Contains(t, body, "profile")
	assert.Contains(t, body, "heap")
	assert.Contains(t, body, "goroutine")
	assert.Contains(t, body, "block")
	assert.Contains(t, body, "mutex")
	assert.Contains(t, body, "threadcreate")
	assert.Contains(t, body, "allocs")
	assert.Contains(t, body, "trace")
}

// TestHTTPHandler_ServeCPUProfile_QuickTest tests CPU profile with very short duration.
func TestHTTPHandler_ServeCPUProfile_QuickTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CPU profile test in short mode")
	}

	h := NewHTTPHandler(nil)
	h.Enable()
	// Set very short max duration
	_ = h.SetMaxCPUProfileDuration(100 * time.Millisecond)

	// Request with seconds=0 which uses default, but limited by max
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/profile?seconds=0", nil)
	rec := httptest.NewRecorder()

	start := time.Now()
	h.ServeCPUProfile(rec, req)
	elapsed := time.Since(start)

	// Should complete within max duration + some buffer
	assert.Less(t, elapsed, 500*time.Millisecond)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/octet-stream", rec.Header().Get("Content-Type"))
}

// TestHTTPHandler_ServeTrace_QuickTest tests trace with very short duration.
func TestHTTPHandler_ServeTrace_QuickTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping trace test in short mode")
	}

	h := NewHTTPHandler(nil)
	h.Enable()
	// Set very short max duration
	_ = h.SetMaxTraceDuration(100 * time.Millisecond)

	// Request with seconds=0 which uses default (1s), but limited by max
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/trace?seconds=0", nil)
	rec := httptest.NewRecorder()

	start := time.Now()
	h.ServeTrace(rec, req)
	elapsed := time.Since(start)

	// Should complete within max duration + some buffer
	assert.Less(t, elapsed, 500*time.Millisecond)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/octet-stream", rec.Header().Get("Content-Type"))
}

// TestHTTPHandler_ServeCPUProfile_DurationLimiting tests that duration is limited.
func TestHTTPHandler_ServeCPUProfile_DurationLimiting(t *testing.T) {
	h := NewHTTPHandler(nil)
	h.Enable()
	// Set very short max duration
	_ = h.SetMaxCPUProfileDuration(50 * time.Millisecond)

	// Request 10 seconds, should be limited to 50ms
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/profile?seconds=10", nil)
	rec := httptest.NewRecorder()

	start := time.Now()
	h.ServeCPUProfile(rec, req)
	elapsed := time.Since(start)

	// Should complete much faster than 10 seconds
	assert.Less(t, elapsed, 500*time.Millisecond)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestHTTPHandler_ServeTrace_DurationLimiting tests that trace duration is limited.
func TestHTTPHandler_ServeTrace_DurationLimiting(t *testing.T) {
	h := NewHTTPHandler(nil)
	h.Enable()
	// Set very short max duration
	_ = h.SetMaxTraceDuration(50 * time.Millisecond)

	// Request 10 seconds, should be limited to 50ms
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/trace?seconds=10", nil)
	rec := httptest.NewRecorder()

	start := time.Now()
	h.ServeTrace(rec, req)
	elapsed := time.Since(start)

	// Should complete much faster than 10 seconds
	assert.Less(t, elapsed, 500*time.Millisecond)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestHTTPHandler_ServeCPUProfile_ZeroSeconds tests zero seconds parameter.
func TestHTTPHandler_ServeCPUProfile_ZeroSeconds(t *testing.T) {
	h := NewHTTPHandler(nil)
	h.Enable()
	// Set very short max duration
	_ = h.SetMaxCPUProfileDuration(50 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/profile?seconds=0", nil)
	rec := httptest.NewRecorder()

	start := time.Now()
	h.ServeCPUProfile(rec, req)
	elapsed := time.Since(start)

	// Should complete within max duration
	assert.Less(t, elapsed, 500*time.Millisecond)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestHTTPHandler_ServeTrace_ValidSeconds tests valid seconds parameter.
func TestHTTPHandler_ServeTrace_ValidSeconds(t *testing.T) {
	h := NewHTTPHandler(nil)
	h.Enable()
	// Set very short max duration
	_ = h.SetMaxTraceDuration(50 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/trace?seconds=1", nil)
	rec := httptest.NewRecorder()

	start := time.Now()
	h.ServeTrace(rec, req)
	elapsed := time.Since(start)

	// Should complete within max duration (limited from 1s to 50ms)
	assert.Less(t, elapsed, 500*time.Millisecond)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestHTTPHandler_MethodNotAllowed tests method not allowed for all endpoints.
func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	h := NewHTTPHandler(nil)
	h.Enable()

	endpoints := []struct {
		name    string
		path    string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{"heap", "/debug/pprof/heap", h.ServeHeapProfile},
		{"goroutine", "/debug/pprof/goroutine", h.ServeGoroutineProfile},
		{"block", "/debug/pprof/block", h.ServeBlockProfile},
		{"mutex", "/debug/pprof/mutex", h.ServeMutexProfile},
		{"threadcreate", "/debug/pprof/threadcreate", h.ServeThreadcreateProfile},
		{"allocs", "/debug/pprof/allocs", h.ServeAllocsProfile},
		{"index", "/debug/pprof/", h.ServeIndex},
		{"cpu", "/debug/pprof/profile", h.ServeCPUProfile},
		{"trace", "/debug/pprof/trace", h.ServeTrace},
	}

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, ep := range endpoints {
		for _, method := range methods {
			t.Run(ep.name+"_"+method, func(t *testing.T) {
				req := httptest.NewRequest(method, ep.path, nil)
				rec := httptest.NewRecorder()

				ep.handler(rec, req)

				assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
			})
		}
	}
}

// BenchmarkHTTPHandler_ServeHeapProfile benchmarks heap profile serving.
func BenchmarkHTTPHandler_ServeHeapProfile(b *testing.B) {
	h := NewHTTPHandler(nil)
	h.Enable()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/debug/pprof/heap", nil)
		rec := httptest.NewRecorder()
		h.ServeHeapProfile(rec, req)
	}
}

// BenchmarkHTTPHandler_ServeGoroutineProfile benchmarks goroutine profile serving.
func BenchmarkHTTPHandler_ServeGoroutineProfile(b *testing.B) {
	h := NewHTTPHandler(nil)
	h.Enable()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/debug/pprof/goroutine?debug=1", nil)
		rec := httptest.NewRecorder()
		h.ServeGoroutineProfile(rec, req)
	}
}

// BenchmarkHTTPHandler_ServeIndex benchmarks index page serving.
func BenchmarkHTTPHandler_ServeIndex(b *testing.B) {
	h := NewHTTPHandler(nil)
	h.Enable()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
		rec := httptest.NewRecorder()
		h.ServeIndex(rec, req)
	}
}
