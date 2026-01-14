package resolver

import (
	"context"
	"fmt"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/repository"
)

type Service struct {
	linkRepo repository.LinkRepository
}

func NewResolverService(repository repository.LinkRepository) *Service {
	return &Service{repository}
}

func (service *Service) GetFullLink(ctx context.Context, shortKey string) (string, error) {
	if shortKey == "" || !service.linkRepo.IsExistShortKey(ctx, shortKey) {
		return "", fmt.Errorf("unknown shortKey: %s", shortKey)
	}
	return service.linkRepo.GetByShortKey(ctx, shortKey), nil
}

func (service *Service) GetShortKeyByURL(ctx context.Context, URL string) string {
	return service.linkRepo.GetShortKeyByURL(ctx, URL)
}

func (service *Service) GetByUserID(ctx context.Context, userID string) ([]*model.Link, error) {
	return service.linkRepo.GetByUserID(ctx, userID)
}
