package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/AeroAgency/golang-helpers-lib/dto"
	"google.golang.org/grpc/codes"
)

// Сервис работы с привилегиями
type Privileges struct {
	meta           *Meta
	errorFormatter *ErrorFormatter
}

// Конструктор
func NewPrivileges() *Privileges {
	metaService := &Meta{}
	errorFormatter := &ErrorFormatter{}
	return &Privileges{
		errorFormatter: errorFormatter,
		meta:           metaService,
	}
}

// Получение привилегий из контекста
func (p Privileges) GetPrivilegesByContext(ctx context.Context) (dto.Privileges, error) {
	privileges := dto.Privileges{}
	privilegesData, err := p.meta.GetDecodedParam(ctx, "privileges")
	if err != nil {
		return dto.Privileges{}, err
	}
	err = json.Unmarshal([]byte(privilegesData), &privileges)
	return privileges, nil
}

// Получение привилегий для авторизованного пользователя
func (p Privileges) GetPrivilegesForAuthorizedUser(ctx context.Context) (dto.Privileges, error) {
	privileges, err := p.GetPrivilegesByContext(ctx)
	if err != nil {
		err = p.errorFormatter.Wrap(err, "Caught error while getting privileges from meta")
		err = p.errorFormatter.ReturnError(codes.PermissionDenied, err, "")
		return dto.Privileges{}, err
	}
	// Проверка авторизованности
	if privileges.IsAuthorized == false {
		return dto.Privileges{}, p.errorFormatter.ReturnError(codes.Unauthenticated, errors.New("unauthenticated"), "")
	}
	return privileges, nil
}
