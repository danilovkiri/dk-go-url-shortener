// Package handlers provides http.HandlerFunc handler functions to be used for endpoints.
package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/shortener"
	storageErrors "github.com/danilovkiri/dk_go_url_shortener/internal/storage/errors"
	"github.com/go-chi/chi"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// URLHandler defines data structure handling and provides support for adding new implementations.
type URLHandler struct {
	processor shortener.Processor
}

// InitURLHandler initializes a URLHandler object and sets its attributes.
func InitURLHandler(processor shortener.Processor) (*URLHandler, error) {
	if processor == nil {
		return nil, fmt.Errorf("nil Shortener Service was passed to service URL Handler initializer")
	}
	return &URLHandler{processor: processor}, nil
}

// HandleGetURL provides functionality for handling GET requests.
func (h *URLHandler) HandleGetURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// set context timeout to 500 ms for timing DB operations
		ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
		defer cancel()

		// execute main body
		sURL := chi.URLParam(r, "urlID")
		log.Println("GET request detected for", sURL)
		URL, err := h.processor.Decode(ctx, sURL)
		if err != nil {
			if errors.Is(err, storageErrors.ContextTimeoutExceededError{}) {
				log.Println("HandleGetURL:", err)
				w.WriteHeader(http.StatusGatewayTimeout)
			} else {
				log.Println("HandleGetURL:", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}
		log.Println("HandleGetURL: retrieved URL", URL)
		w.Header().Set("Location", URL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

// HandlePostURL provides functionality for handling POST requests.
func (h *URLHandler) HandlePostURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// set context timeout to 500 ms for timing DB operations
		ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
		defer cancel()

		// execute main body
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println("POST request detected for", string(b))
		id, err := h.processor.Encode(ctx, string(b))
		if err != nil {
			if errors.Is(err, storageErrors.ContextTimeoutExceededError{}) {
				log.Println("HandlePostURL:", err)
				w.WriteHeader(http.StatusGatewayTimeout)
			} else {
				log.Println("HandlePostURL:", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}
		log.Println("HandlePostURL: stored", string(b), "as", id)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("http://" + r.Host + "/" + id))
	}
}
