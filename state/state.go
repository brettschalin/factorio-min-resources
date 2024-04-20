package state

import (
	"github.com/brettschalin/factorio-min-resources/building"
	"github.com/brettschalin/factorio-min-resources/constants"
	"github.com/brettschalin/factorio-min-resources/data"
	"github.com/brettschalin/factorio-min-resources/shims/slices"
)

type State struct {
	// character main inventory
	Inventory map[string]uint

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
	s := &State{
		TechResearched: make(map[string]bool),
		Buildings:      make(map[string]bool),
	}

	// Starting inventory
	s.Inventory = copyMap(constants.StartingInventory)
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

func (s *State) GetProductivityBonus(recipe *data.Recipe) float64 {
	if s.Furnace != nil && s.Furnace.Entity.CanCraft(recipe) {
		return s.Furnace.ProductivityBonus(recipe.Name)
	}
	if s.Assembler != nil && s.Assembler.Entity.CanCraft(recipe) {
		return s.Assembler.ProductivityBonus(recipe.Name)
	}
	if s.Chem != nil && s.Chem.Entity.CanCraft(recipe) {
		return s.Chem.ProductivityBonus(recipe.Name)
	}
	if s.Refinery != nil && s.Refinery.Entity.CanCraft(recipe) {
		return s.Refinery.ProductivityBonus(recipe.Name)
	}
	return 0
}

// Construct a building. Returns whether it could be placed
func (s *State) ConstructBuilding(name string) bool {

	ok := true

	switch {
	case slices.Contains(constants.Furnaces, name):
		if s.Furnace != nil {
			return false
		}
		s.Furnace = building.NewFurnace(data.GetFurnace(name))
		ok = s.Furnace != nil

	case slices.Contains(constants.AssemblingMachines, name):
		if s.Assembler != nil {
			return false
		}
		s.Assembler = building.NewAssembler(data.GetAssemblingMachine(name))
		ok = s.Assembler != nil

	case slices.Contains(constants.ChemicalPlants, name):
		if s.Chem != nil {
			return false
		}
		s.Chem = building.NewAssembler(data.GetAssemblingMachine(name))
		ok = s.Chem != nil

	case slices.Contains(constants.Refineries, name):
		if s.Refinery != nil {
			return false
		}
		s.Refinery = building.NewAssembler(data.GetAssemblingMachine(name))
		ok = s.Refinery != nil

	case slices.Contains(constants.Labs, name):
		if s.Lab != nil {
			return false
		}
		s.Lab = building.NewLab(data.GetLab(name))
		ok = s.Lab != nil

	case slices.Contains(constants.Boilers, name):
		if s.Boiler != nil {
			return false
		}
		s.Boiler = building.NewBoiler(data.GetBoiler(name))
		ok = s.Boiler != nil
	}
	return ok
}

// Mine a building. Returns whether it could be mined
func (s *State) MineBuilding(name string) bool {
	if slices.Contains(constants.Furnaces, name) {
		if s.Furnace == nil {
			return false
		}
		s.Furnace = nil
		return true
	}

	if slices.Contains(constants.AssemblingMachines, name) {
		if s.Assembler == nil {
			return false
		}
		s.Assembler = nil
		return true
	}

	if slices.Contains(constants.ChemicalPlants, name) {
		if s.Chem == nil {
			return false
		}
		s.Chem = nil
		return true
	}

	if slices.Contains(constants.Refineries, name) {
		if s.Refinery == nil {
			return false
		}
		s.Refinery = nil
		return true
	}

	if slices.Contains(constants.Labs, name) {
		if s.Lab == nil {
			return false
		}
		s.Lab = nil
		return true
	}

	if slices.Contains(constants.Boilers, name) {
		if s.Boiler == nil {
			return false
		}
		s.Boiler = nil
		return true
	}

	return true
}

func (s *State) GetBuilding(name string) building.Building {
	switch {
	case s.Assembler != nil && s.Assembler.Name() == name:
		return s.Assembler
	case s.Chem != nil && s.Chem.Name() == name:
		return s.Chem
	case s.Furnace != nil && s.Furnace.Name() == name:
		return s.Furnace
	case s.Refinery != nil && s.Refinery.Name() == name:
		return s.Refinery
	case s.Lab != nil && s.Lab.Name() == name:
		return s.Lab
	case s.Boiler != nil && s.Boiler.Name() == name:
		return s.Boiler
	}

	return nil
}
