package urlutil

import "testing"

func BenchmarkNormalizeURL(b *testing.B) {
	url := "https://example.com/path/to/page/"
	for b.Loop() {
		_ = NormalizeURL(url)
	}
}

func BenchmarkNormalizeURL_Short(b *testing.B) {
	url := "https://a.co/"
	for b.Loop() {
		_ = NormalizeURL(url)
	}
}

func BenchmarkNormalizeURL_WithSpaces(b *testing.B) {
	url := "  https://example.com/  "
	for b.Loop() {
		_ = NormalizeURL(url)
	}
}
