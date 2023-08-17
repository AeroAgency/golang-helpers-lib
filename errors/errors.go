package errors

import (
	"github.com/pkg/errors"
)

const (
	NoType = ErrorType(iota)
	BadRequest
	NotFound
	Conflict
	RequestTimeout
	ServiceUnavailable
	Forbidden
	Unauthorized
	//add any type you want

	Internal = NoType
)

type ErrorType uint

type AppError struct {
	errorType     ErrorType
	originalError error
	Trans         *Trans
}

type Trans struct {
	Msg    string
	Params []string
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// Error returns the message of a AppError
func (e AppError) Error() string {
	return e.originalError.Error()
}

// New creates a new AppError
func (t ErrorType) New(msg string) AppError {
	return AppError{errorType: t, originalError: errors.New(msg)}
}

// Newf creates a new AppError with formatted message
func (t ErrorType) Newf(msg string, args ...interface{}) AppError {
	return AppError{errorType: t, originalError: errors.Errorf(msg, args...)}
}

// Wrap creates a new wrapped error
func (t ErrorType) Wrap(err error, msg string) AppError {
	return t.Wrapf(err, msg)
}

// Wrapf creates a new wrapped error with formatted message
func (t ErrorType) Wrapf(err error, msg string, args ...interface{}) AppError {
	wrappedError := errors.Wrapf(err, msg, args...)
	if customErr, ok := err.(AppError); ok {
		return AppError{
			errorType:     customErr.errorType,
			originalError: wrappedError,
			Trans:         customErr.Trans,
		}
	}

	return AppError{errorType: t, originalError: wrappedError}
}

func (e AppError) Unwrap() error {
	return errors.Unwrap(e.originalError)
}

func (e AppError) T(msg string, params ...string) AppError {
	return AppError{
		errorType:     e.errorType,
		originalError: e.originalError,
		Trans: &Trans{
			Msg:    msg,
			Params: params,
		},
	}
}

// Cause gives the original error
func Cause(err error) error {
	return errors.Cause(err)
}

// GetType returns the error type
func GetType(err error) ErrorType {
	if customErr, ok := err.(AppError); ok {
		return customErr.errorType
	}

	return NoType
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}

func GetStackTrace(err error) errors.StackTrace {
	if customErr, ok := err.(AppError); ok {
		err = customErr.originalError
	}
	if stacked, ok := err.(stackTracer); ok {
		return stacked.StackTrace()
	}
	return errors.StackTrace{}
}
