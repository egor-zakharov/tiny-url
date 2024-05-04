package storage

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/egor-zakharov/tiny-url/internal/app/models"
	"os"
	"sync"
)

type storage struct {
	urls map[string]string
	mu   sync.RWMutex
	file *os.File
}

func NewMemStorage(file string) Storage {
	store := storage{}
	store.restore(file)
	return &store
}

func (s *storage) Add(_ context.Context, shortURL string, url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var errs []error
	_, ok := s.urls[shortURL]
	if ok {
		errs = append(errs, ErrConflict)
	}
	s.urls[shortURL] = url
	err := s.addToFile(shortURL, url)
	if err != nil {
		errs = append(errs, err)
	}
	err = s.close()
	if err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (s *storage) AddBatch(_ context.Context, URLs map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var errs []error
	for shortURL, url := range URLs {
		_, ok := s.urls[shortURL]
		if ok {
			errs = append(errs, ErrConflict)
		}
		s.urls[shortURL] = url
		err := s.addToFile(shortURL, url)
		if err != nil {
			errs = append(errs, err)
		}
	}
	err := s.file.Close()
	if err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
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

func (s *storage) addToFile(shortURL string, url string) error {
	writer := json.NewEncoder(s.file)
	shortenURL := models.Data{
		ShortURL:    shortURL,
		OriginalURL: url,
	}
	err := writer.Encode(&shortenURL)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) close() error {
	return s.file.Close()
}
