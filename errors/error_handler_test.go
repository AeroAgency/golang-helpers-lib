package errors

import (
	"bufio"
	"errors"
	tracerAdapter "github.com/AeroAgency/go-gin-tracer"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net"
	"net/http"
	"testing"
)

func TestErrorResponse(t *testing.T) {
	err := BadRequest.New("bad request")
	respErr := getResponseError(err)

	assert.Equal(t, http.StatusText(http.StatusBadRequest), respErr.Error)
	assert.Equal(t, errorMessages[BadRequest], respErr.Message)
	assert.Equal(t, err.Error(), respErr.Debug)
}

func TestErrorResponse_UnknownError(t *testing.T) {
	err := errors.New("unknown error")
	respErr := getResponseError(err)

	assert.Equal(t, http.StatusText(http.StatusInternalServerError), respErr.Error)
	assert.Equal(t, errorMessages[NoType], respErr.Message)
	assert.Equal(t, err.Error(), respErr.Debug)
}

// -----

type ResponseWriterMock struct {
}

func (r ResponseWriterMock) Header() http.Header {
	return http.Header{}
}

func (r ResponseWriterMock) Write(bytes []byte) (int, error) {
	return 0, nil
}

func (r ResponseWriterMock) WriteHeader(statusCode int) {
	return
}

func (r ResponseWriterMock) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	//TODO implement me
	panic("implement me2")
}

func (r ResponseWriterMock) Flush() {
	//TODO implement me
	panic("implement me3")
}

func (r ResponseWriterMock) CloseNotify() <-chan bool {
	//TODO implement me
	panic("implement me")
}

func (r ResponseWriterMock) Status() int {
	//TODO implement me
	panic("implement me")
}

func (r ResponseWriterMock) Size() int {
	//TODO implement me
	panic("implement me")
}

func (r ResponseWriterMock) WriteString(s string) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (r ResponseWriterMock) Written() bool {
	//TODO implement me
	panic("implement me")
}

func (r ResponseWriterMock) WriteHeaderNow() {
	//TODO implement me
	panic("implement me")
}

func (r ResponseWriterMock) Pusher() http.Pusher {
	//TODO implement me
	panic("implement me")
}

func TestHandleError(t *testing.T) {
	err := BadRequest.New("bad request")

	// Mocks
	mockTracer := &mockTracerAdapter{}
	mockLogger := &mockAppLoggerInterface{}

	c := &gin.Context{
		Writer: ResponseWriterMock{},
	}
	c.Set("tracer", mockTracer)

	h := ErrorHandler{logger: mockLogger}
	h.HandleError(c, err)

	assert.Len(t, mockTracer.loggedErrors, 1)
	assert.Len(t, mockLogger.loggedErrors, 0)
	mockTracer.loggedErrors = nil
	mockLogger.loggedErrors = nil
}

type mockTracerAdapter struct {
	mock.Mock
	loggedErrors []interface{}
}

func (m *mockTracerAdapter) GetScope() *tracerAdapter.Scope {
	//TODO implement me
	panic("implement me")
}

func (m *mockTracerAdapter) Close() {
	//TODO implement me
	panic("implement me")
}

func (m *mockTracerAdapter) SetTag(key string, value interface{}) {
	//TODO implement me
	panic("implement me")
}

func (m *mockTracerAdapter) SetTags(list map[string]interface{}) {
	//TODO implement me
	panic("implement me")
}

func (m *mockTracerAdapter) LogMessage(message string) {
	//TODO implement me
	panic("implement me")
}

func (m *mockTracerAdapter) Log(key, message string) {
	//TODO implement me
	panic("implement me")
}

func (m *mockTracerAdapter) LogData(key string, data interface{}) {
	//TODO implement me
	panic("implement me")
}

func (m *mockTracerAdapter) LogError(err interface{}) {
	m.loggedErrors = append(m.loggedErrors, err)
}

/*func (mt *mockTracerAdapter) LogError(err error) {
	mt.loggedErrors = append(mt.loggedErrors, err)
}*/

type mockAppLoggerInterface struct {
	mock.Mock
	loggedErrors []error
}

func (ml *mockAppLoggerInterface) Debug(msg string, keysAndValues ...interface{}) {
	//TODO implement me
	panic("implement me")
}

func (ml *mockAppLoggerInterface) Info(msg string, keysAndValues ...interface{}) {
	//TODO implement me
	panic("implement me")
}

func (ml *mockAppLoggerInterface) Error(err error, message string, keysAndValues ...interface{}) {
	ml.loggedErrors = append(ml.loggedErrors, err)
}

func TestHandleError_NoType(t *testing.T) {
	err := BadRequest.New("bad request")

	// Mocks
	mockTracer := &mockTracerAdapter{}
	mockLogger := &mockAppLoggerInterface{}

	c := &gin.Context{
		Writer: ResponseWriterMock{},
	}
	c.Set("tracer", mockTracer)

	h := ErrorHandler{logger: mockLogger}
	h.HandleError(c, err)

	assert.Len(t, mockTracer.loggedErrors, 1)
	assert.Len(t, mockLogger.loggedErrors, 0)
	mockTracer.loggedErrors = nil
	mockLogger.loggedErrors = nil
}

func TestHandleError_NoType_ErrorWithoutStackTrace(t *testing.T) {
	unknownErr := errors.New("unknown error")

	// Mocks
	mockTracer := &mockTracerAdapter{}
	mockLogger := &mockAppLoggerInterface{}

	c := &gin.Context{
		Writer: ResponseWriterMock{},
	}
	c.Set("tracer", mockTracer)

	h := ErrorHandler{logger: mockLogger}
	h.HandleError(c, unknownErr)

	assert.Len(t, mockTracer.loggedErrors, 1)
	assert.Len(t, mockLogger.loggedErrors, 1)
	mockTracer.loggedErrors = nil
	mockLogger.loggedErrors = nil
}

func TestHandleError_NoType_ErrorWithStackTrace(t *testing.T) {
	err := BadRequest.New("bad request")

	// Mocks
	mockTracer := &mockTracerAdapter{}
	mockLogger := &mockAppLoggerInterface{}

	c := &gin.Context{
		Writer: ResponseWriterMock{},
	}
	c.Set("tracer", mockTracer)

	h := ErrorHandler{logger: mockLogger}
	h.HandleError(c, err)

	assert.Len(t, mockTracer.loggedErrors, 1)
	assert.Len(t, mockLogger.loggedErrors, 0)
	mockTracer.loggedErrors = nil
	mockLogger.loggedErrors = nil
}

func TestHandleError_NoType_WrappedErrorWithStackTrace(t *testing.T) {
	err := errors.New("unknown error")
	wrappedErr := BadRequest.Wrap(err, "wrapped error")

	// Mocks
	mockTracer := &mockTracerAdapter{}
	mockLogger := &mockAppLoggerInterface{}

	c := &gin.Context{
		Writer: ResponseWriterMock{},
	}
	c.Set("tracer", mockTracer)

	h := ErrorHandler{logger: mockLogger}
	h.HandleError(c, wrappedErr)

	assert.Len(t, mockTracer.loggedErrors, 1)
	assert.Len(t, mockLogger.loggedErrors, 0)
	mockTracer.loggedErrors = nil
	mockLogger.loggedErrors = nil
}

func TestNewErrorHandler(t *testing.T) {
	NewErrorHandler(&mockAppLoggerInterface{})
}
