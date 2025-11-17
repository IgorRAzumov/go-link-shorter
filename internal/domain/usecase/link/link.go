package link

import (
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

func (usecase *Usecase) GetFullURLByShorKey(shortKey string) (string, error) {
	return usecase.resolver.GetFullLink(shortKey)
}

func (usecase *Usecase) CreateShortKey(URL string) string {
	shortKey := usecase.resolver.GetShortKeyByURL(URL)
	if shortKey == "" {
		return usecase.shorter.CreateShortKey(URL)
	}
	return shortKey
}

func (usecase *Usecase) GetBaseUrl() string {
	return usecase.baseURL
}
