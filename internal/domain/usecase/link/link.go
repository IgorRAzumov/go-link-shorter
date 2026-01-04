package link

import (
	"context"
	"errors"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
)

type Usecase struct {
	resolver service.ResolverService
	shorter  service.ShorterService
	baseURL  string
}

func NewLinkUsecase(resolver service.ResolverService, shorter service.ShorterService, baseURL string) *Usecase {
	return &Usecase{resolver: resolver, shorter: shorter, baseURL: baseURL}
}

func (usecase *Usecase) GetFullURLByShorKey(context context.Context, shortKey string) (string, error) {
	return usecase.resolver.GetFullLink(context, shortKey)
}

func (usecase *Usecase) CreateShortKey(context context.Context, URL string) (string, error) {
	existedShortKey := usecase.resolver.GetShortKeyByURL(context, URL)
	if existedShortKey != "" {
		return existedShortKey, nil
	}

	newShortKey, err := usecase.shorter.CreateShortKey(context, URL)
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

func (usecase *Usecase) CreateShortKeysBatch(context context.Context, urls []string) (map[string]string, error) {
	result := make(map[string]string)
	linksToSave := make([]*model.Link, 0)

	for _, url := range urls {
		normalizedURL := common.NormalizeURL(url)
		existedShortKey := usecase.resolver.GetShortKeyByURL(context, normalizedURL)
		if existedShortKey != "" {
			result[normalizedURL] = existedShortKey
			continue
		}

		shortKey := usecase.shorter.GenerateShortKey(normalizedURL)
		if shortKey == "" {
			return nil, errors.New("short key creation error")
		}

		result[normalizedURL] = shortKey
		linksToSave = append(linksToSave, &model.Link{
			ShortKey: shortKey,
			FullURL:  normalizedURL,
		})
	}

	if len(linksToSave) > 0 {
		err := usecase.shorter.CreateShortKeys(context, linksToSave)
		if err != nil {
			var conflictErr *model.URLConflictError
			if errors.As(err, &conflictErr) {
				for _, link := range linksToSave {
					normalizedURL := common.NormalizeURL(link.FullURL)
					if usecase.resolver.GetShortKeyByURL(context, normalizedURL) == conflictErr.ExistingShortKey {
						result[normalizedURL] = conflictErr.ExistingShortKey
					}
				}
			} else {
				return nil, err
			}
		}
	}

	return result, nil
}
