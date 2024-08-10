package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/egor-zakharov/tiny-url/internal/app/storage"
	"github.com/golang/mock/gomock"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_encodeURL(t *testing.T) {
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

func TestService_ValidateURL(t *testing.T) {
	s := &service{}
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

func Test_service_Delete(t *testing.T) {
	shortURL := "short"
	id := "1"
	type fields struct {
		storage func(ctrl *gomock.Controller) storage.Storage
	}
	type args struct {
		shortURLs string
		ID        string
	}
	tests := []struct {
		fields  fields
		wantErr assert.ErrorAssertionFunc
		args    args
		name    string
	}{
		{
			name: "Delete success",
			fields: fields{
				storage: func(ctrl *gomock.Controller) storage.Storage {
					mock := storage.NewMockStorage(ctrl)
					mock.EXPECT().Delete(shortURL, id).Return(nil)
					return mock
				},
			},
			args: args{
				shortURLs: shortURL,
				ID:        id,
			},
			wantErr: assert.NoError,
		},
		{
			name: "Delete error",
			fields: fields{
				storage: func(ctrl *gomock.Controller) storage.Storage {
					mock := storage.NewMockStorage(ctrl)
					mock.EXPECT().Delete(shortURL, id).Return(errors.New("error"))
					return mock
				},
			},
			args: args{
				shortURLs: shortURL,
				ID:        id,
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &service{
				storage: tt.fields.storage(ctrl),
			}
			tt.wantErr(t, s.Delete(tt.args.shortURLs, tt.args.ID), fmt.Sprintf("Delete(%v, %v)", tt.args.shortURLs, tt.args.ID))
		})
	}
}

func Test_service_GetAll(t *testing.T) {
	ctx := context.Background()
	id := "1"
	type fields struct {
		storage func(ctrl *gomock.Controller) storage.Storage
	}
	type args struct {
		ctx context.Context
		ID  string
	}
	tests := []struct {
		fields  fields
		want    map[string]string
		wantErr assert.ErrorAssertionFunc
		args    args
		name    string
	}{
		{
			name: "GetAll success",
			fields: fields{
				storage: func(ctrl *gomock.Controller) storage.Storage {
					mock := storage.NewMockStorage(ctrl)
					mock.EXPECT().GetAll(ctx, id).Return(map[string]string{"1": "1"}, nil)
					return mock
				},
			},
			args: args{
				ctx: ctx,
				ID:  id,
			},
			want:    map[string]string{"1": "1"},
			wantErr: assert.NoError,
		},
		{
			name: "GetAll error",
			fields: fields{
				storage: func(ctrl *gomock.Controller) storage.Storage {
					mock := storage.NewMockStorage(ctrl)
					mock.EXPECT().GetAll(ctx, id).Return(nil, errors.New("error"))
					return mock
				},
			},
			args: args{
				ctx: ctx,
				ID:  id,
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &service{
				storage: tt.fields.storage(ctrl),
			}
			got, err := s.GetAll(tt.args.ctx, tt.args.ID)
			if !tt.wantErr(t, err, fmt.Sprintf("GetAll(%v, %v)", tt.args.ctx, tt.args.ID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetAll(%v, %v)", tt.args.ctx, tt.args.ID)
		})
	}
}

func Test_service_Add(t *testing.T) {
	ctx := context.Background()
	shortURL := "xvbmdsb2"
	longURL := "longlonglong"
	id := "1"
	type fields struct {
		storage func(ctrl *gomock.Controller) storage.Storage
	}
	type args struct {
		ctx context.Context
		url string
		ID  string
	}
	tests := []struct {
		fields  fields
		wantErr assert.ErrorAssertionFunc
		args    args
		name    string
		want    string
	}{
		{
			name: "Add success",
			fields: fields{
				storage: func(ctrl *gomock.Controller) storage.Storage {
					mock := storage.NewMockStorage(ctrl)
					mock.EXPECT().Add(ctx, shortURL, longURL, id).Return(nil)
					return mock
				},
			},
			args: args{
				ctx: ctx,
				url: longURL,
				ID:  id,
			},
			want:    shortURL,
			wantErr: assert.NoError,
		},
		{
			name: "Add error",
			fields: fields{
				storage: func(ctrl *gomock.Controller) storage.Storage {
					mock := storage.NewMockStorage(ctrl)
					mock.EXPECT().Add(ctx, shortURL, longURL, id).Return(errors.New("error"))
					return mock
				},
			},
			args: args{
				ctx: ctx,
				url: longURL,
				ID:  id,
			},
			want:    shortURL,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &service{
				storage: tt.fields.storage(ctrl),
			}
			got, err := s.Add(tt.args.ctx, tt.args.url, tt.args.ID)
			if !tt.wantErr(t, err, fmt.Sprintf("Add(%v, %v, %v)", tt.args.ctx, tt.args.url, tt.args.ID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Add(%v, %v, %v)", tt.args.ctx, tt.args.url, tt.args.ID)
		})
	}
}

func Test_service_Get(t *testing.T) {
	ctx := context.Background()
	shortURL := "short"
	longURL := "long"
	type fields struct {
		storage func(ctrl *gomock.Controller) storage.Storage
	}
	type args struct {
		ctx      context.Context
		shortURL string
	}
	tests := []struct {
		fields  fields
		wantErr assert.ErrorAssertionFunc
		args    args
		name    string
		want    string
	}{
		{
			name: "Get success",
			fields: fields{
				storage: func(ctrl *gomock.Controller) storage.Storage {
					mock := storage.NewMockStorage(ctrl)
					mock.EXPECT().Get(ctx, shortURL).Return(longURL, nil)
					return mock
				},
			},
			args: args{
				ctx:      ctx,
				shortURL: shortURL,
			},
			want:    longURL,
			wantErr: assert.NoError,
		},
		{
			name: "Get error",
			fields: fields{
				storage: func(ctrl *gomock.Controller) storage.Storage {
					mock := storage.NewMockStorage(ctrl)
					mock.EXPECT().Get(ctx, shortURL).Return("", errors.New("error"))
					return mock
				},
			},
			args: args{
				ctx:      ctx,
				shortURL: shortURL,
			},
			want:    "",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &service{
				storage: tt.fields.storage(ctrl),
			}
			got, err := s.Get(tt.args.ctx, tt.args.shortURL)
			if !tt.wantErr(t, err, fmt.Sprintf("Get(%v, %v)", tt.args.ctx, tt.args.shortURL)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Get(%v, %v)", tt.args.ctx, tt.args.shortURL)
		})
	}
}

func Test_service_AddBatch(t *testing.T) {
	ctx := context.Background()
	shortURL := "xvbmdsb2"
	longURL := "longlonglong"
	URLs := map[string]string{shortURL: longURL}
	id := "1"
	type fields struct {
		storage func(ctrl *gomock.Controller) storage.Storage
	}
	type args struct {
		ctx  context.Context
		URLs map[string]string
		ID   string
	}
	tests := []struct {
		fields  fields
		want    map[string]string
		wantErr assert.ErrorAssertionFunc
		args    args
		name    string
	}{
		{
			name: "Add batch success",
			fields: fields{
				storage: func(ctrl *gomock.Controller) storage.Storage {
					mock := storage.NewMockStorage(ctrl)
					mock.EXPECT().AddBatch(ctx, URLs, id).Return(nil)
					return mock
				},
			},
			args: args{
				ctx:  ctx,
				URLs: map[string]string{shortURL: longURL},
				ID:   id,
			},
			want:    map[string]string{shortURL: shortURL},
			wantErr: assert.NoError,
		},
		{
			name: "Add batch error",
			fields: fields{
				storage: func(ctrl *gomock.Controller) storage.Storage {
					mock := storage.NewMockStorage(ctrl)
					mock.EXPECT().AddBatch(ctx, URLs, id).Return(errors.New("error"))
					return mock
				},
			},
			args: args{
				ctx:  ctx,
				URLs: map[string]string{shortURL: longURL},
				ID:   id,
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &service{
				storage: tt.fields.storage(ctrl),
			}
			got, err := s.AddBatch(tt.args.ctx, tt.args.URLs, tt.args.ID)
			if !tt.wantErr(t, err, fmt.Sprintf("AddBatch(%v, %v, %v)", tt.args.ctx, tt.args.URLs, tt.args.ID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "AddBatch(%v, %v, %v)", tt.args.ctx, tt.args.URLs, tt.args.ID)
		})
	}
}
