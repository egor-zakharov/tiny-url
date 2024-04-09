package main

import (
	"net/http"
	"net/url"

	"github.com/egor-zakharov/tiny-url/internal/app/config"
	"github.com/egor-zakharov/tiny-url/internal/app/handlers"
)

func main() {
	conf := config.NewConfig()
	conf.ParseFlag()

	runURL, err := url.Parse(conf.FlagShortAddr)
	if err != nil {
		panic(err)
	}

	err = http.ListenAndServe(conf.FlagRunAddr, handlers.ChiRouter(handlers.New(*runURL)))
	if err != nil {
		panic(err)
	}
}
