package building

import (
	"fmt"

	"github.com/brettschalin/factorio-min-resources/data"
	"github.com/brettschalin/factorio-min-resources/shims/slices"
)

type Inventory interface {
	Put(item string, amount int) error
	Take(item string, amount int) error
}

type inventory struct {
	maxSlots    int
	limitations []string // what items can be added to the slots
	data        map[string]int
}

func (i *inventory) Put(item string, amount int) error {

	if len(i.limitations) > 0 && !slices.Contains(i.limitations, item) {
		return fmt.Errorf("inventory: could not add %s (not allowed in this inventory)", item)
	}

	s := data.GetItem(item).StackSize

	newN := i.data[item] + amount
	if newN > s {
		return fmt.Errorf("inventory: could not add %s (wanted %d but only had space for %d)", item, newN, s)
	}

	if len(i.data) == i.maxSlots {
		return fmt.Errorf("inventory: could not add %s (no available slots)", item)
	}

	i.data[item] = newN

	return nil
}

func (i *inventory) Take(item string, amount int) error {

	if len(i.limitations) > 0 && !slices.Contains(i.limitations, item) {
		return fmt.Errorf("inventory: could not take %s (not allowed in this inventory)", item)
	}

	newN := i.data[item] - amount
	if newN < 0 {
		return fmt.Errorf("inventory: could not take %s (wanted %d but only had %d)", item, amount, i.data[item])
	}

	i.data[item] = newN
	if i.data[item] <= 0 {
		delete(i.data, item)
	}
	return nil
}

func newInventory(maxSlots int, limitations []string) *inventory {
	return &inventory{
		maxSlots:    maxSlots,
		limitations: limitations,
		data:        make(map[string]int, maxSlots),
	}
}

type ErrTooManyModules struct {
	machine  string
	max, got int
}

func (e ErrTooManyModules) Error() string {
	return fmt.Sprintf(`tried to add %d modules but machine %q can only accept %d`, e.got, e.machine, e.max)
}

type ErrNotEnoughModules struct {
	machine     string
	wanted, got int
}

func (e ErrNotEnoughModules) Error() string {
	return fmt.Sprintf(`tried to take %d modules but machine %q only had %d`, e.wanted, e.machine, e.got)
}

type ErrForbiddenModules struct {
	machine string
	module  string
}

func (e ErrForbiddenModules) Error() string {
	return fmt.Sprintf(`module %q not allowed in machine %q`, e.module, e.machine)
}

type Modules struct {
	maxSlots    int
	machine     Building
	modules     []*data.Module
	limitations []string // what modules this accepts
}

func (m Modules) ProductivityBonus(recipe string) float64 {
	var bonus float64

	for _, mod := range m.modules {
		if recipe == "" || mod.AppliesTo(recipe) {
			bonus += mod.ProductivityBonus()
		}
	}

	return bonus
}

func (m *Modules) Put(module string, amount int) error {

	if len(m.limitations) > 0 && !slices.Contains(m.limitations, module) {
		return ErrForbiddenModules{machine: m.machine.Name(), module: module}
	}

	if n := len(m.modules) + amount; n > m.maxSlots {
		return ErrTooManyModules{machine: m.machine.Name(), max: len(m.modules), got: n}
	}

	for i := 0; i < amount; i++ {
		m.modules = append(m.modules, data.GetModule(module))
	}

	return nil

}

func (m *Modules) Take(module string, amount int) error {

	if len(m.limitations) > 0 && !slices.Contains(m.limitations, module) {
		return ErrForbiddenModules{machine: m.machine.Name(), module: module}
	}

	n := amount
	taken := 0
	newMods := []*data.Module{}
	for _, mod := range m.modules {
		if mod.Name == module {
			n--
			if n < 0 {
				return ErrNotEnoughModules{machine: m.machine.Name(), wanted: amount, got: taken}
			}
			taken++
		} else {
			newMods = append(newMods, mod)
		}
	}

	m.modules = newMods

	return nil
}
