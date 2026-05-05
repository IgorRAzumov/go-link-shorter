package shortenergrpc_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/api/shortenerpb"
	shortenergrpc "github.com/IgorRAzumov/go-link-shorter/internal/controller/grpc/shortener"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	authservice "github.com/IgorRAzumov/go-link-shorter/internal/domain/service/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

// --- мок-usecase-ы ---

type mockCreate struct {
	baseURL string
	key     string
	keyErr  error

	gotURL string
}

func (m *mockCreate) CreateShortKey(_ context.Context, u string) (string, error) {
	m.gotURL = u
	return m.key, m.keyErr
}
func (m *mockCreate) CreateShortKeysBatch(_ context.Context, _ []string) (map[string]string, error) {
	return nil, errors.New("not implemented")
}
func (m *mockCreate) ProcessBatchShortenRequests(_ context.Context, _ []model.BatchShortenRequest) ([]model.BatchShortenResult, error) {
	return nil, errors.New("not implemented")
}
func (m *mockCreate) GetBaseURL() string { return m.baseURL }

type mockRead struct {
	baseURL string
	full    string
	fullErr error
	urls    []*model.Link
	urlsErr error
}

func (m *mockRead) GetFullURLByShorKey(_ context.Context, _ string) (string, error) {
	return m.full, m.fullErr
}
func (m *mockRead) GetUserURLs(_ context.Context) ([]*model.Link, error) {
	return m.urls, m.urlsErr
}
func (m *mockRead) GetBaseURL() string { return m.baseURL }

// --- общий bufconn-стенд ---

const secret = "grpc-test-secret"
const baseURL = "http://short.test"

type testEnv struct {
	t       *testing.T
	client  shortenerpb.ShortenerServiceClient
	authSvc *authservice.Service
	create  *mockCreate
	read    *mockRead
	conn    *grpc.ClientConn
	server  *grpc.Server
}

func newTestEnv(t *testing.T, create *mockCreate, read *mockRead) *testEnv {
	t.Helper()

	authSvc := authservice.NewAuthService(secret)
	srv, err := shortenergrpc.NewGRPCServer(authSvc, shortenergrpc.TLSConfig{})
	if err != nil {
		t.Fatalf("NewGRPCServer: %v", err)
	}
	shortenergrpc.NewServer(authSvc, create, read, baseURL).Register(srv)

	lis := bufconn.Listen(1 << 16)
	go func() {
		if serveErr := srv.Serve(lis); serveErr != nil && !errors.Is(serveErr, grpc.ErrServerStopped) {
			t.Logf("server exited: %v", serveErr)
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) { return lis.Dial() }
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc.NewClient: %v", err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
		srv.Stop()
	})

	return &testEnv{
		t:       t,
		client:  shortenerpb.NewShortenerServiceClient(conn),
		authSvc: authSvc,
		create:  create,
		read:    read,
		conn:    conn,
		server:  srv,
	}
}

func (e *testEnv) authorizedContext(userID string) context.Context {
	signed := e.authSvc.SignUserID(userID)
	return metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", signed))
}

// --- тесты ---

func TestShortenURL_Success(t *testing.T) {
	env := newTestEnv(t, &mockCreate{key: "abc"}, &mockRead{})

	resp, err := env.client.ShortenURL(context.Background(), &shortenerpb.URLShortenRequest{Url: "https://example.com/x"})
	if err != nil {
		t.Fatalf("ShortenURL: %v", err)
	}
	if want := baseURL + "/abc"; resp.GetResult() != want {
		t.Fatalf("result=%q want %q", resp.GetResult(), want)
	}
}

func TestShortenURL_InvalidURL(t *testing.T) {
	env := newTestEnv(t, &mockCreate{}, &mockRead{})

	_, err := env.client.ShortenURL(context.Background(), &shortenerpb.URLShortenRequest{Url: "not-a-url"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code=%v", status.Code(err))
	}
}

func TestShortenURL_Conflict(t *testing.T) {
	conflict := &model.URLConflictError{ExistingShortKey: "dup"}
	env := newTestEnv(t, &mockCreate{keyErr: conflict}, &mockRead{})

	resp, err := env.client.ShortenURL(context.Background(), &shortenerpb.URLShortenRequest{Url: "https://example.com/y"})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("code=%v err=%v", status.Code(err), err)
	}
	if resp != nil && resp.GetResult() != "" {
		t.Fatalf("response body is not transmitted for non-OK gRPC statuses; got %q", resp.GetResult())
	}
}

func TestExpandURL_OK(t *testing.T) {
	env := newTestEnv(t, &mockCreate{}, &mockRead{full: "https://final.example/a"})

	resp, err := env.client.ExpandURL(context.Background(), &shortenerpb.URLExpandRequest{Id: "abc"})
	if err != nil {
		t.Fatalf("ExpandURL: %v", err)
	}
	if resp.GetResult() != "https://final.example/a" {
		t.Fatalf("result=%q", resp.GetResult())
	}
}

func TestExpandURL_NotFound(t *testing.T) {
	env := newTestEnv(t, &mockCreate{}, &mockRead{fullErr: errors.New("not found")})

	_, err := env.client.ExpandURL(context.Background(), &shortenerpb.URLExpandRequest{Id: "missing"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("code=%v", status.Code(err))
	}
}

func TestExpandURL_Deleted(t *testing.T) {
	env := newTestEnv(t, &mockCreate{}, &mockRead{fullErr: model.ErrURLDeleted})

	_, err := env.client.ExpandURL(context.Background(), &shortenerpb.URLExpandRequest{Id: "x"})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("code=%v", status.Code(err))
	}
}

func TestExpandURL_EmptyID(t *testing.T) {
	env := newTestEnv(t, &mockCreate{}, &mockRead{})

	_, err := env.client.ExpandURL(context.Background(), &shortenerpb.URLExpandRequest{Id: ""})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code=%v", status.Code(err))
	}
}

func TestListUserURLs_Unauthenticated(t *testing.T) {
	env := newTestEnv(t, &mockCreate{}, &mockRead{urls: []*model.Link{{ShortKey: "k", FullURL: "u"}}})

	_, err := env.client.ListUserURLs(context.Background(), &emptypb.Empty{})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("code=%v", status.Code(err))
	}
}

func TestListUserURLs_OK(t *testing.T) {
	env := newTestEnv(t, &mockCreate{}, &mockRead{urls: []*model.Link{
		{ShortKey: "k1", FullURL: "https://a"},
		{ShortKey: "k2", FullURL: "https://b"},
	}})

	resp, err := env.client.ListUserURLs(env.authorizedContext("user-1"), &emptypb.Empty{})
	if err != nil {
		t.Fatalf("ListUserURLs: %v", err)
	}
	if len(resp.GetUrl()) != 2 {
		t.Fatalf("got %d urls", len(resp.GetUrl()))
	}
	if resp.GetUrl()[0].GetShortUrl() != baseURL+"/k1" {
		t.Fatalf("short url=%q", resp.GetUrl()[0].GetShortUrl())
	}
	if resp.GetUrl()[1].GetOriginalUrl() != "https://b" {
		t.Fatalf("original url=%q", resp.GetUrl()[1].GetOriginalUrl())
	}
}

func TestAuth_InvalidTokenTreatedAsAnonymous(t *testing.T) {
	env := newTestEnv(t, &mockCreate{}, &mockRead{urls: []*model.Link{{ShortKey: "k"}}})

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "totally-bogus"))
	_, err := env.client.ListUserURLs(ctx, &emptypb.Empty{})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("code=%v", status.Code(err))
	}
}

func TestAuth_BearerPrefixAccepted(t *testing.T) {
	env := newTestEnv(t, &mockCreate{}, &mockRead{})

	signed := env.authSvc.SignUserID("alice")
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+signed))

	if _, err := env.client.ListUserURLs(ctx, &emptypb.Empty{}); err != nil {
		t.Fatalf("ListUserURLs: %v", err)
	}
}

func TestNewGRPCServer_TLSMissingFiles(t *testing.T) {
	auth := authservice.NewAuthService(secret)
	_, err := shortenergrpc.NewGRPCServer(auth, shortenergrpc.TLSConfig{
		Enable:   true,
		CertFile: "/nonexistent/cert.pem",
		KeyFile:  "/nonexistent/key.pem",
	})
	if err == nil {
		t.Fatal("expected error for missing tls files")
	}
}
