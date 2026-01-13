package shorter

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	commontesting "github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common/testing"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/shorter/model"
)

func TestShortenHandler_WrongMethod(t *testing.T) {
	mockUsecase := &commontesting.MockLinkUsecase{}
	handler := Handler(mockUsecase)

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			request := httptest.NewRequest(method, "/", nil)
			writer := httptest.NewRecorder()

			handler(writer, request)

			if writer.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status code %d for method %s, got %d", http.StatusBadRequest, method, writer.Code)
			}
		})
	}
}

func TestShortenHandler_WrongContentType(t *testing.T) {
	testCases := []struct {
		contentType string
		description string
	}{
		{"application/json", "JSON content type"},
		{"text/html", "HTML content type"},
		{"application/xml", "XML content type"},
		{"", "Empty content type"},
		{"text/plain; charset=utf-8", "text/plain with charset - should be valid"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockUsecase := &commontesting.MockLinkUsecase{
				GetBaseURLFunc: func() string {
					return "http://localhost:8080"
				},
			}
			if strings.Contains(testCase.contentType, "text/plain") && testCase.contentType != "" {
				mockUsecase.CreateShortKeyFunc = func(context context.Context, URL string) (string, error) {
					return "short-key", nil
				}
			}
			handler := Handler(mockUsecase)

			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
			if testCase.contentType != "" {
				request.Header.Set(common.ContentType, testCase.contentType)
			}
			writer := httptest.NewRecorder()

			handler(writer, request)

			if strings.Contains(testCase.contentType, "text/plain") && testCase.contentType != "" {
				if writer.Code == http.StatusMethodNotAllowed {
					t.Errorf("Expected success for content type '%s', got BadRequest", testCase.contentType)
				}
			} else {
				if writer.Code != http.StatusMethodNotAllowed {
					t.Errorf("Expected status code %d for content type '%s', got %d", http.StatusBadRequest, testCase.contentType, writer.Code)
				}
			}
		})
	}
}

func TestShortenHandler_EmptyBody(t *testing.T) {
	mockUsecase := &commontesting.MockLinkUsecase{}
	handler := Handler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	request.Header.Set(common.ContentType, common.TextPlain)
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, writer.Code)
	}
}

func TestShortenHandler_BodyReadError(t *testing.T) {
	mockUsecase := &commontesting.MockLinkUsecase{}
	handler := Handler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/", &errorReader{})
	request.Header.Set(common.ContentType, common.TextPlain)
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, writer.Code)
	}
}

func TestShortenHandler_InvalidURL(t *testing.T) {
	mockUsecase := &commontesting.MockLinkUsecase{}
	handler := Handler(mockUsecase)

	testCases := []struct {
		body        string
		description string
	}{
		{"not-a-url", "Invalid URL format"},
		{"://example.com", "Missing scheme and host"},
		{"http://", "Missing host"},
		{"ftp://", "Missing host"},
		{"", "Empty URL"},
		{"just-text", "Plain text without scheme"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testCase.body))
			request.Header.Set(common.ContentType, common.TextPlain)
			writer := httptest.NewRecorder()

			handler(writer, request)

			if writer.Code != http.StatusBadRequest {
				t.Errorf("Expected status code %d for body '%s', got %d", http.StatusBadRequest, testCase.body, writer.Code)
			}
		})
	}
}

func TestShortenHandler_CreateShortKeyReturnsError(t *testing.T) {
	mockUsecase := &commontesting.MockLinkUsecase{
		CreateShortKeyFunc: func(context context.Context, URL string) (string, error) {
			return "", errors.New("generation short link error")
		},
	}
	handler := Handler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	request.Header.Set(common.ContentType, common.TextPlain)
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, writer.Code)
	}
}

func TestShortenHandler_Success_HTTP(t *testing.T) {
	expectedShortKey := "abc123"
	mockUsecase := &commontesting.MockLinkUsecase{
		GetBaseURLFunc: func() string {
			return "http://localhost:8080"
		},
		CreateShortKeyFunc: func(context context.Context, URL string) (string, error) {
			if URL != "https://example.com" {
				t.Errorf("Expected URL 'https://example.com', got '%s'", URL)
			}
			return expectedShortKey, nil
		},
	}
	handler := Handler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	request.Header.Set(common.ContentType, common.TextPlain)
	request.Host = "localhost:8080"
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, writer.Code)
	}

	expectedURL := "http://localhost:8080/" + expectedShortKey
	body := writer.Body.String()
	if body != expectedURL {
		t.Errorf("Expected body '%s', got '%s'", expectedURL, body)
	}

	contentType := writer.Header().Get(common.ContentType)
	if contentType != common.TextPlain {
		t.Errorf("Expected Content-Type '%s', got '%s'", common.TextPlain, contentType)
	}
}

func TestShortenHandler_Success_WithoutBaseURL_HTTP(t *testing.T) {
	expectedShortKey := "xyz789"
	mockUsecase := &commontesting.MockLinkUsecase{
		GetBaseURLFunc: func() string {
			return ""
		},
		CreateShortKeyFunc: func(context context.Context, URL string) (string, error) {
			return expectedShortKey, nil
		},
	}
	handler := Handler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	request.Header.Set(common.ContentType, common.TextPlain)
	request.Host = "localhost:8080"
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, writer.Code)
	}

	expectedURL := "http://localhost:8080/" + expectedShortKey
	body := writer.Body.String()
	if body != expectedURL {
		t.Errorf("Expected body '%s', got '%s'", expectedURL, body)
	}
}

func TestShortenHandler_Success_WithoutBaseURL_HTTPS_TLS(t *testing.T) {
	expectedShortKey := "def456"
	mockUsecase := &commontesting.MockLinkUsecase{
		GetBaseURLFunc: func() string {
			return ""
		},
		CreateShortKeyFunc: func(context context.Context, URL string) (string, error) {
			return expectedShortKey, nil
		},
	}
	handler := Handler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	request.Header.Set(common.ContentType, common.TextPlain)
	request.Host = "example.com"
	request.TLS = &tls.ConnectionState{}
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, writer.Code)
	}

	expectedURL := "https://example.com/" + expectedShortKey
	body := writer.Body.String()
	if body != expectedURL {
		t.Errorf("Expected body '%s', got '%s'", expectedURL, body)
	}
}

func TestShortenHandler_Success_WithoutBaseURL_HTTPS_XForwardedProto(t *testing.T) {
	expectedShortKey := "ghi789"
	mockUsecase := &commontesting.MockLinkUsecase{
		GetBaseURLFunc: func() string {
			return ""
		},
		CreateShortKeyFunc: func(context context.Context, URL string) (string, error) {
			return expectedShortKey, nil
		},
	}
	handler := Handler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	request.Header.Set(common.ContentType, common.TextPlain)
	request.Header.Set("X-Forwarded-Proto", "https")
	request.Host = "example.com"
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, writer.Code)
	}

	expectedURL := "https://example.com/" + expectedShortKey
	body := writer.Body.String()
	if body != expectedURL {
		t.Errorf("Expected body '%s', got '%s'", expectedURL, body)
	}
}

func TestShortenHandler_Success_HTTPS_TLS(t *testing.T) {
	expectedShortKey := "xyz789"
	mockUsecase := &commontesting.MockLinkUsecase{
		GetBaseURLFunc: func() string {
			return "https://example.com"
		},
		CreateShortKeyFunc: func(context context.Context, URL string) (string, error) {
			return expectedShortKey, nil
		},
	}
	handler := Handler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	request.Header.Set(common.ContentType, common.TextPlain)
	request.Host = "example.com"
	request.TLS = &tls.ConnectionState{}
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, writer.Code)
	}

	expectedURL := "https://example.com/" + expectedShortKey
	body := writer.Body.String()
	if body != expectedURL {
		t.Errorf("Expected body '%s', got '%s'", expectedURL, body)
	}
}

func TestShortenHandler_Success_HTTPS_XForwardedProto(t *testing.T) {
	expectedShortKey := "def456"
	mockUsecase := &commontesting.MockLinkUsecase{
		GetBaseURLFunc: func() string {
			return "https://example.com"
		},
		CreateShortKeyFunc: func(context context.Context, URL string) (string, error) {
			return expectedShortKey, nil
		},
	}
	handler := Handler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	request.Header.Set(common.ContentType, common.TextPlain)
	request.Header.Set("X-Forwarded-Proto", "https")
	request.Host = "example.com"
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, writer.Code)
	}

	expectedURL := "https://example.com/" + expectedShortKey
	body := writer.Body.String()
	if body != expectedURL {
		t.Errorf("Expected body '%s', got '%s'", expectedURL, body)
	}
}

func TestShortenHandler_ResponseWriteError(t *testing.T) {
	expectedShortKey := "test123"
	mockUsecase := &commontesting.MockLinkUsecase{
		GetBaseURLFunc: func() string {
			return "http://localhost:8080"
		},
		CreateShortKeyFunc: func(context context.Context, URL string) (string, error) {
			return expectedShortKey, nil
		},
	}
	handler := Handler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	request.Header.Set(common.ContentType, common.TextPlain)
	request.Host = "localhost:8080"

	writer := &errorResponseWriter{
		ResponseRecorder: httptest.NewRecorder(),
		shouldError:      true,
	}

	handler(writer, request)

	if writer.statusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, writer.statusCode)
	}
}

func TestIsTextPlain(t *testing.T) {
	testCases := []struct {
		contentType string
		expected    bool
		description string
	}{
		{"text/plain", true, "Simple text/plain"},
		{"text/plain; charset=utf-8", true, "text/plain with charset"},
		{"text/plain; charset=UTF-8", true, "text/plain with UTF-8 charset"},
		{"application/json", false, "JSON content type"},
		{"text/html", false, "HTML content type"},
		{"", false, "Empty content type"},
		{"invalid", false, "Invalid content type"},
		{"text/plain; boundary=something", true, "text/plain with boundary"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			result := IsTextPlain(testCase.contentType)
			if result != testCase.expected {
				t.Errorf("IsTextPlain('%s') = %v, expected %v", testCase.contentType, result, testCase.expected)
			}
		})
	}
}

func TestParseURL(t *testing.T) {
	testCases := []struct {
		input       string
		shouldError bool
		description string
	}{
		{"https://example.com", false, "Valid HTTPS URL"},
		{"http://example.com", false, "Valid HTTP URL"},
		{"http://example.com/path", false, "Valid URL with path"},
		{"http://example.com/", false, "Valid URL with trailing slash"},
		{"https://example.com:8080", false, "Valid URL with port"},
		{"not-a-url", true, "Invalid URL"},
		{"://example.com", true, "Missing scheme"},
		{"http://", true, "Missing host"},
		{"", true, "Empty URL"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			parsedURL, err := parseURL(testCase.input)

			if testCase.shouldError {
				if err == nil {
					t.Errorf("parseURL('%s') should return error, got nil", testCase.input)
				}
				if parsedURL != nil && (parsedURL.Scheme != "" || parsedURL.Host != "") {
					t.Errorf("parseURL('%s') should return nil or empty URL on error", testCase.input)
				}
			} else {
				if err != nil {
					t.Errorf("parseURL('%s') should not return error, got: %v", testCase.input, err)
				}
				if parsedURL == nil {
					t.Errorf("parseURL('%s') should return non-nil URL", testCase.input)
				} else if parsedURL.Scheme == "" || parsedURL.Host == "" {
					t.Errorf("parseURL('%s') should return URL with scheme and host, got: %+v", testCase.input, parsedURL)
				}
			}
		})
	}
}

func TestReadBody(t *testing.T) {
	testCases := []struct {
		body        string
		shouldError bool
		description string
	}{
		{"test body", false, "Valid body"},
		{"", false, "Empty body"},
		{"long body with multiple lines\nand content", false, "Multiline body"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testCase.body))
			body, err := readBody(req)

			if testCase.shouldError {
				if err == nil {
					t.Errorf("readBody should return error for '%s', got nil", testCase.description)
				}
			} else {
				if err != nil {
					t.Errorf("readBody should not return error for '%s', got: %v", testCase.description, err)
				}
				if string(body) != testCase.body {
					t.Errorf("readBody returned '%s', expected '%s'", string(body), testCase.body)
				}
			}
		})
	}
}

func TestReadBody_ReadError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", &errorReader{})
	body, err := readBody(req)

	if err == nil {
		t.Error("readBody should return error when reader fails")
	}
	if len(body) > 0 {
		t.Errorf("readBody should return empty body on error, got: %s", string(body))
	}
}

func TestSendResponse(t *testing.T) {
	writer := httptest.NewRecorder()
	shortURL := "http://localhost:8080/abc123"

	sendResponse(writer, shortURL)

	if writer.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, writer.Code)
	}

	contentType := writer.Header().Get(common.ContentType)
	if contentType != common.TextPlain {
		t.Errorf("Expected Content-Type '%s', got '%s'", common.TextPlain, contentType)
	}

	body := writer.Body.String()
	if body != shortURL {
		t.Errorf("Expected body '%s', got '%s'", shortURL, body)
	}
}

func TestSendResponse_WriteError(t *testing.T) {
	writer := &errorResponseWriter{
		ResponseRecorder: httptest.NewRecorder(),
		shouldError:      true,
	}
	shortURL := "http://localhost:8080/abc123"

	sendResponse(writer, shortURL)

	if writer.statusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, writer.statusCode)
	}
}

func TestShortenHandler_URLCreatesWithCorrectFormat(t *testing.T) {
	mockUsecase := &commontesting.MockLinkUsecase{
		GetBaseURLFunc: func() string {
			return "http://myserver.com:9090"
		},
		CreateShortKeyFunc: func(context context.Context, URL string) (string, error) {
			return "short123", nil
		},
	}
	handler := Handler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://example.com/path"))
	request.Header.Set(common.ContentType, common.TextPlain)
	request.Host = "myserver.com:9090"
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, writer.Code)
	}

	expectedURL := "http://myserver.com:9090/short123"
	body := writer.Body.String()
	if body != expectedURL {
		t.Errorf("Expected body '%s', got '%s'", expectedURL, body)
	}
}

func TestShortenHandler_NormalizesURL(t *testing.T) {
	mockUsecase := &commontesting.MockLinkUsecase{
		GetBaseURLFunc: func() string {
			return "http://localhost:8080"
		},
		CreateShortKeyFunc: func(context context.Context, URL string) (string, error) {
			if URL == "https://example.com" {
				return "normalized123", nil
			}
			t.Errorf("Expected normalized URL 'https://example.com', got '%s'", URL)
			return "error", nil
		},
	}
	handler := Handler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com/"))
	request.Header.Set(common.ContentType, common.TextPlain)
	request.Host = "localhost:8080"
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, writer.Code)
	}
}

func TestShortenHandler_HandlesBodyWithWhitespace(t *testing.T) {
	mockUsecase := &commontesting.MockLinkUsecase{
		GetBaseURLFunc: func() string {
			return "http://localhost:8080"
		},
		CreateShortKeyFunc: func(context context.Context, URL string) (string, error) {
			if URL == "https://example.com" {
				return "trimmed123", nil
			}
			return "error", nil
		},
	}
	handler := Handler(mockUsecase)

	requestq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("  https://example.com  \n"))
	requestq.Header.Set(common.ContentType, common.TextPlain)
	requestq.Host = "localhost:8080"
	writer := httptest.NewRecorder()

	handler(writer, requestq)

	if writer.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, writer.Code)
	}
}

func TestReadBody_ClosesBody(t *testing.T) {
	bodyContent := "test body"
	body := &testCloser{
		Reader: strings.NewReader(bodyContent),
		closed: false,
	}

	req := httptest.NewRequest(http.MethodPost, "/", body)
	_, _ = readBody(req)

	if !body.closed {
		t.Error("readBody should close the request body")
	}
}

type testCloser struct {
	io.Reader
	closed bool
}

func (t *testCloser) Close() error {
	t.closed = true
	return nil
}

type errorReader struct{}

func (e *errorReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("read error")
}
func (e *errorReader) Close() error {
	return nil
}

type errorResponseWriter struct {
	*httptest.ResponseRecorder
	shouldError bool
	statusCode  int
}

func (w *errorResponseWriter) Write(p []byte) (n int, err error) {
	if w.shouldError {
		return 0, errors.New("write error")
	}
	return w.ResponseRecorder.Write(p)
}

func (w *errorResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseRecorder.WriteHeader(statusCode)
}

// Tests for APIHandler

func TestAPIHandler_WrongContentType(t *testing.T) {
	testCases := []struct {
		contentType string
		description string
	}{
		{"text/plain", "text/plain content type"},
		{"text/html", "HTML content type"},
		{"application/xml", "XML content type"},
		{"", "Empty content type"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockUsecase := &commontesting.MockLinkUsecase{}
			handler := APIHandler(mockUsecase)

			request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://example.com"}`))
			if testCase.contentType != "" {
				request.Header.Set(common.ContentType, testCase.contentType)
			}
			writer := httptest.NewRecorder()

			handler(writer, request)

			if writer.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status code %d for content type '%s', got %d", http.StatusMethodNotAllowed, testCase.contentType, writer.Code)
			}
		})
	}
}

func TestAPIHandler_InvalidJSON(t *testing.T) {
	testCases := []struct {
		body        string
		description string
	}{
		{"invalid json", "Invalid JSON format"},
		{"{url:}", "Invalid JSON syntax"},
		{"", "Empty body"},
		{"{", "Incomplete JSON"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockUsecase := &commontesting.MockLinkUsecase{}
			handler := APIHandler(mockUsecase)

			request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(testCase.body))
			request.Header.Set(common.ContentType, common.ApplicationJSON)
			writer := httptest.NewRecorder()

			handler(writer, request)

			if writer.Code != http.StatusBadRequest {
				t.Errorf("Expected status code %d for body '%s', got %d", http.StatusBadRequest, testCase.body, writer.Code)
			}
		})
	}
}

func TestAPIHandler_InvalidURL(t *testing.T) {
	testCases := []struct {
		body        string
		description string
	}{
		{`{"url":"not-a-url"}`, "Invalid URL format"},
		{`{"url":"://example.com"}`, "Missing scheme and host"},
		{`{"url":"http://"}`, "Missing host"},
		{`{"url":""}`, "Empty URL"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockUsecase := &commontesting.MockLinkUsecase{}
			handler := APIHandler(mockUsecase)

			request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(testCase.body))
			request.Header.Set(common.ContentType, common.ApplicationJSON)
			writer := httptest.NewRecorder()

			handler(writer, request)

			if writer.Code != http.StatusBadRequest {
				t.Errorf("Expected status code %d for body '%s', got %d", http.StatusBadRequest, testCase.body, writer.Code)
			}
		})
	}
}

func TestAPIHandler_CreateShortKeyReturnsError(t *testing.T) {
	mockUsecase := &commontesting.MockLinkUsecase{
		CreateShortKeyFunc: func(context context.Context, URL string) (string, error) {
			return "", errors.New("CreateShortKey error")
		},
	}
	handler := APIHandler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://example.com"}`))
	request.Header.Set(common.ContentType, common.ApplicationJSON)
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, writer.Code)
	}
}

func TestAPIHandler_Success_WithBaseURL(t *testing.T) {
	expectedShortKey := "abc123"
	mockUsecase := &commontesting.MockLinkUsecase{
		GetBaseURLFunc: func() string {
			return "http://localhost:8080"
		},
		CreateShortKeyFunc: func(context context.Context, URL string) (string, error) {
			if URL != "https://example.com" {
				t.Errorf("Expected URL 'https://example.com', got '%s'", URL)
			}
			return expectedShortKey, nil
		},
	}
	handler := APIHandler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://example.com"}`))
	request.Header.Set(common.ContentType, common.ApplicationJSON)
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, writer.Code)
	}

	expectedResult := "http://localhost:8080/abc123"
	var response model.ShortenResponse
	if err := json.Unmarshal(writer.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response.Result != expectedResult {
		t.Errorf("Expected result '%s', got '%s'", expectedResult, response.Result)
	}

	contentType := writer.Header().Get(common.ContentType)
	if contentType != common.ApplicationJSON {
		t.Errorf("Expected Content-Type '%s', got '%s'", common.ApplicationJSON, contentType)
	}
}

func TestAPIHandler_Success_WithoutBaseURL_HTTP(t *testing.T) {
	expectedShortKey := "xyz789"
	mockUsecase := &commontesting.MockLinkUsecase{
		GetBaseURLFunc: func() string {
			return ""
		},
		CreateShortKeyFunc: func(context context.Context, URL string) (string, error) {
			return expectedShortKey, nil
		},
	}
	handler := APIHandler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://example.com"}`))
	request.Header.Set(common.ContentType, common.ApplicationJSON)
	request.Host = "localhost:8080"
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, writer.Code)
	}

	expectedResult := "http://localhost:8080/xyz789"
	var response model.ShortenResponse
	if err := json.Unmarshal(writer.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response.Result != expectedResult {
		t.Errorf("Expected result '%s', got '%s'", expectedResult, response.Result)
	}
}

func TestAPIHandler_Success_WithoutBaseURL_HTTPS_TLS(t *testing.T) {
	expectedShortKey := "def456"
	mockUsecase := &commontesting.MockLinkUsecase{
		GetBaseURLFunc: func() string {
			return ""
		},
		CreateShortKeyFunc: func(context context.Context, URL string) (string, error) {
			return expectedShortKey, nil
		},
	}
	handler := APIHandler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://example.com"}`))
	request.Header.Set(common.ContentType, common.ApplicationJSON)
	request.Host = "example.com"
	request.TLS = &tls.ConnectionState{}
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, writer.Code)
	}

	expectedResult := "https://example.com/def456"
	var response model.ShortenResponse
	if err := json.Unmarshal(writer.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response.Result != expectedResult {
		t.Errorf("Expected result '%s', got '%s'", expectedResult, response.Result)
	}
}

func TestAPIHandler_Success_WithoutBaseURL_HTTPS_XForwardedProto(t *testing.T) {
	expectedShortKey := "ghi789"
	mockUsecase := &commontesting.MockLinkUsecase{
		GetBaseURLFunc: func() string {
			return ""
		},
		CreateShortKeyFunc: func(context context.Context, URL string) (string, error) {
			return expectedShortKey, nil
		},
	}
	handler := APIHandler(mockUsecase)

	request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://example.com"}`))
	request.Header.Set(common.ContentType, common.ApplicationJSON)
	request.Header.Set("X-Forwarded-Proto", "https")
	request.Host = "example.com"
	writer := httptest.NewRecorder()

	handler(writer, request)

	if writer.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, writer.Code)
	}

	expectedResult := "https://example.com/ghi789"
	var response model.ShortenResponse
	if err := json.Unmarshal(writer.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response.Result != expectedResult {
		t.Errorf("Expected result '%s', got '%s'", expectedResult, response.Result)
	}
}

// Tests for GenerateShortenURL

func TestGenerateShortenURL_WithBaseURL(t *testing.T) {
	baseURL := "http://localhost:8080"
	shortKey := "abc123"
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	result := GenerateShortenURL(baseURL, shortKey, request)

	expected := "http://localhost:8080/abc123"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestGenerateShortenURL_WithoutBaseURL_HTTP(t *testing.T) {
	baseURL := ""
	shortKey := "xyz789"
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Host = "example.com"

	result := GenerateShortenURL(baseURL, shortKey, request)

	expected := "http://example.com/xyz789"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestGenerateShortenURL_WithoutBaseURL_HTTPS_TLS(t *testing.T) {
	baseURL := ""
	shortKey := "def456"
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Host = "example.com"
	request.TLS = &tls.ConnectionState{}

	result := GenerateShortenURL(baseURL, shortKey, request)

	expected := "https://example.com/def456"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestGenerateShortenURL_WithoutBaseURL_HTTPS_XForwardedProto(t *testing.T) {
	baseURL := ""
	shortKey := "ghi789"
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Host = "example.com"
	request.Header.Set("X-Forwarded-Proto", "https")

	result := GenerateShortenURL(baseURL, shortKey, request)

	expected := "https://example.com/ghi789"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestGenerateShortenURL_WithoutBaseURL_HTTP_WhenXForwardedProtoIsHTTP(t *testing.T) {
	baseURL := ""
	shortKey := "jkl012"
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Host = "example.com"
	request.Header.Set("X-Forwarded-Proto", "http")

	result := GenerateShortenURL(baseURL, shortKey, request)

	expected := "http://example.com/jkl012"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// Tests for GenerateShortKey

func TestGenerateShortKey_Success(t *testing.T) {
	body := "https://example.com"
	expectedShortKey := "abc123"
	mockUsecase := &commontesting.MockLinkUsecase{
		CreateShortKeyFunc: func(ctx context.Context, URL string) (string, error) {
			if URL != "https://example.com" {
				t.Errorf("Expected URL 'https://example.com', got '%s'", URL)
			}
			return expectedShortKey, nil
		},
	}
	writer := httptest.NewRecorder()
	ctx := context.Background()

	shortKey, err := GenerateShortKey(body, ctx, writer, mockUsecase)

	if err != nil {
		t.Fatalf("Expected no error, got %v. Writer code: %d, body: %s", err, writer.Code, writer.Body.String())
	}
	if shortKey != expectedShortKey {
		t.Errorf("Expected shortKey '%s', got '%s'", expectedShortKey, shortKey)
	}
	if writer.Code == http.StatusBadRequest {
		t.Errorf("Expected no error status code, got BadRequest (%d). Body: %s", writer.Code, writer.Body.String())
	}
}

func TestGenerateShortKey_InvalidURL(t *testing.T) {
	body := "not-a-url"
	mockUsecase := &commontesting.MockLinkUsecase{}
	writer := httptest.NewRecorder()
	ctx := context.Background()

	shortKey, err := GenerateShortKey(body, ctx, writer, mockUsecase)

	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
	if shortKey != "" {
		t.Errorf("Expected empty shortKey, got '%s'", shortKey)
	}
	if writer.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, writer.Code)
	}
}

func TestGenerateShortKey_EmptyURL(t *testing.T) {
	body := ""
	mockUsecase := &commontesting.MockLinkUsecase{}
	writer := httptest.NewRecorder()
	ctx := context.Background()

	shortKey, err := GenerateShortKey(body, ctx, writer, mockUsecase)

	if err == nil {
		t.Error("Expected error for empty URL, got nil")
	}
	if shortKey != "" {
		t.Errorf("Expected empty shortKey, got '%s'", shortKey)
	}
	if writer.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, writer.Code)
	}
}

func TestGenerateShortKey_CreateShortKeyError(t *testing.T) {
	body := "https://example.com"
	mockUsecase := &commontesting.MockLinkUsecase{
		CreateShortKeyFunc: func(ctx context.Context, URL string) (string, error) {
			return "", errors.New("CreateShortKey error")
		},
	}
	writer := httptest.NewRecorder()
	ctx := context.Background()

	shortKey, err := GenerateShortKey(body, ctx, writer, mockUsecase)

	if err == nil {
		t.Error("Expected error from CreateShortKey, got nil")
	}
	if shortKey != "" {
		t.Errorf("Expected empty shortKey, got '%s'", shortKey)
	}
	if writer.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, writer.Code)
	}
}

func TestGenerateShortKey_NormalizesURL(t *testing.T) {
	body := "https://example.com/"
	expectedShortKey := "normalized123"
	mockUsecase := &commontesting.MockLinkUsecase{
		CreateShortKeyFunc: func(ctx context.Context, URL string) (string, error) {
			if URL != "https://example.com" {
				t.Errorf("Expected normalized URL 'https://example.com', got '%s'", URL)
			}
			return expectedShortKey, nil
		},
	}
	writer := httptest.NewRecorder()
	ctx := context.Background()

	shortKey, err := GenerateShortKey(body, ctx, writer, mockUsecase)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if shortKey != expectedShortKey {
		t.Errorf("Expected shortKey '%s', got '%s'", expectedShortKey, shortKey)
	}
}

func TestGenerateShortKey_TrimsWhitespace(t *testing.T) {
	body := "  https://example.com  \n"
	expectedShortKey := "trimmed123"
	mockUsecase := &commontesting.MockLinkUsecase{
		CreateShortKeyFunc: func(ctx context.Context, URL string) (string, error) {
			if URL != "https://example.com" {
				t.Errorf("Expected normalized URL 'https://example.com', got '%s'", URL)
			}
			return expectedShortKey, nil
		},
	}
	writer := httptest.NewRecorder()
	ctx := context.Background()

	shortKey, err := GenerateShortKey(body, ctx, writer, mockUsecase)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if shortKey != expectedShortKey {
		t.Errorf("Expected shortKey '%s', got '%s'", expectedShortKey, shortKey)
	}
}
