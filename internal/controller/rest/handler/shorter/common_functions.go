package shorter

import (
	"context"
	"errors"
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
	"github.com/rs/zerolog/log"
)

func GenerateShortenURL(baseURL string, shortKey string, request *http.Request) string {
	var result string
	if baseURL != "" {
		result = baseURL + "/" + shortKey
	} else {
		scheme := "http"
		if request.TLS != nil || request.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		result = scheme + "://" + request.Host + "/" + shortKey
	}
	return result
}

func GenerateShortKey(body string, context context.Context, writer http.ResponseWriter, usecase usecase.LinkUsecase) (string, error) {
	originalURL, err := parseURL(body)
	log.Logger.Debug().Msg("generate originalURL: " + body)
	if err != nil {
		common.BadRequestError(writer, err)
		return "", err
	}

	shortKey, err := usecase.CreateShortKey(context, common.NormalizeURL(originalURL.String()))
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
