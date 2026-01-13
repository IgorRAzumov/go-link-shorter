package shorter

import (
	"context"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

type mockLinkRepository struct {
	getByShortKeyFunc    func(context context.Context, shortURL string) string
	getShortKeyByURLFunc func(context context.Context, URL string) string
	isExistShortKeyFunc  func(context context.Context, shortURL string) bool
	saveFunc             func(context context.Context, link *model.Link)
	batchSaveFunc        func(context context.Context, links []*model.Link)
}

func (m *mockLinkRepository) GetByShortKey(context context.Context, shortURL string) string {
	if m.getByShortKeyFunc != nil {
		return m.getByShortKeyFunc(context, shortURL)
	}
	return ""
}

func (m *mockLinkRepository) GetShortKeyByURL(context context.Context, URL string) string {
	if m.getShortKeyByURLFunc != nil {
		return m.getShortKeyByURLFunc(context, URL)
	}
	return ""
}

func (m *mockLinkRepository) IsExistShortKey(context context.Context, shortURL string) bool {
	if m.isExistShortKeyFunc != nil {
		return m.isExistShortKeyFunc(context, shortURL)
	}
	return false
}

func (m *mockLinkRepository) Save(context context.Context, link *model.Link) {
	if m.saveFunc != nil {
		m.saveFunc(context, link)
	}
}

func (m *mockLinkRepository) BatchSave(context context.Context, links []*model.Link) {
	if m.batchSaveFunc != nil {
		m.batchSaveFunc(context, links)
	}
}

func TestGenerateShortKey_ReturnsNonEmpty(t *testing.T) {
	repo := &mockLinkRepository{}
	service := NewShorterService(repo)

	shortKey := service.GenerateShortKey("https://example.com")

	if shortKey == "" {
		t.Error("GenerateShortKey should return non-empty string")
	}
}

func TestGenerateShortKey_ConsistentOutput(t *testing.T) {
	repo := &mockLinkRepository{}
	service := NewShorterService(repo)

	url := "https://example.com"
	firstKey := service.GenerateShortKey(url)
	secondKey := service.GenerateShortKey(url)

	if firstKey != secondKey {
		t.Errorf("GenerateShortKey should return consistent output, got '%s' and '%s'", firstKey, secondKey)
	}
}

func TestGenerateShortKey_DifferentURLs(t *testing.T) {
	repo := &mockLinkRepository{}
	service := NewShorterService(repo)

	firstKey := service.GenerateShortKey("https://example1.com")
	secondKey := service.GenerateShortKey("https://example2.com")

	if firstKey == secondKey {
		t.Error("GenerateShortKey should return different keys for different URLs")
	}
}

func TestCreateShortKeys_EmptySlice(t *testing.T) {
	repo := &mockLinkRepository{}
	service := NewShorterService(repo)

	ctx := context.Background()
	service.CreateShortKeys(ctx, []*model.Link{})

	if repo.batchSaveFunc != nil {
		t.Error("BatchSave should not be called for empty slice")
	}
}

func TestCreateShortKeys_CallsBatchSave(t *testing.T) {
	var savedLinks []*model.Link
	repo := &mockLinkRepository{
		batchSaveFunc: func(context context.Context, links []*model.Link) {
			savedLinks = links
		},
	}
	service := NewShorterService(repo)

	ctx := context.Background()
	testLinks := []*model.Link{
		{
			ShortKey: "key1",
			FullURL:  "https://example1.com",
		},
		{
			ShortKey: "key2",
			FullURL:  "https://example2.com",
		},
	}

	service.CreateShortKeys(ctx, testLinks)

	if len(savedLinks) != len(testLinks) {
		t.Fatalf("Expected %d links to be saved, got %d", len(testLinks), len(savedLinks))
	}

	for i, expectedLink := range testLinks {
		if savedLinks[i].ShortKey != expectedLink.ShortKey {
			t.Errorf("Link %d: Expected ShortKey '%s', got '%s'", i, expectedLink.ShortKey, savedLinks[i].ShortKey)
		}
		if savedLinks[i].FullURL != expectedLink.FullURL {
			t.Errorf("Link %d: Expected FullURL '%s', got '%s'", i, expectedLink.FullURL, savedLinks[i].FullURL)
		}
	}
}

func TestCreateShortKeys_MultipleLinks(t *testing.T) {
	var savedLinks []*model.Link
	repo := &mockLinkRepository{
		batchSaveFunc: func(context context.Context, links []*model.Link) {
			savedLinks = links
		},
	}
	service := NewShorterService(repo)

	ctx := context.Background()
	testLinks := []*model.Link{
		{ShortKey: "key1", FullURL: "https://example1.com"},
		{ShortKey: "key2", FullURL: "https://example2.com"},
		{ShortKey: "key3", FullURL: "https://example3.com"},
	}

	service.CreateShortKeys(ctx, testLinks)

	if len(savedLinks) != 3 {
		t.Fatalf("Expected 3 links, got %d", len(savedLinks))
	}
}
