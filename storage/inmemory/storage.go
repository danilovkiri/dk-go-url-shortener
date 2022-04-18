package inmemory

import (
	"context"
	"github.com/danilovkiri/dk_go_url_shortener/storage/errors"
)

type Storage struct {
	DB map[string]string
}

func InitStorage() *Storage {
	db := make(map[string]string)
	return &Storage{DB: db}
}

func (s *Storage) Retrieve(ctx context.Context, sURL string) (URL string, err error) {
	URL, ok := s.DB[sURL]
	if !ok {
		return "", &errors.StorageNotFoundError{ID: sURL}
	}
	return URL, nil
}

func (s *Storage) Dump(ctx context.Context, URL string, sURL string) error {
	_, ok := s.DB[sURL]
	if ok {
		return &errors.StorageAlreadyExistsError{ID: sURL}
	}
	s.DB[sURL] = URL
	return nil
}
