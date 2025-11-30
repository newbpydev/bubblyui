// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"sync"
	"time"
)

// ProfileData aggregates all profiler data for report generation.
//
// This struct collects data from various profiler components to generate
// a comprehensive performance report. Fields can be nil if that data
// source was not used during profiling.
//
// Example:
//
//	data := &ProfileData{
//	    ComponentTracker: tracker,
//	    StartTime:        startTime,
//	    EndTime:          time.Now(),
//	}
//	report := generator.Generate(data)
type ProfileData struct {
	// ComponentTracker provides per-component metrics
	ComponentTracker *ComponentTracker

	// Collector provides timing and counter metrics
	Collector *MetricCollector

	// CPUProfiler provides CPU profiling data
	CPUProfiler *CPUProfiler

	// MemoryProfiler provides memory profiling data
	MemoryProfiler *MemoryProfiler

	// RenderProfiler provides render performance data
	RenderProfiler *RenderProfiler

	// BottleneckDetector provides detected bottlenecks
	BottleneckDetector *BottleneckDetector

	// RecommendationEngine provides optimization recommendations
	RecommendationEngine *RecommendationEngine

	// Bottlenecks is a direct list of bottlenecks (alternative to detector)
	Bottlenecks []*BottleneckInfo

	// Recommendations is a direct list of recommendations (alternative to engine)
	Recommendations []*Recommendation

	// CPUProfile is direct CPU profile data (alternative to profiler)
	CPUProfile *CPUProfileData

	// MemProfile is direct memory profile data (alternative to profiler)
	MemProfile *MemProfileData

	// StartTime is when profiling started
	StartTime time.Time

	// EndTime is when profiling ended
	EndTime time.Time
}

// ReportGenerator generates performance reports from profiling data.
//
// It uses Go's html/template package to render reports as HTML with
// proper escaping for security. Custom templates can be provided for
// different report formats.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	rg := NewReportGenerator()
//	report := rg.Generate(profileData)
//	err := rg.SaveHTML(report, "performance-report.html")
type ReportGenerator struct {
	// templates holds the HTML templates for report generation
	templates *template.Template

	// mu protects concurrent access to generator state
	mu sync.RWMutex
}

// NewReportGenerator creates a new ReportGenerator with the default HTML template.
//
// The default template includes sections for summary, components, bottlenecks,
// CPU profile, memory profile, and recommendations.
//
// Example:
//
//	rg := NewReportGenerator()
//	report := rg.Generate(data)
func NewReportGenerator() *ReportGenerator {
	tmpl := template.New("report").Funcs(templateFuncs())
	tmpl = template.Must(tmpl.Parse(defaultReportTemplate))

	return &ReportGenerator{
		templates: tmpl,
	}
}

// NewReportGeneratorWithTemplate creates a ReportGenerator with a custom template.
//
// The template must define a "report" template that accepts a *Report.
//
// Example:
//
//	tmpl := template.Must(template.New("report").Parse(customTemplate))
//	rg := NewReportGeneratorWithTemplate(tmpl)
func NewReportGeneratorWithTemplate(tmpl *template.Template) *ReportGenerator {
	if tmpl == nil {
		return NewReportGenerator()
	}
	return &ReportGenerator{
		templates: tmpl,
	}
}

// Generate creates a Report from ProfileData.
//
// If data is nil, returns an empty report with default values.
// All fields in the report are properly initialized.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	data := &ProfileData{
//	    ComponentTracker: tracker,
//	    StartTime:        startTime,
//	    EndTime:          time.Now(),
//	}
//	report := rg.Generate(data)
func (rg *ReportGenerator) Generate(data *ProfileData) *Report {
	report := &Report{
		Summary:         &Summary{},
		Components:      make([]*ComponentMetrics, 0),
		Bottlenecks:     make([]*BottleneckInfo, 0),
		CPUProfile:      &CPUProfileData{},
		MemProfile:      &MemProfileData{},
		Recommendations: make([]*Recommendation, 0),
		Timestamp:       time.Now(),
	}

	if data == nil {
		return report
	}

	// Calculate duration
	if !data.StartTime.IsZero() && !data.EndTime.IsZero() {
		report.Summary.Duration = data.EndTime.Sub(data.StartTime)
	}

	// Extract component metrics
	if data.ComponentTracker != nil {
		metrics := data.ComponentTracker.GetAllMetrics()
		for _, m := range metrics {
			report.Components = append(report.Components, m)
		}
	}

	// Extract bottlenecks
	if len(data.Bottlenecks) > 0 {
		report.Bottlenecks = data.Bottlenecks
	}

	// Extract recommendations
	if len(data.Recommendations) > 0 {
		report.Recommendations = data.Recommendations
	}

	// Extract CPU profile data
	if data.CPUProfile != nil {
		report.CPUProfile = data.CPUProfile
	}

	// Extract memory profile data
	if data.MemProfile != nil {
		report.MemProfile = data.MemProfile
	}

	return report
}

// GenerateHTML renders the report as an HTML string.
//
// If report is nil, generates HTML for an empty report.
// Uses html/template for proper escaping to prevent XSS attacks.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	html, err := rg.GenerateHTML(report)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(html)
func (rg *ReportGenerator) GenerateHTML(report *Report) (string, error) {
	rg.mu.RLock()
	tmpl := rg.templates
	rg.mu.RUnlock()

	// Use empty report if nil
	if report == nil {
		report = &Report{
			Summary:         &Summary{},
			Components:      make([]*ComponentMetrics, 0),
			Bottlenecks:     make([]*BottleneckInfo, 0),
			CPUProfile:      &CPUProfileData{},
			MemProfile:      &MemProfileData{},
			Recommendations: make([]*Recommendation, 0),
			Timestamp:       time.Now(),
		}
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "report", report); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// SaveHTML generates HTML and writes it to a file.
//
// Creates or overwrites the file at the specified path.
// Returns an error if the file cannot be created or written.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	err := rg.SaveHTML(report, "performance-report.html")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (rg *ReportGenerator) SaveHTML(report *Report, filename string) error {
	html, err := rg.GenerateHTML(report)
	if err != nil {
		return fmt.Errorf("failed to generate HTML: %w", err)
	}

	if err := os.WriteFile(filename, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// GetTemplate returns the current HTML template.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (rg *ReportGenerator) GetTemplate() *template.Template {
	rg.mu.RLock()
	defer rg.mu.RUnlock()
	return rg.templates
}

// SetTemplate sets a custom HTML template.
//
// The template must define a "report" template that accepts a *Report.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (rg *ReportGenerator) SetTemplate(tmpl *template.Template) {
	rg.mu.Lock()
	defer rg.mu.Unlock()
	rg.templates = tmpl
}

// templateFuncs returns the template functions for report rendering.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"formatDuration": formatDuration,
		"formatBytes":    formatBytesUint,
		"formatPercent":  formatPercent,
		"severityClass":  severityClass,
		"priorityClass":  priorityClass,
		"priorityString": priorityString,
	}
}

// formatDuration formats a duration for display.
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return d.String()
}

// formatBytesUint formats bytes (uint64) for human-readable display.
// This is separate from formatBytes in leak_detector.go which takes int64.
func formatBytesUint(b uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case b >= GB:
		return fmt.Sprintf("%.2f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.2f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.2f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// formatPercent formats a percentage for display.
func formatPercent(p float64) string {
	return fmt.Sprintf("%.1f%%", p)
}

// severityClass returns a CSS class for the severity level.
func severityClass(s Severity) string {
	switch s {
	case SeverityCritical:
		return "severity-critical"
	case SeverityHigh:
		return "severity-high"
	case SeverityMedium:
		return "severity-medium"
	case SeverityLow:
		return "severity-low"
	default:
		return "severity-unknown"
	}
}

// priorityClass returns a CSS class for the priority level.
func priorityClass(p Priority) string {
	switch p {
	case PriorityCritical:
		return "priority-critical"
	case PriorityHigh:
		return "priority-high"
	case PriorityMedium:
		return "priority-medium"
	case PriorityLow:
		return "priority-low"
	default:
		return "priority-unknown"
	}
}

// priorityString returns a human-readable string for the priority level.
func priorityString(p Priority) string {
	switch p {
	case PriorityCritical:
		return "Critical"
	case PriorityHigh:
		return "High"
	case PriorityMedium:
		return "Medium"
	case PriorityLow:
		return "Low"
	default:
		return "Unknown"
	}
}

// defaultReportTemplate is the default HTML template for performance reports.
const defaultReportTemplate = `{{define "report"}}<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>BubblyUI Performance Report</title>
    <style>
        :root {
            --color-critical: #dc2626;
            --color-high: #ea580c;
            --color-medium: #ca8a04;
            --color-low: #16a34a;
            --color-bg: #f8fafc;
            --color-card: #ffffff;
            --color-border: #e2e8f0;
            --color-text: #1e293b;
            --color-muted: #64748b;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--color-bg);
            color: var(--color-text);
            line-height: 1.6;
            padding: 2rem;
        }
        .container { max-width: 1200px; margin: 0 auto; }
        h1 { font-size: 2rem; margin-bottom: 0.5rem; }
        h2 { font-size: 1.5rem; margin: 2rem 0 1rem; border-bottom: 2px solid var(--color-border); padding-bottom: 0.5rem; }
        h3 { font-size: 1.25rem; margin: 1rem 0 0.5rem; }
        .timestamp { color: var(--color-muted); font-size: 0.875rem; margin-bottom: 2rem; }
        .card {
            background: var(--color-card);
            border: 1px solid var(--color-border);
            border-radius: 0.5rem;
            padding: 1.5rem;
            margin-bottom: 1rem;
        }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; }
        .stat { text-align: center; }
        .stat-value { font-size: 2rem; font-weight: bold; }
        .stat-label { color: var(--color-muted); font-size: 0.875rem; }
        table { width: 100%; border-collapse: collapse; margin: 1rem 0; }
        th, td { padding: 0.75rem; text-align: left; border-bottom: 1px solid var(--color-border); }
        th { background: var(--color-bg); font-weight: 600; }
        .severity-critical { color: var(--color-critical); font-weight: bold; }
        .severity-high { color: var(--color-high); font-weight: bold; }
        .severity-medium { color: var(--color-medium); }
        .severity-low { color: var(--color-low); }
        .priority-critical { background: var(--color-critical); color: white; padding: 0.25rem 0.5rem; border-radius: 0.25rem; }
        .priority-high { background: var(--color-high); color: white; padding: 0.25rem 0.5rem; border-radius: 0.25rem; }
        .priority-medium { background: var(--color-medium); color: white; padding: 0.25rem 0.5rem; border-radius: 0.25rem; }
        .priority-low { background: var(--color-low); color: white; padding: 0.25rem 0.5rem; border-radius: 0.25rem; }
        .empty { color: var(--color-muted); font-style: italic; }
        .recommendation { margin-bottom: 1rem; padding: 1rem; border-left: 4px solid var(--color-border); }
        .recommendation.priority-critical { border-left-color: var(--color-critical); }
        .recommendation.priority-high { border-left-color: var(--color-high); }
        .recommendation.priority-medium { border-left-color: var(--color-medium); }
        .recommendation.priority-low { border-left-color: var(--color-low); }
    </style>
</head>
<body>
    <div class="container">
        <h1>BubblyUI Performance Report</h1>
        <p class="timestamp">Generated: {{.Timestamp.Format "2006-01-02 15:04:05"}}</p>

        <h2>Summary</h2>
        <div class="card">
            <div class="grid">
                <div class="stat">
                    <div class="stat-value">{{if .Summary}}{{formatDuration .Summary.Duration}}{{else}}0s{{end}}</div>
                    <div class="stat-label">Duration</div>
                </div>
                <div class="stat">
                    <div class="stat-value">{{if .Summary}}{{.Summary.TotalOperations}}{{else}}0{{end}}</div>
                    <div class="stat-label">Total Operations</div>
                </div>
                <div class="stat">
                    <div class="stat-value">{{if .Summary}}{{printf "%.1f" .Summary.AverageFPS}}{{else}}0.0{{end}}</div>
                    <div class="stat-label">Average FPS</div>
                </div>
                <div class="stat">
                    <div class="stat-value">{{if .Summary}}{{formatBytes .Summary.MemoryUsage}}{{else}}0 B{{end}}</div>
                    <div class="stat-label">Memory Usage</div>
                </div>
                <div class="stat">
                    <div class="stat-value">{{if .Summary}}{{.Summary.GoroutineCount}}{{else}}0{{end}}</div>
                    <div class="stat-label">Goroutines</div>
                </div>
            </div>
        </div>

        <h2>Components</h2>
        <div class="card">
            {{if .Components}}
            <table>
                <thead>
                    <tr>
                        <th>Component</th>
                        <th>Renders</th>
                        <th>Avg Time</th>
                        <th>Max Time</th>
                        <th>Memory</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Components}}
                    <tr>
                        <td>{{.ComponentName}}</td>
                        <td>{{.RenderCount}}</td>
                        <td>{{formatDuration .AvgRenderTime}}</td>
                        <td>{{formatDuration .MaxRenderTime}}</td>
                        <td>{{formatBytes .MemoryUsage}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
            {{else}}
            <p class="empty">No component data available.</p>
            {{end}}
        </div>

        <h2>Bottlenecks</h2>
        <div class="card">
            {{if .Bottlenecks}}
            <table>
                <thead>
                    <tr>
                        <th>Location</th>
                        <th>Type</th>
                        <th>Severity</th>
                        <th>Description</th>
                        <th>Suggestion</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Bottlenecks}}
                    <tr>
                        <td>{{.Location}}</td>
                        <td>{{.Type}}</td>
                        <td class="{{severityClass .Severity}}">{{.Severity}}</td>
                        <td>{{.Description}}</td>
                        <td>{{.Suggestion}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
            {{else}}
            <p class="empty">No bottlenecks detected.</p>
            {{end}}
        </div>

        <h2>CPU Profile</h2>
        <div class="card">
            {{if and .CPUProfile .CPUProfile.HotFunctions}}
            <p>Total Samples: {{.CPUProfile.TotalSamples}}</p>
            <h3>Hot Functions</h3>
            <table>
                <thead>
                    <tr>
                        <th>Function</th>
                        <th>Samples</th>
                        <th>Percent</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .CPUProfile.HotFunctions}}
                    <tr>
                        <td>{{.Name}}</td>
                        <td>{{.Samples}}</td>
                        <td>{{formatPercent .Percent}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
            {{else}}
            <p class="empty">No CPU profile data available.</p>
            {{end}}
        </div>

        <h2>Memory Profile</h2>
        <div class="card">
            {{if .MemProfile}}
            <div class="grid">
                <div class="stat">
                    <div class="stat-value">{{formatBytes .MemProfile.HeapAlloc}}</div>
                    <div class="stat-label">Heap Allocated</div>
                </div>
                <div class="stat">
                    <div class="stat-value">{{.MemProfile.HeapObjects}}</div>
                    <div class="stat-label">Heap Objects</div>
                </div>
            </div>
            {{if .MemProfile.GCPauses}}
            <h3>GC Pauses</h3>
            <p>{{range $i, $pause := .MemProfile.GCPauses}}{{if $i}}, {{end}}{{formatDuration $pause}}{{end}}</p>
            {{end}}
            {{else}}
            <p class="empty">No memory profile data available.</p>
            {{end}}
        </div>

        <h2>Recommendations</h2>
        <div class="card">
            {{if .Recommendations}}
            {{range .Recommendations}}
            <div class="recommendation {{priorityClass .Priority}}">
                <h3><span class="{{priorityClass .Priority}}">{{priorityString .Priority}}</span> {{.Title}}</h3>
                <p>{{.Description}}</p>
                <p><strong>Action:</strong> {{.Action}}</p>
            </div>
            {{end}}
            {{else}}
            <p class="empty">No recommendations at this time.</p>
            {{end}}
        </div>
    </div>
</body>
</html>{{end}}`
