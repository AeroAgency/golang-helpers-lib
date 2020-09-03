package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/thedevsaddam/govalidator"
	"strconv"
	"strings"
)

type Validator struct{}

// Добавление кастомного правила
func (validator Validator) AddCustomRule(name string, fn func(field string, rule string, message string, value interface{}) error) {
	govalidator.AddCustomRule(name, fn)
	return
}

func (validator Validator) addUniversalMaxStringValidationRule() {
	validator.AddCustomRule("max_string_len", func(field string, rule string, message string, value interface{}) error {
		mustLen := strings.TrimPrefix(rule, "max_string_len:")
		maxLen, err := strconv.Atoi(mustLen)
		if err != nil {
			panic(errors.New("govalidator: unable to parse string to integer"))
		}
		val := []rune(value.(string))
		if len(val) > maxLen {
			return fmt.Errorf("The %s field must be maximum %d char", field, maxLen)
		}
		return nil
	})
}

// Валидация объекта
func (validator Validator) ValidateProto(inputStruct interface{}, rules map[string][]string) error {
	validator.addUniversalMaxStringValidationRule()
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
