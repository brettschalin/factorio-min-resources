package main

import (
	"os"

	"github.com/brettschalin/factorio-min-resources/data"
	"github.com/brettschalin/factorio-min-resources/task"
)

func main() {
	var err error
	if err = data.Init(
		"./data/data-raw-dump.json",
	); err != nil {
		panic(err)
	}

	tasks := []*task.Task{
		task.NewBuild("stone-furnace", ""),
		task.NewCraft(map[string]int{
			"steam-engine":        1,
			"offshore-pump":       1,
			"lab":                 1,
			"small-electric-pole": 2,
			"boiler":              1,
		}),

		task.NewBuild("offshore-pump", task.DirectionNorth),
		task.NewBuild("steam-engine", task.DirectionEast),
		task.NewBuild("lab", task.DirectionNorth),
		task.NewBuild("small-electric-pole", ""),
		task.NewBuild("boiler", task.DirectionEast),
		task.NewTech("logistic-science-pack"),
		task.NewTech("solar-energy"),
	}

	// of, err := os.OpenFile("./mods/MinPctTAS_0.0.1/tasks.lua", os.O_CREATE|os.O_APPEND|os.O_WRONLY|os.O_TRUNC, 0666)
	// if err != nil {
	// 	panic(err)
	// }

	// defer of.Close()

	of := os.Stdout

	t := &task.Task{
		Type:          task.TaskMeta,
		Prerequisites: tasks,
	}
	task.Optimize(t)

	if err = task.WriteTasksFile(of, t); err != nil {
		panic(err)
	}

}
