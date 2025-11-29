// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"sort"
	"sync"
	"time"
)

// RecommendationRule defines a rule for generating optimization recommendations.
//
// Each rule has a condition function that evaluates a performance report
// and determines if a recommendation should be generated. Rules are
// evaluated in order and all matching rules produce recommendations.
//
// Example:
//
//	rule := RecommendationRule{
//	    Name:        "suggest_memoization",
//	    Condition:   func(r *Report) bool { return hasFrequentRenders(r) },
//	    Priority:    PriorityHigh,
//	    Category:    CategoryOptimization,
//	    Title:       "Implement Component Memoization",
//	    Description: "Some components render frequently with expensive operations",
//	    Action:      "Add memoization to prevent unnecessary re-renders",
//	    Impact:      ImpactHigh,
//	}
type RecommendationRule struct {
	// Name is a unique identifier for the rule
	Name string

	// Condition is the function that determines if the rule applies
	Condition func(*Report) bool

	// Priority indicates the importance of this recommendation
	Priority Priority

	// Category groups related recommendations
	Category Category

	// Title is a short description of the recommendation
	Title string

	// Description explains the recommendation in detail
	Description string

	// Action suggests what to do to address the issue
	Action string

	// Impact indicates the expected improvement from following this recommendation
	Impact ImpactLevel
}

// RecommendationEngine generates optimization recommendations from performance reports.
//
// It maintains a list of rules and applies them to performance reports
// to generate actionable recommendations. Custom rules can be added
// to extend the built-in recommendation logic.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	re := NewRecommendationEngine()
//	recommendations := re.Generate(report)
//	for _, r := range recommendations {
//	    fmt.Printf("[%s] %s: %s\n", r.Priority, r.Title, r.Action)
//	}
type RecommendationEngine struct {
	// rules is the list of rules to evaluate
	rules []RecommendationRule

	// mu protects concurrent access to engine state
	mu sync.RWMutex
}

// NewRecommendationEngine creates a new RecommendationEngine with default rules.
//
// The default rules include:
//   - suggest_memoization: Recommends memoization for frequently rendering components
//   - reduce_memory_usage: Recommends memory optimization for high-memory components
//   - optimize_slow_renders: Recommends render optimization for slow components
//   - batch_state_updates: Recommends batching for multiple bottlenecks
//   - review_architecture: Recommends architecture review for pattern issues
//
// Example:
//
//	re := NewRecommendationEngine()
//	recommendations := re.Generate(report)
func NewRecommendationEngine() *RecommendationEngine {
	return &RecommendationEngine{
		rules: defaultRules(),
	}
}

// NewRecommendationEngineWithRules creates a new RecommendationEngine with custom rules.
//
// This allows complete control over which rules are evaluated.
// Use AddRule to add rules to an existing engine instead.
//
// Example:
//
//	rules := []RecommendationRule{
//	    {Name: "custom", Condition: func(r *Report) bool { return true }, ...},
//	}
//	re := NewRecommendationEngineWithRules(rules)
func NewRecommendationEngineWithRules(rules []RecommendationRule) *RecommendationEngine {
	if rules == nil {
		rules = make([]RecommendationRule, 0)
	}
	return &RecommendationEngine{
		rules: rules,
	}
}

// Generate evaluates all rules against the report and returns recommendations.
//
// Recommendations are sorted by priority (Critical > High > Medium > Low).
// Returns an empty slice if no rules match or if report is nil.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	re := NewRecommendationEngine()
//	report := profiler.GenerateReport()
//	recommendations := re.Generate(report)
//	for _, r := range recommendations {
//	    fmt.Printf("[%s] %s\n", r.Priority, r.Title)
//	}
func (re *RecommendationEngine) Generate(report *Report) []*Recommendation {
	if report == nil {
		return []*Recommendation{}
	}

	re.mu.RLock()
	rules := make([]RecommendationRule, len(re.rules))
	copy(rules, re.rules)
	re.mu.RUnlock()

	recommendations := make([]*Recommendation, 0)

	for _, rule := range rules {
		if rule.Condition != nil && rule.Condition(report) {
			recommendations = append(recommendations, &Recommendation{
				Title:       rule.Title,
				Description: rule.Description,
				Action:      rule.Action,
				Priority:    rule.Priority,
				Category:    rule.Category,
				Impact:      rule.Impact,
			})
		}
	}

	// Sort by priority (Critical > High > Medium > Low)
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Priority > recommendations[j].Priority
	})

	return recommendations
}

// AddRule adds a custom rule to the engine.
//
// Rules are evaluated in the order they were added.
// Duplicate rule names are allowed but not recommended.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	re.AddRule(RecommendationRule{
//	    Name:        "custom_rule",
//	    Condition:   func(r *Report) bool { return len(r.Bottlenecks) > 10 },
//	    Priority:    PriorityCritical,
//	    Category:    CategoryOptimization,
//	    Title:       "Critical Performance Issues",
//	    Description: "Many bottlenecks detected",
//	    Action:      "Perform comprehensive performance audit",
//	    Impact:      ImpactHigh,
//	})
func (re *RecommendationEngine) AddRule(rule RecommendationRule) {
	re.mu.Lock()
	defer re.mu.Unlock()

	re.rules = append(re.rules, rule)
}

// RemoveRule removes a rule by name.
//
// If multiple rules have the same name, only the first one is removed.
// Returns true if a rule was removed, false if no rule with that name exists.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	removed := re.RemoveRule("suggest_memoization")
//	if removed {
//	    fmt.Println("Rule removed")
//	}
func (re *RecommendationEngine) RemoveRule(name string) bool {
	re.mu.Lock()
	defer re.mu.Unlock()

	for i, r := range re.rules {
		if r.Name == name {
			re.rules = append(re.rules[:i], re.rules[i+1:]...)
			return true
		}
	}
	return false
}

// GetRule returns a rule by name.
//
// Returns nil if no rule with that name exists.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	rule := re.GetRule("suggest_memoization")
//	if rule != nil {
//	    fmt.Printf("Rule priority: %d\n", rule.Priority)
//	}
func (re *RecommendationEngine) GetRule(name string) *RecommendationRule {
	re.mu.RLock()
	defer re.mu.RUnlock()

	for _, r := range re.rules {
		if r.Name == name {
			// Return a copy to prevent external modification
			ruleCopy := r
			return &ruleCopy
		}
	}
	return nil
}

// GetRules returns all registered rules.
//
// Returns a copy of the rules slice to prevent external modification.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	rules := re.GetRules()
//	for _, r := range rules {
//	    fmt.Printf("Rule: %s - %s\n", r.Name, r.Title)
//	}
func (re *RecommendationEngine) GetRules() []RecommendationRule {
	re.mu.RLock()
	defer re.mu.RUnlock()

	result := make([]RecommendationRule, len(re.rules))
	copy(result, re.rules)
	return result
}

// RuleCount returns the number of registered rules.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	count := re.RuleCount()
//	fmt.Printf("Registered rules: %d\n", count)
func (re *RecommendationEngine) RuleCount() int {
	re.mu.RLock()
	defer re.mu.RUnlock()

	return len(re.rules)
}

// Reset removes all rules and restores default rules.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	re.Reset() // Restore default rules
func (re *RecommendationEngine) Reset() {
	re.mu.Lock()
	defer re.mu.Unlock()

	re.rules = defaultRules()
}

// ClearRules removes all rules.
//
// After calling this, Generate will return empty results until rules are added.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	re.ClearRules() // Remove all rules
//	re.AddRule(customRule) // Add only custom rules
func (re *RecommendationEngine) ClearRules() {
	re.mu.Lock()
	defer re.mu.Unlock()

	re.rules = make([]RecommendationRule, 0)
}

// defaultRules returns the default set of recommendation rules.
func defaultRules() []RecommendationRule {
	return []RecommendationRule{
		{
			Name: "suggest_memoization",
			Condition: func(r *Report) bool {
				for _, comp := range r.Components {
					if comp.RenderCount > 100 && comp.AvgRenderTime > 5*time.Millisecond {
						return true
					}
				}
				return false
			},
			Priority:    PriorityHigh,
			Category:    CategoryOptimization,
			Title:       "Implement Component Memoization",
			Description: "Some components render frequently with expensive operations",
			Action:      "Add memoization to prevent unnecessary re-renders",
			Impact:      ImpactHigh,
		},
		{
			Name: "reduce_memory_usage",
			Condition: func(r *Report) bool {
				for _, comp := range r.Components {
					if comp.MemoryUsage > 5*1024*1024 { // 5MB
						return true
					}
				}
				return false
			},
			Priority:    PriorityHigh,
			Category:    CategoryMemory,
			Title:       "Reduce Memory Usage",
			Description: "Components are using excessive memory",
			Action:      "Use object pooling, sync.Pool, or review allocations",
			Impact:      ImpactHigh,
		},
		{
			Name: "optimize_slow_renders",
			Condition: func(r *Report) bool {
				for _, comp := range r.Components {
					if comp.AvgRenderTime > 16*time.Millisecond { // Frame budget
						return true
					}
				}
				return false
			},
			Priority:    PriorityCritical,
			Category:    CategoryRendering,
			Title:       "Optimize Slow Renders",
			Description: "Components exceed frame budget causing dropped frames",
			Action:      "Profile render functions and optimize hot paths",
			Impact:      ImpactHigh,
		},
		{
			Name: "batch_state_updates",
			Condition: func(r *Report) bool {
				// Recommend batching if there are many bottlenecks
				return len(r.Bottlenecks) > 5
			},
			Priority:    PriorityMedium,
			Category:    CategoryOptimization,
			Title:       "Batch State Updates",
			Description: "Multiple state updates causing re-renders",
			Action:      "Combine related state updates into single operations",
			Impact:      ImpactMedium,
		},
		{
			Name: "review_architecture",
			Condition: func(r *Report) bool {
				// Count pattern-type bottlenecks
				patternCount := 0
				for _, b := range r.Bottlenecks {
					if b.Type == BottleneckTypePattern {
						patternCount++
					}
				}
				return patternCount >= 3
			},
			Priority:    PriorityLow,
			Category:    CategoryArchitecture,
			Title:       "Review Component Architecture",
			Description: "Architectural patterns may need improvement",
			Action:      "Consider splitting large components or optimizing state flow",
			Impact:      ImpactMedium,
		},
	}
}
