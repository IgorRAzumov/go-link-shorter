package shorter

import (
	"encoding/json"
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
)

type UserURLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func UserURLsHandler(usecase usecase.LinkReadUsecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		userID := authctx.UserID(request.Context())
		if userID == "" {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		links, err := usecase.GetUserURLs(request.Context())
		if err != nil {
			common.InternalError(writer, err)
			return
		}

		if len(links) == 0 {
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		responses := make([]UserURLResponse, 0, len(links))
		for _, link := range links {
			shortURL := GenerateShortenURL(usecase.GetBaseURL(), link.ShortKey, request)
			responses = append(responses, UserURLResponse{
				ShortURL:    shortURL,
				OriginalURL: link.FullURL,
			})
		}

		responseBytes, jsonErr := json.Marshal(responses)
		if jsonErr != nil {
			common.InternalError(writer, jsonErr)
			return
		}

		writer.Header().Set(common.ContentType, common.ApplicationJSON)
		writer.WriteHeader(http.StatusOK)
		_, writeError := writer.Write(responseBytes)
		if writeError != nil {
			common.InternalError(writer, writeError)
		}
	}
}
