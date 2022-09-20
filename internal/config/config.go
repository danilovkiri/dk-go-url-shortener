// Package config provides types for handling configuration parameters.
package config

import (
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config handles all constants and parameters.
type Config struct {
	ServerAddress   string `json:"server_address" env:"SERVER_ADDRESS"`
	BaseURL         string `json:"base_url" env:"BASE_URL"`
	EnableHTTPS     bool   `json:"enable_https" env:"ENABLE_HTTPS"`
	UseGRPC         bool   `json:"use_grpc" env:"USE_GRPC"`
	FileStoragePath string `json:"file_storage_path" env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `json:"database_dsn" env:"DATABASE_DSN"`
	UserKey         string `env:"USER_KEY" env-default:"jds__63h3_7ds"`
	TrustedSubnet   string `json:"trusted_subnet" env:"TRUSTED_SUBNET"`
	AuthKey         string `env:"AUTH_KEY" env-default:"user"`
}

// NewDefaultConfiguration initializes a configuration struct.
func NewDefaultConfiguration() *Config {
	var cfg Config
	return &cfg
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

func (cfg *Config) assignValues(a, b, f, d, c, t *string, s, g *bool) error {
	// priority: flag -> env -> json config -> default flag
	var err error
	if *c != "" {
		err = cleanenv.ReadConfig(*c, cfg)
	} else {
		err = cleanenv.ReadEnv(cfg)
	}
	// return err here to stop code execution
	if err != nil {
		return err
	}
	if isFlagPassed("a") || cfg.ServerAddress == "" {
		cfg.ServerAddress = *a
	}
	if isFlagPassed("b") || cfg.BaseURL == "" {
		cfg.BaseURL = *b
	}
	if isFlagPassed("f") || cfg.FileStoragePath == "" {
		cfg.FileStoragePath = *f
	}
	if isFlagPassed("d") || cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = *d
	}
	if isFlagPassed("t") || cfg.TrustedSubnet == "" {
		cfg.TrustedSubnet = *t
	}
	if isFlagPassed("s") || !cfg.EnableHTTPS {
		cfg.EnableHTTPS = *s
	}
	if isFlagPassed("g") || !cfg.UseGRPC {
		cfg.UseGRPC = *g
	}
	return nil
}

// Parse parses command line arguments and environment and stores them
func (cfg *Config) Parse() error {
	a := flag.String("a", ":8080", "Server address")
	b := flag.String("b", "http://localhost:8080", "Base url")
	c := flag.String("c", os.Getenv("CONFIG"), "Configuration file path")
	// DatabaseDSN scheme: "postgres://username:password@localhost:5432/database_name"
	d := flag.String("d", "", "PSQL DB connection")
	f := flag.String("f", "url_storage.json", "File storage path")
	s := flag.Bool("s", false, "Use HTTPS connection")
	t := flag.String("t", "", "Trusted subnet")
	g := flag.Bool("g", false, "Use GRPC protocol")
	flag.Parse()
	err := cfg.assignValues(a, b, f, d, c, t, s, g)
	return err
}
