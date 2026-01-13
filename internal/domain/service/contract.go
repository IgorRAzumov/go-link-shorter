package service

import (
	"context"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

type ResolverService interface {
	GetFullLink(ctx context.Context, shortKey string) (string, error)
	GetShortKeyByURL(ctx context.Context, url string) string
}

type ShorterService interface {
	CreateShortKey(ctx context.Context, URL string) (string, error)
	GenerateShortKey(URL string) string
	CreateShortKeys(ctx context.Context, links []*model.Link) error
	ProcessBatchShortenRequests(ctx context.Context, requests []model.BatchShortenRequest, resolver ResolverService) ([]model.BatchShortenResult, error)
}

type HealthCheckService interface {
	CheckStorageConnection(ctx context.Context) (bool, error)
}
