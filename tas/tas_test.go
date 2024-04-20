package tas

import (
	"fmt"
	"testing"

	"github.com/brettschalin/factorio-min-resources/building"
	"github.com/brettschalin/factorio-min-resources/constants"
	"github.com/brettschalin/factorio-min-resources/data"
	"github.com/brettschalin/factorio-min-resources/state"
	"github.com/r3labs/diff/v3"
)

func TestVerifyPrereqs(t *testing.T) {

	var (
		t1 = Craft("iron-gear-wheel", 20)
		t2 = Tech("automation")
		t3 = Recipe("assembling-machine-2", "engine-unit")
	)

	t2.Prerequisites().Add(t1)

	for _, test := range []struct {
		name  string
		input TAS
		err   error
	}{
		{
			name: "valid",
			input: TAS{
				tasks: Tasks{
					t1,
					t2,
				},
			},
		},
		{
			name: "wrong order",
			input: TAS{
				tasks: Tasks{
					t2,
					t3,
					t1,
				},
			},
			err: fmt.Errorf(`task %d references unknown prerequisite %s`, 0, t1.ID()),
		},
	} {
		t.Run(test.name, func(tt *testing.T) {
			err := test.input.verifyPrereqs()
			if d, _ := diff.Diff(err, test.err); len(d) > 0 {
				tt.Fatal(d)
			}
		})
	}
}

func TestVerifyState(t *testing.T) {

	for _, test := range []struct {
		name              string
		input             TAS
		inState, outState *state.State
		err               error
	}{
		{
			name: "valid",
			input: TAS{
				tasks: Tasks{
					Build("lab", 0),
					Craft("automation-science-pack", 10),
					Tech("automation"),
					Transfer("lab", "automation-science-pack", constants.InventoryLabInput, 10, false),
				},
			},
			inState: &state.State{
				Inventory: map[string]uint{
					"lab":          1,
					"iron-plate":   20,
					"copper-plate": 15,
				},
				TechResearched: map[string]bool{},
				Buildings:      map[string]bool{},
			},
			outState: &state.State{
				TechResearched: map[string]bool{
					"automation": true,
				},
				Inventory: map[string]uint{
					"copper-plate": 5,
				},
				Buildings: map[string]bool{"lab": true},
				Lab:       building.NewLab(data.GetLab("lab")),
			},
		}, {
			name: "unresearched technology",
			input: TAS{
				tasks: Tasks{
					Tech("solar-energy"),
				},
			},
			inState:  &state.State{},
			outState: &state.State{},
			err:      fmt.Errorf(`[tech] %q: prerequisite %q not yet researched`, "solar-energy", "optics"),
		},
	} {
		t.Run(test.name, func(tt *testing.T) {
			err := test.input.verifyState(test.inState)
			if d, _ := diff.Diff(err, test.err); len(d) > 0 {
				tt.Fatal(d)
			}
			if d, _ := diff.Diff(test.inState, test.outState); len(d) > 0 {
				tt.Fatal(d)
			}
		})
	}
}
