package profiler

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRecommendationEngine(t *testing.T) {
	tests := []struct {
		name      string
		wantRules int
	}{
		{
			name:      "creates engine with default rules",
			wantRules: 5, // 5 default rules
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := NewRecommendationEngine()
			require.NotNil(t, re)
			assert.Equal(t, tt.wantRules, re.RuleCount())
			assert.NotEmpty(t, re.GetRules())
		})
	}
}

func TestNewRecommendationEngineWithRules(t *testing.T) {
	tests := []struct {
		name      string
		rules     []RecommendationRule
		wantRules int
	}{
		{
			name:      "creates engine with nil rules",
			rules:     nil,
			wantRules: 0,
		},
		{
			name:      "creates engine with empty rules",
			rules:     []RecommendationRule{},
			wantRules: 0,
		},
		{
			name: "creates engine with custom rules",
			rules: []RecommendationRule{
				{Name: "custom1", Priority: PriorityLow},
				{Name: "custom2", Priority: PriorityHigh},
			},
			wantRules: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := NewRecommendationEngineWithRules(tt.rules)
			require.NotNil(t, re)
			assert.Equal(t, tt.wantRules, re.RuleCount())
		})
	}
}

func TestRecommendationEngine_Generate(t *testing.T) {
	tests := []struct {
		name                string
		rules               []RecommendationRule
		report              *Report
		wantRecommendations int
		wantTitles          []string
	}{
		{
			name:                "returns empty for nil report",
			rules:               defaultRules(),
			report:              nil,
			wantRecommendations: 0,
		},
		{
			name:  "returns empty for empty report",
			rules: defaultRules(),
			report: &Report{
				Components:  []*ComponentMetrics{},
				Bottlenecks: []*BottleneckInfo{},
			},
			wantRecommendations: 0,
		},
		{
			name: "evaluates rule condition correctly",
			rules: []RecommendationRule{
				{
					Name:        "always_true",
					Condition:   func(r *Report) bool { return true },
					Priority:    PriorityHigh,
					Title:       "Always True",
					Description: "This always triggers",
					Action:      "Do something",
					Impact:      ImpactHigh,
				},
			},
			report: &Report{
				Components:  []*ComponentMetrics{},
				Bottlenecks: []*BottleneckInfo{},
			},
			wantRecommendations: 1,
			wantTitles:          []string{"Always True"},
		},
		{
			name: "skips rules with nil condition",
			rules: []RecommendationRule{
				{
					Name:      "nil_condition",
					Condition: nil,
					Priority:  PriorityHigh,
					Title:     "Nil Condition",
				},
			},
			report: &Report{
				Components:  []*ComponentMetrics{},
				Bottlenecks: []*BottleneckInfo{},
			},
			wantRecommendations: 0,
		},
		{
			name: "skips rules with false condition",
			rules: []RecommendationRule{
				{
					Name:      "always_false",
					Condition: func(r *Report) bool { return false },
					Priority:  PriorityHigh,
					Title:     "Always False",
				},
			},
			report: &Report{
				Components:  []*ComponentMetrics{},
				Bottlenecks: []*BottleneckInfo{},
			},
			wantRecommendations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := NewRecommendationEngineWithRules(tt.rules)
			recommendations := re.Generate(tt.report)

			assert.Len(t, recommendations, tt.wantRecommendations)

			if len(tt.wantTitles) > 0 {
				for i, title := range tt.wantTitles {
					assert.Equal(t, title, recommendations[i].Title)
				}
			}
		})
	}
}

func TestRecommendationEngine_Generate_PrioritySorting(t *testing.T) {
	tests := []struct {
		name           string
		rules          []RecommendationRule
		wantPriorities []Priority
	}{
		{
			name: "sorts by priority descending",
			rules: []RecommendationRule{
				{Name: "low", Condition: func(r *Report) bool { return true }, Priority: PriorityLow, Title: "Low"},
				{Name: "critical", Condition: func(r *Report) bool { return true }, Priority: PriorityCritical, Title: "Critical"},
				{Name: "medium", Condition: func(r *Report) bool { return true }, Priority: PriorityMedium, Title: "Medium"},
				{Name: "high", Condition: func(r *Report) bool { return true }, Priority: PriorityHigh, Title: "High"},
			},
			wantPriorities: []Priority{PriorityCritical, PriorityHigh, PriorityMedium, PriorityLow},
		},
		{
			name: "maintains order for same priority",
			rules: []RecommendationRule{
				{Name: "high1", Condition: func(r *Report) bool { return true }, Priority: PriorityHigh, Title: "High1"},
				{Name: "high2", Condition: func(r *Report) bool { return true }, Priority: PriorityHigh, Title: "High2"},
			},
			wantPriorities: []Priority{PriorityHigh, PriorityHigh},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := NewRecommendationEngineWithRules(tt.rules)
			report := &Report{Components: []*ComponentMetrics{}, Bottlenecks: []*BottleneckInfo{}}
			recommendations := re.Generate(report)

			require.Len(t, recommendations, len(tt.wantPriorities))
			for i, wantPriority := range tt.wantPriorities {
				assert.Equal(t, wantPriority, recommendations[i].Priority, "index %d", i)
			}
		})
	}
}

func TestRecommendationEngine_Generate_MultipleRecommendations(t *testing.T) {
	tests := []struct {
		name                string
		rules               []RecommendationRule
		report              *Report
		wantRecommendations int
	}{
		{
			name: "generates multiple recommendations",
			rules: []RecommendationRule{
				{Name: "rule1", Condition: func(r *Report) bool { return true }, Priority: PriorityHigh, Title: "Rule1"},
				{Name: "rule2", Condition: func(r *Report) bool { return true }, Priority: PriorityMedium, Title: "Rule2"},
				{Name: "rule3", Condition: func(r *Report) bool { return true }, Priority: PriorityLow, Title: "Rule3"},
			},
			report: &Report{
				Components:  []*ComponentMetrics{},
				Bottlenecks: []*BottleneckInfo{},
			},
			wantRecommendations: 3,
		},
		{
			name: "generates only matching recommendations",
			rules: []RecommendationRule{
				{Name: "rule1", Condition: func(r *Report) bool { return true }, Priority: PriorityHigh, Title: "Rule1"},
				{Name: "rule2", Condition: func(r *Report) bool { return false }, Priority: PriorityMedium, Title: "Rule2"},
				{Name: "rule3", Condition: func(r *Report) bool { return true }, Priority: PriorityLow, Title: "Rule3"},
			},
			report: &Report{
				Components:  []*ComponentMetrics{},
				Bottlenecks: []*BottleneckInfo{},
			},
			wantRecommendations: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := NewRecommendationEngineWithRules(tt.rules)
			recommendations := re.Generate(tt.report)

			assert.Len(t, recommendations, tt.wantRecommendations)
		})
	}
}

func TestRecommendationEngine_Generate_DefaultRules(t *testing.T) {
	tests := []struct {
		name                string
		report              *Report
		wantRecommendations int
		wantRuleNames       []string
	}{
		{
			name: "triggers suggest_memoization rule",
			report: &Report{
				Components: []*ComponentMetrics{
					{
						ComponentName: "FrequentComponent",
						RenderCount:   200,                   // > 100
						AvgRenderTime: 10 * time.Millisecond, // > 5ms
					},
				},
				Bottlenecks: []*BottleneckInfo{},
			},
			wantRecommendations: 1,
			wantRuleNames:       []string{"Implement Component Memoization"},
		},
		{
			name: "triggers reduce_memory_usage rule",
			report: &Report{
				Components: []*ComponentMetrics{
					{
						ComponentName: "MemoryHog",
						RenderCount:   10,
						AvgRenderTime: 1 * time.Millisecond,
						MemoryUsage:   10 * 1024 * 1024, // 10MB > 5MB
					},
				},
				Bottlenecks: []*BottleneckInfo{},
			},
			wantRecommendations: 1,
			wantRuleNames:       []string{"Reduce Memory Usage"},
		},
		{
			name: "triggers optimize_slow_renders rule",
			report: &Report{
				Components: []*ComponentMetrics{
					{
						ComponentName: "SlowComponent",
						RenderCount:   10,
						AvgRenderTime: 20 * time.Millisecond, // > 16ms
					},
				},
				Bottlenecks: []*BottleneckInfo{},
			},
			wantRecommendations: 1,
			wantRuleNames:       []string{"Optimize Slow Renders"},
		},
		{
			name: "triggers batch_state_updates rule",
			report: &Report{
				Components: []*ComponentMetrics{},
				Bottlenecks: []*BottleneckInfo{
					{Type: BottleneckTypeSlow},
					{Type: BottleneckTypeSlow},
					{Type: BottleneckTypeSlow},
					{Type: BottleneckTypeSlow},
					{Type: BottleneckTypeSlow},
					{Type: BottleneckTypeSlow}, // 6 bottlenecks > 5
				},
			},
			wantRecommendations: 1,
			wantRuleNames:       []string{"Batch State Updates"},
		},
		{
			name: "triggers review_architecture rule",
			report: &Report{
				Components: []*ComponentMetrics{},
				Bottlenecks: []*BottleneckInfo{
					{Type: BottleneckTypePattern},
					{Type: BottleneckTypePattern},
					{Type: BottleneckTypePattern}, // 3 pattern bottlenecks
				},
			},
			wantRecommendations: 1,
			wantRuleNames:       []string{"Review Component Architecture"},
		},
		{
			name: "triggers multiple default rules",
			report: &Report{
				Components: []*ComponentMetrics{
					{
						ComponentName: "ProblematicComponent",
						RenderCount:   200,                   // triggers memoization
						AvgRenderTime: 20 * time.Millisecond, // triggers slow_renders
						MemoryUsage:   10 * 1024 * 1024,      // triggers memory
					},
				},
				Bottlenecks: []*BottleneckInfo{
					{Type: BottleneckTypeSlow},
					{Type: BottleneckTypeSlow},
					{Type: BottleneckTypeSlow},
					{Type: BottleneckTypeSlow},
					{Type: BottleneckTypeSlow},
					{Type: BottleneckTypeSlow}, // triggers batch_state_updates
					{Type: BottleneckTypePattern},
					{Type: BottleneckTypePattern},
					{Type: BottleneckTypePattern}, // triggers review_architecture
				},
			},
			wantRecommendations: 5, // All 5 default rules should trigger
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := NewRecommendationEngine()
			recommendations := re.Generate(tt.report)

			assert.Len(t, recommendations, tt.wantRecommendations)

			// Verify expected rule names are present
			titles := make([]string, len(recommendations))
			for i, r := range recommendations {
				titles[i] = r.Title
			}
			for _, wantName := range tt.wantRuleNames {
				assert.Contains(t, titles, wantName)
			}
		})
	}
}

func TestRecommendationEngine_AddRule(t *testing.T) {
	tests := []struct {
		name          string
		initialRules  []RecommendationRule
		ruleToAdd     RecommendationRule
		wantRuleCount int
	}{
		{
			name:         "adds rule to empty engine",
			initialRules: []RecommendationRule{},
			ruleToAdd: RecommendationRule{
				Name:     "new_rule",
				Priority: PriorityHigh,
				Title:    "New Rule",
			},
			wantRuleCount: 1,
		},
		{
			name: "adds rule to existing rules",
			initialRules: []RecommendationRule{
				{Name: "existing", Priority: PriorityLow},
			},
			ruleToAdd: RecommendationRule{
				Name:     "new_rule",
				Priority: PriorityHigh,
				Title:    "New Rule",
			},
			wantRuleCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := NewRecommendationEngineWithRules(tt.initialRules)
			re.AddRule(tt.ruleToAdd)

			assert.Equal(t, tt.wantRuleCount, re.RuleCount())
			assert.NotNil(t, re.GetRule(tt.ruleToAdd.Name))
		})
	}
}

func TestRecommendationEngine_RemoveRule(t *testing.T) {
	tests := []struct {
		name          string
		initialRules  []RecommendationRule
		ruleToRemove  string
		wantRemoved   bool
		wantRuleCount int
	}{
		{
			name:          "returns false for empty engine",
			initialRules:  []RecommendationRule{},
			ruleToRemove:  "nonexistent",
			wantRemoved:   false,
			wantRuleCount: 0,
		},
		{
			name: "returns false for nonexistent rule",
			initialRules: []RecommendationRule{
				{Name: "existing", Priority: PriorityLow},
			},
			ruleToRemove:  "nonexistent",
			wantRemoved:   false,
			wantRuleCount: 1,
		},
		{
			name: "removes existing rule",
			initialRules: []RecommendationRule{
				{Name: "to_remove", Priority: PriorityLow},
				{Name: "to_keep", Priority: PriorityHigh},
			},
			ruleToRemove:  "to_remove",
			wantRemoved:   true,
			wantRuleCount: 1,
		},
		{
			name: "removes first matching rule only",
			initialRules: []RecommendationRule{
				{Name: "duplicate", Priority: PriorityLow},
				{Name: "duplicate", Priority: PriorityHigh},
			},
			ruleToRemove:  "duplicate",
			wantRemoved:   true,
			wantRuleCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := NewRecommendationEngineWithRules(tt.initialRules)
			removed := re.RemoveRule(tt.ruleToRemove)

			assert.Equal(t, tt.wantRemoved, removed)
			assert.Equal(t, tt.wantRuleCount, re.RuleCount())
		})
	}
}

func TestRecommendationEngine_GetRule(t *testing.T) {
	tests := []struct {
		name         string
		initialRules []RecommendationRule
		ruleName     string
		wantNil      bool
		wantPriority Priority
	}{
		{
			name:         "returns nil for empty engine",
			initialRules: []RecommendationRule{},
			ruleName:     "nonexistent",
			wantNil:      true,
		},
		{
			name: "returns nil for nonexistent rule",
			initialRules: []RecommendationRule{
				{Name: "existing", Priority: PriorityLow},
			},
			ruleName: "nonexistent",
			wantNil:  true,
		},
		{
			name: "returns existing rule",
			initialRules: []RecommendationRule{
				{Name: "target", Priority: PriorityHigh},
			},
			ruleName:     "target",
			wantNil:      false,
			wantPriority: PriorityHigh,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := NewRecommendationEngineWithRules(tt.initialRules)
			rule := re.GetRule(tt.ruleName)

			if tt.wantNil {
				assert.Nil(t, rule)
			} else {
				require.NotNil(t, rule)
				assert.Equal(t, tt.wantPriority, rule.Priority)
			}
		})
	}
}

func TestRecommendationEngine_GetRules(t *testing.T) {
	tests := []struct {
		name         string
		initialRules []RecommendationRule
		wantCount    int
	}{
		{
			name:         "returns empty for empty engine",
			initialRules: []RecommendationRule{},
			wantCount:    0,
		},
		{
			name: "returns all rules",
			initialRules: []RecommendationRule{
				{Name: "rule1", Priority: PriorityLow},
				{Name: "rule2", Priority: PriorityHigh},
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := NewRecommendationEngineWithRules(tt.initialRules)
			rules := re.GetRules()

			assert.Len(t, rules, tt.wantCount)

			// Verify it's a copy by modifying returned slice
			if len(rules) > 0 {
				rules[0].Name = "modified"
				originalRules := re.GetRules()
				assert.NotEqual(t, "modified", originalRules[0].Name)
			}
		})
	}
}

func TestRecommendationEngine_RuleCount(t *testing.T) {
	tests := []struct {
		name         string
		initialRules []RecommendationRule
		wantCount    int
	}{
		{
			name:         "returns 0 for empty engine",
			initialRules: []RecommendationRule{},
			wantCount:    0,
		},
		{
			name: "returns correct count",
			initialRules: []RecommendationRule{
				{Name: "rule1"},
				{Name: "rule2"},
				{Name: "rule3"},
			},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := NewRecommendationEngineWithRules(tt.initialRules)
			assert.Equal(t, tt.wantCount, re.RuleCount())
		})
	}
}

func TestRecommendationEngine_Reset(t *testing.T) {
	tests := []struct {
		name         string
		initialRules []RecommendationRule
		wantCount    int
	}{
		{
			name:         "resets empty engine to defaults",
			initialRules: []RecommendationRule{},
			wantCount:    5, // 5 default rules
		},
		{
			name: "resets custom rules to defaults",
			initialRules: []RecommendationRule{
				{Name: "custom1"},
				{Name: "custom2"},
			},
			wantCount: 5, // 5 default rules
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := NewRecommendationEngineWithRules(tt.initialRules)
			re.Reset()

			assert.Equal(t, tt.wantCount, re.RuleCount())
			// Verify default rules are present
			assert.NotNil(t, re.GetRule("suggest_memoization"))
			assert.NotNil(t, re.GetRule("reduce_memory_usage"))
			assert.NotNil(t, re.GetRule("optimize_slow_renders"))
			assert.NotNil(t, re.GetRule("batch_state_updates"))
			assert.NotNil(t, re.GetRule("review_architecture"))
		})
	}
}

func TestRecommendationEngine_ClearRules(t *testing.T) {
	tests := []struct {
		name         string
		initialRules []RecommendationRule
	}{
		{
			name:         "clears empty engine",
			initialRules: []RecommendationRule{},
		},
		{
			name: "clears all rules",
			initialRules: []RecommendationRule{
				{Name: "rule1"},
				{Name: "rule2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := NewRecommendationEngineWithRules(tt.initialRules)
			re.ClearRules()

			assert.Equal(t, 0, re.RuleCount())
			assert.Empty(t, re.GetRules())
		})
	}
}

func TestRecommendationEngine_ConcurrentAccess(t *testing.T) {
	re := NewRecommendationEngine()
	report := &Report{
		Components: []*ComponentMetrics{
			{
				ComponentName: "TestComponent",
				RenderCount:   200,
				AvgRenderTime: 20 * time.Millisecond,
				MemoryUsage:   10 * 1024 * 1024,
			},
		},
		Bottlenecks: []*BottleneckInfo{
			{Type: BottleneckTypeSlow},
			{Type: BottleneckTypeSlow},
			{Type: BottleneckTypeSlow},
			{Type: BottleneckTypeSlow},
			{Type: BottleneckTypeSlow},
			{Type: BottleneckTypeSlow},
		},
	}

	var wg sync.WaitGroup
	goroutines := 50

	// Test concurrent reads
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = re.Generate(report)
			_ = re.GetRules()
			_ = re.RuleCount()
			_ = re.GetRule("suggest_memoization")
		}()
	}

	// Test concurrent writes
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			re.AddRule(RecommendationRule{
				Name:      "concurrent_rule",
				Condition: func(r *Report) bool { return true },
				Priority:  PriorityLow,
			})
			re.RemoveRule("concurrent_rule")
		}(i)
	}

	wg.Wait()

	// Verify engine is still functional
	recommendations := re.Generate(report)
	assert.NotNil(t, recommendations)
}

func TestRecommendationEngine_ActionableSuggestions(t *testing.T) {
	// Test that all default rules produce actionable suggestions
	re := NewRecommendationEngine()
	rules := re.GetRules()

	for _, rule := range rules {
		t.Run(rule.Name, func(t *testing.T) {
			assert.NotEmpty(t, rule.Title, "Rule %s should have a title", rule.Name)
			assert.NotEmpty(t, rule.Description, "Rule %s should have a description", rule.Name)
			assert.NotEmpty(t, rule.Action, "Rule %s should have an action", rule.Name)
			assert.NotNil(t, rule.Condition, "Rule %s should have a condition", rule.Name)
			assert.True(t, rule.Priority >= PriorityLow && rule.Priority <= PriorityCritical,
				"Rule %s should have valid priority", rule.Name)
			assert.NotEmpty(t, rule.Category, "Rule %s should have a category", rule.Name)
			assert.NotEmpty(t, rule.Impact, "Rule %s should have an impact level", rule.Name)
		})
	}
}

func TestRecommendationEngine_RecommendationFields(t *testing.T) {
	// Test that generated recommendations have all required fields
	re := NewRecommendationEngine()
	report := &Report{
		Components: []*ComponentMetrics{
			{
				ComponentName: "TestComponent",
				RenderCount:   200,
				AvgRenderTime: 20 * time.Millisecond,
			},
		},
		Bottlenecks: []*BottleneckInfo{},
	}

	recommendations := re.Generate(report)
	require.NotEmpty(t, recommendations)

	for _, r := range recommendations {
		assert.NotEmpty(t, r.Title, "Recommendation should have a title")
		assert.NotEmpty(t, r.Description, "Recommendation should have a description")
		assert.NotEmpty(t, r.Action, "Recommendation should have an action")
		assert.True(t, r.Priority >= PriorityLow && r.Priority <= PriorityCritical,
			"Recommendation should have valid priority")
		assert.NotEmpty(t, r.Category, "Recommendation should have a category")
		assert.NotEmpty(t, r.Impact, "Recommendation should have an impact level")
	}
}
