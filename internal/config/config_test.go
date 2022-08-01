package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewStorageConfig(t *testing.T) {
	_ = os.Setenv("FILE_STORAGE_PATH", "some_file")
	_ = os.Setenv("DATABASE_DSN", "some_dsn")
	cfg, err := NewStorageConfig()
	expCfg := StorageConfig{
		"some_file",
		"some_dsn",
	}
	assert.Equal(t, nil, err)
	assert.Equal(t, &expCfg, cfg)
}

func TestNewServerConfig(t *testing.T) {
	_ = os.Setenv("SERVER_ADDRESS", "some_server_address")
	_ = os.Setenv("BASE_URL", "some_base_url")
	cfg, err := NewServerConfig()
	expCfg := ServerConfig{
		"some_server_address",
		"some_base_url",
	}
	assert.Equal(t, nil, err)
	assert.Equal(t, &expCfg, cfg)
}

func TestNewSecretConfig(t *testing.T) {
	_ = os.Setenv("USER_KEY", "some_user_key")
	cfg, err := NewSecretConfig()
	expCfg := SecretConfig{
		"some_user_key",
	}
	assert.Equal(t, nil, err)
	assert.Equal(t, &expCfg, cfg)
}

func TestNewDefaultConfiguration(t *testing.T) {
	_ = os.Setenv("FILE_STORAGE_PATH", "some_file")
	_ = os.Setenv("DATABASE_DSN", "some_dsn")
	_ = os.Setenv("SERVER_ADDRESS", "some_server_address")
	_ = os.Setenv("BASE_URL", "some_base_url")
	_ = os.Setenv("USER_KEY", "some_user_key")
	cfg, err := NewDefaultConfiguration()
	expCfg := Config{
		&ServerConfig{
			"some_server_address",
			"some_base_url",
		},
		&StorageConfig{
			"some_file",
			"some_dsn",
		},
		&SecretConfig{
			"some_user_key",
		},
	}
	assert.Equal(t, nil, err)
	assert.Equal(t, &expCfg, cfg)
}
