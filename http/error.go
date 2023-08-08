package http

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Error struct {
	httpCode int
	appCode  string
	error    error
	message  string
}

func New(s string) *Error {
	return &Error{
		error: errors.New(s),
	}
}

func Errorf(s string, a ...interface{}) *Error {
	return &Error{
		error: fmt.Errorf(s, a...),
	}
}

func (e *Error) SetMessage(s string, a ...interface{}) *Error {
	e.message = fmt.Sprintf(s, a...)
	return e
}

func (e *Error) SetAppCode(s string) *Error {
	e.appCode = s
	return e
}

func (e *Error) SetHttpCode(status int) *Error {
	e.httpCode = status
	return e
}

func (e *Error) Message() string {
	return e.message
}

func (e *Error) AppCode() string {
	return e.appCode
}

func (e *Error) HttpCode() int {
	return e.httpCode
}

func (e *Error) Error() string {
	return e.error.Error()
}

func (e *Error) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"applicationErrorCode": e.appCode,
		"message":              e.message,
		"debug":                e.error,
	}
	return json.Marshal(m)
}
