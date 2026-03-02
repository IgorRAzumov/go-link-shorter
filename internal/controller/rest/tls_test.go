package rest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGenerateSelfSignedCert_ProducesValidCertificate(t *testing.T) {
	cert, err := generateSelfSignedCert()
	if err != nil {
		t.Fatalf("generateSelfSignedCert() error: %v", err)
	}

	if len(cert.Certificate) == 0 {
		t.Fatal("Expected at least one certificate in chain")
	}

	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("ParseCertificate error: %v", err)
	}

	if len(x509Cert.Subject.Organization) == 0 {
		t.Error("Expected Organization in certificate subject")
	} else if x509Cert.Subject.Organization[0] != "Shortener" {
		t.Errorf("Expected Organization Shortener, got %q", x509Cert.Subject.Organization[0])
	}

	if x509Cert.NotBefore.After(time.Now()) {
		t.Error("Certificate NotBefore should be in the past")
	}
	if x509Cert.NotAfter.Before(time.Now()) {
		t.Error("Certificate NotAfter should be in the future")
	}

	// Verify certificate can be used for TLS
	if cert.PrivateKey == nil {
		t.Error("Expected PrivateKey to be set")
	}
}

func TestListenAndServeTLS_WithSelfSignedCert_AcceptsConnections(t *testing.T) {
	// Use a random free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen error: %v", err)
	}
	addr := listener.Addr().String()
	_ = listener.Close()

	builder := &Builder{
		serverAddress: addr,
		tlsCertFile:   "nonexistent-cert.pem",
		tlsKeyFile:    "nonexistent-key.pem",
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	done := make(chan error, 1)
	go func() {
		done <- builder.listenAndServeTLS(server)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Connect with InsecureSkipVerify since we use self-signed cert
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get("https://" + addr + "/")
	if err != nil {
		t.Fatalf("HTTPS GET error: %v", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Shutdown server
	if err := server.Close(); err != nil {
		t.Logf("server.Close(): %v", err)
	}

	<-done
}

func TestListenAndServeTLS_WithCertFiles_UsesListenAndServeTLS(t *testing.T) {
	// Create temp cert and key files
	dir := t.TempDir()
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{Organization: []string{"Test"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("CreateCertificate: %v", err)
	}
	certOut, _ := os.Create(certFile)
	_ = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: der})
	_ = certOut.Close()

	keyOut, _ := os.Create(keyFile)
	_ = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	_ = keyOut.Close()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	addr := listener.Addr().String()
	_ = listener.Close()

	builder := &Builder{
		serverAddress: addr,
		tlsCertFile:   certFile,
		tlsKeyFile:    keyFile,
	}

	server := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}

	done := make(chan error, 1)
	go func() {
		done <- builder.listenAndServeTLS(server)
	}()

	time.Sleep(100 * time.Millisecond)

	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		Timeout:   2 * time.Second,
	}
	resp, err := client.Get("https://" + addr + "/")
	if err != nil {
		t.Fatalf("HTTPS GET: %v", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	_ = server.Close()
	<-done
}
