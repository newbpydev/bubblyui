// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReportGenerator(t *testing.T) {
	tests := []struct {
		name string
		want func(*testing.T, *ReportGenerator)
	}{
		{
			name: "creates generator with default template",
			want: func(t *testing.T, rg *ReportGenerator) {
				assert.NotNil(t, rg)
				assert.NotNil(t, rg.GetTemplate())
			},
		},
		{
			name: "template is valid and parseable",
			want: func(t *testing.T, rg *ReportGenerator) {
				tmpl := rg.GetTemplate()
				assert.NotNil(t, tmpl)
				// Template should have "report" defined
				assert.NotNil(t, tmpl.Lookup("report"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rg := NewReportGenerator()
			tt.want(t, rg)
		})
	}
}

func TestReportGenerator_Generate(t *testing.T) {
	tests := []struct {
		name string
		data *ProfileData
		want func(*testing.T, *Report)
	}{
		{
			name: "generates report from nil ProfileData",
			data: nil,
			want: func(t *testing.T, r *Report) {
				assert.NotNil(t, r)
				assert.NotNil(t, r.Summary)
				assert.NotNil(t, r.Components)
				assert.NotNil(t, r.Bottlenecks)
				assert.NotNil(t, r.Recommendations)
				assert.False(t, r.Timestamp.IsZero())
			},
		},
		{
			name: "generates report with empty ProfileData",
			data: &ProfileData{},
			want: func(t *testing.T, r *Report) {
				assert.NotNil(t, r)
				assert.NotNil(t, r.Summary)
				assert.Equal(t, int64(0), r.Summary.TotalOperations)
			},
		},
		{
			name: "generates report with components",
			data: &ProfileData{
				ComponentTracker: func() *ComponentTracker {
					ct := NewComponentTracker()
					ct.RecordRender("comp1", "TestComponent", 10*time.Millisecond)
					ct.RecordRender("comp1", "TestComponent", 15*time.Millisecond)
					return ct
				}(),
			},
			want: func(t *testing.T, r *Report) {
				assert.NotNil(t, r)
				assert.Len(t, r.Components, 1)
				assert.Equal(t, "comp1", r.Components[0].ComponentID)
				assert.Equal(t, "TestComponent", r.Components[0].ComponentName)
				assert.Equal(t, int64(2), r.Components[0].RenderCount)
			},
		},
		{
			name: "generates report with timing data",
			data: &ProfileData{
				StartTime: time.Now().Add(-5 * time.Minute),
				EndTime:   time.Now(),
			},
			want: func(t *testing.T, r *Report) {
				assert.NotNil(t, r)
				assert.NotNil(t, r.Summary)
				// Duration should be approximately 5 minutes
				assert.True(t, r.Summary.Duration >= 4*time.Minute)
				assert.True(t, r.Summary.Duration <= 6*time.Minute)
			},
		},
		{
			name: "generates report with bottlenecks",
			data: &ProfileData{
				Bottlenecks: []*BottleneckInfo{
					{
						Type:        BottleneckTypeSlow,
						Location:    "TestComponent.render",
						Severity:    SeverityHigh,
						Impact:      0.8,
						Description: "Slow render detected",
						Suggestion:  "Optimize render function",
					},
				},
			},
			want: func(t *testing.T, r *Report) {
				assert.NotNil(t, r)
				assert.Len(t, r.Bottlenecks, 1)
				assert.Equal(t, BottleneckTypeSlow, r.Bottlenecks[0].Type)
				assert.Equal(t, SeverityHigh, r.Bottlenecks[0].Severity)
			},
		},
		{
			name: "generates report with recommendations",
			data: &ProfileData{
				Recommendations: []*Recommendation{
					{
						Title:       "Optimize Slow Renders",
						Description: "Some components exceed frame budget",
						Action:      "Profile render functions",
						Priority:    PriorityCritical,
						Category:    CategoryRendering,
						Impact:      ImpactHigh,
					},
				},
			},
			want: func(t *testing.T, r *Report) {
				assert.NotNil(t, r)
				assert.Len(t, r.Recommendations, 1)
				assert.Equal(t, "Optimize Slow Renders", r.Recommendations[0].Title)
				assert.Equal(t, PriorityCritical, r.Recommendations[0].Priority)
			},
		},
		{
			name: "generates report with CPU profile data",
			data: &ProfileData{
				CPUProfile: &CPUProfileData{
					HotFunctions: []*HotFunction{
						{Name: "main.render", Samples: 100, Percent: 50.0},
						{Name: "main.update", Samples: 50, Percent: 25.0},
					},
					TotalSamples: 200,
				},
			},
			want: func(t *testing.T, r *Report) {
				assert.NotNil(t, r)
				assert.NotNil(t, r.CPUProfile)
				assert.Len(t, r.CPUProfile.HotFunctions, 2)
				assert.Equal(t, int64(200), r.CPUProfile.TotalSamples)
			},
		},
		{
			name: "generates report with memory profile data",
			data: &ProfileData{
				MemProfile: &MemProfileData{
					HeapAlloc:   1024 * 1024 * 10, // 10MB
					HeapObjects: 5000,
					GCPauses:    []time.Duration{1 * time.Millisecond, 2 * time.Millisecond},
				},
			},
			want: func(t *testing.T, r *Report) {
				assert.NotNil(t, r)
				assert.NotNil(t, r.MemProfile)
				assert.Equal(t, uint64(10*1024*1024), r.MemProfile.HeapAlloc)
				assert.Equal(t, uint64(5000), r.MemProfile.HeapObjects)
				assert.Len(t, r.MemProfile.GCPauses, 2)
			},
		},
		{
			name: "timestamp is set correctly",
			data: &ProfileData{},
			want: func(t *testing.T, r *Report) {
				assert.NotNil(t, r)
				assert.False(t, r.Timestamp.IsZero())
				// Timestamp should be recent (within last second)
				assert.True(t, time.Since(r.Timestamp) < time.Second)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rg := NewReportGenerator()
			report := rg.Generate(tt.data)
			tt.want(t, report)
		})
	}
}

func TestReportGenerator_GenerateHTML(t *testing.T) {
	tests := []struct {
		name    string
		report  *Report
		wantErr bool
		want    func(*testing.T, string)
	}{
		{
			name:    "generates valid HTML from nil report",
			report:  nil,
			wantErr: false,
			want: func(t *testing.T, html string) {
				assert.Contains(t, html, "<!DOCTYPE html>")
				assert.Contains(t, html, "<html")
				assert.Contains(t, html, "</html>")
			},
		},
		{
			name: "generates HTML with summary section",
			report: &Report{
				Summary: &Summary{
					Duration:        5 * time.Minute,
					TotalOperations: 1000,
					AverageFPS:      58.5,
					MemoryUsage:     10 * 1024 * 1024,
					GoroutineCount:  50,
				},
				Components:      []*ComponentMetrics{},
				Bottlenecks:     []*BottleneckInfo{},
				Recommendations: []*Recommendation{},
				Timestamp:       time.Now(),
			},
			wantErr: false,
			want: func(t *testing.T, html string) {
				assert.Contains(t, html, "Summary")
				assert.Contains(t, html, "58.5") // FPS
			},
		},
		{
			name: "generates HTML with components section",
			report: &Report{
				Summary: &Summary{},
				Components: []*ComponentMetrics{
					{
						ComponentID:     "comp1",
						ComponentName:   "TestComponent",
						RenderCount:     100,
						TotalRenderTime: 500 * time.Millisecond,
						AvgRenderTime:   5 * time.Millisecond,
						MaxRenderTime:   20 * time.Millisecond,
					},
				},
				Bottlenecks:     []*BottleneckInfo{},
				Recommendations: []*Recommendation{},
				Timestamp:       time.Now(),
			},
			wantErr: false,
			want: func(t *testing.T, html string) {
				assert.Contains(t, html, "TestComponent")
				assert.Contains(t, html, "100") // RenderCount
			},
		},
		{
			name: "generates HTML with bottlenecks section",
			report: &Report{
				Summary:    &Summary{},
				Components: []*ComponentMetrics{},
				Bottlenecks: []*BottleneckInfo{
					{
						Type:        BottleneckTypeSlow,
						Location:    "TestComponent.render",
						Severity:    SeverityCritical,
						Description: "Critical performance issue",
						Suggestion:  "Optimize immediately",
					},
				},
				Recommendations: []*Recommendation{},
				Timestamp:       time.Now(),
			},
			wantErr: false,
			want: func(t *testing.T, html string) {
				assert.Contains(t, html, "Bottleneck")
				assert.Contains(t, html, "Critical performance issue")
			},
		},
		{
			name: "generates HTML with recommendations section",
			report: &Report{
				Summary:     &Summary{},
				Components:  []*ComponentMetrics{},
				Bottlenecks: []*BottleneckInfo{},
				Recommendations: []*Recommendation{
					{
						Title:       "Implement Memoization",
						Description: "Components render frequently",
						Action:      "Add memoization",
						Priority:    PriorityHigh,
					},
				},
				Timestamp: time.Now(),
			},
			wantErr: false,
			want: func(t *testing.T, html string) {
				assert.Contains(t, html, "Recommendation")
				assert.Contains(t, html, "Implement Memoization")
			},
		},
		{
			name: "HTML is properly escaped for security",
			report: &Report{
				Summary: &Summary{},
				Components: []*ComponentMetrics{
					{
						ComponentName: "<script>alert('xss')</script>",
					},
				},
				Bottlenecks:     []*BottleneckInfo{},
				Recommendations: []*Recommendation{},
				Timestamp:       time.Now(),
			},
			wantErr: false,
			want: func(t *testing.T, html string) {
				// Script tags should be escaped
				assert.NotContains(t, html, "<script>alert")
				assert.Contains(t, html, "&lt;script&gt;")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rg := NewReportGenerator()
			html, err := rg.GenerateHTML(tt.report)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.want(t, html)
		})
	}
}

func TestReportGenerator_SaveHTML(t *testing.T) {
	tests := []struct {
		name     string
		report   *Report
		filename string
		setup    func(t *testing.T) string
		wantErr  bool
		want     func(*testing.T, string)
	}{
		{
			name: "creates file with HTML content",
			report: &Report{
				Summary:         &Summary{AverageFPS: 60.0},
				Components:      []*ComponentMetrics{},
				Bottlenecks:     []*BottleneckInfo{},
				Recommendations: []*Recommendation{},
				Timestamp:       time.Now(),
			},
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				return filepath.Join(dir, "report.html")
			},
			wantErr: false,
			want: func(t *testing.T, filename string) {
				content, err := os.ReadFile(filename)
				require.NoError(t, err)
				assert.Contains(t, string(content), "<!DOCTYPE html>")
				assert.Contains(t, string(content), "60")
			},
		},
		{
			name: "handles invalid filename",
			report: &Report{
				Summary:         &Summary{},
				Components:      []*ComponentMetrics{},
				Bottlenecks:     []*BottleneckInfo{},
				Recommendations: []*Recommendation{},
				Timestamp:       time.Now(),
			},
			setup: func(t *testing.T) string {
				return "/nonexistent/directory/that/does/not/exist/report.html"
			},
			wantErr: true,
			want:    func(t *testing.T, filename string) {},
		},
		{
			name: "overwrites existing file",
			report: &Report{
				Summary:         &Summary{AverageFPS: 99.0},
				Components:      []*ComponentMetrics{},
				Bottlenecks:     []*BottleneckInfo{},
				Recommendations: []*Recommendation{},
				Timestamp:       time.Now(),
			},
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				filename := filepath.Join(dir, "existing.html")
				err := os.WriteFile(filename, []byte("old content"), 0644)
				require.NoError(t, err)
				return filename
			},
			wantErr: false,
			want: func(t *testing.T, filename string) {
				content, err := os.ReadFile(filename)
				require.NoError(t, err)
				assert.NotContains(t, string(content), "old content")
				assert.Contains(t, string(content), "99")
			},
		},
		{
			name:   "saves nil report as empty report",
			report: nil,
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				return filepath.Join(dir, "nil-report.html")
			},
			wantErr: false,
			want: func(t *testing.T, filename string) {
				content, err := os.ReadFile(filename)
				require.NoError(t, err)
				assert.Contains(t, string(content), "<!DOCTYPE html>")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rg := NewReportGenerator()
			filename := tt.setup(t)
			err := rg.SaveHTML(tt.report, filename)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.want(t, filename)
		})
	}
}

func TestReportGenerator_SetTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantErr  bool
	}{
		{
			name:     "sets custom template",
			template: `{{define "report"}}Custom: {{.Summary.AverageFPS}}{{end}}`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rg := NewReportGenerator()

			// Parse and set custom template
			tmpl, err := rg.GetTemplate().Clone()
			require.NoError(t, err)
			tmpl, err = tmpl.Parse(tt.template)
			require.NoError(t, err)

			rg.SetTemplate(tmpl)

			// Verify custom template is used
			report := &Report{
				Summary:         &Summary{AverageFPS: 42.0},
				Components:      []*ComponentMetrics{},
				Bottlenecks:     []*BottleneckInfo{},
				Recommendations: []*Recommendation{},
				Timestamp:       time.Now(),
			}
			html, err := rg.GenerateHTML(report)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Contains(t, html, "Custom: 42")
		})
	}
}

func TestReportGenerator_ThreadSafety(t *testing.T) {
	rg := NewReportGenerator()
	data := &ProfileData{
		ComponentTracker: func() *ComponentTracker {
			ct := NewComponentTracker()
			ct.RecordRender("comp1", "TestComponent", 10*time.Millisecond)
			return ct
		}(),
		StartTime: time.Now().Add(-time.Minute),
		EndTime:   time.Now(),
	}

	var wg sync.WaitGroup
	const goroutines = 50

	// Test concurrent Generate calls
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			report := rg.Generate(data)
			assert.NotNil(t, report)
		}()
	}

	// Test concurrent GenerateHTML calls
	report := rg.Generate(data)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			html, err := rg.GenerateHTML(report)
			assert.NoError(t, err)
			assert.NotEmpty(t, html)
		}()
	}

	wg.Wait()
}

func TestReportGenerator_Integration(t *testing.T) {
	// Full workflow: ProfileData → Generate → SaveHTML → Read file
	t.Run("full workflow", func(t *testing.T) {
		// Create ProfileData with all components
		ct := NewComponentTracker()
		ct.RecordRender("comp1", "Header", 5*time.Millisecond)
		ct.RecordRender("comp2", "Footer", 3*time.Millisecond)
		ct.RecordRender("comp1", "Header", 7*time.Millisecond)

		data := &ProfileData{
			ComponentTracker: ct,
			StartTime:        time.Now().Add(-2 * time.Minute),
			EndTime:          time.Now(),
			Bottlenecks: []*BottleneckInfo{
				{
					Type:        BottleneckTypeSlow,
					Location:    "Header.render",
					Severity:    SeverityMedium,
					Description: "Moderate slowdown",
				},
			},
			Recommendations: []*Recommendation{
				{
					Title:    "Optimize Header",
					Priority: PriorityMedium,
				},
			},
			CPUProfile: &CPUProfileData{
				HotFunctions: []*HotFunction{
					{Name: "Header.render", Samples: 50, Percent: 30.0},
				},
				TotalSamples: 100,
			},
			MemProfile: &MemProfileData{
				HeapAlloc:   5 * 1024 * 1024,
				HeapObjects: 1000,
			},
		}

		// Generate report
		rg := NewReportGenerator()
		report := rg.Generate(data)

		// Verify report contents
		assert.NotNil(t, report)
		assert.Len(t, report.Components, 2)
		assert.Len(t, report.Bottlenecks, 1)
		assert.Len(t, report.Recommendations, 1)
		assert.NotNil(t, report.CPUProfile)
		assert.NotNil(t, report.MemProfile)

		// Save to file
		dir := t.TempDir()
		filename := filepath.Join(dir, "integration-report.html")
		err := rg.SaveHTML(report, filename)
		require.NoError(t, err)

		// Read and verify file
		content, err := os.ReadFile(filename)
		require.NoError(t, err)

		htmlStr := string(content)
		assert.Contains(t, htmlStr, "<!DOCTYPE html>")
		assert.Contains(t, htmlStr, "Header")
		assert.Contains(t, htmlStr, "Footer")
		assert.Contains(t, htmlStr, "Moderate slowdown")
		assert.Contains(t, htmlStr, "Optimize Header")
	})
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "formats nanoseconds",
			duration: 500 * time.Nanosecond,
			want:     "500ns",
		},
		{
			name:     "formats microseconds",
			duration: 500 * time.Microsecond,
			want:     "500µs",
		},
		{
			name:     "formats milliseconds",
			duration: 500 * time.Millisecond,
			want:     "500ms",
		},
		{
			name:     "formats seconds",
			duration: 5 * time.Second,
			want:     "5s",
		},
		{
			name:     "formats minutes",
			duration: 5 * time.Minute,
			want:     "5m0s",
		},
		{
			name:     "formats zero",
			duration: 0,
			want:     "0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatBytesUint(t *testing.T) {
	tests := []struct {
		name  string
		bytes uint64
		want  string
	}{
		{
			name:  "formats bytes",
			bytes: 500,
			want:  "500 B",
		},
		{
			name:  "formats kilobytes",
			bytes: 1024,
			want:  "1.00 KB",
		},
		{
			name:  "formats megabytes",
			bytes: 1024 * 1024,
			want:  "1.00 MB",
		},
		{
			name:  "formats gigabytes",
			bytes: 1024 * 1024 * 1024,
			want:  "1.00 GB",
		},
		{
			name:  "formats zero",
			bytes: 0,
			want:  "0 B",
		},
		{
			name:  "formats fractional KB",
			bytes: 1536, // 1.5 KB
			want:  "1.50 KB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBytesUint(tt.bytes)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatPercent(t *testing.T) {
	tests := []struct {
		name    string
		percent float64
		want    string
	}{
		{
			name:    "formats whole percent",
			percent: 50.0,
			want:    "50.0%",
		},
		{
			name:    "formats decimal percent",
			percent: 33.33,
			want:    "33.3%",
		},
		{
			name:    "formats zero",
			percent: 0.0,
			want:    "0.0%",
		},
		{
			name:    "formats 100%",
			percent: 100.0,
			want:    "100.0%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPercent(tt.percent)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSeverityClass(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		want     string
	}{
		{
			name:     "critical severity",
			severity: SeverityCritical,
			want:     "severity-critical",
		},
		{
			name:     "high severity",
			severity: SeverityHigh,
			want:     "severity-high",
		},
		{
			name:     "medium severity",
			severity: SeverityMedium,
			want:     "severity-medium",
		},
		{
			name:     "low severity",
			severity: SeverityLow,
			want:     "severity-low",
		},
		{
			name:     "unknown severity",
			severity: Severity("unknown"),
			want:     "severity-unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := severityClass(tt.severity)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPriorityClass(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		want     string
	}{
		{
			name:     "critical priority",
			priority: PriorityCritical,
			want:     "priority-critical",
		},
		{
			name:     "high priority",
			priority: PriorityHigh,
			want:     "priority-high",
		},
		{
			name:     "medium priority",
			priority: PriorityMedium,
			want:     "priority-medium",
		},
		{
			name:     "low priority",
			priority: PriorityLow,
			want:     "priority-low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := priorityClass(tt.priority)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPriorityString(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		want     string
	}{
		{
			name:     "critical priority",
			priority: PriorityCritical,
			want:     "Critical",
		},
		{
			name:     "high priority",
			priority: PriorityHigh,
			want:     "High",
		},
		{
			name:     "medium priority",
			priority: PriorityMedium,
			want:     "Medium",
		},
		{
			name:     "low priority",
			priority: PriorityLow,
			want:     "Low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := priorityString(tt.priority)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReportGenerator_GenerateHTML_AllSections(t *testing.T) {
	// Comprehensive test that all sections are included in HTML
	report := &Report{
		Summary: &Summary{
			Duration:        10 * time.Minute,
			TotalOperations: 5000,
			AverageFPS:      59.5,
			MemoryUsage:     50 * 1024 * 1024,
			GoroutineCount:  100,
		},
		Components: []*ComponentMetrics{
			{
				ComponentID:     "comp1",
				ComponentName:   "MainView",
				RenderCount:     500,
				TotalRenderTime: 2500 * time.Millisecond,
				AvgRenderTime:   5 * time.Millisecond,
				MaxRenderTime:   50 * time.Millisecond,
				MinRenderTime:   1 * time.Millisecond,
				MemoryUsage:     10 * 1024 * 1024,
			},
		},
		Bottlenecks: []*BottleneckInfo{
			{
				Type:        BottleneckTypeSlow,
				Location:    "MainView.render",
				Severity:    SeverityHigh,
				Impact:      0.75,
				Description: "Render exceeds frame budget",
				Suggestion:  "Consider memoization",
			},
		},
		CPUProfile: &CPUProfileData{
			HotFunctions: []*HotFunction{
				{Name: "MainView.render", Samples: 200, Percent: 40.0},
			},
			TotalSamples: 500,
		},
		MemProfile: &MemProfileData{
			HeapAlloc:   50 * 1024 * 1024,
			HeapObjects: 10000,
			GCPauses:    []time.Duration{5 * time.Millisecond},
		},
		Recommendations: []*Recommendation{
			{
				Title:       "Implement Memoization",
				Description: "MainView renders frequently",
				Action:      "Add memoization to prevent re-renders",
				Priority:    PriorityHigh,
				Category:    CategoryOptimization,
				Impact:      ImpactHigh,
			},
		},
		Timestamp: time.Now(),
	}

	rg := NewReportGenerator()
	html, err := rg.GenerateHTML(report)
	require.NoError(t, err)

	// Verify all major sections are present
	sections := []string{
		"Summary",
		"Components",
		"Bottleneck",
		"CPU Profile",
		"Memory Profile",
		"Recommendation",
	}

	for _, section := range sections {
		assert.True(t, strings.Contains(html, section),
			"HTML should contain section: %s", section)
	}

	// Verify specific data is rendered
	assert.Contains(t, html, "MainView")
	assert.Contains(t, html, "59.5")
	assert.Contains(t, html, "Render exceeds frame budget")
	assert.Contains(t, html, "Implement Memoization")
}
