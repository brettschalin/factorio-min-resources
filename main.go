package main

import (
	"io"
	"os"

	"github.com/brettschalin/factorio-min-resources/building"
	"github.com/brettschalin/factorio-min-resources/data"
	"github.com/brettschalin/factorio-min-resources/state"
	"github.com/brettschalin/factorio-min-resources/tas"
)

func main() {

	must(data.Init(
		"./data/data-raw-dump.json",
	))

	out, close, err := getOutputFile()
	must(err)
	defer close()

	t := tas.TAS{}

	// this will take a while, might as well speed it up for us
	t.Add(tas.Speed(100))

	t.Add(makeTechTasks()...)

	var state = state.New()
	state.Assembler = building.NewAssembler(data.GetAssemblingMachine("assembling-machine-2"))
	state.Furnace = building.NewFurnace(data.GetFurnace("stone-furnace"))
	state.Lab = building.NewLab(data.GetLab("lab"))
	state.Boiler = building.NewBoiler(data.GetBoiler("boiler"))
	state.Refinery = building.NewAssembler(data.GetAssemblingMachine("oil-refinery"))
	state.Chem = building.NewAssembler(data.GetAssemblingMachine("chemical-plant"))

	tasks, f := makePowerSetup(state)
	must(t.Add(tasks...))

	tasks, f = researchRGTech("steel-processing", state, f)
	must(t.Add(tasks...))

	tasks, f = researchRGTech("logistic-science-pack", state, f)
	must(t.Add(tasks...))

	tasks, f = researchRGTech("automation", state, f)
	must(t.Add(tasks...))

	tasks, f = researchRGTech("electronics", state, f)
	must(t.Add(tasks...))

	tasks, f = researchRGTech("optics", state, f)
	must(t.Add(tasks...))

	tasks, f = buildSolarPanel(state, f)
	must(t.Add(tasks...))

	tasks, f = buildSteelFurnace(state, f)
	must(t.Add(tasks...))

	tasks, f = researchRGTech("automation-2", state, f)
	must(t.Add(tasks...))

	tasks, f = researchRGTech("engine", state, f)
	must(t.Add(tasks...))

	tasks, f = researchRGTech("fluid-handling", state, f)
	must(t.Add(tasks...))

	tasks, f = researchRGTech("oil-processing", state, f)
	must(t.Add(tasks...))

	tasks, f = researchModules(state, f)
	must(t.Add(tasks...))

	tasks, f = buildOilSetup(state, f)
	must(t.Add(tasks...))

	tasks, f = prodmod1(state, f)
	must(t.Add(tasks...))

	must(t.Add(buildElectricFurnace(state, f)...))

	t.Add(tas.Speed(1))

	must(t.Export(out))

}

func must(e error) {
	if e != nil {
		panic(e)
	}
}

func getOutputFile() (file io.Writer, close func() error, err error) {
	if len(os.Args) > 1 {
		f, err := os.Create(os.Args[1])
		if err != nil {
			return nil, nil, err
		}
		return f, f.Close, nil
	}

	return os.Stdout, func() error { return nil }, nil
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
	"plastics",
	"advanced-electronics",
	"modules",
	"productivity-module", // first modules!
	// "sulfur-processing",
	// "chemical-science-pack",          // blue science packs
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
