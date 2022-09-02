// Package config provides types for handling configuration parameters.
package config

import (
	"bufio"
	"encoding/json"
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
	"os"
)

// Config handles server-related constants and parameters.
type Config struct {
	ServerConfig  *ServerConfig
	StorageConfig *StorageConfig
	SecretConfig  *SecretConfig
}

// AppConfigJSON handles parameters passed in a configuration JSON file.
type AppConfigJSON struct {
	ServerAddress   string `json:"server_address,omitempty"`
	BaseURL         string `json:"base_url,omitempty"`
	FileStoragePath string `json:"file_storage_path,omitempty"`
	DatabaseDSN     string `json:"database_dsn,omitempty"`
	EnableHTTPS     bool   `json:"enable_https,omitempty"`
}

// ServerConfig defines default server-relates constants and parameters and overwrites them with environment variables.
type ServerConfig struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	EnableHTTPS   bool   `env:"ENABLE_HTTPS"`
}

// StorageConfig retrieves file storage-related parameters from environment.
type StorageConfig struct {
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

// SecretConfig retrieves a secret user key for hashing.
type SecretConfig struct {
	UserKey string `env:"USER_KEY" envDefault:"jds__63h3_7ds"`
}

// NewStorageConfig sets up a storage configuration.
func NewStorageConfig() *StorageConfig {
	cfg := StorageConfig{}
	_ = env.Parse(&cfg)
	return &cfg
}

// NewServerConfig sets up a server configuration.
func NewServerConfig() *ServerConfig {
	cfg := ServerConfig{}
	_ = env.Parse(&cfg)
	return &cfg
}

// NewSecretConfig sets up a secret configuration.
func NewSecretConfig() *SecretConfig {
	cfg := SecretConfig{}
	_ = env.Parse(&cfg)
	return &cfg
}

// NewDefaultConfiguration sets up a total configuration.
func NewDefaultConfiguration() *Config {
	serverCfg := NewServerConfig()
	storageCfg := NewStorageConfig()
	secretConfig := NewSecretConfig()
	return &Config{
		ServerConfig:  serverCfg,
		StorageConfig: storageCfg,
		SecretConfig:  secretConfig,
	}
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

// parseAppConfig parses JSON configuration file
func (cfg *Config) parseAppConfig(path string) (*AppConfigJSON, error) {
	configFile, err := os.Open(path)
	defer func(configFile *os.File) {
		err1 := configFile.Close()
		if err1 != nil {
			log.Fatalf("Could not close file: %s", err1)
		}
	}(configFile)
	reader := bufio.NewReader(configFile)
	if err != nil {
		return nil, err
	}
	fi, _ := configFile.Stat()
	var appConfigBytes = make([]byte, fi.Size())
	_, _ = reader.Read(appConfigBytes)
	var appConfig AppConfigJSON
	err = json.Unmarshal(appConfigBytes, &appConfig)
	if err != nil {
		return nil, err
	}
	return &appConfig, nil
}

// redefineConfig implements prioritization logic over app parameters
func (cfg *Config) redefineConfig(a, b, f, d, c *string, s *bool) {
	// priority: flag -> env -> json config -> default flag
	// note that env parsing precedes flag parsing and JSON parsing
	if *c != "" {
		jsonConfig, err := cfg.parseAppConfig(*c)
		if err != nil {
			log.Fatalf("%s configuration file could not be processed: %s", *c, err)
		}
		if cfg.ServerConfig.ServerAddress == "" && jsonConfig.ServerAddress != "" {
			cfg.ServerConfig.ServerAddress = jsonConfig.ServerAddress
		}
		if cfg.ServerConfig.BaseURL == "" && jsonConfig.BaseURL != "" {
			cfg.ServerConfig.BaseURL = jsonConfig.BaseURL
		}
		if cfg.ServerConfig.ServerAddress == "" && jsonConfig.ServerAddress != "" {
			cfg.ServerConfig.ServerAddress = jsonConfig.ServerAddress
		}
		if cfg.StorageConfig.FileStoragePath == "" && jsonConfig.FileStoragePath != "" {
			cfg.StorageConfig.FileStoragePath = jsonConfig.FileStoragePath
		}
		if cfg.StorageConfig.DatabaseDSN == "" && jsonConfig.DatabaseDSN != "" {
			cfg.StorageConfig.DatabaseDSN = jsonConfig.DatabaseDSN
		}
		if !cfg.ServerConfig.EnableHTTPS && jsonConfig.EnableHTTPS {
			cfg.ServerConfig.EnableHTTPS = jsonConfig.EnableHTTPS
		}
	}

	if isFlagPassed("a") || cfg.ServerConfig.ServerAddress == "" {
		cfg.ServerConfig.ServerAddress = *a
	}
	if isFlagPassed("b") || cfg.ServerConfig.BaseURL == "" {
		cfg.ServerConfig.BaseURL = *b
	}
	if isFlagPassed("f") || cfg.StorageConfig.FileStoragePath == "" {
		cfg.StorageConfig.FileStoragePath = *f
	}
	if isFlagPassed("d") || cfg.StorageConfig.DatabaseDSN == "" {
		cfg.StorageConfig.DatabaseDSN = *d
	}
	if isFlagPassed("s") || !cfg.ServerConfig.EnableHTTPS {
		cfg.ServerConfig.EnableHTTPS = *s
	}
}

// ParseFlags parses command line arguments and stores them
func (cfg *Config) ParseFlags() {
	a := flag.String("a", ":8080", "Server address")
	b := flag.String("b", "http://localhost:8080", "Base url")
	c := flag.String("q", os.Getenv("CONFIG"), "Configuration file path")
	// DatabaseDSN scheme: "postgres://username:password@localhost:5432/database_name"
	d := flag.String("d", "", "PSQL DB connection")
	f := flag.String("f", "url_storage.json", "File storage path")
	s := flag.Bool("s", false, "Use HTTPS connection")
	flag.Parse()
	cfg.redefineConfig(a, b, f, d, c, s)

}
