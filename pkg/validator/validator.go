package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ValidateStruct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errMsgs []string
			for _, e := range validationErrors {
				errMsgs = append(errMsgs, fmt.Sprintf("field '%s' failed on '%s' tag", e.Field(), e.Tag()))
			}
			return fmt.Errorf("validation error: %s", strings.Join(errMsgs, "; "))
		}
		return err
	}
	return nil
}
