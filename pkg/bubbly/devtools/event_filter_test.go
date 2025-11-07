package devtools

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewEventFilter tests the constructor
func TestNewEventFilter(t *testing.T) {
	filter := NewEventFilter()
	require.NotNil(t, filter)
	assert.Empty(t, filter.GetNames())
	assert.Empty(t, filter.GetSources())
	assert.Nil(t, filter.GetTimeRange())
}

// TestEventFilter_WithNames tests name filtering
func TestEventFilter_WithNames(t *testing.T) {
	tests := []struct {
		name        string
		filterNames []string
		event       EventRecord
		wantMatch   bool
	}{
		{
			name:        "matches single name",
			filterNames: []string{"click"},
			event:       EventRecord{Name: "click"},
			wantMatch:   true,
		},
		{
			name:        "matches one of multiple names",
			filterNames: []string{"click", "submit", "change"},
			event:       EventRecord{Name: "submit"},
			wantMatch:   true,
		},
		{
			name:        "does not match",
			filterNames: []string{"click", "submit"},
			event:       EventRecord{Name: "change"},
			wantMatch:   false,
		},
		{
			name:        "empty filter matches all",
			filterNames: []string{},
			event:       EventRecord{Name: "anything"},
			wantMatch:   true,
		},
		{
			name:        "case insensitive match",
			filterNames: []string{"CLICK"},
			event:       EventRecord{Name: "click"},
			wantMatch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewEventFilter().WithNames(tt.filterNames...)
			assert.Equal(t, tt.wantMatch, filter.Matches(tt.event))
		})
	}
}

// TestEventFilter_WithSources tests source filtering
func TestEventFilter_WithSources(t *testing.T) {
	tests := []struct {
		name          string
		filterSources []string
		event         EventRecord
		wantMatch     bool
	}{
		{
			name:          "matches single source",
			filterSources: []string{"button-1"},
			event:         EventRecord{SourceID: "button-1"},
			wantMatch:     true,
		},
		{
			name:          "matches one of multiple sources",
			filterSources: []string{"button-1", "form-1", "input-1"},
			event:         EventRecord{SourceID: "form-1"},
			wantMatch:     true,
		},
		{
			name:          "does not match",
			filterSources: []string{"button-1", "form-1"},
			event:         EventRecord{SourceID: "input-1"},
			wantMatch:     false,
		},
		{
			name:          "empty filter matches all",
			filterSources: []string{},
			event:         EventRecord{SourceID: "anything"},
			wantMatch:     true,
		},
		{
			name:          "case insensitive match",
			filterSources: []string{"BUTTON-1"},
			event:         EventRecord{SourceID: "button-1"},
			wantMatch:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewEventFilter().WithSources(tt.filterSources...)
			assert.Equal(t, tt.wantMatch, filter.Matches(tt.event))
		})
	}
}

// TestEventFilter_WithTimeRange tests time range filtering
func TestEventFilter_WithTimeRange(t *testing.T) {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)
	oneHourLater := now.Add(1 * time.Hour)

	tests := []struct {
		name      string
		start     time.Time
		end       time.Time
		event     EventRecord
		wantMatch bool
	}{
		{
			name:      "within range",
			start:     twoHoursAgo,
			end:       oneHourLater,
			event:     EventRecord{Timestamp: now},
			wantMatch: true,
		},
		{
			name:      "at start boundary",
			start:     now,
			end:       oneHourLater,
			event:     EventRecord{Timestamp: now},
			wantMatch: true,
		},
		{
			name:      "at end boundary",
			start:     twoHoursAgo,
			end:       now,
			event:     EventRecord{Timestamp: now},
			wantMatch: true,
		},
		{
			name:      "before range",
			start:     oneHourAgo,
			end:       oneHourLater,
			event:     EventRecord{Timestamp: twoHoursAgo},
			wantMatch: false,
		},
		{
			name:      "after range",
			start:     twoHoursAgo,
			end:       oneHourAgo,
			event:     EventRecord{Timestamp: now},
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewEventFilter().WithTimeRange(tt.start, tt.end)
			assert.Equal(t, tt.wantMatch, filter.Matches(tt.event))
		})
	}
}

// TestEventFilter_MultipleFilters tests combining multiple filter criteria
func TestEventFilter_MultipleFilters(t *testing.T) {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	oneHourLater := now.Add(1 * time.Hour)

	tests := []struct {
		name      string
		filter    *EventFilter
		event     EventRecord
		wantMatch bool
	}{
		{
			name: "matches all criteria",
			filter: NewEventFilter().
				WithNames("click").
				WithSources("button-1").
				WithTimeRange(oneHourAgo, oneHourLater),
			event: EventRecord{
				Name:      "click",
				SourceID:  "button-1",
				Timestamp: now,
			},
			wantMatch: true,
		},
		{
			name: "fails name check",
			filter: NewEventFilter().
				WithNames("click").
				WithSources("button-1"),
			event: EventRecord{
				Name:     "submit",
				SourceID: "button-1",
			},
			wantMatch: false,
		},
		{
			name: "fails source check",
			filter: NewEventFilter().
				WithNames("click").
				WithSources("button-1"),
			event: EventRecord{
				Name:     "click",
				SourceID: "form-1",
			},
			wantMatch: false,
		},
		{
			name: "fails time range check",
			filter: NewEventFilter().
				WithNames("click").
				WithTimeRange(oneHourAgo, now),
			event: EventRecord{
				Name:      "click",
				Timestamp: oneHourLater,
			},
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantMatch, tt.filter.Matches(tt.event))
		})
	}
}

// TestEventFilter_Apply tests batch filtering
func TestEventFilter_Apply(t *testing.T) {
	now := time.Now()
	events := []EventRecord{
		{ID: "1", Name: "click", SourceID: "button-1", Timestamp: now},
		{ID: "2", Name: "submit", SourceID: "form-1", Timestamp: now},
		{ID: "3", Name: "click", SourceID: "button-2", Timestamp: now},
		{ID: "4", Name: "change", SourceID: "input-1", Timestamp: now},
	}

	tests := []struct {
		name    string
		filter  *EventFilter
		wantIDs []string
	}{
		{
			name:    "filter by name",
			filter:  NewEventFilter().WithNames("click"),
			wantIDs: []string{"1", "3"},
		},
		{
			name:    "filter by source",
			filter:  NewEventFilter().WithSources("button-1", "button-2"),
			wantIDs: []string{"1", "3"},
		},
		{
			name:    "filter by multiple criteria",
			filter:  NewEventFilter().WithNames("click").WithSources("button-1"),
			wantIDs: []string{"1"},
		},
		{
			name:    "no filter returns all",
			filter:  NewEventFilter(),
			wantIDs: []string{"1", "2", "3", "4"},
		},
		{
			name:    "no matches returns empty",
			filter:  NewEventFilter().WithNames("nonexistent"),
			wantIDs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := tt.filter.Apply(events)
			var gotIDs []string
			for _, e := range filtered {
				gotIDs = append(gotIDs, e.ID)
			}
			assert.Equal(t, tt.wantIDs, gotIDs)
		})
	}
}

// TestEventFilter_ApplyWithTimeRange tests batch filtering with time ranges
func TestEventFilter_ApplyWithTimeRange(t *testing.T) {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)

	events := []EventRecord{
		{ID: "1", Name: "click", Timestamp: twoHoursAgo},
		{ID: "2", Name: "click", Timestamp: oneHourAgo},
		{ID: "3", Name: "click", Timestamp: now},
	}

	filter := NewEventFilter().
		WithNames("click").
		WithTimeRange(oneHourAgo, now)

	filtered := filter.Apply(events)
	require.Len(t, filtered, 2)
	assert.Equal(t, "2", filtered[0].ID)
	assert.Equal(t, "3", filtered[1].ID)
}

// TestEventFilter_Clear tests clearing filter criteria
func TestEventFilter_Clear(t *testing.T) {
	filter := NewEventFilter().
		WithNames("click").
		WithSources("button-1").
		WithTimeRange(time.Now(), time.Now())

	// Should have filters
	assert.NotEmpty(t, filter.GetNames())
	assert.NotEmpty(t, filter.GetSources())
	assert.NotNil(t, filter.GetTimeRange())

	// Clear
	filter.Clear()

	// Should be empty
	assert.Empty(t, filter.GetNames())
	assert.Empty(t, filter.GetSources())
	assert.Nil(t, filter.GetTimeRange())

	// Should match everything
	event := EventRecord{Name: "anything", SourceID: "anything"}
	assert.True(t, filter.Matches(event))
}

// TestEventFilter_Concurrent tests thread-safe concurrent access
func TestEventFilter_Concurrent(t *testing.T) {
	filter := NewEventFilter()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			filter.WithNames("click", "submit")
			filter.WithSources("button-1", "form-1")
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			event := EventRecord{Name: "click", SourceID: "button-1"}
			_ = filter.Matches(event)
			_ = filter.GetNames()
			_ = filter.GetSources()
		}()
	}

	// Concurrent Apply
	events := []EventRecord{
		{Name: "click", SourceID: "button-1"},
		{Name: "submit", SourceID: "form-1"},
	}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = filter.Apply(events)
		}()
	}

	wg.Wait()
}

// TestEventFilter_GettersSetters tests getter/setter methods
func TestEventFilter_GettersSetters(t *testing.T) {
	filter := NewEventFilter()

	// Names
	filter.WithNames("click", "submit")
	names := filter.GetNames()
	assert.Equal(t, []string{"click", "submit"}, names)

	// Sources
	filter.WithSources("button-1", "form-1")
	sources := filter.GetSources()
	assert.Equal(t, []string{"button-1", "form-1"}, sources)

	// TimeRange
	start := time.Now()
	end := start.Add(1 * time.Hour)
	filter.WithTimeRange(start, end)
	tr := filter.GetTimeRange()
	require.NotNil(t, tr)
	assert.Equal(t, start, tr.Start)
	assert.Equal(t, end, tr.End)
}

// TestEventFilter_EmptyEvents tests filtering empty event slice
func TestEventFilter_EmptyEvents(t *testing.T) {
	filter := NewEventFilter().WithNames("click")
	filtered := filter.Apply([]EventRecord{})
	assert.Empty(t, filtered)
}

// TestEventFilter_NilTimeRange tests behavior with nil time range
func TestEventFilter_NilTimeRange(t *testing.T) {
	filter := NewEventFilter()
	event := EventRecord{Name: "click", Timestamp: time.Now()}

	// Nil time range should match all
	assert.True(t, filter.Matches(event))
}

// TestEventFilter_PartialMatch tests partial string matching
func TestEventFilter_PartialMatch(t *testing.T) {
	tests := []struct {
		name       string
		filterName string
		eventName  string
		wantMatch  bool
	}{
		{
			name:       "exact match",
			filterName: "click",
			eventName:  "click",
			wantMatch:  true,
		},
		{
			name:       "substring match",
			filterName: "click",
			eventName:  "onclick",
			wantMatch:  true,
		},
		{
			name:       "no match",
			filterName: "click",
			eventName:  "submit",
			wantMatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewEventFilter().WithNames(tt.filterName)
			event := EventRecord{Name: tt.eventName}
			assert.Equal(t, tt.wantMatch, filter.Matches(event))
		})
	}
}

// TestTimeRange_Contains tests TimeRange.Contains method
func TestTimeRange_Contains(t *testing.T) {
	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now.Add(1 * time.Hour)
	tr := &TimeRange{Start: start, End: end}

	tests := []struct {
		name string
		t    time.Time
		want bool
	}{
		{
			name: "within range",
			t:    now,
			want: true,
		},
		{
			name: "at start",
			t:    start,
			want: true,
		},
		{
			name: "at end",
			t:    end,
			want: true,
		},
		{
			name: "before range",
			t:    start.Add(-1 * time.Minute),
			want: false,
		},
		{
			name: "after range",
			t:    end.Add(1 * time.Minute),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tr.Contains(tt.t))
		})
	}
}
