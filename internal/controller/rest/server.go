package rest

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
	"github.com/rs/zerolog/log"
)

// Builder — билдер HTTP-сервера с dependency injection.
type Builder struct {
	ctx                context.Context
	linkCreateUsecase  usecase.LinkCreateUsecase
	linkReadUsecase    usecase.LinkReadUsecase
	linkDeleteUsecase  usecase.LinkDeleteUsecase
	healthCheckUsecase usecase.HealthCheckUsecase
	authService        service.AuthService
	auditor            service.AuditorService
	serverAddress      string
	enablePprof        bool
}

// NewServerBuilder создаёт билдер сервера.
func NewServerBuilder() *Builder { return &Builder{} }

// WithContext задаёт контекст для graceful shutdown.
func (builder *Builder) WithContext(ctx context.Context) *Builder {
	builder.ctx = ctx
	return builder
}

// WithLinkCreateUsecase задаёт use case создания ссылок.
func (builder *Builder) WithLinkCreateUsecase(usecase usecase.LinkCreateUsecase) *Builder {
	builder.linkCreateUsecase = usecase
	return builder
}

// WithLinkReadUsecase задаёт use case чтения ссылок.
func (builder *Builder) WithLinkReadUsecase(usecase usecase.LinkReadUsecase) *Builder {
	builder.linkReadUsecase = usecase
	return builder
}

// WithLinkDeleteUsecase задаёт use case удаления ссылок.
func (builder *Builder) WithLinkDeleteUsecase(usecase usecase.LinkDeleteUsecase) *Builder {
	builder.linkDeleteUsecase = usecase
	return builder
}

// WithHealthCheckUsecase задаёт use case проверки здоровья.
func (builder *Builder) WithHealthCheckUsecase(usecase usecase.HealthCheckUsecase) *Builder {
	builder.healthCheckUsecase = usecase
	return builder
}

// WithAuthService задаёт сервис аутентификации.
func (builder *Builder) WithAuthService(service service.AuthService) *Builder {
	builder.authService = service
	return builder
}

// WithAuditor задаёт сервис аудита.
func (builder *Builder) WithAuditor(auditor service.AuditorService) *Builder {
	builder.auditor = auditor
	return builder
}

// WithServerAddress задаёт адрес сервера.
func (builder *Builder) WithServerAddress(addr string) *Builder {
	builder.serverAddress = addr
	return builder
}

// WithEnablePprof включает эндпоинты pprof.
func (builder *Builder) WithEnablePprof(enable bool) *Builder {
	builder.enablePprof = enable
	return builder
}

// Start запускает HTTP-сервер. Блокируется до остановки сервера.
func (builder *Builder) Start() error {
	router := NewRouter(
		builder.linkCreateUsecase,
		builder.linkReadUsecase,
		builder.linkDeleteUsecase,
		builder.healthCheckUsecase,
		builder.authService,
		builder.auditor,
		builder.enablePprof,
	)

	server := &http.Server{
		Addr:    builder.serverAddress,
		Handler: router,
	}

	go func() {
		<-builder.ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("Http server shutdown error")
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
