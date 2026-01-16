package deleter

import (
	"context"
	"errors"
	"time"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/repository"
	"github.com/rs/zerolog/log"
)

type Task struct {
	UserID   string
	ShortKey string
}

type Service struct {
	repo    repository.LinkRepository
	channel chan Task

	maxBatchSize int
}

const DefaultMaxBatch = 100
const defaultTimeSeconds = 1
const defaultQueueCap = 1024

func New(repo repository.LinkRepository, maxBatchSize int) *Service {
	maxBatch := maxBatchSize

	queueCap := defaultQueueCap

	return &Service{
		repo:         repo,
		channel:      make(chan Task, queueCap),
		maxBatchSize: maxBatch,
	}
}

func (service *Service) Start(ctx context.Context) {
	go service.run(ctx)
}

func (service *Service) Enqueue(userID string, shortKeys []string) error {
	if userID == "" {
		return errors.New("empty userID")
	}
	if len(shortKeys) == 0 {
		return nil
	}

	for _, key := range shortKeys {
		if key == "" {
			continue
		}
		service.channel <- Task{UserID: userID, ShortKey: key}
	}
	return nil
}

func (service *Service) run(ctx context.Context) {
	ticker := time.NewTicker(defaultTimeSeconds * time.Second)
	defer ticker.Stop()

	batch := make([]Task, 0, service.maxBatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		byUser := make(map[string][]string)
		for _, task := range batch {
			byUser[task.UserID] = append(byUser[task.UserID], task.ShortKey)
		}

		for userID, keys := range byUser {
			if err := service.repo.MarkDeleted(context.Background(), userID, keys); err != nil {
				log.Error().Err(err).Str("user_id", userID).Int("count", len(keys)).Msg("async delete failed")
			}
		}

		batch = batch[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case task := <-service.channel:
			batch = append(batch, task)
			if len(batch) >= service.maxBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}
