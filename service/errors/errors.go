package errors

type (
	ServiceHashWriteError struct {
		Msg string
	}
)

func (e *ServiceHashWriteError) Error() string {
	return e.Msg
}
