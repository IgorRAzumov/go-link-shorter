package resolver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	commontesting "github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common/testing"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/go-chi/chi/v5"
)

type recordingAuditor struct {
	events []model.AuditEvent
}

func (r *recordingAuditor) AuditNewEvent(_ context.Context, e model.AuditEvent) {
	r.events = append(r.events, e)
}

func TestResolveHandler_EmptyPath(t *testing.T) {
	mockUsecase := &commontesting.MockLinkReadUsecase{}
	handler := Handler(mockUsecase, nil)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, writer.Code)
	}
}

func TestResolveHandler_SingleCharPath_NotFound(t *testing.T) {
	mockUsecase := &commontesting.MockLinkReadUsecase{
		GetFullURLByShortKeyFunc: func(ctx context.Context, shortKey string) (string, error) {
			return "", model.ErrEmptyUser
		},
	}
	handler := Handler(mockUsecase, nil)

	request := httptest.NewRequest(http.MethodGet, "/a", nil)
	request = setURLParam(request, "shortKey", "a")
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, writer.Code)
	}
}

func TestResolveHandler_PathLengthLessThanTwo(t *testing.T) {
	mockUsecase := &commontesting.MockLinkReadUsecase{}
	handler := Handler(mockUsecase, nil)

	testCases := []struct {
		shortKey    string
		description string
	}{
		{"", "Empty shortKey"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if testCase.shortKey != "" {
				req = setURLParam(req, "shortKey", testCase.shortKey)
			}
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status code %d for shortKey '%s', got %d", http.StatusBadRequest, testCase.shortKey, w.Code)
			}
		})
	}
}

func TestResolveHandler_UsecaseReturnsError(t *testing.T) {
	mockUsecase := &commontesting.MockLinkReadUsecase{
		GetFullURLByShortKeyFunc: func(ctx context.Context, shortKey string) (string, error) {
			return "", errors.New("not found")
		},
	}
	handler := Handler(mockUsecase, nil)

	request := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	request = setURLParam(request, "shortKey", "abc123")
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, writer.Code)
	}
}

func TestResolveHandler_UsecaseReturnsEmptyString(t *testing.T) {
	mockUsecase := &commontesting.MockLinkReadUsecase{
		GetFullURLByShortKeyFunc: func(ctx context.Context, shortKey string) (string, error) {
			return "", nil
		},
	}
	handler := Handler(mockUsecase, nil)

	request := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	request = setURLParam(request, "shortKey", "abc123")
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, writer.Code)
	}
}

func TestResolveHandler_Returns410WhenDeleted(t *testing.T) {
	shortKey := "abc123"
	mockUsecase := &commontesting.MockLinkReadUsecase{
		GetFullURLByShortKeyFunc: func(ctx context.Context, key string) (string, error) {
			return "", model.ErrURLDeleted
		},
	}
	handler := Handler(mockUsecase, nil)

	request := httptest.NewRequest(http.MethodGet, "/"+shortKey, nil)
	request = setURLParam(request, "shortKey", shortKey)
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusGone {
		t.Errorf("Expected status code %d, got %d", http.StatusGone, writer.Code)
	}
}

func TestResolveHandler_Success(t *testing.T) {
	expectedURL := "https://example.com/full-url"
	shortKey := "abc123"
	mockUsecase := &commontesting.MockLinkReadUsecase{
		GetFullURLByShortKeyFunc: func(ctx context.Context, key string) (string, error) {
			if key != shortKey {
				t.Errorf("Expected shortKey '%s', got '%s'", shortKey, key)
			}
			return expectedURL, nil
		},
	}
	rec := &recordingAuditor{}
	handler := Handler(mockUsecase, rec)

	request := httptest.NewRequest(http.MethodGet, "/"+shortKey, nil)
	request = setURLParam(request, "shortKey", shortKey)
	request = request.WithContext(authctx.WithUserID(request.Context(), "u3"))
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status code %d, got %d", http.StatusTemporaryRedirect, writer.Code)
	}

	location := writer.Header().Get("Location")
	if location != expectedURL {
		t.Errorf("Expected Location header '%s', got '%s'", expectedURL, location)
	}

	if len(rec.events) != 1 {
		t.Fatalf("Expected 1 audit event, got %d", len(rec.events))
	}
	ev := rec.events[0]
	if ev.Action != "follow" {
		t.Errorf("Expected action 'follow', got '%s'", ev.Action)
	}
	if ev.UserID != "u3" {
		t.Errorf("Expected user_id 'u3', got '%s'", ev.UserID)
	}
	if ev.URL != expectedURL {
		t.Errorf("Expected url '%s', got '%s'", expectedURL, ev.URL)
	}
}

func TestResolveHandler_ExtractsShortKeyCorrectly(t *testing.T) {
	testCases := []struct {
		shortKey string
		expected string
		path     string
	}{
		{"abc", "abc", "/abc"},
		{"abc123", "abc123", "/abc123"},
		{"very/long/pathsdsdsdsdsdsdsdsdsdsdsdsdsdsdsd", "very/long/pathsdsdsdsdsdsdsdsdsdsdsdsdsdsdsd", "/very/long/pathsdsdsdsdsdsdsdsdsdsdsdsdsdsdsd"},
		{"1234567890", "1234567890", "/1234567890"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.path, func(t *testing.T) {
			var receivedKey string
			mockUsecase := &commontesting.MockLinkReadUsecase{
				GetFullURLByShortKeyFunc: func(ctx context.Context, key string) (string, error) {
					receivedKey = key
					return "https://example.com", nil
				},
			}
			handler := Handler(mockUsecase, nil)

			request := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			request = setURLParam(request, "shortKey", testCase.shortKey)
			writer := httptest.NewRecorder()

			handler(writer, request)

			if receivedKey != testCase.expected {
				t.Errorf("Expected shortKey '%s', got '%s'", testCase.expected, receivedKey)
			}

			if writer.Code != http.StatusTemporaryRedirect {
				t.Errorf("Expected status code %d, got %d", http.StatusTemporaryRedirect, writer.Code)
			}
		})
	}
}

func setURLParam(r *http.Request, key, value string) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeContext))
}
