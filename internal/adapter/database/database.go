package database

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(dsn string) (*Storage, error) {
	if dsn == "" {
		log.Warn().Msg("Database DSN is empty, database storage will not be initialized")
		return &Storage{}, nil
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		closeErr := db.Close()
		if closeErr != nil {
			log.Error().Err(closeErr).Msg("Error closing database connection")
		}
		return nil, err
	}

	log.Info().Msg("Database connection established successfully")
	return &Storage{db}, nil
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
