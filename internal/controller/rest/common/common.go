package common

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

const TextPlain = "text/plain"
const ContentType = "Content-Type"

func BadRequestError(writer http.ResponseWriter, err error) {
	log.Warn().Err(err).Msg("Validation error")
	http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func MethodNotAllowedError(writer http.ResponseWriter) {
	message := http.StatusText(http.StatusMethodNotAllowed)
	log.Warn().Err(nil).Msg(message)
	http.Error(writer, message, http.StatusMethodNotAllowed)
}

func InternalError(writer http.ResponseWriter, err error) {
	message := http.StatusText(http.StatusInternalServerError)
	log.Error().Err(err).Msg(message)
	http.Error(writer, message, http.StatusInternalServerError)
}

func NormalizeURL(url string) string {
	return strings.TrimSuffix(strings.TrimSpace(url), "/")
}
