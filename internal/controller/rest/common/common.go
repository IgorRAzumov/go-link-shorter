package common

import (
	"log"
	"net/http"
	"strings"
)

const TextPlain = "text/plain"
const ContentType = "Content-Type"

func BadRequestError(writer http.ResponseWriter, err error) {
	if err != nil {
		log.Printf("Validation error: %v", err)
	} else {
		log.Println("Validation error: incorrect request")
	}

	http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func InternalError(writer http.ResponseWriter, err error) {
	errorText := http.StatusText(http.StatusInternalServerError)
	log.Printf("%s %v", errorText, err)
	http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func NormalizeURL(url string) string {
	return strings.TrimSuffix(strings.TrimSpace(url), "/")
}
