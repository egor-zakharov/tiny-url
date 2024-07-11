package storage

import (
	"context"
	"encoding/json"
	"github.com/egor-zakharov/tiny-url/internal/app/models"
	"os"
	"sync"
)

type storage struct {
	urls models.MemData
	mu   sync.RWMutex
	file *os.File
}

// NewMemStorage - constructor mem storage
func NewMemStorage(file string) Storage {
	store := storage{}
	store.restore(file)
	return &store
}

// Delete - delete url
func (s *storage) Delete(shortURLs string, ID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	shorts, ok := s.urls.UserID[ID]
	if !ok {
		return nil
	}
	for k, v := range shorts.ShortURL {
		if k == shortURLs {
			shorts.ShortURL[k] = models.URL{
				OriginalURL: v.OriginalURL,
				IsDeleted:   "true",
			}
		}
	}
	return nil
}

// GetAll - get urls
func (s *storage) GetAll(_ context.Context, ID string) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.urls.UserID[ID]
	if !ok {
		return nil, ErrNotFound
	}
	res := make(map[string]string)
	for short, url := range v.ShortURL {
		res[short] = url.OriginalURL
	}
	return res, nil

}

// Add - add url
func (s *storage) Add(_ context.Context, shortURL string, url string, ID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	modURL := models.URL{
		OriginalURL: url,
		IsDeleted:   "false",
	}
	shorts, ok := s.urls.UserID[ID]
	if !ok {
		short := models.ShortURL{ShortURL: make(map[string]models.URL)}
		short.ShortURL[shortURL] = modURL
		s.urls.UserID[ID] = short
		return nil
	}

	_, ok = shorts.ShortURL[shortURL]
	if ok {
		return ErrConflict
	}

	shorts.ShortURL[shortURL] = modURL
	s.urls.UserID[ID] = shorts
	return nil
}

// AddBatch - add urls
func (s *storage) AddBatch(_ context.Context, URLs map[string]string, ID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	short := models.ShortURL{ShortURL: make(map[string]models.URL)}
	for shortURL, url := range URLs {
		modURL := models.URL{
			OriginalURL: url,
			IsDeleted:   "false",
		}
		_, ok := short.ShortURL[shortURL]
		if ok {
			return ErrConflict
		}
		short.ShortURL[shortURL] = modURL
	}

	s.urls.UserID[ID] = short
	return nil
}

// Get - get url
func (s *storage) Get(_ context.Context, shortURL string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.urls.UserID {
		url, ok := v.ShortURL[shortURL]
		if !ok {
			continue
		}
		return url.OriginalURL, nil
	}
	return "", ErrNotFound
}

func (s *storage) restore(file string) {
	s.file, _ = os.OpenFile(file, os.O_CREATE|os.O_RDWR, 0644)

	s.urls = models.MemData{UserID: make(map[string]models.ShortURL)}

	decoder := json.NewDecoder(s.file)
	rest := &models.Data{}
	for {
		if err := decoder.Decode(&rest); err != nil {
			break
		}
		url := models.URL{
			OriginalURL: rest.OriginalURL,
			IsDeleted:   rest.IsDeleted,
		}
		short := models.ShortURL{ShortURL: make(map[string]models.URL)}
		short.ShortURL[rest.ShortURL] = url
		s.urls.UserID[rest.UserID] = short
	}
}

// Backup - restore from file
func (s *storage) Backup() {
	writer := json.NewEncoder(s.file)
	for userID, shortURLs := range s.urls.UserID {
		for k, v := range shortURLs.ShortURL {
			shortenURL := models.Data{
				ShortURL:    k,
				OriginalURL: v.OriginalURL,
				UserID:      userID,
				IsDeleted:   v.IsDeleted,
			}
			_ = writer.Encode(&shortenURL)
		}
	}
	_ = s.file.Close()
}
