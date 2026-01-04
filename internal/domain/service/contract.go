package service

import (
	"context"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

type ResolverService interface {
	GetFullLink(context context.Context, shortKey string) (string, error)
	GetShortKeyByURL(context context.Context, url string) string
}

type ShorterService interface {
	CreateShortKey(context context.Context, URL string) (string, error)
	GenerateShortKey(URL string) string
	CreateShortKeys(context context.Context, links []*model.Link) error
}

type HealthCheckService interface {
	CheckStorageConnection(context context.Context) (bool, error)
}
