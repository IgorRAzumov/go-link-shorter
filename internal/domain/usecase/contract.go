package usecase

import (
	"context"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

// LinkCreateUsecase — use case создания сокращённых ссылок.
type LinkCreateUsecase interface {
	CreateShortKey(ctx context.Context, URL string) (string, error)
	CreateShortKeysBatch(ctx context.Context, urls []string) (map[string]string, error)
	ProcessBatchShortenRequests(ctx context.Context, requests []model.BatchShortenRequest) ([]model.BatchShortenResult, error)
	GetBaseURL() string
}

// LinkReadUsecase — use case чтения ссылок и получения полного URL.
type LinkReadUsecase interface {
	GetFullURLByShorKey(ctx context.Context, shortKey string) (string, error)
	GetUserURLs(ctx context.Context) ([]*model.Link, error)
	GetBaseURL() string
}

// LinkDeleteUsecase — use case мягкого удаления ссылок пользователя.
type LinkDeleteUsecase interface {
	DeleteUserURLs(ctx context.Context, shortKeys []string) error
}

// HealthCheckUsecase — use case проверки доступности системы.
type HealthCheckUsecase interface {
	CheckSystemConnections(ctx context.Context) (bool, error)
}

// StatsUsecase — use case получения статистики сервиса.
type StatsUsecase interface {
	GetStats(ctx context.Context) (model.ServiceStats, error)
}
