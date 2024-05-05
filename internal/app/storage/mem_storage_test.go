package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_Add(t *testing.T) {

	tests := []struct {
		name     string
		urls     map[string]string
		shortURL string
		longURL  string
		wantErr  bool
	}{
		{name: "Добавление в dbStorage. Успех", urls: map[string]string{}, shortURL: "thisShort", longURL: "thisLong", wantErr: false},
		{name: "Добавление в dbStorage. Ошибка", urls: map[string]string{"want": "err"}, shortURL: "want", longURL: "err", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemStorage("")
			if !tt.wantErr {
				s = NewMemStorage("test")
				defer os.Remove("test")
			}
			err := s.Add(context.Background(), tt.shortURL, tt.longURL)
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
				_ = s.Add(context.Background(), tt.shortURL, tt.longURL)
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
