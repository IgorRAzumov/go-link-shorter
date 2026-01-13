package link

import (
	"context"
	"errors"
	"fmt"

	handlermodel "github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/shorter/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"github.com/IgorRAzumov/go-link-shorter/pkg/urlutil"
)

type Usecase struct {
	resolver service.ResolverService
	shorter  service.ShorterService
	baseURL  string
}

func NewLinkUsecase(resolver service.ResolverService, shorter service.ShorterService, baseURL string) *Usecase {
	return &Usecase{resolver: resolver, shorter: shorter, baseURL: baseURL}
}

func (usecase *Usecase) GetFullURLByShorKey(ctx context.Context, shortKey string) (string, error) {
	return usecase.resolver.GetFullLink(ctx, shortKey)
}

func (usecase *Usecase) CreateShortKey(ctx context.Context, URL string) (string, error) {
	newShortKey, err := usecase.shorter.CreateShortKey(ctx, URL)
	if err != nil {
		var conflictErr *model.URLConflictError
		if errors.As(err, &conflictErr) {
			return conflictErr.ExistingShortKey, conflictErr
		}
		return "", err
	}
	return newShortKey, nil
}

func (usecase *Usecase) GetBaseURL() string {
	return usecase.baseURL
}

func (usecase *Usecase) CreateShortKeysBatch(ctx context.Context, urls []string) (map[string]string, error) {
	result := make(map[string]string)
	linksToSave := make([]*model.Link, 0)

	for _, url := range urls {
		normalizedURL := urlutil.NormalizeURL(url)
		existedShortKey := usecase.resolver.GetShortKeyByURL(ctx, normalizedURL)
		if existedShortKey != "" {
			result[normalizedURL] = existedShortKey
			continue
		}

		shortKey := usecase.shorter.GenerateShortKey(normalizedURL)
		if shortKey == "" {
			return nil, model.ErrShortKeyCreation
		}

		result[normalizedURL] = shortKey
		linksToSave = append(linksToSave, &model.Link{
			ShortKey: shortKey,
			FullURL:  normalizedURL,
		})
	}

	if len(linksToSave) > 0 {
		err := usecase.shorter.CreateShortKeys(ctx, linksToSave)
		if err != nil {
			var conflictErr *model.URLConflictError
			if errors.As(err, &conflictErr) {
				for _, link := range linksToSave {
					if usecase.resolver.GetShortKeyByURL(ctx, link.FullURL) == conflictErr.ExistingShortKey {
						result[link.FullURL] = conflictErr.ExistingShortKey
					}
				}
			} else {
				return nil, err
			}
		}
	}

	return result, nil
}

func (usecase *Usecase) ProcessBatchShortenRequests(ctx context.Context, requests []handlermodel.BatchShortenRequest, scheme, host string) ([]handlermodel.BatchShortenResponse, error) {
	domainRequests := make([]model.BatchShortenRequest, 0, len(requests))
	for _, req := range requests {
		domainRequests = append(domainRequests, model.BatchShortenRequest{
			CorrelationID: req.CorrelationID,
			OriginalURL:   req.OriginalURL,
		})
	}

	batchResults, err := usecase.shorter.ProcessBatchShortenRequests(ctx, domainRequests, usecase.resolver)
	if err != nil {
		return nil, err
	}

	batchResponses := make([]handlermodel.BatchShortenResponse, 0, len(batchResults))
	for _, result := range batchResults {
		shortURL := usecase.buildShortURL(result.ShortKey, scheme, host)
		batchResponses = append(batchResponses, handlermodel.BatchShortenResponse{
			CorrelationID: result.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	return batchResponses, nil
}

func (usecase *Usecase) buildShortURL(shortKey, scheme, host string) string {
	if usecase.baseURL != "" {
		return fmt.Sprintf("%s/%s", usecase.baseURL, shortKey)
	}
	return fmt.Sprintf("%s://%s/%s", scheme, host, shortKey)
}
