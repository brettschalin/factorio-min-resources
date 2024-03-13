package calc

import (
	"github.com/brettschalin/factorio-min-resources/building"
	"github.com/brettschalin/factorio-min-resources/constants"
	"github.com/brettschalin/factorio-min-resources/data"
)

// TechEnergyCost returns the energy required for the given lab to research the tech
func TechEnergyCost(lab *building.Lab, tech string) float64 {
	t := data.GetTech(tech)
	e := lab.Entity.EnergyUsage
	time := t.Unit.Time
	n := t.Unit.Count

	return float64(time) * float64(n) * float64(e)
}

// BoilerFuelCost returns the amount of fuel required to create the given amount of energy.
func BoilerFuelCost(boiler *building.Boiler, fuel string, energy float64) float64 {
	item := data.GetItem(fuel)

	// TODO: factor in b.Entity.Effectivity. Vanilla boiler is 1

	// note: panics if the fuel value is zero or the item doesn't exist
	return energy / float64(item.FuelValue)
}

// RecipesFromFuel returns the number of recipes that can be crafted with the given amount of fuel
func RecipesFromFuel(m building.CraftingBuilding, recipe *data.Recipe, fuel float64, fuelType string) float64 {

	// energy required for one smelt
	energy := recipe.CraftingTime() * float64(m.EnergyUsage()) / float64(m.CraftingSpeed())

	fItem := data.GetItem(fuelType)

	ret := (float64(fItem.FuelValue) * fuel) / energy

	return ret
}

// FuelFromRecipes returns the amount of fuel required to craft the given number of recipes. It returns 0
// if the machine's energy source is not chemical
func FuelFromRecipes(m building.CraftingBuilding, recipe *data.Recipe, count int, fuel string) float64 {

	if m.EnergySource().FuelCategory != constants.FuelCategoryChemical {
		return 0
	}

	e := float64(data.GetItem(fuel).FuelValue)
	c := float64(m.EnergyUsage())

	timeToCraft := recipe.CraftingTime() / m.CraftingSpeed()

	energy := timeToCraft * float64(c)

	nFuel := float64(count) * (energy / e)

	return nFuel
}
