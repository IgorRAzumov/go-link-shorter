package healthcheck

import (
	"context"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/repository"
)

type Service struct {
	repository repository.HealthCheckRepository
}

func NewHealthCheckService(repository repository.HealthCheckRepository) *Service {
	return &Service{repository: repository}
}
func (service *Service) CheckStorageConnection(ctx context.Context) (bool, error) {
	return service.repository.CheckStorageConnection(ctx)
}
