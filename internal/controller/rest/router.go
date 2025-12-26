package rest

import (
	"net/http"

	resolver "github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/resolver"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/shorter"
	middleware "github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/middlewear"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/middlewear/gzip"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase/link"
	"github.com/go-chi/chi/v5"
)

func NewRouter(usecase *link.Usecase) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.HTTPLogger, gzip.GZIP)
	router.Post("/", shorter.Handler(usecase))
	router.Post("/api/shorten", shorter.APIHandler(usecase))
	router.Get("/{shortKey}", resolver.Handler(usecase))
	return router
}
