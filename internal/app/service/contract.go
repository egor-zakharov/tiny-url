package service

import "context"

//go:generate mockgen -source=contract.go -destination=contract_mock.go -package=service

type Service interface {
	Add(ctx context.Context, url string, ID string) (string, error)
	// AddBatch принимает map[correlation_id]original_url - возвращает map[correlation_id]short_url
	AddBatch(ctx context.Context, URLs map[string]string, ID string) (map[string]string, error)
	Get(ctx context.Context, shortURL string, ID string) (string, error)
	GetAll(ctx context.Context, ID string) (map[string]string, error)
	ValidateURL(url string) error
}
