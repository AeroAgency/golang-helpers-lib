package helpers

import (
	"errors"
	"google.golang.org/grpc/codes"
)

// Сервис для разграничению доступа согласно ролевой модели
type Access struct {
	errorFormatter *ErrorFormatter
}

// Конструктор
func NewAccess() *Access {
	errorFormatter := &ErrorFormatter{}
	return &Access{errorFormatter: errorFormatter}
}

// Возвращает ошибку доступа
func (a Access) PermissionDenied() error {
	return a.errorFormatter.ReturnError(codes.PermissionDenied, errors.New(""), "")
}

// Возвращает признак присутствия переданного скоупа в списке скоупов, выданных пользователю
func (a Access) HasScope(scopes []string, scope string) bool {
	res := StringInSlice(scope, scopes)
	return res
}
