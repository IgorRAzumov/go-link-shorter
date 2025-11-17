package reolver

import (
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
	"github.com/go-chi/chi/v5"
)

func Handler(usecase usecase.LinkUsecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		shortKey := chi.URLParam(request, "shortKey")
		if shortKey == "" {
			common.BadRequestError(writer, nil)
			return
		}

		fullLink, err := usecase.GetFullURLByShorKey(shortKey)
		if err != nil || fullLink == "" {
			common.BadRequestError(writer, err)
			return
		}

		writer.Header().Set("Location", fullLink)
		writer.WriteHeader(http.StatusTemporaryRedirect)
	}
}
