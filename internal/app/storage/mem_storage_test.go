package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

const userID = "someString"

func Test_Add(t *testing.T) {
	tests := []struct {
		name     string
		urls     map[string]map[string]string
		shortURL string
		longURL  string
		userID   string
		wantErr  bool
	}{
		{name: "Добавление в dbStorage. Успех", urls: map[string]map[string]string{}, shortURL: "thisShort", longURL: "thisLong", userID: userID, wantErr: false},
		{name: "Добавление в dbStorage. Ошибка", urls: map[string]map[string]string{userID: {"want": "err"}}, shortURL: "want", longURL: "err", userID: userID, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := storage{
				urls: tt.urls,
			}
			err := s.Add(context.Background(), tt.shortURL, tt.longURL, userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddNew() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

func Test_AddBatch(t *testing.T) {

	tests := []struct {
		name     string
		prepURLs map[string]map[string]string
		inURLs   map[string]string
		userID   string
		wantErr  bool
	}{
		{name: "Добавление batch в dbStorage. Успех", prepURLs: map[string]map[string]string{}, inURLs: map[string]string{"thisShort": "thisLong", "thisShort2": "thisLong2"}, userID: userID, wantErr: false},
		{name: "Добавление batch в dbStorage. Ошибка", prepURLs: map[string]map[string]string{userID: {"want": "err"}}, inURLs: map[string]string{"want": "err"}, userID: userID, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := storage{
				urls: tt.prepURLs,
			}
			err := s.AddBatch(context.Background(), tt.inURLs, userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddNew() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

func Test_Get(t *testing.T) {
	tests := []struct {
		name     string
		urls     map[string]string
		shortURL string
		longURL  string
		wantErr  bool
	}{
		{name: "Получение из dbStorage. Успех", urls: map[string]string{"short": "long"}, shortURL: "short", longURL: "long", wantErr: false},
		{name: "Получение из dbStorage. Ошибка", urls: map[string]string{}, shortURL: "want", longURL: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemStorage("")
			if !tt.wantErr {
				_ = s.Add(context.Background(), tt.shortURL, tt.longURL, "")
			}
			long, err := s.Get(context.Background(), tt.shortURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddNew() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.longURL, long)

		})
	}
}

func Test_GetAll(t *testing.T) {
	tests := []struct {
		name     string
		urls     map[string]string
		shortURL string
		longURL  string
		userID   string
		wantErr  bool
	}{
		{name: "Получение из dbStorage. Успех", urls: map[string]string{"short": "long"}, shortURL: "short", longURL: "long", userID: userID, wantErr: false},
		{name: "Получение из dbStorage. Ошибка", urls: nil, shortURL: "want", longURL: "", userID: userID, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemStorage("")
			if !tt.wantErr {
				_ = s.Add(context.Background(), tt.shortURL, tt.longURL, userID)
			}
			urls, err := s.GetAll(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddNew() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.urls, urls)

		})
	}
}
