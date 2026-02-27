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

	observer := NewFileObserver(tmpFile.Name())
	event := model.AuditEvent{
		Timestamp: 1700000000,
		Action:    "shorten",
		UserID:    "u-1",
		URL:       "https://example.com/path",
	}

	if err := observer.Save(context.Background(), event); err != nil {
		t.Fatalf("Save returned error: %v", err)
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
	if err := json.Unmarshal([]byte(lines[0]), &decoded); err != nil {
		t.Fatalf("failed to unmarshal json line: %v", err)
	}
	if decoded != event {
		t.Fatalf("decoded event mismatch: %#v", decoded)
	}
}
