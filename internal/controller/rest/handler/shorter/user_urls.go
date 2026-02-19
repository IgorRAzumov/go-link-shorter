package shorter

import (
	"encoding/json"
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
)

// UserURLResponse — элемент ответа GET /api/user/urls.
type UserURLResponse struct {
	ShortURL    string `json:"short_url"`    // Сокращённый URL
	OriginalURL string `json:"original_url"` // Исходный URL
}

// UserURLsHandler возвращает обработчик GET /api/user/urls — список ссылок текущего пользователя.
//
// @Summary      Get user URLs
// @Description  Список сокращённых ссылок текущего пользователя
// @Tags         user
// @Produce      json
// @Success      200   {array}   UserURLResponse
// @Success      204   {string}  string  "Нет ссылок"
// @Failure      401   {string}  string  "Не авторизован"
// @Router       /api/user/urls [get]
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
