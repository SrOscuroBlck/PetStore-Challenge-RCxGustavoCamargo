package graph

import (
	"roboticCrewChallenge/internal/domain"
	"roboticCrewChallenge/internal/graph/generated"
	"roboticCrewChallenge/internal/platform/pagination"
)

// picturePath maps a stored object key to the API path that serves the picture.
// It is the read side of the route mounted at GET /pictures/{objectKey...} in
// internal/server, kept same-origin so a browser fetches it over the API's TLS.
func picturePath(objectKey string) string {
	return "/pictures/" + objectKey
}

func toGraphPet(pet domain.Pet) *generated.Pet {
	return &generated.Pet{
		ID:               pet.ID.String(),
		Name:             pet.Name,
		Species:          generated.Species(pet.Species),
		AgeYears:         pet.AgeYears,
		Description:      pet.Description,
		Status:           generated.PetStatus(pet.Status),
		CreatedAt:        pet.CreatedAt,
		SoldAt:           pet.SoldAt,
		BreederName:      pet.BreederName,
		BreederEmail:     pet.BreederEmail,
		PictureObjectKey: pet.PictureObjectKey,
	}
}

func toGraphPublicPet(pet domain.Pet) *generated.PublicPet {
	return &generated.PublicPet{
		ID:               pet.ID.String(),
		Name:             pet.Name,
		Species:          generated.Species(pet.Species),
		AgeYears:         pet.AgeYears,
		Description:      pet.Description,
		Status:           generated.PetStatus(pet.Status),
		CreatedAt:        pet.CreatedAt,
		SoldAt:           pet.SoldAt,
		PictureObjectKey: pet.PictureObjectKey,
	}
}

func toPublicPets(pets []domain.Pet) []generated.PublicPet {
	out := make([]generated.PublicPet, 0, len(pets))
	for _, pet := range pets {
		out = append(out, *toGraphPublicPet(pet))
	}
	return out
}

func toPublicConnection(pets []domain.Pet, nextCursor string) *generated.PublicPetConnection {
	edges := make([]generated.PublicPetEdge, 0, len(pets))
	for _, pet := range pets {
		edges = append(edges, generated.PublicPetEdge{
			Node:   toGraphPublicPet(pet),
			Cursor: pagination.Encode(availableCursor(pet)),
		})
	}

	pageInfo := &generated.PageInfo{HasNextPage: nextCursor != ""}
	if len(edges) > 0 {
		endCursor := edges[len(edges)-1].Cursor
		pageInfo.EndCursor = &endCursor
	}

	return &generated.PublicPetConnection{Edges: edges, PageInfo: pageInfo}
}

// availableCursor and soldCursor derive a pet's keyset cursor: the available
// listing orders by created_at, the sold listing by sold_at.
func availableCursor(pet domain.Pet) pagination.Cursor {
	return pagination.Cursor{SortKey: pet.CreatedAt, ID: pet.ID}
}

func soldCursor(pet domain.Pet) pagination.Cursor {
	key := pagination.Cursor{ID: pet.ID}
	if pet.SoldAt != nil {
		key.SortKey = *pet.SoldAt
	}
	return key
}

func toConnection(pets []domain.Pet, nextCursor string, cursorKey func(domain.Pet) pagination.Cursor) *generated.PetConnection {
	edges := make([]generated.PetEdge, 0, len(pets))
	for _, pet := range pets {
		edges = append(edges, generated.PetEdge{
			Node:   toGraphPet(pet),
			Cursor: pagination.Encode(cursorKey(pet)),
		})
	}

	pageInfo := &generated.PageInfo{HasNextPage: nextCursor != ""}
	if len(edges) > 0 {
		endCursor := edges[len(edges)-1].Cursor
		pageInfo.EndCursor = &endCursor
	}

	return &generated.PetConnection{Edges: edges, PageInfo: pageInfo}
}
