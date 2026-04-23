package stats

import (
	"context"
	"fmt"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/repository"
)

// Service — доменный сервис агрегированной статистики (чтение из хранилища).
type Service struct {
	repository repository.StatsRepository
}

// NewService создаёт сервис статистики.
func NewService(repository repository.StatsRepository) *Service {
	return &Service{repository: repository}
}

// GetStats возвращает количество сокращённых URL и уникальных пользователей.
func (service *Service) GetStats(ctx context.Context) (model.ServiceStats, error) {
	urls, err := service.repository.CountURLs(ctx)
	if err != nil {
		return model.ServiceStats{}, fmt.Errorf("count urls: %w", err)
	}
	users, err := service.repository.CountUsers(ctx)
	if err != nil {
		return model.ServiceStats{}, fmt.Errorf("count users: %w", err)
	}
	return model.ServiceStats{URLs: urls, Users: users}, nil
}
