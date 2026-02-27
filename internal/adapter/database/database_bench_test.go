package database

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

func getBenchDSN(b *testing.B) string {
	dsn := os.Getenv("BENCH_DATABASE_DSN")
	if dsn == "" {
		dsn = os.Getenv("TEST_DATABASE_DSN")
	}
	if dsn == "" {
		b.Skip("BENCH_DATABASE_DSN or TEST_DATABASE_DSN not set, skipping database benchmark")
	}
	return dsn
}

func BenchmarkSave(b *testing.B) {
	storage, err := NewStorage(getBenchDSN(b))
	if err != nil {
		b.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		link := &model.Link{
			ShortKey: fmt.Sprintf("bench_save_%d_%d", i, b.N),
			FullURL:  fmt.Sprintf("https://example.com/bench/save/%d/%d", i, b.N),
			UserID:   "bench-user",
		}
		if err := storage.Save(ctx, link); err != nil {
			b.Fatalf("Save failed: %v", err)
		}
	}
}

func BenchmarkGetByShortKey(b *testing.B) {
	storage, err := NewStorage(getBenchDSN(b))
	if err != nil {
		b.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()
	_ = storage.Save(ctx, &model.Link{
		ShortKey: "bench_getkey",
		FullURL:  "https://example.com/bench-get",
		UserID:   "bench-user",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.GetByShortKey(ctx, "bench_getkey")
	}
}

func BenchmarkGetByUserID(b *testing.B) {
	storage, err := NewStorage(getBenchDSN(b))
	if err != nil {
		b.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()
	userID := "bench_getbyuser"
	for i := 0; i < 100; i++ {
		_ = storage.Save(ctx, &model.Link{
			ShortKey: fmt.Sprintf("bench_getbyuser_%d", i),
			FullURL:  fmt.Sprintf("https://example.com/bench/getbyuser/%d", i),
			UserID:   userID,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.GetByUserID(ctx, userID)
	}
}

func BenchmarkBatchSave(b *testing.B) {
	storage, err := NewStorage(getBenchDSN(b))
	if err != nil {
		b.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		links := make([]*model.Link, 100)
		for j := range links {
			links[j] = &model.Link{
				ShortKey: fmt.Sprintf("bench_batch_%d_%d_%d", i, j, b.N),
				FullURL:  fmt.Sprintf("https://example.com/bench/batch/%d/%d/%d", i, j, b.N),
				UserID:   "bench-user",
			}
		}
		if err := storage.BatchSave(ctx, links); err != nil {
			b.Fatalf("BatchSave failed: %v", err)
		}
	}
}
