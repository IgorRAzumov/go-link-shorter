package shorter

import (
	"context"
	"errors"
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
	"github.com/IgorRAzumov/go-link-shorter/pkg/urlutil"
	"github.com/rs/zerolog/log"
)

// GenerateShortenURL формирует полный сокращённый URL из baseURL или scheme+host запроса.
func GenerateShortenURL(baseURL string, shortKey string, request *http.Request) string {
	scheme := "http"
	if request.TLS != nil || request.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return urlutil.BuildShortURL(baseURL, scheme, request.Host, shortKey)
}

// GenerateShortKey вызывает use case для создания короткого ключа по URL.
func GenerateShortKey(body string, ctx context.Context, writer http.ResponseWriter, usecase usecase.LinkCreateUsecase) (string, error) {
	originalURL, err := parseURL(body)
	log.Logger.Debug().Msg("generate originalURL: " + body)
	if err != nil {
		common.BadRequestError(writer, err)
		return "", err
	}

	shortKey, err := usecase.CreateShortKey(ctx, common.NormalizeURL(originalURL.String()))
	if err != nil {
		var conflictErr *model.URLConflictError
		if !errors.As(err, &conflictErr) {
			common.BadRequestError(writer, err)
			return "", err
		}
		return "", err
	}
	log.Logger.Debug().Msg("returned shortKey: " + shortKey)

	return shortKey, nil
}
