package building

import (
	"fmt"

	"github.com/brettschalin/factorio-min-resources/constants"
	"github.com/brettschalin/factorio-min-resources/data"
)

type modules struct {
	m      []string
	maxLen int
}

type slots struct {
	Input   constants.Inventory
	Output  constants.Inventory
	Fuel    constants.Inventory
	Modules constants.Inventory
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

type Building interface {
	Name() string
	Slots() slots
	Modules() modules
}

type Assembler struct {
	Entity  *data.AssemblingMachine
	slots   slots
	modules modules
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
		Entity: spec,
		slots: slots{
			Input:   constants.InventoryAssemblingMachineInput,
			Output:  constants.InventoryAssemblingMachineOutput,
			Modules: constants.InventoryAssemblingMachineModules,
		},
		modules: mods,
	}
}

func (a *Assembler) Name() string {
	return a.Entity.Name
}

func (a *Assembler) Slots() slots {
	return a.slots
}

func (a *Assembler) Modules() modules {
	return a.modules
}

type Furnace struct {
	Entity  *data.Furnace
	slots   slots
	modules modules
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
		Entity: spec,
		slots: slots{
			Input:   constants.InventoryFurnaceSource,
			Output:  constants.InventoryFurnaceResult,
			Fuel:    constants.InventoryFuel,
			Modules: constants.InventoryFurnaceModules,
		},
		modules: mods,
	}
}

func (f *Furnace) Name() string {
	return f.Entity.Name
}

func (f *Furnace) Slots() slots {
	return f.slots
}

func (f *Furnace) Modules() modules {
	return f.modules
}

type Boiler struct {
	Entity *data.Boiler
	slots  slots

	// I only care about the effective power conversion,
	// so this will act like a combined boiler/steam engine
	// where water and fuel directly produce electricity.
}

func NewBoiler(spec *data.Boiler) *Boiler {
	return &Boiler{
		Entity: spec,
		slots: slots{
			Fuel: constants.InventoryFuel,
		},
	}
}

func (b *Boiler) Name() string {
	return b.Entity.Name
}

func (b *Boiler) Slots() slots {
	return b.slots
}

func (b *Boiler) Modules() modules {
	return modules{}
}

type Lab struct {
	Entity  *data.Lab
	slots   slots
	modules modules
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
		Entity: spec,
		slots: slots{
			Input:   constants.InventoryLabInput,
			Modules: constants.InventoryLabModules,
		},
		modules: mods,
	}
}

func (l *Lab) Name() string {
	return l.Entity.Name
}

func (l *Lab) Slots() slots {
	return l.slots
}

func (l *Lab) Modules() modules {
	return l.modules
}
