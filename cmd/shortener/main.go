package main

import (
	"github.com/danilovkiri/dk_go_url_shortener/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/app"
	"github.com/danilovkiri/dk_go_url_shortener/service/shortener"
	"github.com/danilovkiri/dk_go_url_shortener/storage/inmemory"
	"log"
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
	short, err := shortener.InitShortener()
	if err != nil {
		log.Fatal(err.Error())
	}
	// starting up service
	application := app.App{Config: &configuration, Db: db, Short: short}
	application.Start()
}
