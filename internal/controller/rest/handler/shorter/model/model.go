package model

// ShortenRequest — тело запроса POST /api/shorten.
type ShortenRequest struct {
	URL string `json:"url"` // Исходный URL для сокращения
}

// ShortenResponse — тело ответа POST /api/shorten.
type ShortenResponse struct {
	Result string `json:"result"` // Сокращённый URL
}

// BatchShortenRequest — элемент запроса POST /api/shorten/batch.
type BatchShortenRequest struct {
	CorrelationID string `json:"correlation_id"` // ID для сопоставления с ответом
	OriginalURL   string `json:"original_url"`   // Исходный URL
}

// BatchShortenResponse — элемент ответа POST /api/shorten/batch.
type BatchShortenResponse struct {
	CorrelationID string `json:"correlation_id"` // ID для сопоставления с запросом
	ShortURL      string `json:"short_url"`      // Сокращённый URL
}
