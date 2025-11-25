package rest

import (
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase/link"
	"github.com/rs/zerolog/log"
)

func Start(linkUsecase *link.Usecase, serverAddress string) {
	router := NewRouter(linkUsecase)
	if err := http.ListenAndServe(serverAddress, router); err != nil {
		log.Fatal().Err(err).Msg("Http server start error")
	}
}
