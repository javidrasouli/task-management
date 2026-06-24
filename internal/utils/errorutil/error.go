package errorutil

import "fmt"

// ErrorType pairs an HTTP status code with its localized user-facing message.
// Define all domain error kinds as ErrorType vars in message.go.
type ErrorType struct {
	Code    int
	Message string
}

type Error struct {
	message string
	code    int
	err     error
}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) Code() int {
	return e.code
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.code == t.code
}

func (e *Error) String() string {
	if e.err != nil {
		return fmt.Sprintf("code=%d message=%q cause=%v", e.code, e.message, e.err)
	}
	return fmt.Sprintf("code=%d message=%q", e.code, e.message)
}

// NewError constructs a domain Error from an ErrorType and an optional cause.
func NewError(errType ErrorType, err error) *Error {
	return &Error{
		code:    errType.Code,
		message: errType.Message,
		err:     err,
	}
}
