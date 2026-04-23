package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/IgorRAzumov/go-link-shorter/internal/controller/rest/common"
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Storage — PostgreSQL-реализация LinkRepository.
type Storage struct {
	db *sql.DB
}

// NewStorage создаёт хранилище ссылок в PostgreSQL. При dsn == "" возвращает пустой Storage (nil db).
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

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
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

func (storage *Storage) GetByShortKey(ctx context.Context, shortURL string) (fullURL string, isDeleted bool) {
	if storage.db == nil {
		return "", false
	}

	err := storage.db.QueryRowContext(ctx, "SELECT full_url, is_deleted FROM links WHERE short_key = $1", shortURL).Scan(&fullURL, &isDeleted)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Error().Err(err).Str("short_key", shortURL).Msg("Failed to get link by short key")
		}
		return "", false
	}

	return fullURL, isDeleted
}

func (storage *Storage) GetShortKeyByURL(ctx context.Context, URL string) string {
	if storage.db == nil {
		return ""
	}

	normalizedURL := common.NormalizeURL(URL)
	var shortKey string
	err := storage.db.QueryRowContext(ctx, "SELECT short_key FROM links WHERE full_url = $1", normalizedURL).Scan(&shortKey)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
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

func (storage *Storage) Save(ctx context.Context, link *model.Link) error {
	if storage.db == nil {
		return nil
	}

	linkUUID := uuid.NewString()

	_, err := storage.db.ExecContext(ctx,
		"INSERT INTO links (uuid, short_key, full_url, user_id) VALUES ($1, $2, $3, $4)",
		linkUUID, link.ShortKey, link.FullURL, link.UserID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == pgerrcode.UniqueViolation {
				existingShortKey := storage.GetShortKeyByURL(ctx, link.FullURL)
				if existingShortKey != "" {
					return &model.URLConflictError{ExistingShortKey: existingShortKey}
				}
				existingURL, _ := storage.GetByShortKey(ctx, link.ShortKey)
				if existingURL == link.FullURL {
					return &model.URLConflictError{ExistingShortKey: link.ShortKey}
				}
				return model.ErrURLConflict
			}
		}
		log.Error().Err(err).Str("short_key", link.ShortKey).Str("url", link.FullURL).Msg("Failed to save link")
		return err
	}
	return nil
}

func (storage *Storage) BatchSave(ctx context.Context, links []*model.Link) error {
	if storage.db == nil {
		return nil
	}

	if len(links) == 0 {
		return nil
	}

	err := storage.batchSaveBulk(ctx, links)
	if err == nil {
		return nil
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == pgerrcode.UniqueViolation {
		return storage.batchSavePerRow(ctx, links)
	}
	return err
}

func (storage *Storage) batchSaveBulk(ctx context.Context, links []*model.Link) error {
	tx, err := storage.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	valueStrings := make([]string, 0, len(links))
	args := make([]interface{}, 0, len(links)*4)
	for i, link := range links {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d,$%d,$%d,$%d)", i*4+1, i*4+2, i*4+3, i*4+4))
		args = append(args, uuid.NewString(), link.ShortKey, link.FullURL, link.UserID)
	}
	query := "INSERT INTO links (uuid, short_key, full_url, user_id) VALUES " + strings.Join(valueStrings, ", ")
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (storage *Storage) batchSavePerRow(ctx context.Context, links []*model.Link) error {
	tx, err := storage.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		return err
	}

	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			log.Error().Err(rollbackErr).Msg("Failed to rollback transaction")
		}
	}()

	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO links (uuid, short_key, full_url, user_id) VALUES ($1, $2, $3, $4)")
	if err != nil {
		log.Error().Err(err).Msg("Failed to prepare batch insert statement")
		return err
	}
	defer func(stmt *sql.Stmt) {
		if closeErr := stmt.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("Failed to close prepared statement")
		}
	}(stmt)

	for _, link := range links {
		_, err := stmt.ExecContext(ctx, uuid.NewString(), link.ShortKey, link.FullURL, link.UserID)
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == pgerrcode.UniqueViolation {
				existingShortKey := storage.GetShortKeyByURL(ctx, link.FullURL)
				if existingShortKey != "" {
					return &model.URLConflictError{ExistingShortKey: existingShortKey}
				}
				existingURL, _ := storage.GetByShortKey(ctx, link.ShortKey)
				if existingURL == link.FullURL {
					return &model.URLConflictError{ExistingShortKey: link.ShortKey}
				}
				return model.ErrURLConflict
			}
			log.Error().Err(err).Str("short_key", link.ShortKey).Str("url", link.FullURL).Msg("Failed to save link in batch")
			return err
		}
	}

	return tx.Commit()
}

func (storage *Storage) GetByUserID(ctx context.Context, userID string) ([]*model.Link, error) {
	if storage.db == nil {
		return []*model.Link{}, nil
	}

	rows, err := storage.db.QueryContext(ctx,
		"SELECT short_key, full_url FROM links WHERE user_id = $1 AND is_deleted = FALSE ORDER BY uuid",
		userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to get links by user ID")
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close rows")
		}
	}(rows)

	links := make([]*model.Link, 0, 32)
	for rows.Next() {
		var link model.Link
		if err := rows.Scan(&link.ShortKey, &link.FullURL); err != nil {
			log.Error().Err(err).Str("user_id", userID).Msg("Failed to scan link")
			return nil, err
		}
		link.UserID = userID
		links = append(links, &link)
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Error iterating rows")
		return nil, err
	}

	return links, nil
}

// CountURLs возвращает количество не удалённых сокращённых ссылок в БД.
func (storage *Storage) CountURLs(ctx context.Context) (int, error) {
	if storage.db == nil {
		return 0, nil
	}

	var count int
	err := storage.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM links WHERE is_deleted = FALSE").Scan(&count)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count URLs")
		return 0, err
	}
	return count, nil
}

// CountUsers возвращает количество уникальных пользователей, сокращавших ссылки.
func (storage *Storage) CountUsers(ctx context.Context) (int, error) {
	if storage.db == nil {
		return 0, nil
	}

	var count int
	err := storage.db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT user_id) FROM links WHERE user_id <> ''").Scan(&count)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count users")
		return 0, err
	}
	return count, nil
}

func (storage *Storage) MarkDeleted(ctx context.Context, userID string, shortKeys []string) error {
	if storage.db == nil {
		return nil
	}

	_, err := storage.db.ExecContext(
		ctx,
		"UPDATE links SET is_deleted = TRUE WHERE user_id = $1 AND short_key = ANY($2)",
		userID,
		pq.Array(shortKeys),
	)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Int("count", len(shortKeys)).Msg("Failed to mark links deleted")
		return err
	}
	return nil
}
