// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFlameGraphGenerator(t *testing.T) {
	tests := []struct {
		name string
		want func(*testing.T, *FlameGraphGenerator)
	}{
		{
			name: "creates generator with default dimensions",
			want: func(t *testing.T, fgg *FlameGraphGenerator) {
				assert.NotNil(t, fgg)
				assert.Equal(t, DefaultFlameGraphWidth, fgg.GetWidth())
				assert.Equal(t, DefaultFlameGraphHeight, fgg.GetHeight())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fgg := NewFlameGraphGenerator()
			tt.want(t, fgg)
		})
	}
}

func TestNewFlameGraphGeneratorWithDimensions(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		want   func(*testing.T, *FlameGraphGenerator)
	}{
		{
			name:   "creates generator with custom dimensions",
			width:  1200,
			height: 800,
			want: func(t *testing.T, fgg *FlameGraphGenerator) {
				assert.NotNil(t, fgg)
				assert.Equal(t, 1200, fgg.GetWidth())
				assert.Equal(t, 800, fgg.GetHeight())
			},
		},
		{
			name:   "uses defaults for zero width",
			width:  0,
			height: 500,
			want: func(t *testing.T, fgg *FlameGraphGenerator) {
				assert.Equal(t, DefaultFlameGraphWidth, fgg.GetWidth())
				assert.Equal(t, 500, fgg.GetHeight())
			},
		},
		{
			name:   "uses defaults for zero height",
			width:  800,
			height: 0,
			want: func(t *testing.T, fgg *FlameGraphGenerator) {
				assert.Equal(t, 800, fgg.GetWidth())
				assert.Equal(t, DefaultFlameGraphHeight, fgg.GetHeight())
			},
		},
		{
			name:   "uses defaults for negative dimensions",
			width:  -100,
			height: -50,
			want: func(t *testing.T, fgg *FlameGraphGenerator) {
				assert.Equal(t, DefaultFlameGraphWidth, fgg.GetWidth())
				assert.Equal(t, DefaultFlameGraphHeight, fgg.GetHeight())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fgg := NewFlameGraphGeneratorWithDimensions(tt.width, tt.height)
			tt.want(t, fgg)
		})
	}
}

func TestFlameGraphGenerator_Generate(t *testing.T) {
	tests := []struct {
		name    string
		profile *CPUProfileData
		want    func(*testing.T, *CallNode)
	}{
		{
			name:    "returns nil for nil profile",
			profile: nil,
			want: func(t *testing.T, root *CallNode) {
				assert.Nil(t, root)
			},
		},
		{
			name: "returns nil for empty profile",
			profile: &CPUProfileData{
				HotFunctions: []*HotFunction{},
				CallGraph:    make(map[string][]string),
				TotalSamples: 0,
			},
			want: func(t *testing.T, root *CallNode) {
				assert.Nil(t, root)
			},
		},
		{
			name: "builds call tree from single function",
			profile: &CPUProfileData{
				HotFunctions: []*HotFunction{
					{Name: "main.main", Samples: 100, Percent: 100.0},
				},
				CallGraph:    make(map[string][]string),
				TotalSamples: 100,
			},
			want: func(t *testing.T, root *CallNode) {
				require.NotNil(t, root)
				assert.Equal(t, "root", root.Name)
				assert.Equal(t, int64(100), root.Samples)
				assert.Len(t, root.Children, 1)
				assert.Equal(t, "main.main", root.Children[0].Name)
			},
		},
		{
			name: "builds call tree with nested calls",
			profile: &CPUProfileData{
				HotFunctions: []*HotFunction{
					{Name: "main.main", Samples: 100, Percent: 50.0},
					{Name: "main.render", Samples: 60, Percent: 30.0},
					{Name: "main.update", Samples: 40, Percent: 20.0},
				},
				CallGraph: map[string][]string{
					"main.main":   {"main.render", "main.update"},
					"main.render": {},
					"main.update": {},
				},
				TotalSamples: 200,
			},
			want: func(t *testing.T, root *CallNode) {
				require.NotNil(t, root)
				assert.Equal(t, "root", root.Name)
				assert.Equal(t, int64(200), root.Samples)
				// Should have main.main as child
				require.GreaterOrEqual(t, len(root.Children), 1)
			},
		},
		{
			name: "calculates percentages correctly",
			profile: &CPUProfileData{
				HotFunctions: []*HotFunction{
					{Name: "funcA", Samples: 50, Percent: 50.0},
					{Name: "funcB", Samples: 30, Percent: 30.0},
					{Name: "funcC", Samples: 20, Percent: 20.0},
				},
				CallGraph:    make(map[string][]string),
				TotalSamples: 100,
			},
			want: func(t *testing.T, root *CallNode) {
				require.NotNil(t, root)
				assert.Equal(t, float64(100.0), root.Percent)
				// Children should have correct percentages
				for _, child := range root.Children {
					switch child.Name {
					case "funcA":
						assert.Equal(t, float64(50.0), child.Percent)
					case "funcB":
						assert.Equal(t, float64(30.0), child.Percent)
					case "funcC":
						assert.Equal(t, float64(20.0), child.Percent)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fgg := NewFlameGraphGenerator()
			root := fgg.Generate(tt.profile)
			tt.want(t, root)
		})
	}
}

func TestFlameGraphGenerator_GenerateSVG(t *testing.T) {
	tests := []struct {
		name    string
		profile *CPUProfileData
		want    func(*testing.T, string)
	}{
		{
			name:    "generates empty SVG for nil profile",
			profile: nil,
			want: func(t *testing.T, svg string) {
				assert.Contains(t, svg, "<svg")
				assert.Contains(t, svg, "</svg>")
				assert.Contains(t, svg, "No profile data")
			},
		},
		{
			name: "generates SVG with valid header",
			profile: &CPUProfileData{
				HotFunctions: []*HotFunction{
					{Name: "main.main", Samples: 100, Percent: 100.0},
				},
				CallGraph:    make(map[string][]string),
				TotalSamples: 100,
			},
			want: func(t *testing.T, svg string) {
				assert.Contains(t, svg, "<svg")
				assert.Contains(t, svg, "xmlns=\"http://www.w3.org/2000/svg\"")
				assert.Contains(t, svg, "</svg>")
			},
		},
		{
			name: "generates SVG with rectangles for functions",
			profile: &CPUProfileData{
				HotFunctions: []*HotFunction{
					{Name: "main.main", Samples: 100, Percent: 100.0},
				},
				CallGraph:    make(map[string][]string),
				TotalSamples: 100,
			},
			want: func(t *testing.T, svg string) {
				assert.Contains(t, svg, "<rect")
				assert.Contains(t, svg, "main.main")
			},
		},
		{
			name: "generates SVG with text labels",
			profile: &CPUProfileData{
				HotFunctions: []*HotFunction{
					{Name: "myFunction", Samples: 50, Percent: 50.0},
				},
				CallGraph:    make(map[string][]string),
				TotalSamples: 100,
			},
			want: func(t *testing.T, svg string) {
				assert.Contains(t, svg, "<text")
				assert.Contains(t, svg, "myFunction")
			},
		},
		{
			name: "generates SVG with nested call structure",
			profile: &CPUProfileData{
				HotFunctions: []*HotFunction{
					{Name: "parent", Samples: 100, Percent: 100.0},
					{Name: "child", Samples: 50, Percent: 50.0},
				},
				CallGraph: map[string][]string{
					"parent": {"child"},
				},
				TotalSamples: 100,
			},
			want: func(t *testing.T, svg string) {
				assert.Contains(t, svg, "parent")
				assert.Contains(t, svg, "child")
				// Should have multiple rect elements
				rectCount := strings.Count(svg, "<rect")
				assert.GreaterOrEqual(t, rectCount, 2)
			},
		},
		{
			name: "escapes special characters in function names",
			profile: &CPUProfileData{
				HotFunctions: []*HotFunction{
					{Name: "func<T>", Samples: 100, Percent: 100.0},
				},
				CallGraph:    make(map[string][]string),
				TotalSamples: 100,
			},
			want: func(t *testing.T, svg string) {
				// Should escape < and > for valid XML
				assert.NotContains(t, svg, "<T>")
				assert.Contains(t, svg, "&lt;T&gt;")
			},
		},
		{
			name: "includes fill colors for rectangles",
			profile: &CPUProfileData{
				HotFunctions: []*HotFunction{
					{Name: "main.main", Samples: 100, Percent: 100.0},
				},
				CallGraph:    make(map[string][]string),
				TotalSamples: 100,
			},
			want: func(t *testing.T, svg string) {
				assert.Contains(t, svg, "fill=")
				// Should have flame-like colors (orange/red spectrum)
				assert.True(t, strings.Contains(svg, "rgb(") || strings.Contains(svg, "#"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fgg := NewFlameGraphGenerator()
			svg := fgg.GenerateSVG(tt.profile)
			tt.want(t, svg)
		})
	}
}

func TestFlameGraphGenerator_SetDimensions(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		want   func(*testing.T, *FlameGraphGenerator)
	}{
		{
			name:   "sets valid dimensions",
			width:  1000,
			height: 600,
			want: func(t *testing.T, fgg *FlameGraphGenerator) {
				assert.Equal(t, 1000, fgg.GetWidth())
				assert.Equal(t, 600, fgg.GetHeight())
			},
		},
		{
			name:   "ignores invalid width",
			width:  -100,
			height: 600,
			want: func(t *testing.T, fgg *FlameGraphGenerator) {
				assert.Equal(t, DefaultFlameGraphWidth, fgg.GetWidth())
				assert.Equal(t, 600, fgg.GetHeight())
			},
		},
		{
			name:   "ignores invalid height",
			width:  1000,
			height: 0,
			want: func(t *testing.T, fgg *FlameGraphGenerator) {
				assert.Equal(t, 1000, fgg.GetWidth())
				assert.Equal(t, DefaultFlameGraphHeight, fgg.GetHeight())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fgg := NewFlameGraphGenerator()
			fgg.SetDimensions(tt.width, tt.height)
			tt.want(t, fgg)
		})
	}
}

func TestCallNode_TotalSamples(t *testing.T) {
	tests := []struct {
		name string
		node *CallNode
		want int64
	}{
		{
			name: "returns samples for leaf node",
			node: &CallNode{
				Name:     "leaf",
				Samples:  100,
				Children: nil,
			},
			want: 100,
		},
		{
			name: "returns samples for node with children",
			node: &CallNode{
				Name:    "parent",
				Samples: 100,
				Children: []*CallNode{
					{Name: "child1", Samples: 50},
					{Name: "child2", Samples: 30},
				},
			},
			want: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.node.TotalSamples()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCallNode_AddChild(t *testing.T) {
	tests := []struct {
		name  string
		child *CallNode
		want  func(*testing.T, *CallNode)
	}{
		{
			name: "adds child to empty node",
			child: &CallNode{
				Name:    "child",
				Samples: 50,
			},
			want: func(t *testing.T, parent *CallNode) {
				assert.Len(t, parent.Children, 1)
				assert.Equal(t, "child", parent.Children[0].Name)
			},
		},
		{
			name:  "handles nil child gracefully",
			child: nil,
			want: func(t *testing.T, parent *CallNode) {
				assert.Len(t, parent.Children, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := &CallNode{
				Name:     "parent",
				Samples:  100,
				Children: make([]*CallNode, 0),
			}
			parent.AddChild(tt.child)
			tt.want(t, parent)
		})
	}
}

func TestFlameGraphGenerator_ThreadSafety(t *testing.T) {
	fgg := NewFlameGraphGenerator()
	profile := &CPUProfileData{
		HotFunctions: []*HotFunction{
			{Name: "main.main", Samples: 100, Percent: 100.0},
			{Name: "main.render", Samples: 50, Percent: 50.0},
		},
		CallGraph: map[string][]string{
			"main.main": {"main.render"},
		},
		TotalSamples: 100,
	}

	var wg sync.WaitGroup
	const goroutines = 50

	// Test concurrent Generate calls
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			root := fgg.Generate(profile)
			assert.NotNil(t, root)
		}()
	}

	// Test concurrent GenerateSVG calls
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svg := fgg.GenerateSVG(profile)
			assert.NotEmpty(t, svg)
		}()
	}

	// Test concurrent dimension changes
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			fgg.SetDimensions(800+idx, 600+idx)
			_ = fgg.GetWidth()
			_ = fgg.GetHeight()
		}(i)
	}

	wg.Wait()
}

func TestFlameGraphGenerator_GetColor(t *testing.T) {
	tests := []struct {
		name  string
		depth int
		want  func(*testing.T, string)
	}{
		{
			name:  "returns valid color for depth 0",
			depth: 0,
			want: func(t *testing.T, color string) {
				assert.NotEmpty(t, color)
				assert.True(t, strings.HasPrefix(color, "rgb(") || strings.HasPrefix(color, "#"))
			},
		},
		{
			name:  "returns valid color for depth 5",
			depth: 5,
			want: func(t *testing.T, color string) {
				assert.NotEmpty(t, color)
			},
		},
		{
			name:  "returns valid color for large depth",
			depth: 100,
			want: func(t *testing.T, color string) {
				assert.NotEmpty(t, color)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := getFlameColor(tt.depth)
			tt.want(t, color)
		})
	}
}

func TestTruncateLabel(t *testing.T) {
	tests := []struct {
		name     string
		label    string
		maxWidth int
		want     string
	}{
		{
			name:     "returns full label when fits",
			label:    "short",
			maxWidth: 100,
			want:     "short",
		},
		{
			name:     "truncates long label",
			label:    "this_is_a_very_long_function_name_that_should_be_truncated",
			maxWidth: 50,
			want:     "this...", // 50/7 = 7 chars max, minus 3 for "..." = 4 chars
		},
		{
			name:     "handles empty label",
			label:    "",
			maxWidth: 100,
			want:     "",
		},
		{
			name:     "handles zero width",
			label:    "test",
			maxWidth: 0,
			want:     "",
		},
		{
			name:     "handles very small width",
			label:    "test",
			maxWidth: 10,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateLabel(tt.label, tt.maxWidth)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "escapes less than",
			input: "func<T>",
			want:  "func&lt;T&gt;",
		},
		{
			name:  "escapes ampersand",
			input: "a & b",
			want:  "a &amp; b",
		},
		{
			name:  "escapes quotes",
			input: `"quoted"`,
			want:  "&quot;quoted&quot;",
		},
		{
			name:  "escapes apostrophe",
			input: "it's",
			want:  "it&#39;s",
		},
		{
			name:  "handles empty string",
			input: "",
			want:  "",
		},
		{
			name:  "handles no special chars",
			input: "normalFunction",
			want:  "normalFunction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeXML(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFlameGraphGenerator_Reset(t *testing.T) {
	fgg := NewFlameGraphGeneratorWithDimensions(1200, 800)
	fgg.Reset()

	assert.Equal(t, DefaultFlameGraphWidth, fgg.GetWidth())
	assert.Equal(t, DefaultFlameGraphHeight, fgg.GetHeight())
}

func TestFlameGraphGenerator_Integration(t *testing.T) {
	// Full workflow: CPUProfileData → Generate → GenerateSVG → Validate
	t.Run("full workflow", func(t *testing.T) {
		profile := &CPUProfileData{
			HotFunctions: []*HotFunction{
				{Name: "main.main", Samples: 200, Percent: 40.0},
				{Name: "main.render", Samples: 150, Percent: 30.0},
				{Name: "main.update", Samples: 100, Percent: 20.0},
				{Name: "main.handleEvent", Samples: 50, Percent: 10.0},
			},
			CallGraph: map[string][]string{
				"main.main":        {"main.render", "main.update"},
				"main.render":      {"main.handleEvent"},
				"main.update":      {},
				"main.handleEvent": {},
			},
			TotalSamples: 500,
		}

		fgg := NewFlameGraphGeneratorWithDimensions(1200, 600)

		// Generate call tree
		root := fgg.Generate(profile)
		require.NotNil(t, root)
		assert.Equal(t, "root", root.Name)
		assert.Equal(t, int64(500), root.Samples)

		// Generate SVG
		svg := fgg.GenerateSVG(profile)
		require.NotEmpty(t, svg)

		// Validate SVG structure
		assert.Contains(t, svg, "<svg")
		assert.Contains(t, svg, "</svg>")
		assert.Contains(t, svg, "main.main")
		assert.Contains(t, svg, "main.render")
		assert.Contains(t, svg, "main.update")

		// Should have multiple rectangles
		rectCount := strings.Count(svg, "<rect")
		assert.GreaterOrEqual(t, rectCount, 4) // At least one per function
	})
}
