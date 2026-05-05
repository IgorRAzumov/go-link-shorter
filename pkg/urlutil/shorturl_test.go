package urlutil

import "testing"

func TestBuildShortURL(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		scheme   string
		host     string
		shortKey string
		want     string
	}{
		{"base url без слеша", "http://x.io", "", "", "abc", "http://x.io/abc"},
		{"base url со слешем", "http://x.io/", "", "", "abc", "http://x.io/abc"},
		{"fallback http", "", "http", "example.com:8080", "abc", "http://example.com:8080/abc"},
		{"fallback https", "", "https", "example.com", "abc", "https://example.com/abc"},
		{"fallback пустой scheme", "", "", "example.com", "abc", "http://example.com/abc"},
		{"приоритет у base", "http://base", "https", "ignored", "key", "http://base/key"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildShortURL(tc.base, tc.scheme, tc.host, tc.shortKey)
			if got != tc.want {
				t.Fatalf("BuildShortURL(%q,%q,%q,%q)=%q, want %q",
					tc.base, tc.scheme, tc.host, tc.shortKey, got, tc.want)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		got, err := ValidateURL(" https://example.com/a/ ")
		if err != nil {
			t.Fatal(err)
		}
		if got != "https://example.com/a" {
			t.Fatalf("unexpected: %q", got)
		}
	})
	t.Run("no scheme", func(t *testing.T) {
		if _, err := ValidateURL("example.com"); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("no host", func(t *testing.T) {
		if _, err := ValidateURL("http://"); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("empty", func(t *testing.T) {
		if _, err := ValidateURL(""); err == nil {
			t.Fatal("expected error")
		}
	})
}
