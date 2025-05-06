package core

import (
	"fmt"
	"reflect"
	"strings"
)

// ComponentDiff represents the changes between two component trees
type ComponentDiff struct {
	// Components that need updates (property changes)
	Updates []ComponentUpdate

	// Components that need to be added
	Additions []ComponentAddition

	// Components that need to be removed
	Removals []ComponentRemoval

	// Components that need to be reordered
	Reorders []ComponentReorder
}

// ComponentUpdate represents a component that needs to be updated
type ComponentUpdate struct {
	Component    *ComponentManager
	ChangedProps []string
}

// ComponentAddition represents a component that needs to be added
type ComponentAddition struct {
	Parent    *ComponentManager
	Component *ComponentManager
	Index     int // Position to insert at
}

// ComponentRemoval represents a component that needs to be removed
type ComponentRemoval struct {
	Parent    *ComponentManager
	Component *ComponentManager
}

// ComponentReorder represents a change in the order of child components
type ComponentReorder struct {
	Parent   *ComponentManager
	NewOrder []*ComponentManager
}

// ComponentSnapshot represents a snapshot of a component tree
type ComponentSnapshot struct {
	// Map component paths to their snapshot data
	Components map[string]*ComponentData

	// Store the root component reference
	Root *ComponentManager

	// Track component hierarchy
	Hierarchy map[string][]string // parent path -> child paths
}

// ComponentData stores the snapshot data for a component
type ComponentData struct {
	Name     string
	Path     string
	Props    map[string]interface{}
	Children []string // Paths to child components
}

// ComponentDiffer performs the diffing between component trees
type ComponentDiffer struct {
	// Key property used for reconciliation
	reconciliationKey string

	// Whether to use key-based reconciliation
	useKeyReconciliation bool
}

// NewComponentDiffer creates a new component differ
func NewComponentDiffer() *ComponentDiffer {
	return &ComponentDiffer{
		reconciliationKey:    "",
		useKeyReconciliation: false,
	}
}

// EnableKeyReconciliation enables key-based reconciliation using the specified property
func (d *ComponentDiffer) EnableKeyReconciliation(keyProp string) {
	d.reconciliationKey = keyProp
	d.useKeyReconciliation = true
}

// DisableKeyReconciliation disables key-based reconciliation
func (d *ComponentDiffer) DisableKeyReconciliation() {
	d.reconciliationKey = ""
	d.useKeyReconciliation = false
}

// NewComponentSnapshot creates a snapshot of the component tree
func NewComponentSnapshot(root *ComponentManager) *ComponentSnapshot {
	snapshot := &ComponentSnapshot{
		Components: make(map[string]*ComponentData),
		Root:       root,
		Hierarchy:  make(map[string][]string),
	}

	// Recursively build the snapshot
	snapshot.captureComponent(root, "")

	return snapshot
}

// captureComponent captures a component and its children for the snapshot
func (s *ComponentSnapshot) captureComponent(component *ComponentManager, parentPath string) string {
	// Build this component's path
	path := component.GetName()
	if parentPath != "" {
		path = parentPath + "/" + path
	}

	// Create component data
	data := &ComponentData{
		Name:     component.GetName(),
		Path:     path,
		Props:    make(map[string]interface{}),
		Children: make([]string, 0),
	}

	// Capture all props
	props := component.props // Directly access props field for efficiency
	for key, value := range props {
		data.Props[key] = value
	}

	// Add to components map
	s.Components[path] = data

	// Setup parent-child relationship in hierarchy
	if parentPath != "" {
		if _, exists := s.Hierarchy[parentPath]; !exists {
			s.Hierarchy[parentPath] = make([]string, 0)
		}
		s.Hierarchy[parentPath] = append(s.Hierarchy[parentPath], path)
	}

	// Recursively capture children
	for _, child := range component.GetChildren() {
		childPath := s.captureComponent(child, path)
		data.Children = append(data.Children, childPath)
	}

	return path
}

// Diff calculates the changes between the snapshot and the new component tree
func (d *ComponentDiffer) Diff(snapshot *ComponentSnapshot, newRoot *ComponentManager) *ComponentDiff {
	diff := &ComponentDiff{
		Updates:   make([]ComponentUpdate, 0),
		Additions: make([]ComponentAddition, 0),
		Removals:  make([]ComponentRemoval, 0),
		Reorders:  make([]ComponentReorder, 0),
	}

	// Track component paths we've processed to identify removals later
	processedPaths := make(map[string]bool)

	// For key-based reconciliation, build a map of keys to component data first
	keyToComponentData := make(map[interface{}]*ComponentData)
	if d.useKeyReconciliation && d.reconciliationKey != "" {
		for _, data := range snapshot.Components {
			if key, exists := data.Props[d.reconciliationKey]; exists {
				keyToComponentData[key] = data
			}
		}
	}

	// Process the new tree and compare with snapshot
	d.diffComponent(snapshot, newRoot, "", diff, processedPaths, keyToComponentData)

	// Find components that were removed
	d.findRemovals(snapshot, processedPaths, diff)

	return diff
}

// diffComponent compares a component and its children with the snapshot
func (d *ComponentDiffer) diffComponent(
	snapshot *ComponentSnapshot,
	component *ComponentManager,
	parentPath string,
	diff *ComponentDiff,
	processedPaths map[string]bool,
	keyToComponentData map[interface{}]*ComponentData,
) {
	// Special handling for root component - check for reordering of children
	if parentPath == "" && d.useKeyReconciliation && len(component.GetChildren()) > 0 {
		// Handle the root component's children reordering specially
		d.detectRootReordering(snapshot, component, diff)
	}
	// Build this component's path
	path := component.GetName()
	if parentPath != "" {
		path = parentPath + "/" + path
	}

	// Mark this path as processed
	processedPaths[path] = true

	// Track if we need to process this component or if it was handled by key reconciliation
	reconciled := false

	// If we're using key-based reconciliation and this component has a key
	var oldData *ComponentData
	if d.useKeyReconciliation && d.reconciliationKey != "" {
		if key, ok := component.GetProp(d.reconciliationKey); ok {
			// Create a special path format that indicates this key has been processed
			keyPath := "__KEY__:" + fmt.Sprintf("%v", key)
			processedPaths[keyPath] = true

			// Check if we have old data for this key
			if oldComponentData, exists := keyToComponentData[key]; exists {
				// Mark the old path as processed
				processedPaths[oldComponentData.Path] = true

				// Check if the name has changed
				if oldComponentData.Name != component.GetName() {
					// This is a rename - only add to updates if name changed
					diff.Updates = append(diff.Updates, ComponentUpdate{
						Component:    component,
						ChangedProps: []string{"name"},
					})
				}

				// Keep track of old data for prop comparison
				oldData = oldComponentData

				// This component was handled by key reconciliation
				reconciled = true
			}
		}
	}

	// If this component wasn't handled by key reconciliation
	if !reconciled {
		// Check if this component exists in the snapshot
		snapshotData, exists := snapshot.Components[path]

		if !exists {
			// This is a new component
			var parent *ComponentManager
			if parentPath != "" {
				// Find the parent in the snapshot
				for _, c := range snapshot.Components {
					if c.Path == parentPath {
						parent = findComponentByPath(snapshot.Root, parentPath)
						break
					}
				}
			}

			diff.Additions = append(diff.Additions, ComponentAddition{
				Parent:    parent,
				Component: component,
				Index:     len(component.GetParent().GetChildren()) - 1, // Approximate position
			})
		} else {
			// Use the snapshot data for prop comparison
			oldData = snapshotData
		}
	}

	// Compare props to find changes if we have old data
	if oldData != nil {
		changedProps := d.diffProps(oldData.Props, component)

		if len(changedProps) > 0 {
			// Skip adding to updates if this is a name change and we already added it
			if reconciled && len(changedProps) == 1 && changedProps[0] == "name" {
				// Skip, already added as an update for name change
			} else {
				// Add to updates for property changes
				diff.Updates = append(diff.Updates, ComponentUpdate{
					Component:    component,
					ChangedProps: changedProps,
				})
			}
		}
	}

	// Check for reordering of children when key reconciliation is enabled
	if d.useKeyReconciliation && d.reconciliationKey != "" && len(component.GetChildren()) > 0 {
		d.checkForReordering(snapshot, component, path, diff)
	}

	// Process children
	for _, child := range component.GetChildren() {
		d.diffComponent(snapshot, child, path, diff, processedPaths, keyToComponentData)
	}
}

// diffProps compares properties between snapshot and current component
func (d *ComponentDiffer) diffProps(
	snapshotProps map[string]interface{},
	component *ComponentManager,
) []string {
	changedProps := make([]string, 0)

	// Check for changed or new props
	for key, value := range component.props {
		snapshotValue, exists := snapshotProps[key]
		if !exists || !reflect.DeepEqual(value, snapshotValue) {
			changedProps = append(changedProps, key)
		}
	}

	// Check for removed props
	for key := range snapshotProps {
		_, exists := component.props[key]
		if !exists {
			changedProps = append(changedProps, key)
		}
	}

	return changedProps
}

// findRemovals identifies components that were in the snapshot but not in the new tree
func (d *ComponentDiffer) findRemovals(
	snapshot *ComponentSnapshot,
	processedPaths map[string]bool,
	diff *ComponentDiff,
) {
	// Gather all keys in the old tree if key reconciliation is enabled
	keysInOldTree := make(map[interface{}]string) // key -> path
	if d.useKeyReconciliation && d.reconciliationKey != "" {
		for path, component := range snapshot.Components {
			if key, exists := component.Props[d.reconciliationKey]; exists {
				keysInOldTree[key] = path
			}
		}
	}

	for path := range snapshot.Components {
		if !processedPaths[path] {
			// If using key reconciliation, check if this component's key exists in the new tree
			if d.useKeyReconciliation && d.reconciliationKey != "" {
				componentData := snapshot.Components[path]
				if key, exists := componentData.Props[d.reconciliationKey]; exists {
					// Check for our special key marker
					keyPath := "__KEY__:" + fmt.Sprintf("%v", key)
					if processedPaths[keyPath] {
						// This key has been processed, which means the component was included
						// in the new tree but possibly with a different name or path
						continue
					}
				}
			}

			// This component was in the snapshot but not in the new tree
			// It's been removed
			parentPath := ""
			parts := splitPath(path)
			if len(parts) > 1 {
				parentPath = joinPath(parts[:len(parts)-1])
			}

			// Find the parent component reference
			var parent *ComponentManager
			if parentPath != "" {
				parent = findComponentByPath(snapshot.Root, parentPath)
			}

			// Find the component reference in the snapshot
			component := findComponentByPath(snapshot.Root, path)

			diff.Removals = append(diff.Removals, ComponentRemoval{
				Parent:    parent,
				Component: component,
			})
		}
	}
}

// checkForReordering checks if children have been reordered based on keys
func (d *ComponentDiffer) checkForReordering(
	snapshot *ComponentSnapshot,
	component *ComponentManager,
	parentPath string,
	diff *ComponentDiff,
) {
	if !d.useKeyReconciliation || d.reconciliationKey == "" || len(component.GetChildren()) == 0 {
		return
	}

	// Create a full path for this component
	path := component.GetName()
	if parentPath != "" {
		path = parentPath + "/" + path
	}

	// For root components, we need to explicitly check as paths might differ
	// due to name changes
	var originalComp *ComponentManager
	if path == "Root" || parentPath == "" {
		// This is likely the root component, find it directly
		originalComp = snapshot.Root
	} else {
		// For non-root components, find by path
		originalComp = findComponentByPath(snapshot.Root, path)
	}

	if originalComp == nil {
		return // Original component not found
	}

	// Build map of components that already have updates
	updatedComps := make(map[*ComponentManager]bool)
	for _, update := range diff.Updates {
		updatedComps[update.Component] = true
	}

	// Map keys to indices for both old and new trees
	oldKeyToIndex := make(map[interface{}]int)
	newKeyToIndex := make(map[interface{}]int)
	oldKeyToComponent := make(map[interface{}]*ComponentManager)
	newKeyToComponent := make(map[interface{}]*ComponentManager)

	// Map old keys to indices and components
	oldChildren := originalComp.GetChildren()
	for i, child := range oldChildren {
		if key, exists := child.GetProp(d.reconciliationKey); exists {
			oldKeyToIndex[key] = i
			oldKeyToComponent[key] = child
		}
	}

	// Map new keys to indices and components
	newChildren := component.GetChildren()
	for i, child := range newChildren {
		if key, ok := child.GetProp(d.reconciliationKey); ok {
			newKeyToIndex[key] = i
			newKeyToComponent[key] = child
		}
	}

	// Check if the order has changed
	// We need at least 2 common elements to have a meaningful reordering
	commonElements := 0
	for key := range oldKeyToIndex {
		if _, ok := newKeyToIndex[key]; ok {
			commonElements++
		}
	}

	if commonElements >= 2 {
		// Now check if positions differ for any key
		orderChanged := false
		for key, oldIndex := range oldKeyToIndex {
			if newIndex, ok := newKeyToIndex[key]; ok {
				if oldIndex != newIndex {
					orderChanged = true
					break
				}
			}
		}

		if orderChanged {
			diff.Reorders = append(diff.Reorders, ComponentReorder{
				Parent:   component,
				NewOrder: newChildren,
			})
		}
	}

	// Check for name changes with the same key
	for key, oldData := range oldKeyToIndex {
		if newIndex, ok := newKeyToIndex[key]; ok {
			// Find the old and new components
			oldChild := originalComp.GetChildren()[oldData]
			newChild := component.GetChildren()[newIndex]

			// Check if the name has changed
			if oldChild.GetName() != newChild.GetName() {
				// Check if this component has already been updated
				if !updatedComps[newChild] {
					// Add an update for the name change
					diff.Updates = append(diff.Updates, ComponentUpdate{
						Component:    newChild,
						ChangedProps: []string{"name"},
					})
				}
			}
		}
	}
}

// detectRootReordering specifically handles the reordering of children in the root component
func (d *ComponentDiffer) detectRootReordering(
	snapshot *ComponentSnapshot,
	root *ComponentManager,
	diff *ComponentDiff,
) {
	if !d.useKeyReconciliation || d.reconciliationKey == "" {
		return
	}

	// Get the old root and its children
	oldRoot := snapshot.Root
	if oldRoot == nil || len(oldRoot.GetChildren()) == 0 || len(root.GetChildren()) == 0 {
		return
	}

	// Map keys to positions in old and new trees
	oldKeyToIndex := make(map[interface{}]int)
	newKeyToIndex := make(map[interface{}]int)

	// Map old keys to indices
	for i, child := range oldRoot.GetChildren() {
		if key, exists := child.GetProp(d.reconciliationKey); exists {
			oldKeyToIndex[key] = i
		}
	}

	// Map new keys to indices
	for i, child := range root.GetChildren() {
		if key, exists := child.GetProp(d.reconciliationKey); exists {
			newKeyToIndex[key] = i
		}
	}

	// Count common elements with the same key
	commonElements := 0
	for key := range oldKeyToIndex {
		if _, exists := newKeyToIndex[key]; exists {
			commonElements++
		}
	}

	// We need at least 2 common elements to detect meaningful reordering
	if commonElements < 2 {
		return
	}

	// Check if any common element changed position
	orderChanged := false
	for key, oldPos := range oldKeyToIndex {
		if newPos, exists := newKeyToIndex[key]; exists {
			if oldPos != newPos {
				orderChanged = true
				break
			}
		}
	}

	if orderChanged {
		// Add a reorder operation
		diff.Reorders = append(diff.Reorders, ComponentReorder{
			Parent:   root,
			NewOrder: root.GetChildren(),
		})
	}
}

// Helper functions
func splitPath(path string) []string {
	// Split a path into component names
	// Example: "Root/Child1/Child2" -> ["Root", "Child1", "Child2"]
	if path == "" {
		return []string{}
	}

	// Simple split by '/' - in a real implementation, handle escaping
	parts := strings.Split(path, "/")
	return parts
}

func joinPath(parts []string) string {
	// Join path components into a full path
	// Example: ["Root", "Child1", "Child2"] -> "Root/Child1/Child2"
	return strings.Join(parts, "/")
}

func findComponentByPath(root *ComponentManager, path string) *ComponentManager {
	// If path is empty, return nil
	if path == "" {
		return nil
	}

	// Split the path into parts
	parts := splitPath(path)

	// Start from the root
	current := root

	// If the first part doesn't match the root name, it's not a valid path
	if len(parts) > 0 && parts[0] != root.GetName() {
		return nil
	}

	// If the path is just the root name, return the root
	if len(parts) == 1 {
		return root
	}

	// Traverse the tree to find the component
	for i := 1; i < len(parts); i++ {
		found := false

		// Search through children
		for _, child := range current.GetChildren() {
			if child.GetName() == parts[i] {
				current = child
				found = true
				break
			}
		}

		// If no matching child found, the path is invalid
		if !found {
			return nil
		}
	}

	return current
}
