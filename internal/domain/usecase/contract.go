package usecase

import "context"

type LinkUsecase interface {
	GetFullURLByShorKey(context context.Context, shortKey string) (string, error)
	CreateShortKey(context context.Context, URL string) (string, error)
	GetBaseURL() string
}
