package tas

import (
	"fmt"

	"github.com/brettschalin/factorio-min-resources/building"
	"github.com/brettschalin/factorio-min-resources/calc"
	"github.com/brettschalin/factorio-min-resources/constants"
	"github.com/brettschalin/factorio-min-resources/data"
	"github.com/brettschalin/factorio-min-resources/state"
)

type TAS struct {
	tasks Tasks
}

func (t *TAS) Add(tasks ...Task) error {
	t.tasks = append(t.tasks, tasks...)
	return t.Verify()
}

func (t *TAS) Verify() error {

	var err error

	if err = t.verifyPrereqs(); err != nil {
		return err
	}

	s := state.New()
	if err = t.verifyState(s); err != nil {
		return err
	}

	return nil
}

func (tas *TAS) verifyPrereqs() error {

	visited := map[string]bool{}
	for i, task := range tas.tasks {
		for _, p := range *task.Prerequisites() {
			if !visited[p.ID()] && p.Type() != taskPrereq {
				return fmt.Errorf(`task %d references unknown prerequisite %s`, i, p.ID())
			}
		}
		visited[task.ID()] = true
	}

	return nil
}

func (tas *TAS) verifyState(s *state.State) error {
	for _, task := range tas.tasks {
		switch t := task.(type) {
		case *taskCraft:

			newInv, err := calc.Handcraft(s.Inventory, data.GetRecipe(t.Recipe), t.Amount)
			if err != nil {
				return fmt.Errorf(`[craft] cannot handcraft %q: %v`, t.Recipe, err)
			}

			s.Inventory = newInv
		case *taskTech:
			if s.TechResearched[t.Tech] {
				return fmt.Errorf(`[tech] %q already researched`, t.Tech)
			}
			tech := data.GetTech(t.Tech)
			for _, p := range tech.Prerequisites {
				if !s.TechResearched[p] {
					return fmt.Errorf(`[tech] %q: prerequisite %q not yet researched`, t.Tech, p)
				}
			}
			s.TechResearched[t.Tech] = true

		case *taskRecipe:
			if !s.Buildings[t.Entity] {
				return fmt.Errorf(`[recipe] building %q not placed`, t.Entity)
			}

			if b, ok := s.GetBuilding(t.Entity).(*building.Assembler); ok {
				inv := b.SetRecipe(data.GetRecipe(t.Recipe))
				for ing, n := range inv {
					s.Inventory[ing] += uint(n)
				}
			} else {
				return fmt.Errorf(`[recipe] cannot set recipes on %q`, t.Entity)
			}
		case *taskBuild:
			if s.Inventory[t.Entity] == 0 {
				return fmt.Errorf(`[build] no %q in inventory`, t.Entity)
			}

			s.Inventory[t.Entity]--
			s.Buildings[t.Entity] = true
			if ok := s.ConstructBuilding(t.Entity); !ok {
				return fmt.Errorf(`[build] could not place %q`, t.Entity)
			}

		case *taskMine:
			if t.Resource != "" {
				s.Inventory[t.Resource] += t.Amount
			} else if t.Entity != "" {
				if !s.Buildings[t.Entity] {
					return fmt.Errorf(`[mine] building %q not placed`, t.Entity)
				}
				delete(s.Buildings, t.Entity)
				s.Inventory[t.Entity]++

				// TODO: inventory transfers like in the *taskRecipe case
				if ok := s.MineBuilding(t.Entity); !ok {
					return fmt.Errorf(`[mine] building %q not placed`, t.Entity)
				}

			}

		case *taskTake:

			b := s.GetBuilding(t.Entity)
			if b == nil {
				return fmt.Errorf(`[take] building %q not placed`, t.Entity)
			}

			inv := b.Inventory(t.Slot)
			if inv == nil {
				return fmt.Errorf(`[take] building %q does not have slot %q`, t.Entity, t.Slot)
			}

			err := inv.Take(t.Item, int(t.Amount))
			if err != nil {
				return fmt.Errorf(`[take] not enough %s in output slot of %q (wanted %d)`, t.Item, t.Entity, t.Amount)
			}

			if m, ok := b.(building.CraftingBuilding); ok {
				for {
					status := m.DoCraft()
					if status != building.CraftStatusRunning {
						break
					}
				}
			}

			s.Inventory[t.Item] += t.Amount

		case *taskPut:

			b := s.GetBuilding(t.Entity)
			if b == nil {
				return fmt.Errorf(`[put] building %q not placed`, t.Entity)
			}

			if s.Inventory[t.Item] < t.Amount {
				return fmt.Errorf(`[put] need %d %q but only have %d`, t.Amount, t.Item, s.Inventory[t.Item])
			}

			inv := b.Inventory(t.Slot)
			if inv == nil {
				return fmt.Errorf(`[put] building %q does not have slot %q`, t.Entity, t.Slot)
			}

			err := inv.Put(t.Item, int(t.Amount))
			if err != nil {
				return fmt.Errorf(`[put] cannot put %s in input slot of %q (wanted %d)`, t.Item, t.Entity, t.Amount)
			}

			if m, ok := b.(building.CraftingBuilding); ok {
				for {
					status := m.DoCraft()
					if status != building.CraftStatusRunning {
						break
					}
				}
			}
			s.Inventory[t.Item] -= t.Amount

			// TODO: in the future we should properly track fuel usage. For now just assume that it's correct and empty the inventory
			if t.Slot == constants.InventoryFuel {
				_ = inv.Take(t.Item, inv.Count(t.Item))
			}

			// TODO: in the future we should also track science pack usage. Assume that's correct as well
			if t.Slot == constants.InventoryLabInput {
				_ = inv.Take(t.Item, inv.Count(t.Item))
			}

		}
		for k, v := range s.Inventory {
			if v == 0 {
				delete(s.Inventory, k)
			}
		}
	}

	return nil
}
