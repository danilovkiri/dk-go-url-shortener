// Package handlers provides http.HandlerFunc handler functions to be used for endpoints.
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

// URLHandler defines data structure handling and provides support for adding new implementations.
type URLHandler struct {
	svc shortener.Processor
}

// InitURLHandler initializes a URLHandler object and sets its attributes.
func InitURLHandler(svc shortener.Processor) (*URLHandler, error) {
	if svc == nil {
		return nil, fmt.Errorf("nil Shortener Service was passed to service URL Handler initializer")
	}
	return &URLHandler{svc: svc}, nil
}

// HandleGetURL provides functionality for handling GET requests.
func (h *URLHandler) HandleGetURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
		defer cancel()
		r = r.WithContext(ctx)
		handleGetDone := make(chan string)
		handleGetError := make(chan string)
		sURL := chi.URLParam(r, "urlID")
		log.Println("GET request detected for", sURL)
		go func() {
			URL, err := h.svc.Decode(ctx, sURL)
			if err != nil {
				handleGetError <- err.Error()
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
		case errString := <-handleGetError:
			log.Println("HandleGetURL:", errString)
			http.Error(w, errString, http.StatusNotFound)
		}
	}
}

// HandlePostURL provides functionality for handling POST requests.
func (h *URLHandler) HandlePostURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2000*time.Millisecond)
		defer cancel()
		r = r.WithContext(ctx)
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		handlePostDone := make(chan string)
		handlePostError := make(chan string)
		log.Println("POST request detected for", string(b))
		go func() {
			id, err := h.svc.Encode(ctx, string(b))
			if err != nil {
				handlePostError <- err.Error()
				return
			}
			handlePostDone <- id
		}()

		select {
		case <-ctx.Done():
			log.Println("HandlePostURL:", ctx.Err())
			w.WriteHeader(http.StatusGatewayTimeout)
		case id := <-handlePostDone:
			log.Println("HandlePostURL: stored", string(b), "as", id)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("http://" + r.Host + "/" + id))
		case errString := <-handlePostError:
			log.Println("HandlePostURL:", errString)
			http.Error(w, errString, http.StatusBadRequest)
		}
	}
}
