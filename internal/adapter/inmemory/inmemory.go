package inmemory

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/IgorRAzumov/go-link-shorter/internal/adapter"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type LinkStorage struct {
	linksByID  sync.Map
	linksByURL sync.Map
	filePath   string
	mu         sync.RWMutex
}

func NewInMemoryFileStorage(filePath string) (*LinkStorage, error) {
	storage := &LinkStorage{
		filePath: filePath,
	}

	if err := storage.loadFromFile(); err != nil {
		log.Warn().Err(err).Msg("Failed to load data from file, starting with empty storage")
	}

	return storage, nil
}

func (storage *LinkStorage) GetByShortKey(ctx context.Context, shortURL string) string {
	storage.mu.RLock()
	defer storage.mu.RUnlock()

	log.Debug().Str("short_key", shortURL).Msg("storage: GetByShortKey")
	value, _ := storage.linksByID.Load(shortURL)
	if link, ok := value.(*adapter.Link); ok {
		return link.FullURL
	}
	return ""
}

func (storage *LinkStorage) IsExistShortKey(ctx context.Context, shortKey string) bool {
	storage.mu.RLock()
	defer storage.mu.RUnlock()

	log.Debug().Str("short_key", shortKey).Msg("storage: IsExistShortKey")
	_, ok := storage.linksByID.Load(shortKey)
	return ok
}

func (storage *LinkStorage) GetShortKeyByURL(ctx context.Context, URL string) string {
	storage.mu.RLock()
	defer storage.mu.RUnlock()

	log.Debug().Str("url", URL).Msg("storage: GetByURL")
	value, ok := storage.linksByURL.Load(URL)
	if ok {
		if link, ok := value.(*adapter.Link); ok {
			return link.ShortKey
		}
	}
	return ""
}

func (storage *LinkStorage) Save(ctx context.Context, domainLink *model.Link) error {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	link := adapter.FromDomainLink(domainLink)
	link.UUID = uuid.New().String()

	storage.linksByID.Store(link.ShortKey, link)
	storage.linksByURL.Store(link.FullURL, link)

	if err := storage.saveToFile(); err != nil {
		log.Error().Err(err).Msg("Failed to save data to file")
		return err
	}
	return nil
}

func (storage *LinkStorage) BatchSave(ctx context.Context, links []*model.Link) error {
	if len(links) == 0 {
		return nil
	}

	storage.mu.Lock()
	defer storage.mu.Unlock()

	for _, domainLink := range links {
		link := adapter.FromDomainLink(domainLink)
		link.UUID = uuid.New().String()

		storage.linksByID.Store(link.ShortKey, link)
		storage.linksByURL.Store(link.FullURL, link)
	}

	if err := storage.saveToFile(); err != nil {
		log.Error().Err(err).Msg("Failed to save data to file")
		return err
	}
	return nil
}

func (storage *LinkStorage) loadFromFile() error {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	if _, err := os.Stat(storage.filePath); os.IsNotExist(err) {
		log.Info().Str("file_path", storage.filePath).Msg("Storage file does not exist, starting with empty storage")
		return nil
	}

	data, err := os.ReadFile(storage.filePath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	var links []adapter.Link
	if err := json.Unmarshal(data, &links); err != nil {
		return err
	}

	for _, link := range links {
		linkCopy := link
		storage.linksByID.Store(linkCopy.ShortKey, &linkCopy)
		storage.linksByURL.Store(linkCopy.FullURL, &linkCopy)
	}

	log.Info().Int("count", len(links)).Str("file_path", storage.filePath).Msg("Loaded links from file")

	return nil
}

func (storage *LinkStorage) saveToFile() error {
	if storage.filePath == "" {
		return nil
	}

	allLinks := storage.getAllLinks()
	data, marshalError := json.MarshalIndent(allLinks, "", "  ")
	if marshalError != nil {
		return marshalError
	}

	if writeError := os.WriteFile(storage.filePath, data, 0644); writeError != nil {
		return writeError
	}

	log.Debug().Int("count", len(allLinks)).Str("file_path", storage.filePath).Msg("Saved links to file")

	return nil
}

func (storage *LinkStorage) getAllLinks() []*adapter.Link {
	var links []*adapter.Link
	storage.linksByID.Range(func(key, value interface{}) bool {
		if link, ok := value.(*adapter.Link); ok {
			links = append(links, link)
		}
		return true
	})
	return links
}
