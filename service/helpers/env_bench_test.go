package helpers

import "testing"

func BenchmarkGetenv(b *testing.B) {
	b.Run("env is defined", func(b *testing.B) {
		b.Setenv("TEST_ENV", "Foo")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Getenv("TEST_ENV", "default")
		}
	})

	b.Run("env is undefined", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Getenv("TEST_ENV", "default")
		}
	})
}
