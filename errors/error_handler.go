package errors

import (
	"fmt"
	tracerAdapter "github.com/AeroAgency/go-gin-tracer"
	appLogger "github.com/AeroAgency/golang-helpers-lib/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

var (
	errorStatuses = map[ErrorType]int{
		NoType:             http.StatusInternalServerError,
		BadRequest:         http.StatusBadRequest,
		Unauthorized:       http.StatusUnauthorized,
		Forbidden:          http.StatusForbidden,
		NotFound:           http.StatusNotFound,
		RequestTimeout:     http.StatusRequestTimeout,
		Conflict:           http.StatusConflict,
		ServiceUnavailable: http.StatusServiceUnavailable,
	}

	errorMessages = map[ErrorType]string{
		NoType:             "Произошла ошибка. Попробуйте выполнить операцию позже.",
		BadRequest:         "Некорректный запрос, отсутствует один из обязательных параметров или один из параметров некорректный.",
		Unauthorized:       "Пользователь не авторизован.",
		Forbidden:          "Доступ запрещен.",
		NotFound:           "Запись не найдена.",
		RequestTimeout:     "Превышено время ожидания запроса.",
		Conflict:           "Конфликт.",
		ServiceUnavailable: "В настоящий момент сервис недоступен.",
	}
)

type ErrorHandler struct {
	logger appLogger.AppLoggerInterface
}

func NewErrorHandler(logger appLogger.AppLoggerInterface) *ErrorHandler {
	return &ErrorHandler{logger: logger}
}

type responseError struct {
	Error   interface{} `json:"error"`
	Message string      `json:"message"`
	Debug   interface{} `json:"debug"`
}

func (h ErrorHandler) HandleError(c *gin.Context, err error) {
	tracer, _ := c.MustGet("tracer").(tracerAdapter.TracerInterface)
	tracer.LogError(err)
	var status int
	errorType := GetType(err)
	errStatus, _ := errorStatuses[errorType]
	status = errStatus

	responseData := getResponseError(err)
	c.JSON(status, responseData)

	if errorType == NoType {
		stackTrace := GetStackTrace(Cause(err))
		h.logger.Error(err, fmt.Sprintf("service error. stackTrace %v", stackTrace))
	}
}

func getResponseError(err error) responseError {
	errorFormatted := responseError{}
	GetType(err)

	status := errorStatuses[GetType(err)]
	statusText := http.StatusText(status)
	message := errorMessages[GetType(err)]

	errorFormatted.Error = statusText
	errorFormatted.Message = message
	errorFormatted.Debug = err.Error()

	return errorFormatted
}
