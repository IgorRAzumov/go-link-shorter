package urlutil

import (
	"fmt"
	"net/url"
	"strings"
)

// BuildShortURL формирует полный сокращённый URL. Если задан baseURL — используется он,
// иначе URL собирается из scheme и host. Параметры scheme/host допускается оставлять пустыми
// (актуально для транспортов без понятия хоста, например gRPC).
func BuildShortURL(baseURL, scheme, host, shortKey string) string {
	if baseURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(baseURL, "/"), shortKey)
	}
	if scheme == "" {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/%s", scheme, host, shortKey)
}

// ValidateURL проверяет, что строка представляет собой корректный абсолютный URL (со схемой и хостом),
// и возвращает нормализованное значение.
func ValidateURL(raw string) (string, error) {
	normalized := NormalizeURL(raw)
	parsed, err := url.Parse(normalized)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("incorrect URL: %q", raw)
	}
	return NormalizeURL(parsed.String()), nil
}
