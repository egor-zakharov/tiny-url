package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/egor-zakharov/tiny-url/internal/app/logger"
	"github.com/egor-zakharov/tiny-url/internal/app/service"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	service       *service.Service
	flagShortAddr url.URL
}

func New(flagShortAddr url.URL) *Handlers {
	return &Handlers{service.New(), flagShortAddr}
}

func ChiRouter(h *Handlers) http.Handler {
	r := chi.NewRouter()

	r.Get("/{link}", logger.RequestLogger(h.Get))
	r.Post("/", logger.RequestLogger(h.Post))

	return r
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	//берем параметр урла
	shortURL := chi.URLParam(r, "link")
	//идем в app
	url, err := h.service.Get(shortURL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(url))
		return
	}
	w.Header().Add("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// Было бы тоже неплохо замокать ответ сервиса в тесте на успешное получение
func (h *Handlers) Post(w http.ResponseWriter, r *http.Request) {
	//проверка тела запроса
	body, err := io.ReadAll(r.Body)
	//проверяем ошибку чтения тела запроса
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	stringBody := string(body)
	//валидируем полученное тело
	err = h.service.ValidateURL(stringBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}

	//Кодируем и добавляем с сторейдж
	shortURL, err := h.service.Add(stringBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}

	newURL := h.flagShortAddr
	newURL.Path = shortURL

	//формирование ответа
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(newURL.String()))
}
