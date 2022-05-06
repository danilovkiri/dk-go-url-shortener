// Package config provides types for handling configuration parameters.
package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

// Config handles server-related constants and parameters.
type Config struct {
	ServerConfig  *ServerConfig
	StorageConfig *StorageConfig
}

// ServerConfig defines default server-relates constants and parameters and overwrites them with environment variables.
type ServerConfig struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

// StorageConfig retrieves file storage-related parameters from environment.
type StorageConfig struct {
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"storage/infile/url_storage.json"`
}

// NewStorageConfig sets up a storage configuration.
func NewStorageConfig() (*StorageConfig, error) {
	cfg := StorageConfig{}
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// NewServerConfig sets up a server configuration.
func NewServerConfig() (*ServerConfig, error) {
	cfg := ServerConfig{}
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// NewDefaultConfiguration sets up a total configuration.
func NewDefaultConfiguration() (*Config, error) {
	serverCfg, err := NewServerConfig()
	if err != nil {
		return nil, err
	}
	storageCfg, err := NewStorageConfig()
	if err != nil {
		return nil, err
	}
	return &Config{
		ServerConfig:  serverCfg,
		StorageConfig: storageCfg,
	}, nil
}

// ParseFlags parses command line arguments and stores them
func (c *Config) ParseFlags() {
	flag.StringVar(&c.ServerConfig.ServerAddress, "a", ":8080", "Server address")
	flag.StringVar(&c.ServerConfig.BaseURL, "b", "http://localhost:8080", "Base url")
	flag.StringVar(&c.StorageConfig.FileStoragePath, "f", "storage/infile/url_storage.json", "File storage path")
	flag.Parse()
}
