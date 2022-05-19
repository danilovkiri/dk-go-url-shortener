package main

import (
	"context"
	"github.com/danilovkiri/dk_go_url_shortener/internal/api/rest"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/inpsql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// add a waiting group
	wg := &sync.WaitGroup{}
	// set number of wg members to 1 (this will be the persistStorage goroutine)
	wg.Add(1)
	// get configuration
	cfg, err := config.NewDefaultConfiguration()
	if err != nil {
		log.Fatal(err)
	}
	cfg.ParseFlags()
	// initialize (or retrieve if present) storage
	storage, err := inpsql.InitStorage(ctx, wg, cfg.StorageConfig)
	if err != nil {
		log.Fatal(err)
	}
	// initialize server
	server, err := rest.InitServer(ctx, cfg, storage)
	if err != nil {
		log.Fatal(err)
	}
	// set a listener for os.Signal
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-done
		log.Print("Server shutdown attempted")
		ctxTO, cancelTO := context.WithTimeout(ctx, 5*time.Second)
		defer cancelTO()
		if err := server.Shutdown(ctxTO); err != nil {
			log.Fatal("Server shutdown failed:", err)
		}
		log.Print("Server shutdown succeeded")
		cancel()
	}()
	// start up the server
	log.Print("Server start attempted")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	// wait for goroutine in InitStorage to finish before exiting
	wg.Wait()
}
