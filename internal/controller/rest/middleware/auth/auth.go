// Package auth реализует middleware HTTP-аутентификации на базе подписанной cookie user_id.
package auth

import (
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"github.com/rs/zerolog/log"
)

const userIDCookieName = "user_id"

// Middleware возвращает HTTP middleware, кладущее в контекст userID из подписанной cookie.
func Middleware(authService service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if userID, ok := userIDFromCookie(r, authService); ok {
				ctx = authctx.WithAuthenticated(authctx.WithUserID(ctx, userID))
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			userID := authService.GenerateUserID()
			setUserIDCookie(w, userID, authService)
			ctx = authctx.WithUserID(ctx, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func userIDFromCookie(r *http.Request, authService service.AuthService) (string, bool) {
	cookie, err := r.Cookie(userIDCookieName)
	if err != nil {
		return "", false
	}
	userID, err := authService.ValidateSignedUserID(cookie.Value)
	if err != nil || userID == "" {
		return "", false
	}
	return userID, true
}

func setUserIDCookie(w http.ResponseWriter, userID string, authService service.AuthService) {
	value := authService.SignUserID(userID)

	cookie := &http.Cookie{
		Name:     userIDCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
	log.Debug().Str("user_id", userID).Msg("Set user ID cookie")
}
