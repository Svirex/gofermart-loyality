package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	SecretKey            string `env:"SECRET_KEY"`
}

func ParseEnv() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse enviroment variables: %w", err)
	}
	return cfg, nil
}

func ParseFlags() (*Config, error) {
	cfg := &Config{}
	flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "<host>:<port>")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "DATABASE_URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "ACCRUAL_SYSTEM_ADDRESS")
	flag.StringVar(&cfg.SecretKey, "k", "fake_secret_key", "secret key for auth")
	flag.Parse()
	return cfg, nil
}

func Parse() (*Config, error) {
	envCfg, err := ParseEnv()
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	flagConfig, err := ParseFlags()
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	cfg := mergeConf(envCfg, flagConfig)
	return cfg, nil
}

func mergeConf(envCfg *Config, flagConfig *Config) *Config {
	cfg := &Config{
		RunAddress:           envCfg.RunAddress,
		DatabaseURI:          envCfg.DatabaseURI,
		AccrualSystemAddress: envCfg.AccrualSystemAddress,
		SecretKey:            envCfg.SecretKey,
	}
	if cfg.RunAddress == "" {
		cfg.RunAddress = flagConfig.RunAddress
	}
	if cfg.DatabaseURI == "" {
		cfg.DatabaseURI = flagConfig.DatabaseURI
	}
	if cfg.AccrualSystemAddress == "" {
		cfg.AccrualSystemAddress = flagConfig.AccrualSystemAddress
	}
	if cfg.SecretKey == "" {
		cfg.SecretKey = flagConfig.SecretKey
	}
	return cfg
}
