package helpers

import (
	"errors"
	"fmt"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrSetting struct {
	Error   string
	Message string
}

var ErrSettings = map[codes.Code]ErrSetting{
	codes.FailedPrecondition: {"400 Bad Request", "Некорректный запрос, например, отсутствует один из обязательных параметров."},
	codes.Unauthenticated:    {"401 Unauthorized", "Авторизуйтесь и попробуйте еще раз."},
	codes.PermissionDenied:   {"403 Forbidden", "Доступ запрещен. У вас нет прав для совершения действия."},
	codes.NotFound:           {"404 Not Found", "Страница не найдена. Вы ввели неверный адрес или такой страницы не существует."},
	codes.Internal:           {"500 Internal Server Error", "Что-то пошло не так. Попробуйте еще раз позже."},
	codes.Unavailable:        {"503 Service Unavailable", "Что-то пошло не так. Попробуйте еще раз позже."},
	codes.DeadlineExceeded:   {"504 Gateway Timeout", "Что-то пошло не так. Перезагрузите страницу и попробуйте еще раз."},
	codes.Unknown:            {"500 Undefined Error", "Что-то пошло не так. Перезагрузите страницу и попробуйте еще раз."},
}

type WrappedError struct {
	Context string
	Err     error
}

type RestError struct {
	Message string
	Err     error
}

func (w *WrappedError) Error() string {
	return fmt.Sprintf("%s: %v", w.Context, w.Err)
}

func (r *RestError) Error() string {
	return fmt.Sprintf("%s|%v", r.Err, r.Message)
}

type ErrorFormatter struct{}

func (ef ErrorFormatter) Wrap(err error, info string) *WrappedError {
	return &WrappedError{
		Context: info,
		Err:     err,
	}
}

func (ef ErrorFormatter) ReturnError(code codes.Code, err error, message string) error {
	if err.Error() == "" { // получаем error из настроек
		errorSettingForCode := ErrSettings[code]
		err = errors.New(errorSettingForCode.Error)
	}
	if message == "" { // получаем message из настроек
		errorSettingForCode := ErrSettings[code]
		message = errorSettingForCode.Message
	}
	errTest := status.New(code, err.Error())
	errorObject := map[string]string{
		"message": message,
	}
	br := &errdetails.ErrorInfo{Metadata: errorObject}
	errTest, _ = errTest.WithDetails(br)
	return errTest.Err()
}
