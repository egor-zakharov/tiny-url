package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/egor-zakharov/tiny-url/internal/app/models"
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
			baseURL, err := url.Parse(baseURL)
			if err != nil {
				t.Errorf("Fail to start test - %v", err)
			}
			ts := httptest.NewServer(ChiRouter(New(*baseURL)))
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
			baseURL, err := url.Parse(baseURL)
			if err != nil {
				t.Errorf("Fail to start test - %v", err)
			}
			ts := httptest.NewServer(ChiRouter(New(*baseURL)))
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

func Test_get(t *testing.T) {
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
			baseURL, err := url.Parse(baseURL)
			if err != nil {
				t.Errorf("Fail to start test - %v", err)
			}
			h := New(*baseURL)
			if tt.expectedLocation != "" {
				h.service.Add(tt.expectedLocation)
			}
			ts := httptest.NewServer(ChiRouter(h))
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
