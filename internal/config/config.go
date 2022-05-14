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
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
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

// isFlagPassed checks whether the flag was set in CLI
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// ParseFlags parses command line arguments and stores them
func (c *Config) ParseFlags() {
	a := flag.String("a", ":8080", "Server address")
	b := flag.String("b", "http://localhost:8080", "Base url")
	f := flag.String("f", "url_storage.json", "File storage path")
	flag.Parse()
	// priority: flag -> env -> default flag
	// note that env parsing precedes flag parsing
	if isFlagPassed("a") || c.ServerConfig.ServerAddress == "" {
		c.ServerConfig.ServerAddress = *a
	}
	if isFlagPassed("b") || c.ServerConfig.BaseURL == "" {
		c.ServerConfig.BaseURL = *b
	}
	if isFlagPassed("f") || c.StorageConfig.FileStoragePath == "" {
		c.StorageConfig.FileStoragePath = *f
	}
}
