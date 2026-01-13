package repository

import (
	"context"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

type LinkRepository interface {
	GetByShortKey(context context.Context, shortURL string) string

	GetShortKeyByURL(context context.Context, URL string) string

	IsExistShortKey(context context.Context, shortURL string) bool

	Save(context context.Context, link *model.Link)

	BatchSave(context context.Context, links []*model.Link)
}

type HealthCheckRepository interface {
	CheckStorageConnection(context context.Context) (bool, error)
}
