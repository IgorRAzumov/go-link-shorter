package shorter

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	commontesting "github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common/testing"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/shorter/model"
	domainmodel "github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

func TestBatchAPIHandler_WrongMethod(t *testing.T) {
	mockUsecase := &commontesting.MockLinkCreateUsecase{}
	handler := BatchAPIHandler(mockUsecase)

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			request := httptest.NewRequest(method, "/api/shorten/batch", nil)
			writer := httptest.NewRecorder()

			handler(writer, request)

			if writer.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status code %d for method %s, got %d", http.StatusMethodNotAllowed, method, writer.Code)
			}
		})
	}
}

func TestBatchAPIHandler_WrongContentType(t *testing.T) {
	testCases := []struct {
		contentType string
		description string
	}{
		{"text/plain", "text/plain content type"},
		{"text/html", "HTML content type"},
		{"application/xml", "XML content type"},
		{"", "Empty content type"},
		{"application/json; charset=utf-8", "application/json with charset - should be valid"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockUsecase := &commontesting.MockLinkCreateUsecase{
				GetBaseURLFunc: func() string {
					return "http://localhost:8080"
				},
			}
			if strings.Contains(testCase.contentType, "application/json") && testCase.contentType != "" {
				mockUsecase.ProcessBatchShortenRequestsFunc = func(ctx context.Context, requests []domainmodel.BatchShortenRequest) ([]domainmodel.BatchShortenResult, error) {
					responses := make([]domainmodel.BatchShortenResult, 0, len(requests))
					for _, req := range requests {
						responses = append(responses, domainmodel.BatchShortenResult{
							CorrelationID: req.CorrelationID,
							ShortKey:      "test-key",
						})
					}
					return responses, nil
				}
			}
			handler := BatchAPIHandler(mockUsecase)

			request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(`[{"correlation_id":"1","original_url":"https://example.com"}]`))
			if testCase.contentType != "" {
				request.Header.Set(common.ContentType, testCase.contentType)
			}
			writer := httptest.NewRecorder()

			handler(writer, request)

			if strings.Contains(testCase.contentType, "application/json") && testCase.contentType != "" {
				if writer.Code != http.StatusCreated {
					t.Errorf("Expected status code %d for content type '%s', got %d", http.StatusCreated, testCase.contentType, writer.Code)
				}
			} else {
				if writer.Code != http.StatusMethodNotAllowed {
					t.Errorf("Expected status code %d for content type '%s', got %d", http.StatusMethodNotAllowed, testCase.contentType, writer.Code)
				}
			}
		})
	}
}

func TestBatchAPIHandler_InvalidJSON(t *testing.T) {
	testCases := []struct {
		body        string
		description string
	}{
		{"invalid json", "Invalid JSON format"},
		{"[{correlation_id:}]", "Invalid JSON syntax"},
		{"", "Empty body"},
		{"[", "Incomplete JSON"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockUsecase := &commontesting.MockLinkCreateUsecase{}
			handler := BatchAPIHandler(mockUsecase)

			request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(testCase.body))
			request.Header.Set(common.ContentType, common.ApplicationJSON)
			writer := httptest.NewRecorder()

			handler(writer, request)

			if writer.Code != http.StatusBadRequest {
				t.Errorf("Expected status code %d for body '%s', got %d", http.StatusBadRequest, testCase.body, writer.Code)
			}
		})
	}
}

func TestBatchAPIHandler_EmptyBatch(t *testing.T) {
	mockUsecase := &commontesting.MockLinkCreateUsecase{}
	handler := BatchAPIHandler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(`[]`))
	request.Header.Set(common.ContentType, common.ApplicationJSON)
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d for empty batch, got %d", http.StatusBadRequest, writer.Code)
	}
}

func TestBatchAPIHandler_CreateShortKeysBatchReturnsError(t *testing.T) {
	mockUsecase := &commontesting.MockLinkCreateUsecase{
		ProcessBatchShortenRequestsFunc: func(ctx context.Context, requests []domainmodel.BatchShortenRequest) ([]domainmodel.BatchShortenResult, error) {
			return nil, errors.New("ProcessBatchShortenRequests error")
		},
	}
	handler := BatchAPIHandler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(`[{"correlation_id":"1","original_url":"https://example.com"}]`))
	request.Header.Set(common.ContentType, common.ApplicationJSON)
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, writer.Code)
	}
}

func TestBatchAPIHandler_Success_WithBaseURL(t *testing.T) {
	expectedShortKey := "abc123"
	mockUsecase := &commontesting.MockLinkCreateUsecase{
		GetBaseURLFunc: func() string { return "http://localhost:8080" },
		ProcessBatchShortenRequestsFunc: func(ctx context.Context, requests []domainmodel.BatchShortenRequest) ([]domainmodel.BatchShortenResult, error) {
			responses := make([]domainmodel.BatchShortenResult, 0, len(requests))
			for _, req := range requests {
				responses = append(responses, domainmodel.BatchShortenResult{
					CorrelationID: req.CorrelationID,
					ShortKey:      expectedShortKey,
				})
			}
			return responses, nil
		},
	}
	handler := BatchAPIHandler(mockUsecase)

	requestBody := `[{"correlation_id":"1","original_url":"https://example.com"}]`
	request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(requestBody))
	request.Header.Set(common.ContentType, common.ApplicationJSON)
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, writer.Code)
	}

	var responses []model.BatchShortenResponse
	if err := json.Unmarshal(writer.Body.Bytes(), &responses); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(responses) != 1 {
		t.Fatalf("Expected 1 response, got %d", len(responses))
	}

	expectedResult := "http://localhost:8080/abc123"
	if responses[0].ShortURL != expectedResult {
		t.Errorf("Expected result '%s', got '%s'", expectedResult, responses[0].ShortURL)
	}

	if responses[0].CorrelationID != "1" {
		t.Errorf("Expected correlation_id '1', got '%s'", responses[0].CorrelationID)
	}

	contentType := writer.Header().Get(common.ContentType)
	if contentType != common.ApplicationJSON {
		t.Errorf("Expected Content-Type '%s', got '%s'", common.ApplicationJSON, contentType)
	}
}

func TestBatchAPIHandler_Success_MultipleURLs(t *testing.T) {
	mockUsecase := &commontesting.MockLinkCreateUsecase{
		GetBaseURLFunc: func() string { return "http://localhost:8080" },
		ProcessBatchShortenRequestsFunc: func(ctx context.Context, requests []domainmodel.BatchShortenRequest) ([]domainmodel.BatchShortenResult, error) {
			keys := []string{"key1", "key2", "key3"}
			responses := make([]domainmodel.BatchShortenResult, 0, len(requests))
			for i, req := range requests {
				responses = append(responses, domainmodel.BatchShortenResult{
					CorrelationID: req.CorrelationID,
					ShortKey:      keys[i],
				})
			}
			return responses, nil
		},
	}
	handler := BatchAPIHandler(mockUsecase)

	requestBody := `[
		{"correlation_id":"1","original_url":"https://example1.com"},
		{"correlation_id":"2","original_url":"https://example2.com"},
		{"correlation_id":"3","original_url":"https://example3.com"}
	]`
	request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(requestBody))
	request.Header.Set(common.ContentType, common.ApplicationJSON)
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, writer.Code)
	}

	var responses []model.BatchShortenResponse
	if err := json.Unmarshal(writer.Body.Bytes(), &responses); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(responses) != 3 {
		t.Fatalf("Expected 3 responses, got %d", len(responses))
	}

	expectedResults := []string{
		"http://localhost:8080/key1",
		"http://localhost:8080/key2",
		"http://localhost:8080/key3",
	}
	expectedCorrelationIDs := []string{"1", "2", "3"}

	for i, response := range responses {
		if response.ShortURL != expectedResults[i] {
			t.Errorf("Response %d: Expected result '%s', got '%s'", i, expectedResults[i], response.ShortURL)
		}
		if response.CorrelationID != expectedCorrelationIDs[i] {
			t.Errorf("Response %d: Expected correlation_id '%s', got '%s'", i, expectedCorrelationIDs[i], response.CorrelationID)
		}
	}
}

func TestBatchAPIHandler_Success_WithoutBaseURL_HTTP(t *testing.T) {
	expectedShortKey := "xyz789"
	mockUsecase := &commontesting.MockLinkCreateUsecase{
		GetBaseURLFunc: func() string { return "" },
		ProcessBatchShortenRequestsFunc: func(ctx context.Context, requests []domainmodel.BatchShortenRequest) ([]domainmodel.BatchShortenResult, error) {
			responses := make([]domainmodel.BatchShortenResult, 0, len(requests))
			for _, req := range requests {
				responses = append(responses, domainmodel.BatchShortenResult{
					CorrelationID: req.CorrelationID,
					ShortKey:      expectedShortKey,
				})
			}
			return responses, nil
		},
	}
	handler := BatchAPIHandler(mockUsecase)

	requestBody := `[{"correlation_id":"1","original_url":"https://example.com"}]`
	request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(requestBody))
	request.Header.Set(common.ContentType, common.ApplicationJSON)
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, writer.Code)
	}

	var responses []model.BatchShortenResponse
	if err := json.Unmarshal(writer.Body.Bytes(), &responses); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(responses) == 0 {
		t.Fatalf("Expected at least 1 response, got 0")
	}

	if !strings.Contains(responses[0].ShortURL, expectedShortKey) {
		t.Errorf("Expected short URL to contain '%s', got '%s'", expectedShortKey, responses[0].ShortURL)
	}
}
