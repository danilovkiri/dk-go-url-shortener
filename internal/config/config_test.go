package config

import (
	"encoding/json"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests

func TestNewStorageConfig(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("FILE_STORAGE_PATH", "some_file")
	_ = os.Setenv("DATABASE_DSN", "some_dsn")
	cfg := NewStorageConfig()
	expCfg := StorageConfig{
		"some_file",
		"some_dsn",
	}
	assert.Equal(t, &expCfg, cfg)
}

func TestNewServerConfig(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("SERVER_ADDRESS", "some_server_address")
	_ = os.Setenv("BASE_URL", "some_base_url")
	_ = os.Setenv("ENABLE_HTTPS", "false")
	cfg := NewServerConfig()
	expCfg := ServerConfig{
		"some_server_address",
		"some_base_url",
		false,
	}
	assert.Equal(t, &expCfg, cfg)
}

func TestNewSecretConfig(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("USER_KEY", "some_user_key")
	cfg := NewSecretConfig()
	expCfg := SecretConfig{
		"some_user_key",
	}
	assert.Equal(t, &expCfg, cfg)
}

func TestNewDefaultConfiguration(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("FILE_STORAGE_PATH", "some_file")
	_ = os.Setenv("DATABASE_DSN", "some_dsn")
	_ = os.Setenv("SERVER_ADDRESS", "some_server_address")
	_ = os.Setenv("BASE_URL", "some_base_url")
	_ = os.Setenv("USER_KEY", "some_user_key")
	_ = os.Setenv("ENABLE_HTTPS", "false")
	cfg := NewDefaultConfiguration()
	expCfg := Config{
		&ServerConfig{
			"some_server_address",
			"some_base_url",
			false,
		},
		&StorageConfig{
			"some_file",
			"some_dsn",
		},
		&SecretConfig{
			"some_user_key",
		},
	}
	assert.Equal(t, &expCfg, cfg)
}

func TestConfig_ParseFlags(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("USER_KEY", "some_user_key")
	cfg := NewDefaultConfiguration()
	os.Args = []string{"test", "-a", ":8080", "-q", "config_test.json", "-f", "url_storage.json", "-d", "postgres://username:password@localhost:5432/database_name", "-s", "true"}
	cfg.ParseFlags()
	expCfg := Config{
		&ServerConfig{
			":8080",
			"json_base_url",
			true,
		},
		&StorageConfig{
			"url_storage.json",
			"postgres://username:password@localhost:5432/database_name",
		},
		&SecretConfig{
			"some_user_key",
		},
	}
	assert.Equal(t, &expCfg, cfg)
}

func TestConfig_parseAppConfigPathError(t *testing.T) {
	os.Clearenv()
	cfg := NewDefaultConfiguration()
	_, err := cfg.parseAppConfig("nonexistent_file.json")
	var error *fs.PathError
	assert.ErrorAs(t, err, &error)
}

func TestConfig_parseAppConfigUnmarshallError(t *testing.T) {
	os.Clearenv()
	cfg := NewDefaultConfiguration()
	_, err := cfg.parseAppConfig("config_test_bad.json")
	var error *json.SyntaxError
	assert.ErrorAs(t, err, &error)
}

// Benchmarks

func BenchmarkNewDefaultConfiguration(b *testing.B) {
	os.Clearenv()
	_ = os.Setenv("FILE_STORAGE_PATH", "some_file")
	_ = os.Setenv("DATABASE_DSN", "some_dsn")
	_ = os.Setenv("SERVER_ADDRESS", "some_server_address")
	_ = os.Setenv("BASE_URL", "some_base_url")
	_ = os.Setenv("USER_KEY", "some_user_key")
	_ = os.Setenv("ENABLE_HTTPS", "false")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewDefaultConfiguration()
	}
}
