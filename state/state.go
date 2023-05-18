package state

import (
	"github.com/brettschalin/factorio-min-resources/building"
	"github.com/brettschalin/factorio-min-resources/data"
)

type State struct {
	// character main inventory
	Inventory map[string]int

	TechResearched map[string]bool

	// What's been built?
	Buildings map[string]bool

	// which machines we have access to
	Furnace   *building.Furnace
	Assembler *building.Assembler
	Chem      *building.Assembler
	Refinery  *building.Assembler
	Boiler    *building.Boiler
	Lab       *building.Lab
}

func New() *State {
	f := data.D.Furnace["stone-furnace"]
	s := &State{
		TechResearched: make(map[string]bool),
		Buildings:      make(map[string]bool),
		Furnace:        building.NewFurnace(&f),
	}

	// Starting inventory
	s.Inventory = map[string]int{
		"stone-furnace":       1,
		"burner-mining-drill": 1,
		"wood":                1,

		// found in the spaceship wreckage. Will be in the inventory
		// so long as the generated "mine the ship" commands are kept
		"iron-plate": 8,

		// In secondary inventory. If these are needed they must be transferred to the
		// main inventory first
		"pistol":           1,
		"firearm-magazine": 2,
	}

	return s
}
