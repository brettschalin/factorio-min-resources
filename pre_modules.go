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
		tas.Build("stone-furnace", 0),
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

		tas.Build("steam-engine", 0),
		tas.Build("offshore-pump", 0),
		tas.Build("lab", 0),
		tas.Build("small-electric-pole", 1),
		tas.Build("small-electric-pole", 2),
		tas.Build("boiler", 0),
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

	// get the fuel cost, and how much of each ore to smelt
	var smeltCoal float64
	rec := data.GetRecipe("iron-plate")
	smeltCoal += calc.FuelFromRecipes(f, rec, baseCost["iron-ore"], "coal")
	rec = data.GetRecipe("copper-plate")
	smeltCoal += calc.FuelFromRecipes(f, rec, baseCost["copper-ore"], "coal")

	tasks.Add(tas.FuelMachine("coal", f.Name(), uint(math.Ceil(smeltCoal)))...)

	// finding smelt products from ores is hard, hardcoding is easy. As a consequence this method only works
	// with red and green science but those are the early/mid technologies anyways so it'll be fine
	tasks.Add(tas.MineAndSmelt("iron-ore", f.Name(), uint(baseCost["iron-ore"]))...)
	tasks.Add(tas.MineAndSmelt("copper-ore", f.Name(), uint(baseCost["copper-ore"]))...)

	// craft the science packs
	t := tas.Craft("automation-science-pack", uint(packs["automation-science-pack"]))
	t.Prerequisites().Add(tasks[len(tasks)-1])
	tasks.Add(t)

	lTasks := tas.Tasks{
		tas.Transfer(l.Name(), "automation-science-pack", constants.InventoryLabInput, uint(uint(packs["automation-science-pack"])), false),
	}
	lTasks[0].Prerequisites().Add(tasks[len(tasks)-1])

	if packs["logistic-science-pack"] > 0 {
		tasks.Add(tas.Craft("logistic-science-pack", uint(packs["logistic-science-pack"])))
		t := tas.Transfer(l.Name(), "logistic-science-pack", constants.InventoryLabInput, uint(uint(packs["logistic-science-pack"])), false)
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
	tasks.Add(tas.MineAndSmelt("iron-ore", furnace, 690)...)

	tasks.Add(tas.FuelMachine("coal", furnace, 49)...)
	tasks.Add(tas.MineAndSmelt("iron-ore", furnace, 690)...)

	tasks.Add(tas.FuelMachine("coal", furnace, 43)...)
	tasks.Add(tas.MineAndSmelt("iron-ore", furnace, 595)...)

	c := tas.Craft("iron-gear-wheel", 675)
	c.Prerequisites().Add(tasks[len(tasks)-1])
	tasks.Add(c)
	c = tas.Craft("transport-belt", 125)
	tasks.Add(c)

	// mine and smelt 675 copper
	tasks.Add(tas.FuelMachine("coal", furnace, 49)...)

	smeltTasks := tas.MineAndSmelt("copper-ore", furnace, 675)

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
	tasks.Add(tas.MineAndSmelt("iron-ore", furnace, 40)...)
	tasks.Add(tas.MineAndSmelt("copper-ore", furnace, 28)...)
	tasks.Add(
		tas.Transfer(furnace, "iron-plate", constants.InventoryFurnaceSource, 25, false),
		tas.WaitInventory(furnace, "steel-plate", constants.InventoryFurnaceResult, 5, true),
		tas.Transfer(furnace, "steel-plate", constants.InventoryFurnaceResult, 5, true),
	)

	tasks.Add(
		tas.Craft("solar-panel", 1),
		tas.MineEntity("steam-engine", 0),
		// boiler's needed for water flow, don't mine it
		tas.Build("solar-panel", 0),
	)
	tasks[len(tasks)-3].Prerequisites().Add(techMap["solar-energy"])
	tasks[len(tasks)-2].Prerequisites().Add(techMap["solar-energy"])
	tasks[len(tasks)-1].Prerequisites().Add(tasks[len(tasks)-3])

	return tasks
}

// outputs the tasks needed to research advanced-material-processing and build a steel-furnace.
// Assumes logistic-science-packs and steel smelting have been unlocked
func buildSteelFurnace(hasSolar bool) tas.Tasks {

	tasks := tas.Tasks{}

	var c tas.Task

	tasks.Add(tas.FuelMachine("coal", "stone-furnace", 41)...)

	tasks.Add(tas.MineAndSmelt("iron-ore", "stone-furnace", 564)...)

	c = tas.Craft("transport-belt", 38)
	c.Prerequisites().Add(tasks[len(tasks)-1])

	tasks.Add(c)
	tasks.Add(
		tas.FuelMachine("coal", "stone-furnace", 13)...,
	)

	// we have one extra copper-cable from the solar panel making, and that's enough to take this
	// from 188 to 187
	tasks.Add(tas.MineAndSmelt("copper-ore", "stone-furnace", 187)...)

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
		c,
		// there's about 10% of a coal left in this, hopefully that doesn't
		// come back to bite us. Done this way because fast replace didn't work and just
		// made two entites overlapping the same space instead
		tas.MineEntity("stone-furnace", 0),
	)

	c = tas.Build("steel-furnace", 0)
	c.Prerequisites().Add(tasks[len(tasks)-2:]...)
	tasks.Add(c)

	return tasks
}

func buildOilSetup(f *building.Furnace) tas.Tasks {
	tasks := tas.Tasks{}

	// how many pipes to build in the map. See locations.lua for where they go
	const nPipes = 38

	tasks.Add(tas.FuelMachine("coal", f.Name(), 15)...)

	tasks.Add(tas.MineAndSmelt("iron-ore", f.Name(), 220+nPipes)...)
	tasks.Add(tas.MineAndSmelt("copper-ore", f.Name(), 30)...)

	t := tas.Craft("electronic-circuit", 20)
	t.Prerequisites().Add(tasks[len(tasks)-1])
	tasks.Add(t)
	tasks.Add(
		tas.Craft("iron-gear-wheel", 25),
		tas.Craft("pipe", 25+nPipes),
	)

	tasks.Add(tas.MineAndSmelt("stone", f.Name(), 20)...)

	// smelt 25 steel. Done in batches
	tasks.Add(tas.MachineCraft("steel-plate", f.Name(), 20)...)
	tasks.Add(tas.MachineCraft("steel-plate", f.Name(), 5)...)

	tasks.Add(
		tas.Craft("oil-refinery", 1),
		tas.Craft("chemical-plant", 1),
		tas.Craft("pumpjack", 1),
	)

	tasks[len(tasks)-3].Prerequisites().Add(tasks[len(tasks)-4], techMap["oil-processing"])

	bTasks := tas.Tasks{tas.Build("pumpjack", 0)}
	bTasks[0].Prerequisites().Add(tasks[len(tasks)-1])
	bTasks.Add(
		tas.Build("chemical-plant", 0),
		tas.Build("oil-refinery", 0),
	)
	for i := 1; i <= nPipes; i++ {
		bTasks.Add(tas.Build("pipe", i))
	}

	tasks.Add(bTasks...)

	return tasks
}
