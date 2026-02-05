package model

type Link struct {
	ShortKey    string
	FullURL     string
	UserID      string
	DeletedFlag bool
}

type BatchShortenRequest struct {
	CorrelationID string
	OriginalURL   string
}

type BatchShortenResult struct {
	CorrelationID string
	ShortKey      string
}
