package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Merchant struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

func NewMerchant(id uuid.UUID, email, passwordHash string, createdAt time.Time) (Merchant, error) {
	if err := validateEmail("email", email); err != nil {
		return Merchant{}, err
	}
	if strings.TrimSpace(passwordHash) == "" {
		return Merchant{}, &ValidationError{Field: "passwordHash", Msg: "is required"}
	}
	return Merchant{ID: id, Email: email, PasswordHash: passwordHash, CreatedAt: createdAt}, nil
}
