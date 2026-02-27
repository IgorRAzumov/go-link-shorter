package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

func TestHTTPObserver_Notify_PostsEvent(t *testing.T) {
	var received model.AuditEvent
	var gotRequest bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequest = true
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("failed to decode body: %v", err)
		}
		_ = r.Body.Close()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	observer := NewHTTPAuditObserver(server.URL)
	event := model.AuditEvent{
		Timestamp: 1700000001,
		Action:    "follow",
		UserID:    "u-2",
		URL:       "https://example.com/long",
	}

	if err := observer.Save(context.Background(), event); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if !gotRequest {
		t.Fatalf("expected server to receive request")
	}
	if received != event {
		t.Fatalf("received event mismatch: %#v", received)
	}
}
