package storage

import (
	"context"
	"errors"
)

//go:generate mockgen -source=contract.go -destination=contract_mock.go -package=storage

var (
	ErrNotFound = errors.New("value not found")
	ErrConflict = errors.New("data conflict")
)

type Storage interface {
	Get(ctx context.Context, shortURL string) (string, error)
	GetAll(ctx context.Context, ID string) (map[string]string, error)
	Add(ctx context.Context, shortURL string, url string, ID string) error
	AddBatch(ctx context.Context, URLs map[string]string, ID string) error
	Backup()
}
