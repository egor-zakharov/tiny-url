package main

import (
	"net/http"
	"net/url"

	"github.com/egor-zakharov/tiny-url/internal/app/config"
	"github.com/egor-zakharov/tiny-url/internal/app/handlers"
	"github.com/egor-zakharov/tiny-url/internal/app/logger"
)

func main() {
	conf := config.NewConfig()
	conf.ParseFlag()

	err := logger.Initialize(conf.FlagLogLevel)
	if err != nil {
		panic(err)
	}

	runURL, err := url.Parse(conf.FlagShortAddr)
	if err != nil {
		panic(err)
	}

	logger.Log.Sugar().Infow("Running server", "address", conf.FlagRunAddr)

	err = http.ListenAndServe(conf.FlagRunAddr, handlers.ChiRouter(handlers.New(*runURL)))
	if err != nil {
		panic(err)
	}
}
