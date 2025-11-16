package rest

import (
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler"
	middleware "github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/middlewear"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase/link"
)

func NewRouter(usecase *link.Usecase) http.Handler {
	return middleware.HTTPLogger(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/" && request.Method == http.MethodPost {
			handler.ShortenHandler(usecase)(writer, request)
			return
		}

		if request.URL.Path != "/" && request.Method == http.MethodGet {
			handler.ResolveHandler(usecase)(writer, request)
			return
		}

		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}))
}
