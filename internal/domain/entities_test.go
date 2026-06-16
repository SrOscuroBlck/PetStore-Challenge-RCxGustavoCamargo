package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewMerchant_Validation(t *testing.T) {
	if _, err := NewMerchant(uuid.New(), "m@example.com", "hash", time.Now()); err != nil {
		t.Fatalf("valid merchant rejected: %v", err)
	}
	var ve *ValidationError
	if _, err := NewMerchant(uuid.New(), "not-an-email", "hash", time.Now()); !errors.As(err, &ve) || ve.Field != "email" {
		t.Fatalf("expected an email ValidationError, got %v", err)
	}
	if _, err := NewMerchant(uuid.New(), "m@example.com", "  ", time.Now()); !errors.As(err, &ve) || ve.Field != "passwordHash" {
		t.Fatalf("expected a passwordHash ValidationError, got %v", err)
	}
}

func TestNewCustomer_Validation(t *testing.T) {
	if _, err := NewCustomer(uuid.New(), "c@example.com", "hash", time.Now()); err != nil {
		t.Fatalf("valid customer rejected: %v", err)
	}
	var ve *ValidationError
	if _, err := NewCustomer(uuid.New(), "nope", "hash", time.Now()); !errors.As(err, &ve) || ve.Field != "email" {
		t.Fatalf("expected an email ValidationError, got %v", err)
	}
}

func TestNewStore_Validation(t *testing.T) {
	if _, err := NewStore(uuid.New(), uuid.New(), "Pluto's Pets", time.Now()); err != nil {
		t.Fatalf("valid store rejected: %v", err)
	}
	var ve *ValidationError
	if _, err := NewStore(uuid.New(), uuid.Nil, "Pluto's Pets", time.Now()); !errors.As(err, &ve) || ve.Field != "merchantID" {
		t.Fatalf("expected a merchantID ValidationError, got %v", err)
	}
	if _, err := NewStore(uuid.New(), uuid.New(), "   ", time.Now()); !errors.As(err, &ve) || ve.Field != "name" {
		t.Fatalf("expected a name ValidationError, got %v", err)
	}
}
