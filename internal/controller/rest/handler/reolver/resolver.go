package reolver

import (
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
)

func Handler(usecase usecase.LinkUsecase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "" || len(request.URL.Path) < 2 {
			common.BadRequestError(writer, nil)
			return
		}

		shortKey := request.URL.Path[1:]
		fullLink, err := usecase.GetFullURLByShorKey(shortKey)
		if err != nil || fullLink == "" {
			common.BadRequestError(writer, err)
			return
		}

		writer.Header().Set("Location", fullLink)
		writer.WriteHeader(http.StatusTemporaryRedirect)
	}
}
