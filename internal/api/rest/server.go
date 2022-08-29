// Package rest provides functionality for initializing a server for the shortening URL service.
package rest

import (
	"context"
	"expvar"
	_ "expvar"
	"net/http"
	"time"

	"github.com/danilovkiri/dk_go_url_shortener/internal/api/rest/handlers"
	"github.com/danilovkiri/dk_go_url_shortener/internal/api/rest/middleware"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/secretary/v1"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/shortener/v1"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1"
	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
)

var (
	serverStart = time.Now()
)

// uptime returns time in seconds since the server start-up.
func uptime() interface{} {
	return int64(time.Since(serverStart).Seconds())
}

// InitServer returns a http.Server object ready to be listening and serving .
func InitServer(ctx context.Context, cfg *config.Config, storage storage.URLStorage) (server *http.Server, err error) {
	shortenerService, err := shortener.InitShortener(storage)
	if err != nil {
		return nil, err
	}
	urlHandler, err := handlers.InitURLHandler(shortenerService, cfg.ServerConfig)
	if err != nil {
		return nil, err
	}
	secretaryService := secretary.NewSecretaryService(cfg.SecretConfig)
	cookieHandler, err := middleware.NewCookieHandler(secretaryService, cfg.SecretConfig)
	if err != nil {
		return nil, err
	}
	r := chi.NewRouter()
	r.Use(cookieHandler.CookieHandle)
	r.Use(middleware.CompressHandle)
	r.Use(middleware.DecompressHandle)
	r.Post("/", urlHandler.HandlePostURL())
	r.Post("/api/shorten", urlHandler.JSONHandlePostURL())
	r.Post("/api/shorten/batch", urlHandler.JSONHandlePostURLBatch())
	r.Get("/{urlID}", urlHandler.HandleGetURL())
	r.Get("/api/user/urls", urlHandler.HandleGetURLsByUserID())
	r.Delete("/api/user/urls", urlHandler.HandleDeleteURLBatch())
	r.Get("/ping", urlHandler.HandlePingDB())
	r.Mount("/debug", chiMiddleware.Profiler()) // see https://github.com/go-chi/chi/blob/master/middleware/profiler.go
	expvar.Publish("system.uptime", expvar.Func(uptime))

	srv := &http.Server{
		Addr: cfg.ServerConfig.ServerAddress,
		//Handler:      http.TimeoutHandler(r, 500*time.Millisecond, "Timeout reached"),
		Handler:      r,
		IdleTimeout:  60 * time.Second,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	return srv, nil
}
