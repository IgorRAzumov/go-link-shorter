package reolver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
)

func TestResolveHandler_EmptyPath(t *testing.T) {
	mockUsecase := &common.MockLinkUsecase{}
	handler := Handler(mockUsecase)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, writer.Code)
	}
}

func TestResolveHandler_SingleCharPath_NotFound(t *testing.T) {
	mockUsecase := &common.MockLinkUsecase{
		GetFullURLByShortKeyFunc: func(shortKey string) (string, error) {
			return "", errors.New("not found")
		},
	}
	handler := Handler(mockUsecase)

	request := httptest.NewRequest(http.MethodGet, "/a", nil)
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, writer.Code)
	}
}

func TestResolveHandler_PathLengthLessThanTwo(t *testing.T) {
	mockUsecase := &common.MockLinkUsecase{}
	handler := Handler(mockUsecase)

	testCases := []struct {
		path        string
		description string
		forceEmpty  bool
	}{
		{"/", "Root path (length 1)", false},
		{"/", "Empty path (length 0)", true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			if testCase.forceEmpty {
				req.URL.Path = ""
			}
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status code %d for path '%s', got %d", http.StatusBadRequest, req.URL.Path, w.Code)
			}
		})
	}
}

func TestResolveHandler_UsecaseReturnsError(t *testing.T) {
	mockUsecase := &common.MockLinkUsecase{
		GetFullURLByShortKeyFunc: func(shortKey string) (string, error) {
			return "", errors.New("not found")
		},
	}
	handler := Handler(mockUsecase)

	request := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, writer.Code)
	}
}

func TestResolveHandler_UsecaseReturnsEmptyString(t *testing.T) {
	mockUsecase := &common.MockLinkUsecase{
		GetFullURLByShortKeyFunc: func(shortKey string) (string, error) {
			return "", nil
		},
	}
	handler := Handler(mockUsecase)

	request := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, writer.Code)
	}
}

func TestResolveHandler_Success(t *testing.T) {
	expectedURL := "https://example.com/full-url"
	shortKey := "abc123"
	mockUsecase := &common.MockLinkUsecase{
		GetFullURLByShortKeyFunc: func(key string) (string, error) {
			if key != shortKey {
				t.Errorf("Expected shortKey '%s', got '%s'", shortKey, key)
			}
			return expectedURL, nil
		},
	}
	handler := Handler(mockUsecase)

	request := httptest.NewRequest(http.MethodGet, "/"+shortKey, nil)
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status code %d, got %d", http.StatusTemporaryRedirect, writer.Code)
	}

	location := writer.Header().Get("Location")
	if location != expectedURL {
		t.Errorf("Expected Location header '%s', got '%s'", expectedURL, location)
	}
}

func TestResolveHandler_ExtractsShortKeyCorrectly(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
	}{
		{"/abc", "abc"},
		{"/abc123", "abc123"},
		{"/very/long/pathsdsdsdsdsdsdsdsdsdsdsdsdsdsdsd", "very/long/pathsdsdsdsdsdsdsdsdsdsdsdsdsdsdsd"},
		{"/1234567890", "1234567890"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.path, func(t *testing.T) {
			var receivedKey string
			mockUsecase := &common.MockLinkUsecase{
				GetFullURLByShortKeyFunc: func(key string) (string, error) {
					receivedKey = key
					return "https://example.com", nil
				},
			}
			handler := Handler(mockUsecase)

			request := httptest.NewRequest(http.MethodGet, testCase.path, nil)
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
