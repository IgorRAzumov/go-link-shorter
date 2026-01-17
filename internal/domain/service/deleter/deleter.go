package deleter

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/repository"
	"github.com/rs/zerolog/log"
)

type Task struct {
	UserID   string
	ShortKey string
}

type Service struct {
	repository   repository.LinkRepository
	inputs       []chan Task
	rr           uint32
	maxBatchSize int
	stop         chan struct{}
	done         chan struct{}
}

const DefaultMaxBatch = 100
const defaultTimeSeconds = 1
const defaultTotalCap = 1024
const defaultFanInShards = 5

func NewDeleterService(linkRepository repository.LinkRepository, maxBatchSize int) *Service {
	maxBatch := maxBatchSize
	if maxBatch <= 0 {
		maxBatch = DefaultMaxBatch
	}
	shards := defaultFanInShards
	perShardCap := defaultTotalCap / shards

	inputs := make([]chan Task, 0, shards)
	for i := 0; i < shards; i++ {
		inputs = append(inputs, make(chan Task, perShardCap))
	}

	return &Service{
		repository:   linkRepository,
		inputs:       inputs,
		maxBatchSize: maxBatch,
		stop:         make(chan struct{}),
		done:         make(chan struct{}),
	}
}

func (service *Service) Start() {
	go service.run()
}

func (service *Service) Stop() {
	select {
	case <-service.stop:
		return
	default:
		close(service.stop)
	}
}

func (service *Service) Wait() {
	<-service.done
}

func (service *Service) Enqueue(userID string, shortKeys []string) error {
	if userID == "" {
		return model.ErrEmptyUser
	}
	if len(shortKeys) == 0 || len(service.inputs) == 0 {
		return nil
	}

	idx := int(atomic.AddUint32(&service.rr, 1)-1) % len(service.inputs)
	ch := service.inputs[idx]

	for _, key := range shortKeys {
		if key == "" {
			continue
		}
		ch <- Task{UserID: userID, ShortKey: key}
	}
	return nil
}

func fanIn(channels ...<-chan Task) <-chan Task {
	out := make(chan Task)
	var wg sync.WaitGroup
	wg.Add(len(channels))

	for _, ch := range channels {
		go func(c <-chan Task) {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func (service *Service) run() {
	defer close(service.done)

	ticker := time.NewTicker(defaultTimeSeconds * time.Second)
	defer ticker.Stop()

	batch := make([]Task, 0, service.maxBatchSize)
	merged := fanIn(sliceToReadOnly(service.inputs)...)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		byUser := make(map[string][]string)
		for _, task := range batch {
			byUser[task.UserID] = append(byUser[task.UserID], task.ShortKey)
		}

		for userID, keys := range byUser {
			if err := service.repository.MarkDeleted(context.Background(), userID, keys); err != nil {
				log.Error().Err(err).Str("user_id", userID).Int("count", len(keys)).Msg("async delete failed")
			}
		}

		batch = batch[:0]
	}

	for {
		select {
		case <-service.stop:
			for _, ch := range service.inputs {
				close(ch)
			}
			for task := range merged {
				batch = append(batch, task)
				if len(batch) >= service.maxBatchSize {
					flush()
				}
			}
			flush()
			return

		case task, ok := <-merged:
			if !ok {
				flush()
				return
			}
			batch = append(batch, task)
			if len(batch) >= service.maxBatchSize {
				flush()
			}

		case <-ticker.C:
			flush()
		}
	}
}

func sliceToReadOnly(chs []chan Task) []<-chan Task {
	out := make([]<-chan Task, 0, len(chs))
	for _, ch := range chs {
		out = append(out, ch)
	}
	return out
}
