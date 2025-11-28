package composables

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestUseNotification_InitialState tests that UseNotification initializes correctly
func TestUseNotification_InitialState(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx)

	require.NotNil(t, notif, "UseNotification should return non-nil")
	require.NotNil(t, notif.Notifications, "Notifications should not be nil")

	// Initial notifications should be empty
	assert.Empty(t, notif.Notifications.GetTyped(),
		"Initial notifications should be empty")
}

// TestUseNotification_ShowAddsNotification tests that Show adds a notification
func TestUseNotification_ShowAddsNotification(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx)

	beforeShow := time.Now()
	notif.Show(NotificationInfo, "Title", "Message", 5*time.Second)
	afterShow := time.Now()

	notifications := notif.Notifications.GetTyped()
	require.Len(t, notifications, 1, "Should have one notification")

	n := notifications[0]
	assert.Equal(t, NotificationInfo, n.Type, "Type should match")
	assert.Equal(t, "Title", n.Title, "Title should match")
	assert.Equal(t, "Message", n.Message, "Message should match")
	assert.Equal(t, 5*time.Second, n.Duration, "Duration should match")
	assert.True(t, n.CreatedAt.After(beforeShow) || n.CreatedAt.Equal(beforeShow),
		"CreatedAt should be after or equal to beforeShow")
	assert.True(t, n.CreatedAt.Before(afterShow) || n.CreatedAt.Equal(afterShow),
		"CreatedAt should be before or equal to afterShow")
	assert.Greater(t, n.ID, 0, "ID should be positive")
}

// TestUseNotification_InfoConvenienceMethod tests Info() convenience method
func TestUseNotification_InfoConvenienceMethod(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx)

	notif.Info("Info Title", "Info Message")

	notifications := notif.Notifications.GetTyped()
	require.Len(t, notifications, 1)

	n := notifications[0]
	assert.Equal(t, NotificationInfo, n.Type)
	assert.Equal(t, "Info Title", n.Title)
	assert.Equal(t, "Info Message", n.Message)
}

// TestUseNotification_SuccessConvenienceMethod tests Success() convenience method
func TestUseNotification_SuccessConvenienceMethod(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx)

	notif.Success("Success Title", "Success Message")

	notifications := notif.Notifications.GetTyped()
	require.Len(t, notifications, 1)

	n := notifications[0]
	assert.Equal(t, NotificationSuccess, n.Type)
	assert.Equal(t, "Success Title", n.Title)
	assert.Equal(t, "Success Message", n.Message)
}

// TestUseNotification_WarningConvenienceMethod tests Warning() convenience method
func TestUseNotification_WarningConvenienceMethod(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx)

	notif.Warning("Warning Title", "Warning Message")

	notifications := notif.Notifications.GetTyped()
	require.Len(t, notifications, 1)

	n := notifications[0]
	assert.Equal(t, NotificationWarning, n.Type)
	assert.Equal(t, "Warning Title", n.Title)
	assert.Equal(t, "Warning Message", n.Message)
}

// TestUseNotification_ErrorConvenienceMethod tests Error() convenience method
func TestUseNotification_ErrorConvenienceMethod(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx)

	notif.Error("Error Title", "Error Message")

	notifications := notif.Notifications.GetTyped()
	require.Len(t, notifications, 1)

	n := notifications[0]
	assert.Equal(t, NotificationError, n.Type)
	assert.Equal(t, "Error Title", n.Title)
	assert.Equal(t, "Error Message", n.Message)
}

// TestUseNotification_DismissRemovesById tests Dismiss() removes notification by ID
func TestUseNotification_DismissRemovesById(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx)

	// Add multiple notifications
	notif.Info("First", "Message 1")
	notif.Info("Second", "Message 2")
	notif.Info("Third", "Message 3")

	notifications := notif.Notifications.GetTyped()
	require.Len(t, notifications, 3)

	// Get ID of second notification
	secondID := notifications[1].ID

	// Dismiss second notification
	notif.Dismiss(secondID)

	notifications = notif.Notifications.GetTyped()
	require.Len(t, notifications, 2, "Should have 2 notifications after dismiss")

	// Verify first and third remain
	assert.Equal(t, "First", notifications[0].Title)
	assert.Equal(t, "Third", notifications[1].Title)
}

// TestUseNotification_DismissNonExistentIsNoOp tests Dismiss() with non-existent ID
func TestUseNotification_DismissNonExistentIsNoOp(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx)

	notif.Info("Test", "Message")
	require.Len(t, notif.Notifications.GetTyped(), 1)

	// Dismiss non-existent ID
	notif.Dismiss(99999)

	// Should still have 1 notification
	assert.Len(t, notif.Notifications.GetTyped(), 1)
}

// TestUseNotification_DismissAllClearsAll tests DismissAll() clears all notifications
func TestUseNotification_DismissAllClearsAll(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx)

	// Add multiple notifications
	notif.Info("First", "Message 1")
	notif.Success("Second", "Message 2")
	notif.Warning("Third", "Message 3")
	notif.Error("Fourth", "Message 4")

	require.Len(t, notif.Notifications.GetTyped(), 4)

	// Dismiss all
	notif.DismissAll()

	assert.Empty(t, notif.Notifications.GetTyped(), "Should be empty after DismissAll")
}

// TestUseNotification_WithDefaultDuration tests default duration option
func TestUseNotification_WithDefaultDuration(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx, WithDefaultDuration(10*time.Second))

	// Use convenience method which uses default duration
	notif.Info("Test", "Message")

	notifications := notif.Notifications.GetTyped()
	require.Len(t, notifications, 1)

	assert.Equal(t, 10*time.Second, notifications[0].Duration,
		"Should use custom default duration")
}

// TestUseNotification_WithMaxNotifications tests max notifications option
func TestUseNotification_WithMaxNotifications(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx, WithMaxNotifications(3))

	// Add 5 notifications
	notif.Info("First", "1")
	notif.Info("Second", "2")
	notif.Info("Third", "3")
	notif.Info("Fourth", "4")
	notif.Info("Fifth", "5")

	notifications := notif.Notifications.GetTyped()
	require.Len(t, notifications, 3, "Should only have max 3 notifications")

	// Should have the 3 most recent (oldest dropped)
	assert.Equal(t, "Third", notifications[0].Title)
	assert.Equal(t, "Fourth", notifications[1].Title)
	assert.Equal(t, "Fifth", notifications[2].Title)
}

// TestUseNotification_AutoDismissAfterDuration tests auto-dismiss functionality
func TestUseNotification_AutoDismissAfterDuration(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx)

	// Add notification with short duration
	notif.Show(NotificationInfo, "Test", "Message", 50*time.Millisecond)

	require.Len(t, notif.Notifications.GetTyped(), 1, "Should have 1 notification initially")

	// Wait for auto-dismiss
	time.Sleep(100 * time.Millisecond)

	assert.Empty(t, notif.Notifications.GetTyped(), "Should be auto-dismissed after duration")
}

// TestUseNotification_ManualDismissCancelsAutoDismiss tests that manual dismiss cancels timer
func TestUseNotification_ManualDismissCancelsAutoDismiss(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx)

	// Add notification with longer duration
	notif.Show(NotificationInfo, "Test", "Message", 500*time.Millisecond)

	notifications := notif.Notifications.GetTyped()
	require.Len(t, notifications, 1)
	id := notifications[0].ID

	// Manually dismiss before auto-dismiss
	notif.Dismiss(id)

	assert.Empty(t, notif.Notifications.GetTyped(), "Should be dismissed")

	// Wait past original duration - should not panic or cause issues
	time.Sleep(100 * time.Millisecond)
}

// TestUseNotification_UniqueIDs tests that each notification gets a unique ID
func TestUseNotification_UniqueIDs(t *testing.T) {
	ctx := createTestContext()
	// Use max 10 to allow all notifications
	notif := UseNotification(ctx, WithMaxNotifications(10))

	// Add multiple notifications
	for i := 0; i < 10; i++ {
		notif.Info("Test", "Message")
	}

	notifications := notif.Notifications.GetTyped()
	require.Len(t, notifications, 10)

	// Check all IDs are unique
	ids := make(map[int]bool)
	for _, n := range notifications {
		assert.False(t, ids[n.ID], "ID %d should be unique", n.ID)
		ids[n.ID] = true
	}
}

// TestUseNotification_MultipleNotificationsStack tests notifications stack correctly
func TestUseNotification_MultipleNotificationsStack(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx)

	notif.Info("First", "1")
	notif.Success("Second", "2")
	notif.Warning("Third", "3")
	notif.Error("Fourth", "4")

	notifications := notif.Notifications.GetTyped()
	require.Len(t, notifications, 4)

	// Verify order (oldest first)
	assert.Equal(t, "First", notifications[0].Title)
	assert.Equal(t, "Second", notifications[1].Title)
	assert.Equal(t, "Third", notifications[2].Title)
	assert.Equal(t, "Fourth", notifications[3].Title)

	// Verify types
	assert.Equal(t, NotificationInfo, notifications[0].Type)
	assert.Equal(t, NotificationSuccess, notifications[1].Type)
	assert.Equal(t, NotificationWarning, notifications[2].Type)
	assert.Equal(t, NotificationError, notifications[3].Type)
}

// TestUseNotification_NotificationsAreReactive tests that Notifications ref is reactive
func TestUseNotification_NotificationsAreReactive(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx)

	// Track changes via Watch
	changeCount := 0
	bubbly.Watch(notif.Notifications, func(newVal, oldVal []Notification) {
		changeCount++
	})

	// Each notification should trigger watcher
	notif.Info("Test", "Message")
	assert.Equal(t, 1, changeCount, "Show should trigger watcher")

	// Dismiss should trigger watcher
	notifications := notif.Notifications.GetTyped()
	notif.Dismiss(notifications[0].ID)
	assert.Equal(t, 2, changeCount, "Dismiss should trigger watcher")
}

// TestUseNotification_WorksWithCreateShared tests shared composable pattern
func TestUseNotification_WorksWithCreateShared(t *testing.T) {
	sharedNotif := CreateShared(func(ctx *bubbly.Context) *NotificationReturn {
		return UseNotification(ctx)
	})

	ctx := createTestContext()
	notif1 := sharedNotif(ctx)
	notif2 := sharedNotif(ctx)

	// Both should be the same instance
	notif1.Info("from notif1", "message")

	notifications := notif2.Notifications.GetTyped()
	assert.Len(t, notifications, 1, "Shared instance should have same notifications")
	assert.Equal(t, "from notif1", notifications[0].Title)
}

// TestUseNotification_NilContext tests that UseNotification works with nil context
func TestUseNotification_NilContext(t *testing.T) {
	// Should not panic with nil context
	assert.NotPanics(t, func() {
		notif := UseNotification(nil)
		assert.NotNil(t, notif)
		notif.Info("Test", "Message")
	})
}

// TestUseNotification_ConcurrentAccess tests thread safety
func TestUseNotification_ConcurrentAccess(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx, WithMaxNotifications(100))

	// Run concurrent notifications
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 10; j++ {
				notif.Info("Test", "Message")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have notifications (exact count may vary due to max limit)
	notifications := notif.Notifications.GetTyped()
	assert.LessOrEqual(t, len(notifications), 100, "Should respect max limit")
	assert.Greater(t, len(notifications), 0, "Should have some notifications")
}

// TestUseNotification_ZeroDurationNoAutoDismiss tests that zero duration means no auto-dismiss
func TestUseNotification_ZeroDurationNoAutoDismiss(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx)

	// Add notification with zero duration
	notif.Show(NotificationInfo, "Test", "Message", 0)

	require.Len(t, notif.Notifications.GetTyped(), 1)

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Should still be there (no auto-dismiss for zero duration)
	assert.Len(t, notif.Notifications.GetTyped(), 1,
		"Zero duration should not auto-dismiss")
}

// TestNotificationType_Constants tests notification type constants
func TestNotificationType_Constants(t *testing.T) {
	assert.Equal(t, NotificationType("info"), NotificationInfo)
	assert.Equal(t, NotificationType("success"), NotificationSuccess)
	assert.Equal(t, NotificationType("warning"), NotificationWarning)
	assert.Equal(t, NotificationType("error"), NotificationError)
}

// TestUseNotification_DismissAllCancelsTimers tests that DismissAll cancels all timers
func TestUseNotification_DismissAllCancelsTimers(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx)

	// Add notifications with auto-dismiss
	notif.Show(NotificationInfo, "Test1", "Message", 500*time.Millisecond)
	notif.Show(NotificationInfo, "Test2", "Message", 500*time.Millisecond)

	require.Len(t, notif.Notifications.GetTyped(), 2)

	// Dismiss all
	notif.DismissAll()

	assert.Empty(t, notif.Notifications.GetTyped())

	// Wait past original duration - should not panic
	time.Sleep(100 * time.Millisecond)
}

// TestUseNotification_DefaultDurationValue tests default duration when not specified
func TestUseNotification_DefaultDurationValue(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx)

	notif.Info("Test", "Message")

	notifications := notif.Notifications.GetTyped()
	require.Len(t, notifications, 1)

	// Default duration should be 3 seconds
	assert.Equal(t, 3*time.Second, notifications[0].Duration,
		"Default duration should be 3 seconds")
}

// TestUseNotification_DefaultMaxNotifications tests default max when not specified
func TestUseNotification_DefaultMaxNotifications(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx)

	// Add more than default max (5)
	for i := 0; i < 10; i++ {
		notif.Info("Test", "Message")
	}

	notifications := notif.Notifications.GetTyped()
	assert.Len(t, notifications, 5, "Default max should be 5")
}

// TestUseNotification_CombinedOptions tests multiple options together
func TestUseNotification_CombinedOptions(t *testing.T) {
	ctx := createTestContext()
	notif := UseNotification(ctx,
		WithDefaultDuration(10*time.Second),
		WithMaxNotifications(2),
	)

	// Add 3 notifications
	notif.Info("First", "1")
	notif.Info("Second", "2")
	notif.Info("Third", "3")

	notifications := notif.Notifications.GetTyped()
	require.Len(t, notifications, 2, "Should respect max")

	// Should have custom duration
	assert.Equal(t, 10*time.Second, notifications[0].Duration)
	assert.Equal(t, 10*time.Second, notifications[1].Duration)

	// Should have most recent
	assert.Equal(t, "Second", notifications[0].Title)
	assert.Equal(t, "Third", notifications[1].Title)
}
