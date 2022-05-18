// Package errors provides custom errors for types implementing Processor interface.
package errors

type (
	ServiceInitHashError struct {
		Msg string
	}
	ServiceEncodingHashError struct {
		Msg string
	}
	ServiceFoundNilStorage struct {
		Msg string
	}
	ServiceFoundNilSecretary struct {
		Msg string
	}
	ServiceIncorrectInputURL struct {
		Msg string
	}
)

func (e *ServiceInitHashError) Error() string {
	return e.Msg
}

func (e *ServiceEncodingHashError) Error() string {
	return e.Msg
}

func (e *ServiceFoundNilStorage) Error() string {
	return e.Msg
}

func (e *ServiceFoundNilSecretary) Error() string {
	return e.Msg
}

func (e *ServiceIncorrectInputURL) Error() string {
	return e.Msg
}
