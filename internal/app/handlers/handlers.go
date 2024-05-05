package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/egor-zakharov/tiny-url/internal/app/config"
	"github.com/egor-zakharov/tiny-url/internal/app/logger"
	"github.com/egor-zakharov/tiny-url/internal/app/models"
	"github.com/egor-zakharov/tiny-url/internal/app/service"
	"github.com/egor-zakharov/tiny-url/internal/app/zipper"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/url"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Handlers struct {
	service *service.Service
	log     *logger.Logger
	config  *config.Config
	zip     *zipper.Zipper
}

func NewHandlers(service *service.Service, config *config.Config, log *logger.Logger, zipper *zipper.Zipper) *Handlers {
	return &Handlers{
		service: service,
		config:  config,
		log:     log,
		zip:     zipper,
	}
}

func (h *Handlers) ChiRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/{link}", h.log.RequestLogger(h.zip.GzipMiddleware(h.get)))
	r.Get("/ping", h.log.RequestLogger(h.zip.GzipMiddleware(h.ping)))
	r.Post("/", h.log.RequestLogger(h.zip.GzipMiddleware(h.post)))
	r.Post("/api/shorten", h.log.RequestLogger(h.zip.GzipMiddleware(h.postShorten)))

	return r
}

func (h *Handlers) get(w http.ResponseWriter, r *http.Request) {
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
func (h *Handlers) post(w http.ResponseWriter, r *http.Request) {
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

	newURL, err := url.Parse(h.config.FlagShortAddr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.GetLog().Sugar().With("error", err).Error("short addr parse")
		return
	}
	newURL.Path = shortURL

	//формирование ответа
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(newURL.String()))
}

// Было бы тоже неплохо замокать ответ сервиса в тесте на успешное получение
func (h *Handlers) postShorten(w http.ResponseWriter, r *http.Request) {
	var req models.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		h.log.GetLog().Sugar().With("error", err).Error("decode request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//валидируем полученное тело
	err := h.service.ValidateURL(req.URL)
	if err != nil {
		h.log.GetLog().Sugar().With("error", err).Error("validate url")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}

	//Кодируем и добавляем с сторейдж
	shortURL, err := h.service.Add(req.URL)
	if err != nil {
		h.log.GetLog().Sugar().With("error", err).Error("add storage")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}

	newURL, err := url.Parse(h.config.FlagShortAddr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.GetLog().Sugar().With("error", err).Error("short addr parse")
		return
	}
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

func (h *Handlers) ping(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("pgx", h.config.FlagDB)
	if err != nil {
		h.log.GetLog().Sugar().With("error", err).Error("can not open DB")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		h.log.GetLog().Sugar().With("error", err).Error("can not ping DB")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
