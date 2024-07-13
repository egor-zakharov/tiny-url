package storage

import (
	"context"
	"errors"
)

//go:generate mockgen -source=contract.go -destination=contract_mock.go -package=storage

// Errors from storage
var (
	ErrNotFound   = errors.New("value not found")
	ErrConflict   = errors.New("data conflict")
	ErrDeletedURL = errors.New("urls is deleted")
)

// Storage interface
type Storage interface {
	// Get - get url
	Get(ctx context.Context, shortURL string) (string, error)
	// GetAll - get urls
	GetAll(ctx context.Context, ID string) (map[string]string, error)
	// Add - add url
	Add(ctx context.Context, shortURL string, url string, ID string) error
	// AddBatch - add urls
	AddBatch(ctx context.Context, URLs map[string]string, ID string) error
	// Backup - only for mem_storage
	Backup()
	// Delete - delete url
	Delete(shortURLs string, ID string) error
}
