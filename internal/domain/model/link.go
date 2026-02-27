package model

// Link содержит данные сокращённой ссылки.
type Link struct {
	ShortKey    string // Короткий идентификатор
	FullURL     string // Исходный URL
	UserID      string // ID владельца
	DeletedFlag bool   // Флаг мягкого удаления
}

// BatchShortenRequest — запрос на сокращение URL в batch-операции.
type BatchShortenRequest struct {
	CorrelationID string // ID для сопоставления с ответом
	OriginalURL   string // Исходный URL
}

// BatchShortenResult — результат batch-сокращения.
type BatchShortenResult struct {
	CorrelationID string // ID для сопоставления с запросом
	ShortKey      string // Сокращённый ключ
}
