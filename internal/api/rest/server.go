// Package rest provides functionality for initializing a server for the shortening URL service.
package rest

import (
	"context"
	"crypto/tls"
	"expvar"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"golang.org/x/crypto/acme/autocert"

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
	urlHandler, err := handlers.InitURLHandler(shortenerService, cfg)
	if err != nil {
		return nil, err
	}
	secretaryService := secretary.NewSecretaryService(cfg)
	cookieHandler, err := middleware.NewCookieHandler(secretaryService, cfg)
	if err != nil {
		return nil, err
	}
	trustedNetHandler := middleware.NewTrustedNetHandler(cfg)
	r := chi.NewRouter()
	r.Use(cookieHandler.CookieHandle)
	r.Use(middleware.CompressHandle)
	r.Use(middleware.DecompressHandle)
	trustedGroup := r.Group(nil)
	trustedGroup.Use(trustedNetHandler.TrustedNetworkHandler)
	trustedGroup.Get("/api/internal/stats", urlHandler.HandleGetStats())
	mainGroup := r.Group(nil)
	mainGroup.Post("/", urlHandler.HandlePostURL())
	mainGroup.Post("/api/shorten", urlHandler.JSONHandlePostURL())
	mainGroup.Post("/api/shorten/batch", urlHandler.JSONHandlePostURLBatch())
	mainGroup.Get("/{urlID}", urlHandler.HandleGetURL())
	mainGroup.Get("/api/user/urls", urlHandler.HandleGetURLsByUserID())
	mainGroup.Delete("/api/user/urls", urlHandler.HandleDeleteURLBatch())
	mainGroup.Get("/ping", urlHandler.HandlePingDB())
	mainGroup.Mount("/debug", chiMiddleware.Profiler()) // see https://github.com/go-chi/chi/blob/master/middleware/profiler.go
	expvar.Publish("system.uptime", expvar.Func(uptime))

	var srv *http.Server
	if !cfg.EnableHTTPS {
		srv = &http.Server{
			Addr:         cfg.ServerAddress,
			Handler:      r,
			IdleTimeout:  60 * time.Second,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
		}
	} else {
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("localhost"), // to be changed to real value
			Cache:      autocert.DirCache("../../certs"),
		}
		tlsConfig := certManager.TLSConfig()
		tlsConfig.GetCertificate = getSelfSignedOrLetsEncryptCert(&certManager)
		srv = &http.Server{
			Addr:         cfg.ServerAddress,
			Handler:      r,
			IdleTimeout:  60 * time.Second,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
			TLSConfig:    tlsConfig,
		}
	}

	return srv, nil
}

// getSelfSignedOrLetsEncryptCert implements fallback for certificate usage in case autocert fails
func getSelfSignedOrLetsEncryptCert(certManager *autocert.Manager) func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		dirCache, ok := certManager.Cache.(autocert.DirCache)
		if !ok {
			dirCache = "../../certs"
		}
		keyFile := filepath.Join(string(dirCache), hello.ServerName+".key")
		crtFile := filepath.Join(string(dirCache), hello.ServerName+".crt")
		certificate, err := tls.LoadX509KeyPair(crtFile, keyFile)
		if err != nil {
			log.Printf("%s\nFalling back to Letsencrypt\n", err)
			return certManager.GetCertificate(hello)
		}
		log.Println("Loading self-signed certificate.")
		return &certificate, err
	}
}
