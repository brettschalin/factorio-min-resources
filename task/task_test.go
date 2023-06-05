package task

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/brettschalin/factorio-min-resources/data"
	"github.com/brettschalin/factorio-min-resources/geo"
	"github.com/r3labs/diff/v3"
)

func TestMain(m *testing.M) {

	data.Init("../data/data-raw-dump.json")
	os.Exit(m.Run())

}

type testSpec struct {
	name     string
	args     []any // type assertion done by individual tests
	expected *Task
}

func cmpTask(t *testing.T, expected, actual *Task) {
	changes, err := diff.Diff(actual, expected, diff.SliceOrdering(true))
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) > 0 {

		c, _ := json.Marshal(changes)

		t.Errorf("tasks not identical: %s", string(c))
	}
}

func TestNewWalk(t *testing.T) {

	for _, test := range []testSpec{
		{
			name: "walking",
			args: []any{&geo.Point{1, 1}},
			expected: &Task{
				Type:     TaskWalk,
				Location: &geo.Point{1, 1},
			},
		},
	} {
		t.Run(test.name, func(tt *testing.T) {
			p := test.args[0].(*geo.Point)
			actual := NewWalk(*p)
			cmpTask(tt, test.expected, actual)
		})
	}
}

func TestNewCraft(t *testing.T) {

	for _, test := range []testSpec{

		{
			name: "iron-plate",
			args: []any{map[string]int{
				"iron-plate": 10,
			}},
			expected: &Task{
				Type:   TaskCraft,
				Item:   "iron-plate",
				Amount: 10,
				Prerequisites: []*Task{
					{
						Type:   TaskMine,
						Item:   "iron-ore",
						Amount: 10,
					},
				},
			},
		},
		{
			name: "electronic-circuit",
			args: []any{map[string]int{
				"electronic-circuit": 8,
			}},
			expected: &Task{
				Type:   TaskCraft,
				Item:   "electronic-circuit",
				Amount: 8,
				Prerequisites: []*Task{
					{
						Type:   TaskMine,
						Item:   "iron-ore",
						Amount: 8,
						Index:  0,
					},
					{
						Type:   TaskCraft,
						Item:   "iron-plate",
						Amount: 8,
					},
					{
						Type:   TaskMine,
						Item:   "copper-ore",
						Amount: 12,
					},
					{
						Type:   TaskCraft,
						Item:   "copper-plate",
						Amount: 12,
					},
					{
						Type:   TaskCraft,
						Item:   "copper-cable",
						Amount: 24,
					},
				},
			},
		},
		{
			name: "multiple ingredients",
			args: []any{map[string]int{
				"automation-science-pack": 1,
				"logistic-science-pack":   4,
			}},
			expected: &Task{
				Type: TaskMeta,
				Prerequisites: []*Task{
					{
						Type:   TaskMine,
						Item:   "copper-ore",
						Amount: 7,
					},
					{
						Type:   TaskCraft,
						Item:   "copper-plate",
						Amount: 7,
					},
					{
						Type:   TaskMine,
						Item:   "iron-ore",
						Amount: 24,
					},
					{
						Type:   TaskCraft,
						Item:   "iron-plate",
						Amount: 24,
					},
					{
						Type:   TaskCraft,
						Item:   "iron-gear-wheel",
						Amount: 7,
					},
					{
						Type:   TaskCraft,
						Item:   "automation-science-pack",
						Amount: 1,
					},
					{
						Type:   TaskCraft,
						Item:   "copper-cable",
						Amount: 12,
					},
					{
						Type:   TaskCraft,
						Item:   "electronic-circuit",
						Amount: 4,
					},
					{
						Type:   TaskCraft,
						Item:   "inserter",
						Amount: 4,
					},
					{
						Type:   TaskCraft,
						Item:   "transport-belt",
						Amount: 4,
					},
					{
						Type:   TaskCraft,
						Item:   "logistic-science-pack",
						Amount: 4,
					},
				},
			},
		},
	} {
		t.Run(test.name, func(tt *testing.T) {
			m := test.args[0].(map[string]int)
			actual := NewCraft(m)
			cmpTask(tt, test.expected, actual)
		})
	}

}

// func TestNewWait(t *testing.T) {

// }

// func TestNewBuild(t *testing.T) {

// }

// func TestNewTransfer(t *testing.T) {

// }

func TestNewTech(t *testing.T) {

	for _, test := range []testSpec{
		{
			name: "automation",
			args: []any{"automation"},
			expected: &Task{
				Type: TaskTech,
				Tech: "automation",
				Prerequisites: []*Task{
					{
						Type:   TaskCraft,
						Item:   "automation-science-pack",
						Amount: 10,
						Prerequisites: []*Task{
							{
								Type:   TaskMine,
								Item:   "copper-ore",
								Amount: 10,
							},
							{
								Type:   TaskCraft,
								Item:   "copper-plate",
								Amount: 10,
							},
							{
								Type:   TaskMine,
								Item:   "iron-ore",
								Amount: 20,
							},
							{
								Type:   TaskCraft,
								Item:   "iron-plate",
								Amount: 20,
							},
							{
								Type:   TaskCraft,
								Item:   "iron-gear-wheel",
								Amount: 10,
							},
						},
					},
				},
			},
		},
		{
			name: "landfill",
			args: []any{"landfill"},
			expected: &Task{
				Type: TaskTech,
				Tech: "landfill",
				Prerequisites: []*Task{
					{
						Type: TaskTech,
						Tech: "logistic-science-pack",
						Prerequisites: []*Task{
							{
								Type:   TaskCraft,
								Item:   "automation-science-pack",
								Amount: 75,
								Prerequisites: []*Task{
									{
										Type:   TaskMine,
										Item:   "copper-ore",
										Amount: 75,
									},
									{
										Type:   TaskCraft,
										Item:   "copper-plate",
										Amount: 75,
									},
									{
										Type:   TaskMine,
										Item:   "iron-ore",
										Amount: 150,
									},
									{
										Type:   TaskCraft,
										Item:   "iron-plate",
										Amount: 150,
									},
									{
										Type:   TaskCraft,
										Item:   "iron-gear-wheel",
										Amount: 75,
									},
								},
							}},
					},
					{
						Type: TaskMeta,
						Prerequisites: []*Task{
							{
								Type:   TaskMine,
								Item:   "copper-ore",
								Amount: 125,
							},
							{
								Type:   TaskCraft,
								Item:   "copper-plate",
								Amount: 125,
							},
							{
								Type:   TaskMine,
								Item:   "iron-ore",
								Amount: 375,
							},
							{
								Type:   TaskCraft,
								Item:   "iron-plate",
								Amount: 375,
							},
							{
								Type:   TaskCraft,
								Item:   "iron-gear-wheel",
								Amount: 125,
							},
							{
								Type:   TaskCraft,
								Item:   "automation-science-pack",
								Amount: 50,
							},
							{
								Type:   TaskCraft,
								Item:   "copper-cable",
								Amount: 150,
							},
							{
								Type:   TaskCraft,
								Item:   "electronic-circuit",
								Amount: 50,
							},
							{
								Type:   TaskCraft,
								Item:   "inserter",
								Amount: 50,
							},
							{
								Type:   TaskCraft,
								Item:   "transport-belt",
								Amount: 50,
							},
							{
								Type:   TaskCraft,
								Item:   "logistic-science-pack",
								Amount: 50,
							},
						},
					},
				},
			},
		},
	} {
		t.Run(test.name, func(tt *testing.T) {
			tech := test.args[0].(string)
			actual := NewTech(tech)
			cmpTask(tt, test.expected, actual)
		})
	}

}

// func TestNewMine(t *testing.T) {

// }

// func TestNewLaunch(t *testing.T) {

// }
