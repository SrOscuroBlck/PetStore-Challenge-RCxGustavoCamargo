package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

func NewCustomer(id uuid.UUID, email, passwordHash string, createdAt time.Time) (Customer, error) {
	if err := validateEmail("email", email); err != nil {
		return Customer{}, err
	}
	if strings.TrimSpace(passwordHash) == "" {
		return Customer{}, &ValidationError{Field: "passwordHash", Msg: "is required"}
	}
	return Customer{ID: id, Email: email, PasswordHash: passwordHash, CreatedAt: createdAt}, nil
}
