package domain

import (
	"net/mail"
	"strings"
)

func validateEmail(field, email string) error {
	if strings.TrimSpace(email) == "" {
		return &ValidationError{Field: field, Msg: "is required"}
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return &ValidationError{Field: field, Msg: "is not a valid email address"}
	}
	return nil
}
