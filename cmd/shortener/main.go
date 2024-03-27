package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

var urls = make(map[string]string)

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
	w.Write([]byte(fmt.Sprintf("%s%s", "http://localhost:8080/", shortURL)))
}

func get(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.String()[1:]

	if value, found := urls[shortURL]; found {
		//формирование ответа
		w.Header().Set("Location", value)
		w.WriteHeader(http.StatusTemporaryRedirect)

	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func router(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		post(w, r)
	case http.MethodGet:
		get(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", router)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}

func encodeURL(url []byte) string {
	encoded := base64.StdEncoding.EncodeToString(url)
	// возвращаем 8 символов, отрезаем последние 2 ==
	return encoded[len(encoded)-10 : len(encoded)-2]
}
