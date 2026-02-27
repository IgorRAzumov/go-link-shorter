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

type Builder struct {
	ctx                context.Context
	linkCreateUsecase  usecase.LinkCreateUsecase
	linkReadUsecase    usecase.LinkReadUsecase
	linkDeleteUsecase  usecase.LinkDeleteUsecase
	healthCheckUsecase usecase.HealthCheckUsecase
	authService        service.AuthService
	auditor            service.AuditorService
	serverAddress      string
}

func NewServerBuilder() *Builder { return &Builder{} }

func (builder *Builder) WithContext(ctx context.Context) *Builder {
	builder.ctx = ctx
	return builder
}

func (builder *Builder) WithLinkCreateUsecase(usecase usecase.LinkCreateUsecase) *Builder {
	builder.linkCreateUsecase = usecase
	return builder
}

func (builder *Builder) WithLinkReadUsecase(usecase usecase.LinkReadUsecase) *Builder {
	builder.linkReadUsecase = usecase
	return builder
}

func (builder *Builder) WithLinkDeleteUsecase(usecase usecase.LinkDeleteUsecase) *Builder {
	builder.linkDeleteUsecase = usecase
	return builder
}

func (builder *Builder) WithHealthCheckUsecase(usecase usecase.HealthCheckUsecase) *Builder {
	builder.healthCheckUsecase = usecase
	return builder
}

func (builder *Builder) WithAuthService(service service.AuthService) *Builder {
	builder.authService = service
	return builder
}

func (builder *Builder) WithAuditor(auditor service.AuditorService) *Builder {
	builder.auditor = auditor
	return builder
}

func (builder *Builder) WithServerAddress(addr string) *Builder {
	builder.serverAddress = addr
	return builder
}

func (builder *Builder) Start() {
	router := NewRouter(
		builder.linkCreateUsecase,
		builder.linkReadUsecase,
		builder.linkDeleteUsecase,
		builder.healthCheckUsecase,
		builder.authService,
		builder.auditor,
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
		log.Fatal().Err(err).Msg("Http server start error")
	}
}
