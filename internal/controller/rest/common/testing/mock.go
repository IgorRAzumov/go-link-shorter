package testing

import (
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
)

type MockLinkUsecase struct {
	GetFullURLByShortKeyFunc func(shortKey string) (string, error)
	CreateShortKeyFunc       func(URL string) string
	GetBaseUrlFunc           func() string
}

func (mock *MockLinkUsecase) GetFullURLByShorKey(shortKey string) (string, error) {
	if mock.GetFullURLByShortKeyFunc != nil {
		return mock.GetFullURLByShortKeyFunc(shortKey)
	}
	return "", nil
}

func (mock *MockLinkUsecase) CreateShortKey(URL string) string {
	if mock.CreateShortKeyFunc != nil {
		return mock.CreateShortKeyFunc(URL)
	}
	return ""
}

func (mock *MockLinkUsecase) GetBaseUrl() string {
	if mock.GetBaseUrlFunc != nil {
		return mock.GetBaseUrlFunc()
	}
	return ""
}

var _ usecase.LinkUsecase = (*MockLinkUsecase)(nil)
