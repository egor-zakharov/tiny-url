package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
)

// Config - struct
type Config struct {
	FlagRunAddr       string `json:"server_address"`
	FlagShortAddr     string `json:"base_url"`
	FlagLogLevel      string
	FlagStoragePath   string `json:"file_storage_path"`
	FlagDB            string `json:"database_dsn"`
	FlagHTTPS         bool   `json:"enable_https"`
	FlagConfigPath    string
	FlagTrustedSubnet string `json:"trusted_subnet"`
}

// NewConfig - constructor Config
func NewConfig() *Config {
	return &Config{}
}

// ParseFlag - parsing command line, env flags or config file
func (c *Config) ParseFlag() {
	flag.StringVar(&c.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.FlagShortAddr, "b", "http://localhost:8080", "address and port to short url")
	flag.StringVar(&c.FlagLogLevel, "l", "info", "log level")
	flag.StringVar(&c.FlagStoragePath, "f", "C:\\Users\\edzakharov\\Documents\\GoAdv\\tiny-url\\short-url-db.json", "file storage path")
	flag.StringVar(&c.FlagDB, "d", "postgres://admin:admin@localhost:5432/db?sslmode=disable", "database dsn")
	flag.BoolVar(&c.FlagHTTPS, "s", false, "https enable")
	flag.StringVar(&c.FlagConfigPath, "c", "", "config path")
	flag.StringVar(&c.FlagTrustedSubnet, "t", "", "CIDR")
	flag.Parse()

	if envConfigPath := os.Getenv("CONFIG"); envConfigPath != "" {
		c.FlagConfigPath = envConfigPath
	}

	fileConfig := Config{}
	if c.FlagConfigPath != "" {
		fileConfig = configFromFile(c.FlagConfigPath)

		if !isFlagPresented("a") {
			c.FlagRunAddr = fileConfig.FlagRunAddr
		}

		if !isFlagPresented("b") {
			c.FlagShortAddr = fileConfig.FlagShortAddr
		}

		if !isFlagPresented("f") {
			c.FlagStoragePath = fileConfig.FlagStoragePath
		}

		if !isFlagPresented("d") {
			c.FlagDB = fileConfig.FlagDB
		}

		if !isFlagPresented("s") {
			c.FlagHTTPS = fileConfig.FlagHTTPS
		}

		if !isFlagPresented("t") {
			c.FlagTrustedSubnet = fileConfig.FlagTrustedSubnet
		}
	}

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

	if envHTTPS := os.Getenv("ENABLE_HTTPS"); envHTTPS != "" {
		c.FlagHTTPS, _ = strconv.ParseBool(envHTTPS)
	}

	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		c.FlagTrustedSubnet = envTrustedSubnet
	}
}

func configFromFile(fileName string) Config {
	file, err := os.Open(fileName)

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	if err != nil {
		log.Fatal(err)
	}
	config := Config{}
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

func isFlagPresented(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
