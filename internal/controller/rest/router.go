package rest

import (
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/heath_check"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/resolver"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/shorter"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/middlewear"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/middlewear/gzip"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
	"github.com/go-chi/chi/v5"
)

func NewRouter(linkUsecase usecase.LinkUsecase, healthCheck usecase.HealthCheckUsecase) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.HTTPLogger, gzip.GZIP)
	router.Post("/", shorter.Handler(linkUsecase))
	router.Post("/api/shorten", shorter.APIHandler(linkUsecase))
	router.Get("/{shortKey}", resolver.Handler(linkUsecase))
	router.Get("/ping", heath_check.Handler(healthCheck))
	return router
}
