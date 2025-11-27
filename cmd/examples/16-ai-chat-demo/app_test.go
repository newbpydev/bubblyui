package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	localComponents "github.com/newbpydev/bubblyui/cmd/examples/16-ai-chat-demo/components"
	localComposables "github.com/newbpydev/bubblyui/cmd/examples/16-ai-chat-demo/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

func TestCreateApp(t *testing.T) {
	app, err := CreateApp()
	require.NoError(t, err)
	require.NotNil(t, app)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(app)
	defer ct.Unmount()

	// Should render the header
	ct.AssertRenderContains("BubblyGPT")
}

func TestMessageList(t *testing.T) {
	comp, err := localComponents.CreateMessageList()
	require.NoError(t, err)
	require.NotNil(t, comp)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(comp)
	defer ct.Unmount()

	// Should render chat header
	ct.AssertRenderContains("Messages")
}

func TestChatSidebar(t *testing.T) {
	comp, err := localComponents.CreateChatSidebar()
	require.NoError(t, err)
	require.NotNil(t, comp)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(comp)
	defer ct.Unmount()

	// Should render sidebar elements
	ct.AssertRenderContains("Conversations")
	ct.AssertRenderContains("New Chat")
}

func TestChatInput(t *testing.T) {
	comp, err := localComponents.CreateChatInput()
	require.NoError(t, err)
	require.NotNil(t, comp)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(comp)
	defer ct.Unmount()

	// Should render input elements
	ct.AssertRenderContains("Send")
}

func TestChatComposable_SendMessage(t *testing.T) {
	// Create a minimal component to get a context
	app, err := CreateApp()
	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(app)
	defer ct.Unmount()

	// Get the shared chat state
	chat := localComposables.UseSharedChat(nil)
	require.NotNil(t, chat)

	// Initial state
	messages := chat.Messages.GetTyped()
	initialCount := len(messages)
	assert.GreaterOrEqual(t, initialCount, 1) // Welcome message

	// Send a message
	chat.SendMessage("Hello!")

	// Should have added user message and AI placeholder
	messages = chat.Messages.GetTyped()
	assert.Equal(t, initialCount+2, len(messages))

	// Last message should be AI typing
	lastMsg := messages[len(messages)-1]
	assert.Equal(t, localComposables.RoleAssistant, lastMsg.Role)
	assert.True(t, lastMsg.IsTyping)
}

func TestChatComposable_TypeNextChar(t *testing.T) {
	app, err := CreateApp()
	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(app)
	defer ct.Unmount()

	chat := localComposables.UseSharedChat(nil)
	require.NotNil(t, chat)

	// Send a message to trigger AI response
	chat.SendMessage("hi")

	// Type some characters
	for i := 0; i < 10; i++ {
		chat.TypeNextChar()
	}

	// AI message should have some content now
	messages := chat.Messages.GetTyped()
	lastMsg := messages[len(messages)-1]
	assert.NotEmpty(t, lastMsg.Content)
}

func TestWindowSizeComposable(t *testing.T) {
	tests := []struct {
		name               string
		width              int
		height             int
		expectedBreakpoint localComposables.Breakpoint
		expectedSidebar    bool
	}{
		{
			name:               "small terminal",
			width:              70,
			height:             24,
			expectedBreakpoint: localComposables.BreakpointSM,
			expectedSidebar:    false,
		},
		{
			name:               "medium terminal",
			width:              100,
			height:             30,
			expectedBreakpoint: localComposables.BreakpointMD,
			expectedSidebar:    true,
		},
		{
			name:               "large terminal",
			width:              140,
			height:             40,
			expectedBreakpoint: localComposables.BreakpointLG,
			expectedSidebar:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := localComposables.UseWindowSize(nil)
			ws.SetSize(tt.width, tt.height)

			assert.Equal(t, tt.expectedBreakpoint, ws.Breakpoint.GetTyped())
			assert.Equal(t, tt.expectedSidebar, ws.SidebarVisible.GetTyped())
		})
	}
}

func TestMessageRoles(t *testing.T) {
	assert.Equal(t, localComposables.MessageRole("user"), localComposables.RoleUser)
	assert.Equal(t, localComposables.MessageRole("assistant"), localComposables.RoleAssistant)
	assert.Equal(t, localComposables.MessageRole("system"), localComposables.RoleSystem)
}

func TestChatComposable_Scroll(t *testing.T) {
	app, err := CreateApp()
	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(app)
	defer ct.Unmount()

	chat := localComposables.UseSharedChat(nil)
	require.NotNil(t, chat)

	// Add some messages
	chat.SendMessage("Message 1")
	chat.SendMessage("Message 2")
	chat.SendMessage("Message 3")

	// Scroll to bottom
	chat.ScrollToBottom()
	offset := chat.ScrollOffset.GetTyped()
	assert.Greater(t, offset, 0)

	// Scroll up
	chat.ScrollUp()
	newOffset := chat.ScrollOffset.GetTyped()
	assert.Equal(t, offset-1, newOffset)

	// Scroll down
	chat.ScrollDown()
	finalOffset := chat.ScrollOffset.GetTyped()
	assert.Equal(t, offset, finalOffset)
}
