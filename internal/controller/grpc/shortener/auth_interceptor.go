package shortenergrpc

import (
	"context"
	"strings"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// authorizationMDKey — имя метаданных с авторизационным токеном (регистр не важен, grpc нормализует).
const authorizationMDKey = "authorization"

// AuthUnaryInterceptor проверяет metadata "authorization":
//   - если там валидный signed userID — запрос считается аутентифицированным,
//     userID кладётся в контекст (authctx.WithUserID + WithAuthenticated);
//   - иначе генерируется анонимный userID, клиенту в header'ах ответа отдаётся свежий токен.
func AuthUnaryInterceptor(authService service.AuthService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if userID, ok := userIDFromMetadata(ctx, authService); ok {
			ctx = authctx.WithAuthenticated(authctx.WithUserID(ctx, userID))
			return handler(ctx, req)
		}

		userID := authService.GenerateUserID()
		signed := authService.SignUserID(userID)
		_ = grpc.SetHeader(ctx, metadata.Pairs(authorizationMDKey, signed))
		ctx = authctx.WithUserID(ctx, userID)
		return handler(ctx, req)
	}
}

func userIDFromMetadata(ctx context.Context, authService service.AuthService) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}
	values := md.Get(authorizationMDKey)
	if len(values) == 0 {
		return "", false
	}
	token := strings.TrimSpace(values[0])
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[len("bearer "):])
	}
	if token == "" {
		return "", false
	}
	userID, err := authService.ValidateSignedUserID(token)
	if err != nil || userID == "" {
		return "", false
	}
	return userID, true
}
