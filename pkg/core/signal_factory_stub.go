// This file contains simplified stubs for the signal factory system
// It's intended to be used as a bridge during development of the reactive state system
// and will be integrated with the full implementation in signal_factory.go

package core

import (
	"fmt"
	"sync"
	"time"
)

// LocalSignalFactory provides a minimal implementation of signal factory functions
// for use during the development of the reactive state system
type LocalSignalFactory struct {
	mutex    sync.RWMutex
	debugOn  bool
	registry map[string]interface{}
}

// NewLocalSignalFactory creates a new instance of the local signal factory
func NewLocalSignalFactory() *LocalSignalFactory {
	return &LocalSignalFactory{
		registry: make(map[string]interface{}),
	}
}

// EnableDebug turns on debug mode for the local signal factory
func (f *LocalSignalFactory) EnableDebug() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.debugOn = true
}

// DisableDebug turns off debug mode for the local signal factory
func (f *LocalSignalFactory) DisableDebug() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.debugOn = false
}

// CreateLocalEffect creates a simplified effect that can be used during development
func (f *LocalSignalFactory) CreateLocalEffect(fn func()) string {
	// Generate a unique ID for this effect
	id := fmt.Sprintf("effect_%d", time.Now().UnixNano())

	// Execute the effect function once
	func() {
		// Use defer/recover to handle any panics during execution
		defer func() {
			if r := recover(); r != nil {
				// In the future, this would use proper error handling
				// For now, just recover and continue
			}
		}()

		// Run the effect function
		fn()
	}()

	return id
}

// RegisterLocalCleanup registers a cleanup function for a local effect
func (f *LocalSignalFactory) RegisterLocalCleanup(effectID string, cleanupFn func() error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// In the full implementation, this would register the cleanup function with the effect
	// For now, just log that it was called (if debug is enabled)
	if f.debugOn {
		fmt.Printf("[DEBUG] Registered cleanup for effect: %s\n", effectID)
	}
}

// Initialization for the local signal factory
var DefaultLocalSignalFactory = NewLocalSignalFactory()
