package app

import (
	"github.com/IgorRAzumov/go-link-shorter/internal/adapter/inmemory"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/repository"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/resolver"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/shorter"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase/link"
	"github.com/rs/zerolog/log"
)

func Run(serverAddress, baseURL, fileStoragePath string) {
	storage := initStorage(fileStoragePath)
	shorterService := shorter.NewShorterService(storage)
	resolverService := resolver.NewResolverService(storage)
	linkUsecase := link.NewLinkUsecase(resolverService, shorterService, baseURL)

	rest.Start(linkUsecase, serverAddress)
}

func initStorage(fileStoragePath string) repository.LinkRepository {
	fileStorage, err := inmemory.NewFileStorage(fileStoragePath)
	if err != nil {
		log.Fatal().Err(err).Str("file_path", fileStoragePath).Msg("Failed to initialize file storage")
	}
	log.Info().Str("file_path", fileStoragePath).Msg("Using file storage")
	return fileStorage
}
