package helpers

import (
	"encoding/json"
	"errors"
	"github.com/thedevsaddam/govalidator"
)

type Validator struct{}

func (validator Validator) ValidateProto(inputStruct interface{}, rules map[string][]string) error {
	opts := govalidator.Options{
		Data:  inputStruct,
		Rules: rules,
	}

	v := govalidator.New(opts)
	e := v.ValidateStruct()
	if len(e) > 0 {
		errorsData, _ := json.MarshalIndent(e, "", "  ")
		return errors.New(string(errorsData))
	}
	return nil
}
