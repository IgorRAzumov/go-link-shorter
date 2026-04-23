package rest

import (
	"net"
	"net/http"
	_ "net/http/pprof"

	_ "github.com/IgorRAzumov/go-link-shorter/docs"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/healthcheck"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/resolver"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/shorter"
	statshandler "github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/stats"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/middleware"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/middleware/auth"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/middleware/gzip"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/middleware/trustedsubnet"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

// RouterDeps — зависимости HTTP-роутера сервиса сокращения ссылок.
type RouterDeps struct {
	LinkCreateUsecase usecase.LinkCreateUsecase
	LinkReadUsecase   usecase.LinkReadUsecase
	LinkDeleteUsecase usecase.LinkDeleteUsecase
	HealthCheck       usecase.HealthCheckUsecase
	StatsUsecase      usecase.StatsUsecase
	AuthService       service.AuthService
	Auditor           service.AuditorService
	EnablePprof       bool
	TrustedSubnet     *net.IPNet
}

// NewRouter создаёт HTTP-роутер со всеми эндпоинтами сервиса сокращения ссылок.
func NewRouter(deps RouterDeps) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.HTTPLogger, gzip.GZIP, auth.Middleware(deps.AuthService))
	router.Post("/", shorter.Handler(deps.LinkCreateUsecase, deps.Auditor))
	router.Get("/api/user/urls", shorter.UserURLsHandler(deps.LinkReadUsecase))
	router.Delete("/api/user/urls", shorter.UserURLsDeleteHandler(deps.LinkDeleteUsecase))
	router.Post("/api/shorten", shorter.APIHandler(deps.LinkCreateUsecase, deps.Auditor))
	router.Post("/api/shorten/batch", shorter.BatchAPIHandler(deps.LinkCreateUsecase))
	router.Get("/{shortKey}", resolver.Handler(deps.LinkReadUsecase, deps.Auditor))

	router.Group(func(r chi.Router) {
		r.Use(trustedsubnet.Middleware(deps.TrustedSubnet))
		r.Get("/api/internal/stats", statshandler.StatisticHandler(deps.StatsUsecase))
	})

	if deps.EnablePprof {
		router.Handle("/debug/pprof/*", http.DefaultServeMux)
		router.Handle("/debug/pprof", http.RedirectHandler("/debug/pprof/", http.StatusMovedPermanently))
	}

	router.Get("/ping", healthcheck.Handler(deps.HealthCheck))
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	return router
}
