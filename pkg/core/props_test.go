package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test struct for Props
type TestProps struct {
	Name     string
	Count    int
	IsActive bool
	Tags     []string
	Nested   NestedProps
}

type NestedProps struct {
	Value string
	ID    int
}

func TestPropsBasics(t *testing.T) {
	t.Run("Create and Access", func(t *testing.T) {
		// Create props with initial values
		props := NewProps(TestProps{
			Name:     "Test Component",
			Count:    42,
			IsActive: true,
			Tags:     []string{"tag1", "tag2"},
			Nested: NestedProps{
				Value: "inner",
				ID:    100,
			},
		})

		// Get values
		values := props.Get()

		// Verify values
		assert.Equal(t, "Test Component", values.Name)
		assert.Equal(t, 42, values.Count)
		assert.Equal(t, true, values.IsActive)
		assert.Equal(t, []string{"tag1", "tag2"}, values.Tags)
		assert.Equal(t, "inner", values.Nested.Value)
		assert.Equal(t, 100, values.Nested.ID)
	})

	t.Run("Default Values", func(t *testing.T) {
		// Create props with zero values and defaults
		props := NewPropsWithDefaults(
			TestProps{
				// Only set Name, other fields should use defaults
				Name: "Partial Data",
			},
			TestProps{
				Name:     "Default Name",
				Count:    10,
				IsActive: true,
				Tags:     []string{"default"},
				Nested: NestedProps{
					Value: "default-nested",
					ID:    999,
				},
			},
		)

		// Get values with defaults
		values := props.GetWithDefaults()

		// Verify the explicitly set values remain
		assert.Equal(t, "Partial Data", values.Name)

		// Verify the default values are used for zero values
		assert.Equal(t, 10, values.Count)
		assert.Equal(t, true, values.IsActive)
		assert.Equal(t, []string{"default"}, values.Tags)
		assert.Equal(t, "default-nested", values.Nested.Value)
		assert.Equal(t, 999, values.Nested.ID)
	})

	t.Run("Field Access", func(t *testing.T) {
		// Create props with initial values
		props := NewProps(TestProps{
			Name:     "Field Test",
			Count:    100,
			IsActive: true,
		})

		// Get specific fields
		name, err := props.GetField("Name")
		assert.NoError(t, err)
		assert.Equal(t, "Field Test", name)

		count, err := props.GetField("Count")
		assert.NoError(t, err)
		assert.Equal(t, 100, count)

		// Check field existence
		assert.True(t, props.HasField("Name"))
		assert.True(t, props.HasField("Count"))
		assert.True(t, props.HasField("IsActive"))
		assert.False(t, props.HasField("NonExistentField"))
	})

	t.Run("Validation", func(t *testing.T) {
		// Create validator function
		validator := func(p TestProps) error {
			if p.Count < 0 {
				return errors.New("count must be non-negative")
			}
			return nil
		}

		// Test valid props
		validProps, err := NewPropsWithValidation(TestProps{
			Name:  "Valid Props",
			Count: 5,
		}, validator)

		assert.NoError(t, err)
		assert.NotNil(t, validProps)

		// Test invalid props
		_, err = NewPropsWithValidation(TestProps{
			Name:  "Invalid Props",
			Count: -10,
		}, validator)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "count must be non-negative")
	})
}

func TestPropsImmutability(t *testing.T) {
	t.Run("Props Immutability", func(t *testing.T) {
		// Create initial props
		initialProps := TestProps{
			Name:     "Original",
			Count:    1,
			IsActive: true,
		}

		props := NewProps(initialProps)

		// Get a copy of the props
		values := props.Get()

		// Modify the copy
		values.Name = "Modified"
		values.Count = 999

		// Get another copy and verify it's still the original
		newValues := props.Get()
		assert.Equal(t, "Original", newValues.Name)
		assert.Equal(t, 1, newValues.Count)
		assert.Equal(t, true, newValues.IsActive)

		// Verify the original struct is untouched
		assert.Equal(t, "Original", initialProps.Name)
		assert.Equal(t, 1, initialProps.Count)
	})
}

func TestPropsWithPointers(t *testing.T) {
	type PointerProps struct {
		Data *string
		List *[]int
	}

	t.Run("Pointer Fields", func(t *testing.T) {
		text := "hello"
		list := []int{1, 2, 3}

		// Create props with pointer fields
		props := NewProps(PointerProps{
			Data: &text,
			List: &list,
		})

		// Access the props
		values := props.Get()

		// Verify values
		assert.Equal(t, &text, values.Data)
		assert.Equal(t, "hello", *values.Data)
		assert.Equal(t, &list, values.List)
		assert.Equal(t, []int{1, 2, 3}, *values.List)

		// Modify the original pointers
		text = "changed"
		list[0] = 999

		// Get new values - props should still hold original pointers,
		// but the data they point to has changed
		newValues := props.Get()
		assert.Equal(t, "changed", *newValues.Data)
		assert.Equal(t, []int{999, 2, 3}, *newValues.List)
	})
}
