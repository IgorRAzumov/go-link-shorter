package rest

import (
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
	"github.com/rs/zerolog/log"
)

func Start(linkUsecase usecase.LinkUsecase, healthCheckUsecase usecase.HealthCheckUsecase, serverAddress string) {
	router := NewRouter(linkUsecase, healthCheckUsecase)
	if err := http.ListenAndServe(serverAddress, router); err != nil {
		log.Fatal().Err(err).Msg("Http server start error")
	}
}
