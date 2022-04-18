package handlers

import (
	"context"
	"fmt"
	"github.com/danilovkiri/dk_go_url_shortener/service/shortener"
	"github.com/go-chi/chi"
	"io/ioutil"
	"log"
	"net/http"
	"time"
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

func (h *URLHandler) HandleGetURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
		defer cancel()
		r = r.WithContext(ctx)
		handleGetDone := make(chan string)
		sURL := chi.URLParam(r, "urlID")
		log.Println("GET request detected for", sURL)
		go func() {
			time.Sleep(time.Second)
			URL, err := h.svc.Decode(ctx, sURL)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
			}
			handleGetDone <- URL
		}()

		select {
		case <-ctx.Done():
			log.Println("HandleGetURL:", ctx.Err())
			w.WriteHeader(http.StatusGatewayTimeout)
		case URL := <-handleGetDone:
			log.Println("HandleGetURL: retrieved URL", URL)
			w.Header().Set("Location", URL)
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	}
}

func (h *URLHandler) HandlePostURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Millisecond)
		defer cancel()
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
