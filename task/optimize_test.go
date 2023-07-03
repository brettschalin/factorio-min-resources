package task

import (
	"encoding/json"
	"testing"

	"github.com/brettschalin/factorio-min-resources/building"
	"github.com/brettschalin/factorio-min-resources/data"
	"github.com/brettschalin/factorio-min-resources/state"
	"github.com/r3labs/diff/v3"
)

type opTestSpec struct {
	name          string
	task          *Task
	s             *state.State
	expected      *Task
	expectedState *state.State
}

func newState(inventory map[string]int, techs map[string]bool) *state.State {
	if inventory == nil {
		inventory = map[string]int{}
	}
	if techs == nil {
		techs = map[string]bool{}
	}

	f := data.GetFurnace("stone-furnace")
	asm := data.GetAssemblingMachine("assembling-machine-1")

	return &state.State{
		Inventory:      inventory,
		TechResearched: techs,
		Furnace:        building.NewFurnace(f),
		Assembler:      building.NewAssembler(asm),
	}
}

func TestOptimize(t *testing.T) {

	assembler := newState(nil, nil).Assembler

	for _, test := range []opTestSpec{
		{
			name: "simple crafting test",
			task: NewCraft(map[string]int{"iron-gear-wheel": 1}),
			s: newState(map[string]int{
				"iron-plate": 3,
			}, nil),
			expected: &Task{
				Type:   TaskHandcraft,
				Item:   "iron-gear-wheel",
				Amount: 1,
			},
			expectedState: newState(map[string]int{
				"iron-plate":      1,
				"iron-gear-wheel": 1,
			}, nil),
		},
		{
			name: "craft with machine",
			task: NewCraft(map[string]int{"engine-unit": 10}),
			s: newState(map[string]int{
				"iron-plate":  40,
				"steel-plate": 10,
			}, nil),
			expected: &Task{
				Type: TaskMeta,
				Prerequisites: []*Task{
					{
						Type:   TaskHandcraft,
						Item:   "iron-gear-wheel",
						Amount: 10,
					},
					{
						Type:   TaskHandcraft,
						Item:   "pipe",
						Amount: 20,
					},
					NewTransfer(assembler.Entity.Name, assembler.Slots().Input, "steel-plate", 10, false),
					NewTransfer(assembler.Entity.Name, assembler.Slots().Input, "iron-gear-wheel", 10, false),
					NewTransfer(assembler.Entity.Name, assembler.Slots().Input, "pipe", 20, false),
					NewWait(assembler.Entity.Name, assembler.Slots().Output, "engine-unit", 10),
					NewTransfer(assembler.Entity.Name, assembler.Slots().Output, "engine-unit", 10, true),
				},
			},
			expectedState: newState(map[string]int{
				"engine-unit": 10,
			}, nil),
		},
	} {

		t.Run(test.name, func(tt *testing.T) {
			task := test.task

			// Optimize() does this but doesn't let us define our own state first
			pass1(task, test.s.Copy())
			task.Prune()
			pass2(task, test.s)
			task.Prune()

			taskDiff, _ := diff.Diff(task, test.expected)
			if len(taskDiff) > 0 {
				tk, err := json.Marshal(map[string]any{
					"got":      task,
					"expected": test.expected,
				})
				if err != nil {
					tt.Fatalf("error encoding tasks: %v", err)
				}

				tt.Logf("%s", tk)

				c, _ := json.Marshal(taskDiff)
				tt.Errorf("tasks not identical: %s", string(c))
			}

			for k, v := range test.s.Inventory {
				if v == 0 {
					delete(test.s.Inventory, k)
				}
			}

			stateDiff, _ := diff.Diff(test.s, test.expectedState)
			if len(stateDiff) > 0 {
				c, _ := json.Marshal(stateDiff)
				tt.Errorf("wrong state: %s", string(c))
			}

		})
	}
}
