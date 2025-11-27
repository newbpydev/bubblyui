package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewTestSetup tests that NewTestSetup creates a properly initialized instance.
func TestNewTestSetup(t *testing.T) {
	setup := NewTestSetup()

	assert.NotNil(t, setup)
	assert.NotNil(t, setup.setupFuncs)
	assert.NotNil(t, setup.teardownFuncs)
	assert.Equal(t, 0, len(setup.setupFuncs))
	assert.Equal(t, 0, len(setup.teardownFuncs))
}

// TestTestSetup_AddSetup tests that AddSetup adds functions and returns self for chaining.
func TestTestSetup_AddSetup(t *testing.T) {
	setup := NewTestSetup()

	// Add first setup
	result1 := setup.AddSetup(func(t *testing.T) {})
	assert.Same(t, setup, result1, "AddSetup should return self for chaining")
	assert.Equal(t, 1, len(setup.setupFuncs))

	// Add second setup
	result2 := setup.AddSetup(func(t *testing.T) {})
	assert.Same(t, setup, result2, "AddSetup should return self for chaining")
	assert.Equal(t, 2, len(setup.setupFuncs))
}

// TestTestSetup_AddTeardown tests that AddTeardown adds functions and returns self for chaining.
func TestTestSetup_AddTeardown(t *testing.T) {
	setup := NewTestSetup()

	// Add first teardown
	result1 := setup.AddTeardown(func(t *testing.T) {})
	assert.Same(t, setup, result1, "AddTeardown should return self for chaining")
	assert.Equal(t, 1, len(setup.teardownFuncs))

	// Add second teardown
	result2 := setup.AddTeardown(func(t *testing.T) {})
	assert.Same(t, setup, result2, "AddTeardown should return self for chaining")
	assert.Equal(t, 2, len(setup.teardownFuncs))
}

// TestTestSetup_FluentAPI tests that the fluent API works correctly.
func TestTestSetup_FluentAPI(t *testing.T) {
	setup := NewTestSetup().
		AddSetup(func(t *testing.T) {}).
		AddSetup(func(t *testing.T) {}).
		AddTeardown(func(t *testing.T) {}).
		AddTeardown(func(t *testing.T) {})

	assert.Equal(t, 2, len(setup.setupFuncs))
	assert.Equal(t, 2, len(setup.teardownFuncs))
}

// TestTestSetup_Run_SetupExecutionOrder tests that setup functions execute in FIFO order.
func TestTestSetup_Run_SetupExecutionOrder(t *testing.T) {
	var executionOrder []int

	setup := NewTestSetup().
		AddSetup(func(t *testing.T) {
			executionOrder = append(executionOrder, 1)
		}).
		AddSetup(func(t *testing.T) {
			executionOrder = append(executionOrder, 2)
		}).
		AddSetup(func(t *testing.T) {
			executionOrder = append(executionOrder, 3)
		})

	setup.Run(t, func(t *testing.T) {
		// Setup should have executed before test
		assert.Equal(t, []int{1, 2, 3}, executionOrder, "Setup functions should execute in FIFO order")
	})
}

// TestTestSetup_Run_TeardownExecutionOrder tests that teardown functions execute in LIFO order.
func TestTestSetup_Run_TeardownExecutionOrder(t *testing.T) {
	var executionOrder []int

	// Use a subtest so we can verify teardown after it completes
	t.Run("subtest", func(t *testing.T) {
		setup := NewTestSetup().
			AddTeardown(func(t *testing.T) {
				executionOrder = append(executionOrder, 1)
			}).
			AddTeardown(func(t *testing.T) {
				executionOrder = append(executionOrder, 2)
			}).
			AddTeardown(func(t *testing.T) {
				executionOrder = append(executionOrder, 3)
			})

		setup.Run(t, func(t *testing.T) {
			// Test executes, teardown happens after via t.Cleanup
		})
	})

	// After subtest completes, teardown should have executed in reverse order (LIFO)
	assert.Equal(t, []int{3, 2, 1}, executionOrder, "Teardown functions should execute in LIFO order")
}

// TestTestSetup_Run_CompleteExecutionOrder tests the full execution order.
func TestTestSetup_Run_CompleteExecutionOrder(t *testing.T) {
	var executionOrder []string

	// Use a subtest so teardown completes before we verify
	t.Run("subtest", func(t *testing.T) {
		setup := NewTestSetup().
			AddSetup(func(t *testing.T) {
				executionOrder = append(executionOrder, "setup1")
			}).
			AddSetup(func(t *testing.T) {
				executionOrder = append(executionOrder, "setup2")
			}).
			AddTeardown(func(t *testing.T) {
				executionOrder = append(executionOrder, "teardown1")
			}).
			AddTeardown(func(t *testing.T) {
				executionOrder = append(executionOrder, "teardown2")
			})

		setup.Run(t, func(t *testing.T) {
			executionOrder = append(executionOrder, "test")
		})
	})

	// Expected order: setup1, setup2, test, teardown2, teardown1
	expected := []string{"setup1", "setup2", "test", "teardown2", "teardown1"}
	assert.Equal(t, expected, executionOrder, "Execution order should be: setup (FIFO), test, teardown (LIFO)")
}

// TestTestSetup_Run_TestFunctionExecutes tests that the test function executes.
func TestTestSetup_Run_TestFunctionExecutes(t *testing.T) {
	testExecuted := false

	setup := NewTestSetup()
	setup.Run(t, func(t *testing.T) {
		testExecuted = true
	})

	assert.True(t, testExecuted, "Test function should execute")
}

// TestTestSetup_Run_NoSetupOrTeardown tests that Run works with no setup/teardown.
func TestTestSetup_Run_NoSetupOrTeardown(t *testing.T) {
	testExecuted := false

	setup := NewTestSetup()
	setup.Run(t, func(t *testing.T) {
		testExecuted = true
	})

	assert.True(t, testExecuted, "Test should execute even without setup/teardown")
}

// TestTestSetup_Run_OnlySetup tests that Run works with only setup functions.
func TestTestSetup_Run_OnlySetup(t *testing.T) {
	setupExecuted := false
	testExecuted := false

	setup := NewTestSetup().
		AddSetup(func(t *testing.T) {
			setupExecuted = true
		})

	setup.Run(t, func(t *testing.T) {
		assert.True(t, setupExecuted, "Setup should execute before test")
		testExecuted = true
	})

	assert.True(t, testExecuted, "Test should execute")
}

// TestTestSetup_Run_OnlyTeardown tests that Run works with only teardown functions.
func TestTestSetup_Run_OnlyTeardown(t *testing.T) {
	teardownExecuted := false

	t.Run("subtest", func(t *testing.T) {
		setup := NewTestSetup().
			AddTeardown(func(t *testing.T) {
				teardownExecuted = true
			})

		setup.Run(t, func(t *testing.T) {
			// Test executes
		})
	})

	assert.True(t, teardownExecuted, "Teardown should execute after test")
}

// TestTestSetup_Run_MultipleSetupAndTeardown tests multiple setup and teardown functions.
func TestTestSetup_Run_MultipleSetupAndTeardown(t *testing.T) {
	var events []string

	t.Run("subtest", func(t *testing.T) {
		setup := NewTestSetup().
			AddSetup(func(t *testing.T) {
				events = append(events, "setup1")
			}).
			AddSetup(func(t *testing.T) {
				events = append(events, "setup2")
			}).
			AddSetup(func(t *testing.T) {
				events = append(events, "setup3")
			}).
			AddTeardown(func(t *testing.T) {
				events = append(events, "teardown1")
			}).
			AddTeardown(func(t *testing.T) {
				events = append(events, "teardown2")
			}).
			AddTeardown(func(t *testing.T) {
				events = append(events, "teardown3")
			})

		setup.Run(t, func(t *testing.T) {
			events = append(events, "test")
		})
	})

	expected := []string{
		"setup1", "setup2", "setup3", // FIFO
		"test",
		"teardown3", "teardown2", "teardown1", // LIFO
	}
	assert.Equal(t, expected, events)
}

// TestTestSetup_Run_TeardownExecutesOnPanic tests that teardown executes even if test panics.
func TestTestSetup_Run_TeardownExecutesOnPanic(t *testing.T) {
	teardownExecuted := false

	setup := NewTestSetup().
		AddTeardown(func(t *testing.T) {
			teardownExecuted = true
		})

	// Use a subtest to contain the panic
	t.Run("subtest", func(t *testing.T) {
		defer func() {
			// Recover from panic - expected in this test
			// teardown will execute via t.Cleanup
			_ = recover()
		}()

		setup.Run(t, func(t *testing.T) {
			panic("test panic")
		})
	})

	// Note: In real Go tests, t.Cleanup runs even on panic
	// The implementation uses t.Cleanup which guarantees execution
	// We verify that the variable was set (though in a real scenario,
	// t.Cleanup would have executed it)
	assert.True(t, teardownExecuted, "Teardown should execute even on panic")
}

// TestTestSetup_Run_StateModification tests that setup can modify state for test.
func TestTestSetup_Run_StateModification(t *testing.T) {
	var counter int

	setup := NewTestSetup().
		AddSetup(func(t *testing.T) {
			counter = 10
		}).
		AddSetup(func(t *testing.T) {
			counter += 5
		})

	setup.Run(t, func(t *testing.T) {
		assert.Equal(t, 15, counter, "Setup should modify state before test")
		counter += 3
	})

	assert.Equal(t, 18, counter, "Test should have modified state")
}

// TestTestSetup_Run_TeardownRestoresState tests that teardown can restore state.
func TestTestSetup_Run_TeardownRestoresState(t *testing.T) {
	originalValue := 42
	testValue := originalValue

	t.Run("subtest", func(t *testing.T) {
		setup := NewTestSetup().
			AddSetup(func(t *testing.T) {
				testValue = 100
			}).
			AddTeardown(func(t *testing.T) {
				testValue = originalValue
			})

		setup.Run(t, func(t *testing.T) {
			assert.Equal(t, 100, testValue, "Setup should change value")
			testValue = 200
		})
	})

	assert.Equal(t, originalValue, testValue, "Teardown should restore original value")
}

// TestTestSetup_Run_IntegrationWithTCleanup tests integration with t.Cleanup.
func TestTestSetup_Run_IntegrationWithTCleanup(t *testing.T) {
	var events []string

	t.Run("subtest", func(t *testing.T) {
		setup := NewTestSetup().
			AddTeardown(func(t *testing.T) {
				events = append(events, "testsetup_teardown")
			})

		setup.Run(t, func(t *testing.T) {
			// Add additional cleanup via t.Cleanup
			t.Cleanup(func() {
				events = append(events, "t_cleanup")
			})
			events = append(events, "test")
		})
	})

	// t.Cleanup executes in LIFO order, so t.Cleanup runs before TestSetup teardown
	expected := []string{"test", "t_cleanup", "testsetup_teardown"}
	assert.Equal(t, expected, events, "t.Cleanup should execute before TestSetup teardown")
}

// TestTestSetup_Run_RealWorldExample tests a realistic use case.
func TestTestSetup_Run_RealWorldExample(t *testing.T) {
	var (
		dbConnected    bool
		cacheConnected bool
		loggerSetup    bool
	)

	t.Run("subtest", func(t *testing.T) {
		setup := NewTestSetup().
			AddSetup(func(t *testing.T) {
				// Simulate database connection
				dbConnected = true
				t.Log("Database connected")
			}).
			AddSetup(func(t *testing.T) {
				// Simulate cache connection
				cacheConnected = true
				t.Log("Cache connected")
			}).
			AddSetup(func(t *testing.T) {
				// Simulate logger setup
				loggerSetup = true
				t.Log("Logger configured")
			}).
			AddTeardown(func(t *testing.T) {
				// Cleanup in reverse order
				loggerSetup = false
				t.Log("Logger shutdown")
			}).
			AddTeardown(func(t *testing.T) {
				cacheConnected = false
				t.Log("Cache disconnected")
			}).
			AddTeardown(func(t *testing.T) {
				dbConnected = false
				t.Log("Database disconnected")
			})

		setup.Run(t, func(t *testing.T) {
			// All resources should be available
			assert.True(t, dbConnected, "Database should be connected")
			assert.True(t, cacheConnected, "Cache should be connected")
			assert.True(t, loggerSetup, "Logger should be configured")

			// Run actual test
			t.Log("Running test with all resources")
		})
	})

	// All resources should be cleaned up
	assert.False(t, dbConnected, "Database should be disconnected")
	assert.False(t, cacheConnected, "Cache should be disconnected")
	assert.False(t, loggerSetup, "Logger should be shutdown")
}

// TestTestSetup_Run_EmptyTestFunction tests that empty test function works.
func TestTestSetup_Run_EmptyTestFunction(t *testing.T) {
	setupExecuted := false
	teardownExecuted := false

	t.Run("subtest", func(t *testing.T) {
		setup := NewTestSetup().
			AddSetup(func(t *testing.T) {
				setupExecuted = true
			}).
			AddTeardown(func(t *testing.T) {
				teardownExecuted = true
			})

		// Empty test function
		setup.Run(t, func(t *testing.T) {})
	})

	assert.True(t, setupExecuted, "Setup should execute")
	assert.True(t, teardownExecuted, "Teardown should execute")
}

// TestTestSetup_Run_NestedSetup tests that TestSetup can be nested.
func TestTestSetup_Run_NestedSetup(t *testing.T) {
	var events []string

	t.Run("subtest", func(t *testing.T) {
		outerSetup := NewTestSetup().
			AddSetup(func(t *testing.T) {
				events = append(events, "outer_setup")
			}).
			AddTeardown(func(t *testing.T) {
				events = append(events, "outer_teardown")
			})

		outerSetup.Run(t, func(t *testing.T) {
			events = append(events, "outer_test_start")

			innerSetup := NewTestSetup().
				AddSetup(func(t *testing.T) {
					events = append(events, "inner_setup")
				}).
				AddTeardown(func(t *testing.T) {
					events = append(events, "inner_teardown")
				})

			innerSetup.Run(t, func(t *testing.T) {
				events = append(events, "inner_test")
			})

			events = append(events, "outer_test_end")
		})
	})

	// Note: inner_teardown executes after outer_test_end because it's registered
	// with the outer test's t.Cleanup, which runs after the test function completes
	expected := []string{
		"outer_setup",
		"outer_test_start",
		"inner_setup",
		"inner_test",
		"outer_test_end",
		"inner_teardown", // Executes after outer test function completes
		"outer_teardown",
	}
	assert.Equal(t, expected, events)
}
