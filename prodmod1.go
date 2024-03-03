package main

import (
	"github.com/brettschalin/factorio-min-resources/constants"
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

	return tasks, extraFuel
}
