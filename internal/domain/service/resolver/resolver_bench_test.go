package resolver

import (
	"context"
	"os"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/adapter/database"
	"github.com/IgorRAzumov/go-link-shorter/internal/adapter/inmemory"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
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

func BenchmarkGetFullLink(b *testing.B) {
	storage, err := inmemory.NewInMemoryFileStorage("")
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	_ = storage.Save(ctx, &model.Link{
		ShortKey: "abc123",
		FullURL:  "https://example.com/long/url/path",
		UserID:   "user-1",
	})

	service := NewResolverService(storage)

	for b.Loop() {
		_, _ = service.GetFullLink(ctx, "abc123")
	}
}

func BenchmarkGetShortKeyByURL(b *testing.B) {
	storage, err := inmemory.NewInMemoryFileStorage("")
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	_ = storage.Save(ctx, &model.Link{
		ShortKey: "abc123",
		FullURL:  "https://example.com/long/url/path",
		UserID:   "user-1",
	})

	service := NewResolverService(storage)

	for b.Loop() {
		_ = service.GetShortKeyByURL(ctx, "https://example.com/long/url/path")
	}
}

func BenchmarkGetFullLink_DB(b *testing.B) {
	storage, err := database.NewStorage(getBenchDSN(b))
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = storage.Close() }()

	ctx := context.Background()
	_ = storage.Save(ctx, &model.Link{
		ShortKey: "bench_resolver_abc",
		FullURL:  "https://example.com/bench/resolver/path",
		UserID:   "user-1",
	})

	service := NewResolverService(storage)

	for b.Loop() {
		_, _ = service.GetFullLink(ctx, "bench_resolver_abc")
	}
}

func BenchmarkGetShortKeyByURL_DB(b *testing.B) {
	storage, err := database.NewStorage(getBenchDSN(b))
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = storage.Close() }()

	ctx := context.Background()
	_ = storage.Save(ctx, &model.Link{
		ShortKey: "bench_resolver_xyz",
		FullURL:  "https://example.com/bench/resolver/url",
		UserID:   "user-1",
	})

	service := NewResolverService(storage)

	for b.Loop() {
		_ = service.GetShortKeyByURL(ctx, "https://example.com/bench/resolver/url")
	}
}
