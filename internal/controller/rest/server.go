package rest

import (
	"log"
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase/link"
)

func Start(linkUsecase *link.Usecase, serverAddress string) {
	router := NewRouter(linkUsecase)
	if err := http.ListenAndServe(serverAddress, router); err != nil {
		log.Fatal("Http server start error:", err)
	}
}
