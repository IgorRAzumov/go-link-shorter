package healthcheck

import (
	"context"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
)

type Usecase struct {
	service service.HealthCheckService
}

func NewHealthCheckUsecase(service service.HealthCheckService) *Usecase {
	return &Usecase{service: service}
}

func (usecase *Usecase) CheckSystemConnections(ctx context.Context) (bool, error) {
	return usecase.service.CheckStorageConnection(ctx)
}
