package resolver

import (
	"errors"
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
	"github.com/go-chi/chi/v5"
)

func Handler(usecase usecase.LinkReadUsecase) http.HandlerFunc {
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
	}
}
