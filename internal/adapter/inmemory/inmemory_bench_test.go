package inmemory

import (
	"context"
	"fmt"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

func BenchmarkSave(b *testing.B) {
	storage, err := NewInMemoryFileStorage("")
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()

	var i int
	for b.Loop() {
		link := &model.Link{
			ShortKey: fmt.Sprintf("key%d", i%100),
			FullURL:  fmt.Sprintf("https://example.com/page%d", i%100),
			UserID:   "user-1",
		}
		_ = storage.Save(ctx, link)
		i++
	}
}

func BenchmarkGetByShortKey(b *testing.B) {
	storage, err := NewInMemoryFileStorage("")
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	_ = storage.Save(ctx, &model.Link{
		ShortKey: "testkey",
		FullURL:  "https://example.com",
		UserID:   "user-1",
	})

	for b.Loop() {
		_, _ = storage.GetByShortKey(ctx, "testkey")
	}
}

func BenchmarkGetByUserID(b *testing.B) {
	storage, err := NewInMemoryFileStorage("")
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		_ = storage.Save(ctx, &model.Link{
			ShortKey: fmt.Sprintf("key%d", i),
			FullURL:  fmt.Sprintf("https://example.com/%d", i),
			UserID:   "user-1",
		})
	}

	for b.Loop() {
		_, _ = storage.GetByUserID(ctx, "user-1")
	}
}

func BenchmarkBatchSave(b *testing.B) {
	storage, err := NewInMemoryFileStorage("")
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	links := make([]*model.Link, 100)
	for i := range links {
		links[i] = &model.Link{
			ShortKey: fmt.Sprintf("key%d", i),
			FullURL:  fmt.Sprintf("https://example.com/%d", i),
			UserID:   "user-1",
		}
	}

	for b.Loop() {
		_ = storage.BatchSave(ctx, links)
	}
}
