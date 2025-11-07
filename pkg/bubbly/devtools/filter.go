package devtools

import "sync"

// FilterFunc is a predicate function that determines if a component should be included in the filter results.
//
// The function receives a ComponentSnapshot and returns true if the component should be included,
// false otherwise. This allows for custom filtering logic beyond type and status filtering.
//
// Example:
//
//	// Filter components with more than 5 refs
//	customFilter := func(c *ComponentSnapshot) bool {
//	    return len(c.Refs) > 5
//	}
type FilterFunc func(*ComponentSnapshot) bool

// ComponentFilter provides flexible filtering of component snapshots.
//
// The filter supports three types of criteria:
// - Type filtering: Match components by their Type field
// - Status filtering: Match components by their Status field
// - Custom filtering: Match components using a custom predicate function
//
// All filter criteria are combined with AND logic - a component must pass all
// active filters to be included in the results.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	filter := devtools.NewComponentFilter().
//	    WithTypes([]string{"button", "input"}).
//	    WithStatuses([]string{"mounted"}).
//	    WithCustom(func(c *ComponentSnapshot) bool {
//	        return len(c.Refs) > 0
//	    })
//
//	filtered := filter.Apply(components)
type ComponentFilter struct {
	mu       sync.RWMutex
	types    []string
	statuses []string
	custom   FilterFunc
}

// NewComponentFilter creates a new component filter with no criteria.
//
// An empty filter returns all components. Use the With* methods to add
// filter criteria.
func NewComponentFilter() *ComponentFilter {
	return &ComponentFilter{
		types:    []string{},
		statuses: []string{},
		custom:   nil,
	}
}

// WithTypes adds type filtering to the filter.
//
// Components will be included if their Type field matches any of the
// provided types. An empty types slice disables type filtering.
//
// This method returns the filter for method chaining.
func (cf *ComponentFilter) WithTypes(types []string) *ComponentFilter {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	cf.types = types
	return cf
}

// WithStatuses adds status filtering to the filter.
//
// Components will be included if their Status field matches any of the
// provided statuses. An empty statuses slice disables status filtering.
//
// This method returns the filter for method chaining.
func (cf *ComponentFilter) WithStatuses(statuses []string) *ComponentFilter {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	cf.statuses = statuses
	return cf
}

// WithCustom adds a custom filter function.
//
// The custom function is called for each component and should return true
// if the component should be included. A nil custom function disables
// custom filtering.
//
// This method returns the filter for method chaining.
func (cf *ComponentFilter) WithCustom(custom FilterFunc) *ComponentFilter {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	cf.custom = custom
	return cf
}

// Apply filters the provided components based on the configured criteria.
//
// The method returns a new slice containing only the components that pass
// all active filters. The original slice is not modified.
//
// Filter logic:
// - If no filters are configured, all components are returned
// - Type filter: component.Type must be in the types list
// - Status filter: component.Status must be in the statuses list
// - Custom filter: custom function must return true
// - All filters are combined with AND logic
//
// Thread Safety:
//
//	This method is thread-safe and can be called concurrently.
func (cf *ComponentFilter) Apply(components []*ComponentSnapshot) []*ComponentSnapshot {
	cf.mu.RLock()
	defer cf.mu.RUnlock()

	// Handle nil or empty input
	if components == nil {
		return []*ComponentSnapshot{}
	}

	// If no filters are configured, return all components
	if len(cf.types) == 0 && len(cf.statuses) == 0 && cf.custom == nil {
		return components
	}

	// Filter components
	result := make([]*ComponentSnapshot, 0, len(components))

	for _, component := range components {
		if cf.matchesFilters(component) {
			result = append(result, component)
		}
	}

	return result
}

// matchesFilters checks if a component passes all active filters.
// Must be called with read lock held.
func (cf *ComponentFilter) matchesFilters(component *ComponentSnapshot) bool {
	if component == nil {
		return false
	}

	// Check type filter
	if len(cf.types) > 0 && !cf.matchesType(component) {
		return false
	}

	// Check status filter
	if len(cf.statuses) > 0 && !cf.matchesStatus(component) {
		return false
	}

	// Check custom filter
	if cf.custom != nil && !cf.matchesCustom(component) {
		return false
	}

	return true
}

// matchesType checks if a component's type matches any of the configured types.
// Must be called with read lock held.
func (cf *ComponentFilter) matchesType(component *ComponentSnapshot) bool {
	for _, t := range cf.types {
		if component.Type == t {
			return true
		}
	}
	return false
}

// matchesStatus checks if a component's status matches any of the configured statuses.
// Must be called with read lock held.
func (cf *ComponentFilter) matchesStatus(component *ComponentSnapshot) bool {
	for _, s := range cf.statuses {
		if component.Status == s {
			return true
		}
	}
	return false
}

// matchesCustom checks if a component passes the custom filter function.
// Must be called with read lock held.
func (cf *ComponentFilter) matchesCustom(component *ComponentSnapshot) bool {
	return cf.custom(component)
}
