package storage

import (
	"context"
	"encoding/json"
	"github.com/egor-zakharov/tiny-url/internal/app/models"
	"os"
	"sync"
)

type storage struct {
	urls map[string]map[string]string
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

func (s *storage) Add(_ context.Context, shortURL string, url string, ID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.urls[ID]
	if !ok {
		s.urls[ID] = map[string]string{
			shortURL: url,
		}
		return nil
	}

	_, ok = v[shortURL]
	if ok {
		return ErrConflict
	}
	v[shortURL] = url
	return nil
}

func (s *storage) AddBatch(_ context.Context, URLs map[string]string, ID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.urls[ID]
	if !ok {
		s.urls[ID] = URLs
		return nil
	}

	for shortURL, url := range URLs {
		_, ok := v[shortURL]
		if ok {
			return ErrConflict
		}
		v[shortURL] = url
	}
	return nil
}

func (s *storage) Get(_ context.Context, shortURL string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.urls {
		url, ok := v[shortURL]
		if !ok {
			continue
		}
		return url, nil
	}
	return "", ErrNotFound
}

func (s *storage) restore(file string) {
	s.file, _ = os.OpenFile(file, os.O_CREATE|os.O_RDWR, 0644)

	s.urls = map[string]map[string]string{}

	decoder := json.NewDecoder(s.file)
	short := &models.Data{}
	for {
		if err := decoder.Decode(&short); err != nil {
			break
		}
		temp := make(map[string]string)
		temp[short.ShortURL] = short.OriginalURL
		s.urls[short.UserID] = temp
	}
}

func (s *storage) Backup() {
	writer := json.NewEncoder(s.file)
	for userID, v := range s.urls {
		for short, origin := range v {
			shortenURL := models.Data{
				ShortURL:    short,
				OriginalURL: origin,
				UserID:      userID,
			}
			_ = writer.Encode(&shortenURL)
			_ = s.file.Close()
		}

	}
}
