package main

import (
	"context"
	"github.com/danilovkiri/dk_go_url_shortener/internal/api/rest"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1/infile"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1/inpsql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	// make a top-level file logger for logging critical errors
	flog, err := os.OpenFile(`server.log`, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer flog.Close()
	mainlog := log.New(flog, `main `, log.LstdFlags|log.Lshortfile)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// add a waiting group
	wg := &sync.WaitGroup{}
	// set number of wg members to 1 (this will be the persistStorage goroutine)
	wg.Add(1)
	// get configuration
	cfg, err := config.NewDefaultConfiguration()
	if err != nil {
		mainlog.Fatal(err)
	}
	cfg.ParseFlags()
	// initialize (or retrieve if present) storage, switch between "infile" and "inpsql" modules
	var errInit error
	var storageInit storage.URLStorage
	switch cfg.StorageConfig.DatabaseDSN {
	case "":
		storageInit, errInit = infile.InitStorage(ctx, wg, cfg.StorageConfig)
	default:
		storageInit, errInit = inpsql.InitStorage(ctx, wg, cfg.StorageConfig)
	}

	if err != nil {
		mainlog.Fatal(errInit)
	}
	// initialize server
	server, err := rest.InitServer(ctx, cfg, storageInit)
	if err != nil {
		mainlog.Fatal(err)
	}
	// set a listener for os.Signal
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-done
		mainlog.Print("Server shutdown attempted")
		ctxTO, cancelTO := context.WithTimeout(ctx, 5*time.Second)
		defer cancelTO()
		if err := server.Shutdown(ctxTO); err != nil {
			mainlog.Fatal("Server shutdown failed:", err)
		}
		cancel()
	}()
	// start up the server
	mainlog.Print("Server start attempted")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		mainlog.Fatal(err)
	}
	// wait for goroutine in InitStorage to finish before exiting
	wg.Wait()
	mainlog.Print("Server shutdown succeeded")
}
