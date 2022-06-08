// Package errors provides custom errors for types implementing URLSetter, URLGetter and URLStorage interfaces.
package errors

import (
	"fmt"
)

type (
	NotFoundError struct {
		SURL string
		Err  error
	}
	AlreadyExistsError struct {
		URL       string
		ValidSURL string
		Err       error
	}
	DeletedError struct {
		SURL string
		Err  error
	}
	ContextTimeoutExceededError struct {
		Err error
	}
	StatementPSQLError struct {
		Err error
	}
	ScanningPSQLError struct {
		Err error
	}
	ExecutionPSQLError struct {
		Err error
	}
	FileWriteError struct {
		Err error
	}
)

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s: not found in storage", e.SURL)
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("%s: already exists in storage", e.URL)
}

func (e *DeletedError) Error() string {
	return fmt.Sprintf("%s: was deleted", e.SURL)
}

func (e *ContextTimeoutExceededError) Error() string {
	return fmt.Sprintf("%s: context timeout exceeded", e.Err.Error())
}

func (e *ScanningPSQLError) Error() string {
	return fmt.Sprintf("%s: could not scan rows", e.Err.Error())
}

func (e *StatementPSQLError) Error() string {
	return fmt.Sprintf("%s: could not compile statement", e.Err.Error())
}

func (e *ExecutionPSQLError) Error() string {
	return fmt.Sprintf("%s: could not query", e.Err.Error())
}

func (e *FileWriteError) Error() string {
	return fmt.Sprintf("%s: could not add to file", e.Err.Error())
}

func (e *NotFoundError) Unwrap() error {
	return e.Err
}

func (e *AlreadyExistsError) Unwrap() error {
	return e.Err
}

func (e *ContextTimeoutExceededError) Unwrap() error {
	return e.Err
}

func (e *ScanningPSQLError) Unwrap() error {
	return e.Err
}

func (e *StatementPSQLError) Unwrap() error {
	return e.Err
}

func (e *ExecutionPSQLError) Unwrap() error {
	return e.Err
}

func (e *FileWriteError) Unwrap() error {
	return e.Err
}
