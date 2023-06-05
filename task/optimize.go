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

	pass1(task, s.Copy())
	task.Prune()
	pass2(task, s.Copy())
	task.Prune()

	return task
}

// pass1 iterates through the task tree and updates crafting amounts and
// ensuring we don't attempt to research the same tech twice
func pass1(task *Task, s *state.State) {
	sBackup := s.Copy()

	for _, p := range task.Prerequisites {
		pass1(p, s)
	}

	switch task.Type {
	case TaskCraft:

		// reduce crafting amount if we have some of the item in our inventory
		if n := s.Inventory[task.Item]; n > 0 {
			remaining := max(task.Amount-n, 0)
			toSub, _ := calc.RecipeCost(task.Item, n)
			pass1Reverse(task.Prev(), nil, toSub)
			task.Amount = remaining
		}

		if task.Amount == 0 {
			// get state back to what we expect
			task.eval(sBackup, true)
			*s = *sBackup
			return
		}

		rec := data.GetRecipe(task.Item)

		// TODO: update this check. Machine crafting should also happen when
		// we can use prod modules
		shouldHandcraft := rec.CanHandcraft()

		if shouldHandcraft {
			// handcrafting. We want the recipe count instead of total products
			div := float64(rec.ProductCount(task.Item))
			if div != 1 {
				task.Amount = int(math.Ceil(float64(task.Amount) / div))
			}
			task.Type = TaskHandcraft
		}

		// get state back to what we expect
		task.eval(sBackup, true)
		*s = *sBackup

	case TaskMine:
		if n := s.Inventory[task.Item]; n > 0 {
			remaining := max(task.Amount-n, 0)
			task.Amount = remaining
		}
		task.eval(s, false)

	case TaskBuild:
		task.eval(s, false)

	case TaskTech:
		if s.TechResearched[task.Tech] {
			task.Tech = "" // mark for pruning
		}
		task.eval(s, false)

	default:
		task.eval(s, false)
	}

}

func pass1Reverse(task, until *Task, toSub map[string]int) {

	for ; task != until; task = task.Prev() {
		if toSub[task.Item] == 0 {
			continue
		}
		switch task.Type {
		case TaskHandcraft:

			// task.Amount is how many recipes to craft, rather than the total number of items we want at the end

			rec := data.GetRecipe(task.Item)
			totalAmount := task.Amount * rec.ProductCount(task.Item)

			oldAmt := totalAmount

			totalAmount = max(totalAmount-toSub[task.Item], 0)
			cost, _ := calc.RecipeCost(task.Item, max(oldAmt-totalAmount, 0))
			toSub[task.Item] -= totalAmount
			for i, n := range cost {
				toSub[i] += n
			}

			task.Amount = int(math.Ceil(float64(totalAmount) / float64(rec.ProductCount(task.Item))))

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

// pass2 takes crafting/research tasks and replaces them with ones to batch
// inputting ingredients/fuel to and removing results from the machines
func pass2(task *Task, s *state.State) {
	sBackup := s.Copy()

	for _, p := range task.Prerequisites {
		pass2(p, s)
	}

	switch task.Type {
	case TaskHandcraft:
		task.eval(s, false)

	case TaskCraft:

		rec := data.GetRecipe(task.Item)

		// get the building/slots to use
		var (
			buildingName string
			inSlot       string
			outSlot      string
			fuelSlot     string
			needsFuel    bool
		)

		switch rec.Category {
		case "smelting":
			buildingName = s.Furnace.Entity.Name
			inSlot = s.Furnace.Slots.Input
			outSlot = s.Furnace.Slots.Output
			fuelSlot = s.Furnace.Slots.Fuel
			// TODO: move this check somewhere better
			needsFuel = s.Furnace.Entity.EnergySource.FuelCategory == "chemical"
		case "oil-processing":
			buildingName = s.Refinery.Entity.Name
			inSlot = s.Refinery.Slots.Input
			outSlot = s.Refinery.Slots.Output
		case "chemistry":
			buildingName = s.Chem.Entity.Name
			inSlot = s.Chem.Slots.Input
			outSlot = s.Chem.Slots.Output
		default:
			buildingName = s.Assembler.Entity.Name
			inSlot = s.Assembler.Slots.Input
			outSlot = s.Assembler.Slots.Output
		}

		// now for the batching
		cost, prod := calc.RecipeCost(task.Item, task.Amount)
		osc := rec.OneStackCount()
		recipesToMake := int(math.Floor(float64(prod[task.Item]) / float64(rec.ProductCount(task.Item))))
		osc = min(osc, recipesToMake)

		done := len(cost)
		for ; done > 0; done = len(cost) {
			if needsFuel {
				// mine and transfer fuel for this batch

				// TODO: we shouldn't hardcode furnaces being the only machines
				// that take fuel. It's mostly true for vanilla
				fuelCost := s.Furnace.FuelCost(preferredFuel, task.Item, osc)
				if fuelCost > 0 {
					task.AddPrereq(NewMine(preferredFuel, fuelCost))
					task.AddPrereq(NewTransfer(buildingName, fuelSlot, preferredFuel, fuelCost, false))
				}

			} else {
				// set recipe on the building
				// TODO: the check needs to be something other than "it's not a furnace that burns fuel"
				// but this works for now
			}

			for _, ing := range rec.Ingredients {
				task.AddPrereq(NewTransfer(buildingName, inSlot, ing.Name, ing.Amount*osc, false))
				cost[ing.Name] -= ing.Amount * osc
				if cost[ing.Name] <= 0 {
					done--
					delete(cost, ing.Name)
				}
			}
			task.AddPrereq(NewWait(buildingName, outSlot, task.Item, rec.ProductCount(task.Item)*osc))
			task.AddPrereq(NewTransfer(buildingName, outSlot, task.Item, rec.ProductCount(task.Item)*osc, true))

			recipesToMake -= osc
			osc = min(osc, recipesToMake)
		}

		task.Type = TaskMeta
		task.Item = ""
		task.Amount = 0

		// get state back to what we expect
		task.eval(sBackup, true)
		*s = *sBackup

	case TaskMine:
		task.eval(s, false)

	case TaskBuild:
		task.eval(s, false)

	case TaskTech:

		// TODO: we need to to similar batching as in TaskCraft, both to put packs in the lab
		// but also to fuel the boiler before we get a solar panel
		task.eval(s, false)

	default:
		task.eval(s, false)
	}

}

// eval performs the effect of the task on the provided state object
func (t *Task) eval(s *state.State, doPrereqs bool) {
	if doPrereqs {
		for _, p := range t.Prerequisites {
			p.eval(s, true)
		}
	}

	switch t.Type {
	case TaskHandcraft:
		if t.Amount == 0 {
			return
		}
		rec := data.GetRecipe(t.Item)
		for _, ing := range rec.Ingredients {
			s.Inventory[ing.Name] -= ing.Amount * t.Amount
		}

		for _, prod := range rec.GetResults() {
			s.Inventory[prod.Name] += prod.Amount * t.Amount
		}

	case TaskCraft:
		if t.Amount == 0 {
			return
		}
		rec := data.GetRecipe(t.Item)
		div := float64(rec.ProductCount(t.Item))
		count := 1
		if div != 1 {
			count = int(math.Ceil(float64(t.Amount) / div))
		}

		for _, ing := range rec.Ingredients {
			s.Inventory[ing.Name] -= ing.Amount * count
		}

		for _, prod := range rec.GetResults() {
			s.Inventory[prod.Name] += prod.Amount * count
		}

	case TaskMine:
		if f := data.GetFurnace(t.Entity); f.Name != "" {
			s.Furnace = nil
		}
		if b := data.GetBoiler(t.Entity); b.Name != "" {
			s.Boiler = nil
		}
		if l := data.GetLab(t.Entity); l.Name != "" {
			s.Lab = nil
		}
		if a := data.GetAssemblingMachine(t.Entity); a.Name != "" {
			switch a.Name {
			case s.Chem.Entity.Name:
				s.Chem = nil
			case s.Refinery.Entity.Name:
				s.Refinery = nil
			case s.Assembler.Entity.Name:
				s.Assembler = nil
			}
		}

		if r := t.Item; r != "" {
			s.Inventory[r] += t.Amount
		}

		if e := t.Entity; e != "" {
			s.Inventory[e]++
			delete(s.Buildings, t.Entity)
		}

	case TaskBuild:

		s.Inventory[t.Entity]--
		s.Buildings[t.Entity] = true

		if f := data.GetFurnace(t.Entity); f.Name != "" {
			s.Furnace = building.NewFurnace(f)
		}
		if b := data.GetBoiler(t.Entity); b.Name != "" {
			s.Boiler = building.NewBoiler(b)
		}
		if l := data.GetLab(t.Entity); l.Name != "" {
			s.Lab = building.NewLab(l)
		}
		if a := data.GetAssemblingMachine(t.Entity); a.Name != "" {
			b := building.NewAssembler(a)

			// extremely hacky but it works for vanilla so it's staying for now
			switch a.Name {
			case "chemical-plant":
				s.Chem = b
			case "oil-refinery":
				s.Refinery = b
			default:
				s.Assembler = b
			}
		}
	case TaskTake:
		s.Inventory[t.Item] += t.Amount

	case TaskPut:
		s.Inventory[t.Item] -= t.Amount
		// TODO: if it's a fuel we need to keep track of how much whatever we're putting it into is using

	case TaskTech:
		s.TechResearched[t.Tech] = true

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

// pass3 performs some purely optional optimizations to speed up the run, including
// * reordering waits so that mining and such can be done in the meantime
func pass3(task *Task) {
	return
}
