package testutil

import (
	"math/rand"
)

// DataFactory is a generic factory for generating test data.
// It uses a generator function to produce values of type T on demand.
//
// DataFactory is useful for creating test fixtures with consistent or random data.
// The generator function is called each time Generate() or GenerateN() is invoked.
//
// Example usage:
//
//	// Create a factory for constant values
//	factory := testutil.NewFactory(func() int { return 42 })
//	value := factory.Generate()  // Returns 42
//
//	// Create a factory for sequential values
//	counter := 0
//	factory := testutil.NewFactory(func() int {
//	    counter++
//	    return counter
//	})
//	values := factory.GenerateN(3)  // Returns []int{1, 2, 3}
//
//	// Create a factory for complex types
//	type User struct { Name string; Age int }
//	factory := testutil.NewFactory(func() User {
//	    return User{Name: "John", Age: 30}
//	})
type DataFactory[T any] struct {
	generator func() T
}

// NewFactory creates a new DataFactory with the given generator function.
// The generator function will be called each time Generate() or GenerateN() is invoked.
//
// The generator function can be stateless (returning constant values) or stateful
// (using closures to maintain state between calls).
//
// Example:
//
//	// Stateless generator
//	factory := NewFactory(func() int { return 42 })
//
//	// Stateful generator with closure
//	counter := 0
//	factory := NewFactory(func() int {
//	    counter++
//	    return counter
//	})
func NewFactory[T any](generator func() T) *DataFactory[T] {
	return &DataFactory[T]{
		generator: generator,
	}
}

// Generate produces a single value by calling the generator function.
//
// Example:
//
//	factory := NewFactory(func() int { return 42 })
//	value := factory.Generate()  // Returns 42
func (df *DataFactory[T]) Generate() T {
	return df.generator()
}

// GenerateN produces n values by calling the generator function n times.
// Returns a slice containing all generated values.
//
// If n is 0, returns an empty slice (not nil).
// If n is negative, returns an empty slice.
//
// Example:
//
//	factory := NewFactory(func() int { return 42 })
//	values := factory.GenerateN(5)  // Returns []int{42, 42, 42, 42, 42}
//
//	counter := 0
//	factory := NewFactory(func() int {
//	    counter++
//	    return counter
//	})
//	values := factory.GenerateN(3)  // Returns []int{1, 2, 3}
func (df *DataFactory[T]) GenerateN(n int) []T {
	if n <= 0 {
		return []T{}
	}

	result := make([]T, n)
	for i := 0; i < n; i++ {
		result[i] = df.generator()
	}

	return result
}

// IntFactory creates a DataFactory that generates random integers in the range [min, max] (inclusive).
//
// If min == max, always returns that value.
// If min > max, the behavior is undefined (may panic or return unexpected values).
//
// Example:
//
//	factory := IntFactory(0, 100)
//	value := factory.Generate()  // Returns random int between 0 and 100
//	values := factory.GenerateN(10)  // Returns 10 random ints
//
//	// Generate constant value
//	factory := IntFactory(42, 42)
//	value := factory.Generate()  // Always returns 42
func IntFactory(min, max int) *DataFactory[int] {
	return NewFactory(func() int {
		if min == max {
			return min
		}
		return min + rand.Intn(max-min+1)
	})
}

// StringFactory creates a DataFactory that generates random strings of the specified length.
// The strings contain only ASCII letters (a-z, A-Z).
//
// If length is 0, returns empty strings.
// If length is negative, returns empty strings.
//
// Example:
//
//	factory := StringFactory(10)
//	value := factory.Generate()  // Returns random 10-character string
//	values := factory.GenerateN(5)  // Returns 5 random strings
//
//	// Generate empty strings
//	factory := StringFactory(0)
//	value := factory.Generate()  // Returns ""
func StringFactory(length int) *DataFactory[string] {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	return NewFactory(func() string {
		if length <= 0 {
			return ""
		}

		result := make([]byte, length)
		for i := 0; i < length; i++ {
			result[i] = letters[rand.Intn(len(letters))]
		}

		return string(result)
	})
}
