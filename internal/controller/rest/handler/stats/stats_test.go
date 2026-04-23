package stats

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

type mockStatsUsecase struct {
	stats model.ServiceStats
	err   error
}

func (m *mockStatsUsecase) GetStats(_ context.Context) (model.ServiceStats, error) {
	return m.stats, m.err
}

func TestStatisticHandler_OK(t *testing.T) {
	handler := StatisticHandler(&mockStatsUsecase{stats: model.ServiceStats{URLs: 10, Users: 3}})
	req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("expected application/json, got %q", got)
	}

	var resp model.ServiceStats
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.URLs != 10 || resp.Users != 3 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestStatisticHandler_Error(t *testing.T) {
	handler := StatisticHandler(&mockStatsUsecase{err: errors.New("boom")})
	req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}
