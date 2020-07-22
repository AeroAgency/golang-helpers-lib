package helpers

import (
	"context"
	"encoding/json"
	"github.com/AeroAgency/golang-helpers-lib/dto"
)

type Privileges struct {
	meta *Meta
}

func NewPrivileges() *Privileges {
	metaService := &Meta{}
	return &Privileges{meta: metaService}
}

func (p Privileges) GetPrivilegesByContext(ctx context.Context) (dto.Privileges, error) {
	privileges := dto.Privileges{}
	privilegesData, err := p.meta.GetDecodedParam(ctx, "privileges")
	if err != nil {
		return dto.Privileges{}, err
	}
	err = json.Unmarshal([]byte(privilegesData), &privileges)
	return privileges, nil
}
