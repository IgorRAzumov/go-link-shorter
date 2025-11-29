package gzip

import (
	"net/http"
	"strings"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/rs/zerolog/log"
)

func GZIPMiddllear(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		contentType := request.Header.Get(common.ContentType)
		if contentType != common.ApplicationJSON && contentType != common.TextHTML {
			next.ServeHTTP(writer, request)
			return
		}

		if strings.Contains(request.Header.Get(AcceptEncoding), HeaderCode) {
			compressWriter := newCompressWriter(writer)
			defer func() {
				err := compressWriter.Close()
				if err != nil {
					log.Logger.Err(err).Str("method", request.Method).Str("path", request.URL.Path)
				}
			}()

			writer = compressWriter
		}

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
			// Заменяем Body на распакованный reader
			request.Body = compressReader
		}

		next.ServeHTTP(writer, request)
	})
}
