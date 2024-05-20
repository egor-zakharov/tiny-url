package service

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/egor-zakharov/tiny-url/internal/app/storage"
	"regexp"
)

var (
	re            = regexp.MustCompile(`(https?:\/\/)?(www\.)?\S+\.\S+`)
	errInvalidURL = errors.New("url is invalid")
)

type service struct {
	storage storage.Storage
}

func (s *service) Delete(shortURLs string, ID string) error {
	err := s.storage.Delete(shortURLs, ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) GetAll(ctx context.Context, ID string) (map[string]string, error) {
	urls, err := s.storage.GetAll(ctx, ID)
	if err != nil {
		return urls, err
	}
	return urls, nil
}

func NewService(storage storage.Storage) Service {
	return &service{storage: storage}
}

// Для тестов нужен mockgen для storage
func (s *service) Add(ctx context.Context, url string, ID string) (string, error) {
	shortURL := encodeURL(url)
	err := s.storage.Add(ctx, shortURL, url, ID)
	return shortURL, err
}

// AddBatch принимает map[correlation_id]original_url - возвращает map[correlation_id]short_url
func (s *service) AddBatch(ctx context.Context, URLs map[string]string, ID string) (map[string]string, error) {
	inStore := make(map[string]string, len(URLs))
	res := make(map[string]string, len(URLs))
	for corID, url := range URLs {
		shortURL := encodeURL(url)
		inStore[shortURL] = url
		res[corID] = shortURL
	}
	err := s.storage.AddBatch(ctx, inStore, ID)
	if err != nil {
		return nil, err
	}
	return res, err
}

// Для тестов нужен mockgen для storage
func (s *service) Get(ctx context.Context, shortURL string) (string, error) {
	url, err := s.storage.Get(ctx, shortURL)
	if err != nil {
		return url, err
	}
	return url, nil
}

// Будем явно валидировать в хендлере
func (s *service) ValidateURL(url string) error {
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
