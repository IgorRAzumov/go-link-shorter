package urlutil

import "strings"

func NormalizeURL(url string) string {
	return strings.TrimSuffix(strings.TrimSpace(url), "/")
}
