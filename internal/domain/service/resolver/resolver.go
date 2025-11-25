package resolver

import (
	"context"
	"fmt"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/repository"
)

type Service struct {
	linkRepo repository.LinkRepository
}

func NewResolverService(repository repository.LinkRepository) *Service {
	return &Service{repository}
}

func (service *Service) GetFullLink(context context.Context, shortKey string) (string, error) {
	if shortKey == "" || !service.linkRepo.IsExistShortKey(context, shortKey) {
		return "", fmt.Errorf("unknown shortKey: %s", shortKey)
	}
	return service.linkRepo.GetByShortKey(context, shortKey), nil
}

func (service *Service) GetShortKeyByURL(context context.Context, URL string) string {
	return service.linkRepo.GetShortKeyByURL(context, URL)
}
