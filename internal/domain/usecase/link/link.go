package link

import (
	"context"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
)

type Usecase struct {
	resolver service.ResolverService
	shorter  service.ShorterService
	baseURL  string
}

func NewLinkUsecase(resolver service.ResolverService, shorter service.ShorterService, baseURL string) *Usecase {
	return &Usecase{resolver: resolver, shorter: shorter, baseURL: baseURL}
}

func (usecase *Usecase) GetFullURLByShorKey(context context.Context, shortKey string) (string, error) {
	return usecase.resolver.GetFullLink(context, shortKey)
}

func (usecase *Usecase) CreateShortKey(context context.Context, URL string) string {
	shortKey := usecase.resolver.GetShortKeyByURL(context, URL)
	if shortKey == "" {
		return usecase.shorter.CreateShortKey(context, URL)
	}
	return shortKey
}

func (usecase *Usecase) GetBaseURL() string {
	return usecase.baseURL
}
