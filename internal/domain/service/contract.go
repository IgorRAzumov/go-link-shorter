package service

import "context"

type ResolverService interface {
	GetFullLink(context context.Context, shortKey string) (string, error)
	GetShortKeyByURL(context context.Context, url string) string
}

type ShorterService interface {
	CreateShortKey(context context.Context, URL string) (string, error)
}

type HealthCheckService interface {
	CheckStorageConnection(context context.Context) (bool, error)
}
