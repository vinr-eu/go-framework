package app

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/vinr-eu/go-framework/code"
)

type Error struct {
	err  error
	code code.Code
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func NewError(err error) *Error {
	return &Error{err: err}
}

func NewErrorWithCode(err error, code code.Code) *Error {
	return &Error{err: err, code: code}
}

func NewErrorWithCodeAndMsg(code code.Code) *Error {
	return &Error{err: errors.WithStack(errors.New(code.GetText())), code: code}
}

func (e *Error) Error() string {
	return e.err.Error()
}

func (e *Error) GetError() error {
	return e.err
}

func (e *Error) GetStackTrace() errors.StackTrace {
	if err, ok := e.err.(stackTracer); ok {
		return err.StackTrace()
	} else {
		return nil
	}
}

func (e *Error) SetCode(code code.Code) {
	e.code = code
}

func (e *Error) GetCode() code.Code {
	return e.code
}

func (e *Error) String() string {
	if e.code.GetText() != "" {
		return fmt.Sprintf("%s: %s", e.code.GetText(), e.Error())
	}
	return e.Error()
}
