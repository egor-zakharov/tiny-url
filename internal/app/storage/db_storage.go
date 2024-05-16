package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

const timeOut = 500 * time.Millisecond

type dbStorage struct {
	db *sql.DB
}

func NewDBStorage(ctx context.Context, db *sql.DB) Storage {
	dbs := &dbStorage{db: db}
	_ = dbs.init(ctx)
	return dbs
}

func (db *dbStorage) Add(ctx context.Context, shortURL string, url string, ID string) error {
	ctx, cancel := context.WithTimeout(ctx, timeOut)
	defer cancel()

	_, err := db.db.ExecContext(ctx, `INSERT INTO urls(short_url, original_url, user_id) VALUES ($1, $2, $3)`, shortURL, url, ID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		err = ErrConflict
	}
	return err
}
func (db *dbStorage) AddBatch(ctx context.Context, URLs map[string]string, ID string) error {
	// начинаем транзакцию
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	for k, v := range URLs {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO urls(short_url, original_url, user_id) VALUES($1, $2, $3) ON CONFLICT DO NOTHING`, k, v, ID)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (db *dbStorage) Get(ctx context.Context, shortURL string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeOut)
	defer cancel()

	row := db.db.QueryRowContext(ctx, `SELECT original_url FROM urls WHERE short_url=$1`, shortURL)
	url := ""
	err := row.Scan(&url)
	return url, err
}

func (db *dbStorage) GetAll(ctx context.Context, ID string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeOut)
	defer cancel()

	urls := make(map[string]string, 0)
	rows, err := db.db.QueryContext(ctx, "SELECT short_url, original_url FROM urls WHERE user_id=$1;", ID)
	if err != nil {
		return nil, err
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		shortURL := ""
		originalURL := ""
		err = rows.Scan(&shortURL, &originalURL)
		if err != nil {
			return nil, err
		}
		urls[shortURL] = originalURL
	}
	return urls, err
}

func (db *dbStorage) init(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, timeOut)
	defer cancel()

	_, err := db.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS urls (
		    short_url VARCHAR NOT NULL UNIQUE,
		    original_url VARCHAR NOT NULL UNIQUE,
		    user_id VARCHAR
		    )
		`)
	return err
}

func (db *dbStorage) Backup() {
}
