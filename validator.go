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

func init() {
	// добавляем правило на максимальную длину строки с учетом кириллицы
	govalidator.AddCustomRule("max_string_len", func(field string, rule string, message string, value interface{}) error {
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

	// добавляем правило на минимальную длину строки с учетом кириллицы
	govalidator.AddCustomRule("min_string_len", func(field string, rule string, message string, value interface{}) error {
		mustLen := strings.TrimPrefix(rule, "min_string_len:")
		minLen, err := strconv.Atoi(mustLen)
		if err != nil {
			panic(errors.New("govalidator: unable to parse string to integer"))
		}
		val := []rune(value.(string))
		if len(val) < minLen {
			return fmt.Errorf("The %s field must be minium %d char", field, minLen)
		}
		return nil
	})
}

// ValidateProto Валидация объекта
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

// ValidateProtoWithCustomMessages Валидация объекта c возможностью передачи настраиваимых сообщений
func (validator Validator) ValidateProtoWithCustomMessages(inputStruct interface{}, rules map[string][]string, messages map[string][]string) error {
	opts := govalidator.Options{
		Data:     inputStruct,
		Rules:    rules,
		Messages: messages,
	}
	v := govalidator.New(opts)
	e := v.ValidateStruct()
	if len(e) > 0 {
		for k := range e {
			return errors.New(e.Get(k))
		}
	}
	return nil
}

// ValidateStruct Общий метод для проверки корректности структуры
func ValidateStruct(data interface{}, rules map[string][]string, messages map[string][]string) error {
	opts := govalidator.Options{
		Data:            data,
		Rules:           rules,    // rules map
		Messages:        messages, // custom message map (Optional)
		RequiredDefault: true,     // all the field to be pass the rules
	}
	v := govalidator.New(opts)
	validationResult := v.ValidateStruct()
	var firstErrorMsg string
	if len(validationResult) > 0 {
		for _, item := range validationResult {
			firstErrorMsg = item[0]
			break
		}
		return fmt.Errorf(firstErrorMsg)
	}
	return nil
}
