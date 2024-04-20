package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Add(t *testing.T) {

	tests := []struct {
		name     string
		urls     map[string]string
		shortURL string
		longURL  string
		wantErr  bool
	}{
		{name: "Добавление в storage. Успех", urls: map[string]string{}, shortURL: "thisShort", longURL: "thisLong", wantErr: false},
		{name: "Добавление в storage. Ошибка", urls: map[string]string{"want": "err"}, shortURL: "want", longURL: "err", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Storage{urls: tt.urls}
			if !tt.wantErr {
				s = New("test")
				defer os.Remove("test")
			}
			err := s.Add(tt.shortURL, tt.longURL)
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
		{name: "Получение из storage. Успех", urls: map[string]string{"short": "long"}, shortURL: "short", longURL: "long", wantErr: false},
		{name: "Получение из storage. Ошибка", urls: map[string]string{}, shortURL: "want", longURL: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Storage{
				urls: tt.urls,
			}
			long, err := s.Get(tt.shortURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddNew() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.longURL, long)

		})
	}
}
