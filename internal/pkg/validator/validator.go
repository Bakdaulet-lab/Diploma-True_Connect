package validator

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var kzPhoneRegex = regexp.MustCompile(`^\+7\d{10}$`)

// Validate is the shared validator instance with custom rules registered.
var Validate *validator.Validate

func init() {
	Validate = validator.New()

	err := Validate.RegisterValidation("kz_phone", func(fl validator.FieldLevel) bool {
		return kzPhoneRegex.MatchString(fl.Field().String())
	})
	if err != nil {
		panic(fmt.Sprintf("registering kz_phone validation: %v", err))
	}
}

// FormatErrors extracts human-readable messages from validation errors.
func FormatErrors(err error) []FieldError {
	var fieldErrors []FieldError

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return []FieldError{{Field: "unknown", Reason: err.Error()}}
	}

	for _, e := range validationErrors {
		fieldErrors = append(fieldErrors, FieldError{
			Field:  e.Field(),
			Reason: formatReason(e),
		})
	}

	return fieldErrors
}

type FieldError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

func formatReason(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", e.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", e.Param())
	case "kz_phone":
		return "must be a valid KZ phone number (+7XXXXXXXXXX)"
	default:
		return fmt.Sprintf("failed on '%s' validation", e.Tag())
	}
}
