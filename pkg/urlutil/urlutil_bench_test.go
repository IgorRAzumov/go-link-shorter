package urlutil

import "testing"

func BenchmarkNormalizeURL(b *testing.B) {
	url := "https://example.com/path/to/page/"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NormalizeURL(url)
	}
}

func BenchmarkNormalizeURL_Short(b *testing.B) {
	url := "https://a.co/"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NormalizeURL(url)
	}
}

func BenchmarkNormalizeURL_WithSpaces(b *testing.B) {
	url := "  https://example.com/  "
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NormalizeURL(url)
	}
}
