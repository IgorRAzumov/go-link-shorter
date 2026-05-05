package shorter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	mocktesting "github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common/testing"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

func TestUserURLsHandler_Returns401WhenNoUserID(t *testing.T) {
	mockUsecase := &mocktesting.MockLinkReadUsecase{}
	handler := UserURLsHandler(mockUsecase)

	req := httptest.NewRequest("GET", "/api/user/urls", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestUserURLsHandler_Returns204WhenNoURLs(t *testing.T) {
	mockUsecase := &mocktesting.MockLinkReadUsecase{
		GetUserURLsFunc: func(ctx context.Context) ([]*model.Link, error) {
			return []*model.Link{}, nil
		},
	}
	handler := UserURLsHandler(mockUsecase)

	req := httptest.NewRequest("GET", "/api/user/urls", nil)
	ctx := authctx.WithAuthenticated(authctx.WithUserID(req.Context(), "test-user-id"))
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}

func TestUserURLsHandler_Returns200WithURLs(t *testing.T) {
	testLinks := []*model.Link{
		{ShortKey: "abc123", FullURL: "http://example.com", UserID: "test-user-id"},
		{ShortKey: "def456", FullURL: "http://example.org", UserID: "test-user-id"},
	}

	mockUsecase := &mocktesting.MockLinkReadUsecase{
		GetUserURLsFunc: func(ctx context.Context) ([]*model.Link, error) {
			return testLinks, nil
		},
		GetBaseURLFunc: func() string {
			return "http://localhost:8080"
		},
	}
	handler := UserURLsHandler(mockUsecase)

	req := httptest.NewRequest("GET", "/api/user/urls", nil)
	ctx := authctx.WithAuthenticated(authctx.WithUserID(req.Context(), "test-user-id"))
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var responses []UserURLResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &responses); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(responses) != 2 {
		t.Errorf("Expected 2 responses, got %d", len(responses))
	}

	if responses[0].OriginalURL != "http://example.com" {
		t.Errorf("Expected original URL http://example.com, got %s", responses[0].OriginalURL)
	}

	if responses[0].ShortURL != "http://localhost:8080/abc123" {
		t.Errorf("Expected short URL http://localhost:8080/abc123, got %s", responses[0].ShortURL)
	}
}
