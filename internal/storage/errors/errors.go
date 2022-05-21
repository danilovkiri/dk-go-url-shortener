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
	StoragePSQLAlreadyExistsError struct {
		UserID string
		URL    string
	}
	ContextTimeoutExceededError struct {
	}
	StatementPSQLError struct {
		Msg string
	}
	StorageFileWriteError struct {
	}
)

func (e StorageNotFoundError) Error() string {
	return fmt.Sprintf("%s not found in storage", e.ID)
}

func (e StorageAlreadyExistsError) Error() string {
	return fmt.Sprintf("%s already exists", e.ID)
}

func (e StoragePSQLAlreadyExistsError) Error() string {
	return fmt.Sprintf("%s for user %s already exists", e.URL, e.UserID)
}

func (e ContextTimeoutExceededError) Error() string {
	return fmt.Sprintln("context timeout exceeded")
}

func (e StorageFileWriteError) Error() string {
	return fmt.Sprintln("Add to file error")
}

func (e StatementPSQLError) Error() string {
	return fmt.Sprintf("%s could not compile", e.Msg)
}
