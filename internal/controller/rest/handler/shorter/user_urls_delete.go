package shorter

import (
	"encoding/json"
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
)

// UserURLsDeleteHandler возвращает обработчик DELETE /api/user/urls — мягкое удаление ссылок пользователя.
//
// @Summary      Delete user URLs
// @Description  Мягкое удаление ссылок по short_key
// @Tags         user
// @Accept       json
// @Param        short_keys  body  []string  true  "Список short_key для удаления"
// @Success      202  {string}  string  "Принято"
// @Failure      401  {string}  string  "Не авторизован"
// @Router       /api/user/urls [delete]
func UserURLsDeleteHandler(usecase usecase.LinkDeleteUsecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		userID := authctx.UserID(request.Context())
		if userID == "" {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		var shortKeys []string
		if err := json.NewDecoder(request.Body).Decode(&shortKeys); err != nil {
			common.BadRequestError(writer, err)
			return
		}

		if err := usecase.DeleteUserURLs(request.Context(), shortKeys); err != nil {
			common.InternalError(writer, err)
			return
		}

		writer.WriteHeader(http.StatusAccepted)
	}
}
