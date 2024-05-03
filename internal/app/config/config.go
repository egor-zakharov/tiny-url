package config

import (
	"flag"
	"os"
)

type Config struct {
	FlagRunAddr     string
	FlagShortAddr   string
	FlagLogLevel    string
	FlagStoragePath string
	FlagDB          string
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) ParseFlag() {
	flag.StringVar(&c.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.FlagShortAddr, "b", "http://localhost:8080", "address and port to short url")
	flag.StringVar(&c.FlagLogLevel, "l", "info", "log level")
	flag.StringVar(&c.FlagStoragePath, "f", "/tmp/short-url-db.json", "file storage path")
	flag.StringVar(&c.FlagDB, "d", "postgres://postgres:admin@localhost:5432/postgres?sslmode=disable", "database dsn")

	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		c.FlagRunAddr = envRunAddr
	}

	if envShortAddr := os.Getenv("BASE_URL"); envShortAddr != "" {
		c.FlagShortAddr = envShortAddr
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		c.FlagLogLevel = envLogLevel
	}

	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
		c.FlagStoragePath = envFilePath
	}

	if envDB := os.Getenv("DATABASE_DSN"); envDB != "" {
		c.FlagDB = envDB
	}
}
