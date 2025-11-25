package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func HTTPLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Msg("REQUEST")

		start := time.Now()
		responseLogger := &httpLogger{w, http.StatusOK}

		defer func() {
			duration := time.Since(start)
			log.Info().
				Int("status", responseLogger.status).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Dur("duration", duration).
				Msg("RESPONSE")
		}()

		next.ServeHTTP(responseLogger, r)
	})
}

type httpLogger struct {
	http.ResponseWriter
	status int
}

func (logger *httpLogger) WriteHeader(status int) {
	logger.status = status
	logger.ResponseWriter.WriteHeader(status)
}
