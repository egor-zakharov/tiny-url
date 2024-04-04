package service

import (
	"encoding/base64"
	"errors"
	"regexp"

	"github.com/egor-zakharov/tiny-url/internal/app/storage"
)

var (
	re            = regexp.MustCompile(`(https?:\/\/)?(www\.)?\S+\.\S+`)
	errInvalidURL = errors.New("url is invalid")
)

type Service struct {
	store *storage.Storage
}

func New() *Service {
	return &Service{storage.New()}
}

// Для тестов нужен mockgen для storage
func (s *Service) Add(url string) (shortURL string, err error) {
	shortURL = encodeURL(url)
	err = s.store.Add(shortURL, url)
	if err != nil {
		return shortURL, nil
	}
	return shortURL, err
}

// Для тестов нужен mockgen для storage
func (s *Service) Get(shortURL string) (string, error) {
	url, err := s.store.Get(shortURL)
	if err != nil {
		return url, err
	}
	return url, nil
}

// Будем явно валидировать в хендлере
func (s *Service) ValidateURL(url string) error {
	if ok := re.MatchString(url); !ok {
		return errInvalidURL
	}
	return nil
}

// Приватно кодируем URL
func encodeURL(url string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(url))
	// возвращаем 8 символов, отрезаем последние 2 ==
	return encoded[len(encoded)-10 : len(encoded)-2]
}
