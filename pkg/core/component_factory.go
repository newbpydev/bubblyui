package core

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ComponentOption is a functional option for configuring a component
type ComponentOption func(Component)

// StatefulComponentOption is a functional option for configuring a stateful component
type StatefulComponentOption func(StatefulComponent)

// ComponentConfigurator is a collection of ComponentOptions that can be applied together
type ComponentConfigurator []ComponentOption

// ApplyTo applies all component options in the configurator to the given component
func (c ComponentConfigurator) ApplyTo(component Component) {
	for _, option := range c {
		option(component)
	}
}

// ConfigurableComponent extends Component with a method to apply configuration options
type ConfigurableComponent interface {
	Component
	ApplyOption(option ComponentOption)
}

// ConfigurableBaseComponent adds option application to BaseComponent
type ConfigurableBaseComponent struct {
	*BaseComponent
	renderFunc    func(Component) string
	initFunc      func(Component) error
	disposeFunc   func(Component) error
	updateFunc    func(Component, tea.Msg) (tea.Cmd, error)
	key           string // Component key for reconciliation
	configOptions []ComponentOption
}

// NewConfigurableBaseComponent creates a new configurable component with the given ID
func NewConfigurableBaseComponent(id string) *ConfigurableBaseComponent {
	return &ConfigurableBaseComponent{
		BaseComponent: NewBaseComponent(id),
		renderFunc:    nil,
		initFunc:      nil,
		disposeFunc:   nil,
		updateFunc:    nil,
		key:           "",
		configOptions: make([]ComponentOption, 0),
	}
}

// ApplyOption applies a configuration option to the component
func (c *ConfigurableBaseComponent) ApplyOption(option ComponentOption) {
	option(c)
	c.configOptions = append(c.configOptions, option)
}

// Initialize first calls the custom init function if set, then initializes the base component
func (c *ConfigurableBaseComponent) Initialize() error {
	if c.initFunc != nil {
		if err := c.initFunc(c); err != nil {
			return err
		}
	}
	return c.BaseComponent.Initialize()
}

// Update first calls the custom update function if set, otherwise delegates to the base component
func (c *ConfigurableBaseComponent) Update(msg tea.Msg) (tea.Cmd, error) {
	if c.updateFunc != nil {
		return c.updateFunc(c, msg)
	}
	return c.BaseComponent.Update(msg)
}

// Render uses the custom render function if set, otherwise delegates to the base component
func (c *ConfigurableBaseComponent) Render() string {
	if c.renderFunc != nil {
		return c.renderFunc(c)
	}
	return c.BaseComponent.Render()
}

// Dispose first calls the custom dispose function if set, then disposes the base component
func (c *ConfigurableBaseComponent) Dispose() error {
	if c.disposeFunc != nil {
		if err := c.disposeFunc(c); err != nil {
			return err
		}
	}
	return c.BaseComponent.Dispose()
}

// Key returns the component's key used for reconciliation
func (c *ConfigurableBaseComponent) Key() string {
	return c.key
}

// ConfigurableStatefulComponent adds option application to BaseStatefulComponent
type ConfigurableStatefulComponent struct {
	*BaseStatefulComponent
	renderFunc    func(Component) string
	initFunc      func(Component) error
	disposeFunc   func(Component) error
	updateFunc    func(Component, tea.Msg) (tea.Cmd, error)
	key           string // Component key for reconciliation
	configOptions []ComponentOption
}

// NewConfigurableStatefulComponent creates a new configurable stateful component
func NewConfigurableStatefulComponent(id, name string) *ConfigurableStatefulComponent {
	return &ConfigurableStatefulComponent{
		BaseStatefulComponent: NewBaseStatefulComponent(id, name),
		renderFunc:            nil,
		initFunc:              nil,
		disposeFunc:           nil,
		updateFunc:            nil,
		key:                   "",
		configOptions:         make([]ComponentOption, 0),
	}
}

// ApplyOption applies a configuration option to the component
func (c *ConfigurableStatefulComponent) ApplyOption(option ComponentOption) {
	option(c)
	c.configOptions = append(c.configOptions, option)
}

// Initialize first calls the custom init function if set, then initializes the base component
func (c *ConfigurableStatefulComponent) Initialize() error {
	if c.initFunc != nil {
		if err := c.initFunc(c); err != nil {
			return err
		}
	}
	return c.BaseStatefulComponent.Initialize()
}

// Update first calls the custom update function if set, otherwise delegates to the base component
func (c *ConfigurableStatefulComponent) Update(msg tea.Msg) (tea.Cmd, error) {
	if c.updateFunc != nil {
		return c.updateFunc(c, msg)
	}
	return c.BaseStatefulComponent.Update(msg)
}

// Render uses the custom render function if set, otherwise delegates to the base component
func (c *ConfigurableStatefulComponent) Render() string {
	if c.renderFunc != nil {
		return c.renderFunc(c)
	}
	return c.BaseStatefulComponent.Render()
}

// Dispose first calls the custom dispose function if set, then disposes the base component
func (c *ConfigurableStatefulComponent) Dispose() error {
	if c.disposeFunc != nil {
		if err := c.disposeFunc(c); err != nil {
			return err
		}
	}
	return c.BaseStatefulComponent.Dispose()
}

// Key returns the component's key used for reconciliation
func (c *ConfigurableStatefulComponent) Key() string {
	return c.key
}

// Factory functions

// CreateComponent creates a new configurable component with the given ID and applies options
func CreateComponent(id string, options ...ComponentOption) Component {
	component := NewConfigurableBaseComponent(id)
	for _, option := range options {
		component.ApplyOption(option)
	}
	return component
}

// CreateStatefulComponent creates a new configurable stateful component with the given ID and name
func CreateStatefulComponent(id, name string, options ...ComponentOption) StatefulComponent {
	component := NewConfigurableStatefulComponent(id, name)
	for _, option := range options {
		component.ApplyOption(option)
	}
	return component
}

// Component configuration options

// WithRender creates an option that sets a custom render function for a component
func WithRender(renderFunc func(Component) string) ComponentOption {
	return func(c Component) {
		if cc, ok := c.(*ConfigurableBaseComponent); ok {
			cc.renderFunc = renderFunc
		} else if csc, ok := c.(*ConfigurableStatefulComponent); ok {
			csc.renderFunc = renderFunc
		}
	}
}

// WithInit creates an option that sets a custom initialization function for a component
func WithInit(initFunc func(Component) error) ComponentOption {
	return func(c Component) {
		if cc, ok := c.(*ConfigurableBaseComponent); ok {
			cc.initFunc = initFunc
		} else if csc, ok := c.(*ConfigurableStatefulComponent); ok {
			csc.initFunc = initFunc
		}
	}
}

// WithDispose creates an option that sets a custom dispose function for a component
func WithDispose(disposeFunc func(Component) error) ComponentOption {
	return func(c Component) {
		if cc, ok := c.(*ConfigurableBaseComponent); ok {
			cc.disposeFunc = disposeFunc
		} else if csc, ok := c.(*ConfigurableStatefulComponent); ok {
			csc.disposeFunc = disposeFunc
		}
	}
}

// WithUpdate creates an option that sets a custom update function for a component
func WithUpdate(updateFunc func(Component, tea.Msg) (tea.Cmd, error)) ComponentOption {
	return func(c Component) {
		if cc, ok := c.(*ConfigurableBaseComponent); ok {
			cc.updateFunc = updateFunc
		} else if csc, ok := c.(*ConfigurableStatefulComponent); ok {
			csc.updateFunc = updateFunc
		}
	}
}

// WithInitialChildren creates an option that adds initial children to a component
func WithInitialChildren(children ...Component) ComponentOption {
	return func(c Component) {
		for _, child := range children {
			c.AddChild(child)
		}
	}
}

// WithKey creates an option that sets a key for component reconciliation
func WithKey(key string) ComponentOption {
	return func(c Component) {
		if cc, ok := c.(*ConfigurableBaseComponent); ok {
			cc.key = key
		} else if csc, ok := c.(*ConfigurableStatefulComponent); ok {
			csc.key = key
		}
	}
}

// WithMounted creates an option that sets the mounted state of a stateful component
func WithMounted(mounted bool) ComponentOption {
	return func(c Component) {
		if sc, ok := c.(StatefulComponent); ok {
			sc.SetMounted(mounted)
		}
	}
}

// Tree traversal and component operations

// FindComponentByID searches for a component with the given ID in the component tree
func FindComponentByID(root Component, id string) Component {
	if root.ID() == id {
		return root
	}

	for _, child := range root.Children() {
		if found := FindComponentByID(child, id); found != nil {
			return found
		}
	}

	return nil
}

// TraverseComponentTree traverses the component tree in depth-first order and applies the visitor function
func TraverseComponentTree(root Component, visitor func(Component)) {
	visitor(root)

	for _, child := range root.Children() {
		TraverseComponentTree(child, visitor)
	}
}

// ComponentReconciliationResult contains the results of a component reconciliation operation
type ComponentReconciliationResult struct {
	Reused  []Component // Components that were reused from the old tree
	Added   []Component // Components that were added to the new tree
	Removed []Component // Components that were removed from the old tree
}

// ComponentReconciler manages component reconciliation with keys
type ComponentReconciler struct {
	components     []Component
	keyToComponent map[string]Component
}

// NewComponentReconciler creates a new component reconciler
func NewComponentReconciler() *ComponentReconciler {
	return &ComponentReconciler{
		components:     make([]Component, 0),
		keyToComponent: make(map[string]Component),
	}
}

// AddComponents adds components to the reconciler
func (r *ComponentReconciler) AddComponents(components ...Component) {
	for _, c := range components {
		r.components = append(r.components, c)

		// If the component has a key, add it to the key-to-component map
		if keyed, ok := c.(*ConfigurableBaseComponent); ok && keyed.key != "" {
			r.keyToComponent[keyed.key] = c
		} else if keyed, ok := c.(*ConfigurableStatefulComponent); ok && keyed.key != "" {
			r.keyToComponent[keyed.key] = c
		}
	}
}

// FindByKey returns the component with the given key, or nil if not found
func (r *ComponentReconciler) FindByKey(key string) Component {
	return r.keyToComponent[key]
}

// Reconcile reconciles the current components with a new set of components
func (r *ComponentReconciler) Reconcile(newComponents []Component) ComponentReconciliationResult {
	result := ComponentReconciliationResult{
		Reused:  make([]Component, 0),
		Added:   make([]Component, 0),
		Removed: make([]Component, 0),
	}

	// Track which old components are reused
	reused := make(map[string]bool)

	// First pass: identify reused and added components
	for _, newComp := range newComponents {
		var key string

		// Get the component's key
		if keyed, ok := newComp.(*ConfigurableBaseComponent); ok {
			key = keyed.key
		} else if keyed, ok := newComp.(*ConfigurableStatefulComponent); ok {
			key = keyed.key
		}

		if key != "" && r.keyToComponent[key] != nil {
			// Component with this key exists, mark as reused
			reused[key] = true
			result.Reused = append(result.Reused, newComp)
		} else {
			// New component
			result.Added = append(result.Added, newComp)
		}
	}

	// Second pass: identify removed components
	for _, oldComp := range r.components {
		var key string

		// Get the component's key
		if keyed, ok := oldComp.(*ConfigurableBaseComponent); ok {
			key = keyed.key
		} else if keyed, ok := oldComp.(*ConfigurableStatefulComponent); ok {
			key = keyed.key
		}

		if key != "" && !reused[key] {
			// Component with this key was not reused
			result.Removed = append(result.Removed, oldComp)
		}
	}

	// Update the reconciler's state
	r.components = newComponents
	r.keyToComponent = make(map[string]Component)
	for _, c := range newComponents {
		var key string

		// Get the component's key
		if keyed, ok := c.(*ConfigurableBaseComponent); ok {
			key = keyed.key
		} else if keyed, ok := c.(*ConfigurableStatefulComponent); ok {
			key = keyed.key
		}

		if key != "" {
			r.keyToComponent[key] = c
		}
	}

	return result
}
