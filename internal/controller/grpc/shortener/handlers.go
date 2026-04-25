package shortenergrpc

import (
	"context"
	"errors"

	"github.com/IgorRAzumov/go-link-shorter/api/shortenerpb"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/pkg/urlutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ShortenURL — gRPC-аналог POST /api/shorten. При конфликте возвращает result
// с уже существующим коротким ключом и код codes.AlreadyExists — клиент может
// отличить повторную отправку по коду статуса, но получает рабочий URL.
func (server *Server) ShortenURL(ctx context.Context, in *shortenerpb.URLShortenRequest) (*shortenerpb.URLShortenResponse, error) {
	raw := in.GetUrl()
	normalized, err := urlutil.ValidateURL(raw)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "url is required and must be absolute")
	}

	shortKey, err := server.linkCreateUsecase.CreateShortKey(ctx, normalized)
	if err != nil {
		var conflictErr *model.URLConflictError
		if errors.As(err, &conflictErr) {
			resp := &shortenerpb.URLShortenResponse{Result: server.buildShortURL(conflictErr.ExistingShortKey)}
			return resp, status.Error(codes.AlreadyExists, "url already shortened")
		}
		return nil, status.Error(codes.Internal, "shorten url")
	}

	return &shortenerpb.URLShortenResponse{Result: server.buildShortURL(shortKey)}, nil
}

// ExpandURL — gRPC-аналог GET /{shortKey}. Возвращает NotFound, если ключ не найден,
// и FailedPrecondition, если ссылка помечена как удалённая.
func (server *Server) ExpandURL(ctx context.Context, in *shortenerpb.URLExpandRequest) (*shortenerpb.URLExpandResponse, error) {
	id := in.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	fullURL, err := server.linkReadUsecase.GetFullURLByShorKey(ctx, id)
	switch {
	case errors.Is(err, model.ErrURLDeleted):
		return nil, status.Error(codes.FailedPrecondition, "url is deleted")
	case err != nil || fullURL == "":
		return nil, status.Error(codes.NotFound, "url not found")
	}

	return &shortenerpb.URLExpandResponse{Result: fullURL}, nil
}

// ListUserURLs — gRPC-аналог GET /api/user/urls. Требует аутентифицированный контекст:
func (server *Server) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*shortenerpb.UserURLsResponse, error) {
	if !authctx.IsAuthenticated(ctx) {
		return nil, status.Error(codes.Unauthenticated, "authorization metadata required")
	}

	links, err := server.linkReadUsecase.GetUserURLs(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "list user urls")
	}

	response := &shortenerpb.UserURLsResponse{Url: make([]*shortenerpb.URLData, 0, len(links))}
	for _, l := range links {
		response.Url = append(response.Url, &shortenerpb.URLData{
			ShortUrl:    server.buildShortURL(l.ShortKey),
			OriginalUrl: l.FullURL,
		})
	}
	return response, nil
}

// buildShortURL формирует полный сокращённый URL из настроенного baseURL.
func (server *Server) buildShortURL(shortKey string) string {
	return urlutil.BuildShortURL(server.baseURL, "", "", shortKey)
}
