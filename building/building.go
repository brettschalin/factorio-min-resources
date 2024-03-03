package building

import (
	"fmt"

	"github.com/brettschalin/factorio-min-resources/constants"
	"github.com/brettschalin/factorio-min-resources/data"
)

type slots struct {
	Input   constants.Inventory
	Output  constants.Inventory
	Fuel    constants.Inventory
	Modules constants.Inventory
}

type Modules []*data.Module

func (m Modules) ProductivityBonus(recipe string) float64 {
	var bonus float64

	for _, mod := range m {
		if mod.AppliesTo(recipe) {
			bonus += mod.ProductivityBonus()
		}
	}

	return bonus
}

type ErrTooManyModules struct {
	machine  string
	max, got int
}

func (e ErrTooManyModules) Error() string {
	return fmt.Sprintf(`tried to add %d modules but machine %q can only accept %d`, e.got, e.machine, e.max)
}

type Building interface {
	Name() string
	Slots() *slots
	GetModules() Modules
	SetModules(mod Modules) error
	ProductivityBonus(recipe string) float64
}

type CraftingBuilding interface {
	Building
	EnergySource() data.EnergySource
	EnergyUsage() float64
	CraftingSpeed() float64
}

type Assembler struct {
	Entity  *data.AssemblingMachine
	slots   slots
	modules Modules
}

func NewAssembler(spec *data.AssemblingMachine) *Assembler {
	var mods Modules
	if n := spec.ModuleSpecification.ModuleSlots; n > 0 {
		mods = make(Modules, n)
	}

	// vanilla contains no burner assemblers of any kind (only furnaces) but mods might so account for that here
	var fuelSlot constants.Inventory
	if spec.IsBurner() {
		fuelSlot = constants.InventoryFuel
	}

	return &Assembler{
		Entity: spec,
		slots: slots{
			Input:   constants.InventoryAssemblingMachineInput,
			Output:  constants.InventoryAssemblingMachineOutput,
			Fuel:    fuelSlot,
			Modules: constants.InventoryAssemblingMachineModules,
		},
		modules: mods,
	}
}

func (a *Assembler) Name() string {
	return a.Entity.Name
}

func (a *Assembler) Slots() *slots {
	return &a.slots
}

func (a *Assembler) GetModules() Modules {
	return a.modules
}

func (a *Assembler) SetModules(mod Modules) error {
	if modSlots := a.Entity.ModuleSpecification.ModuleSlots; len(mod) > modSlots {
		return ErrTooManyModules{machine: a.Name(), max: modSlots, got: len(mod)}
	}

	a.modules = mod
	return nil
}

func (a *Assembler) EnergySource() data.EnergySource {
	return a.Entity.EnergySource
}

func (a *Assembler) EnergyUsage() float64 {
	return float64(a.Entity.EnergyUsage)
}

func (a *Assembler) CraftingSpeed() float64 {
	return a.Entity.CraftingSpeed
}

func (a *Assembler) ProductivityBonus(recipe string) float64 {
	if a == nil {
		return 0
	}
	return a.modules.ProductivityBonus(recipe)
}

type Furnace struct {
	Entity  *data.Furnace
	slots   slots
	modules Modules
}

func NewFurnace(spec *data.Furnace) *Furnace {
	var mods Modules
	if n := spec.ModuleSpecification.ModuleSlots; n > 0 {
		mods = make(Modules, n)
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

func (f *Furnace) Slots() *slots {
	return &f.slots
}

func (f *Furnace) GetModules() Modules {
	return f.modules
}

func (f *Furnace) SetModules(mod Modules) error {
	if modSlots := f.Entity.ModuleSpecification.ModuleSlots; len(mod) > modSlots {
		return ErrTooManyModules{machine: f.Name(), max: modSlots, got: len(mod)}
	}

	f.modules = mod
	return nil
}

func (f *Furnace) EnergySource() data.EnergySource {
	return f.Entity.EnergySource
}

func (f *Furnace) EnergyUsage() float64 {
	return float64(f.Entity.EnergyUsage)
}

func (f *Furnace) CraftingSpeed() float64 {
	return f.Entity.CraftingSpeed
}

func (f *Furnace) ProductivityBonus(recipe string) float64 {
	if f == nil {
		return 0
	}
	return f.modules.ProductivityBonus(recipe)
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

func (b *Boiler) Slots() *slots {
	return &b.slots
}

func (b *Boiler) GetModules() Modules {
	return Modules{}
}

type Lab struct {
	Entity  *data.Lab
	slots   slots
	modules Modules
}

func NewLab(spec *data.Lab) *Lab {
	var mods Modules
	if n := spec.ModuleSpecification.ModuleSlots; n > 0 {
		mods = make(Modules, n)
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

func (l *Lab) Slots() *slots {
	return &l.slots
}

func (l *Lab) GetModules() Modules {
	return l.modules
}

func (l *Lab) SetModules(mod Modules) error {
	if modSlots := l.Entity.ModuleSpecification.ModuleSlots; len(mod) > modSlots {
		return ErrTooManyModules{machine: l.Name(), max: modSlots, got: len(mod)}
	}

	l.modules = mod
	return nil
}

func (l *Lab) ProductivityBonus(recipe string) float64 {
	return l.modules.ProductivityBonus(recipe)
}
