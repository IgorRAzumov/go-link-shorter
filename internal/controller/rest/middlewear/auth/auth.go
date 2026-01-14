package auth

import (
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"github.com/rs/zerolog/log"
)

const userIDCookieName = "user_id"

func Middleware(authService service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			cookie, cookieErr := r.Cookie(userIDCookieName)
			if cookieErr != nil {
				userID := authService.GenerateUserID()
				setUserIDCookie(w, userID, authService)
				ctx = authctx.WithUserID(ctx, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			userID, err := authService.ValidateSignedUserID(cookie.Value)
			if err != nil {
				userID = authService.GenerateUserID()
				setUserIDCookie(w, userID, authService)
				ctx = authctx.WithUserID(ctx, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if userID == "" {
				userID = authService.GenerateUserID()
				setUserIDCookie(w, userID, authService)
				if r.Method == http.MethodGet && r.URL != nil && r.URL.Path == "/api/user/urls" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			}

			ctx = authctx.WithUserID(ctx, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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
