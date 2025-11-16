package rest

import (
	"log"
	"net/http"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase/link"
)

func Start(linkUsecase *link.Usecase) {
	router := NewRouter(linkUsecase)
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal("Http server start error:", err)
	}
}
