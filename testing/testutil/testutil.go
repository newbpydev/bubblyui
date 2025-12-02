// Package testutil provides comprehensive test utilities for BubblyUI applications.
//
// This package contains testers, mocks, assertions, simulators, and snapshot
// testing utilities for thorough BubblyUI component and composable testing.
//
// This package is an alias for github.com/newbpydev/bubblyui/pkg/bubbly/testutil,
// providing a cleaner import path for users.
//
// # Features
//
//   - Component testers for all built-in features
//   - Mock implementations (Ref, Router, Storage, etc.)
//   - Snapshot testing with diff visualization
//   - Time simulation for async testing
//   - Event tracking and inspection
//   - Command queue inspection
//   - Data factories for test fixtures
//
// # Example
//
//	import "github.com/newbpydev/bubblyui/testing/testutil"
//
//	func TestComponent(t *testing.T) {
//	    // Create harness
//	    h := testutil.NewHarness(t)
//
//	    // Create mock ref
//	    ref := testutil.NewMockRef(42)
//
//	    // Snapshot testing
//	    testutil.MatchSnapshot(t, component.View())
//	}
package testutil

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/commands"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
	"github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

// =============================================================================
// Test Harness
// =============================================================================

// TestHarness provides a complete test environment.
type TestHarness = testutil.TestHarness

// NewHarness creates a new test harness.
func NewHarness(t *testing.T, opts ...HarnessOption) *TestHarness {
	return testutil.NewHarness(t, opts...)
}

// HarnessOption configures a test harness.
type HarnessOption = testutil.HarnessOption

// =============================================================================
// Test Setup
// =============================================================================

// TestSetup provides common test setup utilities.
type TestSetup = testutil.TestSetup

// NewTestSetup creates a new test setup.
var NewTestSetup = testutil.NewTestSetup

// TestIsolation ensures test isolation.
type TestIsolation = testutil.TestIsolation

// NewTestIsolation creates new test isolation.
var NewTestIsolation = testutil.NewTestIsolation

// =============================================================================
// Fixtures
// =============================================================================

// FixtureBuilder builds test fixtures.
type FixtureBuilder = testutil.FixtureBuilder

// NewFixture creates a new fixture builder.
var NewFixture = testutil.NewFixture

// =============================================================================
// Mock Ref
// =============================================================================

// MockRef is a mock implementation of Ref for testing.
type MockRef[T any] = testutil.MockRef[T]

// NewMockRef creates a new mock ref.
func NewMockRef[T any](initial T) *MockRef[T] {
	return testutil.NewMockRef(initial)
}

// CreateMockRef creates a mock ref in a factory.
func CreateMockRef[T any](mf *MockFactory, name string, initial T) *MockRef[T] {
	return testutil.CreateMockRef(mf, name, initial)
}

// GetMockRef gets a mock ref from a factory.
func GetMockRef[T any](mf *MockFactory, name string) *MockRef[T] {
	return testutil.GetMockRef[T](mf, name)
}

// MockFactory manages mock refs.
type MockFactory = testutil.MockFactory

// NewMockFactory creates a new mock factory.
var NewMockFactory = testutil.NewMockFactory

// =============================================================================
// Mock Components
// =============================================================================

// MockComponent is a mock component for testing.
type MockComponent = testutil.MockComponent

// NewMockComponent creates a new mock component.
var NewMockComponent = testutil.NewMockComponent

// =============================================================================
// Mock Router
// =============================================================================

// MockRouter is a mock router for testing.
type MockRouter = testutil.MockRouter

// NewMockRouter creates a new mock router.
var NewMockRouter = testutil.NewMockRouter

// =============================================================================
// Mock Storage
// =============================================================================

// MockStorage is a mock storage for testing.
type MockStorage = testutil.MockStorage

// NewMockStorage creates a new mock storage.
var NewMockStorage = testutil.NewMockStorage

// =============================================================================
// Mock Commands
// =============================================================================

// MockCommand is a mock command for testing.
type MockCommand = testutil.MockCommand

// NewMockCommand creates a new mock command.
func NewMockCommand(msg tea.Msg) (*MockCommand, tea.Cmd) {
	return testutil.NewMockCommand(msg)
}

// NewMockCommandWithError creates a mock command that returns an error.
var NewMockCommandWithError = testutil.NewMockCommandWithError

// MockCommandGenerator is a mock command generator.
type MockCommandGenerator = testutil.MockCommandGenerator

// NewMockCommandGenerator creates a new mock command generator.
func NewMockCommandGenerator(returnCmd tea.Cmd) *MockCommandGenerator {
	return testutil.NewMockCommandGenerator(returnCmd)
}

// =============================================================================
// Mock Error Reporter
// =============================================================================

// MockErrorReporter is a mock error reporter for testing.
type MockErrorReporter = testutil.MockErrorReporter

// NewMockErrorReporter creates a new mock error reporter.
var NewMockErrorReporter = testutil.NewMockErrorReporter

// MockErrorMsg represents a mock error message.
type MockErrorMsg = testutil.MockErrorMsg

// =============================================================================
// Snapshot Testing
// =============================================================================

// MatchSnapshot compares output against a saved snapshot.
func MatchSnapshot(t *testing.T, actual string) {
	testutil.MatchSnapshot(t, actual)
}

// MatchNamedSnapshot compares output against a named snapshot.
func MatchNamedSnapshot(t *testing.T, name, actual string) {
	testutil.MatchNamedSnapshot(t, name, actual)
}

// MatchComponentSnapshot compares a component's view against a snapshot.
func MatchComponentSnapshot(t *testing.T, component bubbly.Component) {
	testutil.MatchComponentSnapshot(t, component)
}

// MatchSnapshotWithOptions compares with custom options.
func MatchSnapshotWithOptions(t *testing.T, name, actual, dir string, update bool) {
	testutil.MatchSnapshotWithOptions(t, name, actual, dir, update)
}

// GetSnapshotPath returns the path for a snapshot.
func GetSnapshotPath(t *testing.T, name string) string {
	return testutil.GetSnapshotPath(t, name)
}

// ReadSnapshot reads a snapshot from disk.
func ReadSnapshot(t *testing.T, name string) (string, error) {
	return testutil.ReadSnapshot(t, name)
}

// SnapshotExists checks if a snapshot exists.
func SnapshotExists(t *testing.T, name string) bool {
	return testutil.SnapshotExists(t, name)
}

// UpdateSnapshots returns whether snapshots should be updated.
func UpdateSnapshots(t *testing.T) bool {
	return testutil.UpdateSnapshots(t)
}

// SnapshotManager manages snapshots.
type SnapshotManager = testutil.SnapshotManager

// NewSnapshotManager creates a new snapshot manager.
var NewSnapshotManager = testutil.NewSnapshotManager

// GetSnapshotManager returns the snapshot manager for a test.
func GetSnapshotManager(t *testing.T) *SnapshotManager {
	return testutil.GetSnapshotManager(t)
}

// =============================================================================
// Normalization
// =============================================================================

// NormalizeIDs normalizes generated IDs in output.
var NormalizeIDs = testutil.NormalizeIDs

// NormalizeTimestamps normalizes timestamps in output.
var NormalizeTimestamps = testutil.NormalizeTimestamps

// NormalizeUUIDs normalizes UUIDs in output.
var NormalizeUUIDs = testutil.NormalizeUUIDs

// NormalizeAll applies all normalizations.
var NormalizeAll = testutil.NormalizeAll

// Normalizer applies custom normalizations.
type Normalizer = testutil.Normalizer

// NewNormalizer creates a new normalizer.
var NewNormalizer = testutil.NewNormalizer

// NormalizePattern defines a normalization pattern.
type NormalizePattern = testutil.NormalizePattern

// =============================================================================
// Event Tracking
// =============================================================================

// EventTracker tracks events.
type EventTracker = testutil.EventTracker

// NewEventTracker creates a new event tracker.
var NewEventTracker = testutil.NewEventTracker

// EventInspector inspects tracked events.
type EventInspector = testutil.EventInspector

// NewEventInspector creates a new event inspector.
var NewEventInspector = testutil.NewEventInspector

// Event represents a tracked event.
type Event = testutil.Event

// EmittedEvent represents an emitted event.
type EmittedEvent = testutil.EmittedEvent

// =============================================================================
// Time Simulation
// =============================================================================

// TimeSimulator simulates time for testing async behavior.
type TimeSimulator = testutil.TimeSimulator

// NewTimeSimulator creates a new time simulator.
var NewTimeSimulator = testutil.NewTimeSimulator

// SimulatedTimer is a simulated timer.
type SimulatedTimer = testutil.SimulatedTimer

// =============================================================================
// Wait Utilities
// =============================================================================

// WaitFor waits for a condition to be true.
func WaitFor(t *testing.T, condition func() bool, opts WaitOptions) {
	testutil.WaitFor(t, condition, opts)
}

// WaitOptions configures WaitFor behavior.
type WaitOptions = testutil.WaitOptions

// =============================================================================
// Command Testing
// =============================================================================

// CommandQueueInspector inspects command queues.
type CommandQueueInspector = testutil.CommandQueueInspector

// NewCommandQueueInspector creates a new command queue inspector.
func NewCommandQueueInspector(queue *bubbly.CommandQueue) *CommandQueueInspector {
	return testutil.NewCommandQueueInspector(queue)
}

// AssertCommandEnqueued asserts that commands were enqueued.
func AssertCommandEnqueued(t *testing.T, queue *CommandQueueInspector, count int) {
	testutil.AssertCommandEnqueued(t, queue, count)
}

// LoopDetectionVerifier verifies loop detection.
type LoopDetectionVerifier = testutil.LoopDetectionVerifier

// NewLoopDetectionVerifier creates a new loop detection verifier.
func NewLoopDetectionVerifier(detector *commands.LoopDetector) *LoopDetectionVerifier {
	return testutil.NewLoopDetectionVerifier(detector)
}

// AssertNoCommandLoop asserts no command loop occurred.
func AssertNoCommandLoop(t *testing.T, detector *LoopDetectionVerifier) {
	testutil.AssertNoCommandLoop(t, detector)
}

// LoopEvent represents a detected loop event.
type LoopEvent = testutil.LoopEvent

// =============================================================================
// Auto Command Tester
// =============================================================================

// AutoCommandTester tests automatic command generation.
type AutoCommandTester = testutil.AutoCommandTester

// NewAutoCommandTester creates a new auto command tester.
var NewAutoCommandTester = testutil.NewAutoCommandTester

// =============================================================================
// Batcher Tester
// =============================================================================

// BatcherTester tests command batching.
type BatcherTester = testutil.BatcherTester

// NewBatcherTester creates a new batcher tester.
func NewBatcherTester(batcher *commands.CommandBatcher) *BatcherTester {
	return testutil.NewBatcherTester(batcher)
}

// =============================================================================
// Component Testers
// =============================================================================

// ComponentTest is a base for component tests.
type ComponentTest = testutil.ComponentTest

// ChildrenManagementTester tests children management.
type ChildrenManagementTester = testutil.ChildrenManagementTester

// NewChildrenManagementTester creates a new children management tester.
var NewChildrenManagementTester = testutil.NewChildrenManagementTester

// PropsVerifier verifies component props.
type PropsVerifier = testutil.PropsVerifier

// NewPropsVerifier creates a new props verifier.
var NewPropsVerifier = testutil.NewPropsVerifier

// PropsMutation represents a props mutation.
type PropsMutation = testutil.PropsMutation

// KeyBindingsTester tests key bindings.
type KeyBindingsTester = testutil.KeyBindingsTester

// NewKeyBindingsTester creates a new key bindings tester.
var NewKeyBindingsTester = testutil.NewKeyBindingsTester

// MessageHandlerTester tests message handlers.
type MessageHandlerTester = testutil.MessageHandlerTester

// NewMessageHandlerTester creates a new message handler tester.
var NewMessageHandlerTester = testutil.NewMessageHandlerTester

// =============================================================================
// State Inspection
// =============================================================================

// StateInspector inspects component state.
type StateInspector = testutil.StateInspector

// NewStateInspector creates a new state inspector.
var NewStateInspector = testutil.NewStateInspector

// =============================================================================
// Directive Testers
// =============================================================================

// IfTester tests v-if directive behavior.
type IfTester = testutil.IfTester

// NewIfTester creates a new if tester.
var NewIfTester = testutil.NewIfTester

// ForEachTester tests v-for directive behavior.
type ForEachTester = testutil.ForEachTester

// NewForEachTester creates a new for-each tester.
var NewForEachTester = testutil.NewForEachTester

// ShowTester tests v-show directive behavior.
type ShowTester = testutil.ShowTester

// NewShowTester creates a new show tester.
var NewShowTester = testutil.NewShowTester

// BindTester tests v-bind directive behavior.
type BindTester = testutil.BindTester

// NewBindTester creates a new bind tester.
var NewBindTester = testutil.NewBindTester

// BoolRefTester tests boolean ref behavior.
type BoolRefTester = testutil.BoolRefTester

// NewBoolRefTester creates a new bool ref tester.
var NewBoolRefTester = testutil.NewBoolRefTester

// OnTester tests v-on directive behavior.
type OnTester = testutil.OnTester

// NewOnTester creates a new on tester.
var NewOnTester = testutil.NewOnTester

// =============================================================================
// Composable Testers
// =============================================================================

// UseStateTester tests useState behavior.
type UseStateTester[T any] = testutil.UseStateTester[T]

// NewUseStateTester creates a new use state tester.
func NewUseStateTester[T any](comp bubbly.Component) *UseStateTester[T] {
	return testutil.NewUseStateTester[T](comp)
}

// UseEffectTester tests useEffect behavior.
type UseEffectTester = testutil.UseEffectTester

// NewUseEffectTester creates a new use effect tester.
var NewUseEffectTester = testutil.NewUseEffectTester

// UseDebounceTester tests useDebounce behavior.
type UseDebounceTester = testutil.UseDebounceTester

// NewUseDebounceTester creates a new use debounce tester.
var NewUseDebounceTester = testutil.NewUseDebounceTester

// UseThrottleTester tests useThrottle behavior.
type UseThrottleTester = testutil.UseThrottleTester

// NewUseThrottleTester creates a new use throttle tester.
var NewUseThrottleTester = testutil.NewUseThrottleTester

// UseAsyncTester tests useAsync behavior.
type UseAsyncTester = testutil.UseAsyncTester

// NewUseAsyncTester creates a new use async tester.
var NewUseAsyncTester = testutil.NewUseAsyncTester

// UseEventListenerTester tests useEventListener behavior.
type UseEventListenerTester = testutil.UseEventListenerTester

// NewUseEventListenerTester creates a new use event listener tester.
var NewUseEventListenerTester = testutil.NewUseEventListenerTester

// UseFormTester tests useForm behavior.
type UseFormTester[T any] = testutil.UseFormTester[T]

// NewUseFormTester creates a new use form tester.
func NewUseFormTester[T any](comp bubbly.Component) *UseFormTester[T] {
	return testutil.NewUseFormTester[T](comp)
}

// UseLocalStorageTester tests useLocalStorage behavior.
type UseLocalStorageTester[T any] = testutil.UseLocalStorageTester[T]

// NewUseLocalStorageTester creates a new use local storage tester.
func NewUseLocalStorageTester[T any](comp bubbly.Component, storage composables.Storage) *UseLocalStorageTester[T] {
	return testutil.NewUseLocalStorageTester[T](comp, storage)
}

// =============================================================================
// Watch Testers
// =============================================================================

// WatchEffectTester tests watchEffect behavior.
type WatchEffectTester = testutil.WatchEffectTester

// NewWatchEffectTester creates a new watch effect tester.
var NewWatchEffectTester = testutil.NewWatchEffectTester

// DeepWatchTester tests deep watch behavior.
type DeepWatchTester = testutil.DeepWatchTester

// NewDeepWatchTester creates a new deep watch tester.
var NewDeepWatchTester = testutil.NewDeepWatchTester

// CustomComparatorTester tests custom comparator behavior.
type CustomComparatorTester = testutil.CustomComparatorTester

// NewCustomComparatorTester creates a new custom comparator tester.
var NewCustomComparatorTester = testutil.NewCustomComparatorTester

// ComputedCacheVerifier verifies computed caching.
type ComputedCacheVerifier = testutil.ComputedCacheVerifier

// NewComputedCacheVerifier creates a new computed cache verifier.
var NewComputedCacheVerifier = testutil.NewComputedCacheVerifier

// FlushModeController controls flush modes for testing.
type FlushModeController = testutil.FlushModeController

// NewFlushModeController creates a new flush mode controller.
var NewFlushModeController = testutil.NewFlushModeController

// =============================================================================
// Provide/Inject Tester
// =============================================================================

// ProvideInjectTester tests provide/inject behavior.
type ProvideInjectTester = testutil.ProvideInjectTester

// NewProvideInjectTester creates a new provide/inject tester.
var NewProvideInjectTester = testutil.NewProvideInjectTester

// =============================================================================
// Router Testers
// =============================================================================

// PathMatchingTester tests path matching.
type PathMatchingTester = testutil.PathMatchingTester

// NewPathMatchingTester creates a new path matching tester.
func NewPathMatchingTester(r *router.Router) *PathMatchingTester {
	return testutil.NewPathMatchingTester(r)
}

// NamedRoutesTester tests named routes.
type NamedRoutesTester = testutil.NamedRoutesTester

// NewNamedRoutesTester creates a new named routes tester.
func NewNamedRoutesTester(r *router.Router) *NamedRoutesTester {
	return testutil.NewNamedRoutesTester(r)
}

// NestedRoutesTester tests nested routes.
type NestedRoutesTester = testutil.NestedRoutesTester

// NewNestedRoutesTester creates a new nested routes tester.
func NewNestedRoutesTester(r *router.Router) *NestedRoutesTester {
	return testutil.NewNestedRoutesTester(r)
}

// QueryParamsTester tests query parameters.
type QueryParamsTester = testutil.QueryParamsTester

// NewQueryParamsTester creates a new query params tester.
func NewQueryParamsTester(r *router.Router) *QueryParamsTester {
	return testutil.NewQueryParamsTester(r)
}

// RouteGuardTester tests route guards.
type RouteGuardTester = testutil.RouteGuardTester

// NewRouteGuardTester creates a new route guard tester.
func NewRouteGuardTester(r *router.Router) *RouteGuardTester {
	return testutil.NewRouteGuardTester(r)
}

// HistoryTester tests router history.
type HistoryTester = testutil.HistoryTester

// NewHistoryTester creates a new history tester.
func NewHistoryTester(r *router.Router) *HistoryTester {
	return testutil.NewHistoryTester(r)
}

// NavigationSimulator simulates navigation.
type NavigationSimulator = testutil.NavigationSimulator

// NewNavigationSimulator creates a new navigation simulator.
func NewNavigationSimulator(r *router.Router) *NavigationSimulator {
	return testutil.NewNavigationSimulator(r)
}

// =============================================================================
// Dependency Tracking
// =============================================================================

// DependencyTrackingInspector inspects dependency tracking.
type DependencyTrackingInspector = testutil.DependencyTrackingInspector

// NewDependencyTrackingInspector creates a new dependency tracking inspector.
var NewDependencyTrackingInspector = testutil.NewDependencyTrackingInspector

// DependencyGraph represents a dependency graph.
type DependencyGraph = testutil.DependencyGraph

// DependencyNode represents a node in the dependency graph.
type DependencyNode = testutil.DependencyNode

// DependencyEdge represents an edge in the dependency graph.
type DependencyEdge = testutil.DependencyEdge

// =============================================================================
// Data Factories
// =============================================================================

// DataFactory generates test data.
type DataFactory[T any] = testutil.DataFactory[T]

// NewFactory creates a new data factory.
func NewFactory[T any](generator func() T) *DataFactory[T] {
	return testutil.NewFactory(generator)
}

// IntFactory creates an int factory.
var IntFactory = testutil.IntFactory

// StringFactory creates a string factory.
var StringFactory = testutil.StringFactory

// GenerateArgs configures data generation.
type GenerateArgs = testutil.GenerateArgs

// =============================================================================
// Error Testing
// =============================================================================

// ErrorTesting provides error testing utilities.
type ErrorTesting = testutil.ErrorTesting

// NewErrorTesting creates new error testing utilities.
var NewErrorTesting = testutil.NewErrorTesting

// =============================================================================
// Template Safety
// =============================================================================

// TemplateSafetyTester tests template safety.
type TemplateSafetyTester = testutil.TemplateSafetyTester

// NewTemplateSafetyTester creates a new template safety tester.
var NewTemplateSafetyTester = testutil.NewTemplateSafetyTester

// SafetyViolation represents a safety violation.
type SafetyViolation = testutil.SafetyViolation

// =============================================================================
// Observability Assertions
// =============================================================================

// ObservabilityAssertions provides observability assertions.
type ObservabilityAssertions = testutil.ObservabilityAssertions

// NewObservabilityAssertions creates new observability assertions.
var NewObservabilityAssertions = testutil.NewObservabilityAssertions

// =============================================================================
// Test Hooks
// =============================================================================

// TestHooks provides test lifecycle hooks.
type TestHooks = testutil.TestHooks

// NewTestHooks creates new test hooks.
var NewTestHooks = testutil.NewTestHooks

// =============================================================================
// Matchers
// =============================================================================

// Matcher is the interface for custom matchers.
type Matcher = testutil.Matcher

// BeNil returns a matcher that checks for nil.
var BeNil = testutil.BeNil

// BeEmpty returns a matcher that checks for empty.
var BeEmpty = testutil.BeEmpty

// HaveLength returns a matcher that checks length.
var HaveLength = testutil.HaveLength
