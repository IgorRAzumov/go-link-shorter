package rest

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

func (builder *Builder) listenAndServeTLS(server *http.Server) error {
	cert, err := tls.LoadX509KeyPair(builder.tlsCertFile, builder.tlsKeyFile)
	if err != nil {
		return fmt.Errorf(
			"invalid TLS cert/key pair (cert=%q, key=%q): %w",
			builder.tlsCertFile,
			builder.tlsKeyFile,
			err,
		)
	}
	server.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}
	if err := server.ListenAndServeTLS("", ""); err != nil {
		return fmt.Errorf("failed to start HTTP-server with TLS enabled: %w", err)
	}
	return nil
}
