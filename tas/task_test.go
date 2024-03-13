package tas

import (
	"log"
	"math"
	"os"
	"testing"

	"github.com/brettschalin/factorio-min-resources/building"
	"github.com/brettschalin/factorio-min-resources/constants"
	"github.com/brettschalin/factorio-min-resources/data"
)

func TestMain(m *testing.M) {

	err := data.Init(
		"../data/data-raw-dump.json",
	)

	if err != nil {
		log.Fatalf("could not load data: %v", err)
	}

	assembler = building.NewAssembler(data.GetAssemblingMachine("assembling-machine-1"))
	stoneFurnace = building.NewFurnace(data.GetFurnace("stone-furnace"))

	os.Exit(m.Run())
}

type craftTestCase struct {
	name                 string
	fuel                 string
	ore                  string
	recipe               string
	amount               uint
	machine              building.CraftingBuilding
	extraFuel            float64
	expectedTasks        Tasks
	expectedLeftoverFuel *float64
}

func addr[T any](x T) *T {
	return &x
}

func floatsEqual[T ~float32 | ~float64](f1, f2 T) bool {
	const tolerance = 1e-10
	return math.Abs(float64(f1-f2)) <= tolerance
}

var (
	assembler    *building.Assembler
	stoneFurnace *building.Furnace
)

func (c *craftTestCase) verify(t *testing.T, tasks Tasks, leftoverFuel float64) {

	if c.expectedLeftoverFuel != nil {

		if !floatsEqual(*c.expectedLeftoverFuel, leftoverFuel) {
			t.Errorf(`wrong amount of leftover fuel (wanted %f but got %f)`, *c.expectedLeftoverFuel, leftoverFuel)
		}
	}

	if len(c.expectedTasks) > 0 {

		if len(tasks) != len(c.expectedTasks) {
			t.Fatalf(`wrong number of tasks returned (wanted %d but got %d)`, len(c.expectedTasks), len(tasks))
		}

		for i, task := range tasks {
			expected := c.expectedTasks[i]
			if task.Type() != expected.Type() {
				t.Errorf(`wrong type for task %d (wanted %q but got %q)`, i, expected.Type(), task.Type())
			}
			e := string(task.Export())
			delete(ids, task.Type())
			e2 := string(expected.Export())
			delete(ids, task.Type())

			if e != e2 {
				t.Errorf(`tasks at index %d not equal (wanted %q but got %q)`, i, e2, e)
			}
		}
	}
}

func TestMachineCraft(t *testing.T) {

	// MachineCraft does not do batching of any kind. Future iterations of this code will handle that and modules but we are not there yet
	for _, test := range []craftTestCase{
		{
			name:    "craft copper cables",
			machine: assembler,
			recipe:  "copper-cable",
			amount:  20,
			expectedTasks: Tasks{
				Recipe(assembler.Name(), "copper-cable"),
				Transfer(assembler.Name(), "copper-plate", constants.InventoryAssemblingMachineInput, 20, false),
				WaitInventory(assembler.Name(), "copper-cable", constants.InventoryAssemblingMachineOutput, 40, true),
				Transfer(assembler.Name(), "copper-cable", constants.InventoryAssemblingMachineOutput, 40, true),
			},
			expectedLeftoverFuel: addr(float64(0)),
		},
		{
			name:    ">1 stack of ingredients",
			machine: assembler,
			recipe:  "iron-gear-wheel",
			amount:  60,
			expectedTasks: Tasks{
				Recipe(assembler.Name(), "iron-gear-wheel"),
				Transfer(assembler.Name(), "iron-plate", constants.InventoryAssemblingMachineInput, 120, false),
				WaitInventory(assembler.Name(), "iron-gear-wheel", constants.InventoryAssemblingMachineOutput, 60, true),
				Transfer(assembler.Name(), "iron-gear-wheel", constants.InventoryAssemblingMachineOutput, 60, true),
			},
			expectedLeftoverFuel: addr(float64(0)),
		},
		{
			name:    ">1 stack of products",
			machine: assembler,
			recipe:  "copper-cable",
			amount:  150,
			expectedTasks: Tasks{
				Recipe(assembler.Name(), "copper-cable"),
				Transfer(assembler.Name(), "copper-plate", constants.InventoryAssemblingMachineInput, 150, false),
				WaitInventory(assembler.Name(), "copper-cable", constants.InventoryAssemblingMachineOutput, 300, true),
				Transfer(assembler.Name(), "copper-cable", constants.InventoryAssemblingMachineOutput, 300, true),
			},
			expectedLeftoverFuel: addr(float64(0)),
		},
		{
			name:    "multiple ingredients",
			machine: assembler,
			recipe:  "electronic-circuit",
			amount:  50,
			expectedTasks: Tasks{
				Recipe(assembler.Name(), "electronic-circuit"),
				Transfer(assembler.Name(), "iron-plate", constants.InventoryAssemblingMachineInput, 50, false),
				Transfer(assembler.Name(), "copper-cable", constants.InventoryAssemblingMachineInput, 150, false),
				WaitInventory(assembler.Name(), "electronic-circuit", constants.InventoryAssemblingMachineOutput, 50, true),
				Transfer(assembler.Name(), "electronic-circuit", constants.InventoryAssemblingMachineOutput, 50, true),
			},
			expectedLeftoverFuel: addr(float64(0)),
		},
	} {
		t.Run(test.name, func(tt *testing.T) {
			tasks, fuel := MachineCraft(test.recipe, test.machine, test.amount, test.fuel)
			test.verify(tt, tasks, fuel)
		})
	}
}

func TestMineAndSmelt(t *testing.T) {
	for _, test := range []craftTestCase{
		{
			name:                 "smelt iron ore",
			fuel:                 constants.PreferredFuel,
			ore:                  "iron-ore",
			amount:               25,
			machine:              stoneFurnace,
			expectedLeftoverFuel: addr(float64(1.8)),
			expectedTasks: Tasks{
				MineResource("iron-ore", 25),
				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 25, false),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 25, true),
			},
		},
		{
			name:                 "smelt iron plates",
			fuel:                 constants.PreferredFuel,
			ore:                  "iron-plate",
			amount:               50,
			machine:              stoneFurnace,
			expectedLeftoverFuel: addr(float64(3.6)),
			expectedTasks: Tasks{
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceSource, 50, false),
				WaitInventory(stoneFurnace.Name(), "steel-plate", constants.InventoryFurnaceResult, 10, true),
				Transfer(stoneFurnace.Name(), "steel-plate", constants.InventoryFurnaceResult, 10, true),
			},
		},
		{
			name:                 "smelt a lot of iron ore",
			fuel:                 constants.PreferredFuel,
			ore:                  "iron-ore",
			amount:               200,
			machine:              stoneFurnace,
			expectedLeftoverFuel: addr(float64(14.4)),
			expectedTasks: Tasks{
				MineResource("iron-ore", 50),
				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
			},
		},
		{
			name:                 "smelt stone bricks",
			fuel:                 constants.PreferredFuel,
			ore:                  "stone",
			amount:               100,
			machine:              stoneFurnace,
			expectedLeftoverFuel: addr(float64(3.6)),
			expectedTasks: Tasks{
				MineResource("stone", 50),
				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
			},
		},
	} {
		t.Run(test.name, func(tt *testing.T) {
			tasks, fuel := MineAndSmelt(test.ore, test.machine, test.amount, test.fuel)
			test.verify(tt, tasks, fuel)
		})
	}
}

func TestMineFuelAndSmelt(t *testing.T) {

	for _, test := range []craftTestCase{
		{
			name:                 "smelt iron ore",
			fuel:                 constants.PreferredFuel,
			ore:                  "iron-ore",
			amount:               25,
			machine:              stoneFurnace,
			expectedLeftoverFuel: addr(float64(0.2)),
			expectedTasks: Tasks{
				MineResource(constants.PreferredFuel, 2),
				Transfer(stoneFurnace.Name(), constants.PreferredFuel, constants.InventoryFuel, 2, false),
				MineResource("iron-ore", 25),
				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 25, false),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 25, true),
			},
		},
		{
			name:                 "smelt iron ore with extra fuel in the furnace",
			fuel:                 constants.PreferredFuel,
			ore:                  "iron-ore",
			amount:               25,
			machine:              stoneFurnace,
			extraFuel:            0.9,
			expectedLeftoverFuel: addr(float64(0.1)),
			expectedTasks: Tasks{
				MineResource(constants.PreferredFuel, 1),
				Transfer(stoneFurnace.Name(), constants.PreferredFuel, constants.InventoryFuel, 1, false),
				MineResource("iron-ore", 25),
				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 25, false),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 25, true),
			},
		},
		{
			name:                 "smelt iron plates",
			fuel:                 constants.PreferredFuel,
			ore:                  "iron-plate",
			amount:               50,
			machine:              stoneFurnace,
			expectedLeftoverFuel: addr(float64(0.4)),
			expectedTasks: Tasks{
				MineResource(constants.PreferredFuel, 4),
				Transfer(stoneFurnace.Name(), constants.PreferredFuel, constants.InventoryFuel, 4, false),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceSource, 50, false),
				WaitInventory(stoneFurnace.Name(), "steel-plate", constants.InventoryFurnaceResult, 10, true),
				Transfer(stoneFurnace.Name(), "steel-plate", constants.InventoryFurnaceResult, 10, true),
			},
		},
		{
			name:                 "smelt a lot of iron ore",
			fuel:                 constants.PreferredFuel,
			ore:                  "iron-ore",
			amount:               200,
			machine:              stoneFurnace,
			expectedLeftoverFuel: addr(float64(0.6)),
			expectedTasks: Tasks{

				MineResource(constants.PreferredFuel, 15),
				Transfer(stoneFurnace.Name(), constants.PreferredFuel, constants.InventoryFuel, 15, false),

				MineResource("iron-ore", 50),
				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
			},
		},
		{
			name:                 "smelt a LOT of iron ore",
			fuel:                 constants.PreferredFuel,
			ore:                  "iron-ore",
			amount:               800,
			machine:              stoneFurnace,
			extraFuel:            0.65,
			expectedLeftoverFuel: addr(float64(0.05)),
			expectedTasks: Tasks{
				MineResource(constants.PreferredFuel, 50),
				Transfer(stoneFurnace.Name(), constants.PreferredFuel, constants.InventoryFuel, 50, false),

				MineResource("iron-ore", 50),
				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 50),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 3),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 3, false),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 3, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 3, true),

				MineResource(constants.PreferredFuel, 7),
				Transfer(stoneFurnace.Name(), constants.PreferredFuel, constants.InventoryFuel, 7, false),

				MineResource("iron-ore", 50),
				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 50, false),
				MineResource("iron-ore", 47),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 50, true),

				Transfer(stoneFurnace.Name(), "iron-ore", constants.InventoryFurnaceSource, 47, false),
				WaitInventory(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 47, true),
				Transfer(stoneFurnace.Name(), "iron-plate", constants.InventoryFurnaceResult, 47, true),
			},
		},
		{
			name:                 "smelt stone",
			fuel:                 constants.PreferredFuel,
			ore:                  "stone",
			amount:               1400,
			machine:              stoneFurnace,
			expectedLeftoverFuel: addr(float64(0.6)),
			expectedTasks: Tasks{
				MineResource(constants.PreferredFuel, 50),
				Transfer(stoneFurnace.Name(), constants.PreferredFuel, constants.InventoryFuel, 50, false),

				MineResource("stone", 50),
				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 50),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 50, false),
				MineResource("stone", 38),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 25, true),

				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 38, false),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 19, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 19, true),

				MineResource(constants.PreferredFuel, 1),
				Transfer(stoneFurnace.Name(), constants.PreferredFuel, constants.InventoryFuel, 1, false),

				MineResource("stone", 12),
				Transfer(stoneFurnace.Name(), "stone", constants.InventoryFurnaceSource, 12, false),
				WaitInventory(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 6, true),
				Transfer(stoneFurnace.Name(), "stone-brick", constants.InventoryFurnaceResult, 6, true),
			},
		},
	} {
		t.Run(test.name, func(tt *testing.T) {
			tasks, fuel := MineFuelAndSmelt(test.ore, test.fuel, test.machine, test.amount, test.extraFuel)
			test.verify(tt, tasks, fuel)
		})
	}
}
