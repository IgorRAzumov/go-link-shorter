package database

import (
	"context"
	"os"
	"testing"
	"time"
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

	if err := storage.Close(); err != nil {
		t.Fatalf("Failed to close storage: %v", err)
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

	if err := storage.Close(); err != nil {
		t.Fatalf("First Close failed: %v", err)
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

	if err := storage.Close(); err != nil {
		t.Fatalf("Failed to close storage: %v", err)
	}

	result, err = storage.CheckStorageConnection(ctx)
	if err == nil {
		t.Error("CheckStorageConnection should return error after close")
	}
	if result {
		t.Error("CheckStorageConnection should return false after close")
	}
}
