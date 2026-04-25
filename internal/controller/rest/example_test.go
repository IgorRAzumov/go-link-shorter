package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest"
	commontesting "github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common/testing"
	handlermodel "github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/handler/shorter/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/auth"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/usecase"
)

const exampleSecret = "example-secret"

func makeRouter(createUC usecase.LinkCreateUsecase, readUC usecase.LinkReadUsecase, deleteUC usecase.LinkDeleteUsecase) http.Handler {
	authSvc := auth.NewAuthService(exampleSecret)
	healthUC := &commontesting.MockHealthCheckUsecase{}
	return rest.NewRouter(rest.RouterDeps{
		LinkCreateUsecase: createUC,
		LinkReadUsecase:   readUC,
		LinkDeleteUsecase: deleteUC,
		HealthCheck:       healthUC,
		AuthService:       authSvc,
	})
}

// signedUserCookie возвращает валидную cookie user_id для Example-тестов,
// чтобы middleware auth помечал запрос как аутентифицированный.
func signedUserCookie(userID string) *http.Cookie {
	authSvc := auth.NewAuthService(exampleSecret)
	return &http.Cookie{Name: "user_id", Value: authSvc.SignUserID(userID)}
}

// ExampleNewRouter_shortenText демонстрирует POST / — сокращение URL из тела (Content-Type: text/plain).
func ExampleNewRouter_shortenText() {
	uc := &commontesting.MockLinkCreateUsecase{
		GetBaseURLFunc: func() string { return "http://localhost:8080" },
		CreateShortKeyFunc: func(_ context.Context, URL string) (string, error) {
			return "abc123", nil
		},
	}
	router := makeRouter(uc, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com/page"))
	req.Header.Set("Content-Type", "text/plain")
	req = req.WithContext(authctx.WithUserID(req.Context(), "user1"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	fmt.Println(strings.TrimSpace(w.Body.String()))
	// Output:
	// 201
	// http://localhost:8080/abc123
}

// ExampleNewRouter_shortenJSON демонстрирует POST /api/shorten — сокращение URL (Content-Type: application/json).
func ExampleNewRouter_shortenJSON() {
	uc := &commontesting.MockLinkCreateUsecase{
		GetBaseURLFunc: func() string { return "http://localhost:8080" },
		CreateShortKeyFunc: func(_ context.Context, URL string) (string, error) {
			return "xyz789", nil
		},
	}
	router := makeRouter(uc, nil, nil)

	body, _ := json.Marshal(handlermodel.ShortenRequest{URL: "https://yandex.ru/search"})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(authctx.WithUserID(req.Context(), "user2"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp handlermodel.ShortenResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	fmt.Println(w.Code)
	fmt.Println(resp.Result)
	// Output:
	// 201
	// http://localhost:8080/xyz789
}

// ExampleNewRouter_batchShorten демонстрирует POST /api/shorten/batch — пакетное сокращение URL.
func ExampleNewRouter_batchShorten() {
	uc := &commontesting.MockLinkCreateUsecase{
		GetBaseURLFunc: func() string { return "http://localhost:8080" },
		ProcessBatchShortenRequestsFunc: func(_ context.Context, reqs []handlermodel.BatchShortenRequest, scheme, host string) ([]handlermodel.BatchShortenResponse, error) {
			res := make([]handlermodel.BatchShortenResponse, len(reqs))
			for i, r := range reqs {
				res[i] = handlermodel.BatchShortenResponse{
					CorrelationID: r.CorrelationID,
					ShortURL:      scheme + "://" + host + "/key" + r.CorrelationID,
				}
			}
			return res, nil
		},
	}
	router := makeRouter(uc, nil, nil)

	batch := []handlermodel.BatchShortenRequest{
		{CorrelationID: "1", OriginalURL: "https://a.com"},
		{CorrelationID: "2", OriginalURL: "https://b.com"},
	}
	body, _ := json.Marshal(batch)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(authctx.WithUserID(req.Context(), "user3"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp []handlermodel.BatchShortenResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	fmt.Println(w.Code)
	fmt.Println(len(resp))
	fmt.Println(resp[0].ShortURL)
	// Output:
	// 201
	// 2
	// http://example.com/key1
}

// ExampleNewRouter_userURLs демонстрирует GET /api/user/urls — список ссылок пользователя.
func ExampleNewRouter_userURLs() {
	readUC := &commontesting.MockLinkReadUsecase{
		GetBaseURLFunc: func() string { return "http://localhost:8080" },
		GetUserURLsFunc: func(_ context.Context) ([]*model.Link, error) {
			return []*model.Link{
				{ShortKey: "k1", FullURL: "https://x.com", UserID: "u", DeletedFlag: false},
			}, nil
		},
	}
	router := makeRouter(nil, readUC, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req.AddCookie(signedUserCookie("u"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	fmt.Println(strings.Contains(w.Body.String(), "https://x.com"))
	// Output:
	// 200
	// true
}

// ExampleNewRouter_deleteUserURLs демонстрирует DELETE /api/user/urls — мягкое удаление ссылок.
func ExampleNewRouter_deleteUserURLs() {
	deleteUC := &commontesting.MockLinkDeleteUsecase{
		DeleteUserURLsFunc: func(_ context.Context, keys []string) error {
			return nil
		},
	}
	router := makeRouter(nil, nil, deleteUC)

	body, _ := json.Marshal([]string{"key1", "key2"})
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(authctx.WithUserID(req.Context(), "u"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output:
	// 202
}

// ExampleNewRouter_resolve демонстрирует GET /{shortKey} — редирект на полный URL.
func ExampleNewRouter_resolve() {
	readUC := &commontesting.MockLinkReadUsecase{
		GetFullURLByShortKeyFunc: func(_ context.Context, key string) (string, error) {
			if key == "abc" {
				return "https://example.com/target", nil
			}
			return "", nil
		},
	}
	router := makeRouter(nil, readUC, nil)

	req := httptest.NewRequest(http.MethodGet, "/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	fmt.Println(w.Header().Get("Location"))
	// Output:
	// 307
	// https://example.com/target
}

// ExampleNewRouter_ping демонстрирует GET /ping — проверка доступности БД.
func ExampleNewRouter_ping() {
	healthUC := &commontesting.MockHealthCheckUsecase{
		CheckSystemConnectionsFunc: func(_ context.Context) (bool, error) {
			return true, nil
		},
	}
	authSvc := auth.NewAuthService("example-secret")
	router := rest.NewRouter(rest.RouterDeps{
		HealthCheck: healthUC,
		AuthService: authSvc,
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output:
	// 200
}
