package main

import (
	"context"
	"github.com/danilovkiri/dk_go_url_shortener/internal/api/rest"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/infile"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()
	// get configuration
	cfg, err := config.NewDefaultConfiguration()
	if err != nil {
		log.Fatal(err)
	}
	cfg.ParseFlags()
	// initialize (or retrieve if present) storage
	storage, err := infile.InitStorage(cfg.StorageConfig)
	if err != nil {
		log.Fatal(err)
	}
	// initialize server
	server, err := rest.InitServer(ctx, cfg, storage)
	if err != nil {
		log.Fatal(err)
	}
	// start up the server
	go func() {
		log.Print("Server start attempted")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	// graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done
	log.Print("Server shutdown attempted")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server shutdown failed:", err)
	}
	log.Print("Server shutdown succeeded")
	// defer dumping tmpfs storage into file upon server shutdown attempt
	defer func() {
		cancel()
		err := storage.PersistStorage()
		if err != nil {
			log.Fatal(err)
		}
	}()
}
