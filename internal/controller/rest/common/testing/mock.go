package testing

import (
	"context"
	"errors"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
)

type MockLinkCreateUsecase struct {
	CreateShortKeyFunc              func(ctx context.Context, URL string) (string, error)
	CreateShortKeysBatchFunc        func(ctx context.Context, urls []string) (map[string]string, error)
	ProcessBatchShortenRequestsFunc func(ctx context.Context, requests []model.BatchShortenRequest) ([]model.BatchShortenResult, error)
	GetBaseURLFunc                  func() string
}

func (mock *MockLinkCreateUsecase) CreateShortKey(ctx context.Context, URL string) (string, error) {
	if mock.CreateShortKeyFunc != nil {
		return mock.CreateShortKeyFunc(ctx, URL)
	}
	return "", errors.New("generation short link error")
}

func (mock *MockLinkCreateUsecase) CreateShortKeysBatch(ctx context.Context, urls []string) (map[string]string, error) {
	if mock.CreateShortKeysBatchFunc != nil {
		return mock.CreateShortKeysBatchFunc(ctx, urls)
	}
	return make(map[string]string), nil
}

func (mock *MockLinkCreateUsecase) GetBaseURL() string {
	if mock.GetBaseURLFunc != nil {
		return mock.GetBaseURLFunc()
	}
	return ""
}

func (mock *MockLinkCreateUsecase) ProcessBatchShortenRequests(ctx context.Context, requests []model.BatchShortenRequest) ([]model.BatchShortenResult, error) {
	if mock.ProcessBatchShortenRequestsFunc != nil {
		return mock.ProcessBatchShortenRequestsFunc(ctx, requests)
	}
	return []model.BatchShortenResult{}, nil
}

var _ usecase.LinkCreateUsecase = (*MockLinkCreateUsecase)(nil)

type MockLinkReadUsecase struct {
	GetFullURLByShortKeyFunc func(ctx context.Context, shortKey string) (string, error)
	GetUserURLsFunc          func(ctx context.Context) ([]*model.Link, error)
	GetBaseURLFunc           func() string
}

func (mock *MockLinkReadUsecase) GetFullURLByShorKey(ctx context.Context, shortKey string) (string, error) {
	if mock.GetFullURLByShortKeyFunc != nil {
		return mock.GetFullURLByShortKeyFunc(ctx, shortKey)
	}
	return "", nil
}

func (mock *MockLinkReadUsecase) GetUserURLs(ctx context.Context) ([]*model.Link, error) {
	if mock.GetUserURLsFunc != nil {
		return mock.GetUserURLsFunc(ctx)
	}
	return []*model.Link{}, nil
}

func (mock *MockLinkReadUsecase) GetBaseURL() string {
	if mock.GetBaseURLFunc != nil {
		return mock.GetBaseURLFunc()
	}
	return ""
}

var _ usecase.LinkReadUsecase = (*MockLinkReadUsecase)(nil)

type MockLinkDeleteUsecase struct {
	DeleteUserURLsFunc func(ctx context.Context, shortKeys []string) error
}

func (mock *MockLinkDeleteUsecase) DeleteUserURLs(ctx context.Context, shortKeys []string) error {
	if mock.DeleteUserURLsFunc != nil {
		return mock.DeleteUserURLsFunc(ctx, shortKeys)
	}
	return nil
}

var _ usecase.LinkDeleteUsecase = (*MockLinkDeleteUsecase)(nil)

// MockHealthCheckUsecase — мок для HealthCheckUsecase.
type MockHealthCheckUsecase struct {
	CheckSystemConnectionsFunc func(ctx context.Context) (bool, error)
}

func (m *MockHealthCheckUsecase) CheckSystemConnections(ctx context.Context) (bool, error) {
	if m.CheckSystemConnectionsFunc != nil {
		return m.CheckSystemConnectionsFunc(ctx)
	}
	return true, nil
}

var _ usecase.HealthCheckUsecase = (*MockHealthCheckUsecase)(nil)
