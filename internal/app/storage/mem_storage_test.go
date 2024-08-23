package storage

import (
	"context"
	"fmt"
	"github.com/egor-zakharov/tiny-url/internal/app/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

const userID = "someString"

func Test_Add(t *testing.T) {
	tests := []struct {
		name     string
		urls     models.MemData
		shortURL string
		longURL  string
		userID   string
		wantErr  bool
	}{
		{name: "Добавление в memStorage. Успех", urls: models.MemData{UserID: make(map[string]models.ShortURL)}, shortURL: "thisShort", longURL: "thisLong", userID: userID, wantErr: false},
		{name: "Добавление в memStorage. Ошибка", urls: models.MemData{
			UserID: map[string]models.ShortURL{userID: {ShortURL: map[string]models.URL{"want": {
				OriginalURL: "err",
				IsDeleted:   "false",
			},
			}}},
		}, shortURL: "want", longURL: "err", userID: userID, wantErr: true},
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
		prepURLs models.MemData
		inURLs   map[string]string
		userID   string
		wantErr  bool
	}{
		{name: "Добавление batch в memStorage. Успех", prepURLs: models.MemData{UserID: make(map[string]models.ShortURL)}, inURLs: map[string]string{"thisShort": "thisLong", "thisShort2": "thisLong2"}, userID: userID, wantErr: false},
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
		{name: "Получение из memStorage. Успех", urls: map[string]string{"short": "long"}, shortURL: "short", longURL: "long", wantErr: false},
		{name: "Получение из memStorage. Ошибка", urls: map[string]string{}, shortURL: "want", longURL: "", wantErr: true},
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
		{name: "Получение из memStorage. Успех", urls: map[string]string{"short": "long"}, shortURL: "short", longURL: "long", userID: userID, wantErr: false},
		{name: "Получение из memStorage. Ошибка", urls: nil, shortURL: "want", longURL: "", userID: userID, wantErr: true},
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

func Test_storage_Delete(t *testing.T) {
	tests := []struct {
		name     string
		urls     map[string]string
		shortURL string
		longURL  string
		userID   string
		wantErr  bool
	}{
		{name: "Удаление из memStorage. Успех", urls: map[string]string{"short": "long"}, shortURL: "short", longURL: "long", userID: userID, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemStorage("")
			if !tt.wantErr {
				_ = s.Add(context.Background(), tt.shortURL, tt.longURL, userID)
			}
			err := s.Delete(tt.shortURL, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddNew() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_storage_GetStats(t *testing.T) {
	tests := []struct {
		name     string
		urls     models.Stats
		shortURL string
		longURL  string
		userID   string
		wantErr  bool
	}{
		{name: "GetStats из memStorage. Успех", urls: models.Stats{
			Users: 1,
			Urls:  1,
		}, shortURL: "short", longURL: "long", userID: userID, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemStorage("")
			if !tt.wantErr {
				_ = s.Add(context.Background(), tt.shortURL, tt.longURL, userID)
			}
			urls, err := s.GetStats(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.urls, urls)

		})
	}
}

func BenchmarkStorage_AddBatch(b *testing.B) {
	s := storage{
		urls: models.MemData{UserID: make(map[string]models.ShortURL)},
	}
	for i := 0; i < b.N; i++ {
		short := fmt.Sprintf("%d%s", i, "short")
		long := fmt.Sprintf("%d%s", i, "long")
		shortNext := fmt.Sprintf("%d%s", i, "shortNext")
		longNext := fmt.Sprintf("%d%s", i, "longnext")
		_ = s.AddBatch(context.Background(), map[string]string{short: long, shortNext: longNext}, userID)
	}
}

func BenchmarkStorage_Add(b *testing.B) {
	s := storage{
		urls: models.MemData{UserID: make(map[string]models.ShortURL)},
	}
	for i := 0; i < b.N; i++ {
		short := fmt.Sprintf("%d%s", i, "short")
		long := fmt.Sprintf("%d%s", i, "long")
		_ = s.AddBatch(context.Background(), map[string]string{short: long}, userID)
	}
}
