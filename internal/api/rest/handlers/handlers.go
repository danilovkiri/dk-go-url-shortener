// Package handlers provides http.HandlerFunc handler functions to be used for endpoints.
package handlers

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/danilovkiri/dk_go_url_shortener/internal/api/rest/middleware"
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
			}
			log.Println("HandleGetURL:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
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
		userID, err := getUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("POST request detected for", string(b))
		// encode URL into sURL and store
		sURL, err := h.processor.Encode(ctx, string(b), userID)
		if err != nil {
			if errors.Is(err, storageErrors.ContextTimeoutExceededError{}) {
				log.Println("HandlePostURL:", err)
				w.WriteHeader(http.StatusGatewayTimeout)
			}
			log.Println("HandlePostURL:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println("HandlePostURL: stored", string(b), "as", sURL)
		// set and send response
		w.WriteHeader(http.StatusCreated)
		u, err := url.Parse(h.serverConfig.BaseURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		u.Path = sURL
		_, err = w.Write([]byte(u.String()))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

// HandleGetURLsByUserID provides shortening service using modeldto.ResponseFullURL schema.
func (h *URLHandler) HandleGetURLsByUserID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// set context timeout to 500 ms for timing DB operations
		ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
		defer cancel()
		var responseURLs []modeldto.ResponseFullURL
		// retrieve user identifier
		userID, err := getUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// retrieve all pairs of sURL:URL for that particular user
		URLs, err := h.processor.DecodeByUserID(ctx, userID)
		if err != nil {
			if errors.Is(err, storageErrors.ContextTimeoutExceededError{}) {
				log.Println("HandleGetURLsByUserID:", err)
				w.WriteHeader(http.StatusGatewayTimeout)
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// response with HTTP code 204 if no content was found for that user
		if len(URLs) == 0 {
			http.Error(w, "", http.StatusNoContent)
			return
		}
		// create and serialize response object into JSON
		u, err := url.Parse(h.serverConfig.BaseURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		for _, fullURL := range URLs {
			u.Path = fullURL.SURL
			responseURL := modeldto.ResponseFullURL{
				URL:  fullURL.URL,
				SURL: u.String(),
			}
			responseURLs = append(responseURLs, responseURL)
		}
		resBody, err := json.Marshal(responseURLs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// set and send response body
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(resBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

// JSONHandlePostURL provides shortening service for single URL processing using modeldto.RequestURL and
// modeldto.ResponseURL schemas.
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
		userID, err := getUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("JSON POST request detected for", post.URL)
		// encode URL into sURL and store them
		id, err := h.processor.Encode(ctx, post.URL, userID)
		if err != nil {
			if errors.Is(err, storageErrors.ContextTimeoutExceededError{}) {
				log.Println("JSONHandlePostURL:", err)
				w.WriteHeader(http.StatusGatewayTimeout)
			}
			log.Println("JSONHandlePostURL:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
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
			SURL: u.String(),
		}
		resBody, err := json.Marshal(resData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// set and send response body
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(resBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

// HandlePingDB handles PSQL DB pinging to check connection status.
func (h *URLHandler) HandlePingDB() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h.processor.PingDB()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
}

// getUserID retrieves user identifier as a value of cookie with key middleware.UserCookieKey.
func getUserID(r *http.Request) (string, error) {
	userCookie, err := r.Cookie(middleware.UserCookieKey)
	if err != nil {
		return "", err
	}
	token := userCookie.Value
	data, err := hex.DecodeString(token)
	if err != nil {
		return "", err
	}
	userID := data
	return hex.EncodeToString(userID), nil
}

// JSONHandlePostURLBatch provides shortening service for batch processing using modeldto.RequestBatchURL and
// modeldto.ResponseBatchURL schemas.
func (h *URLHandler) JSONHandlePostURLBatch() http.HandlerFunc {
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
		var post []modeldto.RequestBatchURL
		err = json.Unmarshal(b, &post)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		userID, err := getUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("JSON POST batch request detected for", post)
		// prepare url schema for sURL
		u, err := url.Parse(h.serverConfig.BaseURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		// encode URLs into sURLs and store them
		var responseBatchURLs []modeldto.ResponseBatchURL
		for _, requestBatchURL := range post {
			// if error occurs at any iteration - response with 400 or 504 even if some URLs were processed
			id, err := h.processor.Encode(ctx, requestBatchURL.URL, userID)
			if err != nil {
				if errors.Is(err, storageErrors.ContextTimeoutExceededError{}) {
					log.Println("JSONHandlePostURLBatch:", err)
					w.WriteHeader(http.StatusGatewayTimeout)
				}
				log.Println("JSONHandlePostURLBatch:", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			log.Println("JSONHandlePostURLBatch: stored", requestBatchURL.URL, "as", id)
			u.Path = id
			responseBatchURL := modeldto.ResponseBatchURL{
				CorrelationID: requestBatchURL.CorrelationID,
				SURL:          u.String(),
			}
			responseBatchURLs = append(responseBatchURLs, responseBatchURL)
		}
		// serialize struct into JSON
		resBody, err := json.Marshal(responseBatchURLs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// set and send response body
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(resBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}
