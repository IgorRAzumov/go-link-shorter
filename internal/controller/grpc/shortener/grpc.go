package shortenergrpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// TLSConfig — опции TLS для gRPC-сервера.
type TLSConfig struct {
	Enable   bool
	CertFile string
	KeyFile  string
}

// NewGRPCServer создаёт gRPC-сервер с unary-интерцептором аутентификации и опциональным TLS.
func NewGRPCServer(authService service.AuthService, tlsCfg TLSConfig) (*grpc.Server, error) {
	creds, err := buildCredentials(tlsCfg)
	if err != nil {
		return nil, err
	}
	return grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(AuthUnaryInterceptor(authService)),
	), nil
}

// Serve стартует gRPC-сервер в отдельной горутине и возвращает канал, через который
// приходят фатальные ошибки сервера (либо закрытый канал при корректной остановке).
func Serve(ctx context.Context, addr string, grpcServer *grpc.Server) (<-chan error, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("grpc listen on %q: %w", addr, err)
	}

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		if err := grpcServer.Serve(listener); err != nil {
			errCh <- fmt.Errorf("grpc serve: %w", err)
		}
	}()
	return errCh, nil
}

func buildCredentials(cfg TLSConfig) (credentials.TransportCredentials, error) {
	if !cfg.Enable {
		return insecure.NewCredentials(), nil
	}
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("grpc load tls key pair: %w", err)
	}
	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}), nil
}
