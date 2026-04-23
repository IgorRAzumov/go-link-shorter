package trustedsubnet

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mustParseCIDR(t *testing.T, cidr string) *net.IPNet {
	t.Helper()
	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatalf("ParseCIDR(%q) error: %v", cidr, err)
	}
	return subnet
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestMiddleware_NilSubnet_Forbidden(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	req.Header.Set("X-Real-IP", "10.0.0.1")
	rr := httptest.NewRecorder()

	Middleware(nil)(okHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 for nil subnet, got %d", rr.Code)
	}
}

func TestMiddleware_MissingHeader_Forbidden(t *testing.T) {
	subnet := mustParseCIDR(t, "10.0.0.0/8")
	req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	rr := httptest.NewRecorder()

	Middleware(subnet)(okHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 without X-Real-IP, got %d", rr.Code)
	}
}

func TestMiddleware_InvalidIP_Forbidden(t *testing.T) {
	subnet := mustParseCIDR(t, "10.0.0.0/8")
	req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	req.Header.Set("X-Real-IP", "not-an-ip")
	rr := httptest.NewRecorder()

	Middleware(subnet)(okHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 for invalid IP, got %d", rr.Code)
	}
}

func TestMiddleware_OutsideSubnet_Forbidden(t *testing.T) {
	subnet := mustParseCIDR(t, "10.0.0.0/8")
	req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	req.Header.Set("X-Real-IP", "192.168.1.1")
	rr := httptest.NewRecorder()

	Middleware(subnet)(okHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 for outside IP, got %d", rr.Code)
	}
}

func TestMiddleware_InsideSubnet_OK(t *testing.T) {
	subnet := mustParseCIDR(t, "10.0.0.0/8")
	req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	req.Header.Set("X-Real-IP", "10.1.2.3")
	rr := httptest.NewRecorder()

	Middleware(subnet)(okHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for IP in subnet, got %d", rr.Code)
	}
}
