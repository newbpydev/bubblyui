package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIcon_Creation(t *testing.T) {
	tests := []struct {
		name  string
		props IconProps
	}{
		{
			name: "Simple icon",
			props: IconProps{
				Symbol: "✓",
			},
		},
		{
			name: "Colored icon",
			props: IconProps{
				Symbol: "⚠",
				Color:  lipgloss.Color("226"),
			},
		},
		{
			name: "Icon with size",
			props: IconProps{
				Symbol: "★",
				Size:   SizeLarge,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon := Icon(tt.props)
			require.NotNil(t, icon, "Icon should not be nil")

			icon.Init()
			view := icon.View()

			assert.NotEmpty(t, view, "Icon view should not be empty")
			assert.Contains(t, view, tt.props.Symbol, "Icon should contain symbol")
		})
	}
}

func TestIcon_Symbols(t *testing.T) {
	tests := []struct {
		name   string
		symbol string
	}{
		{"Checkmark", "✓"},
		{"Cross", "✗"},
		{"Warning", "⚠"},
		{"Info", "ℹ"},
		{"Star", "★"},
		{"Heart", "♥"},
		{"Arrow right", "→"},
		{"Arrow left", "←"},
		{"Circle", "●"},
		{"Square", "■"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon := Icon(IconProps{
				Symbol: tt.symbol,
			})
			require.NotNil(t, icon)

			icon.Init()
			view := icon.View()

			assert.NotEmpty(t, view)
			assert.Contains(t, view, tt.symbol)
		})
	}
}

func TestIcon_Colors(t *testing.T) {
	tests := []struct {
		name  string
		color lipgloss.Color
	}{
		{"Red", lipgloss.Color("196")},
		{"Green", lipgloss.Color("46")},
		{"Blue", lipgloss.Color("63")},
		{"Yellow", lipgloss.Color("226")},
		{"No color", lipgloss.Color("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon := Icon(IconProps{
				Symbol: "●",
				Color:  tt.color,
			})
			require.NotNil(t, icon)

			icon.Init()
			view := icon.View()

			assert.NotEmpty(t, view)
			assert.Contains(t, view, "●")
		})
	}
}

func TestIcon_Sizes(t *testing.T) {
	tests := []struct {
		name string
		size Size
	}{
		{"Small", SizeSmall},
		{"Medium", SizeMedium},
		{"Large", SizeLarge},
		{"No size", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon := Icon(IconProps{
				Symbol: "★",
				Size:   tt.size,
			})
			require.NotNil(t, icon)

			icon.Init()
			view := icon.View()

			assert.NotEmpty(t, view)
		})
	}
}

func TestIcon_BubbleteatIntegration(t *testing.T) {
	icon := Icon(IconProps{
		Symbol: "✓",
	})
	require.NotNil(t, icon)

	cmd := icon.Init()
	assert.Nil(t, cmd, "Init should return nil cmd")

	updated, cmd := icon.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, updated, "Update should return updated component")
	assert.Nil(t, cmd, "Update should return nil cmd")

	view := icon.View()
	assert.NotEmpty(t, view, "View should return rendered output")
}

func TestIcon_Props(t *testing.T) {
	props := IconProps{
		Symbol: "⚠",
		Color:  lipgloss.Color("226"),
		Size:   SizeLarge,
	}

	icon := Icon(props)
	require.NotNil(t, icon)

	icon.Init()

	retrievedProps := icon.Props()
	assert.NotNil(t, retrievedProps)

	iconProps, ok := retrievedProps.(IconProps)
	assert.True(t, ok, "Props should be IconProps type")
	assert.Equal(t, "⚠", iconProps.Symbol)
	assert.Equal(t, lipgloss.Color("226"), iconProps.Color)
	assert.Equal(t, SizeLarge, iconProps.Size)
}

func TestIcon_EmptySymbol(t *testing.T) {
	icon := Icon(IconProps{
		Symbol: "",
	})
	require.NotNil(t, icon)

	icon.Init()
	view := icon.View()

	// Should render even with empty symbol
	assert.NotNil(t, view)
}

func TestIcon_ThemeIntegration(t *testing.T) {
	icon := Icon(IconProps{
		Symbol: "✓",
		Color:  lipgloss.Color(""), // Empty color should use theme
	})
	require.NotNil(t, icon)

	icon.Init()
	view := icon.View()

	assert.NotEmpty(t, view)
	assert.Contains(t, view, "✓")
}

func TestIcon_CustomStyle(t *testing.T) {
	customStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Padding(1, 2)

	icon := Icon(IconProps{
		Symbol: "★",
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})
	require.NotNil(t, icon)

	icon.Init()
	view := icon.View()

	assert.NotEmpty(t, view)
	assert.Contains(t, view, "★")
}
