package task

import (
	"math"

	"github.com/brettschalin/factorio-min-resources/building"
	"github.com/brettschalin/factorio-min-resources/calc"
	"github.com/brettschalin/factorio-min-resources/data"
	"github.com/brettschalin/factorio-min-resources/state"
)

// various configurations. This should be ideal for vanilla but YMMV
const (

	// what the boiler/furnace should be fueled with. This is assumed to be a minable resource
	preferredFuel = "coal"
)

func Optimize(task *Task) *Task {

	s := state.New()
	pass1(s, task)

	task.Prune()

	s = state.New()
	pass2(s, task)

	// tasks = pass3(tasks)

	return task
}

// pass1 performs basic optimizations, including
// * technologies are not researched twice (and the extra science packs aren't crafted)
// * Inventory is taken into account and alters the number of things to craft/mine
// * (coming soon) applying module bonuses
func pass1(s *state.State, task *Task) {

	for _, p := range task.Prerequisites {
		pass1(s, p)
	}

	switch task.Type {
	case TaskTech:
		// do not research completed techs
		if s.TechResearched[task.Tech] {
			task.Tech = ""
			return
		}
		s.TechResearched[task.Tech] = true
		task.Prune()

	case TaskMine:
		n := s.Inventory[task.Item]
		if n > 0 {
			remaining := max(task.Amount-n, 0)
			task.Amount = remaining
		}
		s.Inventory[task.Item] += task.Amount

	case TaskCraft:
		cost, prod := calc.RecipeCost(task.Item, task.Amount)
		n := s.Inventory[task.Item]
		if n > 0 {
			// we have some in the inventory. Adjust the subtasks as needed
			remaining := max(task.Amount-n, 0)
			toSub := make(map[string]int)
			cost, _ := calc.RecipeCost(task.Item, n)
			for i, n := range cost {
				toSub[i] = n
			}
			pass1Reverse(task, toSub)
			task.Amount = remaining
		}
		for c, n := range cost {
			s.Inventory[c] -= n
		}
		for p, n := range prod {
			s.Inventory[p] += n
		}
	}

}

func pass1Reverse(task *Task, toSub map[string]int) {
	for ; task != nil; task = task.Prev() {
		if toSub[task.Item] == 0 {
			continue
		}
		switch task.Type {
		case TaskCraft:
			oldAmt := task.Amount
			task.Amount = max(task.Amount-toSub[task.Item], 0)
			cost, _ := calc.RecipeCost(task.Item, max(oldAmt-task.Amount, 0))
			toSub[task.Item] -= task.Amount
			for i, n := range cost {
				toSub[i] += n
			}
		case TaskMine:
			task.Amount = max(task.Amount-toSub[task.Item], 0)
			toSub[task.Item] = 0
		}
	}
}

func max[T ~int | ~float64](n1, n2 T) T {
	if n1 < n2 {
		return n2
	}
	return n1
}

func min[T ~int | ~float64](n1, n2 T) T {
	if n1 > n2 {
		return n2
	}
	return n1
}

// pass2 performs the next phase of optimizations, including
// * determining if a recipe should be handcrafted
// * transferring batches of material into machines for crafting and taking products out
func pass2(s *state.State, task *Task) {
	for _, p := range task.Prerequisites {
		pass2(s, p)
	}

	switch task.Type {
	case TaskCraft:
		if task.Amount <= 0 {
			// pruning should make sure we never get here, but just in case we do it anyways
			return
		}
		rec := data.D.GetRecipe(task.Item)

		// TODO: we also want to use a machine when the recipe accepts prod mods
		// and we have some available
		shouldHandcraft := rec.CanHandcraft()

		if shouldHandcraft {
			// handcrafting. We want the recipe count instead of total products
			div := float64(rec.ProductCount(task.Item))
			if div != 1 {
				task.Amount = int(math.Ceil(float64(task.Amount) / div))
			}
		} else {

			var building string

			inSlot := "defines.inventory.assembling_machine_input"
			outSlot := "defines.inventory.assembling_machine_output"
			isSmelting := false
			switch rec.Category {
			case "smelting":
				building = s.Furnace.Entity.Name
				inSlot = "defines.inventory.furnace_source"
				outSlot = "defines.inventory.furnace_result"
				isSmelting = true
			case "oil-processing":
				building = s.Refinery.Entity.Name
			case "chemistry":
				building = s.Chem.Entity.Name
			default:
				building = s.Assembler.Entity.Name
			}

			task.Type = TaskMeta

			cost, products := calc.RecipeCost(task.Item, task.Amount)
			osc := rec.OneStackCount()

			recipesToMake := int(math.Floor(float64(products[task.Item]) / float64(rec.ProductCount(task.Item))))
			osc = min(osc, recipesToMake)

			f := s.Furnace

			done := len(cost)
			for done > 0 {
				done = len(cost)
				if isSmelting {

					fuelCost := f.FuelCost(preferredFuel, task.Item, osc)

					if fuelCost > 0 {
						task.AddPrereq(NewMine(preferredFuel, fuelCost))
						task.AddPrereq(NewTransfer(f.Entity.Name, "defines.inventory.fuel", preferredFuel, fuelCost, false))
					}

				} else {
					// TODO: set a recipe on the machine. There's probably a better check for when to do this
					// than "it's not a furnace," but I'm doing it for now because it's easy
				}
				for _, ing := range rec.Ingredients {
					task.AddPrereq(NewTransfer(building, inSlot, ing.Name, ing.Amount*osc, false))
					cost[ing.Name] -= ing.Amount * osc
					if cost[ing.Name] <= 0 {
						done--
					}
				}

				task.AddPrereq(NewWait(building, outSlot, task.Item, rec.ProductCount(task.Item)*osc))
				task.AddPrereq(NewTransfer(building, outSlot, task.Item, rec.ProductCount(task.Item)*osc, true))

				recipesToMake -= osc
				osc = min(osc, recipesToMake)
			}
		}

	case TaskMine:
		if task.Amount <= 0 {
			return
		}
		if _, ok := data.D.Furnace[task.Entity]; ok {
			s.Furnace = nil
		}
		if _, ok := data.D.Boiler[task.Entity]; ok {
			s.Boiler = nil
		}
		if _, ok := data.D.Lab[task.Entity]; ok {
			s.Lab = nil
		}
		if a, ok := data.D.AssemblingMachine[task.Entity]; ok {
			// quick hack, but works for vanilla
			if a.Name == "chemical-plant" {
				s.Chem = nil
			} else if a.Name == "oil-refinery" {
				s.Refinery = nil
			} else {
				s.Assembler = nil
			}
		}

		if r := task.Item; r != "" {
			s.Inventory[r]++
		}

		if e := task.Entity; e != "" {
			s.Inventory[e]++
			delete(s.Buildings, task.Entity)
		}

	case TaskBuild:
		if f, ok := data.D.Furnace[task.Entity]; ok {
			s.Furnace = building.NewFurnace(&f)
		}
		if b, ok := data.D.Boiler[task.Entity]; ok {
			s.Boiler = building.NewBoiler(&b)
		}
		if l, ok := data.D.Lab[task.Entity]; ok {
			s.Lab = building.NewLab(&l)
		}
		if a, ok := data.D.AssemblingMachine[task.Entity]; ok {
			// quick hack, but works for vanilla
			if a.Name == "chemical-plant" {
				s.Chem = building.NewAssembler(&a)
			} else if a.Name == "oil-refinery" {
				s.Refinery = building.NewAssembler(&a)
			} else {
				s.Assembler = building.NewAssembler(&a)
			}
		}

		s.Buildings[task.Entity] = true
		s.Inventory[task.Entity]--
	case TaskTech:

		var nFuel int
		// fuelStackSize := data.D.Item[preferredFuel].StackSize

		if !s.Buildings["solar-panel"] {
			nFuel = s.Boiler.FuelCost(preferredFuel, s.Lab.EnergyCost(task.Tech))
			task.AddPrereq(NewMine(preferredFuel, nFuel))
		}

		// TODO: various math to determine when to refuel the boiler vs when to add more
		// science packs to the lab
	}
}

// pass3 performs some purely optional optimizations to speed up the run, including
// * reordering waits so that mining and such can be done in the meantime
func pass3(task *Task) {

	panic("not implemented")

	return
}
