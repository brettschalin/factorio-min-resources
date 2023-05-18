package building

import (
	"fmt"
	"math"

	"github.com/brettschalin/factorio-min-resources/data"
)

type modules struct {
	m      []string
	maxLen int
}

func (m *modules) add(mod string) {
	if len(m.m) == m.maxLen {
		// TODO: handle this better
		panic(fmt.Sprintf(`module: trying to add %q, machine already has %s`, mod, m.m))
	}
	m.m = append(m.m, mod)
}

func (m *modules) remove(mod string) {
	newM := make([]string, 0, m.maxLen)
	idx := 0
	for ; idx < len(m.m); idx++ {
		if m.m[idx] == mod {
			idx++
			break
		}
		newM = append(newM, m.m[idx])
	}
	if idx < len(m.m) {
		newM = append(newM, m.m[idx:]...)
	}
	m.m = newM
}

type Assembler struct {
	Entity  *data.AssemblingMachine
	Modules modules
}

func NewAssembler(spec *data.AssemblingMachine) *Assembler {
	var mods modules
	if n := spec.ModuleSpecification.ModuleSlots; n > 0 {
		mods = modules{
			m:      make([]string, 0, n),
			maxLen: n,
		}
	}

	return &Assembler{
		Entity:  spec,
		Modules: mods,
	}
}

type Furnace struct {
	Entity  *data.Furnace
	Modules modules
}

func NewFurnace(spec *data.Furnace) *Furnace {
	var mods modules
	if n := spec.ModuleSpecification.ModuleSlots; n > 0 {
		mods = modules{
			m:      make([]string, 0, n),
			maxLen: n,
		}
	}
	return &Furnace{
		Entity:  spec,
		Modules: mods,
	}
}

func (f *Furnace) FuelCost(fuel, item string, nRecipes int) int {
	// electric furnaces don't have a fuel slot
	if f.Entity.EnergySource.FuelCategory != "chemical" {
		return 0
	}

	rec := data.D.GetRecipe(item)

	e := float64(data.D.Item[fuel].FuelValue)
	c := float64(f.Entity.EnergyUsage)

	timeToCraft := rec.CraftingTime() / f.Entity.CraftingSpeed

	energy := timeToCraft * float64(c)

	n := float64(nRecipes) * (energy / e)

	// TODO: we're rounding up for the required fuel. That fraction should be accounted for
	// but isn't yet

	return int(math.Ceil(n))
}

type Boiler struct {
	Entity *data.Boiler

	// I only care about the effective power conversion,
	// so this will act like a combined boiler/steam engine
	// where water and fuel directly produce electricity.
}

func NewBoiler(spec *data.Boiler) *Boiler {
	return &Boiler{
		Entity: spec,
	}
}

func (b *Boiler) FuelCost(fuel string, energy float64) int {

	item := data.D.Item[fuel]

	// TODO: factor in b.Entity.Effectivity. Vanilla boiler says "1"

	return int(math.Ceil(float64(energy) / float64(item.FuelValue)))

}

type Lab struct {
	Entity  *data.Lab
	Modules modules
}

func NewLab(spec *data.Lab) *Lab {
	var mods modules
	if n := spec.ModuleSpecification.ModuleSlots; n > 0 {
		mods = modules{
			m:      make([]string, 0, n),
			maxLen: n,
		}
	}

	return &Lab{
		Entity:  spec,
		Modules: mods,
	}
}

func (l *Lab) EnergyCost(tech string) float64 {
	t := data.D.GetTech(tech)
	e := l.Entity.EnergyUsage
	time := t.Unit.Time
	n := t.Unit.Count

	return float64(time) * float64(n) * float64(e)
}
