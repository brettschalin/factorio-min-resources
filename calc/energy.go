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
func BoilerFuelCost(boiler building.Boiler, fuel string, energy float64) float64 {
	item := data.GetItem(fuel)

	// TODO: factor in b.Entity.Effectivity. Vanilla boiler is 1

	// note: panics if the fuel value is zero or the item doesn't exist
	return energy / float64(item.FuelValue)
}

// // RecipesFromFuel returns the number of recipes that can be crafted with the given amount of fuel
// func RecipesFromFuel(f *building.Furnace, recipe *data.Recipe, fuel int, fuelType string) int {
// 			// TODO: we most likely don't need this yet, maybe ever
// }

// FuelFromRecipes returns the amount of fuel required to craft the given number of recipes.
// This only accepts furnaces for now since that's what works for vanilla but shouldn't
// be too difficult to extend to other crafting machines
func FuelFromRecipes(f *building.Furnace, recipe *data.Recipe, count int, fuel string) float64 {
	// electric furnaces don't have a fuel slot
	if f.Entity.EnergySource.FuelCategory != constants.FuelCategoryChemical {
		return 0
	}

	e := float64(data.GetItem(fuel).FuelValue)
	c := float64(f.Entity.EnergyUsage)

	timeToCraft := recipe.CraftingTime() / f.Entity.CraftingSpeed

	energy := timeToCraft * float64(c)

	nFuel := float64(count) * (energy / e)

	return nFuel
}
