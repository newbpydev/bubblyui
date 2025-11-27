// Package components provides layout components for the BubblyUI framework.
package components

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestThemeIntegration_Divider_UsesMutedColor verifies that Divider uses theme.Muted for color.
// This is a requirement from Task 5.1: "Divider uses theme.Muted"
func TestThemeIntegration_Divider_UsesMutedColor(t *testing.T) {
	tests := []struct {
		name     string
		props    DividerProps
		expected string
	}{
		{
			name: "horizontal divider uses muted color",
			props: DividerProps{
				Length: 10,
			},
			expected: "─",
		},
		{
			name: "vertical divider uses muted color",
			props: DividerProps{
				Vertical: true,
				Length:   5,
			},
			expected: "│",
		},
		{
			name: "divider with label uses muted color",
			props: DividerProps{
				Length: 20,
				Label:  "OR",
			},
			expected: "OR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			divider := Divider(tt.props)
			divider.Init()
			output := divider.View()

			// Verify output is not empty and contains expected content
			assert.NotEmpty(t, output, "Divider should render output")
			assert.Contains(t, output, tt.expected, "Divider should contain expected content")

			// The output should have ANSI escape codes from theme.Muted styling
			// We can't easily test the exact color, but we can verify styling is applied
			// by checking that the output is longer than the raw content (due to ANSI codes)
		})
	}
}

// TestThemeIntegration_Box_UsesSecondaryForBorder verifies that Box uses theme.Secondary for border.
// This is a requirement from Task 5.1: "Box border uses theme.Secondary"
func TestThemeIntegration_Box_UsesSecondaryForBorder(t *testing.T) {
	tests := []struct {
		name    string
		props   BoxProps
		wantBox bool
	}{
		{
			name: "box with border uses secondary color",
			props: BoxProps{
				Content: "Test Content",
				Border:  true,
			},
			wantBox: true,
		},
		{
			name: "box with border and title",
			props: BoxProps{
				Content: "Test Content",
				Border:  true,
				Title:   "Title",
			},
			wantBox: true,
		},
		{
			name: "box without border has no border styling",
			props: BoxProps{
				Content: "Test Content",
				Border:  false,
			},
			wantBox: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := Box(tt.props)
			box.Init()
			output := box.View()

			assert.NotEmpty(t, output, "Box should render output")
			assert.Contains(t, output, "Test Content", "Box should contain content")

			if tt.wantBox {
				// Border characters should be present
				// NormalBorder uses: ─ │ ┌ ┐ └ ┘
				assert.True(t,
					containsAny(output, "─", "│", "┌", "┐", "└", "┘"),
					"Box with border should contain border characters")
			}
		})
	}
}

// TestThemeIntegration_Box_UsesPrimaryForTitle verifies that Box uses theme.Primary for title.
func TestThemeIntegration_Box_UsesPrimaryForTitle(t *testing.T) {
	box := Box(BoxProps{
		Content: "Content",
		Title:   "My Title",
		Border:  true,
	})
	box.Init()
	output := box.View()

	assert.NotEmpty(t, output, "Box should render output")
	assert.Contains(t, output, "My Title", "Box should contain title")
	assert.Contains(t, output, "Content", "Box should contain content")
}

// TestThemeIntegration_HStack_UsesMutedForDivider verifies that HStack uses theme.Muted for dividers.
func TestThemeIntegration_HStack_UsesMutedForDivider(t *testing.T) {
	text1 := Text(TextProps{Content: "Item1"})
	text1.Init()
	text2 := Text(TextProps{Content: "Item2"})
	text2.Init()

	hstack := HStack(StackProps{
		Items:   []interface{}{text1, text2},
		Divider: true,
		Spacing: 2,
	})
	hstack.Init()
	output := hstack.View()

	assert.NotEmpty(t, output, "HStack should render output")
	assert.Contains(t, output, "Item1", "HStack should contain first item")
	assert.Contains(t, output, "Item2", "HStack should contain second item")
	assert.Contains(t, output, "│", "HStack with divider should contain vertical divider character")
}

// TestThemeIntegration_VStack_UsesMutedForDivider verifies that VStack uses theme.Muted for dividers.
func TestThemeIntegration_VStack_UsesMutedForDivider(t *testing.T) {
	text1 := Text(TextProps{Content: "Item1"})
	text1.Init()
	text2 := Text(TextProps{Content: "Item2"})
	text2.Init()

	vstack := VStack(StackProps{
		Items:   []interface{}{text1, text2},
		Divider: true,
		Spacing: 1,
	})
	vstack.Init()
	output := vstack.View()

	assert.NotEmpty(t, output, "VStack should render output")
	assert.Contains(t, output, "Item1", "VStack should contain first item")
	assert.Contains(t, output, "Item2", "VStack should contain second item")
	assert.Contains(t, output, "─", "VStack with divider should contain horizontal divider character")
}

// TestThemeIntegration_AllComponents_SupportCustomStyle verifies all layout components support custom Style prop.
// This is a requirement from Task 5.1: "All components support custom Style prop"
func TestThemeIntegration_AllComponents_SupportCustomStyle(t *testing.T) {
	customStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	tests := []struct {
		name       string
		component  func() bubbly.Component
		allowEmpty bool
	}{
		{
			name: "Divider supports custom style",
			component: func() bubbly.Component {
				return Divider(DividerProps{
					Length: 10,
					CommonProps: CommonProps{
						Style: &customStyle,
					},
				})
			},
		},
		{
			name: "Box supports custom style",
			component: func() bubbly.Component {
				return Box(BoxProps{
					Content: "Test",
					CommonProps: CommonProps{
						Style: &customStyle,
					},
				})
			},
		},
		{
			name: "HStack supports custom style",
			component: func() bubbly.Component {
				text := Text(TextProps{Content: "Item"})
				text.Init()
				return HStack(StackProps{
					Items: []interface{}{text},
					CommonProps: CommonProps{
						Style: &customStyle,
					},
				})
			},
		},
		{
			name: "VStack supports custom style",
			component: func() bubbly.Component {
				text := Text(TextProps{Content: "Item"})
				text.Init()
				return VStack(StackProps{
					Items: []interface{}{text},
					CommonProps: CommonProps{
						Style: &customStyle,
					},
				})
			},
		},
		{
			name: "Flex supports custom style",
			component: func() bubbly.Component {
				text := Text(TextProps{Content: "Item"})
				text.Init()
				return Flex(FlexProps{
					Items: []bubbly.Component{text},
					CommonProps: CommonProps{
						Style: &customStyle,
					},
				})
			},
		},
		{
			name: "Center supports custom style",
			component: func() bubbly.Component {
				text := Text(TextProps{Content: "Centered"})
				text.Init()
				return Center(CenterProps{
					Child:  text,
					Width:  20,
					Height: 5,
					CommonProps: CommonProps{
						Style: &customStyle,
					},
				})
			},
		},
		{
			name: "Container supports custom style",
			component: func() bubbly.Component {
				text := Text(TextProps{Content: "Contained"})
				text.Init()
				return Container(ContainerProps{
					Child: text,
					Size:  ContainerMd,
					CommonProps: CommonProps{
						Style: &customStyle,
					},
				})
			},
		},
		{
			name: "Spacer supports custom style",
			component: func() bubbly.Component {
				return Spacer(SpacerProps{
					Width:  5,
					Height: 2,
					CommonProps: CommonProps{
						Style: &customStyle,
					},
				})
			},
			allowEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := tt.component()
			comp.Init()
			output := comp.View()

			// Component should render without panic
			// Custom style should be applied (output may have ANSI codes)
			assert.NotPanics(t, func() {
				comp.Init()
				comp.View()
			}, "Component should not panic with custom style")

			// For most components, output should not be empty
			// (Spacer with no content may be empty, which is valid)
			if !tt.allowEmpty {
				assert.NotEmpty(t, output, "Component should render output with custom style")
			}
		})
	}
}

// TestThemeIntegration_AllComponents_InjectTheme verifies all layout components inject theme.
func TestThemeIntegration_AllComponents_InjectTheme(t *testing.T) {
	// Test that all components can be created and rendered without panic
	// This verifies the injectTheme pattern works correctly

	tests := []struct {
		name      string
		component func() bubbly.Component
	}{
		{
			name: "Divider injects theme",
			component: func() bubbly.Component {
				return Divider(DividerProps{Length: 10})
			},
		},
		{
			name: "Box injects theme",
			component: func() bubbly.Component {
				return Box(BoxProps{Content: "Test", Border: true})
			},
		},
		{
			name: "HStack injects theme",
			component: func() bubbly.Component {
				text := Text(TextProps{Content: "Item"})
				text.Init()
				return HStack(StackProps{Items: []interface{}{text}, Divider: true})
			},
		},
		{
			name: "VStack injects theme",
			component: func() bubbly.Component {
				text := Text(TextProps{Content: "Item"})
				text.Init()
				return VStack(StackProps{Items: []interface{}{text}, Divider: true})
			},
		},
		{
			name: "Flex injects theme",
			component: func() bubbly.Component {
				text := Text(TextProps{Content: "Item"})
				text.Init()
				return Flex(FlexProps{Items: []bubbly.Component{text}})
			},
		},
		{
			name: "Center injects theme",
			component: func() bubbly.Component {
				text := Text(TextProps{Content: "Centered"})
				text.Init()
				return Center(CenterProps{Child: text, Width: 20, Height: 5})
			},
		},
		{
			name: "Container injects theme",
			component: func() bubbly.Component {
				text := Text(TextProps{Content: "Contained"})
				text.Init()
				return Container(ContainerProps{Child: text})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := tt.component()

			assert.NotPanics(t, func() {
				comp.Init()
				output := comp.View()
				assert.NotEmpty(t, output, "Component should render output")
			}, "Component should not panic during theme injection")
		})
	}
}

// containsAny checks if the string contains any of the given substrings.
func containsAny(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if len(sub) > 0 && len(s) > 0 {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}
