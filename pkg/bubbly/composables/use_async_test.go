package composables

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
)

// TestUseAsync_ExecuteTriggersFetch verifies that calling Execute triggers the fetcher function
func TestUseAsync_ExecuteTriggersFetch(t *testing.T) {
	// Arrange
	ctx := &bubbly.Context{}
	var mu sync.Mutex
	fetchCalled := false
	fetcher := func() (*string, error) {
		mu.Lock()
		fetchCalled = true
		mu.Unlock()
		result := "data"
		return &result, nil
	}

	// Act
	async := UseAsync(ctx, fetcher)
	async.Execute()

	// Wait for async operation
	time.Sleep(50 * time.Millisecond)

	// Assert
	mu.Lock()
	wasCalled := fetchCalled
	mu.Unlock()
	assert.True(t, wasCalled, "Fetcher should be called when Execute is invoked")
}

// TestUseAsync_LoadingStateManaged verifies loading state transitions correctly
func TestUseAsync_LoadingStateManaged(t *testing.T) {
	// Arrange
	ctx := &bubbly.Context{}
	fetcher := func() (*int, error) {
		time.Sleep(20 * time.Millisecond)
		result := 42
		return &result, nil
	}

	// Act
	async := UseAsync(ctx, fetcher)

	// Assert initial state
	assert.False(t, async.Loading.GetTyped(), "Loading should be false initially")

	// Execute
	async.Execute()

	// Assert loading is true during fetch
	assert.True(t, async.Loading.GetTyped(), "Loading should be true immediately after Execute")

	// Wait for completion
	time.Sleep(50 * time.Millisecond)

	// Assert loading is false after completion
	assert.False(t, async.Loading.GetTyped(), "Loading should be false after fetch completes")
}

// TestUseAsync_DataPopulatedOnSuccess verifies data is set when fetch succeeds
func TestUseAsync_DataPopulatedOnSuccess(t *testing.T) {
	// Arrange
	ctx := &bubbly.Context{}
	expected := "success data"
	fetcher := func() (*string, error) {
		return &expected, nil
	}

	// Act
	async := UseAsync(ctx, fetcher)
	async.Execute()

	// Wait for async operation
	time.Sleep(50 * time.Millisecond)

	// Assert
	assert.NotNil(t, async.Data.GetTyped(), "Data should not be nil after successful fetch")
	assert.Equal(t, expected, *async.Data.GetTyped(), "Data should match fetched value")
	assert.Nil(t, async.Error.GetTyped(), "Error should be nil on success")
}

// TestUseAsync_ErrorSetOnFailure verifies error is set when fetch fails
func TestUseAsync_ErrorSetOnFailure(t *testing.T) {
	// Arrange
	ctx := &bubbly.Context{}
	expectedError := errors.New("fetch failed")
	fetcher := func() (*string, error) {
		return nil, expectedError
	}

	// Act
	async := UseAsync(ctx, fetcher)
	async.Execute()

	// Wait for async operation
	time.Sleep(50 * time.Millisecond)

	// Assert
	assert.NotNil(t, async.Error.GetTyped(), "Error should be set when fetch fails")
	assert.Equal(t, expectedError, async.Error.GetTyped(), "Error should match the returned error")
	assert.Nil(t, async.Data.GetTyped(), "Data should be nil on error")
	assert.False(t, async.Loading.GetTyped(), "Loading should be false after error")
}

// TestUseAsync_ResetClearsState verifies Reset clears all state
func TestUseAsync_ResetClearsState(t *testing.T) {
	// Arrange
	ctx := &bubbly.Context{}
	data := "test data"
	fetcher := func() (*string, error) {
		return &data, nil
	}

	async := UseAsync(ctx, fetcher)
	async.Execute()
	time.Sleep(50 * time.Millisecond)

	// Verify data is set
	assert.NotNil(t, async.Data.GetTyped(), "Data should be set before reset")

	// Act
	async.Reset()

	// Assert
	assert.Nil(t, async.Data.GetTyped(), "Data should be nil after reset")
	assert.False(t, async.Loading.GetTyped(), "Loading should be false after reset")
	assert.Nil(t, async.Error.GetTyped(), "Error should be nil after reset")
}

// TestUseAsync_ConcurrentExecutions verifies concurrent Execute calls are handled safely
func TestUseAsync_ConcurrentExecutions(t *testing.T) {
	// Arrange
	ctx := &bubbly.Context{}
	var mu sync.Mutex
	callCount := 0

	fetcher := func() (*int, error) {
		mu.Lock()
		callCount++
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		result := 42
		return &result, nil
	}

	async := UseAsync(ctx, fetcher)

	// Act - call Execute multiple times concurrently
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			async.Execute()
		}()
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	// Assert
	mu.Lock()
	actualCallCount := callCount
	mu.Unlock()

	assert.Equal(t, 5, actualCallCount, "All concurrent Execute calls should trigger fetcher")
	assert.NotNil(t, async.Data.GetTyped(), "Data should be set after concurrent executions")
	assert.False(t, async.Loading.GetTyped(), "Loading should be false after all executions complete")
}

// TestUseAsync_TypeSafety verifies type safety with different types
func TestUseAsync_TypeSafety(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "int type",
			testFunc: func(t *testing.T) {
				ctx := &bubbly.Context{}
				expected := 123
				fetcher := func() (*int, error) {
					return &expected, nil
				}

				async := UseAsync(ctx, fetcher)
				async.Execute()
				time.Sleep(50 * time.Millisecond)

				assert.Equal(t, expected, *async.Data.GetTyped())
			},
		},
		{
			name: "string type",
			testFunc: func(t *testing.T) {
				ctx := &bubbly.Context{}
				expected := "hello"
				fetcher := func() (*string, error) {
					return &expected, nil
				}

				async := UseAsync(ctx, fetcher)
				async.Execute()
				time.Sleep(50 * time.Millisecond)

				assert.Equal(t, expected, *async.Data.GetTyped())
			},
		},
		{
			name: "struct type",
			testFunc: func(t *testing.T) {
				type User struct {
					Name string
					Age  int
				}

				ctx := &bubbly.Context{}
				expected := User{Name: "Alice", Age: 30}
				fetcher := func() (*User, error) {
					return &expected, nil
				}

				async := UseAsync(ctx, fetcher)
				async.Execute()
				time.Sleep(50 * time.Millisecond)

				assert.Equal(t, expected, *async.Data.GetTyped())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

// TestUseAsync_InitialState verifies initial state is correct
func TestUseAsync_InitialState(t *testing.T) {
	// Arrange
	ctx := &bubbly.Context{}
	fetcher := func() (*string, error) {
		result := "data"
		return &result, nil
	}

	// Act
	async := UseAsync(ctx, fetcher)

	// Assert
	assert.Nil(t, async.Data.GetTyped(), "Data should be nil initially")
	assert.False(t, async.Loading.GetTyped(), "Loading should be false initially")
	assert.Nil(t, async.Error.GetTyped(), "Error should be nil initially")
	assert.NotNil(t, async.Execute, "Execute function should be provided")
	assert.NotNil(t, async.Reset, "Reset function should be provided")
}

// TestUseAsync_ErrorClearedOnRetry verifies error is cleared when retrying after failure
func TestUseAsync_ErrorClearedOnRetry(t *testing.T) {
	// Arrange
	ctx := &bubbly.Context{}
	shouldFail := true
	fetcher := func() (*string, error) {
		if shouldFail {
			return nil, errors.New("first attempt failed")
		}
		result := "success"
		return &result, nil
	}

	async := UseAsync(ctx, fetcher)

	// First attempt - should fail
	async.Execute()
	time.Sleep(50 * time.Millisecond)
	assert.NotNil(t, async.Error.GetTyped(), "Error should be set after first failed attempt")

	// Second attempt - should succeed
	shouldFail = false
	async.Execute()
	time.Sleep(50 * time.Millisecond)

	// Assert
	assert.Nil(t, async.Error.GetTyped(), "Error should be cleared on successful retry")
	assert.NotNil(t, async.Data.GetTyped(), "Data should be set on successful retry")
}
