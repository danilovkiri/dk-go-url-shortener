// Package errors provides custom errors for types implementing URLSetter, URLGetter and URLStorage interfaces.
package errors

import (
	"fmt"
)

type (
	StorageNotFoundError struct {
		ID string
	}
	StorageAlreadyExistsError struct {
		ID string
	}
)

func (e *StorageNotFoundError) Error() string {
	return fmt.Sprintf("%s not found in storage", e.ID)
}

func (e *StorageAlreadyExistsError) Error() string {
	return fmt.Sprintf("%s already exists", e.ID)
}
