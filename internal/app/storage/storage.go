package storage

import (
	"errors"
	"sync"
)

var (
	errAlreadyExist = errors.New("already exists")
	errNotFound     = errors.New("value not found")
)

type Storage struct {
	urls map[string]string
	mu   sync.RWMutex
}

func New() *Storage {
	return &Storage{
		urls: make(map[string]string),
	}
}

func (s *Storage) Add(shortURL, url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.urls[shortURL]
	if ok {
		return errAlreadyExist
	}
	s.urls[shortURL] = url
	return nil
}

func (s *Storage) Get(shortURL string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.urls[shortURL]
	if !ok {
		return "", errNotFound
	}
	return url, nil
}
