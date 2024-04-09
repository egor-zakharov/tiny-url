package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_post(t *testing.T) {
	tests := []struct {
		name                 string
		method               string
		requestBody          string
		expectedCode         int
		expectedResponseBody string
	}{
		{name: "Проверка запроса без тела", method: http.MethodPost, requestBody: "", expectedCode: http.StatusBadRequest, expectedResponseBody: ""},
		{name: "Проверка запроса без тела", method: http.MethodPost, requestBody: "https://practicum.yandex.ru/", expectedCode: http.StatusCreated, expectedResponseBody: "http://localhost:8080/V4LnJ1Lw"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stringReader := strings.NewReader(tt.requestBody)
			r := httptest.NewRequest(tt.method, "/", stringReader)
			w := httptest.NewRecorder()

			// вызовем хендлер как обычную функцию, без запуска самого сервера
			post(w, r)

			assert.Equal(t, tt.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			// проверим корректность полученного тела ответа, если мы его ожидаем
			if tt.expectedResponseBody != "" {
				// проверка тела ответа
				assert.Equal(t, tt.expectedResponseBody, w.Body.String(), "Тело ответа не совпадает с ожидаемым")
			}
		})
	}
}

func Test_get(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		expectedCode     int
		expectedLocation string
	}{
		{name: "Проверка отсутствующего URL", method: http.MethodGet, path: "/urlNotFound", expectedCode: http.StatusBadRequest, expectedLocation: ""},
		{name: "Проверка Location", method: http.MethodGet, path: "/someTest", expectedCode: http.StatusTemporaryRedirect, expectedLocation: "https://practicum.yandex.ru/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			urls["someTest"] = "https://practicum.yandex.ru/"

			// вызовем хендлер как обычную функцию, без запуска самого сервера
			get(w, r)

			assert.Equal(t, tt.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			// проверим корректность полученного тела ответа, если мы его ожидаем
			if tt.expectedLocation != "" {
				// проверка заголовка ответа
				assert.Equal(t, tt.expectedLocation, w.Header().Get("Location"), "Заголовок ответа не совпадает с ожидаемым")
			}
		})
	}
}

func Test_router(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		expectedCode int
	}{
		{name: "Проверка роута метода GET", method: http.MethodGet, expectedCode: http.StatusBadRequest},
		{name: "Проверка роута метода POST", method: http.MethodPost, expectedCode: http.StatusBadRequest},
		{name: "Проверка роута метода PUT", method: http.MethodPut, expectedCode: http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()

			// вызовем хендлер как обычную функцию, без запуска самого сервера
			router(w, r)
			// проверка кода ответа
			assert.Equal(t, tt.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}
