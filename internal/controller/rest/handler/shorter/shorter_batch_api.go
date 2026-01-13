package shorter

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/shorter/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
)

func BatchAPIHandler(usecase usecase.LinkUsecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get(common.ContentType) != common.ApplicationJSON {
			common.MethodNotAllowedError(writer)
			return
		}

		var batchRequests []model.BatchShortenRequest
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

		urls := make([]string, 0, len(batchRequests))
		urlToCorrelationMap := make(map[string]string)

		for _, req := range batchRequests {
			normalizedURL := common.NormalizeURL(req.OriginalURL)
			urls = append(urls, normalizedURL)
			urlToCorrelationMap[normalizedURL] = req.CorrelationID
		}

		shortKeysMap, err := usecase.CreateShortKeysBatch(request.Context(), urls)
		if err != nil {
			common.InternalError(writer, err)
			return
		}

		batchResponses := make([]model.BatchShortenResponse, 0, len(batchRequests))
		for _, req := range batchRequests {
			normalizedURL := common.NormalizeURL(req.OriginalURL)
			shortKey := shortKeysMap[normalizedURL]
			if shortKey != "" {
				batchResponses = append(batchResponses, model.BatchShortenResponse{
					CorrelationID: req.CorrelationID,
					ShortURL:      GenerateShortenURL(usecase.GetBaseURL(), shortKey, request),
				})
			}
		}

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
