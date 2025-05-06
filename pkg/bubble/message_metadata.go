package bubble

import (
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Default priority for messages
const defaultPriority = 0

// MessageContext holds metadata and routing information for messages
type MessageContext struct {
	// Priority determines the order of processing (higher is processed first)
	Priority int

	// Handled indicates whether the message has been handled
	Handled bool

	// TargetPath is the path to the target component (e.g., "root/child1/child2")
	TargetPath string

	// SourcePath is the path to the component that sent the message
	SourcePath string

	// Timestamp when the message was created
	Timestamp time.Time

	// Metadata contains arbitrary key-value pairs associated with the message
	Metadata map[string]interface{}

	// mutex for thread safety
	mutex sync.RWMutex
}

// MessageWithContext wraps a tea.Msg with a MessageContext
type MessageWithContext struct {
	// The original message
	OriginalMsg tea.Msg

	// The message context
	Context *MessageContext
}

// MessageContextOption is a function that configures a MessageContext
type MessageContextOption func(*MessageContext)

// NewMessageContext creates a new message context with default values
func NewMessageContext() *MessageContext {
	return &MessageContext{
		Priority:   defaultPriority,
		Handled:    false,
		TargetPath: "",
		SourcePath: "",
		Timestamp:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}
}

// NewMessageContextWithOptions creates a new message context with the given options
func NewMessageContextWithOptions(options ...MessageContextOption) *MessageContext {
	ctx := NewMessageContext()
	for _, option := range options {
		option(ctx)
	}
	return ctx
}

// WithPriority sets the priority for a message context
func WithPriority(priority int) MessageContextOption {
	return func(ctx *MessageContext) {
		ctx.Priority = priority
	}
}

// WithTargetPath sets the target path for a message context
func WithTargetPath(targetPath string) MessageContextOption {
	return func(ctx *MessageContext) {
		ctx.TargetPath = targetPath
	}
}

// WithSourcePath sets the source path for a message context
func WithSourcePath(sourcePath string) MessageContextOption {
	return func(ctx *MessageContext) {
		ctx.SourcePath = sourcePath
	}
}

// WithTimestamp sets the timestamp for a message context
func WithTimestamp(timestamp time.Time) MessageContextOption {
	return func(ctx *MessageContext) {
		ctx.Timestamp = timestamp
	}
}

// WithMetadata sets the metadata for a message context
func WithMetadata(metadata map[string]interface{}) MessageContextOption {
	return func(ctx *MessageContext) {
		for k, v := range metadata {
			ctx.Metadata[k] = v
		}
	}
}

// GetMetadata retrieves a metadata value by key
func (ctx *MessageContext) GetMetadata(key string) (interface{}, bool) {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	value, ok := ctx.Metadata[key]
	return value, ok
}

// SetMetadata sets a metadata value by key
func (ctx *MessageContext) SetMetadata(key string, value interface{}) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	ctx.Metadata[key] = value
}

// NewMessageWithContext creates a new message with context
func NewMessageWithContext(msg tea.Msg, options ...MessageContextOption) *MessageWithContext {
	return &MessageWithContext{
		OriginalMsg: msg,
		Context:     NewMessageContextWithOptions(options...),
	}
}

// GetMessageContext attempts to extract a MessageContext from a tea.Msg
func GetMessageContext(msg tea.Msg) (*MessageContext, bool) {
	if msgWithCtx, ok := msg.(*MessageWithContext); ok {
		return msgWithCtx.Context, true
	}
	return nil, false
}

// DispatcherFunc represents a function that processes a message
type DispatcherFunc func(msg tea.Msg) tea.Cmd

// MessageMiddleware is a function that can process messages before they are handled
type MessageMiddleware func(msg tea.Msg, next DispatcherFunc) tea.Cmd

// MessageDispatcher is a central dispatcher for messages with middleware support
type MessageDispatcher struct {
	// Middleware chain
	middleware []MessageMiddleware

	// Message queue for async processing
	queue chan tea.Msg

	// Mutex for thread safety
	mutex sync.RWMutex
}

// NewMessageDispatcher creates a new message dispatcher
func NewMessageDispatcher() *MessageDispatcher {
	return &MessageDispatcher{
		middleware: make([]MessageMiddleware, 0),
		queue:      make(chan tea.Msg, 100), // Buffer size of 100 messages
	}
}

// Use adds middleware to the dispatcher
func (d *MessageDispatcher) Use(middleware MessageMiddleware) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.middleware = append(d.middleware, middleware)
}

// Dispatch sends a message through the middleware chain to the handler
func (d *MessageDispatcher) Dispatch(msg tea.Msg, handler DispatcherFunc) tea.Cmd {
	d.mutex.RLock()
	middleware := make([]MessageMiddleware, len(d.middleware))
	copy(middleware, d.middleware)
	d.mutex.RUnlock()

	// Create the final function that will be called after all middleware
	var finalFunc DispatcherFunc = handler

	// Build the middleware chain from the end to the start
	for i := len(middleware) - 1; i >= 0; i-- {
		currentMiddleware := middleware[i]
		nextFunc := finalFunc

		finalFunc = func(currentMsg tea.Msg) tea.Cmd {
			return currentMiddleware(currentMsg, nextFunc)
		}
	}

	// Start the middleware chain with the original message
	return finalFunc(msg)
}

// QueueMessage adds a message to the async processing queue
func (d *MessageDispatcher) QueueMessage(msg tea.Msg) {
	d.queue <- msg
}

// ProcessQueueAsync processes messages in the queue asynchronously
func (d *MessageDispatcher) ProcessQueueAsync(handler DispatcherFunc) {
	go func() {
		for msg := range d.queue {
			d.Dispatch(msg, handler)
		}
	}()
}

// ProcessQueue processes a single message from the queue synchronously
// Returns true if a message was processed, false if the queue is empty
func (d *MessageDispatcher) ProcessQueue(handler DispatcherFunc) bool {
	select {
	case msg := <-d.queue:
		d.Dispatch(msg, handler)
		return true
	default:
		return false
	}
}
