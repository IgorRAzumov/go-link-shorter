package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/IgorRAzumov/go-link-shorter/internal/adapter/database"
	"github.com/IgorRAzumov/go-link-shorter/internal/adapter/inmemory"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/repository"
	authservice "github.com/IgorRAzumov/go-link-shorter/internal/domain/service/auth"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/deleter"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/healthcheck"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/resolver"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/shorter"
	healthcheckusecase "github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase/healthcheck"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase/link"
	"github.com/rs/zerolog/log"
)

func Run(serverAddress, baseURL, fileStoragePath, databaseDSN, secretKey string) {
	storage := initStorage(databaseDSN, fileStoragePath)
	resolverService := resolver.NewResolverService(storage)
	authSvc := authservice.NewAuthService(secretKey)
	deleteService := deleter.NewDeleterService(storage, deleter.DefaultMaxBatch)
	deleteService.Start()

	var dataBaseStorage *database.LinkStorage
	if dbStorage, ok := storage.(*database.LinkStorage); ok {
		dataBaseStorage = dbStorage
	} else {
		dataBaseStorage = &database.LinkStorage{}
	}
	healthCheckService := healthcheck.NewHealthCheckService(dataBaseStorage)

	healthCheckUsecase := healthcheckusecase.NewHealthCheckUsecase(healthCheckService)
	linkCreateUsecase := link.NewLinkCreateUsecase(resolverService, shorter.NewShorterService(storage), baseURL)
	linkReadUsecase := link.NewLinkReadUsecase(resolverService, baseURL)
	linkDeleteUsecase := link.NewLinkDeleteUsecase(deleteService)

	serverCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rest.Start(serverCtx, linkCreateUsecase, linkReadUsecase, linkDeleteUsecase, healthCheckUsecase, authSvc, serverAddress)

	deleteService.Stop()
	deleteService.Wait()
}

func initStorage(databaseDSN, fileStoragePath string) repository.LinkRepository {
	if databaseDSN != "" {
		dbStorage, err := database.NewStorage(databaseDSN)
		if err != nil {
			log.Fatal().Err(err).Str("dsn", databaseDSN).Msg("Failed to initialize database storage")
		}
		log.Info().Msg("Using PostgreSQL database storage")
		return dbStorage
	}

	if fileStoragePath != "" {
		fileStorage, err := inmemory.NewInMemoryFileStorage(fileStoragePath)
		if err != nil {
			log.Fatal().Err(err).Str("file_path", fileStoragePath).Msg("Failed to initialize file storage")
		}
		log.Info().Str("file_path", fileStoragePath).Msg("Using file storage")
		return fileStorage
	}

	memoryStorage, err := inmemory.NewInMemoryFileStorage("")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize in-memory storage")
	}
	log.Info().Msg("Using in-memory storage")
	return memoryStorage
}
