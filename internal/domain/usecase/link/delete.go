package link

import (
	"context"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
)

type DeleteUsecase struct {
	deleter service.DeleteService
}

func NewLinkDeleteUsecase(deleter service.DeleteService) *DeleteUsecase {
	return &DeleteUsecase{deleter: deleter}
}

func (usecase *DeleteUsecase) DeleteUserURLs(ctx context.Context, shortKeys []string) error {
	userID := authctx.UserID(ctx)
	return usecase.deleter.Enqueue(userID, shortKeys)
}
