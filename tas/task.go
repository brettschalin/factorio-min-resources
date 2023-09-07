package tas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"

	"github.com/brettschalin/factorio-min-resources/building"
	"github.com/brettschalin/factorio-min-resources/calc"
	"github.com/brettschalin/factorio-min-resources/constants"
	"github.com/brettschalin/factorio-min-resources/data"
	"github.com/brettschalin/factorio-min-resources/geo"
)

/**** DEFINITIONS ****/

type Task interface {
	ID() string
	Type() TaskType
	Prerequisites() *Tasks
	Export() []byte
}

type Tasks []Task

func (t *Tasks) Add(tasks ...Task) {
	*t = append(*t, tasks...)
}

type TaskType int

const (
	TaskUnknown TaskType = iota
	TaskWalk
	TaskWait
	TaskCraft
	TaskBuild
	TaskTake
	TaskPut
	TaskRecipe
	TaskTech
	TaskMine
	TaskSpeed
	TaskLaunch

	// used internally for prerequisite definitions
	taskPrereq
)

func (t TaskType) String() string {
	switch t {
	case TaskWalk:
		return "walk"
	case TaskWait:
		return "wait"
	case TaskCraft:
		return "craft"
	case TaskBuild:
		return "build"
	case TaskTake:
		return "take"
	case TaskPut:
		return "put"
	case TaskRecipe:
		return "recipe"
	case TaskTech:
		return "tech"
	case TaskMine:
		return "mine"
	case TaskSpeed:
		return "speed"
	case TaskLaunch:
		return "launch"
	default:
		return "unknown"
	}
}

func (t *TaskType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

type baseTask struct {
	id      string
	prereqs Tasks
}

var ids = map[TaskType]int{}

func (t *baseTask) getID(typ TaskType) string {
	if t.id != "" {
		return t.id
	}

	n := ids[typ]
	ids[typ]++

	t.id = fmt.Sprintf("%s_%d", typ.String(), n)
	return t.id
}

func (t *baseTask) ID() string {
	return t.getID(t.Type())
}

func (t *baseTask) Type() TaskType {
	return TaskUnknown
}

func (t *baseTask) Prerequisites() *Tasks {
	return &t.prereqs
}

func (t baseTask) fmtPrereqs() []byte {

	if len(t.prereqs) == 0 {
		return []byte("nil")
	}
	out := bytes.Buffer{}
	out.WriteByte('{')
	out.Write([]byte(t.prereqs[0].ID()))

	for _, p := range t.prereqs[1:] {
		out.WriteByte(',')
		out.Write([]byte(" " + p.ID()))
	}
	out.WriteByte('}')
	return out.Bytes()
}

func (t baseTask) export(id, args string, typ TaskType) []byte {
	out := bytes.Buffer{}

	out.WriteString(id + " = add_task(")
	out.WriteString(`"` + typ.String() + `", `)
	out.Write(t.fmtPrereqs())
	out.Write([]byte(", {"))
	out.WriteString(args)

	out.WriteString("})\n")

	return out.Bytes()
}

type taskCraft struct {
	baseTask
	Recipe string
	Amount uint
}

func (t *taskCraft) ID() string {
	return t.getID(t.Type())
}

func (t *taskCraft) Type() TaskType {
	return TaskCraft
}

func (t *taskCraft) Export() []byte {
	return t.export(
		t.ID(),
		fmt.Sprintf(`item = %q, amount = %d`, t.Recipe, t.Amount),
		TaskCraft,
	)
}

type taskWalk struct {
	baseTask
	Location geo.Point
}

func (t *taskWalk) ID() string {
	return t.getID(t.Type())
}

func (t *taskWalk) Type() TaskType {
	return TaskWalk
}

func (t *taskWalk) Export() []byte {
	return t.export(
		t.ID(),
		fmt.Sprintf(`location = {x = %.2f, y = %.2f}`, t.Location.X, t.Location.Y),
		TaskWalk,
	)
}

type taskWait struct {
	baseTask

	// inventory has a certain # of items
	Entity string
	Slot   constants.Inventory
	Item   string
	Amount uint
	Exact  bool

	// n ticks
	// NTicks uint
}

func (t *taskWait) ID() string {
	return t.getID(t.Type())
}

func (t *taskWait) Type() TaskType {
	return TaskWait
}

func (t *taskWait) Export() []byte {
	return t.export(
		t.ID(),
		fmt.Sprintf(`done = has_inventory(%q, %q, %d, %t, %s)`,
			t.Entity, t.Item, t.Amount, t.Exact, t.Slot),
		TaskWait,
	)
}

type taskMine struct {
	baseTask

	// minable resource
	Resource string
	Amount   uint

	// building
	Entity string
	N      int
}

func (t *taskMine) ID() string {
	return t.getID(t.Type())
}

func (t *taskMine) Type() TaskType {
	return TaskMine
}

func (t *taskMine) Export() []byte {
	var args string
	if t.Resource != "" {
		args = fmt.Sprintf(`resource = %q, amount = %d`, t.Resource, t.Amount)
	} else {
		if t.N != 0 {
			args = fmt.Sprintf(`entity = %q, n = %d`, t.Entity, t.N)
		} else {
			args = fmt.Sprintf(`entity = %q`, t.Entity)
		}
	}
	return t.export(
		t.ID(),
		args,
		TaskMine,
	)
}

type taskBuild struct {
	baseTask
	Entity string
	N      int
}

func (t *taskBuild) ID() string {
	return t.getID(t.Type())
}

func (t *taskBuild) Type() TaskType {
	return TaskBuild
}

func (t *taskBuild) Export() []byte {
	args := fmt.Sprintf(`entity = %q`, t.Entity)

	if t.N != 0 {
		args += fmt.Sprintf(`, n = %d`, t.N)
	}

	return t.export(
		t.ID(),
		args,
		TaskBuild,
	)
}

type taskTake struct {
	baseTask

	Entity string
	Slot   constants.Inventory
	Item   string
	Amount uint
}

func (t *taskTake) Export() []byte {

	return t.export(
		t.ID(),
		fmt.Sprintf(`entity = %q, inventory = %s, item = %q, amount = %d`,
			t.Entity, t.Slot, t.Item, t.Amount),
		TaskTake,
	)
}

func (t *taskTake) ID() string {
	return t.getID(t.Type())
}

func (t *taskTake) Type() TaskType {
	return TaskTake
}

type taskPut struct {
	baseTask

	Entity string
	Slot   constants.Inventory
	Item   string
	Amount uint
}

func (t *taskPut) ID() string {
	return t.getID(t.Type())
}

func (t *taskPut) Type() TaskType {
	return TaskPut
}

func (t *taskPut) Export() []byte {

	return t.export(
		t.ID(),
		fmt.Sprintf(`entity = %q, inventory = %s, item = %q, amount = %d`,
			t.Entity, t.Slot, t.Item, t.Amount),
		TaskPut,
	)
}

type taskRecipe struct {
	baseTask
	Entity string
	Recipe string
}

func (t *taskRecipe) ID() string {
	return t.getID(t.Type())
}

func (t *taskRecipe) Type() TaskType {
	return TaskRecipe
}

func (t *taskRecipe) Export() []byte {
	return t.export(
		t.ID(),
		fmt.Sprintf(`entity = %q, recipe = %q`, t.Entity, t.Recipe),
		TaskRecipe,
	)
}

type taskTech struct {
	baseTask
	Tech string
}

func (t *taskTech) ID() string {
	return t.getID(t.Type())
}

func (t *taskTech) Type() TaskType {
	return TaskTech
}

func (t *taskTech) Export() []byte {
	return t.export(
		t.ID(),
		fmt.Sprintf(`tech = %q`, t.Tech),
		TaskTech,
	)
}

type taskSpeed struct {
	baseTask
	Speed float64
}

func (t *taskSpeed) ID() string {
	return t.getID(t.Type())
}

func (t *taskSpeed) Type() TaskType {
	return TaskSpeed
}

func (t *taskSpeed) Export() []byte {
	return t.export(
		t.ID(),
		fmt.Sprintf(`n = %.2f`, t.Speed),
		TaskSpeed,
	)
}

type taskLaunch struct {
	baseTask
}

func (t *taskLaunch) ID() string {
	return t.getID(t.Type())
}

func (t *taskLaunch) Type() TaskType {
	return TaskLaunch
}

func (t *taskLaunch) Export() []byte {
	return t.export(
		t.ID(),
		``,
		TaskLaunch,
	)
}

// this is intentionally given TaskUnknown for its type. Do not use it as anything except a prerequisite
// as it does hacky stuff to the IDs to make the prereqs work
type taskPrereqWait struct {
	baseTask
	Entity string
	Slot   constants.Inventory
	Item   string
	Amount uint
	Exact  bool
}

func (t *taskPrereqWait) ID() string {
	return fmt.Sprintf(`has_inventory(%q, %q, %d, %t, %s)`,
		t.Entity, t.Item, t.Amount, t.Exact, t.Slot)
}

func (t *taskPrereqWait) Export() []byte {
	return []byte(t.ID())
}

func (t *taskPrereqWait) Type() TaskType {
	return taskPrereq
}

/**** FUNCTIONS ****/

// Build constructs a building facing the given direction. Locations
// are hardcoded in locations.lua
func Build(entity string, n int) Task {
	return &taskBuild{
		Entity: entity,
		N:      n,
	}
}

// Craft starts a handcrafting action
func Craft(recipe string, amount uint) Task {
	return &taskCraft{
		Recipe: recipe,
		Amount: amount,
	}
}

// Launch starts the rocket launch sequence
func Launch() Task {
	return &taskLaunch{}
}

// MineEntity mines a building. Locations are hardcoded
// in locations.lua
func MineEntity(entity string, n int) Task {
	return &taskMine{
		Entity: entity,
		N:      n,
	}
}

// MineResource mines a resource (likely ore). As with the other
// methods, the locations are hardcoded in locations.lua
func MineResource(resource string, amount uint) Task {
	return &taskMine{
		Resource: resource,
		Amount:   amount,
	}
}

// Transfer takes resources from or adds them to the given inventory
func Transfer(entity, item string, slot constants.Inventory, amount uint, take bool) Task {
	if take {
		return &taskTake{
			Entity: entity,
			Item:   item,
			Slot:   slot,
			Amount: amount,
		}
	}
	return &taskPut{
		Entity: entity,
		Item:   item,
		Slot:   slot,
		Amount: amount,
	}
}

// Recipe sets the current recipe on the given machine
func Recipe(entity, recipe string) Task {
	return &taskRecipe{
		Entity: entity,
		Recipe: recipe,
	}
}

// Tech starts researching the given technology
func Tech(tech string) Task {
	return &taskTech{
		Tech: tech,
	}
}

// WaitInventory pauses task execution until certain inventory criteria are met
func WaitInventory(entity, item string, slot constants.Inventory, amount uint, exact bool) Task {
	return &taskWait{
		Entity: entity,
		Item:   item,
		Slot:   slot,
		Amount: amount,
		Exact:  exact,
	}
}

// PrereqWait is like WaitInventory but used for prerequisite definitions
func PrereqWait(entity, item string, slot constants.Inventory, amount uint, exact ...bool) Task {
	var e bool
	if len(exact) > 0 {
		e = exact[0]
	}
	return &taskPrereqWait{
		Entity: entity,
		Item:   item,
		Slot:   slot,
		Amount: amount,
		Exact:  e,
	}
}

// Speed sets the game speed. This can also safely be done with the in-game console by typing `/c game.speed = <n>`
func Speed(speed float64) Task {
	return &taskSpeed{
		Speed: speed,
	}
}

// func WaitN(ticks uint) Task {
// 	return &taskWait{
// 		NTicks: ticks,
// 	}
// }

// Walk moves the character to the given location
func Walk(location geo.Point) Task {
	return &taskWalk{
		Location: location,
	}
}

func MachineCraft(recipe, machine string, amount uint) Tasks {
	tasks := Tasks{}

	// only assemblers can have set recipes
	if m := data.GetAssemblingMachine(machine); m != nil {
		tasks.Add(Recipe(machine, recipe))
	}

	rec := data.GetRecipe(recipe)
	for _, i := range rec.Ingredients {
		if i.IsFluid {
			continue
		}
		tasks.Add(Transfer(machine, i.Name, constants.InventoryAssemblingMachineInput, amount*uint(i.Amount), false))
	}

	for _, p := range rec.GetResults() {
		if p.IsFluid {
			continue
		}
		tasks.Add(
			WaitInventory(machine, p.Name, constants.InventoryAssemblingMachineOutput, amount*uint(p.Amount), true),
			Transfer(machine, p.Name, constants.InventoryAssemblingMachineOutput, amount*uint(p.Amount), true),
		)
	}

	return tasks
}

func min[T ~int | ~uint](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// MineAndSmelt properly intersperses mining, waiting, and transferring
// ores to work around stack size limitations. This does assume the furnace is properly fueled
func MineAndSmelt(ore, furnace string, amount uint) Tasks {
	item := data.GetItem(ore)
	batchSize := uint(item.StackSize)

	amt := min(amount, batchSize)

	tasks := Tasks{
		MineResource(ore, amt),
	}

	rec := data.GetSmeltingRecipe(ore)
	plate := rec.GetResults()[0]

	for amount > 0 {

		tasks.Add(
			Transfer(furnace, ore, constants.InventoryFurnaceSource, amt, false),
		)

		// mine the next batch, but only if we need to
		amount -= amt
		nextBatch := min(amount, batchSize)
		if nextBatch > 0 {
			tasks.Add(MineResource(ore, nextBatch))
		}

		// some recipes (eg, stone-brick) are not a 1 ore:1 plate ratio. Adjust as needed
		takeAmt := uint(float64(amt) * (float64(plate.Amount) / float64(rec.Ingredients.Amount(ore))))

		tasks.Add(
			WaitInventory(furnace, plate.Name, constants.InventoryFurnaceResult, takeAmt, true),
			Transfer(furnace, plate.Name, constants.InventoryFurnaceResult, takeAmt, true),
		)

		amt = nextBatch
	}

	return tasks
}

// MineFuelAndSmelt does the same batching as MineAndSmelt but also handles fueling the furnace
func MineFuelAndSmelt(ore, fuel string, furnace *building.Furnace, amount uint) Tasks {
	tasks := Tasks{}

	fuelItem := data.GetItem(fuel)

	recipesPerBatch := calc.RecipesFromFuel(furnace, data.GetSmeltingRecipe(ore), fuelItem.StackSize, fuel)

	amt := min(uint(recipesPerBatch), amount)

	for {

		if amt < uint(recipesPerBatch) {
			// adjust how much we end up fueling/smelting and break the loop
			nFuel := calc.FuelFromRecipes(furnace, data.GetSmeltingRecipe(ore), int(amt), fuel)

			tasks.Add(FuelMachine(fuel, furnace.Name(), uint(math.Ceil(nFuel)))...)
			tasks.Add(MineAndSmelt(ore, furnace.Name(), amt)...)

			break
		}

		// fuel the machine and do some smelting
		tasks.Add(FuelMachine(fuel, furnace.Name(), uint(fuelItem.StackSize))...)
		tasks.Add(MineAndSmelt(ore, furnace.Name(), min(uint(recipesPerBatch), amount))...)

		amount -= amt
		amt = min(uint(recipesPerBatch), amount)
	}
	return tasks
}

func FuelMachine(fuel, entity string, amount uint) Tasks {
	return Tasks{
		MineResource(fuel, amount),
		Transfer(entity, fuel, constants.InventoryFuel, amount, false),
	}
}
