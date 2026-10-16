package game

type createGameError struct {
	Status  int
	Code    string
	Message string
	Cause   error
}

func (e *createGameError) Error() string {
	return e.Message
}

func (e *createGameError) Unwrap() error {
	return e.Cause
}

func newCreateGameError(
	status int,
	code string,
	message string,
	cause error,
) *createGameError {
	return &createGameError{
		Status:  status,
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}
