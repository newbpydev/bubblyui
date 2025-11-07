package devtools

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCaptureComponent_BasicFields tests capturing basic component information
func TestCaptureComponent_BasicFields(t *testing.T) {
	tests := []struct {
		name          string
		componentName string
		componentID   string
		expectedName  string
		expectedID    string
	}{
		{
			name:          "simple component",
			componentName: "Counter",
			componentID:   "component-1",
			expectedName:  "Counter",
			expectedID:    "component-1",
		},
		{
			name:          "button component",
			componentName: "Button",
			componentID:   "component-42",
			expectedName:  "Button",
			expectedID:    "component-42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock component
			comp := &mockComponent{
				name: tt.componentName,
				id:   tt.componentID,
			}

			// Capture snapshot
			before := time.Now()
			snapshot := CaptureComponent(comp)
			after := time.Now()

			// Verify basic fields
			assert.Equal(t, tt.expectedID, snapshot.ID)
			assert.Equal(t, tt.expectedName, snapshot.Name)
			assert.NotEmpty(t, snapshot.Type, "Type should not be empty")
			assert.Contains(t, snapshot.Type, "mockComponent", "Type should contain mockComponent")
			assert.True(t, snapshot.Timestamp.After(before) || snapshot.Timestamp.Equal(before))
			assert.True(t, snapshot.Timestamp.Before(after) || snapshot.Timestamp.Equal(after))
		})
	}
}

// TestCaptureComponent_State tests capturing component state (refs)
func TestCaptureComponent_State(t *testing.T) {
	tests := []struct {
		name          string
		state         map[string]interface{}
		expectedRefs  int
		checkRefNames []string
	}{
		{
			name:          "no state",
			state:         map[string]interface{}{},
			expectedRefs:  0,
			checkRefNames: []string{},
		},
		{
			name: "single ref",
			state: map[string]interface{}{
				"count": &mockRef{
					id:    "ref-1",
					name:  "count",
					value: 42,
					typ:   "int",
				},
			},
			expectedRefs:  1,
			checkRefNames: []string{"count"},
		},
		{
			name: "multiple refs",
			state: map[string]interface{}{
				"count": &mockRef{
					id:    "ref-1",
					name:  "count",
					value: 42,
					typ:   "int",
				},
				"name": &mockRef{
					id:    "ref-2",
					name:  "name",
					value: "Alice",
					typ:   "string",
				},
			},
			expectedRefs:  2,
			checkRefNames: []string{"count", "name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &mockComponent{
				name:  "TestComponent",
				id:    "component-1",
				state: tt.state,
			}

			snapshot := CaptureComponent(comp)

			assert.Equal(t, tt.expectedRefs, len(snapshot.Refs))

			// Verify ref names are present
			refNames := make(map[string]bool)
			for _, ref := range snapshot.Refs {
				refNames[ref.Name] = true
			}

			for _, expectedName := range tt.checkRefNames {
				assert.True(t, refNames[expectedName], "Expected ref %s not found", expectedName)
			}
		})
	}
}

// TestCaptureComponent_Props tests capturing component props
func TestCaptureComponent_Props(t *testing.T) {
	tests := []struct {
		name          string
		props         interface{}
		expectedProps map[string]interface{}
	}{
		{
			name:          "nil props",
			props:         nil,
			expectedProps: map[string]interface{}{},
		},
		{
			name: "struct props",
			props: struct {
				Title string
				Count int
			}{
				Title: "Hello",
				Count: 42,
			},
			expectedProps: map[string]interface{}{
				"Title": "Hello",
				"Count": 42,
			},
		},
		{
			name: "map props",
			props: map[string]interface{}{
				"color": "blue",
				"size":  "large",
			},
			expectedProps: map[string]interface{}{
				"color": "blue",
				"size":  "large",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &mockComponent{
				name:  "TestComponent",
				id:    "component-1",
				props: tt.props,
			}

			snapshot := CaptureComponent(comp)

			assert.Equal(t, len(tt.expectedProps), len(snapshot.Props))
			for key, expectedValue := range tt.expectedProps {
				assert.Equal(t, expectedValue, snapshot.Props[key])
			}
		})
	}
}

// TestCaptureComponent_ParentChild tests capturing parent/child relationships
func TestCaptureComponent_ParentChild(t *testing.T) {
	t.Run("component with parent", func(t *testing.T) {
		parent := &mockComponent{
			name: "Parent",
			id:   "component-1",
		}

		child := &mockComponent{
			name:   "Child",
			id:     "component-2",
			parent: parent,
		}

		snapshot := CaptureComponent(child)

		require.NotNil(t, snapshot.Parent)
		assert.Equal(t, "Parent", snapshot.Parent.Name)
		assert.Equal(t, "component-1", snapshot.Parent.ID)
	})

	t.Run("component with children", func(t *testing.T) {
		child1 := &mockComponent{
			name: "Child1",
			id:   "component-2",
		}

		child2 := &mockComponent{
			name: "Child2",
			id:   "component-3",
		}

		parent := &mockComponent{
			name:     "Parent",
			id:       "component-1",
			children: []*mockComponent{child1, child2},
		}

		snapshot := CaptureComponent(parent)

		assert.Equal(t, 2, len(snapshot.Children))
		assert.Equal(t, "Child1", snapshot.Children[0].Name)
		assert.Equal(t, "Child2", snapshot.Children[1].Name)
	})

	t.Run("root component (no parent)", func(t *testing.T) {
		root := &mockComponent{
			name: "Root",
			id:   "component-1",
		}

		snapshot := CaptureComponent(root)

		assert.Nil(t, snapshot.Parent)
	})
}

// TestCaptureComponent_RefSnapshot tests RefSnapshot creation
func TestCaptureComponent_RefSnapshot(t *testing.T) {
	tests := []struct {
		name          string
		ref           *mockRef
		expectedID    string
		expectedName  string
		expectedType  string
		expectedValue interface{}
	}{
		{
			name: "int ref",
			ref: &mockRef{
				id:    "ref-1",
				name:  "count",
				value: 42,
				typ:   "int",
			},
			expectedID:    "ref-1",
			expectedName:  "count",
			expectedType:  "int",
			expectedValue: 42,
		},
		{
			name: "string ref",
			ref: &mockRef{
				id:    "ref-2",
				name:  "message",
				value: "Hello",
				typ:   "string",
			},
			expectedID:    "ref-2",
			expectedName:  "message",
			expectedType:  "string",
			expectedValue: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &mockComponent{
				name: "TestComponent",
				id:   "component-1",
				state: map[string]interface{}{
					tt.ref.name: tt.ref,
				},
			}

			snapshot := CaptureComponent(comp)

			require.Equal(t, 1, len(snapshot.Refs))
			refSnapshot := snapshot.Refs[0]

			assert.Equal(t, tt.expectedID, refSnapshot.ID)
			assert.Equal(t, tt.expectedName, refSnapshot.Name)
			assert.Equal(t, tt.expectedType, refSnapshot.Type)
			assert.Equal(t, tt.expectedValue, refSnapshot.Value)
		})
	}
}

// Mock types for testing

type mockComponent struct {
	name     string
	id       string
	state    map[string]interface{}
	props    interface{}
	parent   *mockComponent
	children []*mockComponent
}

func (m *mockComponent) GetName() string {
	return m.name
}

func (m *mockComponent) GetID() string {
	return m.id
}

func (m *mockComponent) GetState() map[string]interface{} {
	if m.state == nil {
		return make(map[string]interface{})
	}
	return m.state
}

func (m *mockComponent) GetProps() interface{} {
	return m.props
}

func (m *mockComponent) GetParent() ComponentInterface {
	if m.parent == nil {
		return nil
	}
	return m.parent
}

func (m *mockComponent) GetChildren() []ComponentInterface {
	if m.children == nil {
		return []ComponentInterface{}
	}
	result := make([]ComponentInterface, len(m.children))
	for i, child := range m.children {
		result[i] = child
	}
	return result
}

type mockRef struct {
	id    string
	name  string
	value interface{}
	typ   string
}

func (m *mockRef) GetID() string {
	return m.id
}

func (m *mockRef) GetName() string {
	return m.name
}

func (m *mockRef) GetValue() interface{} {
	return m.value
}

func (m *mockRef) GetType() string {
	return m.typ
}

func (m *mockRef) GetWatcherCount() int {
	return 0
}
