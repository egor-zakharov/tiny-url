package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/egor-zakharov/tiny-url/internal/app/auth"
	"github.com/egor-zakharov/tiny-url/internal/app/config"
	"github.com/egor-zakharov/tiny-url/internal/app/grpchandlers"
	"github.com/egor-zakharov/tiny-url/internal/app/handlers"
	"github.com/egor-zakharov/tiny-url/internal/app/logger"
	"github.com/egor-zakharov/tiny-url/internal/app/service"
	"github.com/egor-zakharov/tiny-url/internal/app/storage"
	"github.com/egor-zakharov/tiny-url/internal/app/tls"
	"github.com/egor-zakharov/tiny-url/internal/app/whitelist"
	"github.com/egor-zakharov/tiny-url/internal/app/zipper"
	pb "github.com/egor-zakharov/tiny-url/internal/proto"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"syscall"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
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
		defer store.Backup()
	} else {
		log.GetLog().Sugar().Infow("Use DB", "dsn", conf.FlagDB)
		store = storage.NewDBStorage(context.Background(), db)
		defer db.Close()
	}

	var trustedNet *net.IPNet

	if conf.FlagTrustedSubnet != "" {
		_, trustedNet, err = net.ParseCIDR(conf.FlagTrustedSubnet)
		if err != nil {
			log.GetLog().Sugar().Infow("Subnet", "Can not parse", err)
		}
	}

	srv := service.NewService(store)
	zip := zipper.NewZipper()
	authz := auth.NewAuth()
	whiteList := whitelist.NewWhiteList(trustedNet)
	handls := handlers.NewHandlers(srv, conf, log, zip, authz, whiteList)
	grpcHandlers := grpchandlers.NewShortenerServer(srv, log, authz, conf)

	log.GetLog().Sugar().Infow("Log level", "level", conf.FlagLogLevel)
	log.GetLog().Sugar().Infow("File storage", "file", conf.FlagStoragePath)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	if conf.FlagRunGRPCAddr != "" {
		go func() {
			listen, err := net.Listen("tcp", conf.FlagRunGRPCAddr)
			if err != nil {
				panic(err)
			}
			excludeMethod := []string{
				"/proto.shortener.ShortenerService/Auth",
				"/proto.shortener.ShortenerService/GetURL",
				"/proto.shortener.ShortenerService/Stats"}
			s := grpc.NewServer(grpc.UnaryInterceptor(auth.ExcludeMethodsInterceptor(excludeMethod, auth.Interceptor)))
			pb.RegisterShortenerServiceServer(s, grpcHandlers)
			reflection.Register(s)

			err = s.Serve(listen)
			log.GetLog().Sugar().Infow("Running grpc server", "address", conf.FlagRunGRPCAddr)
			if err != nil {
				panic(err)
			}
		}()
	} else {
		go func() {
			if conf.FlagHTTPS {
				log.GetLog().Sugar().Infow("Running server on https", "enabled", conf.FlagHTTPS, "address", conf.FlagRunAddr)
				const (
					certFilePath = "cert.pem" // certFilePath - path to TLS certificate
					keyFilePath  = "key.pem"  // keyFilePath - path to TLS key
				)
				err = tls.CreateTLSCert(certFilePath, keyFilePath)
				err = http.ListenAndServeTLS(conf.FlagRunAddr, certFilePath, keyFilePath, handls.ChiRouter())
			} else {
				log.GetLog().Sugar().Infow("Running server", "address", conf.FlagRunAddr)
				err = http.ListenAndServe(conf.FlagRunAddr, handls.ChiRouter())
			}

			if err != nil {
				panic(err)
			}
		}()
	}
	<-ctx.Done()

}
