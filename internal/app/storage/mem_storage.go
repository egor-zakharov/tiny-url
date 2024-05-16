package storage

import (
	"context"
	"encoding/json"
	"github.com/egor-zakharov/tiny-url/internal/app/models"
	"os"
	"sync"
)

type storage struct {
	urls map[string]string
	mu   sync.RWMutex
	file *os.File
}

func (s *storage) GetAll(ctx context.Context, ID string) (map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func NewMemStorage(file string) Storage {
	store := storage{}
	store.restore(file)
	return &store
}

func (s *storage) Add(_ context.Context, shortURL string, url string, _ string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.urls[shortURL]
	if ok {
		return ErrConflict
	}
	s.urls[shortURL] = url
	return nil
}

func (s *storage) AddBatch(_ context.Context, URLs map[string]string, _ string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for shortURL, url := range URLs {
		_, ok := s.urls[shortURL]
		if ok {
			return ErrConflict
		}
		s.urls[shortURL] = url
	}
	return nil
}

func (s *storage) Get(_ context.Context, shortURL string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.urls[shortURL]
	if !ok {
		return "", ErrNotFound
	}
	return url, nil
}

func (s *storage) restore(file string) {
	s.file, _ = os.OpenFile(file, os.O_CREATE|os.O_RDWR, 0644)

	s.urls = map[string]string{}

	decoder := json.NewDecoder(s.file)
	short := &models.Data{}
	for {
		if err := decoder.Decode(&short); err != nil {
			break
		}
		s.urls[short.ShortURL] = short.OriginalURL
	}
}

func (s *storage) Backup() {
	writer := json.NewEncoder(s.file)
	for k, v := range s.urls {
		shortenURL := models.Data{
			ShortURL:    k,
			OriginalURL: v,
		}
		_ = writer.Encode(&shortenURL)
		_ = s.file.Close()
	}
}
