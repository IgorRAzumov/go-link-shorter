package link

import (
	"context"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service"
)

type mockResolverService struct {
	getShortKeyByURLFunc func(ctx context.Context, URL string) string
}

func (m *mockResolverService) GetFullLink(ctx context.Context, shortKey string) (string, error) {
	return "", nil
}

func (m *mockResolverService) GetShortKeyByURL(ctx context.Context, URL string) string {
	if m.getShortKeyByURLFunc != nil {
		return m.getShortKeyByURLFunc(ctx, URL)
	}
	return ""
}

type mockShorterService struct {
	generateShortKeyFunc            func(URL string) string
	createShortKeysFunc             func(ctx context.Context, links []*model.Link) error
	processBatchShortenRequestsFunc func(ctx context.Context, requests []model.BatchShortenRequest, resolver service.ResolverService) ([]model.BatchShortenResult, error)
}

func (m *mockShorterService) CreateShortKey(ctx context.Context, URL string) (string, error) {
	return "", nil
}

func (m *mockShorterService) GenerateShortKey(URL string) string {
	if m.generateShortKeyFunc != nil {
		return m.generateShortKeyFunc(URL)
	}
	return ""
}

func (m *mockShorterService) CreateShortKeys(ctx context.Context, links []*model.Link) error {
	if m.createShortKeysFunc != nil {
		return m.createShortKeysFunc(ctx, links)
	}
	return nil
}

func (m *mockShorterService) ProcessBatchShortenRequests(ctx context.Context, requests []model.BatchShortenRequest, resolver service.ResolverService) ([]model.BatchShortenResult, error) {
	if m.processBatchShortenRequestsFunc != nil {
		return m.processBatchShortenRequestsFunc(ctx, requests, resolver)
	}
	return []model.BatchShortenResult{}, nil
}

func TestCreateShortKeysBatch_EmptyURLs(t *testing.T) {
	resolver := &mockResolverService{}
	shorter := &mockShorterService{}
	usecase := NewLinkUsecase(resolver, shorter, "http://localhost:8080")

	ctx := context.Background()
	result, err := usecase.CreateShortKeysBatch(ctx, []string{})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d items", len(result))
	}
}

func TestCreateShortKeysBatch_WithExistingURLs(t *testing.T) {
	existingKey := "existing-key-123"
	resolver := &mockResolverService{
		getShortKeyByURLFunc: func(ctx context.Context, URL string) string {
			if URL == "https://example.com" {
				return existingKey
			}
			return ""
		},
	}
	shorter := &mockShorterService{}
	usecase := NewLinkUsecase(resolver, shorter, "http://localhost:8080")

	ctx := context.Background()
	result, err := usecase.CreateShortKeysBatch(ctx, []string{"https://example.com"})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}
	if result["https://example.com"] != existingKey {
		t.Errorf("Expected existing key '%s', got '%s'", existingKey, result["https://example.com"])
	}
}

func TestCreateShortKeysBatch_WithNewURLs(t *testing.T) {
	newKey := "new-key-456"
	resolver := &mockResolverService{
		getShortKeyByURLFunc: func(ctx context.Context, URL string) string {
			return ""
		},
	}
	var savedLinks []*model.Link
	shorter := &mockShorterService{
		generateShortKeyFunc: func(URL string) string {
			return newKey
		},
		createShortKeysFunc: func(ctx context.Context, links []*model.Link) error {
			savedLinks = links
			return nil
		},
	}
	usecase := NewLinkUsecase(resolver, shorter, "http://localhost:8080")

	ctx := context.Background()
	result, err := usecase.CreateShortKeysBatch(ctx, []string{"https://new-example.com"})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}
	if result["https://new-example.com"] != newKey {
		t.Errorf("Expected new key '%s', got '%s'", newKey, result["https://new-example.com"])
	}
	if len(savedLinks) != 1 {
		t.Fatalf("Expected 1 link to be saved, got %d", len(savedLinks))
	}
	if savedLinks[0].ShortKey != newKey {
		t.Errorf("Expected saved link key '%s', got '%s'", newKey, savedLinks[0].ShortKey)
	}
	if savedLinks[0].FullURL != "https://new-example.com" {
		t.Errorf("Expected saved link URL 'https://new-example.com', got '%s'", savedLinks[0].FullURL)
	}
}

func TestCreateShortKeysBatch_WithMixedURLs(t *testing.T) {
	existingKey := "existing-key"
	newKey := "new-key"
	resolver := &mockResolverService{
		getShortKeyByURLFunc: func(ctx context.Context, URL string) string {
			if URL == "https://existing.com" {
				return existingKey
			}
			return ""
		},
	}
	var savedLinks []*model.Link
	shorter := &mockShorterService{
		generateShortKeyFunc: func(URL string) string {
			return newKey
		},
		createShortKeysFunc: func(ctx context.Context, links []*model.Link) error {
			savedLinks = links
			return nil
		},
	}
	usecase := NewLinkUsecase(resolver, shorter, "http://localhost:8080")

	ctx := context.Background()
	result, err := usecase.CreateShortKeysBatch(ctx, []string{"https://existing.com", "https://new.com"})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(result))
	}
	if result["https://existing.com"] != existingKey {
		t.Errorf("Expected existing key '%s', got '%s'", existingKey, result["https://existing.com"])
	}
	if result["https://new.com"] != newKey {
		t.Errorf("Expected new key '%s', got '%s'", newKey, result["https://new.com"])
	}
	if len(savedLinks) != 1 {
		t.Fatalf("Expected 1 link to be saved (only new one), got %d", len(savedLinks))
	}
}

func TestCreateShortKeysBatch_GenerateShortKeyReturnsEmpty(t *testing.T) {
	resolver := &mockResolverService{
		getShortKeyByURLFunc: func(ctx context.Context, URL string) string {
			return ""
		},
	}
	shorter := &mockShorterService{
		generateShortKeyFunc: func(URL string) string {
			return ""
		},
	}
	usecase := NewLinkUsecase(resolver, shorter, "http://localhost:8080")

	ctx := context.Background()
	result, err := usecase.CreateShortKeysBatch(ctx, []string{"https://example.com"})

	if err == nil {
		t.Error("Expected error when GenerateShortKey returns empty, got nil")
	}
	if result != nil {
		t.Errorf("Expected nil result on error, got %v", result)
	}
}

func TestCreateShortKeysBatch_NormalizesURLs(t *testing.T) {
	newKey := "normalized-key"
	resolver := &mockResolverService{
		getShortKeyByURLFunc: func(ctx context.Context, URL string) string {
			return ""
		},
	}
	var normalizedURL string
	shorter := &mockShorterService{
		generateShortKeyFunc: func(URL string) string {
			normalizedURL = URL
			return newKey
		},
		createShortKeysFunc: func(ctx context.Context, links []*model.Link) error {
			return nil
		},
	}
	usecase := NewLinkUsecase(resolver, shorter, "http://localhost:8080")

	ctx := context.Background()
	_, err := usecase.CreateShortKeysBatch(ctx, []string{"https://example.com/"})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if normalizedURL != "https://example.com" {
		t.Errorf("Expected normalized URL 'https://example.com', got '%s'", normalizedURL)
	}
}
