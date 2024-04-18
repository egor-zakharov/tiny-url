package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/egor-zakharov/tiny-url/internal/app/logger"
	"github.com/egor-zakharov/tiny-url/internal/app/models"
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
	r.Post("/api/shorten", logger.RequestLogger(h.PostShorten))

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

// Было бы тоже неплохо замокать ответ сервиса в тесте на успешное получение
func (h *Handlers) PostShorten(w http.ResponseWriter, r *http.Request) {
	var req models.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Sugar().With("error", err).Debug("decode request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//валидируем полученное тело
	err := h.service.ValidateURL(req.URL)
	if err != nil {
		logger.Log.Sugar().With("error", err).Debug("validate url")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}

	//Кодируем и добавляем с сторейдж
	shortURL, err := h.service.Add(req.URL)
	if err != nil {
		logger.Log.Sugar().With("error", err).Debug("add to storage")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}

	newURL := h.flagShortAddr
	newURL.Path = shortURL

	// заполняем модель ответа
	resp := models.Response{
		Result: newURL.String(),
	}

	//формирование ответа

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// сериализуем ответ сервера

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		return
	}
}
