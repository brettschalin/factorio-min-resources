package building

import (
	"sort"

	"github.com/brettschalin/factorio-min-resources/constants"
	"github.com/brettschalin/factorio-min-resources/data"
)

type slots struct {
	Input   constants.Inventory
	Output  constants.Inventory
	Fuel    constants.Inventory
	Modules constants.Inventory
}

type Building interface {
	Name() string
	Slots() *slots
	Inventory(slot constants.Inventory) Inventory
	PutModules(modules []string) error
	TakeModules(modules []string) error
	ProductivityBonus(recipe string) float64
}

type CraftingBuilding interface {
	Building
	EnergySource() data.EnergySource
	EnergyUsage() float64
	CraftingSpeed() float64
}

func putModules(inv *Modules, modules []string) error {
	if len(modules) == 0 {
		return nil
	}
	sort.Strings(modules)
	i := 1
	last := modules[0]
	for idx := 1; idx < len(modules); idx++ {
		m := modules[idx]
		if m == last {
			i += 1
			continue
		}
		err := inv.Put(last, i)
		if err != nil {
			return err
		}
		last = m
		i = 1
	}
	return inv.Put(last, i)
}

func takeModules(inv *Modules, modules []string) error {
	if len(modules) == 0 {
		return nil
	}
	sort.Strings(modules)
	i := 1
	last := modules[0]
	for idx := 1; idx < len(modules); idx++ {
		m := modules[idx]
		if m == last {
			i += 1
			continue
		}
		err := inv.Take(last, i)
		if err != nil {
			return err
		}
		last = m
		i = 1
	}
	return inv.Take(last, i)
}

type Assembler struct {
	Entity *data.AssemblingMachine
	slots  slots

	fuel   *inventory
	input  *inventory
	output *inventory

	modules *Modules
}

func NewAssembler(spec *data.AssemblingMachine) *Assembler {

	// vanilla contains no burner assemblers of any kind (only furnaces) but mods might so account for that here
	var fuelSlot constants.Inventory
	var fuelInv *inventory
	if spec.IsBurner() {
		fuelSlot = constants.InventoryFuel
		fuelInv = newInventory(1, nil)
	}

	a := &Assembler{
		Entity: spec,
		slots: slots{
			Input:   constants.InventoryAssemblingMachineInput,
			Output:  constants.InventoryAssemblingMachineOutput,
			Fuel:    fuelSlot,
			Modules: constants.InventoryAssemblingMachineModules,
		},
		fuel:   fuelInv,
		input:  newInventory(1, nil),
		output: newInventory(1, nil),
	}
	a.modules = &Modules{machine: a, maxSlots: spec.ModuleSpecification.ModuleSlots}

	return a
}

func (a *Assembler) SetRecipe(recipe *data.Recipe) map[string]float64 {
	// TODO: make new inventories depending on ingredients/products
	return nil
}

func (a *Assembler) Name() string {
	return a.Entity.Name
}

func (a *Assembler) Slots() *slots {
	return &a.slots
}

func (a *Assembler) Inventory(slot constants.Inventory) Inventory {
	switch slot {
	case constants.InventoryFuel:
		return a.fuel
	case constants.InventoryAssemblingMachineInput:
		return a.input
	case constants.InventoryAssemblingMachineOutput:
		return a.output
	case constants.InventoryAssemblingMachineModules:
		return a.modules
	}
	return nil
}

func (a *Assembler) PutModules(modules []string) error {
	return putModules(a.modules, modules)
}

func (a *Assembler) TakeModules(modules []string) error {
	return takeModules(a.modules, modules)
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
	if !a.Entity.CanCraft(data.GetRecipe(recipe)) {
		return 0
	}
	return a.modules.ProductivityBonus(recipe)
}

type Furnace struct {
	Entity *data.Furnace
	slots  slots

	fuel   *inventory
	input  *inventory
	output *inventory

	modules *Modules
}

func NewFurnace(spec *data.Furnace) *Furnace {
	// electric furnaces don't have fuel inputs
	var fuelSlot constants.Inventory
	var fuelInv *inventory
	if spec.IsBurner() {
		fuelSlot = constants.InventoryFuel
		fuelInv = newInventory(1, nil)
	}

	f := &Furnace{
		Entity: spec,
		slots: slots{
			Input:   constants.InventoryFurnaceSource,
			Output:  constants.InventoryFurnaceResult,
			Fuel:    fuelSlot,
			Modules: constants.InventoryFurnaceModules,
		},

		fuel:   fuelInv,
		input:  newInventory(1, nil),
		output: newInventory(1, nil),
	}

	f.modules = &Modules{machine: f, maxSlots: spec.ModuleSpecification.ModuleSlots}

	return f
}

func (f *Furnace) Name() string {
	return f.Entity.Name
}

func (f *Furnace) Slots() *slots {
	return &f.slots
}

func (f *Furnace) Inventory(slot constants.Inventory) Inventory {
	switch slot {
	case constants.InventoryFuel:
		return f.fuel
	case constants.InventoryFurnaceSource:
		return f.input
	case constants.InventoryFurnaceResult:
		return f.output
	case constants.InventoryFurnaceModules:
		return f.modules
	}
	return nil
}

func (f *Furnace) PutModules(modules []string) error {
	return putModules(f.modules, modules)
}

func (f *Furnace) TakeModules(modules []string) error {
	return takeModules(f.modules, modules)
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
	if !f.Entity.CanCraft(data.GetRecipe(recipe)) {
		return 0
	}
	return f.modules.ProductivityBonus(recipe)
}

type Boiler struct {
	Entity *data.Boiler
	slots  slots
	fuel   *inventory

	// boilers can't hold modules. This exists to keep compatibility with the Building interface
	// and is given a max size of zero on initialization
	modules *Modules

	// I only care about the effective power conversion,
	// so this will act like a combined boiler/steam engine
	// where water and fuel directly produce electricity.
}

func NewBoiler(spec *data.Boiler) *Boiler {
	b := &Boiler{
		Entity: spec,
		slots: slots{
			Fuel: constants.InventoryFuel,
		},
		fuel: newInventory(1, nil),
	}

	b.modules = &Modules{machine: b, maxSlots: 0}

	return b
}

func (b *Boiler) Name() string {
	return b.Entity.Name
}

func (b *Boiler) Slots() *slots {
	return &b.slots
}

func (b *Boiler) PutModules(modules []string) error {
	return putModules(b.modules, modules)
}

func (b *Boiler) TakeModules(modules []string) error {
	return takeModules(b.modules, modules)
}

func (b *Boiler) Inventory(slot constants.Inventory) Inventory {
	if slot == constants.InventoryFuel {
		return b.fuel
	}
	return nil
}

func (b *Boiler) ProductivityBonus(recipe string) float64 {
	return 0
}

type Lab struct {
	Entity  *data.Lab
	slots   slots
	input   *inventory
	modules *Modules
}

func NewLab(spec *data.Lab) *Lab {
	l := &Lab{
		Entity: spec,
		slots: slots{
			Input:   constants.InventoryLabInput,
			Modules: constants.InventoryLabModules,
		},
		input: newInventory(len(spec.Inputs), nil),
	}

	l.modules = &Modules{machine: l, maxSlots: spec.ModuleSpecification.ModuleSlots}

	return l
}

func (l *Lab) Name() string {
	return l.Entity.Name
}

func (l *Lab) Slots() *slots {
	return &l.slots
}

func (l *Lab) PutModules(modules []string) error {
	return putModules(l.modules, modules)
}

func (l *Lab) TakeModules(modules []string) error {
	return takeModules(l.modules, modules)
}

func (l *Lab) ProductivityBonus(_ string) float64 {
	return l.modules.ProductivityBonus("")
}

func (l *Lab) Inventory(slot constants.Inventory) Inventory {
	if slot == constants.InventoryLabInput {
		return l.input
	} else if slot == constants.InventoryLabModules {
		return l.modules
	}
	return nil
}
