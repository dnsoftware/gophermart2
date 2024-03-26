package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/dnsoftware/gophermart2/internal/constants"
	"log"
)

type Config struct {
	RunAddress     string `env:"RUN_ADDRESS"`
	DatabaseURI    string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

type confFlags struct {
	runAddress     string
	databaseURI    string
	accrualAddress string
}

func NewServerConfig() *Config {
	cfg := &Config{}
	cFlags := confFlags{}

	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&cFlags.runAddress, "a", constants.RunAddress, "server endpoint")
	flag.StringVar(&cFlags.databaseURI, "d", "", "data source name")
	flag.StringVar(&cFlags.accrualAddress, "r", constants.AccrualAddress, "accrual endpoint")
	flag.Parse()

	// если какого-то параметра нет в переменных окружения - берем значение флага, а если и флага нет - берем по умолчанию
	if cfg.RunAddress == "" {
		cfg.RunAddress = cFlags.runAddress
	}

	if cfg.DatabaseURI == "" {
		cfg.DatabaseURI = cFlags.databaseURI
	}

	if cfg.AccrualAddress == "" {
		cfg.AccrualAddress = cFlags.accrualAddress
	}

	return cfg
}
