package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/egor-zakharov/tiny-url/internal/app/auth"
	"github.com/egor-zakharov/tiny-url/internal/app/config"
	"github.com/egor-zakharov/tiny-url/internal/app/logger"
	"github.com/egor-zakharov/tiny-url/internal/app/models"
	"github.com/egor-zakharov/tiny-url/internal/app/service"
	"github.com/egor-zakharov/tiny-url/internal/app/storage"
	"github.com/egor-zakharov/tiny-url/internal/app/zipper"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/url"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Handlers struct {
	service service.Service
	log     *logger.Logger
	config  *config.Config
	zip     *zipper.Zipper
	auth    *auth.Auth
}

func NewHandlers(service service.Service, config *config.Config, log *logger.Logger, zipper *zipper.Zipper, auth *auth.Auth) *Handlers {
	return &Handlers{
		service: service,
		config:  config,
		log:     log,
		zip:     zipper,
		auth:    auth,
	}
}

func (h *Handlers) ChiRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/{link}", h.log.RequestLogger(h.zip.GzipMiddleware(h.get)))
	r.Get("/ping", h.log.RequestLogger(h.zip.GzipMiddleware(h.ping)))
	r.Get("/api/user/urls", h.log.RequestLogger(h.zip.GzipMiddleware(h.getAll)))
	r.Post("/", h.log.RequestLogger(h.zip.GzipMiddleware(h.post)))
	r.Post("/api/shorten", h.log.RequestLogger(h.zip.GzipMiddleware(h.postShorten)))
	r.Post("/api/shorten/batch", h.log.RequestLogger(h.zip.GzipMiddleware(h.postShortenBatch)))

	return r
}

func (h *Handlers) get(w http.ResponseWriter, r *http.Request) {

	//получаем ID
	ID, err := h.auth.GetID(w, r)
	if err != nil {
		h.log.GetLog().Sugar().With("error", err).Error("get ID from token")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}

	//берем параметр урла
	shortURL := chi.URLParam(r, "link")
	//идем в app
	url, err := h.service.Get(r.Context(), shortURL, ID)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte(url))
		return
	}
	w.Header().Add("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handlers) getAll(w http.ResponseWriter, r *http.Request) {

	//получаем ID
	ID, err := h.auth.GetID(w, r)
	if err != nil {
		h.log.GetLog().Sugar().With("error", err).Error("get ID from token")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}

	//идем в app
	urls, err := h.service.GetAll(r.Context(), ID)
	if err != nil || len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		h.log.GetLog().Sugar().With("error", err).Error("get ID from token")
		return
	}

	newURL, err := url.Parse(h.config.FlagShortAddr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.GetLog().Sugar().With("error", err).Error("short addr parse")
		return
	}

	// заполняем модель ответа
	var resp []models.UserURLsResponse

	for shortURL, originalURL := range urls {
		newURL.Path = shortURL
		resp = append(resp, models.UserURLsResponse{
			ShortURL:    newURL.String(),
			OriginalURL: originalURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		return
	}

}

// Было бы тоже неплохо замокать ответ сервиса в тесте на успешное получение
func (h *Handlers) post(w http.ResponseWriter, r *http.Request) {

	//получаем ID
	ID, err := h.auth.GetID(w, r)
	if err != nil {
		h.log.GetLog().Sugar().With("error", err).Error("get ID from token")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}

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

	newURL, err := url.Parse(h.config.FlagShortAddr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.GetLog().Sugar().With("error", err).Error("short addr parse")
		return
	}

	//Кодируем и добавляем с сторейдж
	shortURL, err := h.service.Add(r.Context(), stringBody, ID)
	newURL.Path = shortURL
	if err != nil {
		if errors.Is(err, storage.ErrConflict) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(fmt.Sprint(newURL)))
			return
		}
		h.log.GetLog().Sugar().With("error", err).Error("add storage")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}

	//формирование ответа
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(newURL.String()))
}

// Было бы тоже неплохо замокать ответ сервиса в тесте на успешное получение
func (h *Handlers) postShorten(w http.ResponseWriter, r *http.Request) {
	//получаем ID
	ID, err := h.auth.GetID(w, r)
	if err != nil {
		h.log.GetLog().Sugar().With("error", err).Error("get ID from token")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}

	var req models.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		h.log.GetLog().Sugar().With("error", err).Error("decode request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//валидируем полученное тело
	err = h.service.ValidateURL(req.URL)
	if err != nil {
		h.log.GetLog().Sugar().With("error", err).Error("validate url")
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
	//Кодируем и добавляем с сторейдж
	shortURL, err := h.service.Add(r.Context(), req.URL, ID)
	newURL.Path = shortURL
	// заполняем модель ответа
	resp := models.Response{
		Result: newURL.String(),
	}

	if err != nil && !errors.Is(err, storage.ErrConflict) {
		h.log.GetLog().Sugar().With("error", err).Error("add storage")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}
	if errors.Is(err, storage.ErrConflict) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
	}
	// сериализуем ответ сервера
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		return
	}
}

// Было бы тоже неплохо замокать ответ сервиса в тесте на успешное получение
func (h *Handlers) postShortenBatch(w http.ResponseWriter, r *http.Request) {
	//получаем ID
	ID, err := h.auth.GetID(w, r)
	if err != nil {
		h.log.GetLog().Sugar().With("error", err).Error("get ID from token")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}

	var req []models.ShortenBatchRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		h.log.GetLog().Sugar().With("error", err).Error("decode request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, v := range req {
		//валидируем полученное тело`
		err := h.service.ValidateURL(v.URL)
		if err != nil {
			h.log.GetLog().Sugar().With("error", err).Error("validate url")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprint(err)))
			return
		}
	}

	urls := make(map[string]string, len(req))

	for _, v := range req {
		urls[v.CorrelationID] = v.URL
	}

	//Кодируем и добавляем с сторейдж
	shortURLs, err := h.service.AddBatch(r.Context(), urls, ID)
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

	// заполняем модель ответа
	var resp []models.ShortenBatchResponse

	for corID, shortURL := range shortURLs {
		newURL.Path = shortURL
		resp = append(resp, models.ShortenBatchResponse{
			CorrelationID: corID,
			ShortURL:      newURL.String(),
		})
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
