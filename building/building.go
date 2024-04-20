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

type CraftStatus byte

const (
	CraftStatusNoRecipe CraftStatus = iota
	CraftStatusRunning
	CraftStatusWaitingForInput
	CraftStatusOutputBlocked
)

type CraftingBuilding interface {
	Building
	EnergySource() data.EnergySource
	EnergyUsage() float64
	CraftingSpeed() float64

	// the current recipe that's set or (for furnaces) computed based on the input
	Recipe() *data.Recipe

	// Current status of the machine. Crafting is simplified and assumed to be instant
	// as long as the ingredients are present and there's space to put the products
	Status() CraftStatus

	// applies one crafting cycle, accounting for inventory limitations and
	// module bonuses, and returns the status of the machine
	DoCraft() CraftStatus
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

	recipe *data.Recipe

	prodBonusProgress float64
	status            CraftStatus

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

func (a *Assembler) SetRecipe(recipe *data.Recipe) map[string]int {

	if a.recipe != nil && a.recipe.Name == recipe.Name {
		return nil
	}

	if !a.Entity.CanCraft(recipe) {
		return nil
	}

	a.prodBonusProgress = 0
	inv := map[string]int{}

	for item, n := range a.input.data {
		inv[item] += n
	}
	limits := make([]string, 0, len(recipe.Ingredients))
	for _, ing := range recipe.Ingredients {
		limits = append(limits, ing.Name)
	}

	a.input = newInventory(len(limits), limits)

	for item, n := range a.output.data {
		inv[item] += n
	}

	limits = make([]string, 0, len(recipe.Results))
	for _, prod := range recipe.GetResults() {
		limits = append(limits, prod.Name)
	}
	a.output = newInventory(len(limits), limits)
	a.recipe = recipe
	a.status = CraftStatusWaitingForInput

	return inv
}

func (a *Assembler) Recipe() *data.Recipe {
	return a.recipe
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

func (a *Assembler) DoCraft() CraftStatus {

	rec := a.recipe

	if rec == nil {
		return CraftStatusNoRecipe
	}

	// check productivity bonus output
	if a.prodBonusProgress >= 1 {
		if !a.output.canAdd(rec, 1, false) {
			a.status = CraftStatusOutputBlocked
			return CraftStatusOutputBlocked
		}

		for _, p := range rec.GetResults() {
			if !p.IsFluid {
				_ = a.output.Put(p.Name, p.Amount)
			}
		}
		a.prodBonusProgress -= 1
	}

	// check if we have enough input in the machine
	canStart := true
	for _, ing := range rec.Ingredients {
		if ing.IsFluid {
			continue
		}
		if a.input.Count(ing.Name) < ing.Amount {
			canStart = false
			break
		}
	}
	if !canStart {
		return CraftStatusWaitingForInput
	}

	// check if we have enough space in the output
	if !a.output.canAdd(rec, 1, false) {
		a.status = CraftStatusOutputBlocked
		return CraftStatusOutputBlocked
	}

	// take one recipe's worth of input
	for _, ing := range rec.Ingredients {
		if ing.IsFluid {
			continue
		}
		_ = a.input.Take(ing.Name, ing.Amount)
	}

	// add one recipe's worth of output
	for _, prod := range rec.GetResults() {
		if prod.IsFluid {
			continue
		}
		_ = a.output.Put(prod.Name, prod.Amount)
	}

	// increment prod bonus
	a.prodBonusProgress += a.ProductivityBonus(rec.Name)

	a.status = CraftStatusRunning
	return CraftStatusRunning

}

func (a *Assembler) Status() CraftStatus {
	return a.status
}

type Furnace struct {
	Entity *data.Furnace
	slots  slots

	fuel   *inventory
	input  *inventory
	output *inventory

	prodBonusProgress float64
	status            CraftStatus
	recipe            *data.Recipe

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

func (f *Furnace) Recipe() *data.Recipe {

	var rec *data.Recipe
	for i, n := range f.input.data {
		if n > 0 {
			rec = data.GetSmeltingRecipe(i)
			break
		}
	}

	if rec != nil {
		if rec != f.recipe {
			f.status = CraftStatusWaitingForInput
			f.prodBonusProgress = 0
		}

		f.recipe = rec
		return rec
	}

	return f.recipe
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

func (f *Furnace) DoCraft() CraftStatus {

	rec := f.Recipe()
	if rec == nil {
		f.status = CraftStatusNoRecipe
		return CraftStatusNoRecipe
	}

	// check productivity bonus output
	if f.prodBonusProgress >= 1 {
		if !f.output.canAdd(rec, 1, false) {
			f.status = CraftStatusOutputBlocked
			return CraftStatusOutputBlocked
		}

		for _, p := range rec.GetResults() {
			if !p.IsFluid {
				_ = f.output.Put(p.Name, p.Amount)
			}
		}
		f.prodBonusProgress -= 1
	}

	// check if we have enough input in the machine
	canStart := true
	for _, ing := range rec.Ingredients {
		if ing.IsFluid {
			continue
		}
		if f.input.Count(ing.Name) < ing.Amount {
			canStart = false
			break
		}
	}
	if !canStart {
		f.status = CraftStatusWaitingForInput
		return CraftStatusWaitingForInput
	}

	// check if we have enough space in the output
	if !f.output.canAdd(rec, 1, false) {
		f.status = CraftStatusOutputBlocked
		return CraftStatusOutputBlocked
	}

	// take one recipe's worth of input
	for _, ing := range rec.Ingredients {
		if ing.IsFluid {
			continue
		}
		_ = f.input.Take(ing.Name, ing.Amount)
	}

	// add one recipe's worth of output
	for _, prod := range rec.GetResults() {
		if prod.IsFluid {
			continue
		}
		_ = f.output.Put(prod.Name, prod.Amount)
	}

	// increment prod bonus
	f.prodBonusProgress += f.ProductivityBonus(rec.Name)

	f.status = CraftStatusRunning
	return CraftStatusRunning
}

func (f *Furnace) Status() CraftStatus {
	return f.status
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
		input: newInventory(len(spec.Inputs), spec.Inputs),
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
