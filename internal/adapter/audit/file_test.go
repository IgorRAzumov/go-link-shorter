package audit

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

func TestFileObserver_Notify_AppendsJSONLine(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audit-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = tmpFile.Close()

	observer, closeFn, err := NewFileObserver(tmpFile.Name())
	if err != nil {
		t.Fatalf("NewFileObserver failed: %v", err)
	}
	defer func() { _ = closeFn() }()

	event := model.AuditEvent{
		Timestamp: 1700000000,
		Action:    "shorten",
		UserID:    "u-1",
		URL:       "https://example.com/path",
	}

	if saveErr := observer.Save(context.Background(), event); saveErr != nil {
		t.Fatalf("Save returned error: %v", saveErr)
	}

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}

	var decoded model.AuditEvent
	if unmarshalErr := json.Unmarshal([]byte(lines[0]), &decoded); unmarshalErr != nil {
		t.Fatalf("failed to unmarshal json line: %v", unmarshalErr)
	}
	if decoded != event {
		t.Fatalf("decoded event mismatch: %#v", decoded)
	}
}

func TestFileObserver_Close_SaveAfterCloseIgnored(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audit-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = tmpFile.Close()

	observer, closeFn, err := NewFileObserver(tmpFile.Name())
	if err != nil {
		t.Fatalf("NewFileObserver failed: %v", err)
	}

	if closeErr := closeFn(); closeErr != nil {
		t.Fatalf("Close failed: %v", closeErr)
	}

	// Save после Close не должен паниковать
	err = observer.Save(context.Background(), model.AuditEvent{
		Timestamp: 1700000000,
		Action:    "shorten",
		UserID:    "u-1",
		URL:       "https://example.com",
	})
	if err != nil {
		t.Fatalf("Save after Close should return nil, got: %v", err)
	}
}
