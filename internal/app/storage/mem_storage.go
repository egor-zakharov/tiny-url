package storage

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/egor-zakharov/tiny-url/internal/app/models"
	"os"
	"sync"
)

var (
	errAlreadyExist = errors.New("already exists")
	errNotFound     = errors.New("value not found")
	errEmptyData    = errors.New("urls not be empty")
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

func (s *storage) Add(ctx context.Context, shortURL string, url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.urls[shortURL]
	if ok {
		return errAlreadyExist
	}
	s.urls[shortURL] = url
	err := s.addToFile(shortURL, url)
	if err != nil {
		return err
	}
	err = s.close()
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) AddBatch(ctx context.Context, URLs map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(URLs) == 0 {
		return errEmptyData
	}
	for shortURL, url := range URLs {
		_, ok := s.urls[shortURL]
		if ok {
			continue
		}
		s.urls[shortURL] = url
		err := s.addToFile(shortURL, url)
		if err != nil {
			return err
		}
	}
	err := s.file.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) Get(ctx context.Context, shortURL string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.urls[shortURL]
	if !ok {
		return "", errNotFound
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
