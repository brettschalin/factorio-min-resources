package main

import (
	"math"

	"github.com/brettschalin/factorio-min-resources/building"
	"github.com/brettschalin/factorio-min-resources/calc"
	"github.com/brettschalin/factorio-min-resources/constants"
	"github.com/brettschalin/factorio-min-resources/data"
	"github.com/brettschalin/factorio-min-resources/tas"
)

/*
	This file will hold every run segment needed prior to crafting the first productivity module.
	Segments are split into multiple functions to make it easier to work with them
*/

// mines, smelts, crafts, and builds the initial power setup
func makePowerSetup() tas.Tasks {

	tasks := tas.Tasks{
		tas.Build("stone-furnace", constants.DirectionNorth),
		tas.MineResource("coal", 7),
		tas.Transfer("stone-furnace", "coal", constants.InventoryFuel, 7, false),
		tas.MineResource("iron-ore", 50),
		tas.Transfer("stone-furnace", "iron-ore", constants.InventoryFurnaceSource, 50, false),
		tas.MineResource("iron-ore", 18),
		tas.MineResource("copper-ore", 19),
		tas.WaitInventory("stone-furnace", "iron-plate", constants.InventoryFurnaceResult, 50, true),
		tas.Transfer("stone-furnace", "iron-plate", constants.InventoryFurnaceResult, 50, true),

		tas.Transfer("stone-furnace", "iron-ore", constants.InventoryFurnaceSource, 18, false),
		tas.WaitInventory("stone-furnace", "iron-plate", constants.InventoryFurnaceResult, 18, true),
		tas.Transfer("stone-furnace", "iron-plate", constants.InventoryFurnaceResult, 18, true),
		tas.MineResource("stone", 5),

		tas.Transfer("stone-furnace", "copper-ore", constants.InventoryFurnaceSource, 19, false),
		tas.WaitInventory("stone-furnace", "copper-plate", constants.InventoryFurnaceResult, 19, true),
		tas.Transfer("stone-furnace", "copper-plate", constants.InventoryFurnaceResult, 19, true),

		tas.Craft("steam-engine", 1),
		tas.Craft("offshore-pump", 1),
		tas.Craft("lab", 1),
		tas.Craft("small-electric-pole", 1),
		tas.Craft("boiler", 1),

		tas.Build("steam-engine", constants.DirectionEast),
		tas.Build("offshore-pump", constants.DirectionNorth),
		tas.Build("lab", constants.DirectionNorth),
		tas.Build("small-electric-pole", constants.DirectionNorth),
		tas.Build("boiler", constants.DirectionEast),
	}

	// Ensure we have materials before crafting
	tasks[16].Prerequisites().Add(tasks[12])
	tasks[17].Prerequisites().Add(tasks[15])

	// make sure the thing's crafted before building it

	tasks[21].Prerequisites().Add(tasks[16])
	tasks[22].Prerequisites().Add(tasks[17])
	tasks[23].Prerequisites().Add(tasks[18])
	tasks[24].Prerequisites().Add(tasks[19])
	tasks[25].Prerequisites().Add(tasks[20])

	return tasks
}

// researchTech mines and crafts the science packs required to research the
// given technology. Currently only works with red and green science.
// Also note that fuel values are rounded up, and the difference will likely need
// to be subtracted from the next function called
func researchTech(tech string, f *building.Furnace, l *building.Lab, b *building.Boiler) tas.Tasks {

	tasks := tas.Tasks{}

	if b != nil {
		// we need to fuel the boiler
		boilerCoal := uint(math.Ceil(calc.BoilerFuelCost(b, "coal", calc.TechEnergyCost(l, tech))))

		tasks.Add(
			tas.FuelMachine("coal", b.Name(), boilerCoal)...,
		)
	}

	// calculate how much mining we'll need to do
	packs := calc.TechCost(tech)
	baseCost := map[string]int{}
	for p, amt := range packs {
		cost, _ := calc.RecipeFullCost(p, amt)
		for c, n := range cost {
			baseCost[c] += n
		}
	}

	// get the fuel cost for it, and how much of each ore to smelt
	var (
		smeltCoal  float64
		toSmelt    = map[string]uint{}
		smeltProds = map[string]string{} // ore -> plate
	)

	// finding recipes from ingredients is hard, hardcoding is easy. This does only work with
	// red and green science but since we're really only using this for steel and green pack research anyways
	// it'll probably be fine
	toSmelt["iron-ore"] = uint(baseCost["iron-ore"])
	toSmelt["copper-ore"] = uint(baseCost["copper-ore"])
	smeltProds["iron-ore"] = "iron-plate"
	smeltProds["copper-ore"] = "copper-plate"

	for c, n := range toSmelt {
		rec := data.GetRecipe(smeltProds[c])

		if rec == nil || !f.Entity.CanCraft(rec) {
			continue
		}

		smeltCoal += calc.FuelFromRecipes(f, rec, int(n), "coal")
	}

	tasks.Add(tas.FuelMachine("coal", f.Name(), uint(math.Ceil(smeltCoal)))...)

	for ore, n := range toSmelt {
		tasks.Add(tas.MineAndSmelt(ore, smeltProds[ore], f.Name(), n)...)
	}

	// craft the science packs
	first := true
	for p, amt := range packs {
		t := tas.Craft(p, uint(amt))
		if first {
			first = false
			t.Prerequisites().Add(tasks[len(tasks)-1])
		}
		tasks.Add(t)
	}

	lTasks := tas.Tasks{}

	for p, n := range packs {
		t := tas.Transfer(l.Name(), p, constants.InventoryLabInput, uint(n), false)
		t.Prerequisites().Add(tasks[len(tasks)-1])
		lTasks.Add(t)
	}

	tasks.Add(lTasks...)

	return tasks
}

// outputs the tasks needed to research solar-energy and build a solar-panel.
// Assumes logistic-science-packs and steel smelting have been unlocked
func buildSolarPanel(furnace string) tas.Tasks {
	tasks := tas.Tasks{}

	// mine and smelt 1975 iron. Done in batches because of fuel capacity limitations

	tasks.Add(tas.FuelMachine("coal", furnace, 49)...)
	tasks.Add(tas.MineAndSmelt("iron-ore", "iron-plate", furnace, 690)...)

	tasks.Add(tas.FuelMachine("coal", furnace, 49)...)
	tasks.Add(tas.MineAndSmelt("iron-ore", "iron-plate", furnace, 690)...)

	tasks.Add(tas.FuelMachine("coal", furnace, 43)...)
	tasks.Add(tas.MineAndSmelt("iron-ore", "iron-plate", furnace, 595)...)

	c := tas.Craft("iron-gear-wheel", 675)
	c.Prerequisites().Add(tasks[len(tasks)-1])
	tasks.Add(c)
	c = tas.Craft("transport-belt", 125)
	tasks.Add(c)

	// mine and smelt 675 copper
	tasks.Add(tas.FuelMachine("coal", furnace, 49)...)

	smeltTasks := tas.MineAndSmelt("copper-ore", "copper-plate", furnace, 675)

	tasks.Add(smeltTasks...)

	// once we get 300 copper, start crafting red science
	c = tas.Craft("automation-science-pack", 300)
	c.Prerequisites().Add(smeltTasks[24])
	tasks.Add(c)

	c = tas.Craft("logistic-science-pack", 50)
	c.Prerequisites().Add(smeltTasks[len(smeltTasks)-1])
	tasks.Add(c)
	tasks.Add(tas.Craft("logistic-science-pack", 200))

	// Dumping the science packs all at once causes a rounding error that leaves solar-energy unresearched with
	// a tiny fraction of a green pack remaining. Not sure why but this seems to fix it
	tasks.Add(
		tas.MineResource("coal", 36),
		tas.Transfer("boiler", "coal", constants.InventoryFuel, 36, false),

		tas.WaitInventory("player", "automation-science-pack", constants.InventoryCharacterMain, 10, false),
		tas.Transfer("lab", "automation-science-pack", constants.InventoryLabInput, 10, false),

		tas.WaitInventory("player", "automation-science-pack", constants.InventoryCharacterMain, 30, false),
		tas.Transfer("lab", "automation-science-pack", constants.InventoryLabInput, 30, false),

		tas.WaitInventory("player", "automation-science-pack", constants.InventoryCharacterMain, 10, false),
		tas.Transfer("lab", "automation-science-pack", constants.InventoryLabInput, 10, false),

		tas.WaitInventory("player", "logistic-science-pack", constants.InventoryCharacterMain, 50, false),
	)
	tasks[len(tasks)-5].Prerequisites().Add(techMap["automation"])
	tasks[len(tasks)-3].Prerequisites().Add(techMap["electronics"])
	tasks[len(tasks)-1].Prerequisites().Add(techMap["optics"])

	tasks.Add(

		tas.Transfer("lab", "automation-science-pack", constants.InventoryLabInput, 50, false),
		tas.Transfer("lab", "logistic-science-pack", constants.InventoryLabInput, 50, false),

		tas.MineResource("coal", 37),
		//tas.Walk(geo.Point{-20, -15}),
		tas.WaitInventory("boiler", "coal", constants.InventoryFuel, 0, true),
		tas.Transfer("boiler", "coal", constants.InventoryFuel, 37, false),

		tas.WaitInventory("lab", "automation-science-pack", constants.InventoryLabInput, 0, true),
		tas.Transfer("lab", "automation-science-pack", constants.InventoryLabInput, 50, false),

		tas.WaitInventory("player", "logistic-science-pack", constants.InventoryCharacterMain, 50, false),
		tas.Transfer("lab", "logistic-science-pack", constants.InventoryLabInput, 50, false),

		tas.WaitInventory("lab", "automation-science-pack", constants.InventoryLabInput, 0, true),
		tas.Transfer("lab", "automation-science-pack", constants.InventoryLabInput, 150, false),
		tas.WaitInventory("lab", "logistic-science-pack", constants.InventoryLabInput, 0, true),
		tas.Transfer("lab", "logistic-science-pack", constants.InventoryLabInput, 150, false),

		tas.MineResource("coal", 50),
		//tas.Walk(geo.Point{-20, -15}),
		tas.WaitInventory("boiler", "coal", constants.InventoryFuel, 0, true),
		tas.Transfer("boiler", "coal", constants.InventoryFuel, 50, false),
	)

	tasks.Add(tas.FuelMachine("coal", furnace, 6)...)
	tasks.Add(tas.MineAndSmelt("iron-ore", "iron-plate", furnace, 40)...)
	tasks.Add(tas.MineAndSmelt("copper-ore", "copper-plate", furnace, 28)...)
	tasks.Add(
		tas.Transfer(furnace, "iron-plate", constants.InventoryFurnaceSource, 25, false),
		tas.WaitInventory(furnace, "steel-plate", constants.InventoryFurnaceResult, 5, true),
		tas.Transfer(furnace, "steel-plate", constants.InventoryFurnaceResult, 5, true),
	)

	tasks.Add(
		tas.Craft("solar-panel", 1),
		tas.MineEntity("steam-engine"),
		tas.MineEntity("boiler"),
		tas.Build("solar-panel", constants.DirectionNone),
	)
	tasks[len(tasks)-4].Prerequisites().Add(techMap["solar-energy"])
	tasks[len(tasks)-3].Prerequisites().Add(techMap["solar-energy"])
	tasks[len(tasks)-1].Prerequisites().Add(tasks[len(tasks)-4])

	return tasks
}

// outputs the tasks needed to research advanced-material-processing and build a steel-furnace.
// Assumes logistic-science-packs and steel smelting have been unlocked
func buildSteelFurnace(hasSolar bool) tas.Tasks {

	tasks := tas.Tasks{}

	var c tas.Task

	tasks.Add(tas.FuelMachine("coal", "stone-furnace", 41)...)

	tasks.Add(tas.MineAndSmelt("iron-ore", "iron-plate", "stone-furnace", 564)...)

	c = tas.Craft("transport-belt", 38)
	c.Prerequisites().Add(tasks[len(tasks)-1])

	tasks.Add(c)
	tasks.Add(
		tas.FuelMachine("coal", "stone-furnace", 13)...,
	)

	// we have one extra copper-cable from the solar panel making, and that's enough to take this
	// from 188 to 187
	tasks.Add(tas.MineAndSmelt("copper-ore", "copper-plate", "stone-furnace", 187)...)

	t := tas.Tasks{
		tas.Transfer("lab", "automation-science-pack", constants.InventoryLabInput, 75, false),
		tas.Transfer("lab", "logistic-science-pack", constants.InventoryLabInput, 75, false),
	}
	c = tas.Craft("automation-science-pack", 75)
	c.Prerequisites().Add(tasks[len(tasks)-8])
	t[0].Prerequisites().Add(c)
	tasks.Add(c)

	c = tas.Craft("inserter", 75)
	c.Prerequisites().Add(tasks[len(tasks)-2])
	tasks.Add(c)

	c = tas.Craft("logistic-science-pack", 75)
	t[1].Prerequisites().Add(c)
	c.Prerequisites().Add(techMap["logistic-science-pack"])
	tasks.Add(c)

	if !hasSolar {
		tasks.Add(
			tas.FuelMachine("coal", "boiler", 34)...,
		)
	}

	tasks.Add(t...)

	tasks.Add(
		tas.FuelMachine("coal", "stone-furnace", 5)...,
	)

	tasks.Add(
		tas.MineResource("iron-ore", 30),
		tas.Transfer("stone-furnace", "iron-ore", constants.InventoryFurnaceSource, 30, false),
		tas.MineResource("stone", 20),

		tas.WaitInventory("stone-furnace", "iron-plate", constants.InventoryFurnaceResult, 30, true),
		tas.Transfer("stone-furnace", "iron-plate", constants.InventoryFurnaceResult, 30, true),

		tas.Transfer("stone-furnace", "stone", constants.InventoryFurnaceSource, 20, false),
		tas.WaitInventory("stone-furnace", "stone-brick", constants.InventoryFurnaceResult, 10, true),
		tas.Transfer("stone-furnace", "stone-brick", constants.InventoryFurnaceResult, 10, true),

		tas.Transfer("stone-furnace", "iron-plate", constants.InventoryFurnaceSource, 30, false),
		tas.WaitInventory("stone-furnace", "steel-plate", constants.InventoryFurnaceResult, 6, true),
		tas.Transfer("stone-furnace", "steel-plate", constants.InventoryFurnaceResult, 6, true),
	)

	c = tas.Craft("steel-furnace", 1)
	c.Prerequisites().Add(techMap["advanced-material-processing"])

	tasks.Add(
		tas.Speed(1),
		c,
		// there's about 10% of a coal left in this, hopefully that doesn't
		// come back to bite us. Done this way because fast replace didn't work and just
		// made two entites overlapping the same space instead
		tas.MineEntity("stone-furnace"),
	)

	c = tas.Build("steel-furnace", constants.DirectionNorth)
	c.Prerequisites().Add(tasks[len(tasks)-2:]...)
	tasks.Add(c)

	return tasks
}
