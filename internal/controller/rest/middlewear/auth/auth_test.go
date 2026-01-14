package auth

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/authctx"
	authservice "github.com/IgorRAzumov/go-link-shorter/internal/domain/service/auth"
)

func TestAuthMiddleware_CreatesCookieIfNotExists(t *testing.T) {
	secretKey := "test-secret-key"
	authSvc := authservice.NewAuthService(secretKey)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := authctx.UserID(r.Context())
		if userID == "" {
			t.Error("Expected user ID in context")
		}
	})

	middleware := Middleware(authSvc)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	cookies := rr.Result().Cookies()
	if len(cookies) == 0 {
		t.Error("Expected cookie to be set")
	}

	cookie := cookies[0]
	if cookie.Name != "user_id" {
		t.Errorf("Expected cookie name %s, got %s", "user_id", cookie.Name)
	}
}

func TestAuthMiddleware_ValidatesExistingCookie(t *testing.T) {
	secretKey := "test-secret-key"
	authSvc := authservice.NewAuthService(secretKey)
	var capturedUserID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = authctx.UserID(r.Context())
		if capturedUserID == "" {
			t.Error("Expected user ID in context")
		}
	})

	middleware := Middleware(authSvc)
	req := httptest.NewRequest("GET", "/", nil)

	// Create a valid cookie
	testUserID := "test-user-id"
	value := authSvc.SignUserID(testUserID)
	req.AddCookie(&http.Cookie{
		Name:  "user_id",
		Value: value,
	})

	rr := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr, req)

	if capturedUserID != testUserID {
		t.Errorf("Expected user ID %s, got %s", testUserID, capturedUserID)
	}
}

func TestAuthMiddleware_RejectsInvalidCookie(t *testing.T) {
	secretKey := "test-secret-key"
	authSvc := authservice.NewAuthService(secretKey)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := authctx.UserID(r.Context())
		if userID == "" {
			t.Error("Expected user ID in context")
		}
	})

	middleware := Middleware(authSvc)
	req := httptest.NewRequest("GET", "/", nil)

	// Create an invalid cookie with wrong signature
	testUserID := "test-user-id"
	value := base64.URLEncoding.EncodeToString([]byte(testUserID)) + ".invalid-signature"
	req.AddCookie(&http.Cookie{
		Name:  "user_id",
		Value: value,
	})

	rr := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr, req)

	// Should create a new cookie since the old one is invalid
	cookies := rr.Result().Cookies()
	if len(cookies) == 0 {
		t.Error("Expected new cookie to be set")
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	ctx := context.Background()
	userID := authctx.UserID(ctx)
	if userID != "" {
		t.Errorf("Expected empty user ID, got %s", userID)
	}

	ctx = authctx.WithUserID(ctx, "test-user-id")
	userID = authctx.UserID(ctx)
	if userID != "test-user-id" {
		t.Errorf("Expected user ID test-user-id, got %s", userID)
	}
}
