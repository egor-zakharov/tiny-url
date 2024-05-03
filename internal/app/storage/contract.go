package storage

import "context"

//go:generate mockgen -source=contract.go -destination=contract_mock.go -package=storage

type Storage interface {
	Get(ctx context.Context, shortURL string) (string, error)
	Add(ctx context.Context, shortURL string, url string) error
}
