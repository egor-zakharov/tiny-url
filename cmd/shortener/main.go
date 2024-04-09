package main

import (
	"encoding/base64"
	"io"
	"net/http"

	"github.com/egor-zakharov/tiny-url/internal/app/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var urls = make(map[string]string)
var flagShortAddr string

func post(w http.ResponseWriter, r *http.Request) {
	//проверка тела запроса
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	shortURL := encodeURL(body)

	//добавляем в мапку
	urls[shortURL] = string(body)
	//формирование ответа
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(flagShortAddr + "/" + shortURL))
}

func get(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "link")

	if value, found := urls[shortURL]; found {
		//формирование ответа
		w.Header().Add("Location", value)
		w.WriteHeader(http.StatusTemporaryRedirect)

	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func chiRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/{link}", get)
	r.Post("/", post)

	return r
}

func main() {
	cfg := config.NewConfig()
	cfg.ParseFlag()
	flagShortAddr = cfg.FlagShortAddr
	err := http.ListenAndServe(cfg.FlagRunAddr, chiRouter())
	if err != nil {
		panic(err)
	}
}

func encodeURL(url []byte) string {
	encoded := base64.StdEncoding.EncodeToString(url)
	// возвращаем 8 символов, отрезаем последние 2 ==
	return encoded[len(encoded)-10 : len(encoded)-2]
}
