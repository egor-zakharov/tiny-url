package storage

import (
	"context"
	"database/sql"
	"time"
)

const timeOut = 300 * time.Millisecond

type dbStorage struct {
	db *sql.DB
}

func NewDBStorage(ctx context.Context, db *sql.DB) Storage {
	dbs := &dbStorage{db: db}
	_ = dbs.init(ctx)
	return dbs
}

func (db *dbStorage) Add(ctx context.Context, shortURL string, url string) error {
	ctx, cancel := context.WithTimeout(ctx, timeOut)
	defer cancel()

	_, err := db.db.ExecContext(ctx, `INSERT INTO urls(short_url, original_url) VALUES ($1, $2) ON CONFLICT DO NOTHING`, shortURL, url)
	return err
}

func (db *dbStorage) Get(ctx context.Context, shortURL string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeOut)
	defer cancel()

	row := db.db.QueryRowContext(ctx, `SELECT original_url FROM urls WHERE short_url=$1`, shortURL)
	url := ""
	err := row.Scan(&url)
	return url, err
}

func (db *dbStorage) init(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, timeOut)
	defer cancel()

	_, err := db.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS urls (
		    short_url VARCHAR NOT NULL UNIQUE,
		    original_url VARCHAR NOT NULL UNIQUE
		    )
		`)
	return err
}
