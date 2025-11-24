package components

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPageLayout_Creation tests that PageLayout component can be created.
func TestPageLayout_Creation(t *testing.T) {
	title := Text(TextProps{Content: "Page Title"})
	content := Text(TextProps{Content: "Page Content"})

	layout := PageLayout(PageLayoutProps{
		Title:   title,
		Content: content,
	})

	assert.NotNil(t, layout, "PageLayout should be created")
}

// TestPageLayout_Rendering tests that PageLayout renders all sections.
func TestPageLayout_Rendering(t *testing.T) {
	title := Text(TextProps{Content: "My Page"})
	title.Init()

	content := Text(TextProps{Content: "Page body content"})
	content.Init()

	actions := Text(TextProps{Content: "Actions"})
	actions.Init()

	layout := PageLayout(PageLayoutProps{
		Title:   title,
		Content: content,
		Actions: actions,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "My Page", "Should render title")
	assert.Contains(t, output, "Page body content", "Should render content")
	assert.Contains(t, output, "Actions", "Should render actions")
}

// TestPageLayout_WithOnlyContent tests PageLayout with only content.
func TestPageLayout_WithOnlyContent(t *testing.T) {
	content := Text(TextProps{Content: "Just content"})
	content.Init()

	layout := PageLayout(PageLayoutProps{
		Content: content,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Just content", "Should render content")
	assert.NotEmpty(t, output, "Should produce output")
}

// TestPageLayout_WithTitleAndContent tests PageLayout with title and content.
func TestPageLayout_WithTitleAndContent(t *testing.T) {
	title := Text(TextProps{Content: "Dashboard"})
	title.Init()

	content := Text(TextProps{Content: "Welcome to dashboard"})
	content.Init()

	layout := PageLayout(PageLayoutProps{
		Title:   title,
		Content: content,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Dashboard", "Should render title")
	assert.Contains(t, output, "Welcome to dashboard", "Should render content")

	// Title should come before content
	titleIdx := strings.Index(output, "Dashboard")
	contentIdx := strings.Index(output, "Welcome to dashboard")
	assert.True(t, titleIdx < contentIdx, "Title should come before content")
}

// TestPageLayout_WithAllSections tests PageLayout with all three sections.
func TestPageLayout_WithAllSections(t *testing.T) {
	title := Text(TextProps{Content: "Settings"})
	title.Init()

	content := Text(TextProps{Content: "Configuration options"})
	content.Init()

	actions := Button(ButtonProps{Label: "Save"})
	actions.Init()

	layout := PageLayout(PageLayoutProps{
		Title:   title,
		Content: content,
		Actions: actions,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Settings", "Should render title")
	assert.Contains(t, output, "Configuration options", "Should render content")
	assert.Contains(t, output, "Save", "Should render actions")

	// Verify section ordering
	titleIdx := strings.Index(output, "Settings")
	contentIdx := strings.Index(output, "Configuration options")
	actionsIdx := strings.Index(output, "Save")

	assert.True(t, titleIdx < contentIdx, "Title should come before content")
	assert.True(t, contentIdx < actionsIdx, "Content should come before actions")
}

// TestPageLayout_WithCustomWidth tests PageLayout with custom width.
func TestPageLayout_WithCustomWidth(t *testing.T) {
	content := Text(TextProps{Content: "Content"})
	content.Init()

	layout := PageLayout(PageLayoutProps{
		Content: content,
		Width:   100,
	})

	layout.Init()
	output := layout.View()

	assert.NotEmpty(t, output, "Should produce output")
	assert.Contains(t, output, "Content", "Should render content")
}

// TestPageLayout_WithSpacing tests PageLayout with custom spacing.
func TestPageLayout_WithSpacing(t *testing.T) {
	title := Text(TextProps{Content: "Title"})
	title.Init()

	content := Text(TextProps{Content: "Content"})
	content.Init()

	layout := PageLayout(PageLayoutProps{
		Title:   title,
		Content: content,
		Spacing: 3,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Title", "Should render title")
	assert.Contains(t, output, "Content", "Should render content")
}

// TestPageLayout_ThemeIntegration tests theme integration.
func TestPageLayout_ThemeIntegration(t *testing.T) {
	content := Text(TextProps{Content: "Themed page"})
	content.Init()

	layout := PageLayout(PageLayoutProps{
		Content: content,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Themed page", "Should render with theme")
	assert.NotEmpty(t, output, "Should produce themed output")
}

// TestPageLayout_BubbleteatIntegration tests Bubbletea integration.
func TestPageLayout_BubbleteatIntegration(t *testing.T) {
	content := Text(TextProps{Content: "TUI Page"})
	content.Init()

	layout := PageLayout(PageLayoutProps{
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
	assert.Contains(t, output, "TUI Page", "Should render content")
}

// TestPageLayout_EmptyLayout tests PageLayout with no sections.
func TestPageLayout_EmptyLayout(t *testing.T) {
	layout := PageLayout(PageLayoutProps{})

	layout.Init()
	output := layout.View()

	assert.NotNil(t, output, "Should produce output even when empty")
}

// TestPageLayout_ComplexChildren tests PageLayout with complex components.
func TestPageLayout_ComplexChildren(t *testing.T) {
	title := Card(CardProps{
		Title:   "User Profile",
		Content: "John Doe",
	})
	title.Init()

	content := Card(CardProps{
		Title:   "Details",
		Content: "Email: john@example.com\nPhone: 555-1234",
	})
	content.Init()

	actions := Button(ButtonProps{
		Label: "Edit Profile",
	})
	actions.Init()

	layout := PageLayout(PageLayoutProps{
		Title:   title,
		Content: content,
		Actions: actions,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "User Profile", "Should render title card")
	assert.Contains(t, output, "Details", "Should render content card")
	assert.Contains(t, output, "Edit Profile", "Should render action button")
}

// ============================================================================
// PAGE LAYOUT ADDITIONAL TESTS - Edge Cases
// ============================================================================

func TestPageLayout_TitleOnly(t *testing.T) {
	title := Text(TextProps{Content: "Page Title"})
	title.Init()

	layout := PageLayout(PageLayoutProps{
		Title: title,
		// No content or actions
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Page Title", "Should render title only")
}

func TestPageLayout_ContentOnly(t *testing.T) {
	content := Text(TextProps{Content: "Page Content"})
	content.Init()

	layout := PageLayout(PageLayoutProps{
		Content: content,
		// No title or actions
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Page Content", "Should render content only")
}

func TestPageLayout_ActionsOnly(t *testing.T) {
	actions := Button(ButtonProps{Label: "Action"})
	actions.Init()

	layout := PageLayout(PageLayoutProps{
		Actions: actions,
		// No title or content
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Action", "Should render actions only")
}

func TestPageLayout_EmptyLayout_NoPanic(t *testing.T) {
	layout := PageLayout(PageLayoutProps{
		// No children at all
	})

	layout.Init()
	output := layout.View()

	// Should render empty without panic
	assert.NotPanics(t, func() {
		_ = layout.View()
	})
	// Empty layout may produce empty or whitespace-only output
	_ = output // Just ensure it doesn't panic
}
