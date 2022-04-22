// Package inmemory provides functionality for dumping/retrieving pairs of URL and sURL to/from local
// storage implemented as a map.
package inmemory

import (
	"context"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/errors"
	"log"
)

// Storage struct defines data structure handling and provides support for adding new implementations.
type Storage struct {
	DB map[string]string
}

// InitStorage initializes a Storage object and sets its attributes.
func InitStorage() *Storage {
	db := make(map[string]string)
	return &Storage{DB: db}
}

// Retrieve returns a URL as a value of a map based on the given sURL as a key of a map.
func (s *Storage) Retrieve(ctx context.Context, sURL string) (URL string, err error) {
	// create channels for listening to the go routine result
	retrieveDone := make(chan string)
	retrieveError := make(chan string)
	go func() {
		URL, ok := s.DB[sURL]
		if !ok {
			retrieveError <- "not found in DB"
			return
		}
		retrieveDone <- URL
	}()

	// wait for the first channel to retrieve a value
	select {
	case <-ctx.Done():
		log.Println("Retrieving URL:", ctx.Err())
		return "", errors.ContextTimeoutExceededError{}
	case errString := <-retrieveError:
		log.Println("Retrieving URL:", errString)
		return "", errors.StorageNotFoundError{ID: sURL}
	case URL := <-retrieveDone:
		log.Println("Retrieving URL:", sURL, "as", URL)
		return URL, nil
	}
}

// Dump stores a pair of sURL and URL as a key-value pair in a map.
func (s *Storage) Dump(ctx context.Context, URL string, sURL string) error {
	// create channels for listening to the go routine result
	dumpDone := make(chan bool)
	dumpError := make(chan string)
	go func() {
		_, ok := s.DB[sURL]
		if ok {
			dumpError <- "already exists in DB"
			return
		}
		s.DB[sURL] = URL
		dumpDone <- true
	}()

	// wait for the first channel to retrieve a value
	select {
	case <-ctx.Done():
		log.Println("Dumping URL:", ctx.Err())
		return errors.ContextTimeoutExceededError{}
	case errString := <-dumpError:
		log.Println("Dumping URL:", errString)
		return errors.StorageAlreadyExistsError{ID: sURL}
	case <-dumpDone:
		log.Println("Dumping URL:", sURL, "as", URL)
		return nil
	}
}
