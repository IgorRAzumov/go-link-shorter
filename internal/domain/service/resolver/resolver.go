package resolver

import (
	"fmt"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/repository"
)

type Service struct {
	linkRepo repository.LinkRepository
}

func NewResolverService(repository repository.LinkRepository) *Service {
	return &Service{repository}
}

func (service *Service) GetFullLink(shortKey string) (string, error) {
	if !service.linkRepo.IsExistShortKey(shortKey) {
		return "", fmt.Errorf("unknown shortKey: %s", shortKey)
	}
	return service.linkRepo.GetByShortKey(shortKey), nil
}

func (service *Service) GetShortKeyByURL(URL string) string {
	return service.linkRepo.GetShortKeyByURL(URL)
}
