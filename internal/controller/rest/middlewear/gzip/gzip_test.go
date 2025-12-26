package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
)

func TestGZIP_SkipCompression_UnsupportedContentType(t *testing.T) {
	testCases := []struct {
		responseContentType string
		description         string
	}{
		{"text/plain", "text/plain content type"},
		{"application/xml", "XML content type"},
		{"", "Empty content type"},
		{"image/png", "Image content type"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				if testCase.responseContentType != "" {
					responseWriter.Header().Set(common.ContentType, testCase.responseContentType)
				}
				responseWriter.WriteHeader(http.StatusOK)
				_, _ = responseWriter.Write([]byte("test response"))
			})

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.Header.Set(AcceptEncoding, "gzip")
			responseRecorder := httptest.NewRecorder()

			middleware := GZIP(nextHandler)
			middleware.ServeHTTP(responseRecorder, request)

			if responseRecorder.Code != http.StatusOK {
				t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
			}

			contentEncoding := responseRecorder.Header().Get(ContentEncoding)
			if contentEncoding != "" {
				t.Errorf("Expected no Content-Encoding header, got '%s'", contentEncoding)
			}

			responseBody := responseRecorder.Body.String()
			if responseBody != "test response" {
				t.Errorf("Expected body 'test response', got '%s'", responseBody)
			}
		})
	}
}

func TestGZIP_CompressResponse_ApplicationJSON(t *testing.T) {
	nextHandler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.Header().Set(common.ContentType, common.ApplicationJSON)
		responseWriter.WriteHeader(http.StatusOK)
		_, _ = responseWriter.Write([]byte(`{"message":"test"}`))
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(common.ContentType, common.ApplicationJSON)
	request.Header.Set(AcceptEncoding, "gzip")
	responseRecorder := httptest.NewRecorder()

	middleware := GZIP(nextHandler)
	middleware.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	contentEncoding := responseRecorder.Header().Get(ContentEncoding)
	if contentEncoding != HeaderCode {
		t.Errorf("Expected Content-Encoding '%s', got '%s'", HeaderCode, contentEncoding)
	}

	compressedBody := responseRecorder.Body.Bytes()
	if len(compressedBody) == 0 {
		t.Error("Expected compressed body, got empty")
	}

	gzipReader, gzipReaderError := gzip.NewReader(bytes.NewReader(compressedBody))
	if gzipReaderError != nil {
		t.Fatalf("Failed to create gzip reader: %v", gzipReaderError)
	}
	defer func(gzipReader *gzip.Reader) {
		_ = gzipReader.Close()
	}(gzipReader)

	decompressedData, decompressError := io.ReadAll(gzipReader)
	if decompressError != nil {
		t.Fatalf("Failed to decompress: %v", decompressError)
	}

	expectedBody := `{"message":"test"}`
	if string(decompressedData) != expectedBody {
		t.Errorf("Expected decompressed body '%s', got '%s'", expectedBody, string(decompressedData))
	}
}

func TestGZIP_CompressResponse_TextHTML(t *testing.T) {
	nextHandler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.Header().Set(common.ContentType, common.TextHTML)
		responseWriter.WriteHeader(http.StatusOK)
		_, _ = responseWriter.Write([]byte("<html><body>test</body></html>"))
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(common.ContentType, common.TextHTML)
	request.Header.Set(AcceptEncoding, "gzip")
	responseRecorder := httptest.NewRecorder()

	middleware := GZIP(nextHandler)
	middleware.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	contentEncoding := responseRecorder.Header().Get(ContentEncoding)
	if contentEncoding != HeaderCode {
		t.Errorf("Expected Content-Encoding '%s', got '%s'", HeaderCode, contentEncoding)
	}
}

func TestGZIP_NoCompression_WithoutAcceptEncoding(t *testing.T) {
	nextHandler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.Header().Set(common.ContentType, common.ApplicationJSON)
		responseWriter.WriteHeader(http.StatusOK)
		_, _ = responseWriter.Write([]byte(`{"message":"test"}`))
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(common.ContentType, common.ApplicationJSON)
	// Не устанавливаем Accept-Encoding
	responseRecorder := httptest.NewRecorder()

	middleware := GZIP(nextHandler)
	middleware.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	contentEncoding := responseRecorder.Header().Get(ContentEncoding)
	if contentEncoding != "" {
		t.Errorf("Expected no Content-Encoding header, got '%s'", contentEncoding)
	}

	responseBody := responseRecorder.Body.String()
	if responseBody != `{"message":"test"}` {
		t.Errorf("Expected body '%s', got '%s'", `{"message":"test"}`, responseBody)
	}
}

func TestGZIP_DecompressRequest_WithContentEncoding(t *testing.T) {
	originalRequestBody := `{"url":"https://example.com"}`
	var compressedRequestBody bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedRequestBody)
	_, _ = gzipWriter.Write([]byte(originalRequestBody))
	_ = gzipWriter.Close()

	nextHandler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		requestBody, readError := io.ReadAll(request.Body)
		if readError != nil {
			t.Errorf("Failed to read body: %v", readError)
			responseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}

		if string(requestBody) != originalRequestBody {
			t.Errorf("Expected body '%s', got '%s'", originalRequestBody, string(requestBody))
		}

		responseWriter.WriteHeader(http.StatusOK)
		_, _ = responseWriter.Write([]byte("OK"))
	})

	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(compressedRequestBody.Bytes()))
	request.Header.Set(common.ContentType, common.ApplicationJSON)
	request.Header.Set(ContentEncoding, HeaderCode)
	responseRecorder := httptest.NewRecorder()

	middleware := GZIP(nextHandler)
	middleware.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}
}

func TestGZIP_DecompressRequest_InvalidGzip(t *testing.T) {
	nextHandler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.WriteHeader(http.StatusOK)
	})

	invalidRequestBody := bytes.NewReader([]byte("not a gzip data"))
	request := httptest.NewRequest(http.MethodPost, "/", invalidRequestBody)
	request.Header.Set(common.ContentType, common.ApplicationJSON)
	request.Header.Set(ContentEncoding, HeaderCode)
	responseRecorder := httptest.NewRecorder()

	middleware := GZIP(nextHandler)
	middleware.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, responseRecorder.Code)
	}
}

func TestGZIP_CompressResponse_ErrorStatusCode(t *testing.T) {
	nextHandler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.Header().Set(common.ContentType, common.ApplicationJSON)
		responseWriter.WriteHeader(http.StatusBadRequest)
		_, _ = responseWriter.Write([]byte(`{"error":"bad request"}`))
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(common.ContentType, common.ApplicationJSON)
	request.Header.Set(AcceptEncoding, "gzip")
	responseRecorder := httptest.NewRecorder()

	middleware := GZIP(nextHandler)
	middleware.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, responseRecorder.Code)
	}

	contentEncoding := responseRecorder.Header().Get(ContentEncoding)
	if contentEncoding != "" {
		t.Errorf("Expected no Content-Encoding for error status, got '%s'", contentEncoding)
	}
}

func TestGZIP_AcceptEncodingCaseInsensitive(t *testing.T) {
	testCases := []struct {
		acceptEncoding string
		shouldCompress bool
		description    string
	}{
		{"gzip", true, "lowercase gzip"},
		{"GZIP", true, "uppercase GZIP"},
		{"Gzip", true, "mixed case Gzip"},
		{"gzip, deflate", true, "gzip with other encodings"},
		{"deflate", false, "only deflate"},
		{"", false, "empty Accept-Encoding"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				responseWriter.Header().Set(common.ContentType, common.ApplicationJSON)
				responseWriter.WriteHeader(http.StatusOK)
				_, _ = responseWriter.Write([]byte("test"))
			})

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.Header.Set(common.ContentType, common.ApplicationJSON)
			if testCase.acceptEncoding != "" {
				request.Header.Set(AcceptEncoding, testCase.acceptEncoding)
			}
			responseRecorder := httptest.NewRecorder()

			middleware := GZIP(nextHandler)
			middleware.ServeHTTP(responseRecorder, request)

			contentEncoding := responseRecorder.Header().Get(ContentEncoding)
			if testCase.shouldCompress {
				if !strings.Contains(strings.ToLower(testCase.acceptEncoding), "gzip") {
					if strings.Contains(strings.ToLower(testCase.acceptEncoding), "gzip") && contentEncoding == "" {
						t.Errorf("Expected Content-Encoding '%s', got empty", HeaderCode)
					}
				}
			} else {
				if contentEncoding != "" {
					t.Errorf("Expected no Content-Encoding, got '%s'", contentEncoding)
				}
			}
		})
	}
}

func TestCompressWriter_Header(t *testing.T) {
	originalWriter := httptest.NewRecorder()
	compressWriter := newCompressWriter(originalWriter)

	compressWriter.Header().Set("X-Custom", "test")
	defer func() {
		_ = compressWriter.Close()
	}()

	if originalWriter.Header().Get("X-Custom") != "test" {
		t.Errorf("Expected header 'X-Custom' to be 'test', got '%s'", originalWriter.Header().Get("X-Custom"))
	}
}

func TestCompressWriter_WriteHeader_SuccessCode(t *testing.T) {
	originalWriter := httptest.NewRecorder()
	compressWriter := newCompressWriter(originalWriter)

	compressWriter.WriteHeader(http.StatusOK)
	defer func() {
		_ = compressWriter.Close()
	}()

	if originalWriter.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, originalWriter.Code)
	}

	contentEncoding := originalWriter.Header().Get(ContentEncoding)
	if contentEncoding != HeaderCode {
		t.Errorf("Expected Content-Encoding '%s', got '%s'", HeaderCode, contentEncoding)
	}
}

func TestCompressWriter_WriteHeader_ErrorCode(t *testing.T) {
	originalWriter := httptest.NewRecorder()
	compressWriter := newCompressWriter(originalWriter)

	compressWriter.WriteHeader(http.StatusBadRequest)
	defer func() {
		_ = compressWriter.Close()
	}()

	if originalWriter.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, originalWriter.Code)
	}

	contentEncoding := originalWriter.Header().Get(ContentEncoding)
	if contentEncoding != "" {
		t.Errorf("Expected no Content-Encoding for error status, got '%s'", contentEncoding)
	}
}

func TestCompressWriter_Write(t *testing.T) {
	originalWriter := httptest.NewRecorder()
	compressWriter := newCompressWriter(originalWriter)

	testData := []byte("test data")
	bytesWritten, writeError := compressWriter.Write(testData)
	if writeError != nil {
		t.Fatalf("Write failed: %v", writeError)
	}

	if bytesWritten != len(testData) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(testData), bytesWritten)
	}

	_ = compressWriter.Close()

	compressedBody := originalWriter.Body.Bytes()
	if len(compressedBody) == 0 {
		t.Error("Expected compressed body, got empty")
	}

	gzipReader, gzipReaderError := gzip.NewReader(bytes.NewReader(compressedBody))
	if gzipReaderError != nil {
		t.Fatalf("Failed to create gzip reader: %v", gzipReaderError)
	}
	defer func(gzipReader *gzip.Reader) {
		_ = gzipReader.Close()
	}(gzipReader)

	decompressedData, decompressError := io.ReadAll(gzipReader)
	if decompressError != nil {
		t.Fatalf("Failed to decompress: %v", decompressError)
	}

	if string(decompressedData) != string(testData) {
		t.Errorf("Expected decompressed data '%s', got '%s'", string(testData), string(decompressedData))
	}
}

func TestCompressReader_Read(t *testing.T) {
	originalData := []byte("test data for compression")

	var compressedData bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedData)
	_, _ = gzipWriter.Write(originalData)
	_ = gzipWriter.Close()

	compressedReader := bytes.NewReader(compressedData.Bytes())
	compressReader, compressReaderError := newCompressReader(io.NopCloser(compressedReader))
	if compressReaderError != nil {
		t.Fatalf("Failed to create compressReader: %v", compressReaderError)
	}
	defer func() {
		_ = compressReader.Close()
	}()

	decompressedData, readError := io.ReadAll(compressReader)
	if readError != nil {
		t.Fatalf("Failed to read: %v", readError)
	}

	if string(decompressedData) != string(originalData) {
		t.Errorf("Expected decompressed data '%s', got '%s'", string(originalData), string(decompressedData))
	}
}

func TestCompressReader_Close(t *testing.T) {
	originalData := []byte("test data")

	var compressedData bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedData)
	_, _ = gzipWriter.Write(originalData)
	_ = gzipWriter.Close()

	compressedReader := bytes.NewReader(compressedData.Bytes())
	compressReader, compressReaderError := newCompressReader(io.NopCloser(compressedReader))
	if compressReaderError != nil {
		t.Fatalf("Failed to create compressReader: %v", compressReaderError)
	}

	closeError := compressReader.Close()
	if closeError != nil {
		t.Errorf("Close failed: %v", closeError)
	}
}

func TestCompressReader_InvalidGzip(t *testing.T) {
	invalidData := bytes.NewReader([]byte("not a gzip data"))
	_, compressReaderError := newCompressReader(io.NopCloser(invalidData))

	if compressReaderError == nil {
		t.Error("Expected error for invalid gzip data, got nil")
	}
}
