// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Default dimensions for timeline visualization.
const (
	// DefaultTimelineWidth is the default width in pixels.
	DefaultTimelineWidth = 1200

	// DefaultTimelineHeight is the default height in pixels.
	DefaultTimelineHeight = 400

	// timelineRowHeight is the height of each event row in pixels.
	timelineRowHeight = 30

	// timelineHeaderHeight is the height of the header area.
	timelineHeaderHeight = 40

	// timelineMargin is the margin around the timeline.
	timelineMargin = 20

	// timelineLabelWidth is the width reserved for event labels.
	timelineLabelWidth = 150

	// minEventWidth is the minimum width for an event bar.
	minEventWidth = 2
)

// EventType categorizes timeline events.
type EventType string

const (
	// EventTypeRender represents a render operation.
	EventTypeRender EventType = "render"

	// EventTypeUpdate represents an update operation.
	EventTypeUpdate EventType = "update"

	// EventTypeLifecycle represents a lifecycle hook.
	EventTypeLifecycle EventType = "lifecycle"

	// EventTypeEvent represents an event handler.
	EventTypeEvent EventType = "event"

	// EventTypeCommand represents a command execution.
	EventTypeCommand EventType = "command"

	// EventTypeCustom represents a custom event.
	EventTypeCustom EventType = "custom"
)

// TimedEvent represents an event with timing information for timeline visualization.
//
// Each event has a start time, duration, and metadata for display.
//
// Example:
//
//	event := &TimedEvent{
//	    Name:      "Component.Render",
//	    Type:      EventTypeRender,
//	    StartTime: time.Now(),
//	    Duration:  5 * time.Millisecond,
//	}
type TimedEvent struct {
	// Name is the event name (e.g., "Component.Render")
	Name string

	// Type categorizes the event
	Type EventType

	// StartTime is when the event started
	StartTime time.Time

	// Duration is how long the event took
	Duration time.Duration

	// ComponentID is the optional component identifier
	ComponentID string

	// Metadata contains additional event data
	Metadata map[string]string
}

// TimelineData contains processed timeline data for visualization.
//
// It includes sorted events, time range, and computed layout information.
type TimelineData struct {
	// Events is the sorted list of events
	Events []*TimedEvent

	// StartTime is the earliest event start time
	StartTime time.Time

	// EndTime is the latest event end time
	EndTime time.Time

	// TotalDuration is the timeline span
	TotalDuration time.Duration

	// EventCount is the number of events
	EventCount int

	// TypeCounts maps event types to their counts
	TypeCounts map[EventType]int
}

// TimelineGenerator generates timeline visualizations from timed events.
//
// Timelines show events over time, with each event displayed as a bar
// proportional to its duration. Events are color-coded by type.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	tg := NewTimelineGenerator()
//	data := tg.Generate(events)
//	html := tg.GenerateHTML(events)
type TimelineGenerator struct {
	// width is the visualization width in pixels
	width int

	// height is the visualization height in pixels
	height int

	// mu protects concurrent access to generator state
	mu sync.RWMutex
}

// NewTimelineGenerator creates a new TimelineGenerator with default dimensions.
//
// Default dimensions are 1200x400 pixels.
//
// Example:
//
//	tg := NewTimelineGenerator()
//	html := tg.GenerateHTML(events)
func NewTimelineGenerator() *TimelineGenerator {
	return &TimelineGenerator{
		width:  DefaultTimelineWidth,
		height: DefaultTimelineHeight,
	}
}

// NewTimelineGeneratorWithDimensions creates a TimelineGenerator with custom dimensions.
//
// If width or height is <= 0, the default value is used.
//
// Example:
//
//	tg := NewTimelineGeneratorWithDimensions(1600, 600)
func NewTimelineGeneratorWithDimensions(width, height int) *TimelineGenerator {
	tg := &TimelineGenerator{
		width:  DefaultTimelineWidth,
		height: DefaultTimelineHeight,
	}

	if width > 0 {
		tg.width = width
	}
	if height > 0 {
		tg.height = height
	}

	return tg
}

// Generate processes events and returns structured timeline data.
//
// Events are sorted by start time, and time range is calculated.
// Returns nil if events slice is nil or empty.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	data := tg.Generate(events)
//	if data != nil {
//	    fmt.Printf("Timeline spans %v\n", data.TotalDuration)
//	}
func (tg *TimelineGenerator) Generate(events []*TimedEvent) *TimelineData {
	if len(events) == 0 {
		return nil
	}

	// Filter out nil events
	validEvents := make([]*TimedEvent, 0, len(events))
	for _, e := range events {
		if e != nil {
			validEvents = append(validEvents, e)
		}
	}

	if len(validEvents) == 0 {
		return nil
	}

	// Sort events by start time
	sortedEvents := make([]*TimedEvent, len(validEvents))
	copy(sortedEvents, validEvents)
	sort.Slice(sortedEvents, func(i, j int) bool {
		return sortedEvents[i].StartTime.Before(sortedEvents[j].StartTime)
	})

	// Calculate time range
	startTime := sortedEvents[0].StartTime
	endTime := sortedEvents[0].StartTime.Add(sortedEvents[0].Duration)

	typeCounts := make(map[EventType]int)

	for _, e := range sortedEvents {
		if e.StartTime.Before(startTime) {
			startTime = e.StartTime
		}
		eventEnd := e.StartTime.Add(e.Duration)
		if eventEnd.After(endTime) {
			endTime = eventEnd
		}
		typeCounts[e.Type]++
	}

	return &TimelineData{
		Events:        sortedEvents,
		StartTime:     startTime,
		EndTime:       endTime,
		TotalDuration: endTime.Sub(startTime),
		EventCount:    len(sortedEvents),
		TypeCounts:    typeCounts,
	}
}

// GenerateHTML generates an HTML timeline visualization.
//
// Returns an HTML string containing an SVG timeline that can be embedded
// in a web page or saved to a file. If events is nil or empty, returns
// an HTML page with a "No events" message.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	html := tg.GenerateHTML(events)
//	os.WriteFile("timeline.html", []byte(html), 0644)
func (tg *TimelineGenerator) GenerateHTML(events []*TimedEvent) string {
	tg.mu.RLock()
	width := tg.width
	height := tg.height
	tg.mu.RUnlock()

	var html strings.Builder

	// HTML header
	html.WriteString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Performance Timeline</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background: #f5f5f5;
        }
        .timeline-container {
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            padding: 20px;
            overflow-x: auto;
        }
        .timeline-header {
            margin-bottom: 20px;
        }
        .timeline-header h1 {
            margin: 0 0 10px 0;
            color: #333;
            font-size: 24px;
        }
        .timeline-stats {
            display: flex;
            gap: 20px;
            flex-wrap: wrap;
        }
        .stat {
            background: #f8f9fa;
            padding: 10px 15px;
            border-radius: 4px;
        }
        .stat-label {
            font-size: 12px;
            color: #666;
            text-transform: uppercase;
        }
        .stat-value {
            font-size: 18px;
            font-weight: bold;
            color: #333;
        }
        .legend {
            display: flex;
            gap: 15px;
            margin-top: 15px;
            flex-wrap: wrap;
        }
        .legend-item {
            display: flex;
            align-items: center;
            gap: 5px;
            font-size: 12px;
        }
        .legend-color {
            width: 16px;
            height: 16px;
            border-radius: 3px;
        }
        .event-bar:hover {
            opacity: 0.8;
            cursor: pointer;
        }
        .event-label {
            font-family: monospace;
            font-size: 11px;
        }
        .time-marker {
            font-family: monospace;
            font-size: 10px;
            fill: #666;
        }
        .empty-message {
            text-align: center;
            padding: 40px;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="timeline-container">
`)

	// Generate timeline data
	data := tg.Generate(events)

	if data == nil {
		html.WriteString(`        <div class="empty-message">
            <h2>No Events</h2>
            <p>No timeline events to display.</p>
        </div>
    </div>
</body>
</html>`)
		return html.String()
	}

	// Header with stats
	html.WriteString(`        <div class="timeline-header">
            <h1>Performance Timeline</h1>
            <div class="timeline-stats">
`)
	html.WriteString(fmt.Sprintf(`                <div class="stat">
                    <div class="stat-label">Total Events</div>
                    <div class="stat-value">%d</div>
                </div>
`, data.EventCount))
	html.WriteString(fmt.Sprintf(`                <div class="stat">
                    <div class="stat-label">Duration</div>
                    <div class="stat-value">%s</div>
                </div>
`, formatTimelineDuration(data.TotalDuration)))

	// Type counts
	for eventType, count := range data.TypeCounts {
		html.WriteString(fmt.Sprintf(`                <div class="stat">
                    <div class="stat-label">%s</div>
                    <div class="stat-value">%d</div>
                </div>
`, escapeHTML(string(eventType)), count))
	}

	html.WriteString(`            </div>
            <div class="legend">
`)

	// Legend
	legendTypes := []EventType{EventTypeRender, EventTypeUpdate, EventTypeLifecycle, EventTypeEvent, EventTypeCommand, EventTypeCustom}
	for _, t := range legendTypes {
		if data.TypeCounts[t] > 0 {
			html.WriteString(fmt.Sprintf(`                <div class="legend-item">
                    <div class="legend-color" style="background: %s;"></div>
                    <span>%s</span>
                </div>
`, getEventColor(t), escapeHTML(string(t))))
		}
	}

	html.WriteString(`            </div>
        </div>
`)

	// Calculate dimensions
	svgWidth := width
	svgHeight := height
	if len(data.Events)*timelineRowHeight+timelineHeaderHeight > svgHeight {
		svgHeight = len(data.Events)*timelineRowHeight + timelineHeaderHeight + timelineMargin*2
	}

	// SVG timeline
	html.WriteString(fmt.Sprintf(`        <svg width="%d" height="%d" viewBox="0 0 %d %d">
`, svgWidth, svgHeight, svgWidth, svgHeight))

	// Background
	html.WriteString(fmt.Sprintf(`            <rect x="0" y="0" width="%d" height="%d" fill="#fafafa"/>
`, svgWidth, svgHeight))

	// Time axis
	timelineStart := timelineMargin + timelineLabelWidth
	timelineWidth := svgWidth - timelineStart - timelineMargin
	tg.renderTimeAxis(&html, timelineStart, timelineWidth, timelineHeaderHeight-10, data)

	// Events
	y := timelineHeaderHeight
	for _, event := range data.Events {
		tg.renderEvent(&html, event, timelineStart, y, timelineWidth, data)
		y += timelineRowHeight
	}

	html.WriteString(`        </svg>
    </div>
</body>
</html>`)

	return html.String()
}

// renderTimeAxis renders the time axis with markers.
func (tg *TimelineGenerator) renderTimeAxis(html *strings.Builder, x, width, y int, data *TimelineData) {
	// Axis line
	html.WriteString(fmt.Sprintf(`            <line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#ccc" stroke-width="1"/>
`, x, y, x+width, y))

	// Time markers (5 markers)
	numMarkers := 5
	for i := 0; i <= numMarkers; i++ {
		markerX := x + (width * i / numMarkers)
		markerTime := data.TotalDuration * time.Duration(i) / time.Duration(numMarkers)

		// Tick mark
		html.WriteString(fmt.Sprintf(`            <line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#ccc" stroke-width="1"/>
`, markerX, y-5, markerX, y+5))

		// Time label
		html.WriteString(fmt.Sprintf(`            <text x="%d" y="%d" class="time-marker" text-anchor="middle">%s</text>
`, markerX, y-10, formatTimelineDuration(markerTime)))
	}
}

// renderEvent renders a single event bar.
func (tg *TimelineGenerator) renderEvent(html *strings.Builder, event *TimedEvent, timelineX, y, timelineWidth int, data *TimelineData) {
	// Event label
	labelX := timelineMargin
	labelY := y + timelineRowHeight/2 + 4
	label := truncateTimelineLabel(event.Name, timelineLabelWidth)
	html.WriteString(fmt.Sprintf(`            <text x="%d" y="%d" class="event-label">%s</text>
`, labelX, labelY, escapeHTML(label)))

	// Calculate bar position and width
	var eventX, eventWidth int
	if data.TotalDuration > 0 {
		offset := event.StartTime.Sub(data.StartTime)
		eventX = timelineX + int(float64(timelineWidth)*float64(offset)/float64(data.TotalDuration))
		eventWidth = int(float64(timelineWidth) * float64(event.Duration) / float64(data.TotalDuration))
	} else {
		eventX = timelineX
		eventWidth = timelineWidth
	}

	if eventWidth < minEventWidth {
		eventWidth = minEventWidth
	}

	// Event bar
	barY := y + 5
	barHeight := timelineRowHeight - 10
	color := getEventColor(event.Type)

	html.WriteString(fmt.Sprintf(`            <rect x="%d" y="%d" width="%d" height="%d" fill="%s" class="event-bar" rx="2">
`, eventX, barY, eventWidth, barHeight, color))
	html.WriteString(fmt.Sprintf(`                <title>%s
Type: %s
Duration: %s
Start: %s</title>
`, escapeHTML(event.Name), escapeHTML(string(event.Type)), formatTimelineDuration(event.Duration), event.StartTime.Format("15:04:05.000")))
	html.WriteString(`            </rect>
`)
}

// GetWidth returns the current visualization width.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (tg *TimelineGenerator) GetWidth() int {
	tg.mu.RLock()
	defer tg.mu.RUnlock()
	return tg.width
}

// GetHeight returns the current visualization height.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (tg *TimelineGenerator) GetHeight() int {
	tg.mu.RLock()
	defer tg.mu.RUnlock()
	return tg.height
}

// SetDimensions sets the visualization dimensions.
//
// Invalid dimensions (<= 0) are ignored.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (tg *TimelineGenerator) SetDimensions(width, height int) {
	tg.mu.Lock()
	defer tg.mu.Unlock()

	if width > 0 {
		tg.width = width
	}
	if height > 0 {
		tg.height = height
	}
}

// Reset resets the generator to default dimensions.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (tg *TimelineGenerator) Reset() {
	tg.mu.Lock()
	defer tg.mu.Unlock()
	tg.width = DefaultTimelineWidth
	tg.height = DefaultTimelineHeight
}

// getEventColor returns the color for an event type.
func getEventColor(eventType EventType) string {
	switch eventType {
	case EventTypeRender:
		return "#4CAF50" // Green
	case EventTypeUpdate:
		return "#2196F3" // Blue
	case EventTypeLifecycle:
		return "#9C27B0" // Purple
	case EventTypeEvent:
		return "#FF9800" // Orange
	case EventTypeCommand:
		return "#F44336" // Red
	case EventTypeCustom:
		return "#607D8B" // Blue-grey
	default:
		return "#9E9E9E" // Grey
	}
}

// formatTimelineDuration formats a duration for timeline display.
func formatTimelineDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.1fμs", float64(d.Nanoseconds())/1000)
	}
	if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Nanoseconds())/1000000)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}

// truncateTimelineLabel truncates a label to fit within a given width.
func truncateTimelineLabel(label string, maxWidth int) string {
	if label == "" {
		return ""
	}

	// Approximate characters that fit (assuming ~7px per char)
	maxChars := maxWidth / 7
	if maxChars <= 3 {
		return ""
	}

	if len(label) <= maxChars {
		return label
	}

	return label[:maxChars-3] + "..."
}

// escapeHTML escapes special HTML characters in a string.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// AddEvent creates a new TimedEvent with the given parameters.
//
// This is a convenience function for creating events.
//
// Example:
//
//	event := AddEvent("Component.Render", EventTypeRender, startTime, duration)
func AddEvent(name string, eventType EventType, startTime time.Time, duration time.Duration) *TimedEvent {
	return &TimedEvent{
		Name:      name,
		Type:      eventType,
		StartTime: startTime,
		Duration:  duration,
		Metadata:  make(map[string]string),
	}
}

// AddEventWithComponent creates a new TimedEvent with component information.
//
// Example:
//
//	event := AddEventWithComponent("Render", EventTypeRender, startTime, duration, "comp-123")
func AddEventWithComponent(name string, eventType EventType, startTime time.Time, duration time.Duration, componentID string) *TimedEvent {
	return &TimedEvent{
		Name:        name,
		Type:        eventType,
		StartTime:   startTime,
		Duration:    duration,
		ComponentID: componentID,
		Metadata:    make(map[string]string),
	}
}

// GetEndTime returns the end time of the event.
func (e *TimedEvent) GetEndTime() time.Time {
	return e.StartTime.Add(e.Duration)
}

// SetMetadata sets a metadata key-value pair.
func (e *TimedEvent) SetMetadata(key, value string) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
}

// GetMetadata retrieves a metadata value by key.
func (e *TimedEvent) GetMetadata(key string) string {
	if e.Metadata == nil {
		return ""
	}
	return e.Metadata[key]
}
