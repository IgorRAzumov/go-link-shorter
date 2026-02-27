package shorter

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/adapter/database"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/resolver"
)

func getBenchDSN(b *testing.B) string {
	dsn := os.Getenv("BENCH_DATABASE_DSN")
	if dsn == "" {
		dsn = os.Getenv("TEST_DATABASE_DSN")
	}
	if dsn == "" {
		b.Skip("BENCH_DATABASE_DSN or TEST_DATABASE_DSN not set, skipping DB benchmark")
	}
	return dsn
}

type benchLinkRepository struct {
	saveFunc      func(ctx context.Context, link *model.Link) error
	batchSaveFunc func(ctx context.Context, links []*model.Link) error
}

func (b *benchLinkRepository) GetByShortKey(ctx context.Context, shortURL string) (string, bool) {
	return "", false
}

func (b *benchLinkRepository) GetShortKeyByURL(ctx context.Context, URL string) string {
	return ""
}

func (b *benchLinkRepository) IsExistShortKey(ctx context.Context, shortURL string) bool {
	return false
}

func (b *benchLinkRepository) Save(ctx context.Context, link *model.Link) error {
	if b.saveFunc != nil {
		return b.saveFunc(ctx, link)
	}
	return nil
}

func (b *benchLinkRepository) BatchSave(ctx context.Context, links []*model.Link) error {
	if b.batchSaveFunc != nil {
		return b.batchSaveFunc(ctx, links)
	}
	return nil
}

func (b *benchLinkRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Link, error) {
	return nil, nil
}

func (b *benchLinkRepository) MarkDeleted(ctx context.Context, userID string, shortKeys []string) error {
	return nil
}

func BenchmarkGenerateShortKey(b *testing.B) {
	repo := &benchLinkRepository{}
	service := NewShorterService(repo)
	url := "https://example.com/path/to/long/url"

	for b.Loop() {
		_ = service.GenerateShortKey(url)
	}
}

func BenchmarkCreateShortKey(b *testing.B) {
	repo := &benchLinkRepository{
		saveFunc: func(ctx context.Context, link *model.Link) error {
			return nil
		},
	}
	service := NewShorterService(repo)
	ctx := context.Background()
	url := "https://example.com/path"
	userID := "user-123"

	for b.Loop() {
		_, _ = service.CreateShortKey(ctx, url, userID)
	}
}

func BenchmarkProcessBatchShortenRequests(b *testing.B) {
	repo := &benchLinkRepository{
		batchSaveFunc: func(ctx context.Context, links []*model.Link) error {
			return nil
		},
	}
	service := NewShorterService(repo)
	resolver := &benchResolverService{}
	ctx := context.Background()
	requests := make([]model.BatchShortenRequest, 100)
	for i := range requests {
		requests[i] = model.BatchShortenRequest{
			CorrelationID: fmt.Sprintf("corr-%d", i),
			OriginalURL:   fmt.Sprintf("https://example.com/page/%d", i),
		}
	}

	for b.Loop() {
		_, _ = service.ProcessBatchShortenRequests(ctx, requests, resolver, "user-123")
	}
}

type benchResolverService struct{}

func (b *benchResolverService) GetFullLink(ctx context.Context, shortKey string) (string, error) {
	return "", nil
}

func (b *benchResolverService) GetShortKeyByURL(ctx context.Context, url string) string {
	return ""
}

func (b *benchResolverService) GetByUserID(ctx context.Context, userID string) ([]*model.Link, error) {
	return nil, nil
}

func BenchmarkCreateShortKey_DB(b *testing.B) {
	storage, err := database.NewStorage(getBenchDSN(b))
	if err != nil {
		b.Fatal(err)
	}
	defer storage.Close()

	service := NewShorterService(storage)
	ctx := context.Background()
	userID := "bench-user"

	var i int
	for b.Loop() {
		url := fmt.Sprintf("https://example.com/db/create/%d", i)
		_, _ = service.CreateShortKey(ctx, url, userID)
		i++
	}
}

func BenchmarkProcessBatchShortenRequests_DB(b *testing.B) {
	storage, err := database.NewStorage(getBenchDSN(b))
	if err != nil {
		b.Fatal(err)
	}
	defer storage.Close()

	service := NewShorterService(storage)
	resolverSvc := resolver.NewResolverService(storage)
	ctx := context.Background()

	var i int
	for b.Loop() {
		requests := make([]model.BatchShortenRequest, 100)
		for j := range requests {
			requests[j] = model.BatchShortenRequest{
				CorrelationID: fmt.Sprintf("corr-%d", j),
				OriginalURL:   fmt.Sprintf("https://example.com/db/batch/%d/%d", i, j),
			}
		}
		_, _ = service.ProcessBatchShortenRequests(ctx, requests, resolverSvc, "bench-user")
		i++
	}
}
