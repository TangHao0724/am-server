package game

type gameError struct {
	Status  int
	Code    string
	Message string
	Cause   error
}

func (e *gameError) Error() string {
	return e.Message
}

func (e *gameError) Unwrap() error {
	return e.Cause
}

func newGameError(
	status int,
	code string,
	message string,
	cause error,
) *gameError {
	return &gameError{
		Status:  status,
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}
