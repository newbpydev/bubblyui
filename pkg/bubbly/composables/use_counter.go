package composables

import (
	"math"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// counterConfig holds configuration for UseCounter.
type counterConfig struct {
	min    int
	max    int
	step   int
	hasMin bool
	hasMax bool
}

// CounterOption configures UseCounter.
type CounterOption func(*counterConfig)

// WithMin sets minimum counter value.
// When set, the counter will not go below this value.
// Decrement, DecrementBy, and Set operations will clamp to this minimum.
//
// Example:
//
//	counter := UseCounter(ctx, 50, WithMin(0))
//	counter.Set(-10)  // Clamped to 0
func WithMin(min int) CounterOption {
	return func(c *counterConfig) {
		c.min = min
		c.hasMin = true
	}
}

// WithMax sets maximum counter value.
// When set, the counter will not go above this value.
// Increment, IncrementBy, and Set operations will clamp to this maximum.
//
// Example:
//
//	counter := UseCounter(ctx, 50, WithMax(100))
//	counter.Set(150)  // Clamped to 100
func WithMax(max int) CounterOption {
	return func(c *counterConfig) {
		c.max = max
		c.hasMax = true
	}
}

// WithStep sets increment/decrement step size.
// Default step is 1. This affects Increment() and Decrement() methods.
//
// Example:
//
//	counter := UseCounter(ctx, 0, WithStep(5))
//	counter.Increment()  // Now 5
//	counter.Increment()  // Now 10
func WithStep(step int) CounterOption {
	return func(c *counterConfig) {
		c.step = step
	}
}

// CounterReturn is the return value of UseCounter.
// It provides reactive counter state management with methods for
// incrementing, decrementing, and controlling the value within optional bounds.
type CounterReturn struct {
	// Count is the current counter value.
	Count *bubbly.Ref[int]

	// config holds min/max/step settings.
	config counterConfig

	// initial is the starting value for reset.
	initial int
}

// clamp constrains value to configured bounds.
func (c *CounterReturn) clamp(value int) int {
	if c.config.hasMin && value < c.config.min {
		return c.config.min
	}
	if c.config.hasMax && value > c.config.max {
		return c.config.max
	}
	return value
}

// Increment increases count by step (respects max).
// If no step is configured, increments by 1.
//
// Example:
//
//	counter := composables.UseCounter(ctx, 0)
//	counter.Increment()  // Now 1
//	counter.Increment()  // Now 2
func (c *CounterReturn) Increment() {
	current := c.Count.GetTyped()
	newValue := c.clamp(current + c.config.step)
	c.Count.Set(newValue)
}

// Decrement decreases count by step (respects min).
// If no step is configured, decrements by 1.
//
// Example:
//
//	counter := composables.UseCounter(ctx, 10)
//	counter.Decrement()  // Now 9
//	counter.Decrement()  // Now 8
func (c *CounterReturn) Decrement() {
	current := c.Count.GetTyped()
	newValue := c.clamp(current - c.config.step)
	c.Count.Set(newValue)
}

// IncrementBy increases count by n (respects max).
// This ignores the configured step and uses the provided value directly.
//
// Example:
//
//	counter := composables.UseCounter(ctx, 0)
//	counter.IncrementBy(10)  // Now 10
//	counter.IncrementBy(5)   // Now 15
func (c *CounterReturn) IncrementBy(n int) {
	current := c.Count.GetTyped()
	newValue := c.clamp(current + n)
	c.Count.Set(newValue)
}

// DecrementBy decreases count by n (respects min).
// This ignores the configured step and uses the provided value directly.
//
// Example:
//
//	counter := composables.UseCounter(ctx, 20)
//	counter.DecrementBy(5)   // Now 15
//	counter.DecrementBy(10)  // Now 5
func (c *CounterReturn) DecrementBy(n int) {
	current := c.Count.GetTyped()
	newValue := c.clamp(current - n)
	c.Count.Set(newValue)
}

// Set sets the count to a specific value (clamped to bounds).
// If min/max bounds are configured, the value will be clamped.
//
// Example:
//
//	counter := composables.UseCounter(ctx, 0, WithMin(0), WithMax(100))
//	counter.Set(50)   // Now 50
//	counter.Set(150)  // Clamped to 100
//	counter.Set(-10)  // Clamped to 0
func (c *CounterReturn) Set(n int) {
	newValue := c.clamp(n)
	c.Count.Set(newValue)
}

// Reset resets to initial value.
// The initial value is the value passed to UseCounter when creating the composable.
//
// Example:
//
//	counter := composables.UseCounter(ctx, 50)
//	counter.IncrementBy(100)  // Now 150
//	counter.Reset()           // Back to 50
func (c *CounterReturn) Reset() {
	c.Count.Set(c.initial)
}

// UseCounter creates a bounded counter composable.
// It provides a simple API for managing integer counter state with optional
// min/max bounds and configurable step size.
//
// This composable is useful for:
//   - Volume controls (0-100)
//   - Quantity selectors
//   - Pagination (page numbers)
//   - Rating systems
//   - Any bounded integer state management
//
// Parameters:
//   - ctx: The component context (required for all composables)
//   - initial: The initial counter value
//   - opts: Optional configuration (WithMin, WithMax, WithStep)
//
// Returns:
//   - *CounterReturn: A struct containing the reactive counter value and methods
//
// Example:
//
//	Setup(func(ctx *bubbly.Context) {
//	    // Create a volume counter (0-100, step 5)
//	    volume := composables.UseCounter(ctx, 50,
//	        composables.WithMin(0),
//	        composables.WithMax(100),
//	        composables.WithStep(5))
//	    ctx.Expose("volume", volume)
//
//	    // Increment on key press
//	    ctx.On("volumeUp", func(_ interface{}) {
//	        volume.Increment()
//	    })
//
//	    ctx.On("volumeDown", func(_ interface{}) {
//	        volume.Decrement()
//	    })
//
//	    ctx.On("mute", func(_ interface{}) {
//	        volume.Set(0)
//	    })
//
//	    ctx.On("resetVolume", func(_ interface{}) {
//	        volume.Reset()
//	    })
//	}).
//	WithKeyBinding("+", "volumeUp", "Volume up").
//	WithKeyBinding("-", "volumeDown", "Volume down").
//	WithKeyBinding("m", "mute", "Mute").
//	WithKeyBinding("r", "resetVolume", "Reset volume")
//
// Template usage:
//
//	Template(func(ctx bubbly.RenderContext) string {
//	    volume := ctx.Get("volume").(*composables.CounterReturn)
//	    return fmt.Sprintf("Volume: %d%%", volume.Count.GetTyped())
//	})
//
// Integration with CreateShared:
//
//	var UseSharedCounter = composables.CreateShared(
//	    func(ctx *bubbly.Context) *composables.CounterReturn {
//	        return composables.UseCounter(ctx, 0)
//	    },
//	)
func UseCounter(ctx *bubbly.Context, initial int, opts ...CounterOption) *CounterReturn {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseCounter", time.Since(start))
	}()

	// Apply options with defaults
	config := counterConfig{
		min:    math.MinInt,
		max:    math.MaxInt,
		step:   1, // Default step is 1
		hasMin: false,
		hasMax: false,
	}
	for _, opt := range opts {
		opt(&config)
	}

	// Clamp initial value to bounds if needed
	clampedInitial := initial
	if config.hasMin && clampedInitial < config.min {
		clampedInitial = config.min
	}
	if config.hasMax && clampedInitial > config.max {
		clampedInitial = config.max
	}

	// Create reactive ref with clamped initial value
	count := bubbly.NewRef(clampedInitial)

	return &CounterReturn{
		Count:   count,
		config:  config,
		initial: clampedInitial,
	}
}
