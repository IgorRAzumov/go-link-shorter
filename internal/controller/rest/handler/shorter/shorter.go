package shorter

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
)

func Handler(usecase usecase.LinkUsecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || !IsTextPlain(request.Header.Get(common.ContentType)) {
			common.BadRequestError(writer, nil)
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

		shortKey := usecase.CreateShortKey(originalURL.String())
		if shortKey == "" {
			common.BadRequestError(writer, err)
			return
		}

		baseURL := usecase.GetBaseUrl()
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
			log.Printf("Request body error: %v", readerBodyErr)
		}
	}(request.Body)

	return body, err
}

func parseURL(URL string) (*url.URL, error) {
	log.Printf("Parsing URL: %s", URL)
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
