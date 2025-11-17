package app

import (
	"github.com/IgorRAzumov/go-link-shorter/internal/adapter/inmemory"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/resolver"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/shorter"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase/link"
)

func Run(serverAddress, baseURL string) {
	storage := inmemory.NewInMemoryStorage()
	shorterService := shorter.NewShorterService(storage)
	resolverService := resolver.NewResolverService(storage)
	linkUsecase := link.NewLinkUsecase(resolverService, shorterService, baseURL)

	rest.Start(linkUsecase, serverAddress)
}
