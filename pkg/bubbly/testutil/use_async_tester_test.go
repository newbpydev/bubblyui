package testutil

import (
	"errors"
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
	"github.com/stretchr/testify/assert"
)

// TestUseAsyncTester_BasicAsync tests basic async operation
func TestUseAsyncTester_BasicAsync(t *testing.T) {
	tests := []struct {
		name          string
		fetchResult   *string
		fetchError    error
		expectedData  *string
		expectedError error
	}{
		{
			name:          "successful fetch",
			fetchResult:   stringPtr("test data"),
			fetchError:    nil,
			expectedData:  stringPtr("test data"),
			expectedError: nil,
		},
		{
			name:          "fetch with error",
			fetchResult:   nil,
			fetchError:    errors.New("fetch failed"),
			expectedData:  nil,
			expectedError: errors.New("fetch failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component with UseAsync
			comp, err := bubbly.NewComponent("TestAsync").
				Setup(func(ctx *bubbly.Context) {
					async := composables.UseAsync(ctx, func() (*string, error) {
						return tt.fetchResult, tt.fetchError
					})

					ctx.Expose("data", async.Data)
					ctx.Expose("loading", async.Loading)
					ctx.Expose("error", async.Error)
					ctx.Expose("execute", async.Execute)
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()
			assert.NoError(t, err)
			comp.Init()

			tester := NewUseAsyncTester(comp)

			// Initially not loading, no data, no error
			assert.False(t, tester.IsLoading())
			assert.Nil(t, tester.GetData())
			assert.Nil(t, tester.GetError())

			// Trigger async operation
			tester.TriggerAsync()

			// Wait for completion
			tester.WaitForCompletion(t, 100*time.Millisecond)

			// Verify results
			assert.False(t, tester.IsLoading())

			if tt.expectedData != nil {
				assert.NotNil(t, tester.GetData())
				assert.Equal(t, *tt.expectedData, *tester.GetData().(*string))
			} else {
				assert.Nil(t, tester.GetData())
			}

			if tt.expectedError != nil {
				assert.NotNil(t, tester.GetError())
				assert.Equal(t, tt.expectedError.Error(), tester.GetError().Error())
			} else {
				assert.Nil(t, tester.GetError())
			}
		})
	}
}

// TestUseAsyncTester_LoadingState tests loading state tracking
func TestUseAsyncTester_LoadingState(t *testing.T) {
	comp, err := bubbly.NewComponent("TestAsync").
		Setup(func(ctx *bubbly.Context) {
			async := composables.UseAsync(ctx, func() (*string, error) {
				time.Sleep(50 * time.Millisecond)
				return stringPtr("data"), nil
			})

			ctx.Expose("data", async.Data)
			ctx.Expose("loading", async.Loading)
			ctx.Expose("error", async.Error)
			ctx.Expose("execute", async.Execute)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseAsyncTester(comp)

	// Initially not loading
	assert.False(t, tester.IsLoading())

	// Trigger async
	tester.TriggerAsync()

	// Should be loading immediately after trigger
	// (small delay to let goroutine start)
	time.Sleep(10 * time.Millisecond)
	assert.True(t, tester.IsLoading())

	// Wait for completion
	tester.WaitForCompletion(t, 100*time.Millisecond)

	// Should not be loading after completion
	assert.False(t, tester.IsLoading())
}

// TestUseAsyncTester_MultipleExecutions tests multiple async executions
func TestUseAsyncTester_MultipleExecutions(t *testing.T) {
	callCount := 0

	comp, err := bubbly.NewComponent("TestAsync").
		Setup(func(ctx *bubbly.Context) {
			async := composables.UseAsync(ctx, func() (*int, error) {
				callCount++
				return &callCount, nil
			})

			ctx.Expose("data", async.Data)
			ctx.Expose("loading", async.Loading)
			ctx.Expose("error", async.Error)
			ctx.Expose("execute", async.Execute)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseAsyncTester(comp)

	// First execution
	tester.TriggerAsync()
	tester.WaitForCompletion(t, 100*time.Millisecond)
	assert.Equal(t, 1, *tester.GetData().(*int))

	// Second execution
	tester.TriggerAsync()
	tester.WaitForCompletion(t, 100*time.Millisecond)
	assert.Equal(t, 2, *tester.GetData().(*int))

	// Third execution
	tester.TriggerAsync()
	tester.WaitForCompletion(t, 100*time.Millisecond)
	assert.Equal(t, 3, *tester.GetData().(*int))
}

// TestUseAsyncTester_ErrorClearing tests that errors are cleared on new execution
func TestUseAsyncTester_ErrorClearing(t *testing.T) {
	shouldFail := true

	comp, err := bubbly.NewComponent("TestAsync").
		Setup(func(ctx *bubbly.Context) {
			async := composables.UseAsync(ctx, func() (*string, error) {
				if shouldFail {
					return nil, errors.New("error")
				}
				return stringPtr("success"), nil
			})

			ctx.Expose("data", async.Data)
			ctx.Expose("loading", async.Loading)
			ctx.Expose("error", async.Error)
			ctx.Expose("execute", async.Execute)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseAsyncTester(comp)

	// First execution fails
	tester.TriggerAsync()
	tester.WaitForCompletion(t, 100*time.Millisecond)
	assert.NotNil(t, tester.GetError())
	assert.Nil(t, tester.GetData())

	// Second execution succeeds
	shouldFail = false
	tester.TriggerAsync()
	tester.WaitForCompletion(t, 100*time.Millisecond)
	assert.Nil(t, tester.GetError())
	assert.NotNil(t, tester.GetData())
	assert.Equal(t, "success", *tester.GetData().(*string))
}

// TestUseAsyncTester_MissingRefs tests panic when required refs not exposed
func TestUseAsyncTester_MissingRefs(t *testing.T) {
	comp, err := bubbly.NewComponent("TestAsync").
		Setup(func(ctx *bubbly.Context) {
			// Don't expose required refs
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	assert.Panics(t, func() {
		NewUseAsyncTester(comp)
	})
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
