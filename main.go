package main

import (
	"os"

	"github.com/brettschalin/factorio-min-resources/constants"
	"github.com/brettschalin/factorio-min-resources/data"
	"github.com/brettschalin/factorio-min-resources/tas"
)

func main() {
	var err error
	if err = data.Init(
		"./data/data-raw-dump.json",
	); err != nil {
		panic(err)
	}

	t := tas.TAS{}

	tasks := tas.Tasks{
		tas.Tech("automation"),
		tas.Build("stone-furnace", constants.DirectionNorth),
		tas.MineResource("coal", 7),
		tas.Transfer("stone-furnace", "coal", constants.InventoryFuel, 7, false),
		tas.MineResource("iron-ore", 50),
		tas.Transfer("stone-furnace", "iron-ore", constants.InventoryFurnaceSource, 50, false),
		tas.MineResource("iron-ore", 18),
		tas.MineResource("copper-ore", 19),
		tas.WaitInventory("stone-furnace", "iron-plate", constants.InventoryFurnaceResult, 50),
		tas.Transfer("stone-furnace", "iron-plate", constants.InventoryFurnaceResult, 50, true),

		tas.Transfer("stone-furnace", "iron-ore", constants.InventoryFurnaceSource, 18, false),
		tas.WaitInventory("stone-furnace", "iron-plate", constants.InventoryFurnaceResult, 18),
		tas.Transfer("stone-furnace", "iron-plate", constants.InventoryFurnaceResult, 18, true),
		tas.MineResource("stone", 5),

		tas.Transfer("stone-furnace", "copper-ore", constants.InventoryFurnaceSource, 19, false),
		tas.WaitInventory("stone-furnace", "copper-plate", constants.InventoryFurnaceResult, 19),
		tas.Transfer("stone-furnace", "copper-plate", constants.InventoryFurnaceResult, 19, true),
	}

	cTasks := tas.Tasks{
		tas.Craft("steam-engine", 1),
		tas.Craft("offshore-pump", 1),
		tas.Craft("lab", 1),
		tas.Craft("small-electric-pole", 1),
		tas.Craft("boiler", 1),
	}

	bTasks := tas.Tasks{
		tas.Build("steam-engine", constants.DirectionEast),
		tas.Build("offshore-pump", constants.DirectionNorth),
		tas.Build("lab", constants.DirectionNorth),
		tas.Build("small-electric-pole", constants.DirectionNorth),
		tas.Build("boiler", constants.DirectionEast),
	}

	// Ensure we have materials before crafting
	cTasks[0].Prerequisites().Add(tasks[12])
	cTasks[1].Prerequisites().Add(tasks[16])

	// make sure the thing's crafted before building it

	bTasks[0].Prerequisites().Add(cTasks[0])
	bTasks[1].Prerequisites().Add(cTasks[1])
	bTasks[2].Prerequisites().Add(cTasks[2])
	bTasks[3].Prerequisites().Add(cTasks[3])
	bTasks[4].Prerequisites().Add(cTasks[4])

	// science
	sTasks := tas.Tasks{

		tas.MineResource("coal", 15),

		tas.Transfer("stone-furnace", "coal", constants.InventoryFuel, 5, false),

		tas.MineResource("iron-ore", 20),

		tas.Transfer("stone-furnace", "iron-ore", constants.InventoryFurnaceSource, 20, false),
		tas.MineResource("copper-ore", 10),
		tas.WaitInventory("stone-furnace", "iron-plate", constants.InventoryFurnaceResult, 20),
		tas.Transfer("stone-furnace", "iron-plate", constants.InventoryFurnaceResult, 20, true),

		tas.Craft("iron-gear-wheel", 10),

		tas.Transfer("stone-furnace", "copper-ore", constants.InventoryFurnaceSource, 10, false),
		tas.WaitInventory("stone-furnace", "copper-plate", constants.InventoryFurnaceResult, 10),
		tas.Transfer("stone-furnace", "copper-plate", constants.InventoryFurnaceResult, 10, true),

		tas.Craft("automation-science-pack", 10),

		tas.Transfer("boiler", "coal", constants.InventoryFuel, 10, false),
		tas.Transfer("lab", "automation-science-pack", constants.InventoryLabInput, 10, false),
	}

	// get iron before crafting gears
	sTasks[7].Prerequisites().Add(sTasks[6])

	// get copper before crafting packs
	sTasks[11].Prerequisites().Add(sTasks[10])

	// get packs before putting them in the lab
	sTasks[13].Prerequisites().Add(sTasks[11])

	if err = t.Add(tasks...); err != nil {
		panic(err)
	}
	if err = t.Add(cTasks...); err != nil {
		panic(err)
	}
	if err = t.Add(bTasks...); err != nil {
		panic(err)
	}
	if err = t.Add(sTasks...); err != nil {
		panic(err)
	}

	of := os.Stdout

	if err = t.Export(of); err != nil {
		panic(err)
	}

}
