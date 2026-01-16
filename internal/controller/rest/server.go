package rest

import (
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
	"github.com/rs/zerolog/log"
)

func Start(linkCreateUsecase usecase.LinkCreateUsecase, linkReadUsecase usecase.LinkReadUsecase, linkDeleteUsecase usecase.LinkDeleteUsecase, healthCheckUsecase usecase.HealthCheckUsecase, authService service.AuthService, serverAddress string) {
	router := NewRouter(linkCreateUsecase, linkReadUsecase, linkDeleteUsecase, healthCheckUsecase, authService)
	if err := http.ListenAndServe(serverAddress, router); err != nil {
		log.Fatal().Err(err).Msg("Http server start error")
	}
}
