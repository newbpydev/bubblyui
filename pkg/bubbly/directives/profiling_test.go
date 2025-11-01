package directives

import (
	"fmt"
	"os"
	"runtime/pprof"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestProfiling_CPUHotspots generates CPU profile for directive operations
// To analyze: go tool pprof -http=:8080 directives_cpu.prof
func TestProfiling_CPUHotspots(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping profiling in short mode")
	}

	cpuFile, err := os.Create("directives_cpu.prof")
	if err != nil {
		t.Fatalf("Could not create CPU profile: %v", err)
	}
	defer cpuFile.Close()

	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		t.Fatalf("Could not start CPU profile: %v", err)
	}
	defer pprof.StopCPUProfile()

	// Run representative workload
	t.Log("Running CPU profiling workload...")

	// 1. ForEach with large lists (most common operation)
	for i := 0; i < 10000; i++ {
		items := make([]int, 1000)
		for j := range items {
			items[j] = j
		}
		_ = ForEach(items, func(item int, idx int) string {
			return fmt.Sprintf("%d\n", item)
		}).Render()
	}

	// 2. Nested directives
	for i := 0; i < 5000; i++ {
		_ = If(true, func() string {
			return Show(true, func() string {
				items := []string{"a", "b", "c", "d", "e"}
				return ForEach(items, func(item string, idx int) string {
					return fmt.Sprintf("- %s\n", item)
				}).Render()
			}).Render()
		}).Render()
	}

	// 3. Complex conditional chains
	for i := 0; i < 5000; i++ {
		status := i % 4
		_ = If(status == 0, func() string {
			return "loading"
		}).ElseIf(status == 1, func() string {
			return "success"
		}).ElseIf(status == 2, func() string {
			return "error"
		}).Else(func() string {
			return "empty"
		}).Render()
	}

	// 4. Bind operations
	for i := 0; i < 5000; i++ {
		ref := bubbly.NewRef(fmt.Sprintf("value-%d", i))
		_ = Bind(ref).Render()
	}

	// 5. On directive operations
	for i := 0; i < 5000; i++ {
		handler := func(data interface{}) {}
		_ = On("click", handler).
			PreventDefault().
			StopPropagation().
			Render("content")
	}

	t.Log("CPU profiling complete. Analyze with: go tool pprof -http=:8080 directives_cpu.prof")
}

// TestProfiling_MemoryHotspots generates memory profile for directive operations
// To analyze: go tool pprof -http=:8080 directives_mem.prof
func TestProfiling_MemoryHotspots(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping profiling in short mode")
	}

	// Run representative workload
	t.Log("Running memory profiling workload...")

	// Large ForEach operations (high allocation)
	for i := 0; i < 1000; i++ {
		items := make([]string, 10000)
		for j := range items {
			items[j] = fmt.Sprintf("item-%d", j)
		}
		result := ForEach(items, func(item string, idx int) string {
			return item + "\n"
		}).Render()
		// Ensure result is used
		_ = len(result)
	}

	// Nested operations
	for i := 0; i < 500; i++ {
		categories := make([]struct {
			Name  string
			Items []string
		}, 10)
		for j := range categories {
			categories[j].Name = fmt.Sprintf("Category %d", j)
			categories[j].Items = []string{"a", "b", "c", "d", "e"}
		}

		_ = ForEach(categories, func(cat struct {
			Name  string
			Items []string
		}, idx int) string {
			header := fmt.Sprintf("=== %s ===\n", cat.Name)
			items := ForEach(cat.Items, func(item string, idx int) string {
				return fmt.Sprintf("  - %s\n", item)
			}).Render()
			return header + items
		}).Render()
	}

	// Write memory profile
	memFile, err := os.Create("directives_mem.prof")
	if err != nil {
		t.Fatalf("Could not create memory profile: %v", err)
	}
	defer memFile.Close()

	if err := pprof.WriteHeapProfile(memFile); err != nil {
		t.Fatalf("Could not write memory profile: %v", err)
	}

	t.Log("Memory profiling complete. Analyze with: go tool pprof -http=:8080 directives_mem.prof")
}

// TestProfiling_AllocationsHotspots identifies allocation hotspots
// To analyze: go tool pprof -http=:8080 -alloc_space directives_alloc.prof
func TestProfiling_AllocationsHotspots(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping profiling in short mode")
	}

	t.Log("Running allocation profiling workload...")

	// High-allocation scenarios
	for i := 0; i < 5000; i++ {
		// Many small ForEach operations
		for j := 0; j < 10; j++ {
			items := []int{1, 2, 3, 4, 5}
			_ = ForEach(items, func(item int, idx int) string {
				return fmt.Sprintf("%d", item)
			}).Render()
		}

		// Many If/Else operations
		for j := 0; j < 10; j++ {
			_ = If(j%2 == 0, func() string {
				return "even"
			}).Else(func() string {
				return "odd"
			}).Render()
		}

		// Many Bind operations
		for j := 0; j < 10; j++ {
			ref := bubbly.NewRef(j)
			_ = Bind(ref).Render()
		}
	}

	// Write allocation profile
	allocFile, err := os.Create("directives_alloc.prof")
	if err != nil {
		t.Fatalf("Could not create allocation profile: %v", err)
	}
	defer allocFile.Close()

	if err := pprof.Lookup("allocs").WriteTo(allocFile, 0); err != nil {
		t.Fatalf("Could not write allocation profile: %v", err)
	}

	t.Log("Allocation profiling complete. Analyze with: go tool pprof -http=:8080 -alloc_space directives_alloc.prof")
}

// BenchmarkProfiling_RealisticWorkload is a comprehensive benchmark for profiling
func BenchmarkProfiling_RealisticWorkload(b *testing.B) {
	type Todo struct {
		ID        int
		Title     string
		Completed bool
	}

	type Category struct {
		Name  string
		Todos []Todo
	}

	categories := []Category{
		{
			Name: "Work",
			Todos: []Todo{
				{ID: 1, Title: "Review PR #123", Completed: false},
				{ID: 2, Title: "Write documentation", Completed: true},
				{ID: 3, Title: "Fix critical bug", Completed: false},
				{ID: 4, Title: "Update dependencies", Completed: true},
			},
		},
		{
			Name: "Personal",
			Todos: []Todo{
				{ID: 5, Title: "Buy groceries", Completed: false},
				{ID: 6, Title: "Call dentist", Completed: true},
				{ID: 7, Title: "Plan vacation", Completed: false},
			},
		},
		{
			Name: "Learning",
			Todos: []Todo{
				{ID: 8, Title: "Read Go book chapter 5", Completed: false},
				{ID: 9, Title: "Watch Bubbletea tutorial", Completed: false},
				{ID: 10, Title: "Build sample project", Completed: true},
				{ID: 11, Title: "Contribute to OSS", Completed: false},
			},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Render complete todo list UI with all directives
		_ = ForEach(categories, func(cat Category, catIdx int) string {
			header := fmt.Sprintf("\n=== %s ===\n", cat.Name)

			todos := If(len(cat.Todos) > 0, func() string {
				return ForEach(cat.Todos, func(todo Todo, todoIdx int) string {
					checkbox := If(todo.Completed,
						func() string { return "[✓]" },
					).Else(func() string {
						return "[ ]"
					}).Render()

					return Show(true, func() string {
						return fmt.Sprintf("%s %d. %s\n", checkbox, todo.ID, todo.Title)
					}).Render()
				}).Render()
			}).Else(func() string {
				return "  No todos yet\n"
			}).Render()

			stats := fmt.Sprintf("  (%d total)\n", len(cat.Todos))

			return header + todos + stats
		}).Render()
	}
}
