package benchmarks

import (
	"github.com/procyon-projects/logy"
	"net/http"
	"testing"
)

func BenchmarkObtainLogger(b *testing.B) {
	b.Run("logy.Get", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logy.Get()
			}
		})
	})
	b.Run("logy.Named", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logy.Named("github.com/procyon-projects/logy/test/benchmark")
			}
		})
	})
	b.Run("logy.Of", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logy.Of[http.Client]()
			}
		})
	})
}
