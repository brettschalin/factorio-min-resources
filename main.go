package main

import (
	"os"

	"github.com/brettschalin/factorio-min-resources/building"
	"github.com/brettschalin/factorio-min-resources/data"
	"github.com/brettschalin/factorio-min-resources/tas"
)

func main() {

	must(data.Init(
		"./data/data-raw-dump.json",
	))

	t := tas.TAS{}

	// this will take a while, might as well speed it up for us
	t.Add(tas.Speed(200))

	t.Add(makeTechTasks()...)

	t.Add(makePowerSetup()...)

	f := building.NewFurnace(data.GetFurnace("stone-furnace"))
	l := building.NewLab(data.GetLab("lab"))
	b := building.NewBoiler(data.GetBoiler("boiler"))

	must(t.Add(researchTech("steel-processing", f, l, b)...))
	must(t.Add(researchTech("logistic-science-pack", f, l, b)...))
	must(t.Add(buildSolarPanel(f.Name())...))
	must(t.Add(buildSteelFurnace(true)...))

	b = nil
	f = building.NewFurnace(data.GetFurnace("steel-furnace"))

	must(t.Add(researchTech("automation-2", f, l, b)...))
	must(t.Add(researchTech("engine", f, l, b)...))
	must(t.Add(researchTech("fluid-handling", f, l, b)...))
	must(t.Add(researchTech("oil-processing", f, l, b)...))

	must(t.Add(buildOilSetup(f)...))

	t.Add(tas.Speed(1))

	of := os.Stdout

	must(t.Export(of))

}

func must(e error) {
	if e != nil {
		panic(e)
	}
}

// the technologies to research, in this specific order
var techs = []string{
	"steel-processing",
	"logistic-science-pack", // green science packs
	"automation",
	"electronics",
	"optics",
	"solar-energy",                 // solar panel
	"advanced-material-processing", // steel furnace
	"automation-2",
	"engine",
	"fluid-handling",
	"oil-processing", // refinery/chem plant
	// "plastics",
	// "advanced-electronics",
	// "modules",
	// "productivity-module", // first modules!
	// "sulfur-processing",
	// "chemical-science-pack", // blue science packs
	// "advanced-material-processing-2", // electric furnace
	// "advanced-electronics-2",
	// "productivity-module-2", // better modules
	// "logistics",
	// "logistics-2",
	// "railway",
	// "production-science-pack", // purple science packs
	// "productivity-module-3", // the best modules
	// "speed-module",
	// // "automation-3" // *
	// "advanced-oil-processing",
	// "flammables",
	// "rocket-fuel",
	// "concrete",
	// "speed-module-2",
	// "speed-module-3",
	// "lubricant",
	// "electric-engine",
	// "battery",
	// "robotics",
	// "low-density-structure",
	// "utility-science-pack", // yellow science packs
	// "rocket-control-unit",
	// "rocket-silo",

	// * this needs analysis. Are the two extra module slots an assembler 3 have
	// worth the cost of researching the tech, seeing as it isn't required to build the silo?

}

// exists for easier prerequisite definitions
var techMap = map[string]tas.Task{}

func makeTechTasks() tas.Tasks {

	tasks := make(tas.Tasks, len(techs))

	for i, t := range techs {
		tech := tas.Tech(t)
		tasks[i] = tech
		techMap[t] = tech
	}

	return tasks
}
