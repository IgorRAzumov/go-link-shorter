package resolver

import (
	"errors"
	"net/http"
	"time"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
	"github.com/go-chi/chi/v5"
)

// Handler возвращает обработчик GET /{shortKey} — редирект на полный URL по короткому ключу.
//
// @Summary      Resolve short URL
// @Description  Редирект 307 на полный URL по короткому ключу
// @Tags         resolver
// @Param        shortKey  path  string  true  "Короткий ключ"
// @Success      307  {string}  string  "Location: полный URL"
// @Failure      400  {string}  string  "Не найден"
// @Failure      410  {string}  string  "URL удалён"
// @Router       /{shortKey} [get]
func Handler(usecase usecase.LinkReadUsecase, auditor service.AuditorService) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		shortKey := chi.URLParam(request, "shortKey")
		context := request.Context()

		fullLink, err := usecase.GetFullURLByShorKey(context, shortKey)
		if errors.Is(err, model.ErrURLDeleted) {
			writer.WriteHeader(http.StatusGone)
			return
		}
		if err != nil || fullLink == "" {
			common.BadRequestError(writer, err)
			return
		}

		writer.Header().Set("Location", fullLink)
		writer.WriteHeader(http.StatusTemporaryRedirect)

		if auditor != nil {
			sendAuditEvent(auditor, request, fullLink)
		}
	}
}

func sendAuditEvent(auditor service.AuditorService, request *http.Request, fullLink string) {
	auditor.AuditNewEvent(request.Context(), model.AuditEvent{
		Timestamp: time.Now().Unix(),
		Action:    "follow",
		UserID:    authctx.UserID(request.Context()),
		URL:       common.NormalizeURL(fullLink),
	})
}
