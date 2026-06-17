package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"roboticCrewChallenge/internal/domain"
)

type demoPet struct {
	name        string
	species     domain.Species
	ageYears    int
	description string
}

// demoCatalog is the set of pets seeded into the demo store. The species mix is
// deliberately uneven (more dogs than one page holds) so that both the overall
// catalog and a single-species filter page beyond the default page size — making
// pagination demonstrable end to end.
var demoCatalog = []demoPet{
	{"Rex", domain.SpeciesDog, 3, "A loyal shepherd mix who loves long walks."},
	{"Bella", domain.SpeciesDog, 2, "Gentle retriever, great with children."},
	{"Max", domain.SpeciesDog, 5, "Energetic terrier looking for an active home."},
	{"Luna", domain.SpeciesDog, 1, "Playful husky pup with striking blue eyes."},
	{"Charlie", domain.SpeciesDog, 4, "Calm beagle who naps more than he barks."},
	{"Daisy", domain.SpeciesDog, 6, "Affectionate spaniel, fully house-trained."},
	{"Cooper", domain.SpeciesDog, 2, "Smart border collie, knows a dozen tricks."},
	{"Lucy", domain.SpeciesDog, 7, "Senior labrador with a heart of gold."},
	{"Buddy", domain.SpeciesDog, 3, "Friendly mutt who gets along with cats."},
	{"Sadie", domain.SpeciesDog, 4, "Quiet greyhound, happiest on the couch."},
	{"Rocky", domain.SpeciesDog, 5, "Sturdy boxer with boundless enthusiasm."},
	{"Molly", domain.SpeciesDog, 2, "Sweet corgi with very short legs."},
	{"Bailey", domain.SpeciesDog, 8, "Wise old hound who loves a slow stroll."},

	{"Whiskers", domain.SpeciesCat, 4, "Independent tabby who rules the household."},
	{"Oliver", domain.SpeciesCat, 2, "Curious shorthair, always exploring."},
	{"Cleo", domain.SpeciesCat, 6, "Elegant siamese with a soft meow."},
	{"Simba", domain.SpeciesCat, 1, "Tiny orange kitten full of mischief."},
	{"Nala", domain.SpeciesCat, 3, "Cuddly calico who purrs on contact."},
	{"Milo", domain.SpeciesCat, 5, "Laid-back tuxedo cat, loves windowsills."},

	{"Hopkins", domain.SpeciesFrog, 1, "Bright green tree frog, easy to care for."},
	{"Kermit", domain.SpeciesFrog, 2, "Classic pond frog with a deep croak."},
	{"Pepe", domain.SpeciesFrog, 1, "Small dart frog with vivid markings."},
	{"Jade", domain.SpeciesFrog, 3, "Calm bullfrog who enjoys a big tank."},
	{"Bubbles", domain.SpeciesFrog, 2, "Playful aquatic frog, always swimming."},
}

var speciesColors = map[domain.Species]color.RGBA{
	domain.SpeciesDog:  {R: 139, G: 94, B: 60, A: 255},
	domain.SpeciesCat:  {R: 230, G: 140, B: 60, A: 255},
	domain.SpeciesFrog: {R: 80, G: 170, B: 90, A: 255},
}

// speciesPicture renders a solid-color PNG placeholder so every seeded pet has a
// real, servable image — distinct per species — without bundling binary assets in
// the repository.
func speciesPicture(species domain.Species) ([]byte, error) {
	const size = 400
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	fill, ok := speciesColors[species]
	if !ok {
		return nil, fmt.Errorf("no placeholder color for species %q", species)
	}
	draw.Draw(img, img.Bounds(), &image.Uniform{C: fill}, image.Point{}, draw.Src)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode %s placeholder: %w", species, err)
	}
	return buf.Bytes(), nil
}
