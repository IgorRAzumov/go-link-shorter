package shorter

import (
	"crypto/sha256"
	"encoding/base64"
	"log"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/repository"
)

type Service struct {
	linkRepo repository.LinkRepository
}

func NewShorterService(repository repository.LinkRepository) *Service {
	return &Service{repository}
}

func (service *Service) CreateShortKey(URL string) string {
	shortKey := generateShortKey(URL)
	log.Printf("created shortKey%s", shortKey)
	service.linkRepo.Save(model.Link{ShortKey: shortKey, FullURL: URL})
	return shortKey
}

func generateShortKey(URL string) string {
	hash := sha256.Sum256([]byte(URL))
	return base64.RawURLEncoding.EncodeToString(hash[:8])
}
