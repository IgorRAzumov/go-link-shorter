package link

import (
	"context"
	"errors"
	"fmt"

	handlermodel "github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/shorter/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"github.com/IgorRAzumov/go-link-shorter/pkg/urlutil"
)

type CreatorUsecase struct {
	resolver service.ResolverService
	shorter  service.ShorterService
	baseURL  string
}

type ReaderUsecase struct {
	resolver service.ResolverService
	baseURL  string
}

func NewLinkCreateUsecase(resolver service.ResolverService, shorter service.ShorterService, baseURL string) *CreatorUsecase {
	return &CreatorUsecase{resolver: resolver, shorter: shorter, baseURL: baseURL}
}

func NewLinkReadUsecase(resolver service.ResolverService, baseURL string) *ReaderUsecase {
	return &ReaderUsecase{resolver: resolver, baseURL: baseURL}
}

func (usecase *ReaderUsecase) GetFullURLByShorKey(ctx context.Context, shortKey string) (string, error) {
	return usecase.resolver.GetFullLink(ctx, shortKey)
}

func (usecase *CreatorUsecase) CreateShortKey(ctx context.Context, URL string) (string, error) {
	userID := authctx.UserID(ctx)
	newShortKey, err := usecase.shorter.CreateShortKey(ctx, URL, userID)
	if err != nil {
		var conflictErr *model.URLConflictError
		if errors.As(err, &conflictErr) {
			return conflictErr.ExistingShortKey, conflictErr
		}
		return "", err
	}
	return newShortKey, nil
}

func (usecase *CreatorUsecase) GetBaseURL() string {
	return usecase.baseURL
}

func (usecase *ReaderUsecase) GetBaseURL() string {
	return usecase.baseURL
}

func (usecase *CreatorUsecase) CreateShortKeysBatch(ctx context.Context, urls []string) (map[string]string, error) {
	userID := authctx.UserID(ctx)
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
			UserID:   userID,
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

func (usecase *CreatorUsecase) ProcessBatchShortenRequests(ctx context.Context, requests []handlermodel.BatchShortenRequest, scheme, host string) ([]handlermodel.BatchShortenResponse, error) {
	userID := authctx.UserID(ctx)
	domainRequests := make([]model.BatchShortenRequest, 0, len(requests))
	for _, req := range requests {
		domainRequests = append(domainRequests, model.BatchShortenRequest{
			CorrelationID: req.CorrelationID,
			OriginalURL:   req.OriginalURL,
		})
	}

	batchResults, err := usecase.shorter.ProcessBatchShortenRequests(ctx, domainRequests, usecase.resolver, userID)
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

func (usecase *ReaderUsecase) GetUserURLs(ctx context.Context) ([]*model.Link, error) {
	userID := authctx.UserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}
	return usecase.resolver.GetByUserID(ctx, userID)
}

func (usecase *CreatorUsecase) buildShortURL(shortKey, scheme, host string) string {
	if usecase.baseURL != "" {
		return fmt.Sprintf("%s/%s", usecase.baseURL, shortKey)
	}
	return fmt.Sprintf("%s://%s/%s", scheme, host, shortKey)
}
