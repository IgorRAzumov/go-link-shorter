package testing

import (
	"context"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
)

type MockLinkUsecase struct {
	GetFullURLByShortKeyFunc func(context context.Context, shortKey string) (string, error)
	CreateShortKeyFunc       func(context context.Context, URL string) string
	GetBaseURLFunc           func() string
}

func (mock *MockLinkUsecase) GetFullURLByShorKey(context context.Context, shortKey string) (string, error) {
	if mock.GetFullURLByShortKeyFunc != nil {
		return mock.GetFullURLByShortKeyFunc(context, shortKey)
	}
	return "", nil
}

func (mock *MockLinkUsecase) CreateShortKey(context context.Context, URL string) string {
	if mock.CreateShortKeyFunc != nil {
		return mock.CreateShortKeyFunc(context, URL)
	}
	return ""
}

func (mock *MockLinkUsecase) GetBaseURL() string {
	if mock.GetBaseURLFunc != nil {
		return mock.GetBaseURLFunc()
	}
	return ""
}

var _ usecase.LinkUsecase = (*MockLinkUsecase)(nil)
