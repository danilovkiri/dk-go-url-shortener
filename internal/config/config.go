// Package config provides types for handling configuration parameters.
package config

import "github.com/caarlos0/env/v6"

// Config handles server-related constants and parameters.
type Config struct {
	ServerConfig *ServerConfig
}

// ServerConfig defines default server-relates constants and parameters and overwrites them with environment variables.
type ServerConfig struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"localhost:8080"`
}

// NewDefaultConfiguration sets up a server configuration.
func NewDefaultConfiguration() (*Config, error) {
	cfg := ServerConfig{}
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return &Config{
		ServerConfig: &cfg,
	}, nil
}
