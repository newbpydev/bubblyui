package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBadge_Creation(t *testing.T) {
	tests := []struct {
		name  string
		props BadgeProps
	}{
		{
			name: "Simple badge",
			props: BadgeProps{
				Label: "New",
			},
		},
		{
			name: "Badge with variant",
			props: BadgeProps{
				Label:   "Error",
				Variant: VariantDanger,
			},
		},
		{
			name: "Badge with custom color",
			props: BadgeProps{
				Label: "Custom",
				Color: lipgloss.Color("99"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			badge := Badge(tt.props)
			require.NotNil(t, badge, "Badge should not be nil")

			badge.Init()
			view := badge.View()

			assert.NotEmpty(t, view, "Badge view should not be empty")
			assert.Contains(t, view, tt.props.Label, "Badge should contain label")
		})
	}
}

func TestBadge_Variants(t *testing.T) {
	tests := []struct {
		name    string
		variant Variant
	}{
		{"Primary", VariantPrimary},
		{"Secondary", VariantSecondary},
		{"Success", VariantSuccess},
		{"Warning", VariantWarning},
		{"Danger", VariantDanger},
		{"Info", VariantInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			badge := Badge(BadgeProps{
				Label:   "Test",
				Variant: tt.variant,
			})
			require.NotNil(t, badge)

			badge.Init()
			view := badge.View()

			assert.NotEmpty(t, view)
			assert.Contains(t, view, "Test")
		})
	}
}

func TestBadge_Labels(t *testing.T) {
	tests := []struct {
		name  string
		label string
	}{
		{"Short label", "New"},
		{"Medium label", "In Progress"},
		{"Long label", "Waiting for Approval"},
		{"Number", "42"},
		{"Symbol", "✓"},
		{"Empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			badge := Badge(BadgeProps{
				Label: tt.label,
			})
			require.NotNil(t, badge)

			badge.Init()
			view := badge.View()

			assert.NotNil(t, view)
			if tt.label != "" {
				assert.Contains(t, view, tt.label)
			}
		})
	}
}

func TestBadge_CustomColor(t *testing.T) {
	tests := []struct {
		name  string
		color lipgloss.Color
	}{
		{"Red", lipgloss.Color("196")},
		{"Green", lipgloss.Color("46")},
		{"Blue", lipgloss.Color("63")},
		{"Purple", lipgloss.Color("99")},
		{"No color", lipgloss.Color("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			badge := Badge(BadgeProps{
				Label: "Test",
				Color: tt.color,
			})
			require.NotNil(t, badge)

			badge.Init()
			view := badge.View()

			assert.NotEmpty(t, view)
			assert.Contains(t, view, "Test")
		})
	}
}

func TestBadge_BubbleteatIntegration(t *testing.T) {
	badge := Badge(BadgeProps{
		Label:   "Status",
		Variant: VariantSuccess,
	})
	require.NotNil(t, badge)

	cmd := badge.Init()
	assert.Nil(t, cmd, "Init should return nil cmd")

	updated, cmd := badge.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, updated, "Update should return updated component")
	assert.Nil(t, cmd, "Update should return nil cmd")

	view := badge.View()
	assert.NotEmpty(t, view, "View should return rendered output")
}

func TestBadge_Props(t *testing.T) {
	props := BadgeProps{
		Label:   "Important",
		Variant: VariantWarning,
		Color:   lipgloss.Color("226"),
	}

	badge := Badge(props)
	require.NotNil(t, badge)

	badge.Init()

	retrievedProps := badge.Props()
	assert.NotNil(t, retrievedProps)

	badgeProps, ok := retrievedProps.(BadgeProps)
	assert.True(t, ok, "Props should be BadgeProps type")
	assert.Equal(t, "Important", badgeProps.Label)
	assert.Equal(t, VariantWarning, badgeProps.Variant)
	assert.Equal(t, lipgloss.Color("226"), badgeProps.Color)
}

func TestBadge_ThemeIntegration(t *testing.T) {
	badge := Badge(BadgeProps{
		Label:   "Themed",
		Variant: VariantPrimary,
	})
	require.NotNil(t, badge)

	badge.Init()
	view := badge.View()

	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Themed")
}

func TestBadge_CustomStyle(t *testing.T) {
	customStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Padding(0, 3)

	badge := Badge(BadgeProps{
		Label: "Custom",
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})
	require.NotNil(t, badge)

	badge.Init()
	view := badge.View()

	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Custom")
}

func TestBadge_DefaultVariant(t *testing.T) {
	// Test badge with empty variant defaults to primary
	badge := Badge(BadgeProps{
		Label:   "Default",
		Variant: "",
	})
	require.NotNil(t, badge)

	badge.Init()
	view := badge.View()

	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Default")
}
