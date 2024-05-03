package main

import (
	"context"
	"database/sql"
	"github.com/egor-zakharov/tiny-url/internal/app/config"
	"github.com/egor-zakharov/tiny-url/internal/app/handlers"
	"github.com/egor-zakharov/tiny-url/internal/app/logger"
	"github.com/egor-zakharov/tiny-url/internal/app/service"
	"github.com/egor-zakharov/tiny-url/internal/app/storage"
	"github.com/egor-zakharov/tiny-url/internal/app/zipper"
	"net/http"
	"os/signal"
	"syscall"
)

func main() {
	conf := config.NewConfig()
	conf.ParseFlag()
	log := logger.NewLogger()

	err := log.Initialize(conf.FlagLogLevel)
	if err != nil {
		panic(err)
	}

	var store storage.Storage
	if conf.FlagDB == "" {
		store = storage.NewMemStorage(conf.FlagStoragePath)
	} else {
		db, err := sql.Open("pgx", conf.FlagDB)

		if err != nil {
			log.GetLog().Sugar().With("error", err).Error("can not open DB")
			panic(err)
		}
		store = storage.NewDBStorage(context.Background(), db)
		defer db.Close()
	}

	srv := service.NewService(store)
	zip := zipper.NewZipper()
	handls := handlers.NewHandlers(srv, conf, log, zip)

	log.GetLog().Sugar().Infow("Log level", "level", conf.FlagLogLevel)
	log.GetLog().Sugar().Infow("File storage", "file", conf.FlagStoragePath)
	log.GetLog().Sugar().Infow("Running server", "address", conf.FlagRunAddr)
	log.GetLog().Sugar().Infow("DB", "dsn", conf.FlagDB)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer stop()

	go func() {
		err = http.ListenAndServe(conf.FlagRunAddr, handls.ChiRouter())
		if err != nil {
			panic(err)
		}
	}()
	<-ctx.Done()

}
