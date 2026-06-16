package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestUnavailablePetsError_NamesEveryPet(t *testing.T) {
	err := &UnavailablePetsError{Pets: []UnavailablePet{
		{ID: uuid.New(), Name: "Pluto"},
		{ID: uuid.New(), Name: "Rex"},
	}}

	want := "these pets are no longer available: Pluto, Rex"
	if got := err.Error(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
