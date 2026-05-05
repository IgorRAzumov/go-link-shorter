// Package shortenergrpc реализует gRPC-фасад поверх use case'ов сервиса сокращения ссылок.
// Хендлеры (ShortenURL/ExpandURL/ListUserURLs) являются тонкими обёртками над общими
// доменными use case'ами — той же бизнес-логикой пользуется HTTP-слой.
package shortenergrpc

import (
	"github.com/IgorRAzumov/go-link-shorter/api/shortenerpb"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
	"google.golang.org/grpc"
)

// Server реализует ShortenerServiceServer, делегируя бизнес-логику use case'ам.
type Server struct {
	shortenerpb.UnimplementedShortenerServiceServer

	authService       service.AuthService
	linkCreateUsecase usecase.LinkCreateUsecase
	linkReadUsecase   usecase.LinkReadUsecase
	baseURL           string
}

// NewServer создаёт gRPC-фасад сервиса сокращения ссылок.
func NewServer(
	authService service.AuthService,
	linkCreateUsecase usecase.LinkCreateUsecase,
	linkReadUsecase usecase.LinkReadUsecase,
	baseURL string,
) *Server {
	return &Server{
		authService:       authService,
		linkCreateUsecase: linkCreateUsecase,
		linkReadUsecase:   linkReadUsecase,
		baseURL:           baseURL,
	}
}

// Register подключает сервис к gRPC-серверу.
func (server *Server) Register(grpcServer grpc.ServiceRegistrar) {
	shortenerpb.RegisterShortenerServiceServer(grpcServer, server)
}
