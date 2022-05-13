// Package handlers provides http.HandlerFunc handler functions to be used for endpoints.
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/danilovkiri/dk_go_url_shortener/internal/api/rest/modeldto"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/shortener"
	storageErrors "github.com/danilovkiri/dk_go_url_shortener/internal/storage/errors"
	"github.com/go-chi/chi"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

// URLHandler defines data structure handling and provides support for adding new implementations.
type URLHandler struct {
	processor    shortener.Processor
	serverConfig *config.ServerConfig
}

// InitURLHandler initializes a URLHandler object and sets its attributes.
func InitURLHandler(processor shortener.Processor, serverConfig *config.ServerConfig) (*URLHandler, error) {
	if processor == nil {
		return nil, fmt.Errorf("nil Shortener Service was passed to service URL Handler initializer")
	}
	return &URLHandler{processor: processor, serverConfig: serverConfig}, nil
}

// HandleGetURL provides client with a redirect to the original URL accessed by shortened URL.
func (h *URLHandler) HandleGetURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// set context timeout to 500 ms for timing DB operations
		ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
		defer cancel()
		// retrieve sURL from query
		sURL := chi.URLParam(r, "urlID")
		log.Println("GET request detected for", sURL)
		// decode sURL into the original URL
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
		// set and send response
		w.Header().Set("Location", URL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

// HandlePostURL stores the original URL with its shortened version.
func (h *URLHandler) HandlePostURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// set context timeout to 500 ms for timing DB operations
		ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
		defer cancel()
		// read POST body
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println("POST request detected for", string(b))
		// encode URL into sURL and store
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
		// set and send response
		w.WriteHeader(http.StatusCreated)
		u, err := url.Parse(h.serverConfig.BaseURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		u.Path = id
		w.Write([]byte(u.String()))
	}
}

// JSONHandlePostURL accepts JSON as {"url":"<some_url>"} and provides client with JSON as {"result":"<shorten_url>"}.
func (h *URLHandler) JSONHandlePostURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// set context timeout to 500 ms for timing DB operations
		ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
		defer cancel()
		// check for POST body content type compliance
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
		}
		// read POST body
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// deserialize JSON into struct
		var post modeldto.RequestURL
		err = json.Unmarshal(b, &post)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println("JSON POST request detected for", post.URL)
		// encode URL into sURL and store them
		id, err := h.processor.Encode(ctx, post.URL)
		if err != nil {
			if errors.Is(err, storageErrors.ContextTimeoutExceededError{}) {
				log.Println("JSONHandlePostURL:", err)
				w.WriteHeader(http.StatusGatewayTimeout)
			} else {
				log.Println("JSONHandlePostURL:", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}
		log.Println("JSONHandlePostURL: stored", post.URL, "as", id)
		// serialize struct into JSON
		u, err := url.Parse(h.serverConfig.BaseURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		u.Path = id
		resData := modeldto.ResponseURL{
			ShortURL: u.String(),
		}
		resBody, err := json.Marshal(resData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// set and send response body
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(resBody)
	}
}
