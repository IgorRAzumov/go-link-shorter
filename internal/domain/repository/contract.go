package repository

import (
	"context"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

// LinkRepository определяет контракт хранилища ссылок.
type LinkRepository interface {
	GetByShortKey(ctx context.Context, shortURL string) (fullURL string, isDeleted bool)

	GetShortKeyByURL(ctx context.Context, URL string) string

	IsExistShortKey(ctx context.Context, shortURL string) bool

	Save(ctx context.Context, link *model.Link) error

	BatchSave(ctx context.Context, links []*model.Link) error

	GetByUserID(ctx context.Context, userID string) ([]*model.Link, error)

	MarkDeleted(ctx context.Context, userID string, shortKeys []string) error
}

// HealthCheckRepository определяет контракт проверки доступности хранилища.
type HealthCheckRepository interface {
	CheckStorageConnection(ctx context.Context) (bool, error)
}

// AuditRepository определяет контракт сохранения аудит-событий.
type AuditRepository interface {
	Save(ctx context.Context, event model.AuditEvent) error
}
