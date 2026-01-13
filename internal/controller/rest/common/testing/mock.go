package testing

import (
	"context"
	"errors"

	handlermodel "github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/shorter/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
)

type MockLinkUsecase struct {
	GetFullURLByShortKeyFunc        func(ctx context.Context, shortKey string) (string, error)
	CreateShortKeyFunc              func(ctx context.Context, URL string) (string, error)
	CreateShortKeysBatchFunc        func(ctx context.Context, urls []string) (map[string]string, error)
	ProcessBatchShortenRequestsFunc func(ctx context.Context, requests []handlermodel.BatchShortenRequest, scheme, host string) ([]handlermodel.BatchShortenResponse, error)
	GetBaseURLFunc                  func() string
}

func (mock *MockLinkUsecase) GetFullURLByShorKey(ctx context.Context, shortKey string) (string, error) {
	if mock.GetFullURLByShortKeyFunc != nil {
		return mock.GetFullURLByShortKeyFunc(ctx, shortKey)
	}
	return "", nil
}

func (mock *MockLinkUsecase) CreateShortKey(ctx context.Context, URL string) (string, error) {
	if mock.CreateShortKeyFunc != nil {
		return mock.CreateShortKeyFunc(ctx, URL)
	}
	return "", errors.New("generation short link error")
}

func (mock *MockLinkUsecase) CreateShortKeysBatch(ctx context.Context, urls []string) (map[string]string, error) {
	if mock.CreateShortKeysBatchFunc != nil {
		return mock.CreateShortKeysBatchFunc(ctx, urls)
	}
	return make(map[string]string), nil
}

func (mock *MockLinkUsecase) GetBaseURL() string {
	if mock.GetBaseURLFunc != nil {
		return mock.GetBaseURLFunc()
	}
	return ""
}

func (mock *MockLinkUsecase) ProcessBatchShortenRequests(ctx context.Context, requests []handlermodel.BatchShortenRequest, scheme, host string) ([]handlermodel.BatchShortenResponse, error) {
	if mock.ProcessBatchShortenRequestsFunc != nil {
		return mock.ProcessBatchShortenRequestsFunc(ctx, requests, scheme, host)
	}
	return []handlermodel.BatchShortenResponse{}, nil
}

var _ usecase.LinkUsecase = (*MockLinkUsecase)(nil)
