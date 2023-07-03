package constants

const (
	FuelCategoryChemical = "chemical"
	FuelCategoryElectric = "electric"
	FuelCategoryNuclear  = "nuclear"
)

var StartingInventory = map[string]int{
	"stone-furnace":       1,
	"burner-mining-drill": 1,
	"wood":                1,

	// found in the spaceship wreckage. Will be in the inventory
	// so long as the generated "mine the ship" commands are kept
	"iron-plate": 8,

	// In secondary inventory. If these are needed they must be transferred to the
	// main inventory first
	// "pistol":           1,
	// "firearm-magazine": 2,
}

// config options
const (
	// Set true to use expensive variants of recipes
	UseExpensive = false

	// what fuel the boiler/furnace should use. This is assumed to be minable
	PreferredFuel = "coal"
)
