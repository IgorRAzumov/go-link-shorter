package shorter

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/repository"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"github.com/IgorRAzumov/go-link-shorter/pkg/urlutil"
	"github.com/rs/zerolog/log"
)

type Service struct {
	linkRepo repository.LinkRepository
}

func NewShorterService(repository repository.LinkRepository) *Service {
	return &Service{repository}
}

func (service *Service) CreateShortKey(ctx context.Context, URL string) (string, error) {
	normalizedURL := urlutil.NormalizeURL(URL)
	shortKey := service.GenerateShortKey(normalizedURL)
	log.Debug().Str("short_key", shortKey).Str("url", normalizedURL).Msg("created shortKey")
	if shortKey == "" {
		return "", model.ErrShortKeyCreation
	}
	err := service.linkRepo.Save(ctx, &model.Link{ShortKey: shortKey, FullURL: normalizedURL})
	if err != nil {
		return "", err
	}
	return shortKey, nil
}

func (service *Service) GenerateShortKey(URL string) string {
	hash := sha256.Sum256([]byte(URL))
	return base64.RawURLEncoding.EncodeToString(hash[:8])
}

func (service *Service) CreateShortKeys(ctx context.Context, links []*model.Link) error {
	return service.linkRepo.BatchSave(ctx, links)
}

func (service *Service) ProcessBatchShortenRequests(ctx context.Context, requests []model.BatchShortenRequest, resolver service.ResolverService) ([]model.BatchShortenResult, error) {
	normalizedMap, urlToShortKey, linksToSave := service.prepareBatchData(requests, ctx, resolver)

	if err := service.saveBatchLinks(ctx, linksToSave, urlToShortKey); err != nil {
		return nil, err
	}

	return service.buildBatchResults(requests, normalizedMap, urlToShortKey), nil
}

func (service *Service) prepareBatchData(requests []model.BatchShortenRequest, ctx context.Context, resolver service.ResolverService) (map[string]string, map[string]string, []*model.Link) {
	normalizedMap := make(map[string]string, len(requests))
	urlToShortKey := make(map[string]string)
	linksToSave := make([]*model.Link, 0)
	processedURLs := make(map[string]bool)

	for _, req := range requests {
		normalizedURL := urlutil.NormalizeURL(req.OriginalURL)
		normalizedMap[req.OriginalURL] = normalizedURL

		if processedURLs[normalizedURL] {
			continue
		}
		processedURLs[normalizedURL] = true

		existedShortKey := resolver.GetShortKeyByURL(ctx, normalizedURL)
		if existedShortKey != "" {
			urlToShortKey[normalizedURL] = existedShortKey
			continue
		}

		shortKey := service.GenerateShortKey(normalizedURL)
		if shortKey == "" {
			continue
		}

		urlToShortKey[normalizedURL] = shortKey
		linksToSave = append(linksToSave, &model.Link{
			ShortKey: shortKey,
			FullURL:  normalizedURL,
		})
	}

	return normalizedMap, urlToShortKey, linksToSave
}

func (service *Service) saveBatchLinks(ctx context.Context, linksToSave []*model.Link, urlToShortKey map[string]string) error {
	if len(linksToSave) == 0 {
		return nil
	}

	err := service.linkRepo.BatchSave(ctx, linksToSave)
	if err != nil {
		return service.handleBatchSaveError(err, linksToSave, urlToShortKey)
	}

	return nil
}

func (service *Service) handleBatchSaveError(err error, linksToSave []*model.Link, urlToShortKey map[string]string) error {
	var conflictErr *model.URLConflictError
	if !errors.As(err, &conflictErr) {
		return err
	}

	existingShortKey := conflictErr.ExistingShortKey
	for _, link := range linksToSave {
		if link.ShortKey == existingShortKey {
			urlToShortKey[link.FullURL] = existingShortKey
			return nil
		}
	}

	if len(linksToSave) > 0 {
		urlToShortKey[linksToSave[0].FullURL] = existingShortKey
		return nil
	}

	return err
}

func (service *Service) buildBatchResults(requests []model.BatchShortenRequest, normalizedMap map[string]string, urlToShortKey map[string]string) []model.BatchShortenResult {
	batchResults := make([]model.BatchShortenResult, 0, len(requests))
	for _, req := range requests {
		normalizedURL := normalizedMap[req.OriginalURL]
		shortKey := urlToShortKey[normalizedURL]
		if shortKey != "" {
			batchResults = append(batchResults, model.BatchShortenResult{
				CorrelationID: req.CorrelationID,
				ShortKey:      shortKey,
			})
		}
	}
	return batchResults
}
