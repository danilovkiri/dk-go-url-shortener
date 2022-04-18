package handlers

import (
	"context"
	"fmt"
	"github.com/danilovkiri/dk_go_url_shortener/service/shortener"
	"github.com/go-chi/chi"
	"io/ioutil"
	"log"
	"net/http"
)

type URLHandler struct {
	svc shortener.Processor
}

func InitURLHandler(svc shortener.Processor) (*URLHandler, error) {
	if svc == nil {
		return nil, fmt.Errorf("nil Shortener Service was passed to service URL Handler initializer")
	}
	return &URLHandler{svc: svc}, nil
}

func (h *URLHandler) HandleGetURL(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sURL := chi.URLParam(r, "urlID")
		log.Println("GET request detected for", sURL)
		URL, err := h.svc.Decode(ctx, sURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Location", URL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func (h *URLHandler) HandlePostURL(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		log.Println("POST request detected", string(b))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := h.svc.Encode(ctx, string(b))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("http://" + r.Host + "/" + id))
	}
}
