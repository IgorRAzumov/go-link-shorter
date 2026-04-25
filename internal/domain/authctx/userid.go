// Package authctx хранит данные аутентификации пользователя в context.Context.
package authctx

import "context"

type userIDKey struct{}
type authenticatedKey struct{}

// WithUserID кладёт идентификатор пользователя в контекст.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// UserID достаёт идентификатор пользователя из контекста (пустая строка, если отсутствует).
func UserID(ctx context.Context) string {
	userID, ok := ctx.Value(userIDKey{}).(string)
	if !ok {
		return ""
	}
	return userID
}

// WithAuthenticated помечает запрос как аутентифицированный (валидный токен/кука клиента).
func WithAuthenticated(ctx context.Context) context.Context {
	return context.WithValue(ctx, authenticatedKey{}, true)
}

// IsAuthenticated сообщает, был ли контекст помечен как аутентифицированный.
func IsAuthenticated(ctx context.Context) bool {
	v, _ := ctx.Value(authenticatedKey{}).(bool)
	return v
}
