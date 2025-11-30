// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExporter(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "creates exporter with default settings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExporter()

			require.NotNil(t, e)
		})
	}
}

func TestExporter_ExportHTML(t *testing.T) {
	baseTime := time.Date(2024, 11, 29, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		report         *Report
		expectError    bool
		checkContent   bool
		expectedInHTML []string
	}{
		{
			name:         "nil report creates valid HTML",
			report:       nil,
			expectError:  false,
			checkContent: true,
			expectedInHTML: []string{
				"<!DOCTYPE html>",
				"<html",
				"BubblyUI Performance Report",
			},
		},
		{
			name: "report with summary exports correctly",
			report: &Report{
				Summary: &Summary{
					Duration:        5 * time.Second,
					TotalOperations: 1000,
					AverageFPS:      60.0,
					MemoryUsage:     1024 * 1024,
					GoroutineCount:  10,
				},
				Components:      []*ComponentMetrics{},
				Bottlenecks:     []*BottleneckInfo{},
				CPUProfile:      &CPUProfileData{},
				MemProfile:      &MemProfileData{},
				Recommendations: []*Recommendation{},
				Timestamp:       baseTime,
			},
			expectError:  false,
			checkContent: true,
			expectedInHTML: []string{
				"<!DOCTYPE html>",
				"Duration",
				"Total Operations",
			},
		},
		{
			name: "report with components exports correctly",
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
						MemoryUsage:     2048,
					},
				},
				Bottlenecks:     []*BottleneckInfo{},
				CPUProfile:      &CPUProfileData{},
				MemProfile:      &MemProfileData{},
				Recommendations: []*Recommendation{},
				Timestamp:       baseTime,
			},
			expectError:  false,
			checkContent: true,
			expectedInHTML: []string{
				"TestComponent",
			},
		},
		{
			name: "report with bottlenecks exports correctly",
			report: &Report{
				Summary:    &Summary{},
				Components: []*ComponentMetrics{},
				Bottlenecks: []*BottleneckInfo{
					{
						Type:        BottleneckTypeSlow,
						Location:    "render",
						Severity:    SeverityHigh,
						Impact:      0.8,
						Description: "Slow render detected",
						Suggestion:  "Optimize render function",
					},
				},
				CPUProfile:      &CPUProfileData{},
				MemProfile:      &MemProfileData{},
				Recommendations: []*Recommendation{},
				Timestamp:       baseTime,
			},
			expectError:  false,
			checkContent: true,
			expectedInHTML: []string{
				"Slow render detected",
				"Optimize render function",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExporter()

			// Create temp file
			tmpDir := t.TempDir()
			filename := filepath.Join(tmpDir, "report.html")

			err := e.ExportHTML(tt.report, filename)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify file exists
			_, err = os.Stat(filename)
			require.NoError(t, err, "file should exist")

			if tt.checkContent {
				content, err := os.ReadFile(filename)
				require.NoError(t, err)

				for _, expected := range tt.expectedInHTML {
					assert.Contains(t, string(content), expected, "HTML should contain: %s", expected)
				}
			}
		})
	}
}

func TestExporter_ExportJSON(t *testing.T) {
	baseTime := time.Date(2024, 11, 29, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		report       *Report
		expectError  bool
		validateJSON bool
		expectedKeys []string
	}{
		{
			name:         "nil report creates valid JSON",
			report:       nil,
			expectError:  false,
			validateJSON: true,
			expectedKeys: []string{"summary", "components", "bottlenecks", "timestamp"},
		},
		{
			name: "report with all fields exports correctly",
			report: &Report{
				Summary: &Summary{
					Duration:        5 * time.Second,
					TotalOperations: 1000,
					AverageFPS:      60.0,
					MemoryUsage:     1024 * 1024,
					GoroutineCount:  10,
				},
				Components: []*ComponentMetrics{
					{
						ComponentID:     "comp1",
						ComponentName:   "TestComponent",
						RenderCount:     100,
						TotalRenderTime: 500 * time.Millisecond,
						AvgRenderTime:   5 * time.Millisecond,
						MaxRenderTime:   20 * time.Millisecond,
						MemoryUsage:     2048,
					},
				},
				Bottlenecks: []*BottleneckInfo{
					{
						Type:        BottleneckTypeSlow,
						Location:    "render",
						Severity:    SeverityHigh,
						Impact:      0.8,
						Description: "Slow render",
						Suggestion:  "Optimize",
					},
				},
				CPUProfile: &CPUProfileData{
					HotFunctions: []*HotFunction{
						{Name: "main.render", Samples: 100, Percent: 50.0},
					},
					TotalSamples: 200,
				},
				MemProfile: &MemProfileData{
					HeapAlloc:   1024 * 1024,
					HeapObjects: 1000,
					GCPauses:    []time.Duration{1 * time.Millisecond, 2 * time.Millisecond},
				},
				Recommendations: []*Recommendation{
					{
						Title:       "Optimize render",
						Description: "Render is slow",
						Action:      "Use memoization",
						Priority:    PriorityHigh,
						Category:    CategoryRendering,
						Impact:      ImpactHigh,
					},
				},
				Timestamp: baseTime,
			},
			expectError:  false,
			validateJSON: true,
			expectedKeys: []string{"summary", "components", "bottlenecks", "cpu_profile", "mem_profile", "recommendations", "timestamp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExporter()

			// Create temp file
			tmpDir := t.TempDir()
			filename := filepath.Join(tmpDir, "report.json")

			err := e.ExportJSON(tt.report, filename)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify file exists
			_, err = os.Stat(filename)
			require.NoError(t, err, "file should exist")

			if tt.validateJSON {
				content, err := os.ReadFile(filename)
				require.NoError(t, err)

				// Verify valid JSON
				var result map[string]interface{}
				err = json.Unmarshal(content, &result)
				require.NoError(t, err, "should be valid JSON")

				// Check expected keys
				for _, key := range tt.expectedKeys {
					_, exists := result[key]
					assert.True(t, exists, "JSON should contain key: %s", key)
				}
			}
		})
	}
}

func TestExporter_ExportCSV(t *testing.T) {
	baseTime := time.Date(2024, 11, 29, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		report          *Report
		expectError     bool
		validateCSV     bool
		expectedHeaders []string
		expectedRows    int
	}{
		{
			name:            "nil report creates valid CSV with headers only",
			report:          nil,
			expectError:     false,
			validateCSV:     true,
			expectedHeaders: []string{"component_id", "component_name", "render_count", "avg_render_time_ns", "max_render_time_ns", "memory_usage"},
			expectedRows:    1, // Header only
		},
		{
			name: "report with components exports correctly",
			report: &Report{
				Summary: &Summary{},
				Components: []*ComponentMetrics{
					{
						ComponentID:     "comp1",
						ComponentName:   "TestComponent1",
						RenderCount:     100,
						TotalRenderTime: 500 * time.Millisecond,
						AvgRenderTime:   5 * time.Millisecond,
						MaxRenderTime:   20 * time.Millisecond,
						MemoryUsage:     2048,
					},
					{
						ComponentID:     "comp2",
						ComponentName:   "TestComponent2",
						RenderCount:     200,
						TotalRenderTime: 1 * time.Second,
						AvgRenderTime:   10 * time.Millisecond,
						MaxRenderTime:   50 * time.Millisecond,
						MemoryUsage:     4096,
					},
				},
				Bottlenecks:     []*BottleneckInfo{},
				CPUProfile:      &CPUProfileData{},
				MemProfile:      &MemProfileData{},
				Recommendations: []*Recommendation{},
				Timestamp:       baseTime,
			},
			expectError:     false,
			validateCSV:     true,
			expectedHeaders: []string{"component_id", "component_name", "render_count", "avg_render_time_ns", "max_render_time_ns", "memory_usage"},
			expectedRows:    3, // Header + 2 components
		},
		{
			name: "empty components creates CSV with headers only",
			report: &Report{
				Summary:         &Summary{},
				Components:      []*ComponentMetrics{},
				Bottlenecks:     []*BottleneckInfo{},
				CPUProfile:      &CPUProfileData{},
				MemProfile:      &MemProfileData{},
				Recommendations: []*Recommendation{},
				Timestamp:       baseTime,
			},
			expectError:     false,
			validateCSV:     true,
			expectedHeaders: []string{"component_id", "component_name", "render_count", "avg_render_time_ns", "max_render_time_ns", "memory_usage"},
			expectedRows:    1, // Header only
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExporter()

			// Create temp file
			tmpDir := t.TempDir()
			filename := filepath.Join(tmpDir, "report.csv")

			err := e.ExportCSV(tt.report, filename)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify file exists
			_, err = os.Stat(filename)
			require.NoError(t, err, "file should exist")

			if tt.validateCSV {
				file, err := os.Open(filename)
				require.NoError(t, err)
				defer file.Close()

				reader := csv.NewReader(file)
				records, err := reader.ReadAll()
				require.NoError(t, err, "should be valid CSV")

				// Check row count
				assert.Equal(t, tt.expectedRows, len(records), "CSV should have expected number of rows")

				// Check headers
				if len(records) > 0 {
					for i, header := range tt.expectedHeaders {
						if i < len(records[0]) {
							assert.Equal(t, header, records[0][i], "header at position %d should match", i)
						}
					}
				}
			}
		})
	}
}

func TestExporter_ExportHTML_InvalidPath(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectError bool
	}{
		{
			name:        "invalid directory path returns error",
			filename:    "/nonexistent/directory/report.html",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExporter()
			report := &Report{
				Summary:         &Summary{},
				Components:      []*ComponentMetrics{},
				Bottlenecks:     []*BottleneckInfo{},
				CPUProfile:      &CPUProfileData{},
				MemProfile:      &MemProfileData{},
				Recommendations: []*Recommendation{},
				Timestamp:       time.Now(),
			}

			err := e.ExportHTML(report, tt.filename)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExporter_ExportJSON_InvalidPath(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectError bool
	}{
		{
			name:        "invalid directory path returns error",
			filename:    "/nonexistent/directory/report.json",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExporter()
			report := &Report{
				Summary:         &Summary{},
				Components:      []*ComponentMetrics{},
				Bottlenecks:     []*BottleneckInfo{},
				CPUProfile:      &CPUProfileData{},
				MemProfile:      &MemProfileData{},
				Recommendations: []*Recommendation{},
				Timestamp:       time.Now(),
			}

			err := e.ExportJSON(report, tt.filename)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExporter_ExportCSV_InvalidPath(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectError bool
	}{
		{
			name:        "invalid directory path returns error",
			filename:    "/nonexistent/directory/report.csv",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExporter()
			report := &Report{
				Summary:         &Summary{},
				Components:      []*ComponentMetrics{},
				Bottlenecks:     []*BottleneckInfo{},
				CPUProfile:      &CPUProfileData{},
				MemProfile:      &MemProfileData{},
				Recommendations: []*Recommendation{},
				Timestamp:       time.Now(),
			}

			err := e.ExportCSV(report, tt.filename)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExporter_ExportAll(t *testing.T) {
	baseTime := time.Date(2024, 11, 29, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		report      *Report
		expectError bool
	}{
		{
			name: "exports all formats successfully",
			report: &Report{
				Summary: &Summary{
					Duration:        5 * time.Second,
					TotalOperations: 1000,
					AverageFPS:      60.0,
					MemoryUsage:     1024 * 1024,
					GoroutineCount:  10,
				},
				Components: []*ComponentMetrics{
					{
						ComponentID:   "comp1",
						ComponentName: "TestComponent",
						RenderCount:   100,
						AvgRenderTime: 5 * time.Millisecond,
						MaxRenderTime: 20 * time.Millisecond,
						MemoryUsage:   2048,
					},
				},
				Bottlenecks:     []*BottleneckInfo{},
				CPUProfile:      &CPUProfileData{},
				MemProfile:      &MemProfileData{},
				Recommendations: []*Recommendation{},
				Timestamp:       baseTime,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExporter()
			tmpDir := t.TempDir()

			htmlFile := filepath.Join(tmpDir, "report.html")
			jsonFile := filepath.Join(tmpDir, "report.json")
			csvFile := filepath.Join(tmpDir, "report.csv")

			err := e.ExportAll(tt.report, htmlFile, jsonFile, csvFile)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify all files exist
			_, err = os.Stat(htmlFile)
			assert.NoError(t, err, "HTML file should exist")

			_, err = os.Stat(jsonFile)
			assert.NoError(t, err, "JSON file should exist")

			_, err = os.Stat(csvFile)
			assert.NoError(t, err, "CSV file should exist")
		})
	}
}

func TestExporter_ExportToString(t *testing.T) {
	baseTime := time.Date(2024, 11, 29, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		report      *Report
		format      ExportFormat
		expectError bool
		checkOutput func(t *testing.T, output string)
	}{
		{
			name: "export to HTML string",
			report: &Report{
				Summary:         &Summary{Duration: 5 * time.Second},
				Components:      []*ComponentMetrics{},
				Bottlenecks:     []*BottleneckInfo{},
				CPUProfile:      &CPUProfileData{},
				MemProfile:      &MemProfileData{},
				Recommendations: []*Recommendation{},
				Timestamp:       baseTime,
			},
			format:      FormatHTML,
			expectError: false,
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "<!DOCTYPE html>")
				assert.Contains(t, output, "BubblyUI Performance Report")
			},
		},
		{
			name: "export to JSON string",
			report: &Report{
				Summary:         &Summary{Duration: 5 * time.Second},
				Components:      []*ComponentMetrics{},
				Bottlenecks:     []*BottleneckInfo{},
				CPUProfile:      &CPUProfileData{},
				MemProfile:      &MemProfileData{},
				Recommendations: []*Recommendation{},
				Timestamp:       baseTime,
			},
			format:      FormatJSON,
			expectError: false,
			checkOutput: func(t *testing.T, output string) {
				var result map[string]interface{}
				err := json.Unmarshal([]byte(output), &result)
				assert.NoError(t, err, "should be valid JSON")
			},
		},
		{
			name: "export to CSV string",
			report: &Report{
				Summary: &Summary{},
				Components: []*ComponentMetrics{
					{ComponentID: "c1", ComponentName: "Test", RenderCount: 10},
				},
				Bottlenecks:     []*BottleneckInfo{},
				CPUProfile:      &CPUProfileData{},
				MemProfile:      &MemProfileData{},
				Recommendations: []*Recommendation{},
				Timestamp:       baseTime,
			},
			format:      FormatCSV,
			expectError: false,
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "component_id")
				assert.Contains(t, output, "Test")
			},
		},
		{
			name: "invalid format returns error",
			report: &Report{
				Summary:         &Summary{},
				Components:      []*ComponentMetrics{},
				Bottlenecks:     []*BottleneckInfo{},
				CPUProfile:      &CPUProfileData{},
				MemProfile:      &MemProfileData{},
				Recommendations: []*Recommendation{},
				Timestamp:       baseTime,
			},
			format:      ExportFormat("invalid"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExporter()

			output, err := e.ExportToString(tt.report, tt.format)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, output)

			if tt.checkOutput != nil {
				tt.checkOutput(t, output)
			}
		})
	}
}

func TestExporter_ThreadSafety(t *testing.T) {
	e := NewExporter()
	report := &Report{
		Summary: &Summary{
			Duration:        5 * time.Second,
			TotalOperations: 1000,
			AverageFPS:      60.0,
			MemoryUsage:     1024 * 1024,
			GoroutineCount:  10,
		},
		Components: []*ComponentMetrics{
			{ComponentID: "c1", ComponentName: "Test", RenderCount: 10},
		},
		Bottlenecks:     []*BottleneckInfo{},
		CPUProfile:      &CPUProfileData{},
		MemProfile:      &MemProfileData{},
		Recommendations: []*Recommendation{},
		Timestamp:       time.Now(),
	}

	tmpDir := t.TempDir()
	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(3)

		// HTML export
		go func(idx int) {
			defer wg.Done()
			filename := filepath.Join(tmpDir, "report_"+string(rune('a'+idx%26))+".html")
			_ = e.ExportHTML(report, filename)
		}(i)

		// JSON export
		go func(idx int) {
			defer wg.Done()
			filename := filepath.Join(tmpDir, "report_"+string(rune('a'+idx%26))+".json")
			_ = e.ExportJSON(report, filename)
		}(i)

		// CSV export
		go func(idx int) {
			defer wg.Done()
			filename := filepath.Join(tmpDir, "report_"+string(rune('a'+idx%26))+".csv")
			_ = e.ExportCSV(report, filename)
		}(i)
	}

	wg.Wait()
	// If we get here without race detector errors, thread safety is verified
}

func TestExporter_JSONFormat_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name        string
		report      *Report
		expectError bool
	}{
		{
			name: "handles special characters in component names",
			report: &Report{
				Summary: &Summary{},
				Components: []*ComponentMetrics{
					{
						ComponentID:   "comp<1>",
						ComponentName: "Test\"Component'With<Special>&Chars",
						RenderCount:   100,
					},
				},
				Bottlenecks:     []*BottleneckInfo{},
				CPUProfile:      &CPUProfileData{},
				MemProfile:      &MemProfileData{},
				Recommendations: []*Recommendation{},
				Timestamp:       time.Now(),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExporter()
			tmpDir := t.TempDir()
			filename := filepath.Join(tmpDir, "report.json")

			err := e.ExportJSON(tt.report, filename)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify valid JSON
			content, err := os.ReadFile(filename)
			require.NoError(t, err)

			var result map[string]interface{}
			err = json.Unmarshal(content, &result)
			require.NoError(t, err, "should be valid JSON even with special characters")
		})
	}
}

func TestExporter_CSVFormat_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name        string
		report      *Report
		expectError bool
	}{
		{
			name: "handles special characters in component names",
			report: &Report{
				Summary: &Summary{},
				Components: []*ComponentMetrics{
					{
						ComponentID:   "comp,1",
						ComponentName: "Test,Component\"With\nNewlines",
						RenderCount:   100,
					},
				},
				Bottlenecks:     []*BottleneckInfo{},
				CPUProfile:      &CPUProfileData{},
				MemProfile:      &MemProfileData{},
				Recommendations: []*Recommendation{},
				Timestamp:       time.Now(),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExporter()
			tmpDir := t.TempDir()
			filename := filepath.Join(tmpDir, "report.csv")

			err := e.ExportCSV(tt.report, filename)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify valid CSV
			file, err := os.Open(filename)
			require.NoError(t, err)
			defer file.Close()

			reader := csv.NewReader(file)
			records, err := reader.ReadAll()
			require.NoError(t, err, "should be valid CSV even with special characters")
			assert.Equal(t, 2, len(records), "should have header and 1 data row")
		})
	}
}

func TestExporter_HTMLFormat_XSSProtection(t *testing.T) {
	tests := []struct {
		name             string
		report           *Report
		expectError      bool
		shouldNotContain []string
	}{
		{
			name: "escapes XSS in component names",
			report: &Report{
				Summary: &Summary{},
				Components: []*ComponentMetrics{
					{
						ComponentID:   "comp1",
						ComponentName: "<script>alert('xss')</script>",
						RenderCount:   100,
					},
				},
				Bottlenecks:     []*BottleneckInfo{},
				CPUProfile:      &CPUProfileData{},
				MemProfile:      &MemProfileData{},
				Recommendations: []*Recommendation{},
				Timestamp:       time.Now(),
			},
			expectError: false,
			shouldNotContain: []string{
				"<script>alert('xss')</script>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExporter()
			tmpDir := t.TempDir()
			filename := filepath.Join(tmpDir, "report.html")

			err := e.ExportHTML(tt.report, filename)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			content, err := os.ReadFile(filename)
			require.NoError(t, err)

			for _, notExpected := range tt.shouldNotContain {
				assert.NotContains(t, string(content), notExpected, "HTML should escape: %s", notExpected)
			}

			// Should contain escaped version
			assert.Contains(t, string(content), "&lt;script&gt;", "should contain escaped script tag")
		})
	}
}

func TestExporter_GetSupportedFormats(t *testing.T) {
	e := NewExporter()
	formats := e.GetSupportedFormats()

	assert.Contains(t, formats, FormatHTML)
	assert.Contains(t, formats, FormatJSON)
	assert.Contains(t, formats, FormatCSV)
	assert.Equal(t, 3, len(formats))
}

func TestExportFormat_String(t *testing.T) {
	tests := []struct {
		format   ExportFormat
		expected string
	}{
		{FormatHTML, "html"},
		{FormatJSON, "json"},
		{FormatCSV, "csv"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.format))
		})
	}
}

func TestExporter_ExportJSON_PrettyPrint(t *testing.T) {
	e := NewExporter()
	report := &Report{
		Summary:         &Summary{Duration: 5 * time.Second},
		Components:      []*ComponentMetrics{},
		Bottlenecks:     []*BottleneckInfo{},
		CPUProfile:      &CPUProfileData{},
		MemProfile:      &MemProfileData{},
		Recommendations: []*Recommendation{},
		Timestamp:       time.Now(),
	}

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "report.json")

	err := e.ExportJSON(report, filename)
	require.NoError(t, err)

	content, err := os.ReadFile(filename)
	require.NoError(t, err)

	// Check for indentation (pretty print)
	assert.True(t, strings.Contains(string(content), "\n"), "JSON should be pretty printed with newlines")
	assert.True(t, strings.Contains(string(content), "  "), "JSON should be pretty printed with indentation")
}

func TestExporter_ExportAll_PartialFailure(t *testing.T) {
	e := NewExporter()
	report := &Report{
		Summary:         &Summary{Duration: 5 * time.Second},
		Components:      []*ComponentMetrics{},
		Bottlenecks:     []*BottleneckInfo{},
		CPUProfile:      &CPUProfileData{},
		MemProfile:      &MemProfileData{},
		Recommendations: []*Recommendation{},
		Timestamp:       time.Now(),
	}

	tmpDir := t.TempDir()

	// Test HTML failure
	err := e.ExportAll(report, "/nonexistent/dir/report.html", filepath.Join(tmpDir, "report.json"), filepath.Join(tmpDir, "report.csv"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTML export failed")

	// Test JSON failure (HTML succeeds)
	err = e.ExportAll(report, filepath.Join(tmpDir, "report.html"), "/nonexistent/dir/report.json", filepath.Join(tmpDir, "report.csv"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JSON export failed")

	// Test CSV failure (HTML and JSON succeed)
	err = e.ExportAll(report, filepath.Join(tmpDir, "report2.html"), filepath.Join(tmpDir, "report2.json"), "/nonexistent/dir/report.csv")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CSV export failed")
}

func TestExporter_ReportToJSON_AllPriorities(t *testing.T) {
	baseTime := time.Date(2024, 11, 29, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		priority Priority
		expected string
	}{
		{name: "critical priority", priority: PriorityCritical, expected: "critical"},
		{name: "high priority", priority: PriorityHigh, expected: "high"},
		{name: "medium priority", priority: PriorityMedium, expected: "medium"},
		{name: "low priority", priority: PriorityLow, expected: "low"},
		{name: "unknown priority", priority: Priority(99), expected: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExporter()
			report := &Report{
				Summary:     &Summary{},
				Components:  []*ComponentMetrics{},
				Bottlenecks: []*BottleneckInfo{},
				CPUProfile:  &CPUProfileData{},
				MemProfile:  &MemProfileData{},
				Recommendations: []*Recommendation{
					{
						Title:       "Test",
						Description: "Test desc",
						Action:      "Test action",
						Priority:    tt.priority,
						Category:    CategoryOptimization,
						Impact:      ImpactHigh,
					},
				},
				Timestamp: baseTime,
			}

			tmpDir := t.TempDir()
			filename := filepath.Join(tmpDir, "report.json")

			err := e.ExportJSON(report, filename)
			require.NoError(t, err)

			content, err := os.ReadFile(filename)
			require.NoError(t, err)

			assert.Contains(t, string(content), tt.expected)
		})
	}
}

func TestExporter_ReportToJSON_NilFields(t *testing.T) {
	e := NewExporter()
	report := &Report{
		Summary: nil,
		Components: []*ComponentMetrics{
			nil, // nil component should be skipped
			{ComponentID: "c1", ComponentName: "Test"},
		},
		Bottlenecks: []*BottleneckInfo{
			nil, // nil bottleneck should be skipped
			{Type: BottleneckTypeSlow, Location: "test"},
		},
		CPUProfile: &CPUProfileData{
			HotFunctions: []*HotFunction{
				nil, // nil hot function should be skipped
				{Name: "main.test", Samples: 100, Percent: 50.0},
			},
		},
		MemProfile: nil,
		Recommendations: []*Recommendation{
			nil, // nil recommendation should be skipped
			{Title: "Test Rec"},
		},
		Timestamp: time.Now(),
	}

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "report.json")

	err := e.ExportJSON(report, filename)
	require.NoError(t, err)

	content, err := os.ReadFile(filename)
	require.NoError(t, err)

	// Verify valid JSON
	var result map[string]interface{}
	err = json.Unmarshal(content, &result)
	require.NoError(t, err)

	// Check that non-nil items are present
	assert.Contains(t, string(content), "Test")
	assert.Contains(t, string(content), "main.test")
	assert.Contains(t, string(content), "Test Rec")
}

func TestExporter_CSVToString_NilComponents(t *testing.T) {
	e := NewExporter()
	report := &Report{
		Summary: &Summary{},
		Components: []*ComponentMetrics{
			nil, // nil component should be skipped
			{ComponentID: "c1", ComponentName: "Test", RenderCount: 10},
			nil, // another nil
		},
		Bottlenecks:     []*BottleneckInfo{},
		CPUProfile:      &CPUProfileData{},
		MemProfile:      &MemProfileData{},
		Recommendations: []*Recommendation{},
		Timestamp:       time.Now(),
	}

	output, err := e.ExportToString(report, FormatCSV)
	require.NoError(t, err)

	// Should have header + 1 data row (nil components skipped)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 2, len(lines), "should have header and 1 data row")
	assert.Contains(t, output, "Test")
}
