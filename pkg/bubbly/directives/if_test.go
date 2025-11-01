package directives

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// TestIfDirective_Simple tests basic If directive functionality
func TestIfDirective_Simple(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		thenFunc  func() string
		expected  string
	}{
		{
			name:      "condition true renders then branch",
			condition: true,
			thenFunc:  func() string { return "then branch" },
			expected:  "then branch",
		},
		{
			name:      "condition false renders empty",
			condition: false,
			thenFunc:  func() string { return "then branch" },
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := If(tt.condition, tt.thenFunc).Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIfDirective_WithElse tests If with Else branch
func TestIfDirective_WithElse(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		thenFunc  func() string
		elseFunc  func() string
		expected  string
	}{
		{
			name:      "condition true renders then branch",
			condition: true,
			thenFunc:  func() string { return "then" },
			elseFunc:  func() string { return "else" },
			expected:  "then",
		},
		{
			name:      "condition false renders else branch",
			condition: false,
			thenFunc:  func() string { return "then" },
			elseFunc:  func() string { return "else" },
			expected:  "else",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := If(tt.condition, tt.thenFunc).Else(tt.elseFunc).Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIfDirective_ElseIfChain tests ElseIf chaining
func TestIfDirective_ElseIfChain(t *testing.T) {
	tests := []struct {
		name       string
		condition1 bool
		condition2 bool
		condition3 bool
		expected   string
	}{
		{
			name:       "first condition true",
			condition1: true,
			condition2: false,
			condition3: false,
			expected:   "first",
		},
		{
			name:       "second condition true",
			condition1: false,
			condition2: true,
			condition3: false,
			expected:   "second",
		},
		{
			name:       "third condition true",
			condition1: false,
			condition2: false,
			condition3: true,
			expected:   "third",
		},
		{
			name:       "all conditions false with else",
			condition1: false,
			condition2: false,
			condition3: false,
			expected:   "else",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := If(tt.condition1, func() string { return "first" }).
				ElseIf(tt.condition2, func() string { return "second" }).
				ElseIf(tt.condition3, func() string { return "third" }).
				Else(func() string { return "else" }).
				Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIfDirective_ElseIfWithoutElse tests ElseIf without final Else
func TestIfDirective_ElseIfWithoutElse(t *testing.T) {
	tests := []struct {
		name       string
		condition1 bool
		condition2 bool
		expected   string
	}{
		{
			name:       "first condition true",
			condition1: true,
			condition2: false,
			expected:   "first",
		},
		{
			name:       "second condition true",
			condition1: false,
			condition2: true,
			expected:   "second",
		},
		{
			name:       "all conditions false returns empty",
			condition1: false,
			condition2: false,
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := If(tt.condition1, func() string { return "first" }).
				ElseIf(tt.condition2, func() string { return "second" }).
				Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIfDirective_Nested tests nested If directives
func TestIfDirective_Nested(t *testing.T) {
	tests := []struct {
		name      string
		outerCond bool
		innerCond bool
		expected  string
	}{
		{
			name:      "both conditions true",
			outerCond: true,
			innerCond: true,
			expected:  "inner true",
		},
		{
			name:      "outer true inner false",
			outerCond: true,
			innerCond: false,
			expected:  "inner false",
		},
		{
			name:      "outer false",
			outerCond: false,
			innerCond: true,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := If(tt.outerCond, func() string {
				return If(tt.innerCond, func() string {
					return "inner true"
				}).Else(func() string {
					return "inner false"
				}).Render()
			}).Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIfDirective_EmptyConditions tests edge cases with empty/nil functions
func TestIfDirective_EmptyConditions(t *testing.T) {
	t.Run("empty then function returns empty", func(t *testing.T) {
		result := If(true, func() string { return "" }).Render()
		assert.Equal(t, "", result)
	})

	t.Run("empty else function returns empty", func(t *testing.T) {
		result := If(false, func() string { return "then" }).
			Else(func() string { return "" }).
			Render()
		assert.Equal(t, "", result)
	})

	t.Run("multiple elseif with empty returns", func(t *testing.T) {
		result := If(false, func() string { return "first" }).
			ElseIf(true, func() string { return "" }).
			Else(func() string { return "else" }).
			Render()
		assert.Equal(t, "", result)
	})
}

// TestIfDirective_ComplexContent tests with complex string content
func TestIfDirective_ComplexContent(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		expected  string
	}{
		{
			name:      "multiline content",
			condition: true,
			expected:  "Line 1\nLine 2\nLine 3",
		},
		{
			name:      "content with special characters",
			condition: true,
			expected:  "Special: !@#$%^&*()",
		},
		{
			name:      "unicode content",
			condition: true,
			expected:  "Hello 世界 🌍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := If(tt.condition, func() string { return tt.expected }).Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIfDirective_ConditionalDirectiveInterface tests interface compliance
func TestIfDirective_ConditionalDirectiveInterface(t *testing.T) {
	t.Run("implements ConditionalDirective interface", func(t *testing.T) {
		var _ ConditionalDirective = If(true, func() string { return "test" })
	})

	t.Run("implements Directive interface", func(t *testing.T) {
		var _ Directive = If(true, func() string { return "test" })
	})
}

// TestIfDirective_PanicRecovery tests panic recovery in branch functions
func TestIfDirective_PanicRecovery(t *testing.T) {
	// Set up a test error reporter to capture panics
	var capturedErrors []error
	var capturedContexts []*observability.ErrorContext

	testReporter := &testErrorReporter{
		onReportError: func(err error, ctx *observability.ErrorContext) {
			capturedErrors = append(capturedErrors, err)
			capturedContexts = append(capturedContexts, ctx)
		},
	}

	// Save original reporter and restore after test
	originalReporter := observability.GetErrorReporter()
	defer observability.SetErrorReporter(originalReporter)

	observability.SetErrorReporter(testReporter)

	t.Run("then branch panic recovered", func(t *testing.T) {
		capturedErrors = nil
		capturedContexts = nil

		result := If(true, func() string {
			panic("then branch panic")
		}).Render()

		// Should return empty string on panic
		assert.Equal(t, "", result)

		// Should report error to observability
		assert.Len(t, capturedErrors, 1)
		assert.Contains(t, capturedErrors[0].Error(), "render function panicked")
		assert.Contains(t, capturedErrors[0].Error(), "then branch panicked")

		// Check error context
		assert.Len(t, capturedContexts, 1)
		ctx := capturedContexts[0]
		assert.Equal(t, "If", ctx.ComponentName)
		assert.Equal(t, "If", ctx.Tags["directive_type"])
		assert.Equal(t, "then", ctx.Tags["branch_name"])
		assert.Equal(t, "render_panic", ctx.Tags["error_type"])
		assert.Equal(t, "then branch panic", ctx.Extra["panic_value"])
		assert.NotNil(t, ctx.StackTrace)
	})

	t.Run("elseif branch panic recovered", func(t *testing.T) {
		capturedErrors = nil
		capturedContexts = nil

		result := If(false, func() string {
			return "then"
		}).ElseIf(true, func() string {
			panic("elseif panic")
		}).Render()

		assert.Equal(t, "", result)
		assert.Len(t, capturedErrors, 1)
		assert.Contains(t, capturedErrors[0].Error(), "elseif[0] branch panicked")

		ctx := capturedContexts[0]
		assert.Equal(t, "elseif[0]", ctx.Tags["branch_name"])
		assert.Equal(t, "elseif panic", ctx.Extra["panic_value"])
	})

	t.Run("else branch panic recovered", func(t *testing.T) {
		capturedErrors = nil
		capturedContexts = nil

		result := If(false, func() string {
			return "then"
		}).Else(func() string {
			panic("else panic")
		}).Render()

		assert.Equal(t, "", result)
		assert.Len(t, capturedErrors, 1)
		assert.Contains(t, capturedErrors[0].Error(), "else branch panicked")

		ctx := capturedContexts[0]
		assert.Equal(t, "else", ctx.Tags["branch_name"])
		assert.Equal(t, "else panic", ctx.Extra["panic_value"])
	})

	t.Run("nil pointer panic recovered", func(t *testing.T) {
		capturedErrors = nil
		capturedContexts = nil

		result := If(true, func() string {
			var ptr *string
			return *ptr // nil pointer dereference
		}).Render()

		assert.Equal(t, "", result)
		assert.Len(t, capturedErrors, 1)
	})

	t.Run("no panic when no reporter configured", func(t *testing.T) {
		// Temporarily disable reporter
		observability.SetErrorReporter(nil)
		defer observability.SetErrorReporter(testReporter)

		// Should not panic even without reporter
		result := If(true, func() string {
			panic("panic without reporter")
		}).Render()

		assert.Equal(t, "", result)
	})

	t.Run("multiple elseif panics - only first executes", func(t *testing.T) {
		capturedErrors = nil
		capturedContexts = nil

		result := If(false, func() string {
			return "then"
		}).ElseIf(false, func() string {
			panic("first elseif - should not execute")
		}).ElseIf(true, func() string {
			panic("second elseif - should execute")
		}).Render()

		assert.Equal(t, "", result)
		assert.Len(t, capturedErrors, 1)
		assert.Contains(t, capturedErrors[0].Error(), "elseif[1] branch panicked")
	})
}

// testErrorReporter is a mock error reporter for testing
type testErrorReporter struct {
	onReportError func(err error, ctx *observability.ErrorContext)
	onReportPanic func(err *observability.HandlerPanicError, ctx *observability.ErrorContext)
}

func (r *testErrorReporter) ReportError(err error, ctx *observability.ErrorContext) {
	if r.onReportError != nil {
		r.onReportError(err, ctx)
	}
}

func (r *testErrorReporter) ReportPanic(err *observability.HandlerPanicError, ctx *observability.ErrorContext) {
	if r.onReportPanic != nil {
		r.onReportPanic(err, ctx)
	}
}

func (r *testErrorReporter) Flush(timeout time.Duration) error {
	return nil
}

// ==================== BENCHMARKS ====================

// BenchmarkIfDirective_SimpleTrue benchmarks simple If with true condition
// Target: < 50ns
func BenchmarkIfDirective_SimpleTrue(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = If(true, func() string {
			return "result"
		}).Render()
	}
}

// BenchmarkIfDirective_SimpleFalse benchmarks simple If with false condition
// Target: < 50ns
func BenchmarkIfDirective_SimpleFalse(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = If(false, func() string {
			return "result"
		}).Render()
	}
}

// BenchmarkIfDirective_IfElse benchmarks If with Else branch
// Target: < 100ns
func BenchmarkIfDirective_IfElse(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = If(true, func() string {
			return "then"
		}).Else(func() string {
			return "else"
		}).Render()
	}
}

// BenchmarkIfDirective_ElseIfChain benchmarks If with ElseIf chain
// Target: < 200ns
func BenchmarkIfDirective_ElseIfChain(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = If(false, func() string {
			return "first"
		}).ElseIf(false, func() string {
			return "second"
		}).ElseIf(true, func() string {
			return "third"
		}).Else(func() string {
			return "else"
		}).Render()
	}
}

// BenchmarkIfDirective_Nested benchmarks nested If directives
// Target: < 300ns
func BenchmarkIfDirective_Nested(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = If(true, func() string {
			return If(true, func() string {
				return "nested"
			}).Render()
		}).Render()
	}
}

// BenchmarkIfDirective_ComplexContent benchmarks If with complex string content
// Target: < 100ns
func BenchmarkIfDirective_ComplexContent(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	content := "Line 1\nLine 2\nLine 3\nSpecial: !@#$%^&*()\nUnicode: 世界 🌍"
	for i := 0; i < b.N; i++ {
		_ = If(true, func() string {
			return content
		}).Render()
	}
}
