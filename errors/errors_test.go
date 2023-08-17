package errors

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
)

func TestErrorType(t *testing.T) {
	expectedMsg := "test message"

	err := BadRequest.New(expectedMsg)
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message: %s, but got: %s", expectedMsg, err.Error())
	}

	errf := NotFound.Newf("not found: %d", 404)
	if errf.Error() != "not found: 404" {
		t.Errorf("Expected error message: not found: 404, but got: %s", errf.Error())
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := BadRequest.Wrap(originalErr, "wrapped error")

	if wrappedErr.errorType != BadRequest {
		t.Errorf("Expected wrapped error type to be BadRequest, but got: %v", wrappedErr.errorType)
	}

}

func TestTrans(t *testing.T) {
	err := BadRequest.New("bad request").T("Translated message", "param1", "param2")

	if err.Trans == nil {
		t.Error("Expected Trans field to be not nil")
	}

	if err.Trans.Msg != "Translated message" {
		t.Errorf("Expected Translated message, but got: %s", err.Trans.Msg)
	}

	if len(err.Trans.Params) != 2 || err.Trans.Params[0] != "param1" || err.Trans.Params[1] != "param2" {
		t.Errorf("Expected params to be [param1 param2], but got: %v", err.Trans.Params)
	}
}

func TestIs(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := BadRequest.Wrap(originalErr, "wrapped error")

	if !Is(wrappedErr, originalErr) {
		t.Errorf("Expected Is to return true for wrapped error and original error")
	}

	otherErr := errors.New("other error")
	if Is(wrappedErr, otherErr) {
		t.Errorf("Expected Is to return false for wrapped error and other error")
	}
}

func TestGetType(t *testing.T) {
	err := BadRequest.New("bad request")
	typ := GetType(err)
	if typ != BadRequest {
		t.Errorf("Expected error type BadRequest, but got: %v", typ)
	}

	typ = GetType(errors.New("generic error"))
	if typ != NoType {
		t.Errorf("Expected error type NoType, but got: %v", typ)
	}
}

func TestGetStackTrace(t *testing.T) {
	appErr := BadRequest.New("bad request")
	errWithStackTrace := appErr.T("Translated message", "param1", "param2")

	stackTrace := GetStackTrace(errWithStackTrace)

	if len(stackTrace) == 0 {
		t.Error("Expected non-empty stack trace for error with stack trace")
	}

	// Now, let's create an error without stack trace
	errWithoutStackTrace := errors.New("error without stack trace")
	wrappedErrWithoutStackTrace := BadRequest.Wrap(errWithoutStackTrace, "wrapped error without stack trace")

	stackTrace = GetStackTrace(wrappedErrWithoutStackTrace)

	if len(stackTrace) == 0 {
		t.Error("Expected non-empty stack trace for error without stack trace, but got empty stack trace")
	}

	stackTrace = GetStackTrace(fmt.Errorf("123"))
}

func TestCause(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := errors.Wrap(originalErr, "wrapped error")
	cause := Cause(wrappedErr)
	if cause.Error() != originalErr.Error() {
		t.Errorf("Expected original error message for cause, but got: %v", cause.Error())
	}
}

func TestWrapf(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := BadRequest.Wrapf(originalErr, "wrapped error with %s", "formatting")

	if wrappedErr.errorType != BadRequest {
		t.Errorf("Expected error type to be BadRequest, but got: %v", wrappedErr.errorType)
	}

	if wrappedErr.Trans != nil {
		t.Error("Expected Trans to be nil")
	}

	if wrappedErr.originalError.Error() != "wrapped error with formatting: original error" {
		t.Errorf("Expected wrapped error message: wrapped error with formatting: original error, but got: %s", wrappedErr.originalError.Error())
	}

	// Now, let's create another wrapped error using the previous wrappedErr
	doubleWrappedErr := BadRequest.Wrapf(wrappedErr, "double wrapped error")

	if doubleWrappedErr.errorType != BadRequest {
		t.Errorf("Expected error type to be BadRequest, but got: %v", doubleWrappedErr.errorType)
	}

	if doubleWrappedErr.Trans != nil {
		t.Error("Expected Trans to be nil")
	}

	if doubleWrappedErr.originalError.Error() != "double wrapped error: wrapped error with formatting: original error" {
		t.Errorf("Expected wrapped error message: double wrapped error: wrapped error with formatting: original error, but got: %s", doubleWrappedErr.originalError.Error())
	}
}
