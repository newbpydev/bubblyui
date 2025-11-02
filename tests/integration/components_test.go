package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// testRoot wraps a component with theme provision for testing
func testRoot(child bubbly.Component) bubbly.Component {
	root, _ := bubbly.NewComponent("TestRoot").
		Children(child).
		Setup(func(ctx *bubbly.Context) {
			// Provide default theme for all child components
			ctx.Provide("theme", components.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Render child directly
			children := ctx.Children()
			if len(children) > 0 {
				return ctx.RenderChild(children[0])
			}
			return ""
		}).
		Build()
	return root
}

// TestFormWithInputs verifies that Form component properly integrates with Input components.
// This tests the molecule (Input) to organism (Form) integration.
func TestFormWithInputs(t *testing.T) {
	t.Run("form collects input values", func(t *testing.T) {
		// Setup: Create refs and components
		nameRef := bubbly.NewRef("")
		emailRef := bubbly.NewRef("")

		nameInput := components.Input(components.InputProps{
			Value:       nameRef,
			Placeholder: "Enter name",
			Type:        components.InputText,
		})

		emailInput := components.Input(components.InputProps{
			Value:       emailRef,
			Placeholder: "Enter email",
			Type:        components.InputEmail,
		})

		type UserData struct {
			Name  string
			Email string
		}

		submitted := false
		var submittedData UserData

		form := components.Form(components.FormProps[UserData]{
			Initial: UserData{},
			Fields: []components.FormField{
				{
					Name:      "Name",
					Label:     "Full Name",
					Component: nameInput,
				},
				{
					Name:      "Email",
					Label:     "Email Address",
					Component: emailInput,
				},
			},
			Validate: func(data UserData) map[string]string {
				errors := make(map[string]string)
				if data.Name == "" {
					errors["Name"] = "Name is required"
				}
				if data.Email == "" {
					errors["Email"] = "Email is required"
				}
				return errors
			},
			OnSubmit: func(data UserData) {
				submitted = true
				submittedData = data
			},
		})

		// Initialize components
		form.Init()

		// Action: Update input values
		nameRef.Set("Alice")
		emailRef.Set("alice@example.com")

		// Wait for reactive updates
		time.Sleep(10 * time.Millisecond)

		// Verify inputs are rendered in form
		view := form.View()
		assert.Contains(t, view, "Full Name")
		assert.Contains(t, view, "Email Address")
		assert.Contains(t, view, "Alice")
		assert.Contains(t, view, "alice@example.com")

		// Note: Form submission would need to be triggered through proper events
		// For now, verify the form structure is correct
		assert.False(t, submitted, "Form not submitted yet")
		assert.Equal(t, UserData{}, submittedData)
	})

	t.Run("form validation with inputs", func(t *testing.T) {
		// Setup
		nameRef := bubbly.NewRef("")
		emailRef := bubbly.NewRef("invalid")

		nameInput := components.Input(components.InputProps{
			Value: nameRef,
		})

		emailInput := components.Input(components.InputProps{
			Value: emailRef,
		})

		type UserData struct {
			Name  string
			Email string
		}

		form := components.Form(components.FormProps[UserData]{
			Initial: UserData{},
			Fields: []components.FormField{
				{Name: "Name", Component: nameInput},
				{Name: "Email", Component: emailInput},
			},
			Validate: func(data UserData) map[string]string {
				errors := make(map[string]string)
				if data.Name == "" {
					errors["Name"] = "Name is required"
				}
				return errors
			},
		})

		form.Init()

		// Action: Try to submit with empty name
		form.Emit("submit", nil)

		time.Sleep(10 * time.Millisecond)

		// Verify validation error appears
		view := form.View()
		assert.Contains(t, view, "Name is required")
	})

	t.Run("form with multiple inputs renders correctly", func(t *testing.T) {
		// Setup: Create form with 3 inputs
		name := bubbly.NewRef("")
		email := bubbly.NewRef("")
		age := bubbly.NewRef("")

		type FormData struct {
			Name  string
			Email string
			Age   string
		}

		form := components.Form(components.FormProps[FormData]{
			Initial: FormData{},
			Fields: []components.FormField{
				{Name: "Name", Label: "Name", Component: components.Input(components.InputProps{Value: name})},
				{Name: "Email", Label: "Email", Component: components.Input(components.InputProps{Value: email})},
				{Name: "Age", Label: "Age", Component: components.Input(components.InputProps{Value: age})},
			},
		})

		form.Init()

		// Verify all fields render
		view := form.View()
		assert.Contains(t, view, "Name")
		assert.Contains(t, view, "Email")
		assert.Contains(t, view, "Age")
	})
}

// TestTableInLayout verifies that Table component works correctly inside layout components.
// This tests the organism (Table) to template (PageLayout) integration.
func TestTableInLayout(t *testing.T) {
	t.Run("table renders in page layout", func(t *testing.T) {
		// Setup: Create table data
		type User struct {
			Name  string
			Email string
			Age   int
		}

		usersData := bubbly.NewRef([]User{
			{Name: "Alice", Email: "alice@example.com", Age: 30},
			{Name: "Bob", Email: "bob@example.com", Age: 25},
			{Name: "Charlie", Email: "charlie@example.com", Age: 35},
		})

		// Create table
		table := components.Table(components.TableProps[User]{
			Data: usersData,
			Columns: []components.TableColumn[User]{
				{Header: "Name", Field: "Name", Width: 20},
				{Header: "Email", Field: "Email", Width: 30},
				{Header: "Age", Field: "Age", Width: 10},
			},
		})

		// Create page layout with table
		title := components.Text(components.TextProps{
			Content: "User Directory",
			Bold:    true,
		})

		// Initialize children first
		title.Init()
		table.Init()

		layout := components.PageLayout(components.PageLayoutProps{
			Title:   title,
			Content: table,
		})

		// Wrap with theme provider
		root := testRoot(layout)

		// Initialize
		root.Init()

		// Verify layout renders with table
		view := root.View()
		assert.Contains(t, view, "User Directory")
		assert.Contains(t, view, "Name")
		assert.Contains(t, view, "Email")
		assert.Contains(t, view, "Age")
		assert.Contains(t, view, "Alice")
		assert.Contains(t, view, "Bob")
		assert.Contains(t, view, "Charlie")
	})

	t.Run("table updates work inside layout", func(t *testing.T) {
		// Setup
		type Item struct {
			Name string
		}

		itemsData := bubbly.NewRef([]Item{
			{Name: "Item 1"},
		})

		table := components.Table(components.TableProps[Item]{
			Data: itemsData,
			Columns: []components.TableColumn[Item]{
				{Header: "Name", Field: "Name", Width: 20},
			},
		})

		// Initialize table first
		table.Init()

		layout := components.PageLayout(components.PageLayoutProps{
			Content: table,
		})

		root := testRoot(layout)
		root.Init()

		// Action: Update table data
		itemsData.Set([]Item{
			{Name: "Item 1"},
			{Name: "Item 2"},
			{Name: "Item 3"},
		})

		time.Sleep(10 * time.Millisecond)

		// Verify updated data appears
		view := root.View()
		assert.Contains(t, view, "Item 1")
		assert.Contains(t, view, "Item 2")
		assert.Contains(t, view, "Item 3")
	})

	t.Run("table with actions in layout", func(t *testing.T) {
		// Setup
		type User struct {
			Name string
		}

		usersData := bubbly.NewRef([]User{{Name: "Alice"}})

		table := components.Table(components.TableProps[User]{
			Data: usersData,
			Columns: []components.TableColumn[User]{
				{Header: "Name", Field: "Name", Width: 20},
			},
		})

		refreshButton := components.Button(components.ButtonProps{
			Label:   "Refresh",
			Variant: components.ButtonPrimary,
		})

		// Initialize children first
		table.Init()
		refreshButton.Init()

		layout := components.PageLayout(components.PageLayoutProps{
			Content: table,
			Actions: refreshButton,
		})

		root := testRoot(layout)
		root.Init()

		// Verify layout includes table and actions
		view := root.View()
		assert.Contains(t, view, "Alice")
		assert.Contains(t, view, "Refresh")
	})
}

// TestModalWithForm verifies that Modal component can contain and display Form components.
// This tests the organism (Form) inside another organism (Modal) composition.
func TestModalWithForm(t *testing.T) {
	t.Run("modal displays form", func(t *testing.T) {
		// Setup
		visible := bubbly.NewRef(true)
		nameRef := bubbly.NewRef("")

		type UserData struct {
			Name string
		}

		inputComp := components.Input(components.InputProps{
			Value: nameRef,
		})
		inputComp.Init()

		form := components.Form(components.FormProps[UserData]{
			Initial: UserData{},
			Fields: []components.FormField{
				{
					Name:      "Name",
					Label:     "Name",
					Component: inputComp,
				},
			},
		})
		form.Init()

		modal := components.Modal(components.ModalProps{
			Title:   "Edit User",
			Visible: visible,
			Buttons: []bubbly.Component{form},
		})

		// Initialize
		modal.Init()

		// Verify modal shows form when visible
		view := modal.View()
		assert.Contains(t, view, "Edit User")
		assert.Contains(t, view, "Name")
	})

	t.Run("modal hide/show with form", func(t *testing.T) {
		// Setup
		visible := bubbly.NewRef(false)
		nameRef := bubbly.NewRef("")

		type UserData struct {
			Name string
		}

		inputComp := components.Input(components.InputProps{Value: nameRef})
		inputComp.Init()

		form := components.Form(components.FormProps[UserData]{
			Initial: UserData{},
			Fields: []components.FormField{
				{
					Name:      "Name",
					Label:     "Name", // Add label so it appears in output
					Component: inputComp,
				},
			},
		})
		form.Init()

		modal := components.Modal(components.ModalProps{
			Title:   "Form Modal",
			Visible: visible,
			Buttons: []bubbly.Component{form},
		})

		modal.Init()

		// Initially hidden
		view := modal.View()
		assert.NotContains(t, view, "Form Modal")

		// Show modal
		visible.Set(true)
		time.Sleep(10 * time.Millisecond)

		view = modal.View()
		assert.Contains(t, view, "Form Modal")
		// Form should be visible with label
		assert.Contains(t, view, "Form")

		// Hide modal
		visible.Set(false)
		time.Sleep(10 * time.Millisecond)

		view = modal.View()
		assert.NotContains(t, view, "Form Modal")
	})

	t.Run("form submission in modal", func(t *testing.T) {
		// Setup
		visible := bubbly.NewRef(true)
		nameRef := bubbly.NewRef("")
		submitted := false

		type UserData struct {
			Name string
		}

		inputComp := components.Input(components.InputProps{Value: nameRef})
		inputComp.Init()

		form := components.Form(components.FormProps[UserData]{
			Initial: UserData{},
			Fields: []components.FormField{
				{
					Name:      "Name",
					Component: inputComp,
				},
			},
			OnSubmit: func(data UserData) {
				submitted = true
				visible.Set(false)
			},
		})
		form.Init()

		modal := components.Modal(components.ModalProps{
			Title:   "Submit Form",
			Visible: visible,
			Buttons: []bubbly.Component{form},
		})

		modal.Init()

		// Set value and submit
		nameRef.Set("Alice")
		form.Emit("submit", nil)

		time.Sleep(10 * time.Millisecond)

		// Verify submission worked and modal closed
		assert.True(t, submitted)
		assert.False(t, visible.GetTyped())
	})
}

// TestFullAppComposition verifies complex multi-component hierarchies work correctly.
// This tests the full atomic design hierarchy: Atoms → Molecules → Organisms → Templates.
func TestFullAppComposition(t *testing.T) {
	t.Run("complete app with all component levels", func(t *testing.T) {
		// Setup: Build a complete app structure
		type User struct {
			Name  string
			Email string
		}

		// Data layer
		usersData := bubbly.NewRef([]User{
			{Name: "Alice", Email: "alice@example.com"},
			{Name: "Bob", Email: "bob@example.com"},
		})

		modalVisible := bubbly.NewRef(false)
		selectedUser := bubbly.NewRef("")

		// Atoms
		titleText := components.Text(components.TextProps{
			Content: "User Management",
			Bold:    true,
		})

		addButton := components.Button(components.ButtonProps{
			Label:   "Add User",
			Variant: components.ButtonPrimary,
			OnClick: func() {
				modalVisible.Set(true)
			},
		})

		// Molecules (Input is created inline in Form)
		nameRef := bubbly.NewRef("")
		emailRef := bubbly.NewRef("")

		// Organisms
		usersTable := components.Table(components.TableProps[User]{
			Data: usersData,
			Columns: []components.TableColumn[User]{
				{Header: "Name", Field: "Name", Width: 20},
				{Header: "Email", Field: "Email", Width: 30},
			},
			OnRowClick: func(user User, index int) {
				selectedUser.Set(user.Name)
			},
		})

		type FormData struct {
			Name  string
			Email string
		}

		userForm := components.Form(components.FormProps[FormData]{
			Initial: FormData{},
			Fields: []components.FormField{
				{
					Name:  "Name",
					Label: "Name",
					Component: components.Input(components.InputProps{
						Value: nameRef,
					}),
				},
				{
					Name:  "Email",
					Label: "Email",
					Component: components.Input(components.InputProps{
						Value: emailRef,
					}),
				},
			},
			OnSubmit: func(data FormData) {
				// Add user to table
				current := usersData.GetTyped()
				usersData.Set(append(current, User{Name: data.Name, Email: data.Email}))
				modalVisible.Set(false)
				nameRef.Set("")
				emailRef.Set("")
			},
		})

		addModal := components.Modal(components.ModalProps{
			Title:   "Add New User",
			Visible: modalVisible,
			Buttons: []bubbly.Component{userForm},
		})

		// Initialize all child components first
		titleText.Init()
		addButton.Init()
		usersTable.Init()
		userForm.Init()
		addModal.Init()

		// Template
		app := components.PageLayout(components.PageLayoutProps{
			Title:   titleText,
			Content: usersTable,
			Actions: addButton,
		})

		// Wrap with theme provider
		root := testRoot(app)

		// Initialize entire hierarchy
		root.Init()

		// Verify: All components render in hierarchy
		view := root.View()
		require.NotEmpty(t, view)

		// Verify atoms
		assert.Contains(t, view, "User Management")
		assert.Contains(t, view, "Add User")

		// Verify organisms (table data)
		assert.Contains(t, view, "Alice")
		assert.Contains(t, view, "Bob")
		assert.Contains(t, view, "alice@example.com")

		// Modal should not be visible initially
		modalView := addModal.View()
		assert.NotContains(t, modalView, "Add New User")
	})

	t.Run("full user interaction flow", func(t *testing.T) {
		// Setup: Realistic user flow
		type User struct {
			Name string
		}

		usersData := bubbly.NewRef([]User{
			{Name: "Alice"},
		})

		modalVisible := bubbly.NewRef(false)
		nameRef := bubbly.NewRef("")

		// Build app
		table := components.Table(components.TableProps[User]{
			Data: usersData,
			Columns: []components.TableColumn[User]{
				{Header: "Name", Field: "Name", Width: 20},
			},
		})

		type FormData struct {
			Name string
		}

		// Initialize input first
		inputComp := components.Input(components.InputProps{Value: nameRef})
		inputComp.Init()

		// Create form with initialized input
		form := components.Form(components.FormProps[FormData]{
			Initial: FormData{},
			Fields: []components.FormField{
				{
					Name:      "Name",
					Label:     "Name",
					Component: inputComp,
				},
			},
			Validate: func(data FormData) map[string]string {
				// No validation errors - allow submission
				return make(map[string]string)
			},
			OnSubmit: func(data FormData) {
				// Get current data from ref
				current := usersData.GetTyped()
				// Append new user with value from nameRef
				newUser := User{Name: nameRef.GetTyped()}
				usersData.Set(append(current, newUser))
				modalVisible.Set(false)
			},
		})

		// Initialize children
		table.Init()
		form.Init()

		// Create modal with initialized form
		modal := components.Modal(components.ModalProps{
			Title:   "Add User",
			Visible: modalVisible,
			Buttons: []bubbly.Component{form},
		})
		modal.Init()

		app := components.PageLayout(components.PageLayoutProps{
			Content: table,
		})

		root := testRoot(app)
		root.Init()

		// Initialize modal separately with theme
		modalRoot := testRoot(modal)
		modalRoot.Init()

		// User Flow 1: View table
		view := root.View()
		assert.Contains(t, view, "Alice")

		// User Flow 2: Open modal
		modalVisible.Set(true)
		time.Sleep(10 * time.Millisecond)

		modalView := modalRoot.View()
		assert.Contains(t, modalView, "Add User")

		// User Flow 3: Fill form
		nameRef.Set("Bob")
		time.Sleep(10 * time.Millisecond)

		// User Flow 4: Submit form
		form.Emit("submit", nil)
		time.Sleep(50 * time.Millisecond) // Increase wait time for reactive updates

		// Verify: New user added to table
		view = root.View()
		assert.Contains(t, view, "Alice")
		assert.Contains(t, view, "Bob", "Bob should be added to the table after form submission")

		// Verify: Modal closed
		assert.False(t, modalVisible.GetTyped())
	})

	t.Run("event propagation through hierarchy", func(t *testing.T) {
		// Setup: Test event bubbling from deep component
		clicked := false

		button := components.Button(components.ButtonProps{
			Label: "Click Me",
			OnClick: func() {
				clicked = true
			},
		})

		card := components.Card(components.CardProps{
			Title:    "Card",
			Children: []bubbly.Component{button},
		})

		// Initialize children first
		button.Init()
		card.Init()

		layout := components.PageLayout(components.PageLayoutProps{
			Content: card,
		})

		root := testRoot(layout)
		root.Init()

		// Verify button is in hierarchy
		view := root.View()
		assert.Contains(t, view, "Click Me")

		// Trigger button click through emit
		button.Emit("click", nil)
		time.Sleep(10 * time.Millisecond)

		// Verify event handled
		assert.True(t, clicked)
	})
}
