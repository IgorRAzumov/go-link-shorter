package app

import (
	"context"
	"fmt"
	"net"
	"os/signal"
	"syscall"

	auditadapter "github.com/IgorRAzumov/go-link-shorter/internal/adapter/audit"
	"github.com/IgorRAzumov/go-link-shorter/internal/adapter/database"
	"github.com/IgorRAzumov/go-link-shorter/internal/adapter/inmemory"
	"github.com/IgorRAzumov/go-link-shorter/internal/config"
	shortenergrpc "github.com/IgorRAzumov/go-link-shorter/internal/controller/grpc/shortener"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/repository"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/audit"
	authservice "github.com/IgorRAzumov/go-link-shorter/internal/domain/service/auth"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/deleter"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/healthcheck"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/resolver"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/shorter"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/stats"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
	healthcheckusecase "github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase/healthcheck"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase/link"
	statisticusecase "github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase/statistic"
	"github.com/rs/zerolog/log"
)

func Run(config *config.Config) error {
	linkRepository, err := initStorage(config.DatabaseAddress, config.FileStoragePath)
	if err != nil {
		return err
	}

	trustedSubnet, err := parseTrustedSubnet(config.TrustedSubnet)
	if err != nil {
		return err
	}

	resolverService := resolver.NewResolverService(linkRepository)
	authService := authservice.NewAuthService(config.SecretKey)
	auditorService, closeAudit, err := initAuditorService(config.AuditFile, config.AuditURL)
	if err != nil {
		return err
	}
	if auditorService != nil && closeAudit != nil {
		defer func() {
			if closeErr := closeAudit(); closeErr != nil {
				log.Error().Err(closeErr).Msg("Failed to close audit file")
			}
		}()
	}

	deleteService := deleter.NewDeleterService(linkRepository, deleter.DefaultMaxBatch)
	deleteService.Start()

	var dataBaseStorage *database.Storage
	if dbStorage, ok := linkRepository.(*database.Storage); ok {
		dataBaseStorage = dbStorage
	} else {
		dataBaseStorage = &database.Storage{}
	}

	defer func() {
		if closeErr := dataBaseStorage.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("Failed to close database storage")
		}
	}()

	healthCheckService := healthcheck.NewHealthCheckService(dataBaseStorage)
	healthCheckUsecase := healthcheckusecase.NewHealthCheckUsecase(healthCheckService)

	linkCreateUsecase := link.NewLinkCreateUsecase(resolverService, shorter.NewShorterService(linkRepository), config.BaseShortURL)
	linkReadUsecase := link.NewLinkReadUsecase(resolverService, config.BaseShortURL)
	linkDeleteUsecase := link.NewLinkDeleteUsecase(deleteService)

	statsRepo, ok := linkRepository.(repository.StatsRepository)
	if !ok {
		return fmt.Errorf("link repository does not implement StatsRepository")
	}
	statsService := stats.NewService(statsRepo)
	statsUsecase := statisticusecase.NewUsecase(statsService)

	serverCtx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	grpcErrCh, err := startGrpc(serverCtx, config, authService, linkCreateUsecase, linkReadUsecase)
	if err != nil {
		return err
	}
	go func() {
		if grpcErr := <-grpcErrCh; grpcErr != nil {
			log.Error().Err(grpcErr).Msg("gRPC server failed, shutting down")
			stop()
		}
	}()

	if serverError := rest.NewServerBuilder().
		WithContext(serverCtx).
		WithLinkCreateUsecase(linkCreateUsecase).
		WithLinkReadUsecase(linkReadUsecase).
		WithLinkDeleteUsecase(linkDeleteUsecase).
		WithHealthCheckUsecase(healthCheckUsecase).
		WithStatsUsecase(statsUsecase).
		WithAuthService(authService).
		WithAuditor(auditorService).
		WithServerAddress(config.ServerAddress).
		WithEnablePprof(config.EnablePprof).
		WithEnableHTTPS(config.EnableHTTPS, config.TLSCertFile, config.TLSKeyFile).
		WithTrustedSubnet(trustedSubnet).
		Start(); serverError != nil {
		return serverError
	}

	deleteService.Stop()
	deleteService.Wait()
	return nil
}

func startGrpc(
	ctx context.Context,
	config *config.Config,
	authService service.AuthService,
	linkCreateUsecase usecase.LinkCreateUsecase,
	linkReadUsecase usecase.LinkReadUsecase,
) (<-chan error, error) {
	grpcServer, err := shortenergrpc.NewGRPCServer(authService, shortenergrpc.TLSConfig{
		Enable:   config.EnableHTTPS,
		CertFile: config.TLSCertFile,
		KeyFile:  config.TLSKeyFile,
	})
	if err != nil {
		return nil, err
	}

	svc := shortenergrpc.NewServer(authService, linkCreateUsecase, linkReadUsecase, config.BaseShortURL)
	svc.Register(grpcServer)

	log.Info().Str("addr", config.GRPCAddress).Msg("Starting gRPC server")
	return shortenergrpc.Serve(ctx, config.GRPCAddress, grpcServer)
}

func parseTrustedSubnet(value string) (*net.IPNet, error) {
	if value == "" {
		return nil, nil
	}
	_, subnet, err := net.ParseCIDR(value)
	if err != nil {
		return nil, fmt.Errorf("invalid trusted subnet CIDR %q: %w", value, err)
	}
	return subnet, nil
}

func initAuditorService(auditFile, auditURL string) (service.AuditorService, func() error, error) {
	if auditFile == "" && auditURL == "" {
		return nil, nil, nil
	}

	var closeAudit func() error = nil
	publisher := audit.NewAuditService()

	if auditFile != "" {
		observer, closer, err := auditadapter.NewFileObserver(auditFile)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open audit file %q: %w", auditFile, err)
		}
		publisher.Register(observer)
		closeAudit = closer
	}
	if auditURL != "" {
		publisher.Register(auditadapter.NewHTTPAuditObserver(auditURL))
	}
	return publisher, closeAudit, nil
}

func initStorage(databaseDSN, fileStoragePath string) (repository.LinkRepository, error) {
	if databaseDSN != "" {
		dbStorage, err := database.NewStorage(databaseDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database storage (dsn=%q): %w", databaseDSN, err)
		}
		log.Info().Msg("Using PostgreSQL database storage")
		return dbStorage, nil
	}

	if fileStoragePath != "" {
		fileStorage, err := inmemory.NewInMemoryFileStorage(fileStoragePath)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize file storage (path=%q): %w", fileStoragePath, err)
		}
		log.Info().Str("file_path", fileStoragePath).Msg("Using file storage")
		return fileStorage, nil
	}

	memoryStorage, err := inmemory.NewInMemoryFileStorage("")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize in-memory storage: %w", err)
	}
	log.Info().Msg("Using in-memory storage")
	return memoryStorage, nil
}
