package bubbly

// ProvideKey is a type-safe key for provide/inject dependency injection.
// It wraps a string key with type information to enable compile-time type checking.
//
// The generic type parameter T specifies what type of value this key expects.
// This prevents accidentally injecting the wrong type at compile time.
//
// Example:
//
//	// Define a typed key for theme configuration
//	var ThemeKey = NewProvideKey[string]("app.theme")
//
//	// Provide a value
//	ProvideTyped(ctx, ThemeKey, "dark")
//
//	// Inject with type safety - returns string, not interface{}
//	theme := InjectTyped(ctx, ThemeKey, "light")
type ProvideKey[T any] struct {
	key string
}

// NewProvideKey creates a new type-safe provide/inject key.
//
// The generic type parameter T specifies what type of value this key will be used for.
// The key string should be unique within your application to avoid conflicts.
//
// Common patterns for key naming:
//   - Dot notation: "app.theme", "user.preferences"
//   - Slash notation: "config/database/url"
//   - Simple names: "theme", "user", "config"
//
// Example:
//
//	// Create keys for different types
//	var ThemeKey = NewProvideKey[string]("theme")
//	var CountKey = NewProvideKey[int]("count")
//	var UserKey = NewProvideKey[*User]("currentUser")
//	var ConfigKey = NewProvideKey[*Ref[Config]]("appConfig")
func NewProvideKey[T any](key string) ProvideKey[T] {
	return ProvideKey[T]{key: key}
}

// ProvideTyped stores a typed value in the component's provides map,
// making it available for injection by child components via InjectTyped().
//
// This is a type-safe wrapper around Context.Provide() that ensures
// the provided value matches the key's type parameter at compile time.
//
// The value can be of any type, including:
//   - Primitive types: string, int, bool, etc.
//   - Structs and pointers
//   - Reactive values: *Ref[T], *Computed[T]
//   - Functions and interfaces
//
// When a reactive value (Ref or Computed) is provided, all injecting
// components share the same instance and see updates automatically.
//
// Example:
//
//	// Provide a simple string
//	themeKey := NewProvideKey[string]("theme")
//	ProvideTyped(ctx, themeKey, "dark")
//
//	// Provide a reactive Ref
//	countKey := NewProvideKey[*Ref[int]]("count")
//	count := ctx.Ref(0)
//	ProvideTyped(ctx, countKey, count)
//
//	// Provide a struct
//	configKey := NewProvideKey[Config]("config")
//	ProvideTyped(ctx, configKey, Config{Theme: "dark", Debug: true})
func ProvideTyped[T any](ctx *Context, key ProvideKey[T], value T) {
	ctx.Provide(key.key, value)
}

// InjectTyped retrieves a typed value provided by an ancestor component.
//
// This is a type-safe wrapper around Context.Inject() that:
//   - Returns the actual type T instead of interface{}
//   - Ensures the default value matches the key's type at compile time
//   - Eliminates the need for type assertions
//
// It walks up the component tree looking for the first component that
// provided the specified key. If not found, returns the defaultValue.
//
// The nearest provider wins - if both a parent and grandparent provide
// the same key, the parent's value is returned.
//
// Example:
//
//	// Inject with default - returns string, not interface{}
//	themeKey := NewProvideKey[string]("theme")
//	theme := InjectTyped(ctx, themeKey, "light")
//
//	// Inject reactive Ref - no type assertion needed
//	countKey := NewProvideKey[*Ref[int]]("count")
//	count := InjectTyped(ctx, countKey, ctx.Ref(0))
//	count.Set(count.GetTyped() + 1)  // Direct access, no casting
//
//	// Inject optional dependency with nil default
//	userKey := NewProvideKey[*User]("currentUser")
//	user := InjectTyped(ctx, userKey, nil)
//	if user != nil {
//	    // Use user
//	}
func InjectTyped[T any](ctx *Context, key ProvideKey[T], defaultValue T) T {
	val := ctx.Inject(key.key, defaultValue)
	return val.(T)
}
