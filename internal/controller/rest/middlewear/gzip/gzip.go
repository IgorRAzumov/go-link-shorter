package gzip

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

func GZIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.Contains(request.Header.Get(ContentEncoding), HeaderCode) {
			compressReader, compressReaderError := newCompressReader(request.Body)
			if compressReaderError != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer func() {
				err := compressReader.Close()
				if err != nil {
					log.Logger.Err(err).Str("method", request.Method).Str("path", request.URL.Path)
				}
			}()
			request.Body = compressReader
		}

		responseWriter := writer
		if strings.Contains(request.Header.Get(AcceptEncoding), HeaderCode) {
			wrapper := &responseWriterWrapper{
				ResponseWriter: writer,
				request:        request,
			}
			defer func() {
				if closeError := wrapper.Close(); closeError != nil {
					log.Logger.Err(closeError).Str("method", request.Method).Str("path", request.URL.Path)
				}
			}()
			responseWriter = wrapper
		}

		next.ServeHTTP(responseWriter, request)
	})
}
