package statistic

import (
	"context"
	"errors"
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/service/stats"
)

type mockStatsRepo struct {
	urls    int
	users   int
	urlErr  error
	userErr error
}

func (m *mockStatsRepo) CountURLs(_ context.Context) (int, error) {
	return m.urls, m.urlErr
}

func (m *mockStatsRepo) CountUsers(_ context.Context) (int, error) {
	return m.users, m.userErr
}

func TestUsecase_GetStats_OK(t *testing.T) {
	u := NewUsecase(stats.NewService(&mockStatsRepo{urls: 5, users: 2}))
	got, err := u.GetStats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.URLs != 5 || got.Users != 2 {
		t.Fatalf("got %+v", got)
	}
}

func TestUsecase_GetStats_Errors(t *testing.T) {
	t.Run("count urls", func(t *testing.T) {
		u := NewUsecase(stats.NewService(&mockStatsRepo{urlErr: errors.New("db down")}))
		_, err := u.GetStats(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("count users", func(t *testing.T) {
		u := NewUsecase(stats.NewService(&mockStatsRepo{urls: 1, userErr: errors.New("db down")}))
		_, err := u.GetStats(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
