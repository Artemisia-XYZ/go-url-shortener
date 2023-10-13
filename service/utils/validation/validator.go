package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gobeam/stringy"
)

type errValidateResp struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

func snakeCaseLower(str string) string {
	return stringy.New(str).SnakeCase().ToLower()
}

func buildErrorMessage(err validator.FieldError) string {
	key := snakeCaseLower(err.Field())
	tag := err.Tag()
	param := err.Param()

	if param == "" {
		return fmt.Sprintf("[%v]: need to implement '%v'", key, tag)
	}

	return fmt.Sprintf("[%v]: need to implement '%v[%v]'", key, tag, param)
}

func ValidateStruct(data interface{}) []*errValidateResp {
	var validate = validator.New()
	var errs []*errValidateResp

	err := validate.Struct(data)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			msg := buildErrorMessage(err)
			element := errValidateResp{
				Field:   err.Field(),
				Tag:     err.Tag(),
				Key:     snakeCaseLower(err.Field()),
				Value:   err.Param(),
				Message: msg,
			}
			errs = append(errs, &element)
		}
	}

	fmt.Print(errs)

	return errs
}
