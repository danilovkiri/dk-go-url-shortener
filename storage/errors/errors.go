package errors

import (
	"fmt"
)

type (
	StorageNotFoundError struct {
		ShortURL string
	}
	StorageAlreadyExistsError struct {
		ShortURL string
	}
)

func (e *StorageNotFoundError) Error() string {
	return fmt.Sprintf("%s not found in storage", e.ShortURL)
}

func (e *StorageAlreadyExistsError) Error() string {
	return fmt.Sprintf("%s already exists", e.ShortURL)
}
