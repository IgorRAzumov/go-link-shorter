package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Storage struct {
	db *sql.DB
}

func NewStorage(dsn string) (*Storage, error) {
	if dsn == "" {
		return &Storage{}, nil
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		closeErr := db.Close()
		if closeErr != nil {
			log.Error().Err(closeErr).Msg("Error closing database connection")
		}
		return nil, err
	}

	storage := &Storage{db: db}

	if err := storage.runMigrations(); err != nil {
		closeErr := db.Close()
		if closeErr != nil {
			log.Error().Err(closeErr).Msg("Error closing database connection")
		}
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info().Msg("Database connection established successfully")
	return storage, nil
}

func (storage *Storage) runMigrations() error {
	driver, err := postgres.WithInstance(storage.db, &postgres.Config{})
	if err != nil {
		return err
	}

	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Info().Msg("Database migrations completed successfully")
	return nil
}

func (storage *Storage) CheckStorageConnection(ctx context.Context) (bool, error) {
	if storage.db == nil {
		return false, nil
	}

	if err := storage.db.PingContext(ctx); err != nil {
		log.Error().Err(err).Msg("Database connection check failed")
		return false, err
	}

	return true, nil
}

func (storage *Storage) Close() error {
	if storage.db == nil {
		return nil
	}
	return storage.db.Close()
}

func (storage *Storage) GetByShortKey(ctx context.Context, shortURL string) string {
	if storage.db == nil {
		return ""
	}

	var fullURL string
	err := storage.db.QueryRowContext(ctx, "SELECT full_url FROM links WHERE short_key = $1", shortURL).Scan(&fullURL)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Error().Err(err).Str("short_key", shortURL).Msg("Failed to get link by short key")
		}
		return ""
	}

	return fullURL
}

func (storage *Storage) GetShortKeyByURL(ctx context.Context, URL string) string {
	if storage.db == nil {
		return ""
	}

	normalizedURL := common.NormalizeURL(URL)
	var shortKey string
	err := storage.db.QueryRowContext(ctx, "SELECT short_key FROM links WHERE full_url = $1", normalizedURL).Scan(&shortKey)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Error().Err(err).Str("url", normalizedURL).Msg("Failed to get short key by URL")
		}
		return ""
	}

	return shortKey
}

func (storage *Storage) IsExistShortKey(ctx context.Context, shortURL string) bool {
	if storage.db == nil {
		return false
	}

	var exists bool
	err := storage.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM links WHERE short_key = $1)", shortURL).Scan(&exists)
	if err != nil {
		log.Error().Err(err).Str("short_key", shortURL).Msg("Failed to check if short key exists")
		return false
	}

	return exists
}

func (storage *Storage) Save(ctx context.Context, link *model.Link) {
	if storage.db == nil {
		return
	}

	normalizedURL := common.NormalizeURL(link.FullURL)
	linkUUID := uuid.New().String()

	_, err := storage.db.ExecContext(ctx,
		"INSERT INTO links (uuid, short_key, full_url) VALUES ($1, $2, $3) ON CONFLICT (short_key) DO UPDATE SET full_url = EXCLUDED.full_url",
		linkUUID, link.ShortKey, normalizedURL)
	if err != nil {
		log.Error().Err(err).Str("short_key", link.ShortKey).Str("url", normalizedURL).Msg("Failed to save link")
	}
}
