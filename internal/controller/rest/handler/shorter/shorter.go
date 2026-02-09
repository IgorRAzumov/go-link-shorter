package shorter

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"time"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
	"github.com/rs/zerolog/log"
)

func Handler(usecase usecase.LinkCreateUsecase, auditor service.AuditorService) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if !isTextPlain(request.Header.Get(common.ContentType)) {
			common.MethodNotAllowedError(writer)
			return
		}

		body, err := readBody(request)
		if err != nil || len(body) == 0 {
			common.BadRequestError(writer, err)
			return
		}

		userID := authctx.UserID(request.Context())

		shortKey, err := GenerateShortKey(string(body), request.Context(), writer, usecase)
		if err != nil {
			var conflictErr *model.URLConflictError
			if errors.As(err, &conflictErr) {
				ok := sendConflictResponse(writer, GenerateShortenURL(usecase.GetBaseURL(), conflictErr.ExistingShortKey, request))
				if ok && auditor != nil {
					sendAuditEvent(auditor, request, userID, string(body))
				}
				return
			}
			return
		}

		ok := sendResponse(writer, GenerateShortenURL(usecase.GetBaseURL(), shortKey, request))
		if ok && auditor != nil {
			sendAuditEvent(auditor, request, userID, string(body))
		}
	}
}

func sendAuditEvent(auditor service.AuditorService, request *http.Request, userID string, auditURLRaw string) {
	auditURL := auditURLFromRaw(auditURLRaw)
	auditor.AuditNewEvent(request.Context(), model.AuditEvent{
		Timestamp: time.Now().Unix(),
		Action:    "shorten",
		UserID:    userID,
		URL:       auditURL,
	})
}

func sendResponse(writer http.ResponseWriter, shortURL string) bool {
	writer.Header().Set(common.ContentType, common.TextPlain)
	writer.WriteHeader(http.StatusCreated)
	_, err := writer.Write([]byte(shortURL))
	if err != nil {
		common.InternalError(writer, err)
		return false
	}
	return true
}

func sendConflictResponse(writer http.ResponseWriter, shortURL string) bool {
	writer.Header().Set(common.ContentType, common.TextPlain)
	writer.WriteHeader(http.StatusConflict)
	_, err := writer.Write([]byte(shortURL))
	if err != nil {
		common.InternalError(writer, err)
		return false
	}
	return true
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
		return nil, fmt.Errorf("incorrect URL: %s", parsedURL)
	}

	return parsedURL, nil
}

func isTextPlain(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return mediaType == common.TextPlain
}

func auditURLFromRaw(raw string) string {
	parsed, err := parseURL(raw)
	if err != nil {
		return common.NormalizeURL(raw)
	}
	return common.NormalizeURL(parsed.String())
}
