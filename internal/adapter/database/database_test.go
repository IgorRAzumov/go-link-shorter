package database

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/rs/zerolog/log"
)

func TestNewStorage(t *testing.T) {
	databaseDSN := os.Getenv("TEST_DATABASE_DSN")
	if databaseDSN == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping integration test")
	}

	storage, err := NewStorage(databaseDSN)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	if storage == nil {
		t.Fatal("NewStorage returned nil")
	}
	defer func(storage *Storage) {
		if closeError := storage.Close(); closeError != nil {
			t.Fatalf("Failed to close storage")
		}
	}(storage)
}

func TestNewStorage_WithEmptyDSN(t *testing.T) {
	storage, err := NewStorage("")
	if err != nil {
		t.Fatalf("NewStorage should not fail with empty DSN: %v", err)
	}
	if storage == nil {
		t.Fatal("NewStorage returned nil")
	}
	if storage.db != nil {
		t.Error("Storage.db should be nil when DSN is empty")
	}
}

func TestNewStorage_WithInvalidDSN(t *testing.T) {
	storage, err := NewStorage("invalid-dsn-format")
	if err == nil {
		if storage != nil && storage.db != nil {
			if closeError := storage.Close(); closeError != nil {
				t.Fatalf("Failed to close storage")
			}
		}
		t.Fatal("NewStorage should return error for invalid DSN")
	}
	if storage != nil && storage.db != nil {
		if closeError := storage.Close(); closeError != nil {
			t.Fatalf("Failed to close storage")
		}
	}
}

func TestStorage_CheckStorageConnection(t *testing.T) {
	databaseDSN := os.Getenv("TEST_DATABASE_DSN")
	if databaseDSN == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping integration test")
	}

	storage, err := NewStorage(databaseDSN)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer func(storage *Storage) {
		if closeError := storage.Close(); closeError != nil {
			t.Fatalf("Failed to close storage")
		}
	}(storage)

	ctx := context.Background()
	result, err := storage.CheckStorageConnection(ctx)
	if err != nil {
		t.Fatalf("CheckStorageConnection failed: %v", err)
	}
	if !result {
		t.Error("CheckStorageConnection should return true for valid connection")
	}
}

func TestStorage_CheckStorageConnection_WithNilDB(t *testing.T) {
	storage := &Storage{db: nil}
	ctx := context.Background()

	result, err := storage.CheckStorageConnection(ctx)
	if err != nil {
		t.Errorf("CheckStorageConnection should not return error with nil db, got: %v", err)
	}
	if result {
		t.Error("CheckStorageConnection should return false with nil db")
	}
}

func TestStorage_CheckStorageConnection_WithClosedConnection(t *testing.T) {
	databaseDSN := os.Getenv("TEST_DATABASE_DSN")
	if databaseDSN == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping integration test")
	}

	storage, err := NewStorage(databaseDSN)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}

	if closeErr := storage.Close(); closeErr != nil {
		t.Fatalf("Failed to close storage: %v", closeErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	result, err := storage.CheckStorageConnection(ctx)
	if err == nil {
		t.Error("CheckStorageConnection should return error for closed connection")
	}
	if result {
		t.Error("CheckStorageConnection should return false for closed connection")
	}
}

func TestStorage_CheckStorageConnection_WithContextTimeout(t *testing.T) {
	databaseDSN := os.Getenv("TEST_DATABASE_DSN")
	if databaseDSN == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping integration test")
	}

	storage, err := NewStorage(databaseDSN)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer func(storage *Storage) {
		if closeError := storage.Close(); closeError != nil {
			t.Fatalf("Failed to close storage")
		}
	}(storage)

	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	result, err := storage.CheckStorageConnection(ctx)
	if err == nil {
		t.Error("CheckStorageConnection should return error for cancelled context")
	}
	if result {
		t.Error("CheckStorageConnection should return false for cancelled context")
	}
}

func TestStorage_Close(t *testing.T) {
	databaseDSN := os.Getenv("TEST_DATABASE_DSN")
	if databaseDSN == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping integration test")
	}

	storage, err := NewStorage(databaseDSN)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}

	err = storage.Close()
	if err != nil {
		t.Errorf("Close should not return error, got: %v", err)
	}
}

func TestStorage_Close_WithNilDB(t *testing.T) {
	storage := &Storage{db: nil}
	err := storage.Close()
	if err != nil {
		t.Errorf("Close should not return error with nil db, got: %v", err)
	}
}

func TestStorage_Close_AlreadyClosed(t *testing.T) {
	databaseDSN := os.Getenv("TEST_DATABASE_DSN")
	if databaseDSN == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping integration test")
	}

	storage, err := NewStorage(databaseDSN)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}

	if closeErr := storage.Close(); closeErr != nil {
		t.Fatalf("First Close failed: %v", closeErr)
	}

	err = storage.Close()
	if err == nil {
		t.Error("Close on already closed connection should return error")
	}
}

func TestStorage_CheckStorageConnection_AfterClose(t *testing.T) {
	databaseDSN := os.Getenv("TEST_DATABASE_DSN")
	if databaseDSN == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping integration test")
	}

	storage, err := NewStorage(databaseDSN)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}

	ctx := context.Background()
	result, err := storage.CheckStorageConnection(ctx)
	if err != nil || !result {
		t.Fatalf("CheckStorageConnection should succeed before close, got result=%v, err=%v", result, err)
	}

	if closeErr := storage.Close(); closeErr != nil {
		t.Fatalf("Failed to close storage: %v", closeErr)
	}

	result, err = storage.CheckStorageConnection(ctx)
	if err == nil {
		t.Error("CheckStorageConnection should return error after close")
	}
	if result {
		t.Error("CheckStorageConnection should return false after close")
	}
}

func TestStorage_BatchSave_EmptySlice(t *testing.T) {
	databaseDSN := os.Getenv("TEST_DATABASE_DSN")
	if databaseDSN == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping integration test")
	}

	storage, err := NewStorage(databaseDSN)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer func(storage *Storage) {
		if closeError := storage.Close(); closeError != nil {
			t.Fatalf("Failed to close storage")
		}
	}(storage)

	ctx := context.Background()
	if err := storage.BatchSave(ctx, []*model.Link{}); err != nil {
		log.Error().Err(err).Msg("BatchSave failed with empty slice")
	}
}

func TestStorage_BatchSave_SingleLink(t *testing.T) {
	databaseDSN := os.Getenv("TEST_DATABASE_DSN")
	if databaseDSN == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping integration test")
	}

	storage, err := NewStorage(databaseDSN)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer func(storage *Storage) {
		if closeError := storage.Close(); closeError != nil {
			t.Fatalf("Failed to close storage")
		}
	}(storage)

	ctx := context.Background()
	testLink := &model.Link{
		ShortKey: "batch-test-key-1",
		FullURL:  "https://batch-test1.com",
	}

	if err := storage.BatchSave(ctx, []*model.Link{testLink}); err != nil {
		log.Error().Err(err).Msg("BatchSave failed for single link")
	}

	if !storage.IsExistShortKey(ctx, "batch-test-key-1") {
		t.Error("Link was not saved in batch")
	}

	retrievedURL, _ := storage.GetByShortKey(ctx, "batch-test-key-1")
	if retrievedURL != testLink.FullURL {
		t.Errorf("Expected URL '%s', got '%s'", testLink.FullURL, retrievedURL)
	}
}

func TestStorage_BatchSave_MultipleLinks(t *testing.T) {
	databaseDSN := os.Getenv("TEST_DATABASE_DSN")
	if databaseDSN == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping integration test")
	}

	storage, err := NewStorage(databaseDSN)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer func(storage *Storage) {
		if closeError := storage.Close(); closeError != nil {
			t.Fatalf("Failed to close storage")
		}
	}(storage)

	ctx := context.Background()
	testLinks := []*model.Link{
		{
			ShortKey: "batch-test-key-2",
			FullURL:  "https://batch-test2.com",
		},
		{
			ShortKey: "batch-test-key-3",
			FullURL:  "https://batch-test3.com",
		},
		{
			ShortKey: "batch-test-key-4",
			FullURL:  "https://batch-test4.com",
		},
	}

	if err := storage.BatchSave(ctx, testLinks); err != nil {
		log.Error().Err(err).Msg("BatchSave failed for multiple links")
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

func TestStorage_BatchSave_WithNilDB(t *testing.T) {
	storage := &Storage{}

	ctx := context.Background()
	testLinks := []*model.Link{
		{
			ShortKey: "test-key",
			FullURL:  "https://test.com",
		},
	}

	if err := storage.BatchSave(ctx, testLinks); err != nil {
		log.Error().Err(err).Msg("BatchSave failed with nil db")
	}

	if storage.db != nil {
		t.Error("BatchSave should not fail with nil db")
	}
}

func TestStorage_Save_WithURLConflict(t *testing.T) {
	databaseDSN := os.Getenv("TEST_DATABASE_DSN")
	if databaseDSN == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping integration test")
	}

	storage, err := NewStorage(databaseDSN)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer func(storage *Storage) {
		if closeError := storage.Close(); closeError != nil {
			t.Fatalf("Failed to close storage")
		}
	}(storage)

	ctx := context.Background()
	testURL := "https://conflict-test.com"
	firstShortKey := "first-conflict-key"
	secondShortKey := "second-conflict-key"

	// Save first link
	firstLink := &model.Link{
		ShortKey: firstShortKey,
		FullURL:  testURL,
	}
	err = storage.Save(ctx, firstLink)
	if err != nil {
		t.Fatalf("Failed to save first link: %v", err)
	}

	// Try to save second link with same URL but different short_key
	secondLink := &model.Link{
		ShortKey: secondShortKey,
		FullURL:  testURL,
	}
	err = storage.Save(ctx, secondLink)
	if err == nil {
		t.Error("Expected URLConflictError when saving duplicate URL, got nil")
	}

	// Check that error is URLConflictError with correct existing short key
	var conflictErr *model.URLConflictError
	if !errors.As(err, &conflictErr) {
		t.Fatalf("Expected URLConflictError, got %T: %v", err, err)
	}

	if conflictErr.ExistingShortKey != firstShortKey {
		t.Errorf("Expected existing short key '%s', got '%s'", firstShortKey, conflictErr.ExistingShortKey)
	}

	// Verify that first link still exists and second link was not saved
	if !storage.IsExistShortKey(ctx, firstShortKey) {
		t.Error("First link should still exist")
	}

	if storage.IsExistShortKey(ctx, secondShortKey) {
		t.Error("Second link should not have been saved")
	}

	// Verify that GetShortKeyByURL returns the first short key
	retrievedShortKey := storage.GetShortKeyByURL(ctx, testURL)
	if retrievedShortKey != firstShortKey {
		t.Errorf("Expected short key '%s', got '%s'", firstShortKey, retrievedShortKey)
	}
}

func TestStorage_BatchSave_WithURLConflict(t *testing.T) {
	databaseDSN := os.Getenv("TEST_DATABASE_DSN")
	if databaseDSN == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping integration test")
	}

	storage, err := NewStorage(databaseDSN)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer func(storage *Storage) {
		if closeError := storage.Close(); closeError != nil {
			t.Fatalf("Failed to close storage")
		}
	}(storage)

	ctx := context.Background()
	testURL := "https://batch-conflict-test.com"
	firstShortKey := "batch-first-conflict-key"
	secondShortKey := "batch-second-conflict-key"

	// Save first link
	firstLink := &model.Link{
		ShortKey: firstShortKey,
		FullURL:  testURL,
	}
	err = storage.Save(ctx, firstLink)
	if err != nil {
		t.Fatalf("Failed to save first link: %v", err)
	}

	// Try to batch save with conflicting URL
	conflictingLinks := []*model.Link{
		{
			ShortKey: secondShortKey,
			FullURL:  testURL,
		},
	}
	err = storage.BatchSave(ctx, conflictingLinks)
	if err == nil {
		t.Error("Expected URLConflictError when batch saving duplicate URL, got nil")
	}

	// Check that error is URLConflictError with correct existing short key
	var conflictErr *model.URLConflictError
	if !errors.As(err, &conflictErr) {
		t.Fatalf("Expected URLConflictError, got %T: %v", err, err)
	}

	if conflictErr.ExistingShortKey != firstShortKey {
		t.Errorf("Expected existing short key '%s', got '%s'", firstShortKey, conflictErr.ExistingShortKey)
	}

	// Verify that first link still exists and second link was not saved
	if !storage.IsExistShortKey(ctx, firstShortKey) {
		t.Error("First link should still exist")
	}

	if storage.IsExistShortKey(ctx, secondShortKey) {
		t.Error("Second link should not have been saved")
	}
}

func TestStorage_GetByUserID_EmptyResult(t *testing.T) {
	databaseDSN := os.Getenv("TEST_DATABASE_DSN")
	if databaseDSN == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping integration test")
	}

	storage, err := NewStorage(databaseDSN)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer func(storage *Storage) {
		if closeError := storage.Close(); closeError != nil {
			t.Fatalf("Failed to close storage")
		}
	}(storage)

	ctx := context.Background()
	links, err := storage.GetByUserID(ctx, "non-existent-user-id")
	if err != nil {
		t.Fatalf("GetByUserID should not return error, got: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("Expected 0 links for non-existent user, got %d", len(links))
	}
}

func TestStorage_GetByUserID_SingleUser(t *testing.T) {
	databaseDSN := os.Getenv("TEST_DATABASE_DSN")
	if databaseDSN == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping integration test")
	}

	storage, err := NewStorage(databaseDSN)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer func(storage *Storage) {
		if closeError := storage.Close(); closeError != nil {
			t.Fatalf("Failed to close storage")
		}
	}(storage)

	ctx := context.Background()
	userID := "test-user-123"
	testLinks := []*model.Link{
		{
			ShortKey: "getbyuserid-key1",
			FullURL:  "https://getbyuserid1.com",
			UserID:   userID,
		},
		{
			ShortKey: "getbyuserid-key2",
			FullURL:  "https://getbyuserid2.com",
			UserID:   userID,
		},
		{
			ShortKey: "getbyuserid-key3",
			FullURL:  "https://getbyuserid3.com",
			UserID:   userID,
		},
	}

	for _, link := range testLinks {
		if saveErr := storage.Save(ctx, link); saveErr != nil {
			t.Fatalf("Failed to save link: %v", saveErr)
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

func TestStorage_GetByUserID_MultipleUsers(t *testing.T) {
	databaseDSN := os.Getenv("TEST_DATABASE_DSN")
	if databaseDSN == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping integration test")
	}

	storage, err := NewStorage(databaseDSN)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer func(storage *Storage) {
		if closeError := storage.Close(); closeError != nil {
			t.Fatalf("Failed to close storage")
		}
	}(storage)

	ctx := context.Background()
	user1ID := "test-user-1"
	user2ID := "test-user-2"

	user1Links := []*model.Link{
		{
			ShortKey: "multiuser-user1-key1",
			FullURL:  "https://multiuser-user1-1.com",
			UserID:   user1ID,
		},
		{
			ShortKey: "multiuser-user1-key2",
			FullURL:  "https://multiuser-user1-2.com",
			UserID:   user1ID,
		},
	}

	user2Links := []*model.Link{
		{
			ShortKey: "multiuser-user2-key1",
			FullURL:  "https://multiuser-user2-1.com",
			UserID:   user2ID,
		},
	}

	for _, link := range user1Links {
		if saveErr := storage.Save(ctx, link); saveErr != nil {
			t.Fatalf("Failed to save user1 link: %v", saveErr)
		}
	}

	for _, link := range user2Links {
		if saveErr := storage.Save(ctx, link); saveErr != nil {
			t.Fatalf("Failed to save user2 link: %v", saveErr)
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

func TestStorage_GetByUserID_WithNilDB(t *testing.T) {
	storage := &Storage{db: nil}
	ctx := context.Background()

	links, err := storage.GetByUserID(ctx, "test-user-id")
	if err != nil {
		t.Fatalf("GetByUserID should not return error with nil db, got: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("Expected 0 links with nil db, got %d", len(links))
	}
}
