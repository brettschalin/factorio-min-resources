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

func copyMap[K comparable, V any](m map[K]V) map[K]V {

	out := make(map[K]V, len(m))

	for k, v := range m {
		out[k] = v
	}

	return out
}

func (s *State) Copy() *State {

	var (
		f       building.Furnace
		a, c, r building.Assembler
		b       building.Boiler
		l       building.Lab
		ret     = &State{
			Inventory:      copyMap(s.Inventory),
			TechResearched: copyMap(s.TechResearched),
			Buildings:      copyMap(s.Buildings),
		}
	)

	if s.Furnace != nil {
		f = *s.Furnace
		ret.Furnace = &f
	}

	if s.Assembler != nil {
		a = *s.Assembler
		ret.Assembler = &a
	}

	if s.Chem != nil {
		c = *s.Chem
		ret.Chem = &c
	}

	if s.Refinery != nil {
		r = *s.Refinery
		ret.Refinery = &r
	}

	if s.Boiler != nil {
		b = *s.Boiler
		ret.Boiler = &b
	}

	if s.Lab != nil {
		l = *s.Lab
		ret.Lab = &l
	}

	return ret
}
