package composables

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// testLocalStorageErrorReporter is a simple error reporter for testing
type testLocalStorageErrorReporter struct {
	onError func(error, *observability.ErrorContext)
}

func (r *testLocalStorageErrorReporter) ReportError(err error, ctx *observability.ErrorContext) {
	if r.onError != nil {
		r.onError(err, ctx)
	}
}

func (r *testLocalStorageErrorReporter) ReportPanic(panicErr *observability.HandlerPanicError, ctx *observability.ErrorContext) {
	// Not needed for these tests
}

func (r *testLocalStorageErrorReporter) Flush(timeout time.Duration) error {
	return nil
}

// TestReportStorageError_WithReporter tests error reporting with a configured reporter
func TestReportStorageError_WithReporter(t *testing.T) {
	// Arrange
	reportedError := false
	reporter := &testLocalStorageErrorReporter{
		onError: func(err error, ctx *observability.ErrorContext) {
			reportedError = true
			assert.Equal(t, "UseLocalStorage", ctx.ComponentName)
			assert.Equal(t, "Load", ctx.EventName)
			assert.Equal(t, "UseLocalStorage", ctx.Tags["component"])
			assert.Equal(t, "Load", ctx.Tags["operation"])
			assert.Equal(t, "test error", ctx.Extra["error_message"])
		},
	}
	observability.SetErrorReporter(reporter)
	defer observability.SetErrorReporter(nil)

	// Act
	err := errors.New("test error")
	tags := map[string]string{"key": "value"}
	extra := map[string]interface{}{"data": "test"}
	reportStorageError("Load", err, tags, extra)

	// Assert
	assert.True(t, reportedError, "Error should have been reported")
}

// TestReportStorageError_WithoutReporter tests that function handles nil reporter gracefully
func TestReportStorageError_WithoutReporter(t *testing.T) {
	// Arrange
	observability.SetErrorReporter(nil)

	// Act & Assert - should not panic
	err := errors.New("test error")
	tags := map[string]string{}
	extra := map[string]interface{}{}
	reportStorageError("Save", err, tags, extra)
}

// TestTruncateData_ShortData tests truncation with data shorter than max length
func TestTruncateData_ShortData(t *testing.T) {
	// Arrange
	data := []byte("short")

	// Act
	result := truncateData(data, 10)

	// Assert
	assert.Equal(t, "short", result, "Short data should not be truncated")
}

// TestTruncateData_LongData tests truncation with data longer than max length
func TestTruncateData_LongData(t *testing.T) {
	// Arrange
	data := []byte("this is a very long string that exceeds the maximum length")

	// Act
	result := truncateData(data, 10)

	// Assert
	assert.Equal(t, "this is a ...", result, "Long data should be truncated with ellipsis")
	assert.Equal(t, 13, len(result), "Truncated string should be maxLen + 3 (for ...)")
}

// TestTruncateData_ExactLength tests truncation with data exactly at max length
func TestTruncateData_ExactLength(t *testing.T) {
	// Arrange
	data := []byte("exactly10!")

	// Act
	result := truncateData(data, 10)

	// Assert
	assert.Equal(t, "exactly10!", result, "Data at exact length should not be truncated")
}

// TestTruncateData_EmptyData tests truncation with empty data
func TestTruncateData_EmptyData(t *testing.T) {
	// Arrange
	data := []byte("")

	// Act
	result := truncateData(data, 10)

	// Assert
	assert.Equal(t, "", result, "Empty data should return empty string")
}

// TestGetTypeName_PrimitiveTypes tests type name extraction for primitive types
func TestGetTypeName_PrimitiveTypes(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"int", 42, "int"},
		{"string", "test", "string"},
		{"bool", true, "bool"},
		{"float64", 3.14, "float64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := getTypeName(tt.value)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetTypeName_ComplexTypes tests type name extraction for complex types
func TestGetTypeName_ComplexTypes(t *testing.T) {
	type TestStruct struct {
		Field string
	}

	tests := []struct {
		name  string
		value interface{}
	}{
		{"struct", TestStruct{Field: "test"}},
		{"pointer", &TestStruct{Field: "test"}},
		{"slice", []int{1, 2, 3}},
		{"map", map[string]int{"key": 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := getTypeName(tt.value)

			// Assert
			assert.NotEmpty(t, result, "Type name should not be empty")
		})
	}
}

// TestGetTypeName_NilValue tests type name extraction for nil value
func TestGetTypeName_NilValue(t *testing.T) {
	// Act
	result := getTypeName(nil)

	// Assert
	assert.Equal(t, "<nil>", result)
}
