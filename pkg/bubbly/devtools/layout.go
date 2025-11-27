package devtools

import (
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// LayoutManager manages the layout of the dev tools UI.
// It supports multiple layout modes (horizontal, vertical, overlay, hidden)
// and configurable split ratios for positioning the dev tools relative to
// the application content.
type LayoutManager struct {
	mu     sync.RWMutex
	mode   LayoutMode
	ratio  float64 // Split ratio (0.0 - 1.0) - app size / total size
	width  int     // Total width available
	height int     // Total height available
}

// NewLayoutManager creates a new LayoutManager with the specified mode and ratio.
// The ratio determines how much space the application takes (0.0 = none, 1.0 = all).
// Default ratio is 0.6 (60% app, 40% tools).
func NewLayoutManager(mode LayoutMode, ratio float64) *LayoutManager {
	// Clamp ratio to valid range
	if ratio < 0.0 {
		ratio = 0.0
	}
	if ratio > 1.0 {
		ratio = 1.0
	}

	return &LayoutManager{
		mode:  mode,
		ratio: ratio,
	}
}

// SetMode sets the layout mode.
func (lm *LayoutManager) SetMode(mode LayoutMode) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.mode = mode
}

// GetMode returns the current layout mode.
func (lm *LayoutManager) GetMode() LayoutMode {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.mode
}

// SetRatio sets the split ratio (0.0 - 1.0).
// The ratio determines how much space the application takes.
func (lm *LayoutManager) SetRatio(ratio float64) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Clamp to valid range
	if ratio < 0.0 {
		ratio = 0.0
	}
	if ratio > 1.0 {
		ratio = 1.0
	}

	lm.ratio = ratio
}

// GetRatio returns the current split ratio.
func (lm *LayoutManager) GetRatio() float64 {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.ratio
}

// SetSize sets the total available width and height.
func (lm *LayoutManager) SetSize(width, height int) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.width = width
	lm.height = height
}

// GetSize returns the current width and height.
func (lm *LayoutManager) GetSize() (width, height int) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.width, lm.height
}

// Render renders the application and dev tools content according to the current layout mode.
// Returns the final rendered output as a string.
func (lm *LayoutManager) Render(appContent, toolsContent string) string {
	lm.mu.RLock()
	mode := lm.mode
	ratio := lm.ratio
	width := lm.width
	height := lm.height
	lm.mu.RUnlock()

	switch mode {
	case LayoutHorizontal:
		return lm.renderHorizontal(appContent, toolsContent, ratio, width, height)
	case LayoutVertical:
		return lm.renderVertical(appContent, toolsContent, ratio, width, height)
	case LayoutOverlay:
		return lm.renderOverlay(appContent, toolsContent, width, height)
	case LayoutHidden:
		return appContent
	default:
		return appContent
	}
}

// calculateSplitDimensions calculates primary and secondary dimensions for split layouts.
// Returns (primary, secondary) dimensions ensuring both are at least 1.
func calculateSplitDimensions(total int, ratio float64) (int, int) {
	primary := int(float64(total) * ratio)
	secondary := total - primary - 1 // -1 for separator border

	// Ensure minimum dimensions
	if primary < 1 {
		primary = 1
	}
	if secondary < 1 {
		secondary = 1
	}

	return primary, secondary
}

// renderHorizontal renders app and tools side-by-side (left/right split).
func (lm *LayoutManager) renderHorizontal(appContent, toolsContent string, ratio float64, width, height int) string {
	// Calculate widths
	appWidth, toolsWidth := calculateSplitDimensions(width, ratio)

	// Style app box with right border separator
	appBox := lipgloss.NewStyle().
		Width(appWidth).
		Height(height).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(lipgloss.Color("240")).
		Render(appContent)

	// Style tools box
	toolsBox := lipgloss.NewStyle().
		Width(toolsWidth).
		Height(height).
		Render(toolsContent)

	// Join horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, appBox, toolsBox)
}

// renderVertical renders app and tools stacked (top/bottom split).
func (lm *LayoutManager) renderVertical(appContent, toolsContent string, ratio float64, width, height int) string {
	// Calculate heights
	appHeight, toolsHeight := calculateSplitDimensions(height, ratio)

	// Style app box with bottom border separator
	appBox := lipgloss.NewStyle().
		Width(width).
		Height(appHeight).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("240")).
		Render(appContent)

	// Style tools box
	toolsBox := lipgloss.NewStyle().
		Width(width).
		Height(toolsHeight).
		Render(toolsContent)

	// Join vertically
	return lipgloss.JoinVertical(lipgloss.Left, appBox, toolsBox)
}

// renderOverlay renders tools on top of the application.
func (lm *LayoutManager) renderOverlay(appContent, toolsContent string, width, height int) string {
	// Place app content as background
	appBox := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Render(appContent)

	// Place tools in center with border
	toolsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1).
		Render(toolsContent)

	// Overlay tools on app (center position)
	return lipgloss.Place(
		width, height,
		lipgloss.Center, lipgloss.Center,
		toolsBox,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.NoColor{}),
	) + "\n" + appBox
}

// CalculateResponsiveLayout determines the optimal layout mode and ratio based on terminal width.
//
// Breakpoints:
//   - < 80 cols: Vertical layout, 50/50 split (too narrow for side-by-side)
//   - 80-120 cols: Horizontal layout, 50/50 split (medium width)
//   - > 120 cols: Horizontal layout, 40/60 split (wide, more space for tools)
//
// Thread Safety:
//
//	Safe to call concurrently (pure function, no shared state).
//
// Example:
//
//	mode, ratio := devtools.CalculateResponsiveLayout(100)
//	// mode = LayoutHorizontal, ratio = 0.5
//
// Parameters:
//   - width: Terminal width in columns
//
// Returns:
//   - LayoutMode: The recommended layout mode
//   - float64: The recommended split ratio (app size / total size)
func CalculateResponsiveLayout(width int) (LayoutMode, float64) {
	switch {
	case width < 80:
		// Narrow terminal: use vertical layout with 50/50 split
		return LayoutVertical, 0.5
	case width <= 120:
		// Medium terminal: use horizontal layout with 50/50 split
		return LayoutHorizontal, 0.5
	default:
		// Wide terminal: use horizontal layout with 40/60 split
		// (40% app, 60% tools for more inspection space)
		return LayoutHorizontal, 0.4
	}
}
