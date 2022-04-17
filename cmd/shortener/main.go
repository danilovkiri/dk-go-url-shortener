package main

import (
	"github.com/danilovkiri/dk_go_url_shortener/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/app"
	"github.com/danilovkiri/dk_go_url_shortener/storage/inmemory"
)

const (
	host = "localhost"
	port = "8080"
)

func main() {
	// setting up configuration
	configuration := config.ServerConfig{
		Host: host,
		Port: port,
	}
	// setting up url mapping storage
	db := inmemory.InitStorage()
	// starting up service
	application := app.App{Config: &configuration, Db: db}
	application.Start()
}
