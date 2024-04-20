package main

import (
	"github.com/egor-zakharov/tiny-url/internal/app/service"
	"github.com/egor-zakharov/tiny-url/internal/app/storage"
	"net/http"
	"net/url"

	"github.com/egor-zakharov/tiny-url/internal/app/config"
	"github.com/egor-zakharov/tiny-url/internal/app/handlers"
	"github.com/egor-zakharov/tiny-url/internal/app/logger"
)

func main() {
	conf := config.NewConfig()
	conf.ParseFlag()
	log := logger.NewLogger()

	err := log.Initialize(conf.FlagLogLevel)
	if err != nil {
		panic(err)
	}

	runURL, err := url.Parse(conf.FlagShortAddr)
	if err != nil {
		panic(err)
	}

	store := storage.New()
	srv := service.NewService(store)
	handls := handlers.NewHandlers(srv, *runURL, log)

	log.GetLog().Sugar().Infow("Log level", "level", conf.FlagLogLevel)
	log.GetLog().Sugar().Infow("Running server", "address", conf.FlagRunAddr)

	err = http.ListenAndServe(conf.FlagRunAddr, handls.ChiRouter())
	if err != nil {
		panic(err)
	}
}
