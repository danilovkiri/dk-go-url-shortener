package rest

import (
	"context"
	"github.com/danilovkiri/dk_go_url_shortener/api/rest/handlers"
	"github.com/danilovkiri/dk_go_url_shortener/service/shortener/v1"
	"github.com/danilovkiri/dk_go_url_shortener/storage/inmemory"
	"github.com/go-chi/chi"
	"net/http"
	"time"
)

const (
	host = "localhost"
	port = "8080"
)

func InitServer(ctx context.Context) (server *http.Server, err error) {
	storage := inmemory.InitStorage()
	shortenerService, err := shortener.InitShortener(storage)
	if err != nil {
		return nil, err
	}
	urlHandler, err := handlers.InitURLHandler(shortenerService)
	if err != nil {
		return nil, err
	}
	r := chi.NewRouter()
	r.Post("/", urlHandler.HandlePostURL())
	r.Get("/{urlID}", urlHandler.HandleGetURL())
	srv := &http.Server{
		Addr: host + ":" + port,
		//Handler:      http.TimeoutHandler(r, 500*time.Millisecond, "Timeout reached"),
		Handler:      r,
		IdleTimeout:  10 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return srv, nil
}
