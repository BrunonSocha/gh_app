package validator

import (
	"strconv"
	"strings"
)

type Validator struct {
	FieldErrors map[string]string
}

func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0
}

func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func NotZero(value float64) bool {
	return value > 0
}

func LengthNIP(nip string) bool {
	return len(nip) == 10
}

func NumberNIP(nip string) bool {
	_, err := strconv.Atoi(nip)
	return err == nil
}
