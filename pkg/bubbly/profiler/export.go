// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// ExportFormat represents the output format for report export.
type ExportFormat string

const (
	// FormatHTML exports the report as an HTML file.
	FormatHTML ExportFormat = "html"

	// FormatJSON exports the report as a JSON file.
	FormatJSON ExportFormat = "json"

	// FormatCSV exports the report as a CSV file (component metrics only).
	FormatCSV ExportFormat = "csv"
)

// Exporter handles exporting performance reports to various formats.
//
// It supports HTML, JSON, and CSV export formats. HTML uses the ReportGenerator
// for rendering, JSON uses encoding/json with pretty printing, and CSV exports
// component metrics in a tabular format.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	e := NewExporter()
//	err := e.ExportHTML(report, "report.html")
//	err = e.ExportJSON(report, "report.json")
//	err = e.ExportCSV(report, "report.csv")
type Exporter struct {
	// reportGenerator handles HTML report generation
	reportGenerator *ReportGenerator

	// mu protects concurrent access to exporter state
	mu sync.RWMutex
}

// NewExporter creates a new Exporter with default settings.
//
// The exporter uses the default ReportGenerator for HTML export.
//
// Example:
//
//	e := NewExporter()
//	err := e.ExportHTML(report, "report.html")
func NewExporter() *Exporter {
	return &Exporter{
		reportGenerator: NewReportGenerator(),
	}
}

// ExportHTML exports the report as an HTML file.
//
// If report is nil, exports an empty report with default values.
// Uses html/template for proper escaping to prevent XSS attacks.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	err := e.ExportHTML(report, "performance-report.html")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (e *Exporter) ExportHTML(report *Report, filename string) error {
	e.mu.RLock()
	rg := e.reportGenerator
	e.mu.RUnlock()

	// Use empty report if nil
	if report == nil {
		report = createEmptyReport()
	}

	return rg.SaveHTML(report, filename)
}

// ExportJSON exports the report as a JSON file.
//
// If report is nil, exports an empty report with default values.
// The JSON is pretty-printed with indentation for readability.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	err := e.ExportJSON(report, "performance-report.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (e *Exporter) ExportJSON(report *Report, filename string) error {
	// Use empty report if nil
	if report == nil {
		report = createEmptyReport()
	}

	// Create JSON-friendly representation
	jsonReport := reportToJSON(report)

	// Marshal with pretty printing
	data, err := json.MarshalIndent(jsonReport, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

// ExportCSV exports the report's component metrics as a CSV file.
//
// If report is nil, exports only the header row.
// The CSV includes columns for component ID, name, render count,
// average render time, max render time, and memory usage.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	err := e.ExportCSV(report, "performance-report.csv")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (e *Exporter) ExportCSV(report *Report, filename string) error {
	// Use empty report if nil
	if report == nil {
		report = createEmptyReport()
	}

	// Create file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"component_id",
		"component_name",
		"render_count",
		"avg_render_time_ns",
		"max_render_time_ns",
		"memory_usage",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write component data
	for _, comp := range report.Components {
		if comp == nil {
			continue
		}
		row := []string{
			comp.ComponentID,
			comp.ComponentName,
			fmt.Sprintf("%d", comp.RenderCount),
			fmt.Sprintf("%d", comp.AvgRenderTime.Nanoseconds()),
			fmt.Sprintf("%d", comp.MaxRenderTime.Nanoseconds()),
			fmt.Sprintf("%d", comp.MemoryUsage),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// ExportAll exports the report to all three formats (HTML, JSON, CSV).
//
// This is a convenience method that calls ExportHTML, ExportJSON, and ExportCSV.
// If any export fails, the error is returned immediately.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	err := e.ExportAll(report, "report.html", "report.json", "report.csv")
func (e *Exporter) ExportAll(report *Report, htmlFile, jsonFile, csvFile string) error {
	if err := e.ExportHTML(report, htmlFile); err != nil {
		return fmt.Errorf("HTML export failed: %w", err)
	}

	if err := e.ExportJSON(report, jsonFile); err != nil {
		return fmt.Errorf("JSON export failed: %w", err)
	}

	if err := e.ExportCSV(report, csvFile); err != nil {
		return fmt.Errorf("CSV export failed: %w", err)
	}

	return nil
}

// ExportToString exports the report to a string in the specified format.
//
// Supported formats: FormatHTML, FormatJSON, FormatCSV.
// Returns an error for unsupported formats.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	html, err := e.ExportToString(report, FormatHTML)
//	json, err := e.ExportToString(report, FormatJSON)
//	csv, err := e.ExportToString(report, FormatCSV)
func (e *Exporter) ExportToString(report *Report, format ExportFormat) (string, error) {
	// Use empty report if nil
	if report == nil {
		report = createEmptyReport()
	}

	switch format {
	case FormatHTML:
		return e.exportHTMLToString(report)
	case FormatJSON:
		return e.exportJSONToString(report)
	case FormatCSV:
		return e.exportCSVToString(report)
	default:
		return "", fmt.Errorf("unsupported export format: %s", format)
	}
}

// GetSupportedFormats returns a list of supported export formats.
//
// Example:
//
//	formats := e.GetSupportedFormats()
//	// Returns: []ExportFormat{FormatHTML, FormatJSON, FormatCSV}
func (e *Exporter) GetSupportedFormats() []ExportFormat {
	return []ExportFormat{FormatHTML, FormatJSON, FormatCSV}
}

// exportHTMLToString generates HTML as a string.
func (e *Exporter) exportHTMLToString(report *Report) (string, error) {
	e.mu.RLock()
	rg := e.reportGenerator
	e.mu.RUnlock()

	return rg.GenerateHTML(report)
}

// exportJSONToString generates JSON as a string.
func (e *Exporter) exportJSONToString(report *Report) (string, error) {
	jsonReport := reportToJSON(report)

	data, err := json.MarshalIndent(jsonReport, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(data), nil
}

// exportCSVToString generates CSV as a string.
func (e *Exporter) exportCSVToString(report *Report) (string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"component_id",
		"component_name",
		"render_count",
		"avg_render_time_ns",
		"max_render_time_ns",
		"memory_usage",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write component data
	for _, comp := range report.Components {
		if comp == nil {
			continue
		}
		row := []string{
			comp.ComponentID,
			comp.ComponentName,
			fmt.Sprintf("%d", comp.RenderCount),
			fmt.Sprintf("%d", comp.AvgRenderTime.Nanoseconds()),
			fmt.Sprintf("%d", comp.MaxRenderTime.Nanoseconds()),
			fmt.Sprintf("%d", comp.MemoryUsage),
		}
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.String(), nil
}

// createEmptyReport creates an empty report with default values.
func createEmptyReport() *Report {
	return &Report{
		Summary:         &Summary{},
		Components:      make([]*ComponentMetrics, 0),
		Bottlenecks:     make([]*BottleneckInfo, 0),
		CPUProfile:      &CPUProfileData{},
		MemProfile:      &MemProfileData{},
		Recommendations: make([]*Recommendation, 0),
		Timestamp:       time.Now(),
	}
}

// jsonReport is the JSON-serializable representation of a Report.
type jsonReport struct {
	Summary         *jsonSummary          `json:"summary"`
	Components      []*jsonComponent      `json:"components"`
	Bottlenecks     []*jsonBottleneck     `json:"bottlenecks"`
	CPUProfile      *jsonCPUProfile       `json:"cpu_profile"`
	MemProfile      *jsonMemProfile       `json:"mem_profile"`
	Recommendations []*jsonRecommendation `json:"recommendations"`
	Timestamp       string                `json:"timestamp"`
}

// jsonSummary is the JSON-serializable representation of Summary.
type jsonSummary struct {
	DurationNs      int64   `json:"duration_ns"`
	DurationStr     string  `json:"duration_str"`
	TotalOperations int64   `json:"total_operations"`
	AverageFPS      float64 `json:"average_fps"`
	MemoryUsage     uint64  `json:"memory_usage"`
	GoroutineCount  int     `json:"goroutine_count"`
}

// jsonComponent is the JSON-serializable representation of ComponentMetrics.
type jsonComponent struct {
	ComponentID       string `json:"component_id"`
	ComponentName     string `json:"component_name"`
	RenderCount       int64  `json:"render_count"`
	TotalRenderTimeNs int64  `json:"total_render_time_ns"`
	AvgRenderTimeNs   int64  `json:"avg_render_time_ns"`
	MaxRenderTimeNs   int64  `json:"max_render_time_ns"`
	MinRenderTimeNs   int64  `json:"min_render_time_ns"`
	MemoryUsage       uint64 `json:"memory_usage"`
}

// jsonBottleneck is the JSON-serializable representation of BottleneckInfo.
type jsonBottleneck struct {
	Type        string  `json:"type"`
	Location    string  `json:"location"`
	Severity    string  `json:"severity"`
	Impact      float64 `json:"impact"`
	Description string  `json:"description"`
	Suggestion  string  `json:"suggestion"`
}

// jsonCPUProfile is the JSON-serializable representation of CPUProfileData.
type jsonCPUProfile struct {
	HotFunctions []*jsonHotFunction  `json:"hot_functions"`
	CallGraph    map[string][]string `json:"call_graph"`
	TotalSamples int64               `json:"total_samples"`
}

// jsonHotFunction is the JSON-serializable representation of HotFunction.
type jsonHotFunction struct {
	Name    string  `json:"name"`
	Samples int64   `json:"samples"`
	Percent float64 `json:"percent"`
}

// jsonMemProfile is the JSON-serializable representation of MemProfileData.
type jsonMemProfile struct {
	HeapAlloc   uint64  `json:"heap_alloc"`
	HeapObjects uint64  `json:"heap_objects"`
	GCPausesNs  []int64 `json:"gc_pauses_ns"`
}

// jsonRecommendation is the JSON-serializable representation of Recommendation.
type jsonRecommendation struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Priority    string `json:"priority"`
	Category    string `json:"category"`
	Impact      string `json:"impact"`
}

// reportToJSON converts a Report to its JSON-serializable representation.
func reportToJSON(report *Report) *jsonReport {
	jr := &jsonReport{
		Summary:         &jsonSummary{},
		Components:      make([]*jsonComponent, 0),
		Bottlenecks:     make([]*jsonBottleneck, 0),
		CPUProfile:      &jsonCPUProfile{},
		MemProfile:      &jsonMemProfile{},
		Recommendations: make([]*jsonRecommendation, 0),
		Timestamp:       report.Timestamp.Format(time.RFC3339),
	}

	// Convert summary
	if report.Summary != nil {
		jr.Summary = &jsonSummary{
			DurationNs:      report.Summary.Duration.Nanoseconds(),
			DurationStr:     report.Summary.Duration.String(),
			TotalOperations: report.Summary.TotalOperations,
			AverageFPS:      report.Summary.AverageFPS,
			MemoryUsage:     report.Summary.MemoryUsage,
			GoroutineCount:  report.Summary.GoroutineCount,
		}
	}

	// Convert components
	for _, comp := range report.Components {
		if comp == nil {
			continue
		}
		jr.Components = append(jr.Components, &jsonComponent{
			ComponentID:       comp.ComponentID,
			ComponentName:     comp.ComponentName,
			RenderCount:       comp.RenderCount,
			TotalRenderTimeNs: comp.TotalRenderTime.Nanoseconds(),
			AvgRenderTimeNs:   comp.AvgRenderTime.Nanoseconds(),
			MaxRenderTimeNs:   comp.MaxRenderTime.Nanoseconds(),
			MinRenderTimeNs:   comp.MinRenderTime.Nanoseconds(),
			MemoryUsage:       comp.MemoryUsage,
		})
	}

	// Convert bottlenecks
	for _, bn := range report.Bottlenecks {
		if bn == nil {
			continue
		}
		jr.Bottlenecks = append(jr.Bottlenecks, &jsonBottleneck{
			Type:        string(bn.Type),
			Location:    bn.Location,
			Severity:    string(bn.Severity),
			Impact:      bn.Impact,
			Description: bn.Description,
			Suggestion:  bn.Suggestion,
		})
	}

	// Convert CPU profile
	if report.CPUProfile != nil {
		jr.CPUProfile = &jsonCPUProfile{
			HotFunctions: make([]*jsonHotFunction, 0),
			CallGraph:    report.CPUProfile.CallGraph,
			TotalSamples: report.CPUProfile.TotalSamples,
		}
		for _, hf := range report.CPUProfile.HotFunctions {
			if hf == nil {
				continue
			}
			jr.CPUProfile.HotFunctions = append(jr.CPUProfile.HotFunctions, &jsonHotFunction{
				Name:    hf.Name,
				Samples: hf.Samples,
				Percent: hf.Percent,
			})
		}
	}

	// Convert memory profile
	if report.MemProfile != nil {
		jr.MemProfile = &jsonMemProfile{
			HeapAlloc:   report.MemProfile.HeapAlloc,
			HeapObjects: report.MemProfile.HeapObjects,
			GCPausesNs:  make([]int64, 0, len(report.MemProfile.GCPauses)),
		}
		for _, pause := range report.MemProfile.GCPauses {
			jr.MemProfile.GCPausesNs = append(jr.MemProfile.GCPausesNs, pause.Nanoseconds())
		}
	}

	// Convert recommendations
	for _, rec := range report.Recommendations {
		if rec == nil {
			continue
		}
		jr.Recommendations = append(jr.Recommendations, &jsonRecommendation{
			Title:       rec.Title,
			Description: rec.Description,
			Action:      rec.Action,
			Priority:    priorityToString(rec.Priority),
			Category:    string(rec.Category),
			Impact:      string(rec.Impact),
		})
	}

	return jr
}

// priorityToString converts Priority to a string representation.
func priorityToString(p Priority) string {
	switch p {
	case PriorityCritical:
		return "critical"
	case PriorityHigh:
		return "high"
	case PriorityMedium:
		return "medium"
	case PriorityLow:
		return "low"
	default:
		return "unknown"
	}
}
