package main

import (
	"fmt"
	"math"

	"github.com/brettschalin/factorio-min-resources/building"
	"github.com/brettschalin/factorio-min-resources/calc"
	"github.com/brettschalin/factorio-min-resources/constants"
	"github.com/brettschalin/factorio-min-resources/data"
	"github.com/brettschalin/factorio-min-resources/state"
	"github.com/brettschalin/factorio-min-resources/tas"
)

// prodmod1 returns the tasks required to craft productivity-modules and place them in the relevant machines
func prodmod1(state *state.State, extraFuel float64) (tas.Tasks, float64) {
	tasks, extraFuel := tas.MineFuelAndSmelt("iron-ore", constants.PreferredFuel, state.Furnace, 150, extraFuel)

	t, extraFuel := tas.MineFuelAndSmelt("copper-ore", constants.PreferredFuel, state.Furnace, 237, extraFuel)
	tasks.Add(t...)

	t, extraFuel = tas.MineFuelAndSmelt("iron-plate", constants.PreferredFuel, state.Furnace, 10, extraFuel)
	tasks.Add(t...)

	tasks.Add(tas.MineResource("coal", 35))

	t, _ = tas.MachineCraft("plastic-bar", state.Chem, 35, constants.PreferredFuel)
	tasks.Add(t...)

	craftTasks := tas.Tasks{
		tas.Craft("assembling-machine-2", 1),
		tas.Craft("productivity-module", 7),
	}
	craftTasks[0].Prerequisites().Add(tasks[len(tasks)-1])
	craftTasks[1].Prerequisites().Add(techMap["productivity-module"])
	tasks.Add(craftTasks...)

	t = tas.Tasks{
		tas.Build("assembling-machine-2", 0),
		tas.Transfer(state.Assembler.Name(), "productivity-module", state.Assembler.Slots().Modules, 2, false),
		tas.Transfer(state.Lab.Name(), "productivity-module", state.Lab.Slots().Modules, 2, false),
		tas.Transfer(state.Chem.Name(), "productivity-module", state.Chem.Slots().Modules, 3, false),
	}
	t[0].Prerequisites().Add(tasks[len(tasks)-1])
	tasks.Add(t...)

	mod := data.GetModule("productivity-module")
	state.Assembler.SetModules(building.Modules{mod, mod})
	state.Lab.SetModules(building.Modules{mod, mod})
	state.Chem.SetModules(building.Modules{mod, mod, mod})

	return tasks, extraFuel
}

// buildElectricFurnace returns the tasks required to research and build the electric-furnace
func buildElectricFurnace(state *state.State, extraFuel float64) tas.Tasks {
	tasks := tas.Tasks{}

	tasks.Add(tas.Speed(5))

	toCraft := map[*data.Recipe]int{}
	for _, tech := range []string{"sulfur-processing", "chemical-science-pack", "advanced-material-processing-2"} {
		for pack, amount := range calc.TechCost(tech) {
			toCraft[data.GetRecipe(pack)] += amount
		}
	}

	// apply lab bonuses to the research and convert it to the number of recipes to craft instead of packs needed
	for r, amount := range toCraft {
		p := r.ProductCount(r.Name)
		b := 1 + state.Lab.ProductivityBonus()

		amt := int(math.Ceil(float64(amount) / (float64(p) * b)))

		extra := amt % p
		if extra != 0 {
			amt += (p - extra)
		}
		toCraft[r] = amt
	}

	toCraft[data.GetRecipe("electric-furnace")] = 1

	ings, _ := calc.RecipeAllIngredients(toCraft, state)
	fmt.Println("Required materials")
	for _, ing := range ings {
		fmt.Printf("\t%s: %d\n", ing.Name, ing.Amount)
	}

	t, extraFuel := tas.MineFuelAndSmelt("iron-ore", constants.PreferredFuel, state.Furnace, uint(ings.Amount("iron-ore")), extraFuel)
	tasks.Add(t...)

	t, extraFuel = tas.MineFuelAndSmelt("copper-ore", constants.PreferredFuel, state.Furnace, uint(ings.Amount("copper-ore")), extraFuel)
	tasks.Add(t...)

	t, extraFuel = tas.MineFuelAndSmelt("stone", constants.PreferredFuel, state.Furnace, uint(ings.Amount("stone")), extraFuel)
	tasks.Add(t...)

	t, extraFuel = tas.MineFuelAndSmelt("iron-plate", constants.PreferredFuel, state.Furnace, uint(ings.Amount("steel-plate")*5), extraFuel)
	tasks.Add(t...)

	_ = extraFuel

	// state.Furnace = building.NewFurnace(data.GetFurnace("electric-furnace"))

	return tasks
}
