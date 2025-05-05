package core

import (
	"fmt"
	"sort"
)

// ScheduleEffect schedules an effect to run with a specific priority.
// Higher priority effects will run before lower priority ones.
func ScheduleEffect(effectID string, priority EffectPriority) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	// Check if effect exists
	if _, exists := effectsRegistry[effectID]; !exists {
		return
	}

	// Set or update the effect's priority
	info, exists := effectInfos[effectID]
	if !exists {
		info = EffectInfo{
			Priority:    PriorityNormal, // Default priority
			IsCancelled: false,
			IsDeferred:  false,
		}
	}
	info.Priority = priority
	effectInfos[effectID] = info

	// Add to queue if not already present
	if pendingEffects[effectID] && !info.IsCancelled {
		addToEffectQueue(effectID)
	}
}

// AddToEffectBatch adds an effect to a batch, so that related effects can be
// grouped and processed together.
func AddToEffectBatch(effectID string, batchID string) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	// Check if effect exists
	if _, exists := effectsRegistry[effectID]; !exists {
		return
	}

	// Set or update the effect's batch ID
	info, exists := effectInfos[effectID]
	if !exists {
		info = EffectInfo{
			Priority:    PriorityNormal, // Default priority
			IsCancelled: false,
			IsDeferred:  false,
		}
	}
	info.BatchID = batchID
	effectInfos[effectID] = info
}

// DeferEffect marks an effect to be run after all non-deferred effects have completed
func DeferEffect(effectID string) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	// Check if effect exists
	if _, exists := effectsRegistry[effectID]; !exists {
		return
	}

	// Set the effect as deferred
	info, exists := effectInfos[effectID]
	if !exists {
		info = EffectInfo{
			Priority:    PriorityNormal, // Default priority
			IsCancelled: false,
			IsDeferred:  false,
		}
	}
	info.IsDeferred = true
	effectInfos[effectID] = info
}

// CancelEffect prevents an effect from executing in the current cycle
func CancelEffect(effectID string) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	// Check if effect exists
	if _, exists := effectsRegistry[effectID]; !exists {
		return
	}

	// Mark the effect as cancelled
	info, exists := effectInfos[effectID]
	if !exists {
		info = EffectInfo{
			Priority:    PriorityNormal, // Default priority
			IsCancelled: false,
			IsDeferred:  false,
		}
	}
	info.IsCancelled = true
	effectInfos[effectID] = info

	// Remove from pending effects
	delete(pendingEffects, effectID)
}

// addToEffectQueue adds an effect to the execution queue
func addToEffectQueue(effectID string) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	// Check if already in queue
	for _, id := range effectQueue {
		if id == effectID {
			return
		}
	}

	// Add to queue
	effectQueue = append(effectQueue, effectID)

	// Process queue if not already processing and not in batch mode
	if !processingQueue && !batchMode {
		// Set processing to true before releasing lock and launching goroutine
		processingQueue = true
		go processEffectQueue()
	}
}

// processEffectQueue processes the effect queue based on priorities and batches
func processEffectQueue() {
	// Note: processingQueue flag is already set to true by addToEffectQueue
	defer func() {
		globalMutex.Lock()
		processingQueue = false
		globalMutex.Unlock()
	}()

	// Already setup the defer at the beginning of the function

	// Create a copy of the queue for processing
	globalMutex.Lock()
	queueCopy := make([]string, len(effectQueue))
	copy(queueCopy, effectQueue)
	effectQueue = []string{} // Clear the queue
	globalMutex.Unlock()

	// Sort effects by priority (highest first) and deferred status
	infoCopy := make(map[string]EffectInfo, len(queueCopy))

	// First, safely collect all the effect info we need for sorting
	globalMutex.RLock()
	for _, id := range queueCopy {
		if info, exists := effectInfos[id]; exists {
			infoCopy[id] = info
		} else {
			// Use default values if not found
			infoCopy[id] = EffectInfo{Priority: PriorityNormal}
		}
	}
	globalMutex.RUnlock()

	// Now sort with the copied data
	sort.SliceStable(queueCopy, func(i, j int) bool {
		deferi := infoCopy[queueCopy[i]].IsDeferred
		deferj := infoCopy[queueCopy[j]].IsDeferred
		priorityi := infoCopy[queueCopy[i]].Priority
		priorityj := infoCopy[queueCopy[j]].Priority

		// Non-deferred effects come before deferred ones
		if deferi != deferj {
			return !deferi
		}

		// Higher priority comes first
		return priorityi > priorityj
	})

	// Group effects by batch
	batchedEffects := make(map[string][]string)
	normalEffects := []string{}

	for _, effectID := range queueCopy {
		// Use our local copy of effect info instead of locking again
		info := infoCopy[effectID]

		if info.IsCancelled {
			continue
		}

		if info.BatchID != "" {
			batchedEffects[info.BatchID] = append(batchedEffects[info.BatchID], effectID)
		} else {
			normalEffects = append(normalEffects, effectID)
		}
	}

	// Execute batched effects first, grouped by batch ID
	for batchID, effects := range batchedEffects {
		if debugMode {
			fmt.Printf("[DEBUG] Processing batch %s with %d effects\n", batchID, len(effects))
		}

		// Sort effects within batch by priority using our copy
		sort.SliceStable(effects, func(i, j int) bool {
			priorityi := infoCopy[effects[i]].Priority
			priorityj := infoCopy[effects[j]].Priority
			return priorityi > priorityj
		})

		// Execute each effect in the batch
		for _, effectID := range effects {
			executeEffect(effectID)
		}
	}

	// Execute non-batched effects
	for _, effectID := range normalEffects {
		executeEffect(effectID)
	}
}

// executeEffect actually runs an effect if it's not cancelled
func executeEffect(effectID string) {
	// Check if effect exists and get a reference to it
	globalMutex.RLock()
	effect, exists := effectsRegistry[effectID]
	info, infoExists := effectInfos[effectID]
	globalMutex.RUnlock()

	// Skip if effect doesn't exist or is cancelled
	if !exists || (infoExists && info.IsCancelled) {
		return
	}

	// Execute the effect
	if debugMode {
		var debugInfo string
		if e, ok := effect.(*Effect); ok && e.debugInfo != "" {
			debugInfo = e.debugInfo
		} else {
			debugInfo = effectID
		}
		fmt.Printf("[DEBUG] Executing effect: %s\n", debugInfo)
	}

	effect.Notify()

	// Remove from pending effects
	globalMutex.Lock()
	delete(pendingEffects, effectID)
	globalMutex.Unlock()
}
