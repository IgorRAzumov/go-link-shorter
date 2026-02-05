package rest

import (
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/healthcheck"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/resolver"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/shorter"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/middlewear"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/middlewear/auth"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/middlewear/gzip"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
	"github.com/go-chi/chi/v5"
)

func NewRouter(linkCreateUsecase usecase.LinkCreateUsecase, linkReadUsecase usecase.LinkReadUsecase, healthCheck usecase.HealthCheckUsecase, authService service.AuthService) http.Handler {
	router := chi.NewRouter()

	router.Use(middlewear.HTTPLogger, gzip.GZIP, auth.Middleware(authService))
	router.Post("/", shorter.Handler(linkCreateUsecase))
	router.Post("/api/shorten", shorter.APIHandler(linkCreateUsecase))
	router.Post("/api/shorten/batch", shorter.BatchAPIHandler(linkCreateUsecase))
	router.Get("/api/user/urls", shorter.UserURLsHandler(linkReadUsecase))
	router.Get("/{shortKey}", resolver.Handler(linkReadUsecase))
	router.Get("/ping", healthcheck.Handler(healthCheck))
	return router
}
