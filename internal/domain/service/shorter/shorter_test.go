package shorter

import (
	"context"
	"errors"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

type mockLinkRepository struct {
	getByShortKeyFunc    func(ctx context.Context, shortURL string) (string, bool)
	getShortKeyByURLFunc func(ctx context.Context, URL string) string
	isExistShortKeyFunc  func(ctx context.Context, shortURL string) bool
	saveFunc             func(ctx context.Context, link *model.Link) error
	batchSaveFunc        func(ctx context.Context, links []*model.Link) error
	getByUserIDFunc      func(ctx context.Context, userID string) ([]*model.Link, error)
	markDeletedFunc      func(ctx context.Context, userID string, shortKeys []string) error
}

func (m *mockLinkRepository) GetByShortKey(ctx context.Context, shortURL string) (string, bool) {
	if m.getByShortKeyFunc != nil {
		return m.getByShortKeyFunc(ctx, shortURL)
	}
	return "", false
}

func (m *mockLinkRepository) GetShortKeyByURL(ctx context.Context, URL string) string {
	if m.getShortKeyByURLFunc != nil {
		return m.getShortKeyByURLFunc(ctx, URL)
	}
	return ""
}

func (m *mockLinkRepository) IsExistShortKey(ctx context.Context, shortURL string) bool {
	if m.isExistShortKeyFunc != nil {
		return m.isExistShortKeyFunc(ctx, shortURL)
	}
	return false
}

func (m *mockLinkRepository) Save(ctx context.Context, link *model.Link) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, link)
	}
	return nil
}

func (m *mockLinkRepository) BatchSave(ctx context.Context, links []*model.Link) error {
	if m.batchSaveFunc != nil {
		return m.batchSaveFunc(ctx, links)
	}
	return nil
}

func (m *mockLinkRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Link, error) {
	if m.getByUserIDFunc != nil {
		return m.getByUserIDFunc(ctx, userID)
	}
	return []*model.Link{}, nil
}

func (m *mockLinkRepository) MarkDeleted(ctx context.Context, userID string, shortKeys []string) error {
	if m.markDeletedFunc != nil {
		return m.markDeletedFunc(ctx, userID, shortKeys)
	}
	return nil
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
	err := service.CreateShortKeys(ctx, []*model.Link{})
	if err != nil {
		t.Fatalf("CreateShortKeys should not return an error")
	}

	if repo.batchSaveFunc != nil {
		t.Error("BatchSave should not be called for empty slice")
	}
}

func TestCreateShortKeys_CallsBatchSave(t *testing.T) {
	var savedLinks []*model.Link
	repo := &mockLinkRepository{
		batchSaveFunc: func(ctx context.Context, links []*model.Link) error {
			savedLinks = links
			return nil
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

	err := service.CreateShortKeys(ctx, testLinks)
	if err != nil {
		t.Fatalf("service.CreateShortKeys should not return an error")
	}

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
		batchSaveFunc: func(ctx context.Context, links []*model.Link) error {
			savedLinks = links
			return nil
		},
	}
	service := NewShorterService(repo)

	ctx := context.Background()
	testLinks := []*model.Link{
		{ShortKey: "key1", FullURL: "https://example1.com"},
		{ShortKey: "key2", FullURL: "https://example2.com"},
		{ShortKey: "key3", FullURL: "https://example3.com"},
	}

	err := service.CreateShortKeys(ctx, testLinks)
	if err != nil {
		t.Fatalf("service.CreateShortKeys should not return an error")
	}

	if len(savedLinks) != 3 {
		t.Fatalf("Expected 3 links, got %d", len(savedLinks))
	}
}

type mockResolverService struct {
	getFullLinkFunc      func(ctx context.Context, shortKey string) (string, error)
	getShortKeyByURLFunc func(ctx context.Context, url string) string
	getByUserIDFunc      func(ctx context.Context, userID string) ([]*model.Link, error)
}

func (m *mockResolverService) GetFullLink(ctx context.Context, shortKey string) (string, error) {
	if m.getFullLinkFunc != nil {
		return m.getFullLinkFunc(ctx, shortKey)
	}
	return "", nil
}

func (m *mockResolverService) GetShortKeyByURL(ctx context.Context, url string) string {
	if m.getShortKeyByURLFunc != nil {
		return m.getShortKeyByURLFunc(ctx, url)
	}
	return ""
}

func (m *mockResolverService) GetByUserID(ctx context.Context, userID string) ([]*model.Link, error) {
	if m.getByUserIDFunc != nil {
		return m.getByUserIDFunc(ctx, userID)
	}
	return []*model.Link{}, nil
}

func TestPrepareBatchData_NormalizesURLs(t *testing.T) {
	repo := &mockLinkRepository{}
	service := NewShorterService(repo)
	resolver := &mockResolverService{
		getShortKeyByURLFunc: func(ctx context.Context, url string) string {
			return ""
		},
	}

	ctx := context.Background()
	requests := []model.BatchShortenRequest{
		{CorrelationID: "1", OriginalURL: "https://example.com/"},
		{CorrelationID: "2", OriginalURL: "https://example.com"},
	}

	normalizedMap, urlToShortKey, linksToSave := service.prepareBatchData(requests, ctx, resolver, "test-user-id")

	if len(normalizedMap) != 2 {
		t.Fatalf("Expected normalizedMap to have 2 entries, got %d", len(normalizedMap))
	}

	normalizedURL1 := normalizedMap["https://example.com/"]
	normalizedURL2 := normalizedMap["https://example.com"]
	if normalizedURL1 != normalizedURL2 {
		t.Errorf("Expected normalized URLs to be equal, got '%s' and '%s'", normalizedURL1, normalizedURL2)
	}

	if len(urlToShortKey) != 1 {
		t.Fatalf("Expected urlToShortKey to have 1 entry (duplicates removed), got %d", len(urlToShortKey))
	}

	if len(linksToSave) != 1 {
		t.Fatalf("Expected 1 link to save (duplicates removed), got %d", len(linksToSave))
	}
}

func TestPrepareBatchData_HandlesExistingURLs(t *testing.T) {
	repo := &mockLinkRepository{}
	service := NewShorterService(repo)
	existingShortKey := "existing_key"
	resolver := &mockResolverService{
		getShortKeyByURLFunc: func(ctx context.Context, url string) string {
			if url == "https://example.com" {
				return existingShortKey
			}
			return ""
		},
	}

	ctx := context.Background()
	requests := []model.BatchShortenRequest{
		{CorrelationID: "1", OriginalURL: "https://example.com"},
		{CorrelationID: "2", OriginalURL: "https://new.com"},
	}

	normalizedMap, urlToShortKey, linksToSave := service.prepareBatchData(requests, ctx, resolver, "test-user-id")

	if len(normalizedMap) != 2 {
		t.Fatalf("Expected normalizedMap to have 2 entries, got %d", len(normalizedMap))
	}

	if urlToShortKey["https://example.com"] != existingShortKey {
		t.Errorf("Expected existing short key '%s', got '%s'", existingShortKey, urlToShortKey["https://example.com"])
	}

	if len(linksToSave) != 1 {
		t.Fatalf("Expected 1 link to save (existing URL skipped), got %d", len(linksToSave))
	}

	if linksToSave[0].FullURL != "https://new.com" {
		t.Errorf("Expected link to save for 'https://new.com', got '%s'", linksToSave[0].FullURL)
	}
}

func TestPrepareBatchData_HandlesDuplicateURLs(t *testing.T) {
	repo := &mockLinkRepository{}
	service := NewShorterService(repo)
	resolver := &mockResolverService{
		getShortKeyByURLFunc: func(ctx context.Context, url string) string {
			return ""
		},
	}

	ctx := context.Background()
	requests := []model.BatchShortenRequest{
		{CorrelationID: "1", OriginalURL: "https://example.com"},
		{CorrelationID: "2", OriginalURL: "https://example.com"},
		{CorrelationID: "3", OriginalURL: "https://example.com"},
	}

	_, urlToShortKey, linksToSave := service.prepareBatchData(requests, ctx, resolver, "test-user-id")

	if len(linksToSave) != 1 {
		t.Fatalf("Expected 1 link to save (duplicates removed), got %d", len(linksToSave))
	}

	shortKey := urlToShortKey["https://example.com"]
	if shortKey == "" {
		t.Error("Expected short key to be generated")
	}
}

func TestSaveBatchLinks_EmptySlice(t *testing.T) {
	repo := &mockLinkRepository{}
	service := NewShorterService(repo)
	urlToShortKey := make(map[string]string)

	ctx := context.Background()
	err := service.saveBatchLinks(ctx, []*model.Link{}, urlToShortKey)

	if err != nil {
		t.Errorf("Expected no error for empty slice, got %v", err)
	}

	if len(urlToShortKey) != 0 {
		t.Errorf("Expected urlToShortKey to remain empty, got %d entries", len(urlToShortKey))
	}
}

func TestSaveBatchLinks_Success(t *testing.T) {
	var savedLinks []*model.Link
	repo := &mockLinkRepository{
		batchSaveFunc: func(ctx context.Context, links []*model.Link) error {
			savedLinks = links
			return nil
		},
	}
	service := NewShorterService(repo)
	urlToShortKey := make(map[string]string)

	ctx := context.Background()
	linksToSave := []*model.Link{
		{ShortKey: "key1", FullURL: "https://example1.com"},
		{ShortKey: "key2", FullURL: "https://example2.com"},
	}

	err := service.saveBatchLinks(ctx, linksToSave, urlToShortKey)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(savedLinks) != 2 {
		t.Fatalf("Expected 2 links to be saved, got %d", len(savedLinks))
	}
}

func TestSaveBatchLinks_NonConflictError(t *testing.T) {
	expectedErr := errors.New("database error")
	repo := &mockLinkRepository{
		batchSaveFunc: func(ctx context.Context, links []*model.Link) error {
			return expectedErr
		},
	}
	service := NewShorterService(repo)
	urlToShortKey := make(map[string]string)

	ctx := context.Background()
	linksToSave := []*model.Link{
		{ShortKey: "key1", FullURL: "https://example1.com"},
	}

	err := service.saveBatchLinks(ctx, linksToSave, urlToShortKey)

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestHandleBatchSaveError_ConflictByShortKey(t *testing.T) {
	repo := &mockLinkRepository{}
	service := NewShorterService(repo)
	urlToShortKey := map[string]string{
		"https://example.com": "new_key",
	}

	conflictErr := &model.URLConflictError{ExistingShortKey: "new_key"}
	linksToSave := []*model.Link{
		{ShortKey: "new_key", FullURL: "https://example.com"},
	}

	err := service.handleBatchSaveError(conflictErr, linksToSave, urlToShortKey)

	if err != nil {
		t.Errorf("Expected no error after handling conflict, got %v", err)
	}

	if urlToShortKey["https://example.com"] != "new_key" {
		t.Errorf("Expected short key to remain 'new_key', got '%s'", urlToShortKey["https://example.com"])
	}
}

func TestHandleBatchSaveError_ConflictByURL(t *testing.T) {
	repo := &mockLinkRepository{}
	service := NewShorterService(repo)
	urlToShortKey := map[string]string{
		"https://example.com": "new_key",
	}

	existingShortKey := "existing_key"
	conflictErr := &model.URLConflictError{ExistingShortKey: existingShortKey}
	linksToSave := []*model.Link{
		{ShortKey: "new_key", FullURL: "https://example.com"},
	}

	err := service.handleBatchSaveError(conflictErr, linksToSave, urlToShortKey)

	if err != nil {
		t.Errorf("Expected no error after handling conflict, got %v", err)
	}

	if urlToShortKey["https://example.com"] != existingShortKey {
		t.Errorf("Expected short key to be updated to '%s', got '%s'", existingShortKey, urlToShortKey["https://example.com"])
	}
}

func TestHandleBatchSaveError_NonConflictError(t *testing.T) {
	repo := &mockLinkRepository{}
	service := NewShorterService(repo)
	urlToShortKey := make(map[string]string)

	expectedErr := errors.New("other error")
	linksToSave := []*model.Link{
		{ShortKey: "key1", FullURL: "https://example.com"},
	}

	err := service.handleBatchSaveError(expectedErr, linksToSave, urlToShortKey)

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestBuildBatchResults(t *testing.T) {
	repo := &mockLinkRepository{}
	service := NewShorterService(repo)

	requests := []model.BatchShortenRequest{
		{CorrelationID: "1", OriginalURL: "https://example1.com"},
		{CorrelationID: "2", OriginalURL: "https://example2.com"},
		{CorrelationID: "3", OriginalURL: "https://example3.com"},
	}

	normalizedMap := map[string]string{
		"https://example1.com": "https://example1.com",
		"https://example2.com": "https://example2.com",
		"https://example3.com": "https://example3.com",
	}

	urlToShortKey := map[string]string{
		"https://example1.com": "key1",
		"https://example2.com": "key2",
		"https://example3.com": "",
	}

	results := service.buildBatchResults(requests, normalizedMap, urlToShortKey)

	if len(normalizedMap) != 3 {
		t.Errorf("Expected normalizedMap to have 3 entries, got %d", len(normalizedMap))
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results (one URL without short key skipped), got %d", len(results))
	}

	if results[0].CorrelationID != "1" || results[0].ShortKey != "key1" {
		t.Errorf("Expected result 1: CorrelationID='1', ShortKey='key1', got CorrelationID='%s', ShortKey='%s'",
			results[0].CorrelationID, results[0].ShortKey)
	}

	if results[1].CorrelationID != "2" || results[1].ShortKey != "key2" {
		t.Errorf("Expected result 2: CorrelationID='2', ShortKey='key2', got CorrelationID='%s', ShortKey='%s'",
			results[1].CorrelationID, results[1].ShortKey)
	}
}

func TestProcessBatchShortenRequests_Integration(t *testing.T) {
	var savedLinks []*model.Link
	repo := &mockLinkRepository{
		batchSaveFunc: func(ctx context.Context, links []*model.Link) error {
			savedLinks = links
			return nil
		},
	}
	service := NewShorterService(repo)
	resolver := &mockResolverService{
		getShortKeyByURLFunc: func(ctx context.Context, url string) string {
			if url == "https://existing.com" {
				return "existing_key"
			}
			return ""
		},
	}

	ctx := context.Background()
	requests := []model.BatchShortenRequest{
		{CorrelationID: "1", OriginalURL: "https://existing.com"},
		{CorrelationID: "2", OriginalURL: "https://new.com"},
		{CorrelationID: "3", OriginalURL: "https://new.com"},
	}

	results, err := service.ProcessBatchShortenRequests(ctx, requests, resolver, "test-user-id")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	if results[0].CorrelationID != "1" || results[0].ShortKey != "existing_key" {
		t.Errorf("Expected result 1 to have existing key, got CorrelationID='%s', ShortKey='%s'",
			results[0].CorrelationID, results[0].ShortKey)
	}

	if results[1].CorrelationID != "2" || results[1].ShortKey == "" {
		t.Errorf("Expected result 2 to have generated key, got CorrelationID='%s', ShortKey='%s'",
			results[1].CorrelationID, results[1].ShortKey)
	}

	if results[2].CorrelationID != "3" || results[2].ShortKey == "" {
		t.Errorf("Expected result 3 to have generated key, got CorrelationID='%s', ShortKey='%s'",
			results[2].CorrelationID, results[2].ShortKey)
	}

	if len(savedLinks) != 1 {
		t.Fatalf("Expected 1 link to be saved (existing skipped, duplicates removed), got %d", len(savedLinks))
	}
}
