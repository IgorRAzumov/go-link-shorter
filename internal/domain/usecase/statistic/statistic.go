package statistic

import (
	"context"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
)

// Usecase — реализация StatsUsecase поверх StatsService.
type Usecase struct {
	statsService service.StatsService
}

// NewUsecase создаёт use case получения статистики сервиса.
func NewUsecase(statsService service.StatsService) *Usecase {
	return &Usecase{statsService: statsService}
}

// GetStats возвращает количество сокращённых URL и уникальных пользователей.
func (usecase *Usecase) GetStats(ctx context.Context) (model.ServiceStats, error) {
	return usecase.statsService.GetStats(ctx)
}
