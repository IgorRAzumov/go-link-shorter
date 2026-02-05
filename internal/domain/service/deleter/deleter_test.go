package deleter

import (
	"context"
	"testing"
	"time"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/repository"
)

type markCall struct {
	userID string
	keys   []string
}

type mockRepo struct {
	markDeleted func(ctx context.Context, userID string, shortKeys []string) error
}

var _ repository.LinkRepository = (*mockRepo)(nil)

func (m *mockRepo) GetByShortKey(ctx context.Context, shortURL string) (string, bool) {
	return "", false
}
func (m *mockRepo) GetShortKeyByURL(ctx context.Context, URL string) string   { return "" }
func (m *mockRepo) IsExistShortKey(ctx context.Context, shortURL string) bool { return false }
func (m *mockRepo) Save(ctx context.Context, link *model.Link) error          { return nil }
func (m *mockRepo) BatchSave(ctx context.Context, links []*model.Link) error  { return nil }
func (m *mockRepo) GetByUserID(ctx context.Context, userID string) ([]*model.Link, error) {
	return []*model.Link{}, nil
}
func (m *mockRepo) MarkDeleted(ctx context.Context, userID string, shortKeys []string) error {
	if m.markDeleted != nil {
		return m.markDeleted(ctx, userID, shortKeys)
	}
	return nil
}

func TestEnqueue_EmptyUserID_ReturnsError(t *testing.T) {
	s := NewDeleterService(&mockRepo{}, DefaultMaxBatch)
	if err := s.Enqueue("", []string{"a"}); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestService_FlushesOnBatchSize(t *testing.T) {
	calls := make(chan markCall, 1)
	repo := &mockRepo{
		markDeleted: func(ctx context.Context, userID string, shortKeys []string) error {
			keysCopy := append([]string(nil), shortKeys...)
			calls <- markCall{userID: userID, keys: keysCopy}
			return nil
		},
	}

	s := NewDeleterService(repo, 2)
	s.Start()
	defer func() {
		s.Stop()
		s.Wait()
	}()

	if err := s.Enqueue("u1", []string{"a", "b"}); err != nil {
		t.Fatalf("enqueue error: %v", err)
	}

	select {
	case got := <-calls:
		if got.userID != "u1" {
			t.Fatalf("expected userID %q, got %q", "u1", got.userID)
		}
		if len(got.keys) != 2 {
			t.Fatalf("expected 2 keys, got %#v", got.keys)
		}
		set := map[string]struct{}{got.keys[0]: {}, got.keys[1]: {}}
		if _, ok := set["a"]; !ok {
			t.Fatalf("expected key %q, got %#v", "a", got.keys)
		}
		if _, ok := set["b"]; !ok {
			t.Fatalf("expected key %q, got %#v", "b", got.keys)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for MarkDeleted call")
	}
}

func TestService_GroupsByUserInSingleFlush(t *testing.T) {
	calls := make(chan markCall, 2)
	repo := &mockRepo{
		markDeleted: func(ctx context.Context, userID string, shortKeys []string) error {
			keysCopy := append([]string(nil), shortKeys...)
			calls <- markCall{userID: userID, keys: keysCopy}
			return nil
		},
	}

	s := NewDeleterService(repo, 2)
	s.Start()
	defer func() {
		s.Stop()
		s.Wait()
	}()

	if err := s.Enqueue("u1", []string{"a"}); err != nil {
		t.Fatalf("enqueue error: %v", err)
	}
	if err := s.Enqueue("u2", []string{"b"}); err != nil {
		t.Fatalf("enqueue error: %v", err)
	}

	got := map[string][]string{}
	for i := 0; i < 2; i++ {
		select {
		case c := <-calls:
			got[c.userID] = c.keys
		case <-time.After(500 * time.Millisecond):
			t.Fatal("timed out waiting for MarkDeleted calls")
		}
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 users, got %#v", got)
	}
	if keys := got["u1"]; len(keys) != 1 || keys[0] != "a" {
		t.Fatalf("expected u1 keys [a], got %#v", keys)
	}
	if keys := got["u2"]; len(keys) != 1 || keys[0] != "b" {
		t.Fatalf("expected u2 keys [b], got %#v", keys)
	}
}
