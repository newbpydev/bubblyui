package reflectcache

import (
	"reflect"
	"testing"
)

// Benchmark struct for testing
type BenchForm struct {
	Name    string
	Email   string
	Age     int
	Address string
	Phone   string
}

// BenchmarkFieldCache_GetFieldIndex_Hit measures cache hit performance
// Target: < 10ns per hit
func BenchmarkFieldCache_GetFieldIndex_Hit(b *testing.B) {
	cache := NewFieldCache()
	formType := reflect.TypeOf(BenchForm{})

	// Warm up cache
	cache.CacheType(formType)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = cache.GetFieldIndex(formType, "Name")
	}
}

// BenchmarkFieldCache_GetFieldIndex_Miss measures cache miss performance
func BenchmarkFieldCache_GetFieldIndex_Miss(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache := NewFieldCache()
		formType := reflect.TypeOf(BenchForm{})
		_, _ = cache.GetFieldIndex(formType, "Name")
	}
}

// BenchmarkFieldCache_GetFieldType_Hit measures type lookup with cache hit
func BenchmarkFieldCache_GetFieldType_Hit(b *testing.B) {
	cache := NewFieldCache()
	formType := reflect.TypeOf(BenchForm{})

	// Warm up cache
	cache.CacheType(formType)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = cache.GetFieldType(formType, "Email")
	}
}

// BenchmarkFieldCache_CacheType measures type caching overhead
func BenchmarkFieldCache_CacheType(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache := NewFieldCache()
		formType := reflect.TypeOf(BenchForm{})
		_ = cache.CacheType(formType)
	}
}

// BenchmarkFieldCache_WarmUp measures WarmUp performance
func BenchmarkFieldCache_WarmUp(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache := NewFieldCache()
		cache.WarmUp(BenchForm{})
	}
}

// BenchmarkFieldCache_Stats measures statistics retrieval overhead
func BenchmarkFieldCache_Stats(b *testing.B) {
	cache := NewFieldCache()
	formType := reflect.TypeOf(BenchForm{})
	cache.GetFieldIndex(formType, "Name")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = cache.Stats()
	}
}

// BenchmarkDirectReflection_FieldByName measures direct reflection without caching
// This is the baseline for comparison
func BenchmarkDirectReflection_FieldByName(b *testing.B) {
	form := BenchForm{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		v := reflect.ValueOf(&form).Elem()
		_ = v.FieldByName("Name")
	}
}

// BenchmarkDirectReflection_Field measures direct reflection with index
func BenchmarkDirectReflection_Field(b *testing.B) {
	form := BenchForm{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		v := reflect.ValueOf(&form).Elem()
		_ = v.Field(0) // Name is at index 0
	}
}

// BenchmarkCachedVsDirect compares cached vs direct reflection
func BenchmarkCachedVsDirect(b *testing.B) {
	b.Run("Cached", func(b *testing.B) {
		cache := NewFieldCache()
		formType := reflect.TypeOf(BenchForm{})
		cache.CacheType(formType)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _ = cache.GetFieldIndex(formType, "Name")
		}
	})

	b.Run("Direct_FieldByName", func(b *testing.B) {
		form := BenchForm{}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			v := reflect.ValueOf(&form).Elem()
			_ = v.FieldByName("Name")
		}
	})
}

// BenchmarkFieldAccess_WithCache simulates UseForm SetField with cache
func BenchmarkFieldAccess_WithCache(b *testing.B) {
	cache := NewFieldCache()
	formType := reflect.TypeOf(BenchForm{})

	// Pre-warm cache
	cache.CacheType(formType)

	form := BenchForm{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate SetField flow with cache
		idx, ok := cache.GetFieldIndex(formType, "Name")
		if ok {
			v := reflect.ValueOf(&form).Elem()
			field := v.Field(idx)
			if field.CanSet() {
				field.SetString("test")
			}
		}
	}
}

// BenchmarkFieldAccess_WithoutCache simulates UseForm SetField without cache
func BenchmarkFieldAccess_WithoutCache(b *testing.B) {
	form := BenchForm{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate SetField flow without cache
		v := reflect.ValueOf(&form).Elem()
		field := v.FieldByName("Name")
		if field.IsValid() && field.CanSet() {
			field.SetString("test")
		}
	}
}

// BenchmarkMultipleFields_Sequential measures accessing multiple fields
func BenchmarkMultipleFields_Sequential(b *testing.B) {
	cache := NewFieldCache()
	formType := reflect.TypeOf(BenchForm{})
	cache.CacheType(formType)

	fields := []string{"Name", "Email", "Age", "Address", "Phone"}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, field := range fields {
			_, _ = cache.GetFieldIndex(formType, field)
		}
	}
}

// BenchmarkConcurrentAccess measures concurrent cache access
func BenchmarkConcurrentAccess(b *testing.B) {
	cache := NewFieldCache()
	formType := reflect.TypeOf(BenchForm{})
	cache.CacheType(formType)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = cache.GetFieldIndex(formType, "Name")
		}
	})
}

// BenchmarkCacheWithMultipleTypes measures cache with many different types
func BenchmarkCacheWithMultipleTypes(b *testing.B) {
	type Form1 struct{ F1 string }
	type Form2 struct{ F2 int }
	type Form3 struct{ F3 bool }
	type Form4 struct{ F4 float64 }
	type Form5 struct{ F5 []byte }

	cache := NewFieldCache()
	types := []reflect.Type{
		reflect.TypeOf(Form1{}),
		reflect.TypeOf(Form2{}),
		reflect.TypeOf(Form3{}),
		reflect.TypeOf(Form4{}),
		reflect.TypeOf(Form5{}),
	}

	// Pre-cache all types
	for _, t := range types {
		cache.CacheType(t)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		t := types[i%len(types)]
		field := "F" + string(rune('1'+i%len(types)))
		_, _ = cache.GetFieldIndex(t, field)
	}
}

// BenchmarkMemoryAllocation compares memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("Cache_ColdStart", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			cache := NewFieldCache()
			formType := reflect.TypeOf(BenchForm{})
			cache.GetFieldIndex(formType, "Name")
		}
	})

	b.Run("Cache_WarmCache", func(b *testing.B) {
		cache := NewFieldCache()
		formType := reflect.TypeOf(BenchForm{})
		// Warmup
		for i := 0; i < 10; i++ {
			cache.GetFieldIndex(formType, "Name")
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.GetFieldIndex(formType, "Name")
		}
	})

	b.Run("Direct_FieldByName", func(b *testing.B) {
		form := BenchForm{}

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			v := reflect.ValueOf(&form).Elem()
			v.FieldByName("Name")
		}
	})
}
