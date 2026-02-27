package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	auditadapter "github.com/IgorRAzumov/go-link-shorter/internal/adapter/audit"
	"github.com/IgorRAzumov/go-link-shorter/internal/adapter/database"
	"github.com/IgorRAzumov/go-link-shorter/internal/adapter/inmemory"
	"github.com/IgorRAzumov/go-link-shorter/internal/config"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/repository"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/audit"
	authservice "github.com/IgorRAzumov/go-link-shorter/internal/domain/service/auth"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/deleter"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/healthcheck"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/resolver"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/shorter"
	healthcheckusecase "github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase/healthcheck"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase/link"
	"github.com/rs/zerolog/log"
)

func Run(config *config.Config) {
	linkRepository := initStorage(config.DatabaseAddress, config.FileStoragePath)

	resolverService := resolver.NewResolverService(linkRepository)
	authService := authservice.NewAuthService(config.SecretKey)
	auditorService := initAuditorService(config.AuditFile, config.AuditURL)

	deleteService := deleter.NewDeleterService(linkRepository, deleter.DefaultMaxBatch)
	deleteService.Start()

	var dataBaseStorage *database.Storage
	if dbStorage, ok := linkRepository.(*database.Storage); ok {
		dataBaseStorage = dbStorage
	} else {
		dataBaseStorage = &database.Storage{}
	}

	healthCheckService := healthcheck.NewHealthCheckService(dataBaseStorage)
	healthCheckUsecase := healthcheckusecase.NewHealthCheckUsecase(healthCheckService)

	linkCreateUsecase := link.NewLinkCreateUsecase(resolverService, shorter.NewShorterService(linkRepository), config.BaseShortURL)
	linkReadUsecase := link.NewLinkReadUsecase(resolverService, config.BaseShortURL)
	linkDeleteUsecase := link.NewLinkDeleteUsecase(deleteService)

	serverCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rest.NewServerBuilder().
		WithContext(serverCtx).
		WithLinkCreateUsecase(linkCreateUsecase).
		WithLinkReadUsecase(linkReadUsecase).
		WithLinkDeleteUsecase(linkDeleteUsecase).
		WithHealthCheckUsecase(healthCheckUsecase).
		WithAuthService(authService).
		WithAuditor(auditorService).
		WithServerAddress(config.ServerAddress).
		Start()

	deleteService.Stop()
	deleteService.Wait()
}

func initAuditorService(auditFile, auditURL string) service.AuditorService {
	if auditFile == "" && auditURL == "" {
		return nil
	}
	publisher := audit.NewAuditService()
	if auditFile != "" {
		publisher.Register(auditadapter.NewFileObserver(auditFile))
	}
	if auditURL != "" {
		publisher.Register(auditadapter.NewHTTPAuditObserver(auditURL))
	}
	return publisher
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
