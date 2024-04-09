package config

import (
	"flag"
)

type Config struct {
	FlagRunAddr   string
	FlagShortAddr string
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) ParseFlag() {
	flag.StringVar(&c.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.FlagShortAddr, "b", "http://localhost:8080", "address and port to short url")
	flag.Parse()
}
