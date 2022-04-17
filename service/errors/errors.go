package errors

type (
	ServiceInitHashError struct {
		Msg string
	}
	ServiceEncodingHashError struct {
		Msg string
	}
	ServiceDecodingHashError struct {
		Msg string
	}
)

func (e *ServiceInitHashError) Error() string {
	return e.Msg
}

func (e *ServiceEncodingHashError) Error() string {
	return e.Msg
}

func (e *ServiceDecodingHashError) Error() string {
	return e.Msg
}
