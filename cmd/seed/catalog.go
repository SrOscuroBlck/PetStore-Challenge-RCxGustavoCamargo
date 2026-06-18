package main

import (
	"embed"
	"fmt"
	"strings"

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

// store2Catalog stocks a second store so multi-tenant isolation is demonstrable:
// the second merchant sees only these pets, and cannot touch the first store's.
var store2Catalog = []demoPet{
	{"Shadow", domain.SpeciesCat, 4, "Sleek black cat, the second store's mascot."},
	{"Ziggy", domain.SpeciesDog, 3, "Spotted dalmatian with endless energy."},
	{"Coco", domain.SpeciesFrog, 1, "Tiny brown frog who loves a humid tank."},
}

// petImages holds the bundled demo photos. They are embedded (not fetched at
// runtime) so the seeded store has real pictures while the system stays fully
// local. See assets/CREDITS.md for sources.
//
//go:embed assets/*.jpg assets/*.png
var petImages embed.FS

var speciesImagePrefix = map[domain.Species]string{
	domain.SpeciesDog:  "dog-",
	domain.SpeciesCat:  "cat-",
	domain.SpeciesFrog: "frog-",
}

// loadPetImages reads the embedded photos grouped by species, using the filename
// prefix to classify each one. ReadDir returns entries sorted by name, so the
// grouping is deterministic across runs.
func loadPetImages() (map[domain.Species][][]byte, error) {
	entries, err := petImages.ReadDir("assets")
	if err != nil {
		return nil, fmt.Errorf("read embedded images: %w", err)
	}
	images := make(map[domain.Species][][]byte)
	for _, entry := range entries {
		for species, prefix := range speciesImagePrefix {
			if !strings.HasPrefix(entry.Name(), prefix) {
				continue
			}
			data, err := petImages.ReadFile("assets/" + entry.Name())
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
			}
			images[species] = append(images[species], data)
		}
	}
	for species, prefix := range speciesImagePrefix {
		if len(images[species]) == 0 {
			return nil, fmt.Errorf("no embedded images with prefix %q for %s", prefix, species)
		}
	}
	return images, nil
}
