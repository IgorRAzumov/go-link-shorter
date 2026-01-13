package usecase

import "context"

type LinkUsecase interface {
	GetFullURLByShorKey(context context.Context, shortKey string) (string, error)
	CreateShortKey(context context.Context, URL string) (string, error)
	CreateShortKeysBatch(context context.Context, urls []string) (map[string]string, error)
	GetBaseURL() string
}

type HealthCheckUsecase interface {
	CheckSystemConnections(context context.Context) (bool, error)
}
