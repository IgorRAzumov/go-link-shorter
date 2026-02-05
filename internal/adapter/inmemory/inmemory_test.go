package inmemory

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/adapter"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/google/uuid"
)

func TestNewInMemoryStorage(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")
	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}
	if storage == nil {
		t.Fatal("NewInMemoryFileStorage returned nil")
	}
}

func TestNewFileStorage(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test-storage.json")

	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}

	if storage == nil {
		t.Fatal("NewInMemoryFileStorage returned nil")
	}
}

func TestNewFileStorage_WithNonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "non-existent.json")

	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage should not fail for non-existent file: %v", err)
	}

	if storage == nil {
		t.Fatal("NewInMemoryFileStorage returned nil")
	}

	if _, err := os.Stat(filePath); err == nil {
		t.Error("File should not exist before first Save")
	}
}

func TestNewFileStorage_WithExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "existing-storage.json")

	existingLinks := []adapter.Link{
		{
			UUID:     "uuid-1",
			ShortKey: "key1",
			FullURL:  "https://example1.com",
		},
		{
			UUID:     "uuid-2",
			ShortKey: "key2",
			FullURL:  "https://example2.com",
		},
	}

	data, err := json.MarshalIndent(existingLinks, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}

	ctx := context.Background()

	if !storage.IsExistShortKey(ctx, "key1") {
		t.Error("Link 'key1' was not loaded from file")
	}

	if !storage.IsExistShortKey(ctx, "key2") {
		t.Error("Link 'key2' was not loaded from file")
	}

	retrievedURL, _ := storage.GetByShortKey(ctx, "key1")
	if retrievedURL != "https://example1.com" {
		t.Errorf("Expected URL 'https://example1.com', got '%s'", retrievedURL)
	}
}

func TestLinkStorage_Save(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")
	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}
	ctx := context.Background()

	testLink := &adapter.Link{
		ShortKey: "short-key-1",
		FullURL:  "https://example.com",
	}

	domainLink := &model.Link{
		ShortKey: testLink.ShortKey,
		FullURL:  testLink.FullURL,
	}
	storage.Save(ctx, domainLink)

	if !storage.IsExistShortKey(ctx, testLink.ShortKey) {
		t.Error("Link was not saved - IsExistShortKey returned false")
	}

	retrievedURL, _ := storage.GetByShortKey(ctx, testLink.ShortKey)
	if retrievedURL != testLink.FullURL {
		t.Errorf("Expected URL '%s', got '%s'", testLink.FullURL, retrievedURL)
	}
}

func TestLinkStorage_Save_WithFileStorage(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "save-test.json")

	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}

	ctx := context.Background()

	testLink := &adapter.Link{
		ShortKey: "save-key",
		FullURL:  "https://save-test.com",
	}

	domainLink := &model.Link{
		ShortKey: testLink.ShortKey,
		FullURL:  testLink.FullURL,
	}
	storage.Save(ctx, domainLink)

	if !storage.IsExistShortKey(ctx, testLink.ShortKey) {
		t.Error("Link was not saved in memory")
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Storage file was not created after Save")
	}

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read storage file: %v", err)
	}

	var savedLinks []adapter.Link
	if err := json.Unmarshal(fileData, &savedLinks); err != nil {
		t.Fatalf("Failed to unmarshal saved data: %v", err)
	}

	if len(savedLinks) != 1 {
		t.Errorf("Expected 1 link in file, got %d", len(savedLinks))
	}

	if savedLinks[0].ShortKey != testLink.ShortKey {
		t.Errorf("Expected short key '%s', got '%s'", testLink.ShortKey, savedLinks[0].ShortKey)
	}

	if savedLinks[0].FullURL != testLink.FullURL {
		t.Errorf("Expected URL '%s', got '%s'", testLink.FullURL, savedLinks[0].FullURL)
	}
}

func TestLinkStorage_Save_GeneratesUUID(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "uuid-test.json")

	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}

	ctx := context.Background()

	domainLink := &model.Link{
		ShortKey: "uuid-test-key",
		FullURL:  "https://uuid-test.com",
	}

	storage.Save(ctx, domainLink)

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read storage file: %v", err)
	}

	var savedLinks []adapter.Link
	if err := json.Unmarshal(fileData, &savedLinks); err != nil {
		t.Fatalf("Failed to unmarshal saved data: %v", err)
	}

	if len(savedLinks) != 1 {
		t.Fatalf("Expected 1 link in file, got %d", len(savedLinks))
	}

	if savedLinks[0].UUID == "" {
		t.Error("UUID was not generated for link without UUID")
	}

	if len(savedLinks[0].UUID) != 36 {
		t.Errorf("UUID has incorrect length: expected 36, got %d", len(savedLinks[0].UUID))
	}
}

func TestLinkStorage_GetByShortKey(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")
	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}
	ctx := context.Background()

	testCases := []struct {
		description string
		link        *adapter.Link
		shortKey    string
		expectedURL string
	}{
		{
			description: "Existing short key",
			link: &adapter.Link{
				ShortKey: "abc123",
				FullURL:  "https://example.com",
			},
			shortKey:    "abc123",
			expectedURL: "https://example.com",
		},
		{
			description: "Non-existing short key",
			link:        nil,
			shortKey:    "nonexistent",
			expectedURL: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.link != nil {
				domainLink := &model.Link{
					ShortKey: testCase.link.ShortKey,
					FullURL:  testCase.link.FullURL,
				}
				storage.Save(ctx, domainLink)
			}

			result, _ := storage.GetByShortKey(ctx, testCase.shortKey)
			if result != testCase.expectedURL {
				t.Errorf("Expected URL '%s', got '%s'", testCase.expectedURL, result)
			}
		})
	}
}

func TestLinkStorage_IsExistShortKey(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")
	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}
	ctx := context.Background()

	testLink := &adapter.Link{
		ShortKey: "existing-key",
		FullURL:  "https://example.com",
	}

	domainLink := &model.Link{
		ShortKey: testLink.ShortKey,
		FullURL:  testLink.FullURL,
	}
	storage.Save(ctx, domainLink)

	testCases := []struct {
		description string
		shortKey    string
		expected    bool
	}{
		{
			description: "Existing short key",
			shortKey:    "existing-key",
			expected:    true,
		},
		{
			description: "Non-existing short key",
			shortKey:    "nonexistent-key",
			expected:    false,
		},
		{
			description: "Empty short key",
			shortKey:    "",
			expected:    false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			result := storage.IsExistShortKey(ctx, testCase.shortKey)
			if result != testCase.expected {
				t.Errorf("Expected %v, got %v for short key '%s'", testCase.expected, result, testCase.shortKey)
			}
		})
	}
}

func TestLinkStorage_GetShortKeyByURL(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		description   string
		link          *adapter.Link
		url           string
		expectedKey   string
		normalizeTest bool
	}{
		{
			description: "Existing URL",
			link: &adapter.Link{
				ShortKey: "key1",
				FullURL:  "https://example.com",
			},
			url:         "https://example.com",
			expectedKey: "key1",
		},
		{
			description: "URL with trailing slash",
			link: &adapter.Link{
				ShortKey: "key2",
				FullURL:  "https://example.com/",
			},
			url:           "https://example.com",
			expectedKey:   "key2",
			normalizeTest: true,
		},
		{
			description: "Non-existing URL",
			link:        nil,
			url:         "https://nonexistent.com",
			expectedKey: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			tempDir := t.TempDir()
			testFilePath := filepath.Join(tempDir, "test.json")
			testStorage, err := NewInMemoryFileStorage(testFilePath)
			if err != nil {
				t.Fatalf("NewInMemoryFileStorage failed: %v", err)
			}
			if testCase.link != nil {
				normalizedFullURL := common.NormalizeURL(testCase.link.FullURL)
				domainLink := &model.Link{
					ShortKey: testCase.link.ShortKey,
					FullURL:  normalizedFullURL,
				}
				testStorage.Save(ctx, domainLink)
			}

			normalizedSearchURL := common.NormalizeURL(testCase.url)
			result := testStorage.GetShortKeyByURL(ctx, normalizedSearchURL)
			if result != testCase.expectedKey {
				t.Errorf("Expected short key '%s', got '%s' for URL '%s'", testCase.expectedKey, result, normalizedSearchURL)
			}
		})
	}
}

func TestLinkStorage_GetAllLinks(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")
	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}
	ctx := context.Background()

	allLinks := storage.getAllLinks()
	if len(allLinks) != 0 {
		t.Errorf("Expected 0 links in empty storage, got %d", len(allLinks))
	}

	links := []*adapter.Link{
		{
			ShortKey: "key1",
			FullURL:  "https://example1.com",
		},
		{
			ShortKey: "key2",
			FullURL:  "https://example2.com",
		},
		{
			ShortKey: "key3",
			FullURL:  "https://example3.com",
		},
	}

	for _, link := range links {
		domainLink := &model.Link{
			ShortKey: link.ShortKey,
			FullURL:  link.FullURL,
		}
		storage.Save(ctx, domainLink)
	}

	allLinks = storage.getAllLinks()
	if len(allLinks) != len(links) {
		t.Errorf("Expected %d links, got %d", len(links), len(allLinks))
	}

	linkMap := make(map[string]*adapter.Link)
	for _, link := range allLinks {
		linkMap[link.ShortKey] = link
	}

	for _, expectedLink := range links {
		foundLink, exists := linkMap[expectedLink.ShortKey]
		if !exists {
			t.Errorf("Link with short key '%s' not found in GetAllLinks result", expectedLink.ShortKey)
			continue
		}

		if foundLink.FullURL != expectedLink.FullURL {
			t.Errorf("Expected URL '%s' for key '%s', got '%s'", expectedLink.FullURL, expectedLink.ShortKey, foundLink.FullURL)
		}

		if foundLink.UUID == "" {
			t.Errorf("UUID was not generated for key '%s'", expectedLink.ShortKey)
		}
	}
}

func TestLinkStorage_MultipleSaves(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")
	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}
	ctx := context.Background()

	firstLink := &adapter.Link{
		ShortKey: "same-key",
		FullURL:  "https://first.com",
	}
	firstDomainLink := &model.Link{
		ShortKey: firstLink.ShortKey,
		FullURL:  firstLink.FullURL,
	}
	storage.Save(ctx, firstDomainLink)

	secondLink := &adapter.Link{
		ShortKey: "same-key",
		FullURL:  "https://second.com",
	}
	secondDomainLink := &model.Link{
		ShortKey: secondLink.ShortKey,
		FullURL:  secondLink.FullURL,
	}
	storage.Save(ctx, secondDomainLink)

	retrievedURL, _ := storage.GetByShortKey(ctx, "same-key")
	if retrievedURL != secondLink.FullURL {
		t.Errorf("Expected URL '%s' after overwrite, got '%s'", secondLink.FullURL, retrievedURL)
	}
}

func TestLinkStorage_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")
	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}
	ctx := context.Background()

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			link := &adapter.Link{
				ShortKey: "key-concurrent",
				FullURL:  "https://concurrent.com",
			}
			domainLink := &model.Link{
				ShortKey: link.ShortKey,
				FullURL:  link.FullURL,
			}
			storage.Save(ctx, domainLink)
			_, _ = storage.GetByShortKey(ctx, "key-concurrent")
			storage.IsExistShortKey(ctx, "key-concurrent")
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if !storage.IsExistShortKey(ctx, "key-concurrent") {
		t.Error("Link was not saved during concurrent access")
	}
}

func TestLinkStorage_Persistence(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "persistence-test.json")

	firstStorage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}

	ctx := context.Background()

	testLinks := []*adapter.Link{
		{
			UUID:     "uuid-persist-1",
			ShortKey: "persist-key-1",
			FullURL:  "https://persist1.com",
		},
		{
			UUID:     "uuid-persist-2",
			ShortKey: "persist-key-2",
			FullURL:  "https://persist2.com",
		},
	}

	for _, link := range testLinks {
		domainLink := &model.Link{
			ShortKey: link.ShortKey,
			FullURL:  link.FullURL,
		}
		firstStorage.Save(ctx, domainLink)
	}

	secondStorage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed on second initialization: %v", err)
	}

	for _, expectedLink := range testLinks {
		if !secondStorage.IsExistShortKey(ctx, expectedLink.ShortKey) {
			t.Errorf("Link with key '%s' was not restored after restart", expectedLink.ShortKey)
		}

		retrievedURL, _ := secondStorage.GetByShortKey(ctx, expectedLink.ShortKey)
		if retrievedURL != expectedLink.FullURL {
			t.Errorf("Expected URL '%s' for key '%s', got '%s'", expectedLink.FullURL, expectedLink.ShortKey, retrievedURL)
		}
	}
}

func TestLinkStorage_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "empty-test.json")

	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}
	defer func(file *os.File) {
		if closeError := file.Close(); closeError != nil {
			t.Fatalf("Failed to close file")
		}
	}(file)

	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage should handle empty file: %v", err)
	}

	ctx := context.Background()

	allLinks := storage.getAllLinks()
	if len(allLinks) != 0 {
		t.Errorf("Expected 0 links in empty file, got %d", len(allLinks))
	}

	testLink := &adapter.Link{
		ShortKey: "empty-key",
		FullURL:  "https://empty-test.com",
	}

	domainLink := &model.Link{
		ShortKey: testLink.ShortKey,
		FullURL:  testLink.FullURL,
	}
	storage.Save(ctx, domainLink)

	if !storage.IsExistShortKey(ctx, "empty-key") {
		t.Error("Failed to save link after loading from empty file")
	}
}

func TestGenerateUUID(t *testing.T) {
	uuidSet := make(map[string]bool)
	uuidCount := 100

	for i := 0; i < uuidCount; i++ {
		generatedUUID := uuid.New().String()

		if len(generatedUUID) != 36 {
			t.Errorf("UUID has incorrect length: expected 36, got %d", len(generatedUUID))
		}

		if uuidSet[generatedUUID] {
			t.Errorf("Duplicate UUID generated: %s", generatedUUID)
		}
		uuidSet[generatedUUID] = true
	}

	if len(uuidSet) != uuidCount {
		t.Errorf("Expected %d unique UUIDs, got %d", uuidCount, len(uuidSet))
	}
}

func TestLinkStorage_BatchSave_EmptySlice(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "batch-test.json")
	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}
	ctx := context.Background()

	storage.BatchSave(ctx, []*model.Link{})

	allLinks := storage.getAllLinks()
	if len(allLinks) != 0 {
		t.Errorf("Expected 0 links after empty batch save, got %d", len(allLinks))
	}
}

func TestLinkStorage_BatchSave_SingleLink(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "batch-single-test.json")
	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}
	ctx := context.Background()

	testLink := &model.Link{
		ShortKey: "batch-key-1",
		FullURL:  "https://batch1.com",
	}

	storage.BatchSave(ctx, []*model.Link{testLink})

	if !storage.IsExistShortKey(ctx, "batch-key-1") {
		t.Error("Link was not saved in batch")
	}

	retrievedURL, _ := storage.GetByShortKey(ctx, "batch-key-1")
	if retrievedURL != testLink.FullURL {
		t.Errorf("Expected URL '%s', got '%s'", testLink.FullURL, retrievedURL)
	}
}

func TestLinkStorage_BatchSave_MultipleLinks(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "batch-multiple-test.json")
	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}
	ctx := context.Background()

	testLinks := []*model.Link{
		{
			ShortKey: "batch-key-1",
			FullURL:  "https://batch1.com",
		},
		{
			ShortKey: "batch-key-2",
			FullURL:  "https://batch2.com",
		},
		{
			ShortKey: "batch-key-3",
			FullURL:  "https://batch3.com",
		},
	}

	storage.BatchSave(ctx, testLinks)

	allLinks := storage.getAllLinks()
	if len(allLinks) != len(testLinks) {
		t.Errorf("Expected %d links, got %d", len(testLinks), len(allLinks))
	}

	for _, expectedLink := range testLinks {
		if !storage.IsExistShortKey(ctx, expectedLink.ShortKey) {
			t.Errorf("Link with key '%s' was not saved", expectedLink.ShortKey)
		}

		retrievedURL, _ := storage.GetByShortKey(ctx, expectedLink.ShortKey)
		if retrievedURL != expectedLink.FullURL {
			t.Errorf("Expected URL '%s' for key '%s', got '%s'", expectedLink.FullURL, expectedLink.ShortKey, retrievedURL)
		}
	}
}

func TestLinkStorage_BatchSave_PersistsToFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "batch-persist-test.json")
	firstStorage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}
	ctx := context.Background()

	testLinks := []*model.Link{
		{
			ShortKey: "batch-persist-1",
			FullURL:  "https://batch-persist1.com",
		},
		{
			ShortKey: "batch-persist-2",
			FullURL:  "https://batch-persist2.com",
		},
	}

	firstStorage.BatchSave(ctx, testLinks)

	secondStorage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed on second initialization: %v", err)
	}

	for _, expectedLink := range testLinks {
		if !secondStorage.IsExistShortKey(ctx, expectedLink.ShortKey) {
			t.Errorf("Link with key '%s' was not restored after restart", expectedLink.ShortKey)
		}

		retrievedURL, _ := secondStorage.GetByShortKey(ctx, expectedLink.ShortKey)
		if retrievedURL != expectedLink.FullURL {
			t.Errorf("Expected URL '%s' for key '%s', got '%s'", expectedLink.FullURL, expectedLink.ShortKey, retrievedURL)
		}
	}
}

func TestLinkStorage_GetByUserID_EmptyResult(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "get-by-user-id-empty.json")
	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}
	ctx := context.Background()

	links, err := storage.GetByUserID(ctx, "non-existent-user-id")
	if err != nil {
		t.Fatalf("GetByUserID should not return error, got: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("Expected 0 links for non-existent user, got %d", len(links))
	}
}

func TestLinkStorage_GetByUserID_SingleUser(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "get-by-user-id-single.json")
	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}
	ctx := context.Background()

	userID := "user-123"
	testLinks := []*model.Link{
		{
			ShortKey: "key1",
			FullURL:  "https://example1.com",
			UserID:   userID,
		},
		{
			ShortKey: "key2",
			FullURL:  "https://example2.com",
			UserID:   userID,
		},
		{
			ShortKey: "key3",
			FullURL:  "https://example3.com",
			UserID:   userID,
		},
	}

	for _, link := range testLinks {
		if err := storage.Save(ctx, link); err != nil {
			t.Fatalf("Failed to save link: %v", err)
		}
	}

	links, err := storage.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByUserID should not return error, got: %v", err)
	}
	if len(links) != len(testLinks) {
		t.Errorf("Expected %d links, got %d", len(testLinks), len(links))
	}

	linkMap := make(map[string]*model.Link)
	for _, link := range links {
		linkMap[link.ShortKey] = link
	}

	for _, expectedLink := range testLinks {
		foundLink, exists := linkMap[expectedLink.ShortKey]
		if !exists {
			t.Errorf("Link with key '%s' not found", expectedLink.ShortKey)
			continue
		}
		if foundLink.FullURL != expectedLink.FullURL {
			t.Errorf("Expected URL '%s' for key '%s', got '%s'", expectedLink.FullURL, expectedLink.ShortKey, foundLink.FullURL)
		}
		if foundLink.UserID != userID {
			t.Errorf("Expected UserID '%s' for key '%s', got '%s'", userID, expectedLink.ShortKey, foundLink.UserID)
		}
	}
}

func TestLinkStorage_GetByUserID_MultipleUsers(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "get-by-user-id-multiple.json")
	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}
	ctx := context.Background()

	user1ID := "user-1"
	user2ID := "user-2"

	user1Links := []*model.Link{
		{
			ShortKey: "user1-key1",
			FullURL:  "https://user1-example1.com",
			UserID:   user1ID,
		},
		{
			ShortKey: "user1-key2",
			FullURL:  "https://user1-example2.com",
			UserID:   user1ID,
		},
	}

	user2Links := []*model.Link{
		{
			ShortKey: "user2-key1",
			FullURL:  "https://user2-example1.com",
			UserID:   user2ID,
		},
	}

	for _, link := range user1Links {
		if err := storage.Save(ctx, link); err != nil {
			t.Fatalf("Failed to save user1 link: %v", err)
		}
	}

	for _, link := range user2Links {
		if err := storage.Save(ctx, link); err != nil {
			t.Fatalf("Failed to save user2 link: %v", err)
		}
	}

	user1Result, err := storage.GetByUserID(ctx, user1ID)
	if err != nil {
		t.Fatalf("GetByUserID should not return error for user1, got: %v", err)
	}
	if len(user1Result) != len(user1Links) {
		t.Errorf("Expected %d links for user1, got %d", len(user1Links), len(user1Result))
	}

	user2Result, err := storage.GetByUserID(ctx, user2ID)
	if err != nil {
		t.Fatalf("GetByUserID should not return error for user2, got: %v", err)
	}
	if len(user2Result) != len(user2Links) {
		t.Errorf("Expected %d links for user2, got %d", len(user2Links), len(user2Result))
	}

	// Verify user1 links don't contain user2 links
	for _, link := range user1Result {
		if link.UserID != user1ID {
			t.Errorf("User1 result contains link with wrong UserID: %s", link.UserID)
		}
	}

	// Verify user2 links don't contain user1 links
	for _, link := range user2Result {
		if link.UserID != user2ID {
			t.Errorf("User2 result contains link with wrong UserID: %s", link.UserID)
		}
	}
}

func TestLinkStorage_GetByUserID_WithEmptyUserID(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "get-by-user-id-empty-id.json")
	storage, err := NewInMemoryFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewInMemoryFileStorage failed: %v", err)
	}
	ctx := context.Background()

	// Save link with empty UserID
	link := &model.Link{
		ShortKey: "empty-user-key",
		FullURL:  "https://empty-user.com",
		UserID:   "",
	}
	if err := storage.Save(ctx, link); err != nil {
		t.Fatalf("Failed to save link: %v", err)
	}

	links, err := storage.GetByUserID(ctx, "")
	if err != nil {
		t.Fatalf("GetByUserID should not return error, got: %v", err)
	}
	if len(links) != 1 {
		t.Errorf("Expected 1 link with empty UserID, got %d", len(links))
	}
}
