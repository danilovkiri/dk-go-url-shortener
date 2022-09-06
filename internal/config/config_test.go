package config

import (
	"io/fs"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests

func TestNewDefaultConfiguration(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("FILE_STORAGE_PATH", "some_file")
	_ = os.Setenv("DATABASE_DSN", "some_dsn")
	_ = os.Setenv("SERVER_ADDRESS", "some_server_address")
	_ = os.Setenv("BASE_URL", "some_base_url")
	_ = os.Setenv("USER_KEY", "some_user_key")
	_ = os.Setenv("ENABLE_HTTPS", "false")
	cfg := NewDefaultConfiguration()
	var a = ""
	var b = ""
	var f = ""
	var d = ""
	var c = ""
	var s = false
	err := cfg.assignValues(&a, &b, &f, &d, &c, &s)
	if err != nil {
		log.Fatal(err)
	}
	expCfg := Config{
		ServerAddress:   "some_server_address",
		BaseURL:         "some_base_url",
		EnableHTTPS:     false,
		FileStoragePath: "some_file",
		DatabaseDSN:     "some_dsn",
		UserKey:         "some_user_key",
	}
	assert.Equal(t, &expCfg, cfg)
}

func TestConfig_ParseFlags(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("USER_KEY", "some_user_key")
	cfg := NewDefaultConfiguration()
	os.Args = []string{"test", "-a", ":8080", "-c", "config_test.json", "-f", "url_storage.json", "-d", "postgres://username:password@localhost:5432/database_name", "-s", "true"}
	err := cfg.Parse()
	if err != nil {
		log.Fatal(err)
	}
	expCfg := Config{
		ServerAddress:   ":8080",
		BaseURL:         "json_base_url",
		EnableHTTPS:     true,
		FileStoragePath: "url_storage.json",
		DatabaseDSN:     "postgres://username:password@localhost:5432/database_name",
		UserKey:         "some_user_key",
	}
	assert.Equal(t, &expCfg, cfg)
}

func TestConfig_parseAppConfigPathError(t *testing.T) {
	os.Clearenv()
	cfg := NewDefaultConfiguration()
	var a = ""
	var b = ""
	var f = ""
	var d = ""
	var c = "nonexistent_file.json"
	var s = false
	err := cfg.assignValues(&a, &b, &f, &d, &c, &s)
	var error *fs.PathError
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
	var a = ""
	var bb = ""
	var f = ""
	var d = ""
	var c = ""
	var s = false
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := NewDefaultConfiguration()
		err := cfg.assignValues(&a, &bb, &f, &d, &c, &s)
		if err != nil {
			log.Fatal(err)
		}

	}
}
