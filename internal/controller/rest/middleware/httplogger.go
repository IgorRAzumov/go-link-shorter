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
			Str("uri", r.RequestURI).
			Str("remote_addr", r.RemoteAddr).
			Msg("REQUEST")

		start := time.Now()
		responseLogger := &httpLogger{
			ResponseWriter: w,
			status:         http.StatusOK,
			size:           0,
		}

		defer func() {
			duration := time.Since(start)
			log.Info().
				Int("status", responseLogger.status).
				Int("size", responseLogger.size).
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
	size   int
}

func (logger *httpLogger) WriteHeader(status int) {
	logger.status = status
	logger.ResponseWriter.WriteHeader(status)
}

func (logger *httpLogger) Write(b []byte) (int, error) {
	size, err := logger.ResponseWriter.Write(b)
	logger.size += size
	return size, err
}
