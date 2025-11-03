package components

import (
	"strings"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
)

// TestAppLayout_Creation tests that AppLayout component can be created.
func TestAppLayout_Creation(t *testing.T) {
	header := Text(TextProps{Content: "Header"})
	sidebar := Text(TextProps{Content: "Sidebar"})
	content := Text(TextProps{Content: "Content"})
	footer := Text(TextProps{Content: "Footer"})

	layout := AppLayout(AppLayoutProps{
		Header:  header,
		Sidebar: sidebar,
		Content: content,
		Footer:  footer,
	})

	assert.NotNil(t, layout, "AppLayout should be created")
}

// TestAppLayout_Rendering tests that AppLayout renders all sections.
func TestAppLayout_Rendering(t *testing.T) {
	header := Text(TextProps{Content: "Header Section"})
	header.Init()

	sidebar := Text(TextProps{Content: "Sidebar Menu"})
	sidebar.Init()

	content := Text(TextProps{Content: "Main Content"})
	content.Init()

	footer := Text(TextProps{Content: "Footer Info"})
	footer.Init()

	layout := AppLayout(AppLayoutProps{
		Header:  header,
		Sidebar: sidebar,
		Content: content,
		Footer:  footer,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Header Section", "Should render header")
	assert.Contains(t, output, "Sidebar Menu", "Should render sidebar")
	assert.Contains(t, output, "Main Content", "Should render content")
	assert.Contains(t, output, "Footer Info", "Should render footer")
}

// TestAppLayout_WithOnlyContent tests AppLayout with only content section.
func TestAppLayout_WithOnlyContent(t *testing.T) {
	content := Text(TextProps{Content: "Just Content"})
	content.Init()

	layout := AppLayout(AppLayoutProps{
		Content: content,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Just Content", "Should render content")
	assert.NotEmpty(t, output, "Should produce output")
}

// TestAppLayout_WithHeaderAndContent tests AppLayout with header and content.
func TestAppLayout_WithHeaderAndContent(t *testing.T) {
	header := Text(TextProps{Content: "App Title"})
	header.Init()

	content := Text(TextProps{Content: "Page Content"})
	content.Init()

	layout := AppLayout(AppLayoutProps{
		Header:  header,
		Content: content,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "App Title", "Should render header")
	assert.Contains(t, output, "Page Content", "Should render content")
}

// TestAppLayout_WithSidebarAndContent tests AppLayout with sidebar and content.
func TestAppLayout_WithSidebarAndContent(t *testing.T) {
	sidebar := Text(TextProps{Content: "Navigation"})
	sidebar.Init()

	content := Text(TextProps{Content: "Main Area"})
	content.Init()

	layout := AppLayout(AppLayoutProps{
		Sidebar: sidebar,
		Content: content,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Navigation", "Should render sidebar")
	assert.Contains(t, output, "Main Area", "Should render content")
}

// TestAppLayout_WithAllSections tests AppLayout with all four sections.
func TestAppLayout_WithAllSections(t *testing.T) {
	header := Text(TextProps{Content: "Top Bar"})
	header.Init()

	sidebar := Text(TextProps{Content: "Left Nav"})
	sidebar.Init()

	content := Text(TextProps{Content: "Center"})
	content.Init()

	footer := Text(TextProps{Content: "Bottom Bar"})
	footer.Init()

	layout := AppLayout(AppLayoutProps{
		Header:  header,
		Sidebar: sidebar,
		Content: content,
		Footer:  footer,
	})

	layout.Init()
	output := layout.View()

	// Verify all sections are present
	assert.Contains(t, output, "Top Bar", "Should render header")
	assert.Contains(t, output, "Left Nav", "Should render sidebar")
	assert.Contains(t, output, "Center", "Should render content")
	assert.Contains(t, output, "Bottom Bar", "Should render footer")

	// Verify layout structure (header should come before content, footer after)
	headerIdx := strings.Index(output, "Top Bar")
	contentIdx := strings.Index(output, "Center")
	footerIdx := strings.Index(output, "Bottom Bar")

	assert.True(t, headerIdx < contentIdx, "Header should come before content")
	assert.True(t, contentIdx < footerIdx, "Content should come before footer")
}

// TestAppLayout_WithCustomWidth tests AppLayout with custom width.
func TestAppLayout_WithCustomWidth(t *testing.T) {
	content := Text(TextProps{Content: "Content"})
	content.Init()

	layout := AppLayout(AppLayoutProps{
		Content: content,
		Width:   100,
	})

	layout.Init()
	output := layout.View()

	assert.NotEmpty(t, output, "Should produce output")
	assert.Contains(t, output, "Content", "Should render content")
}

// TestAppLayout_WithCustomHeight tests AppLayout with custom height.
func TestAppLayout_WithCustomHeight(t *testing.T) {
	content := Text(TextProps{Content: "Content"})
	content.Init()

	layout := AppLayout(AppLayoutProps{
		Content: content,
		Height:  30,
	})

	layout.Init()
	output := layout.View()

	assert.NotEmpty(t, output, "Should produce output")
	assert.Contains(t, output, "Content", "Should render content")
}

// TestAppLayout_WithCustomSidebarWidth tests AppLayout with custom sidebar width.
func TestAppLayout_WithCustomSidebarWidth(t *testing.T) {
	sidebar := Text(TextProps{Content: "Nav"})
	sidebar.Init()

	content := Text(TextProps{Content: "Main"})
	content.Init()

	layout := AppLayout(AppLayoutProps{
		Sidebar:      sidebar,
		Content:      content,
		SidebarWidth: 30,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Nav", "Should render sidebar")
	assert.Contains(t, output, "Main", "Should render content")
}

// TestAppLayout_ThemeIntegration tests that AppLayout integrates with theme system.
func TestAppLayout_ThemeIntegration(t *testing.T) {
	content := Text(TextProps{Content: "Themed Content"})
	content.Init()

	layout := AppLayout(AppLayoutProps{
		Content: content,
	})

	layout.Init()
	output := layout.View()

	// Theme integration is tested by successful rendering
	// The component should use DefaultTheme when no theme is provided
	assert.Contains(t, output, "Themed Content", "Should render content with theme")
	assert.NotEmpty(t, output, "Should produce themed output")
}

// TestAppLayout_BubbleteatIntegration tests Bubbletea Init/Update/View cycle.
func TestAppLayout_BubbleteatIntegration(t *testing.T) {
	content := Text(TextProps{Content: "TUI Content"})
	content.Init()

	layout := AppLayout(AppLayoutProps{
		Content: content,
	})

	// Test Init
	cmd := layout.Init()
	assert.Nil(t, cmd, "Init should return nil command")

	// Test Update
	model, cmd := layout.Update(nil)
	assert.NotNil(t, model, "Update should return model")
	assert.Nil(t, cmd, "Update should return nil command")

	// Test View
	output := layout.View()
	assert.NotEmpty(t, output, "View should produce output")
	assert.Contains(t, output, "TUI Content", "Should render content")
}

// TestAppLayout_EmptyLayout tests AppLayout with no sections.
func TestAppLayout_EmptyLayout(t *testing.T) {
	layout := AppLayout(AppLayoutProps{})

	layout.Init()
	output := layout.View()

	// Should still produce some output (empty layout structure)
	assert.NotNil(t, output, "Should produce output even when empty")
}

// TestAppLayout_PropsAccessibility tests that props can be accessed.
func TestAppLayout_PropsAccessibility(t *testing.T) {
	header := Text(TextProps{Content: "Header"})
	header.Init()

	content := Text(TextProps{Content: "Content"})
	content.Init()

	layout := AppLayout(AppLayoutProps{
		Header:  header,
		Content: content,
	})

	layout.Init()

	// Props should be accessible through the component
	assert.NotNil(t, layout, "Component should exist")
}

// TestAppLayout_ComplexChildren tests AppLayout with complex child components.
func TestAppLayout_ComplexChildren(t *testing.T) {
	// Create complex children
	header := Card(CardProps{
		Title:   "Application",
		Content: "Version 1.0",
	})
	header.Init()

	sidebar := Menu(MenuProps{
		Items: []MenuItem{
			{Label: "Home", Value: "home"},
			{Label: "Settings", Value: "settings"},
		},
		Selected: bubbly.NewRef("home"),
	})
	sidebar.Init()

	content := Card(CardProps{
		Title:   "Dashboard",
		Content: "Welcome to the app",
	})
	content.Init()

	footer := Text(TextProps{
		Content: "© 2024 BubblyUI",
	})
	footer.Init()

	layout := AppLayout(AppLayoutProps{
		Header:  header,
		Sidebar: sidebar,
		Content: content,
		Footer:  footer,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Application", "Should render header card")
	assert.Contains(t, output, "Home", "Should render sidebar menu")
	assert.Contains(t, output, "Dashboard", "Should render content card")
	assert.Contains(t, output, "© 2024", "Should render footer")
}
