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
	db, err := sql.Open("pgx", conf.FlagDB)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		log.GetLog().Sugar().Infow("Use Mem Storage", "Can not ping DB", err)
		store = storage.NewMemStorage(conf.FlagStoragePath)
	} else {
		log.GetLog().Sugar().Infow("Use DB", "dsn", conf.FlagDB)
		store = storage.NewDBStorage(context.Background(), db)
		defer db.Close()
	}

	srv := service.NewService(store)
	zip := zipper.NewZipper()
	handls := handlers.NewHandlers(srv, conf, log, zip)

	log.GetLog().Sugar().Infow("Log level", "level", conf.FlagLogLevel)
	log.GetLog().Sugar().Infow("File storage", "file", conf.FlagStoragePath)
	log.GetLog().Sugar().Infow("Running server", "address", conf.FlagRunAddr)

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
