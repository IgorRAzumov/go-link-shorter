package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/adapter/database"
	"github.com/IgorRAzumov/go-link-shorter/internal/adapter/inmemory"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/shorter"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/middleware"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/middleware/auth"
	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/middleware/gzip"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
	authservice "github.com/IgorRAzumov/go-link-shorter/internal/domain/service/auth"
	resolversvc "github.com/IgorRAzumov/go-link-shorter/internal/domain/service/resolver"
	shortersvc "github.com/IgorRAzumov/go-link-shorter/internal/domain/service/shorter"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase/link"
	"github.com/go-chi/chi/v5"
)

func getBenchDBDSN(b *testing.B) string {
	dsn := os.Getenv("BENCH_DATABASE_DSN")
	if dsn == "" {
		dsn = os.Getenv("TEST_DATABASE_DSN")
	}
	if dsn == "" {
		b.Skip("BENCH_DATABASE_DSN or TEST_DATABASE_DSN not set, skipping DB benchmark")
	}
	return dsn
}

func BenchmarkAPIHandler_Shorten(b *testing.B) {
	router := benchRouter()
	body := []byte(`{"url":"https://example.com/long/url/path"}`)

	for b.Loop() {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(authctx.WithUserID(req.Context(), "user-1"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
	}
}

func BenchmarkBatchAPIHandler(b *testing.B) {
	router := benchRouter()
	requests := make([]map[string]string, 50)
	for i := range requests {
		requests[i] = map[string]string{
			"correlation_id": fmt.Sprintf("corr-%d", i),
			"original_url":   fmt.Sprintf("https://example.com/batch/%d", i),
		}
	}
	body, _ := json.Marshal(requests)

	for b.Loop() {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(authctx.WithUserID(req.Context(), "user-1"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
	}
}

func BenchmarkAPIHandler_Shorten_DB(b *testing.B) {
	router := benchRouterDB(b)

	var i int
	for b.Loop() {
		body := []byte(fmt.Sprintf(`{"url":"https://example.com/db/shorten/%d"}`, i))
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(authctx.WithUserID(req.Context(), "user-1"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		i++
	}
}

func BenchmarkBatchAPIHandler_DB(b *testing.B) {
	router := benchRouterDB(b)

	var i int
	for b.Loop() {
		requests := make([]map[string]string, 50)
		for j := range requests {
			requests[j] = map[string]string{
				"correlation_id": fmt.Sprintf("corr-%d", j),
				"original_url":   fmt.Sprintf("https://example.com/batch/db/%d/%d", i, j),
			}
		}
		body, _ := json.Marshal(requests)
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(authctx.WithUserID(req.Context(), "user-1"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		i++
	}
}

func benchRouter() http.Handler {
	storage, _ := inmemory.NewInMemoryFileStorage("")
	resolverSvc := resolversvc.NewResolverService(storage)
	authSvc := authservice.NewAuthService("test-secret")
	shorterSvc := shortersvc.NewShorterService(storage)

	createUsecase := link.NewLinkCreateUsecase(resolverSvc, shorterSvc, "http://localhost:8080")

	router := chi.NewRouter()
	router.Use(middleware.HTTPLogger, gzip.GZIP, auth.Middleware(authSvc))
	router.Post("/api/shorten", shorter.APIHandler(createUsecase, nil))
	router.Post("/api/shorten/batch", shorter.BatchAPIHandler(createUsecase))
	return router
}

func benchRouterDB(b *testing.B) http.Handler {
	storage, err := database.NewStorage(getBenchDBDSN(b))
	if err != nil {
		b.Fatalf("NewStorage failed: %v", err)
	}
	b.Cleanup(func() { _ = storage.Close() })

	resolverSvc := resolversvc.NewResolverService(storage)
	authSvc := authservice.NewAuthService("test-secret")
	shorterSvc := shortersvc.NewShorterService(storage)

	createUsecase := link.NewLinkCreateUsecase(resolverSvc, shorterSvc, "http://localhost:8080")

	router := chi.NewRouter()
	router.Use(middleware.HTTPLogger, gzip.GZIP, auth.Middleware(authSvc))
	router.Post("/api/shorten", shorter.APIHandler(createUsecase, nil))
	router.Post("/api/shorten/batch", shorter.BatchAPIHandler(createUsecase))
	return router
}
