package tas

import (
	"fmt"

	"github.com/brettschalin/factorio-min-resources/calc"
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
			if !visited[p.ID()] {
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

			newInv, err := calc.Handcraft(s.Inventory, t.Recipe, t.Amount)
			if err != nil {
				return fmt.Errorf(`[craft] cannot handcraft %q: %v`, t.Recipe, err)
			}

			s.Inventory = newInv
		case *taskBuild:
			if s.Inventory[t.Entity] == 0 {
				return fmt.Errorf(`[build] no %q in inventory`, t.Entity)
			}

			s.Inventory[t.Entity]--
			s.Buildings[t.Entity] = true
		case *taskMine:

			if t.Resource != "" {
				s.Inventory[t.Resource] += t.Amount
			} else if t.Entity != "" {

				if !s.Buildings[t.Entity] {
					return fmt.Errorf(`[mine] building %q not placed`, t.Entity)
				}
				delete(s.Buildings, t.Entity)
				s.Inventory[t.Entity]++

			}
		case *taskTake:
			s.Inventory[t.Item] += t.Amount

		case *taskPut:
			if s.Inventory[t.Item] < t.Amount {
				return fmt.Errorf(`[put] need %d %q but only have %d`, t.Amount, t.Item, s.Inventory[t.Item])
			}

			s.Inventory[t.Item] -= t.Amount

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
		}
	}

	return nil
}
