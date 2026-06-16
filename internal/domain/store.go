package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Store struct {
	ID         uuid.UUID
	MerchantID uuid.UUID
	Name       string
	CreatedAt  time.Time
}

func NewStore(id, merchantID uuid.UUID, name string, createdAt time.Time) (Store, error) {
	if merchantID == uuid.Nil {
		return Store{}, &ValidationError{Field: "merchantID", Msg: "is required"}
	}
	if strings.TrimSpace(name) == "" {
		return Store{}, &ValidationError{Field: "name", Msg: "is required"}
	}
	return Store{ID: id, MerchantID: merchantID, Name: name, CreatedAt: createdAt}, nil
}
