package rest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func (builder *Builder) listenAndServeTLS(server *http.Server) error {
	_, err := tls.LoadX509KeyPair(builder.tlsCertFile, builder.tlsKeyFile)
	if err != nil {
		log.Info().Err(err).Msg("TLS cert/key files not found, generating self-signed certificate")

		cert, err := generateSelfSignedCert()
		if err != nil {
			return err
		}
		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}

		listener, err := tls.Listen("tcp", builder.serverAddress, server.TLSConfig)
		if err != nil {
			return err
		}
		defer func(listener net.Listener) {
			err := listener.Close()
			if err != nil {
				log.Error().Err(err).Msg("Error closing TLS listener")
			}
		}(listener)

		return server.Serve(listener)
	}
	return server.ListenAndServeTLS(builder.tlsCertFile, builder.tlsKeyFile)
}

func generateSelfSignedCert() (tls.Certificate, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Shortener"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1)},
	}

	bytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.Certificate{
		Certificate: [][]byte{bytes},
		PrivateKey:  privateKey,
	}, nil
}
