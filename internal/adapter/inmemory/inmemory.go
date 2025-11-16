package inmemory

import (
	"log"
	"sync"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

type LinkStorage struct {
	linksById  sync.Map
	linksByURL sync.Map
}

func NewInMemoryStorage() *LinkStorage {
	return &LinkStorage{}
}

func (storage *LinkStorage) GetByShortKey(shortURL string) string {
	log.Default().Printf("storage :GetByShortKey(%s)", shortURL)
	value, _ := storage.linksById.Load(shortURL)
	return value.(model.Link).FullURL
}

func (storage *LinkStorage) IsExistShortKey(shortKey string) bool {
	log.Default().Printf("storage :IsExistShortKey(%s)", shortKey)
	_, ok := storage.linksById.Load(shortKey)
	return ok
}

func (storage *LinkStorage) GetShortKeyByURL(URL string) string {
	log.Default().Printf("storage :GetByURL(%s)", URL)
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
func (storage *LinkStorage) Save(link model.Link) {
	storage.linksById.Store(link.ShortKey, link)
	storage.linksByURL.Store(common.NormalizeURL(link.FullURL), link)
}
