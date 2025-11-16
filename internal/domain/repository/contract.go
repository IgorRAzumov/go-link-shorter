package repository

import (
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

type LinkRepository interface {
	GetByShortKey(shortURL string) string

	GetShortKeyByURL(URL string) string

	IsExistShortKey(shortURL string) bool

	Save(link model.Link)
}
