package rest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestListenAndServeTLS_MissingCertFiles_ReturnsError(t *testing.T) {
	builder := &Builder{
		serverAddress: "127.0.0.1:0",
		tlsCertFile:   "nonexistent-cert.pem",
		tlsKeyFile:    "nonexistent-key.pem",
	}
	server := &http.Server{
		Addr:    builder.serverAddress,
		Handler: http.NewServeMux(),
	}

	err := builder.listenAndServeTLS(server)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Expected not-exist error, got: %v", err)
	}
}

func TestListenAndServeTLS_WithCertFiles_AcceptsConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long TLS server test in -short mode")
	}

	certFile, keyFile := writeTempCertKeyPair(t)
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
	writePemOrFail(t, certFile, &pem.Block{Type: "CERTIFICATE", Bytes: der})
	writePemOrFail(t, keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

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

	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		Timeout:   2 * time.Second,
	}
	var resp *http.Response
	for i := 0; i < 100; i++ {
		resp, err = client.Get("https://" + addr + "/")
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
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

func writeTempCertKeyPair(t *testing.T) (certFile, keyFile string) {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem")
}

func writePemOrFail(t *testing.T, path string, block *pem.Block) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create(%s): %v", path, err)
	}
	defer func() {
		_ = f.Close()
	}()
	if err := pem.Encode(f, block); err != nil {
		t.Fatalf("pem.Encode(%s): %v", path, err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("Close(%s): %v", path, err)
	}
}
