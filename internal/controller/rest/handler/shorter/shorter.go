package shorter

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
	"github.com/rs/zerolog/log"
)

func Handler(usecase usecase.LinkUsecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if !IsTextPlain(request.Header.Get(common.ContentType)) {
			common.MethodNotAllowedError(writer)
			return
		}

		body, err := readBody(request)
		if err != nil || len(body) == 0 {
			common.BadRequestError(writer, err)
			return
		}

		originalURL, err := parseURL(string(body))
		if err != nil {
			common.BadRequestError(writer, err)
			return
		}

		context := request.Context()
		shortKey := usecase.CreateShortKey(context, originalURL.String())
		if shortKey == "" {
			common.BadRequestError(writer, err)
			return
		}

		baseURL := usecase.GetBaseURL()
		if baseURL != "" {
			sendResponse(writer, baseURL+"/"+shortKey)
		} else {
			scheme := "http"
			if request.TLS != nil || request.Header.Get("X-Forwarded-Proto") == "https" {
				scheme = "https"
			}
			sendResponse(writer, scheme+"://"+request.Host+"/"+shortKey)
		}
	}
}

func sendResponse(writer http.ResponseWriter, shortURL string) {
	writer.Header().Set(common.ContentType, common.TextPlain)
	writer.WriteHeader(http.StatusCreated)
	_, err := writer.Write([]byte(shortURL))
	if err != nil {
		common.InternalError(writer, err)
		return
	}
}

func readBody(request *http.Request) ([]byte, error) {
	body, err := io.ReadAll(request.Body)
	defer func(Body io.ReadCloser) {
		readerBodyErr := Body.Close()
		if err != nil {
			log.Error().Err(readerBodyErr).Msg("Request body error")
		}
	}(request.Body)

	return body, err
}

func parseURL(URL string) (*url.URL, error) {
	log.Debug().Str("url", URL).Msg("Parsing URL")
	parsedURL, err := url.Parse(common.NormalizeURL(URL))
	if err != nil {
		return nil, err
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return &url.URL{}, fmt.Errorf("incorrect URL: %s, error: %w", parsedURL, err)
	}

	return parsedURL, nil
}

func IsTextPlain(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return mediaType == common.TextPlain
}
