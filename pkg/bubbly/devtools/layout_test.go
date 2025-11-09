package devtools

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLayoutManager(t *testing.T) {
	tests := []struct {
		name      string
		mode      LayoutMode
		ratio     float64
		wantMode  LayoutMode
		wantRatio float64
	}{
		{
			name:      "Horizontal with valid ratio",
			mode:      LayoutHorizontal,
			ratio:     0.6,
			wantMode:  LayoutHorizontal,
			wantRatio: 0.6,
		},
		{
			name:      "Vertical with valid ratio",
			mode:      LayoutVertical,
			ratio:     0.7,
			wantMode:  LayoutVertical,
			wantRatio: 0.7,
		},
		{
			name:      "Ratio below 0 clamped to 0",
			mode:      LayoutHorizontal,
			ratio:     -0.5,
			wantMode:  LayoutHorizontal,
			wantRatio: 0.0,
		},
		{
			name:      "Ratio above 1 clamped to 1",
			mode:      LayoutHorizontal,
			ratio:     1.5,
			wantMode:  LayoutHorizontal,
			wantRatio: 1.0,
		},
		{
			name:      "Overlay mode",
			mode:      LayoutOverlay,
			ratio:     0.5,
			wantMode:  LayoutOverlay,
			wantRatio: 0.5,
		},
		{
			name:      "Hidden mode",
			mode:      LayoutHidden,
			ratio:     0.5,
			wantMode:  LayoutHidden,
			wantRatio: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := NewLayoutManager(tt.mode, tt.ratio)
			assert.NotNil(t, lm)
			assert.Equal(t, tt.wantMode, lm.GetMode())
			assert.Equal(t, tt.wantRatio, lm.GetRatio())
		})
	}
}

func TestLayoutManager_SetMode(t *testing.T) {
	lm := NewLayoutManager(LayoutHorizontal, 0.6)

	tests := []struct {
		name string
		mode LayoutMode
	}{
		{"Set to Vertical", LayoutVertical},
		{"Set to Overlay", LayoutOverlay},
		{"Set to Hidden", LayoutHidden},
		{"Set back to Horizontal", LayoutHorizontal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm.SetMode(tt.mode)
			assert.Equal(t, tt.mode, lm.GetMode())
		})
	}
}

func TestLayoutManager_SetRatio(t *testing.T) {
	lm := NewLayoutManager(LayoutHorizontal, 0.6)

	tests := []struct {
		name      string
		ratio     float64
		wantRatio float64
	}{
		{"Valid ratio 0.5", 0.5, 0.5},
		{"Valid ratio 0.8", 0.8, 0.8},
		{"Ratio 0.0", 0.0, 0.0},
		{"Ratio 1.0", 1.0, 1.0},
		{"Negative ratio clamped", -0.3, 0.0},
		{"Ratio > 1 clamped", 1.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm.SetRatio(tt.ratio)
			assert.Equal(t, tt.wantRatio, lm.GetRatio())
		})
	}
}

func TestLayoutManager_SetSize(t *testing.T) {
	lm := NewLayoutManager(LayoutHorizontal, 0.6)

	tests := []struct {
		name       string
		width      int
		height     int
		wantWidth  int
		wantHeight int
	}{
		{"Set 80x24", 80, 24, 80, 24},
		{"Set 120x40", 120, 40, 120, 40},
		{"Set 0x0", 0, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm.SetSize(tt.width, tt.height)
			gotWidth, gotHeight := lm.GetSize()
			assert.Equal(t, tt.wantWidth, gotWidth)
			assert.Equal(t, tt.wantHeight, gotHeight)
		})
	}
}

func TestLayoutManager_Render_Horizontal(t *testing.T) {
	tests := []struct {
		name         string
		ratio        float64
		width        int
		height       int
		appContent   string
		toolsContent string
	}{
		{
			name:         "60/40 split",
			ratio:        0.6,
			width:        100,
			height:       20,
			appContent:   "Application",
			toolsContent: "DevTools",
		},
		{
			name:         "50/50 split",
			ratio:        0.5,
			width:        80,
			height:       24,
			appContent:   "App",
			toolsContent: "Tools",
		},
		{
			name:         "70/30 split",
			ratio:        0.7,
			width:        120,
			height:       30,
			appContent:   "Main App",
			toolsContent: "Inspector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := NewLayoutManager(LayoutHorizontal, tt.ratio)
			lm.SetSize(tt.width, tt.height)

			result := lm.Render(tt.appContent, tt.toolsContent)

			// Verify result is not empty
			assert.NotEmpty(t, result)

			// Verify both contents appear in result
			assert.Contains(t, result, tt.appContent)
			assert.Contains(t, result, tt.toolsContent)

			// Verify result has multiple lines (joined horizontally)
			lines := strings.Split(result, "\n")
			assert.Greater(t, len(lines), 1)
		})
	}
}

func TestLayoutManager_Render_Vertical(t *testing.T) {
	tests := []struct {
		name         string
		ratio        float64
		width        int
		height       int
		appContent   string
		toolsContent string
	}{
		{
			name:         "60/40 split",
			ratio:        0.6,
			width:        80,
			height:       40,
			appContent:   "Application",
			toolsContent: "DevTools",
		},
		{
			name:         "50/50 split",
			ratio:        0.5,
			width:        100,
			height:       30,
			appContent:   "App",
			toolsContent: "Tools",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := NewLayoutManager(LayoutVertical, tt.ratio)
			lm.SetSize(tt.width, tt.height)

			result := lm.Render(tt.appContent, tt.toolsContent)

			// Verify result is not empty
			assert.NotEmpty(t, result)

			// Verify both contents appear in result
			assert.Contains(t, result, tt.appContent)
			assert.Contains(t, result, tt.toolsContent)

			// Verify result has multiple lines (joined vertically)
			lines := strings.Split(result, "\n")
			assert.Greater(t, len(lines), 1)
		})
	}
}

func TestLayoutManager_Render_Overlay(t *testing.T) {
	lm := NewLayoutManager(LayoutOverlay, 0.6)
	lm.SetSize(80, 24)

	appContent := "Application Content"
	toolsContent := "DevTools Panel"

	result := lm.Render(appContent, toolsContent)

	// Verify result is not empty
	assert.NotEmpty(t, result)

	// Verify both contents appear in result
	assert.Contains(t, result, appContent)
	assert.Contains(t, result, toolsContent)
}

func TestLayoutManager_Render_Hidden(t *testing.T) {
	lm := NewLayoutManager(LayoutHidden, 0.6)
	lm.SetSize(80, 24)

	appContent := "Application Content"
	toolsContent := "DevTools Panel"

	result := lm.Render(appContent, toolsContent)

	// Verify only app content is shown
	assert.Equal(t, appContent, result)

	// Verify tools content is NOT shown
	assert.NotContains(t, result, toolsContent)
}

func TestLayoutManager_Render_RatioAdjustment(t *testing.T) {
	lm := NewLayoutManager(LayoutHorizontal, 0.5)
	lm.SetSize(100, 20)

	appContent := "App"
	toolsContent := "Tools"

	// Render with 50/50
	result1 := lm.Render(appContent, toolsContent)
	assert.NotEmpty(t, result1)

	// Change ratio to 70/30
	lm.SetRatio(0.7)
	result2 := lm.Render(appContent, toolsContent)
	assert.NotEmpty(t, result2)

	// Results should be different (different layouts)
	assert.NotEqual(t, result1, result2)
}

func TestLayoutManager_Render_ResponsiveResize(t *testing.T) {
	lm := NewLayoutManager(LayoutHorizontal, 0.6)

	appContent := "Application"
	toolsContent := "DevTools"

	// Render at 80x24
	lm.SetSize(80, 24)
	result1 := lm.Render(appContent, toolsContent)
	assert.NotEmpty(t, result1)

	// Resize to 120x40
	lm.SetSize(120, 40)
	result2 := lm.Render(appContent, toolsContent)
	assert.NotEmpty(t, result2)

	// Results should be different (different sizes)
	assert.NotEqual(t, result1, result2)
}

func TestLayoutManager_Render_EmptyContent(t *testing.T) {
	lm := NewLayoutManager(LayoutHorizontal, 0.6)
	lm.SetSize(80, 24)

	tests := []struct {
		name         string
		appContent   string
		toolsContent string
	}{
		{"Empty app", "", "Tools"},
		{"Empty tools", "App", ""},
		{"Both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lm.Render(tt.appContent, tt.toolsContent)
			// Should not panic, result may be empty or contain borders
			assert.NotNil(t, result)
		})
	}
}

func TestLayoutManager_Render_MinimumSizes(t *testing.T) {
	tests := []struct {
		name   string
		mode   LayoutMode
		width  int
		height int
	}{
		{"Horizontal 1x1", LayoutHorizontal, 1, 1},
		{"Vertical 1x1", LayoutVertical, 1, 1},
		{"Horizontal 10x5", LayoutHorizontal, 10, 5},
		{"Vertical 10x5", LayoutVertical, 10, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := NewLayoutManager(tt.mode, 0.6)
			lm.SetSize(tt.width, tt.height)

			result := lm.Render("App", "Tools")
			// Should not panic with small sizes
			assert.NotNil(t, result)
		})
	}
}

func TestLayoutManager_Concurrent(t *testing.T) {
	lm := NewLayoutManager(LayoutHorizontal, 0.6)
	lm.SetSize(80, 24)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = lm.GetMode()
			_ = lm.GetRatio()
			_, _ = lm.GetSize()
			_ = lm.Render("App", "Tools")
		}()
	}

	// Concurrent writes
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			mode := LayoutMode(idx % 4)
			lm.SetMode(mode)
			lm.SetRatio(float64(idx%10) / 10.0)
			lm.SetSize(80+idx, 24+idx)
		}(i)
	}

	wg.Wait()

	// Verify final state is valid
	mode := lm.GetMode()
	assert.True(t, mode >= LayoutHorizontal && mode <= LayoutHidden)

	ratio := lm.GetRatio()
	assert.True(t, ratio >= 0.0 && ratio <= 1.0)

	width, height := lm.GetSize()
	assert.True(t, width >= 0 && height >= 0)
}

func TestLayoutManager_ModeSwitch(t *testing.T) {
	lm := NewLayoutManager(LayoutHorizontal, 0.6)
	lm.SetSize(100, 30)

	appContent := "Application"
	toolsContent := "DevTools"

	modes := []LayoutMode{
		LayoutHorizontal,
		LayoutVertical,
		LayoutOverlay,
		LayoutHidden,
	}

	for _, mode := range modes {
		t.Run(mode.String(), func(t *testing.T) {
			lm.SetMode(mode)
			result := lm.Render(appContent, toolsContent)
			assert.NotEmpty(t, result)

			// Verify mode-specific behavior
			switch mode {
			case LayoutHidden:
				assert.Equal(t, appContent, result)
			default:
				assert.NotEqual(t, appContent, result)
			}
		})
	}
}
