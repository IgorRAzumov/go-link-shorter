package shorter

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	mocktesting "github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common/testing"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
)

func TestUserURLsDeleteHandler_Returns401WhenNoUserID(t *testing.T) {
	mockUsecase := &mocktesting.MockLinkDeleteUsecase{}
	handler := UserURLsDeleteHandler(mockUsecase)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader([]byte(`["a"]`)))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestUserURLsDeleteHandler_Returns400OnInvalidJSON(t *testing.T) {
	mockUsecase := &mocktesting.MockLinkDeleteUsecase{}
	handler := UserURLsDeleteHandler(mockUsecase)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader([]byte(`not-json`)))
	req = req.WithContext(authctx.WithUserID(req.Context(), "test-user-id"))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestUserURLsDeleteHandler_Returns202AndCallsUsecase(t *testing.T) {
	var got []string
	mockUsecase := &mocktesting.MockLinkDeleteUsecase{
		DeleteUserURLsFunc: func(ctx context.Context, shortKeys []string) error {
			got = append([]string(nil), shortKeys...)
			return nil
		},
	}
	handler := UserURLsDeleteHandler(mockUsecase)

	want := []string{"6qxTVvsy", "RTfd56hn", "Jlfd67ds"}
	body, _ := json.Marshal(want)
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
	req = req.WithContext(authctx.WithUserID(req.Context(), "test-user-id"))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("Expected status %d, got %d", http.StatusAccepted, rr.Code)
	}
	if len(got) != len(want) {
		t.Fatalf("Expected %d ids, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Expected id[%d]=%q, got %q", i, want[i], got[i])
		}
	}
}
