package app

import (
	"github.com/IgorRAzumov/go-link-shorter/internal/adapter/database"
	"github.com/IgorRAzumov/go-link-shorter/internal/adapter/inmemory"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/repository"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/healthcheck"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/resolver"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/shorter"
	healthcheckusecase "github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase/healthcheck"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase/link"
	"github.com/rs/zerolog/log"
)

func Run(serverAddress, baseURL, fileStoragePath, databaseDSN string) {
	storage := initStorage(fileStoragePath)
	shorterService := shorter.NewShorterService(storage)
	resolverService := resolver.NewResolverService(storage)
	linkUsecase := link.NewLinkUsecase(resolverService, shorterService, baseURL)

	dataBaseStorage, err := database.NewStorage(databaseDSN)
	if err != nil {
		log.Fatal().Err(err).Str("dsn", databaseDSN).Msg("Failed to initialize database storage")
	}
	healthCheckService := healthcheck.NewHealthCheckService(dataBaseStorage)
	healthCheckUsecase := healthcheckusecase.NewHealthCheckUsecase(healthCheckService)

	rest.Start(linkUsecase, healthCheckUsecase, serverAddress)
}

func initStorage(fileStoragePath string) repository.LinkRepository {
	fileStorage, err := inmemory.NewInMemoryFileStorage(fileStoragePath)
	if err != nil {
		log.Fatal().Err(err).Str("file_path", fileStoragePath).Msg("Failed to initialize file storage")
	}
	log.Info().Str("file_path", fileStoragePath).Msg("Using file storage")
	return fileStorage
}
