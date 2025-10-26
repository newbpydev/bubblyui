package bubbly

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewComponent tests the NewComponent constructor function.
func TestNewComponent(t *testing.T) {
	tests := []struct {
		name          string
		componentName string
		wantName      string
	}{
		{
			name:          "creates builder with simple name",
			componentName: "Button",
			wantName:      "Button",
		},
		{
			name:          "creates builder with compound name",
			componentName: "FormInput",
			wantName:      "FormInput",
		},
		{
			name:          "creates builder with empty name",
			componentName: "",
			wantName:      "",
		},
		{
			name:          "creates builder with special characters",
			componentName: "My-Component_123",
			wantName:      "My-Component_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			builder := NewComponent(tt.componentName)

			// Assert
			require.NotNil(t, builder, "NewComponent should return non-nil builder")
			assert.NotNil(t, builder.component, "Builder should have component reference")
			assert.Equal(t, tt.wantName, builder.component.name, "Component name should match")
			assert.NotEmpty(t, builder.component.id, "Component should have unique ID")
			assert.NotNil(t, builder.errors, "Builder should have errors slice")
			assert.Empty(t, builder.errors, "Builder should start with no errors")
		})
	}
}

// TestComponentBuilder_Structure tests the ComponentBuilder struct fields.
func TestComponentBuilder_Structure(t *testing.T) {
	t.Run("builder stores component reference", func(t *testing.T) {
		// Arrange & Act
		builder := NewComponent("Test")

		// Assert
		require.NotNil(t, builder.component)
		assert.IsType(t, &componentImpl{}, builder.component)
	})

	t.Run("builder initializes error tracking", func(t *testing.T) {
		// Arrange & Act
		builder := NewComponent("Test")

		// Assert
		require.NotNil(t, builder.errors)
		assert.Empty(t, builder.errors)
		assert.Equal(t, 0, len(builder.errors))
	})

	t.Run("component has initialized fields", func(t *testing.T) {
		// Arrange & Act
		builder := NewComponent("Test")

		// Assert
		c := builder.component
		assert.NotEmpty(t, c.id, "Component should have ID")
		assert.NotNil(t, c.state, "State map should be initialized")
		assert.NotNil(t, c.handlers, "Handlers map should be initialized")
		assert.NotNil(t, c.children, "Children slice should be initialized")
	})
}

// TestComponentBuilder_UniqueIDs tests that each builder creates components with unique IDs.
func TestComponentBuilder_UniqueIDs(t *testing.T) {
	t.Run("generates unique IDs for multiple components", func(t *testing.T) {
		// Arrange
		count := 10
		ids := make(map[string]bool)

		// Act
		for i := 0; i < count; i++ {
			builder := NewComponent("Test")
			id := builder.component.id

			// Assert uniqueness
			assert.False(t, ids[id], "ID %s should be unique", id)
			ids[id] = true
		}

		// Assert
		assert.Equal(t, count, len(ids), "Should have %d unique IDs", count)
	})

	t.Run("IDs follow expected format", func(t *testing.T) {
		// Arrange & Act
		builder := NewComponent("Test")
		id := builder.component.id

		// Assert
		assert.Regexp(t, `^component-\d+$`, id, "ID should match format 'component-N'")
	})
}

// TestComponentBuilder_ErrorTracking tests error tracking functionality.
func TestComponentBuilder_ErrorTracking(t *testing.T) {
	t.Run("errors slice is mutable", func(t *testing.T) {
		// Arrange
		builder := NewComponent("Test")

		// Act - manually add error for testing
		builder.errors = append(builder.errors, assert.AnError)

		// Assert
		assert.Len(t, builder.errors, 1)
		assert.Equal(t, assert.AnError, builder.errors[0])
	})

	t.Run("multiple errors can be tracked", func(t *testing.T) {
		// Arrange
		builder := NewComponent("Test")

		// Act
		builder.errors = append(builder.errors, assert.AnError)
		builder.errors = append(builder.errors, assert.AnError)

		// Assert
		assert.Len(t, builder.errors, 2)
	})
}

// TestComponentBuilder_Concurrency tests thread-safety of component creation.
func TestComponentBuilder_Concurrency(t *testing.T) {
	t.Run("concurrent component creation is safe", func(t *testing.T) {
		// Arrange
		count := 100
		done := make(chan string, count)

		// Act - create components concurrently
		for i := 0; i < count; i++ {
			go func() {
				builder := NewComponent("Test")
				done <- builder.component.id
			}()
		}

		// Collect IDs
		ids := make(map[string]bool)
		for i := 0; i < count; i++ {
			id := <-done
			ids[id] = true
		}

		// Assert
		assert.Equal(t, count, len(ids), "All IDs should be unique even with concurrent creation")
	})
}
