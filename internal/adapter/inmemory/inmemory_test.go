package inmemory

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/adapter"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

func TestNewInMemoryStorage(t *testing.T) {
	storage := NewInMemoryStorage()
	if storage == nil {
		t.Fatal("NewInMemoryStorage returned nil")
	}
}

func TestNewFileStorage(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test-storage.json")

	storage, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewFileStorage failed: %v", err)
	}

	if storage == nil {
		t.Fatal("NewFileStorage returned nil")
	}
}

func TestNewFileStorage_WithNonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "non-existent.json")

	storage, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewFileStorage should not fail for non-existent file: %v", err)
	}

	if storage == nil {
		t.Fatal("NewFileStorage returned nil")
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

	storage, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewFileStorage failed: %v", err)
	}

	ctx := context.Background()

	if !storage.IsExistShortKey(ctx, "key1") {
		t.Error("Link 'key1' was not loaded from file")
	}

	if !storage.IsExistShortKey(ctx, "key2") {
		t.Error("Link 'key2' was not loaded from file")
	}

	retrievedURL := storage.GetByShortKey(ctx, "key1")
	if retrievedURL != "https://example1.com" {
		t.Errorf("Expected URL 'https://example1.com', got '%s'", retrievedURL)
	}
}

func TestLinkStorage_Save(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	testLink := &adapter.Link{
		UUID:     "test-uuid-1",
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

	retrievedURL := storage.GetByShortKey(ctx, testLink.ShortKey)
	if retrievedURL != testLink.FullURL {
		t.Errorf("Expected URL '%s', got '%s'", testLink.FullURL, retrievedURL)
	}
}

func TestLinkStorage_Save_WithFileStorage(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "save-test.json")

	storage, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewFileStorage failed: %v", err)
	}

	ctx := context.Background()

	testLink := &adapter.Link{
		UUID:     "test-uuid-1",
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

	storage, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewFileStorage failed: %v", err)
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
	storage := NewInMemoryStorage()
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
				UUID:     "test-uuid-1",
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

			result := storage.GetByShortKey(ctx, testCase.shortKey)
			if result != testCase.expectedURL {
				t.Errorf("Expected URL '%s', got '%s'", testCase.expectedURL, result)
			}
		})
	}
}

func TestLinkStorage_IsExistShortKey(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	testLink := &adapter.Link{
		UUID:     "test-uuid-1",
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
	storage := NewInMemoryStorage()
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
				UUID:     "test-uuid-1",
				ShortKey: "key1",
				FullURL:  "https://example.com",
			},
			url:         "https://example.com",
			expectedKey: "key1",
		},
		{
			description: "URL with trailing slash",
			link: &adapter.Link{
				UUID:     "test-uuid-2",
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
			if testCase.link != nil {
				domainLink := &model.Link{
					ShortKey: testCase.link.ShortKey,
					FullURL:  testCase.link.FullURL,
				}
				storage.Save(ctx, domainLink)
			}

			result := storage.GetShortKeyByURL(ctx, testCase.url)
			if result != testCase.expectedKey {
				t.Errorf("Expected short key '%s', got '%s' for URL '%s'", testCase.expectedKey, result, testCase.url)
			}
		})
	}
}

func TestLinkStorage_GetAllLinks(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	allLinks := storage.getAllLinks()
	if len(allLinks) != 0 {
		t.Errorf("Expected 0 links in empty storage, got %d", len(allLinks))
	}

	links := []*adapter.Link{
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
		{
			UUID:     "uuid-3",
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
	storage := NewInMemoryStorage()
	ctx := context.Background()

	firstLink := &adapter.Link{
		UUID:     "uuid-1",
		ShortKey: "same-key",
		FullURL:  "https://first.com",
	}
	firstDomainLink := &model.Link{
		ShortKey: firstLink.ShortKey,
		FullURL:  firstLink.FullURL,
	}
	storage.Save(ctx, firstDomainLink)

	secondLink := &adapter.Link{
		UUID:     "uuid-2",
		ShortKey: "same-key",
		FullURL:  "https://second.com",
	}
	secondDomainLink := &model.Link{
		ShortKey: secondLink.ShortKey,
		FullURL:  secondLink.FullURL,
	}
	storage.Save(ctx, secondDomainLink)

	retrievedURL := storage.GetByShortKey(ctx, "same-key")
	if retrievedURL != secondLink.FullURL {
		t.Errorf("Expected URL '%s' after overwrite, got '%s'", secondLink.FullURL, retrievedURL)
	}
}

func TestLinkStorage_ConcurrentAccess(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			link := &adapter.Link{
				UUID:     "uuid-concurrent",
				ShortKey: "key-concurrent",
				FullURL:  "https://concurrent.com",
			}
			domainLink := &model.Link{
				ShortKey: link.ShortKey,
				FullURL:  link.FullURL,
			}
			storage.Save(ctx, domainLink)
			storage.GetByShortKey(ctx, "key-concurrent")
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

	firstStorage, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewFileStorage failed: %v", err)
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

	secondStorage, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewFileStorage failed on second initialization: %v", err)
	}

	for _, expectedLink := range testLinks {
		if !secondStorage.IsExistShortKey(ctx, expectedLink.ShortKey) {
			t.Errorf("Link with key '%s' was not restored after restart", expectedLink.ShortKey)
		}

		retrievedURL := secondStorage.GetByShortKey(ctx, expectedLink.ShortKey)
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

	storage, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("NewFileStorage should handle empty file: %v", err)
	}

	ctx := context.Background()

	allLinks := storage.getAllLinks()
	if len(allLinks) != 0 {
		t.Errorf("Expected 0 links in empty file, got %d", len(allLinks))
	}

	testLink := &adapter.Link{
		UUID:     "uuid-empty",
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
		generatedUUID := generateUUID()

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
