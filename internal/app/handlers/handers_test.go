package handlers

import (
	"context"
	"encoding/json"
	"github.com/egor-zakharov/tiny-url/internal/app/config"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/egor-zakharov/tiny-url/internal/app/logger"
	"github.com/egor-zakharov/tiny-url/internal/app/models"
	"github.com/egor-zakharov/tiny-url/internal/app/service"
	"github.com/egor-zakharov/tiny-url/internal/app/storage"
	"github.com/egor-zakharov/tiny-url/internal/app/zipper"
	"github.com/stretchr/testify/assert"
)

const baseURL = "http://localhost:8080"

func testRequestNoRedirect(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	// http client that doesn't redirect
	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer resp.Body.Close()

	return resp, string(respBody)
}

func Test_Post(t *testing.T) {
	log := logger.NewLogger()
	store := storage.NewMemStorage("")
	srv := service.NewService(store)
	zip := zipper.NewZipper()
	conf := config.NewConfig()
	conf.FlagShortAddr = baseURL
	tests := []struct {
		name                 string
		method               string
		requestBody          string
		expectedCode         int
		expectedResponseBody string
	}{
		{name: "Проверка запроса без тела", method: http.MethodPost, requestBody: "", expectedCode: http.StatusBadRequest, expectedResponseBody: ""},
		{name: "Проверка запроса с телом", method: http.MethodPost, requestBody: "https://practicum.yandex.ru/", expectedCode: http.StatusCreated, expectedResponseBody: "http://localhost:8080/V4LnJ1Lw"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stringReader := strings.NewReader(tt.requestBody)
			ts := httptest.NewServer(NewHandlers(srv, conf, log, zip).ChiRouter())
			defer ts.Close()
			resp, body := testRequestNoRedirect(t, ts, tt.method, "/", stringReader)
			resp.Body.Close()
			// проверка статус кода
			assert.Equal(t, tt.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
			// проверим корректность полученного тела ответа, если мы его ожидаем
			if tt.expectedResponseBody != "" {
				// проверка тела ответа
				assert.Equal(t, tt.expectedResponseBody, body, "Тело ответа не совпадает с ожидаемым")
			}
		})
	}
}

func Test_PostShorten(t *testing.T) {
	tempModel := models.Response{}
	log := logger.NewLogger()
	store := storage.NewMemStorage("")
	srv := service.NewService(store)
	zip := zipper.NewZipper()
	conf := config.NewConfig()
	conf.FlagShortAddr = baseURL
	tests := []struct {
		name                 string
		method               string
		requestBody          models.Request
		expectedCode         int
		expectedResponseBody models.Response
	}{
		{name: "Проверка запроса без тела", method: http.MethodPost, requestBody: models.Request{}, expectedCode: http.StatusBadRequest, expectedResponseBody: models.Response{}},
		{name: "Проверка запроса с телом", method: http.MethodPost, requestBody: models.Request{URL: "https://practicum.yandex.ru/"}, expectedCode: http.StatusCreated, expectedResponseBody: models.Response{Result: "http://localhost:8080/V4LnJ1Lw"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, _ := json.Marshal(tt.requestBody)
			stringReader := strings.NewReader(string(out))
			ts := httptest.NewServer(NewHandlers(srv, conf, log, zip).ChiRouter())
			defer ts.Close()
			resp, body := testRequestNoRedirect(t, ts, tt.method, "/api/shorten", stringReader)
			resp.Body.Close()
			// проверка статус кода
			assert.Equal(t, tt.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
			// проверим корректность полученного тела ответа, если мы его ожидаем
			if tt.expectedResponseBody.Result != "" {

				json.Unmarshal([]byte(body), &tempModel)
				// проверка тела ответа
				assert.Equal(t, tt.expectedResponseBody, tempModel, "Тело ответа не совпадает с ожидаемым")
			}
		})
	}
}

func Test_PostShortenBatch(t *testing.T) {
	var tempModel []models.ShortenBatchResponse
	log := logger.NewLogger()
	store := storage.NewMemStorage("")
	srv := service.NewService(store)
	zip := zipper.NewZipper()
	conf := config.NewConfig()
	conf.FlagShortAddr = baseURL
	tests := []struct {
		name                 string
		method               string
		requestBody          []models.ShortenBatchRequest
		expectedCode         int
		expectedResponseBody []models.ShortenBatchResponse
	}{

		{name: "Проверка запроса с телом", method: http.MethodPost, requestBody: []models.ShortenBatchRequest{
			{
				CorrelationID: "123",
				URL:           "https://yandex.ru",
			}},
			expectedCode: http.StatusCreated, expectedResponseBody: []models.ShortenBatchResponse{
				{
					CorrelationID: "123",
					ShortURL:      "http://localhost:8080/5kZXgucn",
				},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, _ := json.Marshal(tt.requestBody)
			stringReader := strings.NewReader(string(out))
			ts := httptest.NewServer(NewHandlers(srv, conf, log, zip).ChiRouter())
			defer ts.Close()
			resp, body := testRequestNoRedirect(t, ts, tt.method, "/api/shorten/batch", stringReader)
			resp.Body.Close()
			// проверка статус кода
			assert.Equal(t, tt.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")

			json.Unmarshal([]byte(body), &tempModel)
			// проверка тела ответа
			assert.Equal(t, tt.expectedResponseBody, tempModel, "Тело ответа не совпадает с ожидаемым")

		})
	}
}

func Test_get(t *testing.T) {
	log := logger.NewLogger()
	store := storage.NewMemStorage("")
	srv := service.NewService(store)
	zip := zipper.NewZipper()
	conf := config.NewConfig()
	conf.FlagShortAddr = baseURL
	tests := []struct {
		name             string
		method           string
		path             string
		expectedCode     int
		expectedLocation string
	}{
		{name: "Проверка отсутствующего URL", method: http.MethodGet, path: "/urlNotFound", expectedCode: http.StatusBadRequest, expectedLocation: ""},
		{name: "Проверка Location", method: http.MethodGet, path: "/V4LnJ1Lw", expectedCode: http.StatusTemporaryRedirect, expectedLocation: "https://practicum.yandex.ru/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandlers(srv, conf, log, zip)
			if tt.expectedLocation != "" {
				h.service.Add(context.Background(), tt.expectedLocation)
			}
			ts := httptest.NewServer(h.ChiRouter())
			defer ts.Close()
			resp, _ := testRequestNoRedirect(t, ts, tt.method, tt.path, nil)
			resp.Body.Close()
			// проверка статус кода
			assert.Equal(t, tt.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
			// проверим корректность полученного тела ответа, если мы его ожидаем
			if tt.expectedLocation != "" {
				// проверка заголовка ответа
				assert.Equal(t, tt.expectedLocation, resp.Header.Get("Location"), "Заголовок ответа не совпадает с ожидаемым")
			}
		})
	}
}
