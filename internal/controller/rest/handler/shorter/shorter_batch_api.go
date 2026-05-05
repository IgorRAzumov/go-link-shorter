package shorter

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	handlermodel "github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/shorter/model"
	domainmodel "github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
	"github.com/IgorRAzumov/go-link-shorter/pkg/urlutil"
)

// BatchAPIHandler возвращает обработчик POST /api/shorten/batch — пакетное сокращение URL.
//
// @Summary      Batch shorten URLs
// @Description  Пакетное сокращение нескольких URL
// @Tags         shorter
// @Accept       json
// @Produce      json
// @Param        request  body      []handlermodel.BatchShortenRequest  true  "Список URL"
// @Success      201      {array}   handlermodel.BatchShortenResponse
// @Failure      400      {string}  string  "Некорректный запрос"
// @Router       /api/shorten/batch [post]
func BatchAPIHandler(usecase usecase.LinkCreateUsecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if !IsApplicationJSON(request.Header.Get(common.ContentType)) {
			common.MethodNotAllowedError(writer)
			return
		}

		var batchRequests []handlermodel.BatchShortenRequest
		dec := json.NewDecoder(request.Body)
		if err := dec.Decode(&batchRequests); err != nil {
			common.BadRequestError(writer, err)
			return
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(request.Body)

		if len(batchRequests) == 0 {
			common.BadRequestError(writer, nil)
			return
		}

		domainRequests := createDomainRequests(batchRequests)
		batchResults, err := usecase.ProcessBatchShortenRequests(request.Context(), domainRequests)
		if err != nil {
			common.InternalError(writer, err)
			return
		}

		scheme := "http"
		if request.TLS != nil || request.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		baseURL := usecase.GetBaseURL()
		batchResponses := createResponse(batchResults, baseURL, scheme, request)

		responseBytes, jsonErr := json.Marshal(batchResponses)
		if jsonErr != nil {
			common.InternalError(writer, jsonErr)
			return
		}

		writer.Header().Set(common.ContentType, common.ApplicationJSON)
		writer.WriteHeader(http.StatusCreated)
		_, writeError := writer.Write(responseBytes)
		if writeError != nil {
			common.InternalError(writer, writeError)
		}
	}
}

// IsApplicationJSON проверяет, что Content-Type — application/json.
func IsApplicationJSON(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return mediaType == common.ApplicationJSON
}

func createDomainRequests(batchRequests []handlermodel.BatchShortenRequest) []domainmodel.BatchShortenRequest {
	domainRequests := make([]domainmodel.BatchShortenRequest, 0, len(batchRequests))
	for _, req := range batchRequests {
		domainRequests = append(domainRequests, domainmodel.BatchShortenRequest{
			CorrelationID: req.CorrelationID,
			OriginalURL:   req.OriginalURL,
		})
	}
	return domainRequests
}

func createResponse(batchResults []domainmodel.BatchShortenResult, baseURL string, scheme string, request *http.Request) []handlermodel.BatchShortenResponse {
	batchResponses := make([]handlermodel.BatchShortenResponse, 0, len(batchResults))
	for _, r := range batchResults {
		batchResponses = append(batchResponses, handlermodel.BatchShortenResponse{
			CorrelationID: r.CorrelationID,
			ShortURL:      urlutil.BuildShortURL(baseURL, scheme, request.Host, r.ShortKey),
		})
	}
	return batchResponses
}
