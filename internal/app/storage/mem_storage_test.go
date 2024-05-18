package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

//func Test_Add(t *testing.T) {
//
//	tests := []struct {
//		name     string
//		urls     map[string]string
//		shortURL string
//		longURL  string
//		wantErr  bool
//	}{
//		{name: "Добавление в dbStorage. Успех", urls: map[string]string{}, shortURL: "thisShort", longURL: "thisLong", wantErr: false},
//		{name: "Добавление в dbStorage. Ошибка", urls: map[string]string{"want": "err"}, shortURL: "want", longURL: "err", wantErr: true},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := storage{
//				urls: tt.urls,
//			}
//			err := s.Add(context.Background(), tt.shortURL, tt.longURL, "")
//			if (err != nil) != tt.wantErr {
//				t.Errorf("AddNew() error = %v, wantErr %v", err, tt.wantErr)
//			}
//
//		})
//	}
//}
//
//func Test_AddBatch(t *testing.T) {
//
//	tests := []struct {
//		name     string
//		prepURLs map[string]string
//		inURLs   map[string]string
//		wantErr  bool
//	}{
//		{name: "Добавление batch в dbStorage. Успех", prepURLs: map[string]string{}, inURLs: map[string]string{"thisShort": "thisLong", "thisShort2": "thisLong2"}, wantErr: false},
//		{name: "Добавление batch в dbStorage. Ошибка", prepURLs: map[string]string{"want": "err"}, inURLs: map[string]string{"want": "err"}, wantErr: true},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := storage{
//				urls: tt.prepURLs,
//			}
//			err := s.AddBatch(context.Background(), tt.inURLs, "")
//			if (err != nil) != tt.wantErr {
//				t.Errorf("AddNew() error = %v, wantErr %v", err, tt.wantErr)
//			}
//
//		})
//	}
//}

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
