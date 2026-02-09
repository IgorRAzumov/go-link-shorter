package audit

import (
	"context"
	"sync"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/repository"
	"github.com/rs/zerolog/log"
)

type Service struct {
	mu           sync.RWMutex
	repositories []repository.AuditRepository
}

func NewAuditService() *Service {
	return &Service{}
}

func (publisher *Service) Register(repo repository.AuditRepository) {
	if repo == nil {
		return
	}
	publisher.mu.Lock()
	defer publisher.mu.Unlock()

	publisher.repositories = append(publisher.repositories, repo)
}

func (publisher *Service) AuditNewEvent(ctx context.Context, event model.AuditEvent) {
	publisher.mu.RLock()
	repositories := append([]repository.AuditRepository(nil), publisher.repositories...)
	publisher.mu.RUnlock()

	for _, repo := range repositories {
		if repo == nil {
			continue
		}
		if err := repo.Save(ctx, event); err != nil {
			log.Error().Err(err).Str("action", event.Action).Msg("audit notify failed")
		}
	}
}
