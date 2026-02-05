package usecase

import (
	"context"

	handlermodel "github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/shorter/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

type LinkCreateUsecase interface {
	CreateShortKey(ctx context.Context, URL string) (string, error)
	CreateShortKeysBatch(ctx context.Context, urls []string) (map[string]string, error)
	ProcessBatchShortenRequests(ctx context.Context, requests []handlermodel.BatchShortenRequest, scheme, host string) ([]handlermodel.BatchShortenResponse, error)
	GetBaseURL() string
}

type LinkReadUsecase interface {
	GetFullURLByShorKey(ctx context.Context, shortKey string) (string, error)
	GetUserURLs(ctx context.Context) ([]*model.Link, error)
	GetBaseURL() string
}

type HealthCheckUsecase interface {
	CheckSystemConnections(ctx context.Context) (bool, error)
}
