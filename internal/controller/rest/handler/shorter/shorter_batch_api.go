package shorter

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/shorter/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
)

func BatchAPIHandler(usecase usecase.LinkCreateUsecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if !IsApplicationJSON(request.Header.Get(common.ContentType)) {
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

		scheme := "http"
		if request.TLS != nil || request.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := request.Host

		batchResponses, err := usecase.ProcessBatchShortenRequests(request.Context(), batchRequests, scheme, host)
		if err != nil {
			common.InternalError(writer, err)
			return
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

func IsApplicationJSON(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return mediaType == common.ApplicationJSON
}
