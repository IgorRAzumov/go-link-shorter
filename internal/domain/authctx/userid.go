package authctx

import "context"

type userIDKey struct{}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

func UserID(ctx context.Context) string {
	userID, ok := ctx.Value(userIDKey{}).(string)
	if !ok {
		return ""
	}
	return userID
}
