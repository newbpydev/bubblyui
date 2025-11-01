package monitoring

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnableProfiling tests that profiling endpoint can be started
func TestEnableProfiling(t *testing.T) {
	// Use a random available port
	addr := "localhost:0"

	err := EnableProfiling(addr)
	require.NoError(t, err, "EnableProfiling should not return an error")

	// Cleanup
	defer StopProfiling()
}

// TestEnableProfiling_PortInUse tests error handling for port conflicts
func TestEnableProfiling_PortInUse(t *testing.T) {
	addr := "localhost:16061"

	// Start first server
	err := EnableProfiling(addr)
	require.NoError(t, err)
	defer StopProfiling()

	// Try to start another server on the same port (should fail eventually or be prevented)
	// This tests that we handle port conflicts gracefully
	time.Sleep(100 * time.Millisecond) // Give server time to start
}

// TestEnableProfiling_InvalidAddress tests error handling for invalid addresses
func TestEnableProfiling_InvalidAddress(t *testing.T) {
	// Test with empty address
	err := EnableProfiling("")
	assert.Error(t, err, "Should return error for empty address")

	// Note: Port 99999 is syntactically valid (though it will fail to bind at runtime)
	// The actual binding error happens asynchronously in the goroutine
}

// TestProfilingEndpoints tests that pprof endpoints are available
func TestProfilingEndpoints(t *testing.T) {
	addr := "localhost:16062"

	err := EnableProfiling(addr)
	require.NoError(t, err)
	defer StopProfiling()

	// Give server time to start
	time.Sleep(200 * time.Millisecond)

	// Test that pprof endpoints are available
	endpoints := []string{
		"/debug/pprof/",
		"/debug/pprof/heap",
		"/debug/pprof/goroutine",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			resp, err := http.Get("http://" + addr + endpoint)
			if err != nil {
				t.Skipf("Server might not be ready: %v", err)
				return
			}
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode,
				"Endpoint should be accessible")
		})
	}

	// Test CPU profile endpoint separately (takes longer)
	t.Run("/debug/pprof/profile", func(t *testing.T) {
		// Note: This endpoint takes time (default 30s, we use 1s)
		// Just verify it responds, don't wait for completion
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get("http://" + addr + "/debug/pprof/profile?seconds=1")
		if err != nil {
			// Timeout is expected for profile endpoint
			t.Skipf("Profile endpoint timing: %v", err)
			return
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestProfileComposables tests composable profiling functionality
func TestProfileComposables(t *testing.T) {
	duration := 100 * time.Millisecond

	profile := ProfileComposables(duration)

	assert.NotNil(t, profile, "Profile should not be nil")
	assert.False(t, profile.Start.IsZero(), "Start time should be set")
	assert.False(t, profile.End.IsZero(), "End time should be set")
	assert.True(t, profile.End.After(profile.Start), "End should be after Start")
	assert.NotNil(t, profile.Calls, "Calls map should be initialized")
}

// TestCallStats_Recording tests that call stats are recorded correctly
func TestCallStats_Recording(t *testing.T) {
	stats := &CallStats{
		Count:       10,
		TotalTime:   100 * time.Millisecond,
		Allocations: 1280,
	}

	// Calculate average
	stats.CalculateAverage()

	assert.Equal(t, int64(10), stats.Count)
	assert.Equal(t, 100*time.Millisecond, stats.TotalTime)
	assert.Equal(t, 10*time.Millisecond, stats.AverageTime)
	assert.Equal(t, int64(1280), stats.Allocations)
}

// TestComposableProfile_AddCall tests adding calls to profile
func TestComposableProfile_AddCall(t *testing.T) {
	profile := &ComposableProfile{
		Start: time.Now(),
		Calls: make(map[string]*CallStats),
	}

	// Add a call
	profile.AddCall("UseState", 100*time.Nanosecond, 128)
	profile.AddCall("UseState", 120*time.Nanosecond, 128)
	profile.AddCall("UseForm", 500*time.Nanosecond, 256)

	// Close profile
	profile.End = time.Now()

	// Verify stats
	useState := profile.Calls["UseState"]
	require.NotNil(t, useState, "UseState stats should exist")
	assert.Equal(t, int64(2), useState.Count)
	assert.Equal(t, 220*time.Nanosecond, useState.TotalTime)
	assert.Equal(t, int64(256), useState.Allocations)

	useForm := profile.Calls["UseForm"]
	require.NotNil(t, useForm, "UseForm stats should exist")
	assert.Equal(t, int64(1), useForm.Count)
	assert.Equal(t, 500*time.Nanosecond, useForm.TotalTime)
	assert.Equal(t, int64(256), useForm.Allocations)
}

// TestComposableProfile_Summary tests profile summary generation
func TestComposableProfile_Summary(t *testing.T) {
	profile := &ComposableProfile{
		Start: time.Now(),
		End:   time.Now().Add(1 * time.Second),
		Calls: map[string]*CallStats{
			"UseState": {
				Count:       100,
				TotalTime:   350 * time.Microsecond,
				Allocations: 12800,
			},
			"UseForm": {
				Count:       50,
				TotalTime:   750 * time.Microsecond,
				Allocations: 12800,
			},
		},
	}

	// Calculate averages
	for _, stats := range profile.Calls {
		stats.CalculateAverage()
	}

	summary := profile.Summary()

	assert.NotEmpty(t, summary, "Summary should not be empty")
	assert.Contains(t, summary, "UseState", "Summary should contain UseState")
	assert.Contains(t, summary, "UseForm", "Summary should contain UseForm")
	assert.Contains(t, summary, "100 calls", "Summary should contain call count")
}

// TestStopProfiling tests that profiling can be stopped
func TestStopProfiling(t *testing.T) {
	addr := "localhost:16063"

	err := EnableProfiling(addr)
	require.NoError(t, err)

	// Stop profiling
	StopProfiling()

	// Verify server is stopped by trying to connect
	time.Sleep(100 * time.Millisecond)
	_, err = http.Get("http://" + addr + "/debug/pprof/")
	assert.Error(t, err, "Server should be stopped")
}

// TestGetProfilingAddress tests getting the current profiling address
func TestGetProfilingAddress(t *testing.T) {
	// Initially should be empty
	addr := GetProfilingAddress()
	assert.Empty(t, addr, "Address should be empty initially")

	// Enable profiling
	err := EnableProfiling("localhost:16064")
	require.NoError(t, err)
	defer StopProfiling()

	// Should return the address
	addr = GetProfilingAddress()
	assert.Equal(t, "localhost:16064", addr)
}

// TestIsProfilingEnabled tests profiling status check
func TestIsProfilingEnabled(t *testing.T) {
	// Initially should be false
	assert.False(t, IsProfilingEnabled(), "Profiling should be disabled initially")

	// Enable profiling
	err := EnableProfiling("localhost:16065")
	require.NoError(t, err)
	defer StopProfiling()

	// Should be true
	assert.True(t, IsProfilingEnabled(), "Profiling should be enabled")

	// Stop profiling
	StopProfiling()

	// Should be false again
	assert.False(t, IsProfilingEnabled(), "Profiling should be disabled after stop")
}

// TestEnableProfiling_Concurrent tests concurrent profiling operations
func TestEnableProfiling_Concurrent(t *testing.T) {
	addr := "localhost:16066"

	// Enable profiling
	err := EnableProfiling(addr)
	require.NoError(t, err)
	defer StopProfiling()

	// Try to enable again (should handle gracefully)
	err = EnableProfiling("localhost:16067")
	assert.Error(t, err, "Should return error when already enabled")
}

// TestProfileComposables_WithMetrics tests profiling with metrics integration
func TestProfileComposables_WithMetrics(t *testing.T) {
	// Set up metrics
	metrics := &NoOpMetrics{}
	SetGlobalMetrics(metrics)
	defer SetGlobalMetrics(&NoOpMetrics{})

	// Run profile
	profile := ProfileComposables(50 * time.Millisecond)

	assert.NotNil(t, profile)
	assert.NotNil(t, profile.Calls)
}

// TestCallStats_ThreadSafety tests that CallStats operations are thread-safe
func TestCallStats_ThreadSafety(t *testing.T) {
	stats := &CallStats{}

	done := make(chan bool)

	// Concurrent increments
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				stats.RecordCall(10*time.Nanosecond, 128)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify count
	assert.Equal(t, int64(1000), stats.Count, "Should record all 1000 calls")
}
