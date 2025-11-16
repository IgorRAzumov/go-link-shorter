package service

type ResolverService interface {
	GetFullLink(shortKey string) (string, error)
	GetShortKeyByURL(url string) string
}

type ShorterService interface {
	CreateShortKey(URL string) string
}
