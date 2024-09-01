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
	"github.com/egor-zakharov/tiny-url/internal/app/whitelist"
	"github.com/egor-zakharov/tiny-url/internal/app/zipper"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"time"
)

// Handlers - handlers struct
type Handlers struct {
	service   service.Service
	log       *logger.Logger
	config    *config.Config
	zip       *zipper.Zipper
	auth      *auth.Auth
	whitelist *whitelist.WhiteList
}

// NewHandlers - constructor Handlers
func NewHandlers(service service.Service, config *config.Config, log *logger.Logger, zipper *zipper.Zipper, auth *auth.Auth, whitelist *whitelist.WhiteList) *Handlers {
	return &Handlers{
		service:   service,
		config:    config,
		log:       log,
		zip:       zipper,
		auth:      auth,
		whitelist: whitelist,
	}
}

// ChiRouter - router with paths
func (h *Handlers) ChiRouter() http.Handler {
	r := chi.NewRouter()

	r.Mount("/debug", middleware.Profiler())
	r.Get("/{link}", h.log.RequestLogger(h.zip.GzipMiddleware(h.Get)))
	r.Get("/ping", h.log.RequestLogger(h.zip.GzipMiddleware(h.Ping)))
	r.Get("/api/user/urls", h.log.RequestLogger(h.zip.GzipMiddleware(h.GetAll)))
	r.Post("/", h.log.RequestLogger(h.zip.GzipMiddleware(h.Post)))
	r.Post("/api/shorten", h.log.RequestLogger(h.zip.GzipMiddleware(h.PostShorten)))
	r.Post("/api/shorten/batch", h.log.RequestLogger(h.zip.GzipMiddleware(h.PostShortenBatch)))
	r.Delete("/api/user/urls", h.log.RequestLogger(h.zip.GzipMiddleware(h.DeleteBatch)))
	r.Get("/api/internal/stats", h.log.RequestLogger(h.zip.GzipMiddleware(h.whitelist.Handler(h.GetStats))))

	return r
}

// Get - handle get /{link} - get record
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	//берем параметр урла
	shortURL := chi.URLParam(r, "link")
	//идем в app
	url, err := h.service.Get(r.Context(), shortURL)

	if err != nil {
		if errors.Is(err, storage.ErrDeletedURL) {
			w.WriteHeader(http.StatusGone)
			w.Write([]byte(url))
			return
		}
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte(url))
		return
	}
	w.Header().Add("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// GetAll - handle get /api/user/urls - get all user's records
func (h *Handlers) GetAll(w http.ResponseWriter, r *http.Request) {
	//получаем ID
	ID, err := h.auth.GetID(w, r)
	if err != nil {
		h.log.GetLog().Sugar().With("error", err).Error("get ID from token")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}

	//TODO При отсутствии сокращённых пользователем URL хендлер должен отдавать HTTP-статус 204 No Content
	//В тесте /fetch_no_urls  expected: 401, точно ли валидный код ответа?
	urls, err := h.service.GetAll(r.Context(), ID)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		h.log.GetLog().Sugar().With("error", err).Error("service error")
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

// Post - handle / - add record
func (h *Handlers) Post(w http.ResponseWriter, r *http.Request) {

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

// PostShorten - handle /api/shorten - add shorten record
func (h *Handlers) PostShorten(w http.ResponseWriter, r *http.Request) {
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
	if err = dec.Decode(&req); err != nil {
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

// PostShortenBatch - handle /api/shorten/batch - add batch shorten record
func (h *Handlers) PostShortenBatch(w http.ResponseWriter, r *http.Request) {
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
	if err = dec.Decode(&req); err != nil {
		h.log.GetLog().Sugar().With("error", err).Error("decode request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, v := range req {
		//валидируем полученное тело`
		err = h.service.ValidateURL(v.URL)
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

// DeleteBatch - handle /api/user/urls - delete batch record
func (h *Handlers) DeleteBatch(w http.ResponseWriter, r *http.Request) {
	ID, err := h.auth.GetID(w, r)
	if err != nil {
		h.log.GetLog().Sugar().With("error", err).Error("get ID from token")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(fmt.Sprint(err)))
		return
	}

	var req models.DeleteBatchRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		h.log.GetLog().Sugar().With("error", err).Error("decode request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ch := h.generator(req)
	_ = h.deleteURL(ch, ID)

	//формирование ответа
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

}

// GetStats - handle /api/internal/stats - get stats
func (h *Handlers) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := models.StatsResponse(stats)

	//формирование ответа
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// сериализуем ответ сервера
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		return
	}
}

// Ping - handle /ping - ping db
func (h *Handlers) Ping(w http.ResponseWriter, r *http.Request) {
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

func (h *Handlers) generator(input []string) chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		for _, v := range input {
			ch <- v
		}
	}()

	return ch
}

func (h *Handlers) deleteURL(ch <-chan string, ID string) error {
	var errs error
	for URL := range ch {
		err := h.service.Delete(URL, ID)
		if err != nil {
			errs = err
		}
	}
	return errs
}
