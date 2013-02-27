package errors

import (
	"fmt"
	"runtime/debug"
)

type Error struct {
	CallStack   string
	OriginalErr error
	Code        int
}

func (err *Error) Error() string {
	return fmt.Sprintf("[status code %d]\n%s\n\n%s", err.Code, err.OriginalErr, err.CallStack)
}

func New(original error) error {
	if _, ok := original.(*Error); ok {
		return original
	}

	return &Error{
		OriginalErr: original,
		Code:        500,
		CallStack:   fmt.Sprintf("%s", debug.Stack()),
	}
}

func Format(format string, args ...interface{}) error {
	return New(fmt.Errorf(format, args...))
}

func Code(code int) error {
	return &Error{
		Code:      code,
		CallStack: fmt.Sprintf("%s", debug.Stack()),
	}
}
