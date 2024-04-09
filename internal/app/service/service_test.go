package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EncodeURL(t *testing.T) {
	tests := []struct {
		name     string
		longURL  string
		shortURL string
	}{
		{name: "Кодирования строки. Успех", longURL: "https://practicum.yandex.ru/", shortURL: "V4LnJ1Lw"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := encodeURL(tt.longURL)
			assert.Equal(t, tt.shortURL, encoded)

		})
	}
}

func Test_ValidateURL(t *testing.T) {
	s := &Service{}
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "Валидация http. Успех", url: "http://practicum.yandex.ru/", wantErr: false},
		{name: "Валидация https. Успех", url: "https://practicum.yandex.ru/", wantErr: false},
		{name: "Ввалидация. Ошибка", url: "notUrl", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddNew() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}
