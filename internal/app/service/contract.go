package service

import "context"

//go:generate mockgen -source=contract.go -destination=contract_mock.go -package=service

// Service - service interface
type Service interface {
	// Add - add url
	Add(ctx context.Context, url string, ID string) (string, error)
	// AddBatch - принимает map[correlation_id]original_url - возвращает map[correlation_id]short_url
	AddBatch(ctx context.Context, URLs map[string]string, ID string) (map[string]string, error)
	// Get - get url
	Get(ctx context.Context, shortURL string) (string, error)
	// GetAll - get all urls
	GetAll(ctx context.Context, ID string) (map[string]string, error)
	// ValidateURL - validate url by regexp
	ValidateURL(url string) error
	// Delete - delete url
	Delete(shortURLs string, ID string) error
}
