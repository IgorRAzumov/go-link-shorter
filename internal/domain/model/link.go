package model

type Link struct {
	ShortKey string
	FullURL  string
}

type BatchShortenRequest struct {
	CorrelationID string
	OriginalURL   string
}

type BatchShortenResult struct {
	CorrelationID string
	ShortKey      string
}
