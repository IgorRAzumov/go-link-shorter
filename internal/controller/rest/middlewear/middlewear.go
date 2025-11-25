package middleware

import (
	"log"
	"net/http"
	"time"
)

func HTTPLogger(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[REQUEST] %s %s %s",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
		)

		log.Printf("[REQUEST] %s",
			r.RequestURI,
		)

		start := time.Now()
		responseLogger := &httpLogger{w, http.StatusOK}

		defer func() {
			duration := time.Since(start)
			log.Printf("[RESPONSE] %d %s %s %v",
				responseLogger.status,
				r.Method,
				r.URL.Path,
				duration,
			)
		}()

		next.ServeHTTP(responseLogger, r)
	}
}

type httpLogger struct {
	http.ResponseWriter
	status int
}

func (logger *httpLogger) WriteHeader(status int) {
	logger.status = status
	logger.ResponseWriter.WriteHeader(status)
}
