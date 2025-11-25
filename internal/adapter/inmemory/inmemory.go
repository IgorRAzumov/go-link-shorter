package inmemory

import (
	"context"
	"sync"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/rs/zerolog/log"
)

type LinkStorage struct {
	linksByID  sync.Map
	linksByURL sync.Map
}

func NewInMemoryStorage() *LinkStorage {
	return &LinkStorage{}
}

func (storage *LinkStorage) GetByShortKey(context context.Context, shortURL string) string {
	log.Debug().Str("short_key", shortURL).Msg("storage: GetByShortKey")
	value, _ := storage.linksByID.Load(shortURL)
	return value.(model.Link).FullURL
}

func (storage *LinkStorage) IsExistShortKey(context context.Context, shortKey string) bool {
	log.Debug().Str("short_key", shortKey).Msg("storage: IsExistShortKey")
	_, ok := storage.linksByID.Load(shortKey)
	return ok
}

func (storage *LinkStorage) GetShortKeyByURL(context context.Context, URL string) string {
	log.Debug().Str("url", URL).Msg("storage: GetByURL")
	value, ok := storage.linksByURL.Load(common.NormalizeURL(URL))
	if ok {
		return value.(model.Link).ShortKey
	}
	return ""
}

/*
неоптимальное решение, как кажется - но как понимаю в первых итерациях этим можно принебречь,
а с переходом на реализацию в БД это утратит актуальность
*/
func (storage *LinkStorage) Save(context context.Context, link *model.Link) {
	storage.linksByID.Store(link.ShortKey, link)
	storage.linksByURL.Store(common.NormalizeURL(link.FullURL), link)
}
