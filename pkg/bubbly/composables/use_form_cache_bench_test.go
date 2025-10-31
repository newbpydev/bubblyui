package composables

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly/composables/reflectcache"
)

// TestForm for benchmarking
type BenchmarkForm struct {
	Name    string
	Email   string
	Age     int
	Address string
	Phone   string
}

// BenchmarkUseForm_SetField_WithoutCache measures SetField performance without cache
func BenchmarkUseForm_SetField_WithoutCache(b *testing.B) {
	// Ensure cache is disabled
	originalCache := reflectcache.GlobalCache
	reflectcache.GlobalCache = nil
	defer func() { reflectcache.GlobalCache = originalCache }()

	ctx := createTestContext()
	form := UseForm(ctx, BenchmarkForm{}, func(f BenchmarkForm) map[string]string {
		return make(map[string]string)
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		form.SetField("Name", "John Doe")
	}
}

// BenchmarkUseForm_SetField_WithCache measures SetField performance with cache enabled
func BenchmarkUseForm_SetField_WithCache(b *testing.B) {
	// Enable cache
	reflectcache.EnableGlobalCache()
	defer func() { reflectcache.GlobalCache = nil }()

	ctx := createTestContext()
	form := UseForm(ctx, BenchmarkForm{}, func(f BenchmarkForm) map[string]string {
		return make(map[string]string)
	})

	// Warm up cache with one call
	form.SetField("Name", "Warmup")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		form.SetField("Name", "John Doe")
	}
}

// BenchmarkUseForm_SetField_MultipleFields_WithoutCache tests multiple field updates
func BenchmarkUseForm_SetField_MultipleFields_WithoutCache(b *testing.B) {
	originalCache := reflectcache.GlobalCache
	reflectcache.GlobalCache = nil
	defer func() { reflectcache.GlobalCache = originalCache }()

	ctx := createTestContext()
	form := UseForm(ctx, BenchmarkForm{}, func(f BenchmarkForm) map[string]string {
		return make(map[string]string)
	})

	fields := []struct {
		name  string
		value interface{}
	}{
		{"Name", "John Doe"},
		{"Email", "john@example.com"},
		{"Age", 30},
		{"Address", "123 Main St"},
		{"Phone", "555-1234"},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, field := range fields {
			form.SetField(field.name, field.value)
		}
	}
}

// BenchmarkUseForm_SetField_MultipleFields_WithCache tests multiple field updates with cache
func BenchmarkUseForm_SetField_MultipleFields_WithCache(b *testing.B) {
	reflectcache.EnableGlobalCache()
	defer func() { reflectcache.GlobalCache = nil }()

	ctx := createTestContext()
	form := UseForm(ctx, BenchmarkForm{}, func(f BenchmarkForm) map[string]string {
		return make(map[string]string)
	})

	// Warm up cache
	form.SetField("Name", "Warmup")

	fields := []struct {
		name  string
		value interface{}
	}{
		{"Name", "John Doe"},
		{"Email", "john@example.com"},
		{"Age", 30},
		{"Address", "123 Main St"},
		{"Phone", "555-1234"},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, field := range fields {
			form.SetField(field.name, field.value)
		}
	}
}

// BenchmarkUseForm_Complete_WithoutCache measures complete form workflow without cache
func BenchmarkUseForm_Complete_WithoutCache(b *testing.B) {
	originalCache := reflectcache.GlobalCache
	reflectcache.GlobalCache = nil
	defer func() { reflectcache.GlobalCache = originalCache }()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx := createTestContext()
		form := UseForm(ctx, BenchmarkForm{}, func(f BenchmarkForm) map[string]string {
			errors := make(map[string]string)
			if f.Name == "" {
				errors["Name"] = "Name required"
			}
			if f.Email == "" {
				errors["Email"] = "Email required"
			}
			return errors
		})

		form.SetField("Name", "John Doe")
		form.SetField("Email", "john@example.com")
		form.Submit()
	}
}

// BenchmarkUseForm_Complete_WithCache measures complete form workflow with cache
func BenchmarkUseForm_Complete_WithCache(b *testing.B) {
	reflectcache.EnableGlobalCache()
	defer func() { reflectcache.GlobalCache = nil }()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx := createTestContext()
		form := UseForm(ctx, BenchmarkForm{}, func(f BenchmarkForm) map[string]string {
			errors := make(map[string]string)
			if f.Name == "" {
				errors["Name"] = "Name required"
			}
			if f.Email == "" {
				errors["Email"] = "Email required"
			}
			return errors
		})

		form.SetField("Name", "John Doe")
		form.SetField("Email", "john@example.com")
		form.Submit()
	}
}

// BenchmarkCacheComparison compares cached vs non-cached side-by-side
func BenchmarkCacheComparison(b *testing.B) {
	b.Run("WithoutCache", func(b *testing.B) {
		originalCache := reflectcache.GlobalCache
		reflectcache.GlobalCache = nil
		defer func() { reflectcache.GlobalCache = originalCache }()

		ctx := createTestContext()
		form := UseForm(ctx, BenchmarkForm{}, func(f BenchmarkForm) map[string]string {
			return make(map[string]string)
		})

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			form.SetField("Name", "John Doe")
		}
	})

	b.Run("WithCache", func(b *testing.B) {
		reflectcache.EnableGlobalCache()
		defer func() { reflectcache.GlobalCache = nil }()

		ctx := createTestContext()
		form := UseForm(ctx, BenchmarkForm{}, func(f BenchmarkForm) map[string]string {
			return make(map[string]string)
		})

		// Warm up
		form.SetField("Name", "Warmup")

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			form.SetField("Name", "John Doe")
		}
	})
}
