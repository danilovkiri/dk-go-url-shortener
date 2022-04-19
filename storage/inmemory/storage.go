// Package inmemory provides functionality for dumping/retrieving pairs of URL and sURL to/from local
// storage implemented as a map.
package inmemory

import (
	"context"
	"github.com/danilovkiri/dk_go_url_shortener/storage/errors"
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
	URL, ok := s.DB[sURL]
	if !ok {
		return "", &errors.StorageNotFoundError{ID: sURL}
	}
	return URL, nil
}

// Dump stores a pair of sURL and URL as a key-value pair in a map.
func (s *Storage) Dump(ctx context.Context, URL string, sURL string) error {
	_, ok := s.DB[sURL]
	if ok {
		return &errors.StorageAlreadyExistsError{ID: sURL}
	}
	s.DB[sURL] = URL
	return nil
}
