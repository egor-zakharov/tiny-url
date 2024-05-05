package main

import (
	"net/http"

	"github.com/egor-zakharov/tiny-url/internal/app/config"
	"github.com/egor-zakharov/tiny-url/internal/app/handlers"
	"github.com/egor-zakharov/tiny-url/internal/app/logger"
	"github.com/egor-zakharov/tiny-url/internal/app/service"
	"github.com/egor-zakharov/tiny-url/internal/app/storage"
	"github.com/egor-zakharov/tiny-url/internal/app/zipper"
)

func main() {
	conf := config.NewConfig()
	conf.ParseFlag()
	log := logger.NewLogger()

	err := log.Initialize(conf.FlagLogLevel)
	if err != nil {
		panic(err)
	}

	store := storage.New(conf.FlagStoragePath)
	srv := service.NewService(store)
	zip := zipper.NewZipper()
	handls := handlers.NewHandlers(srv, conf, log, zip)

	log.GetLog().Sugar().Infow("Log level", "level", conf.FlagLogLevel)
	log.GetLog().Sugar().Infow("File storage", "file", conf.FlagStoragePath)
	log.GetLog().Sugar().Infow("Running server", "address", conf.FlagRunAddr)
	log.GetLog().Sugar().Infow("DB", "dsn", conf.FlagDB)

	err = http.ListenAndServe(conf.FlagRunAddr, handls.ChiRouter())
	if err != nil {
		panic(err)
	}
}
