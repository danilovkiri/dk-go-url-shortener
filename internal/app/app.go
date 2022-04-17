package app

import (
	"fmt"
	"github.com/danilovkiri/dk_go_url_shortener/config"
	"github.com/danilovkiri/dk_go_url_shortener/service/shortener"
	"github.com/danilovkiri/dk_go_url_shortener/storage/inmemory"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"time"
)

type App struct {
	Config *config.ServerConfig
	Db     *inmemory.Database
}

func (app *App) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.mainHandler)
	srv := &http.Server{
		Addr:         app.Config.Host + ":" + app.Config.Port,
		Handler:      mux,
		IdleTimeout:  10 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	err := srv.ListenAndServe()
	log.Fatal(err)
}

func (app *App) mainHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sUrl := path.Base(r.URL.Path)
		url, err := app.Db.GetFullUrl(inmemory.ShortUrl(sUrl))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Location", string(url))
		w.WriteHeader(http.StatusTemporaryRedirect)
	case http.MethodPost:
		b, _ := ioutil.ReadAll(r.Body)
		url := inmemory.Url(b)
		sUrl, err := shortener.GenereteShortString(string(url))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = app.Db.SaveShortUrl(inmemory.ShortUrl(sUrl), url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(fmt.Sprintf("http://%s/%s", r.Host, sUrl)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Only GET and POST requests are allowed", http.StatusMethodNotAllowed)
	}
}
