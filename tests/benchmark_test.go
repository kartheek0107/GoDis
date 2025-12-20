package tests

import (
	"fmt"
	"testing"

	"github.com/kartheek0107/GoDis/internal/store"
)

// BenchmarkStoreSet measures how fast we can write to the engine.
// Run from root with: go test -bench=BenchmarkStoreSet -benchmem ./tests/...
func BenchmarkStoreSet(b *testing.B) {
	s := store.Newstore(store.Store{})
	b.ResetTimer() // Don't count the setup time

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		val := "value"
		s.Set(key, val)
	}
}

// BenchmarkStoreGet measures read performance.
func BenchmarkStoreGet(b *testing.B) {
	s := store.Newstore(store.Store{})
	s.Set("foo", "bar")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = s.Get("foo")
	}
}

// BenchmarkConcurrentSetGet simulates a real-world scenario:
// 10% Writes and 90% Reads happening concurrently.
func BenchmarkConcurrentSetGet(b *testing.B) {
	s := store.Newstore(store.Store{})
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%10 == 0 {
				s.Set("key", "val")
			} else {
				_, _ = s.Get("key")
			}
			i++
		}
	})
}
