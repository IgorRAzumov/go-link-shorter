package link

import (
	"context"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/repository"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/deleter"
)

type DeleteUsecase struct {
	deleter service.DeleteService
}

func NewLinkDeleteUsecase(repo repository.LinkRepository) *DeleteUsecase {
	if repo == nil {
		return &DeleteUsecase{}
	}

	deleteService := deleter.New(repo, deleter.DefaultMaxBatch)
	deleteService.Start(context.Background())

	return &DeleteUsecase{deleter: deleteService}
}

func (usecase *DeleteUsecase) DeleteUserURLs(ctx context.Context, shortKeys []string) error {
	userID := authctx.UserID(ctx)
	if usecase.deleter == nil {
		return nil
	}
	return usecase.deleter.Enqueue(userID, shortKeys)
}
