package main

import (
	"math"

	"github.com/brettschalin/factorio-min-resources/building"
	"github.com/brettschalin/factorio-min-resources/calc"
	"github.com/brettschalin/factorio-min-resources/constants"
	"github.com/brettschalin/factorio-min-resources/shims"
	"github.com/brettschalin/factorio-min-resources/tas"
)

/*
This file will hold every run segment needed prior to crafting the first productivity module.
Segments are split into multiple functions to make it easier to work with them
*/

func playerHasItem(item string, amount uint) tas.Task {
	return tas.PrereqWait("player", item, constants.InventoryCharacterMain, amount, false)
}

// mines, smelts, crafts, and builds the initial power setup, and returns the amount of "leftover" fuel remaining
func makePowerSetup(furnace *building.Furnace) (tas.Tasks, float64) {

	tasks := tas.Tasks{
		tas.Build("stone-furnace", 0),
	}

	// craft these whenever we get the needed materials
	craftTasks := tas.Tasks{
		tas.Craft("steam-engine", 1),
		tas.Craft("boiler", 1),
		tas.Craft("offshore-pump", 1),
		tas.Craft("lab", 1),
		tas.Craft("small-electric-pole", 1),
	}

	t, extraFuel := tas.MineFuelAndSmelt("iron-ore", constants.PreferredFuel, furnace, 68, 0)
	tasks.Add(t...)

	s := tas.MineResource("stone", 5)
	tasks.Add(s)

	craftTasks[0].Prerequisites().Add(playerHasItem("iron-plate", 31))
	craftTasks[1].Prerequisites().Add(s, playerHasItem("iron-plate", 4), playerHasItem("stone", 5))
	craftTasks[2].Prerequisites().Add(playerHasItem("iron-plate", 5), playerHasItem("copper-plate", 3))
	craftTasks[3].Prerequisites().Add(playerHasItem("iron-plate", 36), playerHasItem("copper-plate", 15))
	craftTasks[4].Prerequisites().Add(playerHasItem("copper-plate", 1)) // we start with one wood piece

	t, extraFuel = tas.MineFuelAndSmelt("copper-ore", constants.PreferredFuel, furnace, 19, extraFuel)
	tasks.Add(t...)

	tasks.Add(craftTasks...)

	buildTasks := tas.Tasks{
		tas.Build("steam-engine", 0),
		tas.Build("boiler", 0),
		tas.Build("offshore-pump", 0),
		tas.Build("lab", 0),
		tas.Build("small-electric-pole", 1),
		tas.Build("small-electric-pole", 2),
	}
	for i := range craftTasks {
		buildTasks[i].Prerequisites().Add(craftTasks[i])
	}
	buildTasks[5].Prerequisites().Add(craftTasks[4])

	tasks.Add(buildTasks...)

	return tasks, extraFuel
}

// researchRGTech mines and crafts the science packs required to research the
// given technology. Only works with red and green science technologies that take
// <= 200 science packs
func researchRGTech(tech string, f *building.Furnace, l *building.Lab, b *building.Boiler, extraFuel float64) (tas.Tasks, float64) {

	tasks := tas.Tasks{}

	// calculate how much mining we'll need to do
	packs := calc.TechCost(tech)
	baseCost := map[string]int{}
	for p, amt := range packs {
		cost, _ := calc.RecipeFullCost(p, amt)
		for c, n := range cost {
			baseCost[c] += n
		}
	}

	var st tas.Tasks
	st, extraFuel = tas.MineFuelAndSmelt("iron-ore", constants.PreferredFuel, f, uint(baseCost["iron-ore"]), extraFuel)
	tasks.Add(st...)

	st, extraFuel = tas.MineFuelAndSmelt("copper-ore", constants.PreferredFuel, f, uint(baseCost["copper-ore"]), extraFuel)
	tasks.Add(st...)

	// craft the science packs. This at least starts crafting when the iron is available
	t := tas.Craft("iron-gear-wheel", uint(packs["automation-science-pack"]))
	t.Prerequisites().Add(playerHasItem("iron-plate", uint(baseCost["iron-ore"])))
	tasks.Add(t)
	t = tas.Craft("automation-science-pack", uint(packs["automation-science-pack"]))
	t.Prerequisites().Add(tasks[len(tasks)-1], tasks[len(tasks)-2])
	tasks.Add(t)

	lTasks := tas.Tasks{
		tas.Transfer(l.Name(), "automation-science-pack", constants.InventoryLabInput, uint(uint(packs["automation-science-pack"])), false),
	}
	lTasks[0].Prerequisites().Add(tasks[len(tasks)-1], tas.PrereqWait(l.Name(), "automation-science-pack", l.Slots().Input, 0, true))

	if b != nil {
		// we need to fuel the boiler
		boilerCoal := uint(math.Ceil(calc.BoilerFuelCost(b, constants.PreferredFuel, calc.TechEnergyCost(l, tech))))

		tasks.Add(tas.FuelMachine(constants.PreferredFuel, b.Name(), boilerCoal)...)
	}

	if packs["logistic-science-pack"] > 0 {
		tasks.Add(tas.Craft("logistic-science-pack", uint(packs["logistic-science-pack"])))
		t := tas.Transfer(l.Name(), "logistic-science-pack", constants.InventoryLabInput, uint(uint(packs["logistic-science-pack"])), false)
		t.Prerequisites().Add(tasks[len(tasks)-1], tas.PrereqWait(l.Name(), "logistic-science-pack", l.Slots().Input, 0, true))

		lTasks.Add(t)
	}

	tasks.Add(lTasks...)

	return tasks, extraFuel
}

// outputs the tasks needed to research solar-energy and build a solar-panel.
// Assumes logistic-science-packs and steel smelting have been unlocked. This is separate from
// the rest of the red/green tech tasks because `solar-energy` requires 250 science packs and researchRGTech can't handle batching transfers
func buildSolarPanel(furnace *building.Furnace, lab *building.Lab, boiler *building.Boiler, extraFuel float64) (tas.Tasks, float64) {

	tasks := tas.Tasks{}

	// this hopefully fixes a rounding error. Yay for floating point math...
	extraFuel = shims.Max(0, extraFuel-0.0001)

	// smelting
	var st tas.Tasks
	st, extraFuel = tas.MineFuelAndSmelt("iron-ore", constants.PreferredFuel, furnace, 1875, extraFuel)
	tasks.Add(st...)

	st, extraFuel = tas.MineFuelAndSmelt("copper-ore", constants.PreferredFuel, furnace, 625, extraFuel)
	tasks.Add(st...)

	c := tas.Craft("iron-gear-wheel", 625)
	c.Prerequisites().Add(playerHasItem("iron-plate", 1250))
	tasks.Add(c)

	c = tas.Craft("transport-belt", 125)
	c.Prerequisites().Add(playerHasItem("iron-plate", 125))
	c.Prerequisites().Add(playerHasItem("iron-gear-wheel", 125))
	tasks.Add(c)

	// craft science while we're still smelting
	c = tas.Craft("automation-science-pack", 250)
	c.Prerequisites().Add(playerHasItem("copper-plate", 250))
	tasks.Add(c)

	c = tas.Craft("logistic-science-pack", 50)
	c.Prerequisites().Add(playerHasItem("transport-belt", 250))
	c.Prerequisites().Add(playerHasItem("copper-plate", 375))
	tasks.Add(c)
	tasks.Add(tas.Craft("logistic-science-pack", 200))

	// do the research. This requires a careful balancing of boiler fuel and science packs
	tasks.Add(
		// Research requires 112.5 coal worth of energy, but all of the previous research was done with researchRGTech which rounds
		// up any fractional requirements, so that means we can mine less here
		tas.MineResource(constants.PreferredFuel, 111),
		tas.WaitInventory(boiler.Name(), constants.PreferredFuel, constants.InventoryFuel, 0, true),
		tas.Transfer(boiler.Name(), constants.PreferredFuel, constants.InventoryFuel, 50, false),

		tas.WaitInventory("player", "automation-science-pack", constants.InventoryCharacterMain, 50, false),
		tas.Transfer(lab.Name(), "automation-science-pack", constants.InventoryLabInput, 50, false),
		tas.WaitInventory("player", "logistic-science-pack", constants.InventoryCharacterMain, 50, false),
		tas.Transfer(lab.Name(), "logistic-science-pack", constants.InventoryLabInput, 50, false),

		tas.WaitInventory("player", "automation-science-pack", constants.InventoryCharacterMain, 200, false),
		tas.WaitInventory("lab", "automation-science-pack", constants.InventoryLabInput, 0, true),
		tas.Transfer(lab.Name(), "automation-science-pack", constants.InventoryLabInput, 200, false),
		tas.WaitInventory("player", "logistic-science-pack", constants.InventoryCharacterMain, 200, false),
		tas.WaitInventory("lab", "logistic-science-pack", constants.InventoryLabInput, 0, true),
		tas.Transfer(lab.Name(), "logistic-science-pack", constants.InventoryLabInput, 200, false),

		tas.WaitInventory(boiler.Name(), constants.PreferredFuel, constants.InventoryFuel, 0, true),
		tas.Transfer(boiler.Name(), constants.PreferredFuel, constants.InventoryFuel, 50, false),

		tas.WaitInventory(boiler.Name(), constants.PreferredFuel, constants.InventoryFuel, 0, true),
		tas.Transfer(boiler.Name(), constants.PreferredFuel, constants.InventoryFuel, 11, false),
	)

	t, extraFuel := tas.MineFuelAndSmelt("iron-ore", constants.PreferredFuel, furnace, 40, extraFuel)
	tasks.Add(t...)

	t, extraFuel = tas.MineFuelAndSmelt("copper-ore", constants.PreferredFuel, furnace, 28, extraFuel)
	tasks.Add(t...)

	c = tas.Craft("electronic-circuit", 15)
	c.Prerequisites().Add(tasks[len(tasks)-1])
	tasks.Add(c)

	t, extraFuel = tas.MineFuelAndSmelt("iron-plate", constants.PreferredFuel, furnace, 25, extraFuel)

	tasks.Add(t...)

	tasks.Add(
		tas.Craft("solar-panel", 1),
		tas.MineEntity("steam-engine", 0),
		// boiler's needed for water flow, don't mine it
		tas.Build("solar-panel", 0),
	)
	tasks[len(tasks)-3].Prerequisites().Add(techMap["solar-energy"])
	tasks[len(tasks)-2].Prerequisites().Add(techMap["solar-energy"])
	tasks[len(tasks)-1].Prerequisites().Add(tasks[len(tasks)-3])

	return tasks, extraFuel
}

// outputs the tasks needed to research advanced-material-processing and build a steel-furnace.
// Assumes logistic-science-packs and steel smelting have been unlocked
func buildSteelFurnace(furnace *building.Furnace, lab *building.Lab, boiler *building.Boiler, extraFuel float64) (tas.Tasks, float64) {

	tasks := tas.Tasks{}

	var c tas.Task

	tasks, extraFuel = tas.MineFuelAndSmelt("iron-ore", constants.PreferredFuel, furnace, 564, extraFuel)

	// we need 188 but the extra copper-cable left over from the solar panel is enough
	st, extraFuel := tas.MineFuelAndSmelt("copper-ore", constants.PreferredFuel, furnace, 187, extraFuel)
	tasks.Add(st...)

	c = tas.Craft("transport-belt", 38)
	c.Prerequisites().Add(playerHasItem("iron-plate", 114))
	tasks.Add(c)

	t := tas.Tasks{
		tas.Transfer("lab", "automation-science-pack", constants.InventoryLabInput, 75, false),
		tas.Transfer("lab", "logistic-science-pack", constants.InventoryLabInput, 75, false),
	}
	c = tas.Craft("automation-science-pack", 75)
	c.Prerequisites().Add(playerHasItem("iron-plate", 150))
	c.Prerequisites().Add(playerHasItem("copper-plate", 75))
	t[0].Prerequisites().Add(c)
	tasks.Add(c)

	c = tas.Craft("inserter", 75)
	c.Prerequisites().Add(playerHasItem("copper-plate", 112))
	tasks.Add(c)

	c = tas.Craft("logistic-science-pack", 75)
	c.Prerequisites().Add(
		playerHasItem("transport-belt", 75),
		playerHasItem("inserter", 75),
	)
	t[1].Prerequisites().Add(c)
	c.Prerequisites().Add(techMap["logistic-science-pack"])
	tasks.Add(c)

	if boiler != nil {
		tasks.Add(
			tas.FuelMachine(constants.PreferredFuel, "boiler", 34)...,
		)
	}

	tasks.Add(t...)

	st, extraFuel = tas.MineFuelAndSmelt("iron-ore", constants.PreferredFuel, furnace, 30, extraFuel)
	tasks.Add(st...)

	st, extraFuel = tas.MineFuelAndSmelt("stone", constants.PreferredFuel, furnace, 20, extraFuel)
	tasks.Add(st...)

	st, _ = tas.MineFuelAndSmelt("iron-plate", constants.PreferredFuel, furnace, 30, extraFuel)
	tasks.Add(st...)

	c = tas.Craft("steel-furnace", 1)
	c.Prerequisites().Add(techMap["advanced-material-processing"])

	// there's about 10% of a coal left in this, hopefully that doesn't
	// come back to bite us. Done this way because fast replace didn't work and just
	// made two entites overlapping the same space instead
	m := tas.MineEntity("stone-furnace", 0)
	m.Prerequisites().Add(c)

	tasks.Add(
		c,
		m,
	)

	c = tas.Build("steel-furnace", 0)
	c.Prerequisites().Add(playerHasItem("steel-furnace", 1))
	tasks.Add(c)

	// furnace is replaced and the extra fuel is reset
	return tasks, 0
}

func buildOilSetup(f *building.Furnace, extraFuel float64) (tas.Tasks, float64) {
	tasks := tas.Tasks{}

	// how many pipes to build in the map. See locations.lua for where they go
	const nPipes = 38

	t, extraFuel := tas.MineFuelAndSmelt("iron-ore", constants.PreferredFuel, f, 220+nPipes, extraFuel)
	tasks.Add(t...)
	t, extraFuel = tas.MineFuelAndSmelt("copper-ore", constants.PreferredFuel, f, 30, extraFuel)
	tasks.Add(t...)

	task := tas.Craft("electronic-circuit", 20)
	task.Prerequisites().Add(tasks[len(tasks)-1])
	tasks.Add(task)
	tasks.Add(
		tas.Craft("iron-gear-wheel", 25),
		tas.Craft("pipe", 25+nPipes),
	)

	t, extraFuel = tas.MineFuelAndSmelt("stone", constants.PreferredFuel, f, 20, extraFuel)
	tasks.Add(t...)

	t, extraFuel = tas.MineFuelAndSmelt("iron-plate", constants.PreferredFuel, f, 125, extraFuel)
	tasks.Add(t...)

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

	return tasks, extraFuel
}

func researchModules(f *building.Furnace, l *building.Lab, b *building.Boiler, extraFuel float64) (tas.Tasks, float64) {
	tasks := tas.Tasks{}

	t, extraFuel := researchRGTech("plastics", f, l, b, extraFuel)
	tasks.Add(t...)

	t, extraFuel = researchRGTech("advanced-electronics", f, l, b, extraFuel)
	tasks.Add(t...)

	t, extraFuel = researchRGTech("modules", f, l, b, extraFuel)
	tasks.Add(t...)

	t, extraFuel = researchRGTech("productivity-module", f, l, b, extraFuel)
	tasks.Add(t...)

	return tasks, extraFuel
}
