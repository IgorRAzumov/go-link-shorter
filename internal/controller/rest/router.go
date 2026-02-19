package rest

import (
	"net/http"
	_ "net/http/pprof"

	_ "github.com/IgorRAzumov/go-link-shorter/docs"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/healthcheck"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/resolver"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/shorter"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/middleware"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/middleware/auth"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/middleware/gzip"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

// NewRouter создаёт HTTP-роутер со всеми эндпоинтами сервиса сокращения ссылок.
func NewRouter(
	linkCreateUsecase usecase.LinkCreateUsecase,
	linkReadUsecase usecase.LinkReadUsecase,
	linkDeleteUsecase usecase.LinkDeleteUsecase,
	healthCheck usecase.HealthCheckUsecase,
	authService service.AuthService,
	auditor service.AuditorService,
	enablePprof bool,
) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.HTTPLogger, gzip.GZIP, auth.Middleware(authService))
	router.Post("/", shorter.Handler(linkCreateUsecase, auditor))
	router.Get("/api/user/urls", shorter.UserURLsHandler(linkReadUsecase))
	router.Post("/api/shorten", shorter.APIHandler(linkCreateUsecase, auditor))
	router.Post("/api/shorten/batch", shorter.BatchAPIHandler(linkCreateUsecase))
	router.Delete("/api/user/urls", shorter.UserURLsDeleteHandler(linkDeleteUsecase))
	router.Get("/ping", healthcheck.Handler(healthCheck))
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	router.Get("/{shortKey}", resolver.Handler(linkReadUsecase, auditor))

	if enablePprof {
		router.Handle("/debug/pprof/*", http.DefaultServeMux)
		router.Handle("/debug/pprof", http.RedirectHandler("/debug/pprof/", http.StatusMovedPermanently))
	}
	return router
}
