package trustedsubnet

import (
	"net"
	"net/http"
)

// Middleware возвращает middleware, допускающий только запросы, у которых
// заголовок X-Real-IP содержит IP, принадлежащий trustedSubnet.
// Если trustedSubnet == nil, все запросы отклоняются со статусом 403 Forbidden.
func Middleware(trustedSubnet *net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if trustedSubnet == nil {
				http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			realIP := request.Header.Get("X-Real-IP")
			ip := net.ParseIP(realIP)
			if ip == nil || !trustedSubnet.Contains(ip) {
				http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			next.ServeHTTP(writer, request)
		})
	}
}
