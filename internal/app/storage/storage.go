package storage

import (
	"encoding/json"
	"errors"
	"github.com/egor-zakharov/tiny-url/internal/app/models"
	"os"
	"sync"
)

var (
	errAlreadyExist = errors.New("already exists")
	errNotFound     = errors.New("value not found")
)

type Storage struct {
	urls map[string]string
	mu   sync.RWMutex
	file *os.File
}

func New(file string) *Storage {
	store := Storage{}
	store.Restore(file)
	return &store
}

func (s *Storage) Add(shortURL string, url string) error {
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

func (s *Storage) Restore(file string) {
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

func (s *Storage) addToFile(shortURL string, url string) error {
	writer := json.NewEncoder(s.file)
	shortenURL := models.Data{
		ShortURL:    shortURL,
		OriginalURL: url,
	}
	err := writer.Encode(&shortenURL)
	if err != nil {
		return err
	}
	return s.file.Close()
}
