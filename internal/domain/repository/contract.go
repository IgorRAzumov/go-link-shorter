package repository

import (
	"context"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

type LinkRepository interface {
	GetByShortKey(ctx context.Context, shortURL string) string

	GetShortKeyByURL(ctx context.Context, URL string) string

	IsExistShortKey(ctx context.Context, shortURL string) bool

	Save(ctx context.Context, link *model.Link) error

	BatchSave(ctx context.Context, links []*model.Link) error
}

type HealthCheckRepository interface {
	CheckStorageConnection(ctx context.Context) (bool, error)
}
