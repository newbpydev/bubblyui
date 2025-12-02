package composables

import (
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// NotificationType defines notification types.
// Each type represents a different severity or purpose of the notification.
type NotificationType string

const (
	// NotificationInfo is for informational notifications.
	NotificationInfo NotificationType = "info"

	// NotificationSuccess is for success notifications.
	NotificationSuccess NotificationType = "success"

	// NotificationWarning is for warning notifications.
	NotificationWarning NotificationType = "warning"

	// NotificationError is for error notifications.
	NotificationError NotificationType = "error"
)

// Default configuration values
const (
	// DefaultNotificationDuration is the default duration before auto-dismiss.
	DefaultNotificationDuration = 3 * time.Second

	// DefaultMaxNotifications is the default maximum number of notifications.
	DefaultMaxNotifications = 5
)

// Notification represents a toast notification.
// Each notification has a unique ID, type, title, message, duration, and creation time.
type Notification struct {
	// ID is the unique identifier for this notification.
	ID int

	// Type is the notification type (info, success, warning, error).
	Type NotificationType

	// Title is the notification title.
	Title string

	// Message is the notification message body.
	Message string

	// Duration is the time before auto-dismiss (0 means no auto-dismiss).
	Duration time.Duration

	// CreatedAt is the timestamp when the notification was created.
	CreatedAt time.Time
}

// notificationConfig holds configuration for UseNotification.
type notificationConfig struct {
	defaultDuration  time.Duration
	maxNotifications int
}

// NotificationOption configures UseNotification.
type NotificationOption func(*notificationConfig)

// WithDefaultDuration sets the default notification duration.
// This duration is used by convenience methods (Info, Success, Warning, Error).
//
// Example:
//
//	notif := UseNotification(ctx, WithDefaultDuration(10*time.Second))
//	notif.Info("Title", "Message") // Will auto-dismiss after 10 seconds
func WithDefaultDuration(d time.Duration) NotificationOption {
	return func(c *notificationConfig) {
		c.defaultDuration = d
	}
}

// WithMaxNotifications sets the maximum number of notifications.
// When this limit is reached, the oldest notification is dismissed.
//
// Example:
//
//	notif := UseNotification(ctx, WithMaxNotifications(3))
//	// Only 3 notifications will be visible at a time
func WithMaxNotifications(max int) NotificationOption {
	return func(c *notificationConfig) {
		if max > 0 {
			c.maxNotifications = max
		}
	}
}

// NotificationReturn is the return value of UseNotification.
// It provides toast notification management with auto-dismiss support.
type NotificationReturn struct {
	// Notifications is the active notification stack.
	// Notifications are ordered from oldest to newest.
	Notifications *bubbly.Ref[[]Notification]

	// nextID is the next notification ID to assign.
	nextID int

	// config holds the notification configuration.
	config notificationConfig

	// mu protects concurrent access to notification operations.
	mu sync.Mutex

	// timers holds auto-dismiss timers by notification ID.
	timers map[int]*time.Timer
}

// Show displays a notification with the specified type, title, message, and duration.
// If duration is 0, the notification will not auto-dismiss.
// If duration is positive, the notification will be automatically dismissed after the duration.
//
// Example:
//
//	notif.Show(NotificationInfo, "Title", "Message", 5*time.Second)
func (n *NotificationReturn) Show(ntype NotificationType, title, message string, duration time.Duration) {
	n.mu.Lock()

	// Generate unique ID
	n.nextID++
	id := n.nextID

	// Create notification
	notification := Notification{
		ID:        id,
		Type:      ntype,
		Title:     title,
		Message:   message,
		Duration:  duration,
		CreatedAt: time.Now(),
	}

	// Get current notifications
	current := n.Notifications.GetTyped()

	// Enforce max notifications (drop oldest)
	for len(current) >= n.config.maxNotifications {
		// Cancel timer for oldest notification
		oldestID := current[0].ID
		if timer, exists := n.timers[oldestID]; exists {
			timer.Stop()
			delete(n.timers, oldestID)
		}
		current = current[1:]
	}

	// Add new notification
	newNotifications := append(current, notification)

	// Set up auto-dismiss timer if duration > 0
	if duration > 0 {
		timer := time.AfterFunc(duration, func() {
			n.Dismiss(id)
		})
		n.timers[id] = timer
	}

	n.mu.Unlock()

	// Update ref outside lock to avoid deadlock with Watch
	n.Notifications.Set(newNotifications)
}

// Info shows an info notification with the default duration.
//
// Example:
//
//	notif.Info("Welcome", "Welcome to the application!")
func (n *NotificationReturn) Info(title, message string) {
	n.Show(NotificationInfo, title, message, n.config.defaultDuration)
}

// Success shows a success notification with the default duration.
//
// Example:
//
//	notif.Success("Saved", "Your changes have been saved.")
func (n *NotificationReturn) Success(title, message string) {
	n.Show(NotificationSuccess, title, message, n.config.defaultDuration)
}

// Warning shows a warning notification with the default duration.
//
// Example:
//
//	notif.Warning("Low Storage", "You are running low on storage space.")
func (n *NotificationReturn) Warning(title, message string) {
	n.Show(NotificationWarning, title, message, n.config.defaultDuration)
}

// Error shows an error notification with the default duration.
//
// Example:
//
//	notif.Error("Failed", "Could not save your changes.")
func (n *NotificationReturn) Error(title, message string) {
	n.Show(NotificationError, title, message, n.config.defaultDuration)
}

// Dismiss removes a specific notification by ID.
// If the notification has an auto-dismiss timer, it is canceled.
// If the ID does not exist, this is a no-op.
//
// Example:
//
//	notif.Dismiss(notificationID)
func (n *NotificationReturn) Dismiss(id int) {
	n.mu.Lock()

	// Cancel timer if exists
	if timer, exists := n.timers[id]; exists {
		timer.Stop()
		delete(n.timers, id)
	}

	// Get current notifications
	current := n.Notifications.GetTyped()

	// Find and remove notification
	var newNotifications []Notification
	for _, notif := range current {
		if notif.ID != id {
			newNotifications = append(newNotifications, notif)
		}
	}

	// If nothing was removed, just unlock and return
	if len(newNotifications) == len(current) {
		n.mu.Unlock()
		return
	}

	n.mu.Unlock()

	// Update ref outside lock
	n.Notifications.Set(newNotifications)
}

// DismissAll removes all notifications.
// All auto-dismiss timers are canceled.
//
// Example:
//
//	notif.DismissAll()
func (n *NotificationReturn) DismissAll() {
	n.mu.Lock()

	// Cancel all timers
	for id, timer := range n.timers {
		timer.Stop()
		delete(n.timers, id)
	}

	n.mu.Unlock()

	// Update ref outside lock
	n.Notifications.Set([]Notification{})
}

// UseNotification creates a notification composable.
// It provides toast notification management with auto-dismiss support.
//
// This composable is useful for:
//   - Displaying success/error messages
//   - Showing informational alerts
//   - Warning users about important events
//   - Providing feedback for user actions
//
// Parameters:
//   - ctx: The component context (can be nil for testing)
//   - opts: Optional configuration options
//
// Returns:
//   - *NotificationReturn: A struct containing the notifications ref and control methods
//
// Example - Basic usage:
//
//	Setup(func(ctx *bubbly.Context) {
//	    notif := composables.UseNotification(ctx)
//	    ctx.Expose("notifications", notif)
//
//	    ctx.On("saved", func(_ interface{}) {
//	        notif.Success("Saved", "Changes saved successfully")
//	    })
//
//	    ctx.On("error", func(data interface{}) {
//	        err := data.(error)
//	        notif.Error("Error", err.Error())
//	    })
//	})
//
// Example - Custom configuration:
//
//	notif := composables.UseNotification(ctx,
//	    composables.WithDefaultDuration(5*time.Second),
//	    composables.WithMaxNotifications(3),
//	)
//
// Example - Manual dismiss:
//
//	notif.Show(composables.NotificationInfo, "Loading", "Please wait...", 0)
//	// ... later ...
//	notif.Dismiss(id)
//
// Example - Rendering notifications:
//
//	Template(func(ctx bubbly.RenderContext) string {
//	    notif := ctx.Get("notifications").(*composables.NotificationReturn)
//	    notifications := notif.Notifications.GetTyped()
//
//	    var lines []string
//	    for _, n := range notifications {
//	        lines = append(lines, fmt.Sprintf("[%s] %s: %s", n.Type, n.Title, n.Message))
//	    }
//	    return strings.Join(lines, "\n")
//	})
//
// Integration with CreateShared:
//
//	var UseSharedNotifications = composables.CreateShared(
//	    func(ctx *bubbly.Context) *composables.NotificationReturn {
//	        return composables.UseNotification(ctx)
//	    },
//	)
//
// Thread Safety:
//
// UseNotification is thread-safe. All notification operations are synchronized with a mutex.
// The Notifications ref can be safely accessed from multiple goroutines.
//
// Cleanup:
//
// All auto-dismiss timers are automatically canceled when the component unmounts.
func UseNotification(ctx *bubbly.Context, opts ...NotificationOption) *NotificationReturn {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseNotification", time.Since(start))
	}()

	// Initialize config with defaults
	config := notificationConfig{
		defaultDuration:  DefaultNotificationDuration,
		maxNotifications: DefaultMaxNotifications,
	}

	// Apply options
	for _, opt := range opts {
		opt(&config)
	}

	// Create return struct
	notif := &NotificationReturn{
		Notifications: bubbly.NewRef([]Notification{}),
		nextID:        0,
		config:        config,
		timers:        make(map[int]*time.Timer),
	}

	// Register cleanup on unmount
	if ctx != nil {
		ctx.OnUnmounted(func() {
			notif.DismissAll()
		})
	}

	return notif
}
