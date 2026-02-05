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

func Start(ctx context.Context, linkCreateUsecase usecase.LinkCreateUsecase, linkReadUsecase usecase.LinkReadUsecase, linkDeleteUsecase usecase.LinkDeleteUsecase, healthCheckUsecase usecase.HealthCheckUsecase, authService service.AuthService, serverAddress string) {
	router := NewRouter(linkCreateUsecase, linkReadUsecase, linkDeleteUsecase, healthCheckUsecase, authService)
	server := &http.Server{
		Addr:    serverAddress,
		Handler: router,
	}

	go func() {
		<-ctx.Done()
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
