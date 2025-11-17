package usecase

type LinkUsecase interface {
	GetFullURLByShorKey(shortKey string) (string, error)
	CreateShortKey(URL string) string
	GetBaseURL() string
}
