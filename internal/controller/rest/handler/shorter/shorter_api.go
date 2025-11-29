package shorter

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/shorter/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
)

func APIHandler(usecase usecase.LinkUsecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get(common.ContentType) != common.ApplicationJson {
			common.MethodNotAllowedError(writer)
			return
		}

		var shortenRequest model.ShortenRequest
		dec := json.NewDecoder(request.Body)
		if err := dec.Decode(&shortenRequest); err != nil {
			common.BadRequestError(writer, err)
			return
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(request.Body)

		shortKey, err := GenerateShortKey(shortenRequest.URL, request.Context(), writer, usecase)
		if err != nil {
			return
		}

		var response = &model.ShortenResponse{
			Result: GenerateShortenURL(usecase.GetBaseURL(), shortKey, request),
		}
		bytes, jsonErr := (*response).MarshalJSON()
		if jsonErr != nil {
			common.InternalError(writer, jsonErr)
			return
		}

		writer.Header().Set(common.ContentType, common.ApplicationJson)
		writer.WriteHeader(http.StatusCreated)
		_, writeError := writer.Write(bytes)
		if writeError != nil {
			common.InternalError(writer, writeError)
		}
	}
}
