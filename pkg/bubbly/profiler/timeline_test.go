// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTimelineGenerator(t *testing.T) {
	tests := []struct {
		name           string
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "creates generator with default dimensions",
			expectedWidth:  DefaultTimelineWidth,
			expectedHeight: DefaultTimelineHeight,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := NewTimelineGenerator()

			require.NotNil(t, tg)
			assert.Equal(t, tt.expectedWidth, tg.GetWidth())
			assert.Equal(t, tt.expectedHeight, tg.GetHeight())
		})
	}
}

func TestNewTimelineGeneratorWithDimensions(t *testing.T) {
	tests := []struct {
		name           string
		width          int
		height         int
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "custom dimensions",
			width:          1600,
			height:         600,
			expectedWidth:  1600,
			expectedHeight: 600,
		},
		{
			name:           "zero width uses default",
			width:          0,
			height:         600,
			expectedWidth:  DefaultTimelineWidth,
			expectedHeight: 600,
		},
		{
			name:           "zero height uses default",
			width:          1600,
			height:         0,
			expectedWidth:  1600,
			expectedHeight: DefaultTimelineHeight,
		},
		{
			name:           "negative dimensions use defaults",
			width:          -100,
			height:         -200,
			expectedWidth:  DefaultTimelineWidth,
			expectedHeight: DefaultTimelineHeight,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := NewTimelineGeneratorWithDimensions(tt.width, tt.height)

			require.NotNil(t, tg)
			assert.Equal(t, tt.expectedWidth, tg.GetWidth())
			assert.Equal(t, tt.expectedHeight, tg.GetHeight())
		})
	}
}

func TestTimelineGenerator_Generate(t *testing.T) {
	baseTime := time.Date(2024, 11, 29, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		events         []*TimedEvent
		expectedNil    bool
		expectedCount  int
		expectedTypes  map[EventType]int
		checkOrdering  bool
		checkTimeRange bool
	}{
		{
			name:        "nil events returns nil",
			events:      nil,
			expectedNil: true,
		},
		{
			name:        "empty events returns nil",
			events:      []*TimedEvent{},
			expectedNil: true,
		},
		{
			name: "single event",
			events: []*TimedEvent{
				{Name: "Event1", Type: EventTypeRender, StartTime: baseTime, Duration: 10 * time.Millisecond},
			},
			expectedNil:   false,
			expectedCount: 1,
			expectedTypes: map[EventType]int{EventTypeRender: 1},
		},
		{
			name: "multiple events sorted by start time",
			events: []*TimedEvent{
				{Name: "Event3", Type: EventTypeCommand, StartTime: baseTime.Add(200 * time.Millisecond), Duration: 5 * time.Millisecond},
				{Name: "Event1", Type: EventTypeRender, StartTime: baseTime, Duration: 10 * time.Millisecond},
				{Name: "Event2", Type: EventTypeUpdate, StartTime: baseTime.Add(100 * time.Millisecond), Duration: 15 * time.Millisecond},
			},
			expectedNil:   false,
			expectedCount: 3,
			expectedTypes: map[EventType]int{EventTypeRender: 1, EventTypeUpdate: 1, EventTypeCommand: 1},
			checkOrdering: true,
		},
		{
			name: "filters nil events",
			events: []*TimedEvent{
				{Name: "Event1", Type: EventTypeRender, StartTime: baseTime, Duration: 10 * time.Millisecond},
				nil,
				{Name: "Event2", Type: EventTypeUpdate, StartTime: baseTime.Add(50 * time.Millisecond), Duration: 5 * time.Millisecond},
				nil,
			},
			expectedNil:   false,
			expectedCount: 2,
			expectedTypes: map[EventType]int{EventTypeRender: 1, EventTypeUpdate: 1},
		},
		{
			name: "all nil events returns nil",
			events: []*TimedEvent{
				nil,
				nil,
				nil,
			},
			expectedNil: true,
		},
		{
			name: "calculates correct time range",
			events: []*TimedEvent{
				{Name: "Event1", Type: EventTypeRender, StartTime: baseTime, Duration: 10 * time.Millisecond},
				{Name: "Event2", Type: EventTypeUpdate, StartTime: baseTime.Add(50 * time.Millisecond), Duration: 100 * time.Millisecond},
			},
			expectedNil:    false,
			expectedCount:  2,
			checkTimeRange: true,
		},
		{
			name: "multiple events of same type",
			events: []*TimedEvent{
				{Name: "Render1", Type: EventTypeRender, StartTime: baseTime, Duration: 10 * time.Millisecond},
				{Name: "Render2", Type: EventTypeRender, StartTime: baseTime.Add(20 * time.Millisecond), Duration: 10 * time.Millisecond},
				{Name: "Render3", Type: EventTypeRender, StartTime: baseTime.Add(40 * time.Millisecond), Duration: 10 * time.Millisecond},
			},
			expectedNil:   false,
			expectedCount: 3,
			expectedTypes: map[EventType]int{EventTypeRender: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := NewTimelineGenerator()
			data := tg.Generate(tt.events)

			if tt.expectedNil {
				assert.Nil(t, data)
				return
			}

			require.NotNil(t, data)
			assert.Equal(t, tt.expectedCount, data.EventCount)
			assert.Equal(t, tt.expectedCount, len(data.Events))

			if tt.expectedTypes != nil {
				assert.Equal(t, tt.expectedTypes, data.TypeCounts)
			}

			if tt.checkOrdering {
				// Verify events are sorted by start time
				for i := 1; i < len(data.Events); i++ {
					assert.True(t, !data.Events[i].StartTime.Before(data.Events[i-1].StartTime),
						"Events should be sorted by start time")
				}
			}

			if tt.checkTimeRange {
				// Event2 ends at baseTime + 50ms + 100ms = baseTime + 150ms
				expectedDuration := 150 * time.Millisecond
				assert.Equal(t, expectedDuration, data.TotalDuration)
				assert.Equal(t, baseTime, data.StartTime)
				assert.Equal(t, baseTime.Add(expectedDuration), data.EndTime)
			}
		})
	}
}

func TestTimelineGenerator_GenerateHTML(t *testing.T) {
	baseTime := time.Date(2024, 11, 29, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name             string
		events           []*TimedEvent
		checkContains    []string
		checkNotContains []string
	}{
		{
			name:   "nil events shows empty message",
			events: nil,
			checkContains: []string{
				"<!DOCTYPE html>",
				"Performance Timeline",
				"No Events",
				"No timeline events to display",
			},
		},
		{
			name:   "empty events shows empty message",
			events: []*TimedEvent{},
			checkContains: []string{
				"No Events",
			},
		},
		{
			name: "single event generates valid HTML",
			events: []*TimedEvent{
				{Name: "Component.Render", Type: EventTypeRender, StartTime: baseTime, Duration: 10 * time.Millisecond},
			},
			checkContains: []string{
				"<!DOCTYPE html>",
				"<html>",
				"</html>",
				"<svg",
				"</svg>",
				"Component.Render",
				"render",
				"Total Events",
				"1",
			},
			checkNotContains: []string{
				"No Events",
			},
		},
		{
			name: "multiple event types show in legend",
			events: []*TimedEvent{
				{Name: "Render", Type: EventTypeRender, StartTime: baseTime, Duration: 10 * time.Millisecond},
				{Name: "Update", Type: EventTypeUpdate, StartTime: baseTime.Add(20 * time.Millisecond), Duration: 5 * time.Millisecond},
				{Name: "Lifecycle", Type: EventTypeLifecycle, StartTime: baseTime.Add(30 * time.Millisecond), Duration: 3 * time.Millisecond},
			},
			checkContains: []string{
				"legend",
				"#4CAF50", // Render color
				"#2196F3", // Update color
				"#9C27B0", // Lifecycle color
			},
		},
		{
			name: "escapes HTML special characters",
			events: []*TimedEvent{
				{Name: "<script>alert('xss')</script>", Type: EventTypeCustom, StartTime: baseTime, Duration: 10 * time.Millisecond},
			},
			checkContains: []string{
				"&lt;script&gt;",
			},
			checkNotContains: []string{
				"<script>alert",
			},
		},
		{
			name: "includes duration formatting",
			events: []*TimedEvent{
				{Name: "FastEvent", Type: EventTypeRender, StartTime: baseTime, Duration: 500 * time.Microsecond},
				{Name: "SlowEvent", Type: EventTypeUpdate, StartTime: baseTime.Add(10 * time.Millisecond), Duration: 2 * time.Second},
			},
			checkContains: []string{
				"Duration",
			},
		},
		{
			name: "all event types have colors",
			events: []*TimedEvent{
				{Name: "E1", Type: EventTypeRender, StartTime: baseTime, Duration: 1 * time.Millisecond},
				{Name: "E2", Type: EventTypeUpdate, StartTime: baseTime.Add(2 * time.Millisecond), Duration: 1 * time.Millisecond},
				{Name: "E3", Type: EventTypeLifecycle, StartTime: baseTime.Add(4 * time.Millisecond), Duration: 1 * time.Millisecond},
				{Name: "E4", Type: EventTypeEvent, StartTime: baseTime.Add(6 * time.Millisecond), Duration: 1 * time.Millisecond},
				{Name: "E5", Type: EventTypeCommand, StartTime: baseTime.Add(8 * time.Millisecond), Duration: 1 * time.Millisecond},
				{Name: "E6", Type: EventTypeCustom, StartTime: baseTime.Add(10 * time.Millisecond), Duration: 1 * time.Millisecond},
			},
			checkContains: []string{
				"#4CAF50", // Render
				"#2196F3", // Update
				"#9C27B0", // Lifecycle
				"#FF9800", // Event
				"#F44336", // Command
				"#607D8B", // Custom
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := NewTimelineGenerator()
			html := tg.GenerateHTML(tt.events)

			require.NotEmpty(t, html)

			for _, expected := range tt.checkContains {
				assert.Contains(t, html, expected, "HTML should contain: %s", expected)
			}

			for _, notExpected := range tt.checkNotContains {
				assert.NotContains(t, html, notExpected, "HTML should not contain: %s", notExpected)
			}
		})
	}
}

func TestTimelineGenerator_SetDimensions(t *testing.T) {
	tests := []struct {
		name           string
		width          int
		height         int
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "set valid dimensions",
			width:          1600,
			height:         800,
			expectedWidth:  1600,
			expectedHeight: 800,
		},
		{
			name:           "zero width ignored",
			width:          0,
			height:         800,
			expectedWidth:  DefaultTimelineWidth,
			expectedHeight: 800,
		},
		{
			name:           "zero height ignored",
			width:          1600,
			height:         0,
			expectedWidth:  1600,
			expectedHeight: DefaultTimelineHeight,
		},
		{
			name:           "negative dimensions ignored",
			width:          -100,
			height:         -200,
			expectedWidth:  DefaultTimelineWidth,
			expectedHeight: DefaultTimelineHeight,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := NewTimelineGenerator()
			tg.SetDimensions(tt.width, tt.height)

			assert.Equal(t, tt.expectedWidth, tg.GetWidth())
			assert.Equal(t, tt.expectedHeight, tg.GetHeight())
		})
	}
}

func TestTimelineGenerator_Reset(t *testing.T) {
	tg := NewTimelineGeneratorWithDimensions(1600, 800)

	assert.Equal(t, 1600, tg.GetWidth())
	assert.Equal(t, 800, tg.GetHeight())

	tg.Reset()

	assert.Equal(t, DefaultTimelineWidth, tg.GetWidth())
	assert.Equal(t, DefaultTimelineHeight, tg.GetHeight())
}

func TestTimelineGenerator_ThreadSafety(t *testing.T) {
	tg := NewTimelineGenerator()
	baseTime := time.Now()

	events := make([]*TimedEvent, 100)
	for i := 0; i < 100; i++ {
		events[i] = &TimedEvent{
			Name:      "Event",
			Type:      EventTypeRender,
			StartTime: baseTime.Add(time.Duration(i) * time.Millisecond),
			Duration:  time.Millisecond,
		}
	}

	var wg sync.WaitGroup
	numGoroutines := 50

	// Concurrent reads and writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Mix of operations
			switch id % 5 {
			case 0:
				tg.Generate(events)
			case 1:
				tg.GenerateHTML(events)
			case 2:
				tg.GetWidth()
				tg.GetHeight()
			case 3:
				tg.SetDimensions(1200+id, 400+id)
			case 4:
				tg.Reset()
			}
		}(i)
	}

	wg.Wait()
	// If we get here without deadlock or panic, test passes
}

func TestTimedEvent_Methods(t *testing.T) {
	baseTime := time.Date(2024, 11, 29, 10, 0, 0, 0, time.UTC)

	t.Run("GetEndTime", func(t *testing.T) {
		event := &TimedEvent{
			Name:      "Test",
			StartTime: baseTime,
			Duration:  100 * time.Millisecond,
		}

		endTime := event.GetEndTime()
		assert.Equal(t, baseTime.Add(100*time.Millisecond), endTime)
	})

	t.Run("SetMetadata and GetMetadata", func(t *testing.T) {
		event := &TimedEvent{Name: "Test"}

		// Initially nil metadata
		assert.Equal(t, "", event.GetMetadata("key"))

		// Set metadata
		event.SetMetadata("key1", "value1")
		event.SetMetadata("key2", "value2")

		assert.Equal(t, "value1", event.GetMetadata("key1"))
		assert.Equal(t, "value2", event.GetMetadata("key2"))
		assert.Equal(t, "", event.GetMetadata("nonexistent"))
	})

	t.Run("GetMetadata with nil map", func(t *testing.T) {
		event := &TimedEvent{Name: "Test", Metadata: nil}
		assert.Equal(t, "", event.GetMetadata("key"))
	})
}

func TestAddEvent(t *testing.T) {
	baseTime := time.Date(2024, 11, 29, 10, 0, 0, 0, time.UTC)

	event := AddEvent("TestEvent", EventTypeRender, baseTime, 50*time.Millisecond)

	require.NotNil(t, event)
	assert.Equal(t, "TestEvent", event.Name)
	assert.Equal(t, EventTypeRender, event.Type)
	assert.Equal(t, baseTime, event.StartTime)
	assert.Equal(t, 50*time.Millisecond, event.Duration)
	assert.NotNil(t, event.Metadata)
	assert.Empty(t, event.ComponentID)
}

func TestAddEventWithComponent(t *testing.T) {
	baseTime := time.Date(2024, 11, 29, 10, 0, 0, 0, time.UTC)

	event := AddEventWithComponent("TestEvent", EventTypeUpdate, baseTime, 25*time.Millisecond, "comp-123")

	require.NotNil(t, event)
	assert.Equal(t, "TestEvent", event.Name)
	assert.Equal(t, EventTypeUpdate, event.Type)
	assert.Equal(t, baseTime, event.StartTime)
	assert.Equal(t, 25*time.Millisecond, event.Duration)
	assert.Equal(t, "comp-123", event.ComponentID)
	assert.NotNil(t, event.Metadata)
}

func TestGetEventColor(t *testing.T) {
	tests := []struct {
		eventType     EventType
		expectedColor string
	}{
		{EventTypeRender, "#4CAF50"},
		{EventTypeUpdate, "#2196F3"},
		{EventTypeLifecycle, "#9C27B0"},
		{EventTypeEvent, "#FF9800"},
		{EventTypeCommand, "#F44336"},
		{EventTypeCustom, "#607D8B"},
		{EventType("unknown"), "#9E9E9E"},
	}

	for _, tt := range tests {
		t.Run(string(tt.eventType), func(t *testing.T) {
			color := getEventColor(tt.eventType)
			assert.Equal(t, tt.expectedColor, color)
		})
	}
}

func TestFormatTimelineDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{500 * time.Nanosecond, "500ns"},
		{500 * time.Microsecond, "500.0μs"},
		{1500 * time.Microsecond, "1.5ms"},
		{500 * time.Millisecond, "500.0ms"},
		{1500 * time.Millisecond, "1.50s"},
		{90 * time.Second, "1.5m"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatTimelineDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateTimelineLabel(t *testing.T) {
	tests := []struct {
		name     string
		label    string
		maxWidth int
		expected string
	}{
		{
			name:     "empty label",
			label:    "",
			maxWidth: 100,
			expected: "",
		},
		{
			name:     "label fits",
			label:    "Short",
			maxWidth: 100,
			expected: "Short",
		},
		{
			name:     "label truncated",
			label:    "VeryLongLabelThatNeedsTruncation",
			maxWidth: 70, // ~10 chars
			expected: "VeryLon...",
		},
		{
			name:     "width too small",
			label:    "Test",
			maxWidth: 20, // ~2 chars, too small
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateTimelineLabel(tt.label, tt.maxWidth)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal text", "normal text"},
		{"<script>", "&lt;script&gt;"},
		{"a & b", "a &amp; b"},
		{`"quoted"`, "&quot;quoted&quot;"},
		{"it's", "it&#39;s"},
		{"<a href=\"test\">link</a>", "&lt;a href=&quot;test&quot;&gt;link&lt;/a&gt;"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := escapeHTML(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimelineData_Fields(t *testing.T) {
	baseTime := time.Date(2024, 11, 29, 10, 0, 0, 0, time.UTC)

	events := []*TimedEvent{
		{Name: "E1", Type: EventTypeRender, StartTime: baseTime, Duration: 10 * time.Millisecond},
		{Name: "E2", Type: EventTypeUpdate, StartTime: baseTime.Add(20 * time.Millisecond), Duration: 30 * time.Millisecond},
	}

	tg := NewTimelineGenerator()
	data := tg.Generate(events)

	require.NotNil(t, data)

	// Verify all fields
	assert.Equal(t, 2, data.EventCount)
	assert.Equal(t, 2, len(data.Events))
	assert.Equal(t, baseTime, data.StartTime)
	assert.Equal(t, baseTime.Add(50*time.Millisecond), data.EndTime) // E2 ends at 20+30=50ms
	assert.Equal(t, 50*time.Millisecond, data.TotalDuration)
	assert.Equal(t, 1, data.TypeCounts[EventTypeRender])
	assert.Equal(t, 1, data.TypeCounts[EventTypeUpdate])
}

func TestTimelineGenerator_GenerateHTML_LargeEventCount(t *testing.T) {
	baseTime := time.Now()
	events := make([]*TimedEvent, 100)

	for i := 0; i < 100; i++ {
		events[i] = &TimedEvent{
			Name:      "Event",
			Type:      EventType([]EventType{EventTypeRender, EventTypeUpdate, EventTypeLifecycle}[i%3]),
			StartTime: baseTime.Add(time.Duration(i) * time.Millisecond),
			Duration:  time.Millisecond,
		}
	}

	tg := NewTimelineGenerator()
	html := tg.GenerateHTML(events)

	require.NotEmpty(t, html)
	assert.Contains(t, html, "100") // Total events count
	assert.Contains(t, html, "<svg")
	assert.Contains(t, html, "</svg>")
}

func TestTimelineGenerator_GenerateHTML_ZeroDuration(t *testing.T) {
	baseTime := time.Now()

	// All events at same time with zero duration
	events := []*TimedEvent{
		{Name: "E1", Type: EventTypeRender, StartTime: baseTime, Duration: 0},
		{Name: "E2", Type: EventTypeUpdate, StartTime: baseTime, Duration: 0},
	}

	tg := NewTimelineGenerator()
	html := tg.GenerateHTML(events)

	require.NotEmpty(t, html)
	// Should still render without errors
	assert.Contains(t, html, "E1")
	assert.Contains(t, html, "E2")
}

func TestEventTypes(t *testing.T) {
	// Verify all event type constants
	assert.Equal(t, EventType("render"), EventTypeRender)
	assert.Equal(t, EventType("update"), EventTypeUpdate)
	assert.Equal(t, EventType("lifecycle"), EventTypeLifecycle)
	assert.Equal(t, EventType("event"), EventTypeEvent)
	assert.Equal(t, EventType("command"), EventTypeCommand)
	assert.Equal(t, EventType("custom"), EventTypeCustom)
}

func TestTimelineGenerator_HTMLStructure(t *testing.T) {
	baseTime := time.Now()
	events := []*TimedEvent{
		{Name: "Test", Type: EventTypeRender, StartTime: baseTime, Duration: 10 * time.Millisecond},
	}

	tg := NewTimelineGenerator()
	html := tg.GenerateHTML(events)

	// Verify HTML structure
	assert.True(t, strings.HasPrefix(html, "<!DOCTYPE html>"))
	assert.Contains(t, html, "<html>")
	assert.Contains(t, html, "</html>")
	assert.Contains(t, html, "<head>")
	assert.Contains(t, html, "</head>")
	assert.Contains(t, html, "<body>")
	assert.Contains(t, html, "</body>")
	assert.Contains(t, html, "<style>")
	assert.Contains(t, html, "</style>")
}

func TestTimelineGenerator_SVGStructure(t *testing.T) {
	baseTime := time.Now()
	events := []*TimedEvent{
		{Name: "Test", Type: EventTypeRender, StartTime: baseTime, Duration: 10 * time.Millisecond},
	}

	tg := NewTimelineGenerator()
	html := tg.GenerateHTML(events)

	// Verify SVG structure
	assert.Contains(t, html, "<svg")
	assert.Contains(t, html, "</svg>")
	assert.Contains(t, html, "viewBox")
	assert.Contains(t, html, "<rect") // Event bars
	assert.Contains(t, html, "<text") // Labels
	assert.Contains(t, html, "<line") // Axis
}

func TestTimelineGenerator_TooltipContent(t *testing.T) {
	baseTime := time.Date(2024, 11, 29, 10, 30, 45, 0, time.UTC)
	events := []*TimedEvent{
		{Name: "TestComponent.Render", Type: EventTypeRender, StartTime: baseTime, Duration: 15 * time.Millisecond},
	}

	tg := NewTimelineGenerator()
	html := tg.GenerateHTML(events)

	// Verify tooltip content
	assert.Contains(t, html, "<title>")
	assert.Contains(t, html, "TestComponent.Render")
	assert.Contains(t, html, "Type: render")
	assert.Contains(t, html, "Duration:")
	assert.Contains(t, html, "Start:")
}

func BenchmarkTimelineGenerator_Generate(b *testing.B) {
	baseTime := time.Now()
	events := make([]*TimedEvent, 1000)
	for i := 0; i < 1000; i++ {
		events[i] = &TimedEvent{
			Name:      "Event",
			Type:      EventTypeRender,
			StartTime: baseTime.Add(time.Duration(i) * time.Millisecond),
			Duration:  time.Millisecond,
		}
	}

	tg := NewTimelineGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tg.Generate(events)
	}
}

func BenchmarkTimelineGenerator_GenerateHTML(b *testing.B) {
	baseTime := time.Now()
	events := make([]*TimedEvent, 100)
	for i := 0; i < 100; i++ {
		events[i] = &TimedEvent{
			Name:      "Event",
			Type:      EventTypeRender,
			StartTime: baseTime.Add(time.Duration(i) * time.Millisecond),
			Duration:  time.Millisecond,
		}
	}

	tg := NewTimelineGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tg.GenerateHTML(events)
	}
}
