package shorter

import (
	"context"
	"crypto/sha256"
	"encoding/base64"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/repository"
	"github.com/rs/zerolog/log"
)

type Service struct {
	linkRepo repository.LinkRepository
}

func NewShorterService(repository repository.LinkRepository) *Service {
	return &Service{repository}
}

func (service *Service) CreateShortKey(context context.Context, URL string) string {
	shortKey := generateShortKey(URL)
	log.Debug().Str("short_key", shortKey).Str("url", URL).Msg("created shortKey")
	service.linkRepo.Save(context, &model.Link{ShortKey: shortKey, FullURL: URL})
	return shortKey
}

func generateShortKey(URL string) string {
	hash := sha256.Sum256([]byte(URL))
	return base64.RawURLEncoding.EncodeToString(hash[:8])
}
