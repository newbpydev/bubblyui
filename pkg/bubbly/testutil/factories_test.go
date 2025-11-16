package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewFactory tests factory creation with custom generators
func TestNewFactory(t *testing.T) {
	tests := []struct {
		name      string
		generator func() int
		expected  int
	}{
		{
			name:      "constant generator",
			generator: func() int { return 42 },
			expected:  42,
		},
		{
			name: "counter generator",
			generator: func() func() int {
				count := 0
				return func() int {
					count++
					return count
				}
			}(),
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewFactory(tt.generator)
			assert.NotNil(t, factory)
			assert.NotNil(t, factory.generator)

			// Verify generator works
			result := factory.generator()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDataFactory_Generate tests single value generation
func TestDataFactory_Generate(t *testing.T) {
	tests := []struct {
		name      string
		generator func() interface{}
		expected  interface{}
	}{
		{
			name:      "int generator",
			generator: func() interface{} { return 42 },
			expected:  42,
		},
		{
			name:      "string generator",
			generator: func() interface{} { return "test" },
			expected:  "test",
		},
		{
			name:      "bool generator",
			generator: func() interface{} { return true },
			expected:  true,
		},
		{
			name:      "struct generator",
			generator: func() interface{} { return struct{ Name string }{Name: "John"} },
			expected:  struct{ Name string }{Name: "John"},
		},
		{
			name:      "slice generator",
			generator: func() interface{} { return []int{1, 2, 3} },
			expected:  []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewFactory(tt.generator)
			result := factory.Generate()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDataFactory_GenerateN tests multiple value generation
func TestDataFactory_GenerateN(t *testing.T) {
	tests := []struct {
		name      string
		generator func() int
		n         int
		validate  func(*testing.T, []int)
	}{
		{
			name:      "generate zero values",
			generator: func() int { return 42 },
			n:         0,
			validate: func(t *testing.T, result []int) {
				assert.Empty(t, result)
				assert.NotNil(t, result) // Should be empty slice, not nil
			},
		},
		{
			name:      "generate one value",
			generator: func() int { return 42 },
			n:         1,
			validate: func(t *testing.T, result []int) {
				assert.Len(t, result, 1)
				assert.Equal(t, 42, result[0])
			},
		},
		{
			name:      "generate multiple values",
			generator: func() int { return 42 },
			n:         5,
			validate: func(t *testing.T, result []int) {
				assert.Len(t, result, 5)
				for _, v := range result {
					assert.Equal(t, 42, v)
				}
			},
		},
		{
			name: "generate sequential values",
			generator: func() func() int {
				count := 0
				return func() int {
					count++
					return count
				}
			}(),
			n: 3,
			validate: func(t *testing.T, result []int) {
				assert.Len(t, result, 3)
				assert.Equal(t, []int{1, 2, 3}, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewFactory(tt.generator)
			result := factory.GenerateN(tt.n)
			tt.validate(t, result)
		})
	}
}

// TestIntFactory tests integer factory
func TestIntFactory(t *testing.T) {
	tests := []struct {
		name     string
		min      int
		max      int
		count    int
		validate func(*testing.T, []int, int, int)
	}{
		{
			name:  "single value in range",
			min:   0,
			max:   10,
			count: 1,
			validate: func(t *testing.T, values []int, min, max int) {
				assert.Len(t, values, 1)
				assert.GreaterOrEqual(t, values[0], min)
				assert.LessOrEqual(t, values[0], max)
			},
		},
		{
			name:  "multiple values in range",
			min:   0,
			max:   100,
			count: 10,
			validate: func(t *testing.T, values []int, min, max int) {
				assert.Len(t, values, 10)
				for _, v := range values {
					assert.GreaterOrEqual(t, v, min)
					assert.LessOrEqual(t, v, max)
				}
			},
		},
		{
			name:  "negative range",
			min:   -50,
			max:   -10,
			count: 5,
			validate: func(t *testing.T, values []int, min, max int) {
				assert.Len(t, values, 5)
				for _, v := range values {
					assert.GreaterOrEqual(t, v, min)
					assert.LessOrEqual(t, v, max)
				}
			},
		},
		{
			name:  "single value range (min == max)",
			min:   42,
			max:   42,
			count: 3,
			validate: func(t *testing.T, values []int, min, max int) {
				assert.Len(t, values, 3)
				for _, v := range values {
					assert.Equal(t, 42, v)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := IntFactory(tt.min, tt.max)
			assert.NotNil(t, factory)

			values := factory.GenerateN(tt.count)
			tt.validate(t, values, tt.min, tt.max)
		})
	}
}

// TestStringFactory tests string factory
func TestStringFactory(t *testing.T) {
	tests := []struct {
		name     string
		length   int
		count    int
		validate func(*testing.T, []string, int)
	}{
		{
			name:   "zero length string",
			length: 0,
			count:  1,
			validate: func(t *testing.T, values []string, length int) {
				assert.Len(t, values, 1)
				assert.Equal(t, "", values[0])
			},
		},
		{
			name:   "single character string",
			length: 1,
			count:  5,
			validate: func(t *testing.T, values []string, length int) {
				assert.Len(t, values, 5)
				for _, v := range values {
					assert.Len(t, v, length)
					// Verify it's a letter
					assert.Regexp(t, "^[a-zA-Z]$", v)
				}
			},
		},
		{
			name:   "short string",
			length: 5,
			count:  3,
			validate: func(t *testing.T, values []string, length int) {
				assert.Len(t, values, 3)
				for _, v := range values {
					assert.Len(t, v, length)
					// Verify all characters are letters
					assert.Regexp(t, "^[a-zA-Z]+$", v)
				}
			},
		},
		{
			name:   "long string",
			length: 100,
			count:  2,
			validate: func(t *testing.T, values []string, length int) {
				assert.Len(t, values, 2)
				for _, v := range values {
					assert.Len(t, v, length)
					assert.Regexp(t, "^[a-zA-Z]+$", v)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := StringFactory(tt.length)
			assert.NotNil(t, factory)

			values := factory.GenerateN(tt.count)
			tt.validate(t, values, tt.length)
		})
	}
}

// TestDataFactory_TypeSafety tests type safety with different types
func TestDataFactory_TypeSafety(t *testing.T) {
	t.Run("int factory", func(t *testing.T) {
		factory := NewFactory(func() int { return 42 })
		result := factory.Generate()
		assert.Equal(t, 42, result)
	})

	t.Run("string factory", func(t *testing.T) {
		factory := NewFactory(func() string { return "test" })
		result := factory.Generate()
		assert.Equal(t, "test", result)
	})

	t.Run("bool factory", func(t *testing.T) {
		factory := NewFactory(func() bool { return true })
		result := factory.Generate()
		assert.Equal(t, true, result)
	})

	t.Run("struct factory", func(t *testing.T) {
		type User struct {
			Name string
			Age  int
		}
		factory := NewFactory(func() User {
			return User{Name: "John", Age: 30}
		})
		result := factory.Generate()
		assert.Equal(t, User{Name: "John", Age: 30}, result)
	})

	t.Run("slice factory", func(t *testing.T) {
		factory := NewFactory(func() []int { return []int{1, 2, 3} })
		result := factory.Generate()
		assert.Equal(t, []int{1, 2, 3}, result)
	})
}

// TestDataFactory_CustomFactories tests custom factory patterns
func TestDataFactory_CustomFactories(t *testing.T) {
	t.Run("email factory", func(t *testing.T) {
		emailFactory := NewFactory(func() string {
			return "user@example.com"
		})
		emails := emailFactory.GenerateN(3)
		assert.Len(t, emails, 3)
		for _, email := range emails {
			assert.Equal(t, "user@example.com", email)
		}
	})

	t.Run("uuid factory", func(t *testing.T) {
		counter := 0
		uuidFactory := NewFactory(func() string {
			counter++
			return "uuid-" + string(rune('0'+counter))
		})
		uuids := uuidFactory.GenerateN(3)
		assert.Len(t, uuids, 3)
		assert.Equal(t, "uuid-1", uuids[0])
		assert.Equal(t, "uuid-2", uuids[1])
		assert.Equal(t, "uuid-3", uuids[2])
	})

	t.Run("complex struct factory", func(t *testing.T) {
		type Address struct {
			Street string
			City   string
		}
		type Person struct {
			Name    string
			Age     int
			Address Address
		}

		personFactory := NewFactory(func() Person {
			return Person{
				Name: "John Doe",
				Age:  30,
				Address: Address{
					Street: "123 Main St",
					City:   "Springfield",
				},
			}
		})

		people := personFactory.GenerateN(2)
		assert.Len(t, people, 2)
		for _, person := range people {
			assert.Equal(t, "John Doe", person.Name)
			assert.Equal(t, 30, person.Age)
			assert.Equal(t, "123 Main St", person.Address.Street)
		}
	})
}
