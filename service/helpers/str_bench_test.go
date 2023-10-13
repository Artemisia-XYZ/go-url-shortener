package helpers

import "testing"

func BenchmarkStrRandom(b *testing.B) {
	for i := 0; i < b.N; i++ {
		StrRandom(16)
	}
}
