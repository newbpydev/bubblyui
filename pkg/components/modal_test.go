package components

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
)

func TestModal_Creation(t *testing.T) {
	title := "Confirm Action"
	content := "Are you sure?"

	modal := Modal(ModalProps{
		Title:   title,
		Content: content,
		Visible: bubbly.NewRef(true),
	})

	assert.NotNil(t, modal, "Modal should be created")
}

func TestModal_Rendering(t *testing.T) {
	title := "Test Modal"
	content := "Modal content here"
	visible := bubbly.NewRef(true)

	modal := Modal(ModalProps{
		Title:   title,
		Content: content,
		Visible: visible,
	})

	modal.Init()
	output := modal.View()

	assert.Contains(t, output, title, "Should render title")
	assert.Contains(t, output, content, "Should render content")
}

func TestModal_HiddenState(t *testing.T) {
	visible := bubbly.NewRef(false)

	modal := Modal(ModalProps{
		Title:   "Hidden",
		Content: "Should not see this",
		Visible: visible,
	})

	modal.Init()
	output := modal.View()

	assert.Empty(t, output, "Hidden modal should render nothing")
}

func TestModal_CloseEvent(t *testing.T) {
	closed := false
	visible := bubbly.NewRef(true)

	modal := Modal(ModalProps{
		Title:   "Test",
		Content: "Content",
		Visible: visible,
		OnClose: func() {
			closed = true
		},
	})

	modal.Init()
	// Manually emit close event (simulating Esc key handling by parent)
	modal.Emit("close", nil)

	assert.True(t, closed, "OnClose should be called")
	assert.False(t, visible.GetTyped(), "Visible should be set to false")
}

func TestModal_ConfirmEvent(t *testing.T) {
	confirmed := false

	modal := Modal(ModalProps{
		Title:   "Confirm",
		Content: "Confirm this action?",
		Visible: bubbly.NewRef(true),
		OnConfirm: func() {
			confirmed = true
		},
	})

	modal.Init()
	// Manually emit confirm event (simulating Enter key handling by parent)
	modal.Emit("confirm", nil)

	assert.True(t, confirmed, "OnConfirm should be called")
}

func TestModal_WithButtons(t *testing.T) {
	confirmBtn := Button(ButtonProps{
		Label:   "Confirm",
		Variant: ButtonPrimary,
	})
	confirmBtn.Init() // Initialize button before use

	cancelBtn := Button(ButtonProps{
		Label:   "Cancel",
		Variant: ButtonSecondary,
	})
	cancelBtn.Init() // Initialize button before use

	modal := Modal(ModalProps{
		Title:   "With Buttons",
		Content: "Choose an action",
		Visible: bubbly.NewRef(true),
		Buttons: []bubbly.Component{confirmBtn, cancelBtn},
	})

	modal.Init()
	output := modal.View()

	assert.Contains(t, output, "Confirm", "Should render confirm button")
	assert.Contains(t, output, "Cancel", "Should render cancel button")
}

func TestModal_ThemeIntegration(t *testing.T) {
	modal := Modal(ModalProps{
		Title:   "Themed",
		Content: "Content",
		Visible: bubbly.NewRef(true),
	})

	modal.Init()
	output := modal.View()

	assert.NotEmpty(t, output, "Should render with theme")
}

func TestModal_CustomStyle(t *testing.T) {
	customStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("99"))

	modal := Modal(ModalProps{
		Title:   "Custom",
		Content: "Styled content",
		Visible: bubbly.NewRef(true),
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})

	modal.Init()
	output := modal.View()

	assert.NotEmpty(t, output, "Should render with custom style")
}

func TestModal_Width(t *testing.T) {
	modal := Modal(ModalProps{
		Title:   "Wide Modal",
		Content: "This is a wider modal",
		Visible: bubbly.NewRef(true),
		Width:   60,
	})

	modal.Init()
	output := modal.View()

	assert.NotEmpty(t, output, "Should render with custom width")
}

func TestModal_BubbleteatIntegration(t *testing.T) {
	modal := Modal(ModalProps{
		Title:   "Integration",
		Content: "Test",
		Visible: bubbly.NewRef(true),
	})

	// Test Init
	cmd := modal.Init()
	assert.Nil(t, cmd, "Init should return nil command")

	// Test Update
	newModel, cmd := modal.Update(nil)
	assert.NotNil(t, newModel, "Update should return model")
	assert.Nil(t, cmd, "Update should return nil command for nil msg")

	// Test View
	output := modal.View()
	assert.NotEmpty(t, output, "View should return output")
}

func TestModal_EmptyContent(t *testing.T) {
	modal := Modal(ModalProps{
		Title:   "Empty",
		Content: "",
		Visible: bubbly.NewRef(true),
	})

	modal.Init()
	output := modal.View()

	assert.Contains(t, output, "Empty", "Should still render title")
}

func TestModal_LongContent(t *testing.T) {
	longContent := "This is a very long content that should be properly wrapped and displayed in the modal. "
	longContent += "It contains multiple sentences and should handle line breaks appropriately."

	modal := Modal(ModalProps{
		Title:   "Long Content",
		Content: longContent,
		Visible: bubbly.NewRef(true),
		Width:   40,
	})

	modal.Init()
	output := modal.View()

	assert.Contains(t, output, "Long Content", "Should render title")
	assert.NotEmpty(t, output, "Should render long content")
}

func TestModal_ToggleVisibility(t *testing.T) {
	visible := bubbly.NewRef(true)

	modal := Modal(ModalProps{
		Title:   "Toggle",
		Content: "Test visibility",
		Visible: visible,
	})

	modal.Init()

	// Initially visible
	output := modal.View()
	assert.NotEmpty(t, output, "Should be visible initially")

	// Hide
	visible.Set(false)
	output = modal.View()
	assert.Empty(t, output, "Should be hidden after toggle")

	// Show again
	visible.Set(true)
	output = modal.View()
	assert.NotEmpty(t, output, "Should be visible after toggle back")
}
