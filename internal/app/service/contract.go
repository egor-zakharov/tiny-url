package service

import "context"

//go:generate mockgen -source=contract.go -destination=contract_mock.go -package=service

type Service interface {
	Add(ctx context.Context, url string) (shortURL string, err error)
	Get(ctx context.Context, shortURL string) (string, error)
	ValidateURL(url string) error
}
