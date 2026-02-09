package service

import (
	"context"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

type ResolverService interface {
	GetFullLink(ctx context.Context, shortKey string) (string, error)
	GetShortKeyByURL(ctx context.Context, url string) string
	GetByUserID(ctx context.Context, userID string) ([]*model.Link, error)
}

type ShorterService interface {
	CreateShortKey(ctx context.Context, URL string, userID string) (string, error)
	GenerateShortKey(URL string) string
	CreateShortKeys(ctx context.Context, links []*model.Link) error
	ProcessBatchShortenRequests(ctx context.Context, requests []model.BatchShortenRequest, resolver ResolverService, userID string) ([]model.BatchShortenResult, error)
}

type HealthCheckService interface {
	CheckStorageConnection(ctx context.Context) (bool, error)
}

type DeleteService interface {
	Enqueue(userID string, shortKeys []string) error
}

type AuthService interface {
	GenerateUserID() string
	SignUserID(userID string) string
	ValidateSignedUserID(value string) (string, error)
}

type AuditorService interface {
	AuditNewEvent(ctx context.Context, event model.AuditEvent)
}
