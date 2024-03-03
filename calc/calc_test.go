package calc

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/brettschalin/factorio-min-resources/building"
	"github.com/brettschalin/factorio-min-resources/data"
	"github.com/brettschalin/factorio-min-resources/shims/maps"
)

func TestMain(m *testing.M) {

	err := data.Init(
		"../data/data-raw-dump.json",
	)

	if err != nil {
		log.Fatalf("could not load data: %v", err)
	}

	assemblerNoModules = building.NewAssembler(data.GetAssemblingMachine("assembling-machine-1"))
	assemblerModules = building.NewAssembler(data.GetAssemblingMachine("assembling-machine-2"))

	prodmod1 = data.GetModule("productivity-module")
	prodmod2 = data.GetModule("productivity-module-2")
	prodmod3 = data.GetModule("productivity-module-3")

	err = assemblerModules.SetModules(building.Modules{prodmod2, prodmod2})

	if err != nil {
		log.Fatalf("could not add modules to assembler: %v", err)
	}

	os.Exit(m.Run())
}

func cmpErr(e1, e2 error) bool {
	if e1 == nil {
		return e2 == nil
	}
	if e2 == nil {
		return false
	}
	return e1.Error() == e2.Error()
}

var (
	assemblerNoModules, assemblerModules *building.Assembler
	prodmod1, prodmod2, prodmod3         *data.Module
)

func TestRecipeCost(t *testing.T) {

	var tests = []struct {
		recipe       string
		amount       int
		building     building.CraftingBuilding
		expectedIng  map[string]int
		expectedProd map[string]int
	}{
		{
			recipe: "iron-plate",
			amount: 1,
			expectedIng: map[string]int{
				"iron-ore": 1,
			},
			expectedProd: map[string]int{
				"iron-plate": 1,
			},
			building: assemblerNoModules,
		},

		{
			recipe: "iron-gear-wheel",
			amount: 3,
			expectedIng: map[string]int{
				"iron-plate": 6,
			},
			expectedProd: map[string]int{
				"iron-gear-wheel": 3,
			},
			building: assemblerNoModules,
		},

		{
			recipe: "copper-cable",
			amount: 1,
			expectedIng: map[string]int{
				"copper-plate": 1,
			},
			expectedProd: map[string]int{
				"copper-cable": 2,
			},
			building: assemblerNoModules,
		},

		{
			recipe: "rocket-control-unit",
			amount: 10,
			expectedIng: map[string]int{
				"speed-module":    10,
				"processing-unit": 10,
			},
			expectedProd: map[string]int{
				"rocket-control-unit": 10,
			},
			building: assemblerNoModules,
		},
		{
			recipe: "utility-science-pack",
			amount: 12,
			expectedIng: map[string]int{
				"low-density-structure": 36,
				"processing-unit":       24,
				"flying-robot-frame":    12,
			},
			expectedProd: map[string]int{
				"utility-science-pack": 36,
			},
			building: assemblerNoModules,
		},
		{
			recipe:   "this-item-does-not-exist",
			amount:   1,
			building: assemblerNoModules,
		},
		{
			recipe: "logistic-science-pack",
			amount: 50,
			expectedIng: map[string]int{
				"inserter":       45,
				"transport-belt": 45,
			},
			expectedProd: map[string]int{
				"logistic-science-pack": 50,
			},
			building: assemblerModules,
		},
	}

	for _, test := range tests {
		actualIng, actualProd := RecipeCost(data.GetRecipe(test.recipe), test.amount, test.building)
		if !maps.Equal(actualIng, test.expectedIng) {
			t.Errorf("wrong ingredients for test '%d %s': wanted %v but got %v", test.amount, test.recipe, test.expectedIng, actualIng)
		}
		if !maps.Equal(actualProd, test.expectedProd) {
			t.Errorf("wrong products for test '%d %s': wanted %v but got %v", test.amount, test.recipe, test.expectedProd, actualProd)
		}
	}
}

func TestRecipeFullCost(t *testing.T) {
	var tests = []struct {
		item         string
		amount       int
		expectedIng  map[string]int
		expectedProd map[string]int
	}{
		{
			item:   "iron-plate",
			amount: 10,
			expectedIng: map[string]int{
				"iron-ore": 10,
			},
			expectedProd: map[string]int{
				"iron-plate": 10,
			},
		},
		{
			item:   "electronic-circuit",
			amount: 3,
			expectedIng: map[string]int{
				"iron-ore":   3,
				"copper-ore": 5,
			},
			expectedProd: map[string]int{
				"electronic-circuit": 3,
			},
		},
		{
			item:   "red-wire",
			amount: 1,
			expectedIng: map[string]int{
				"iron-ore":   1,
				"copper-ore": 2,
			},
			expectedProd: map[string]int{
				"red-wire": 1,
			},
		},
		{
			item:   "rocket-control-unit",
			amount: 20,
			expectedIng: map[string]int{
				"iron-ore":      782,
				"copper-ore":    1450,
				"coal":          140,
				"water":         350,
				"petroleum-gas": 2950,
			},
			expectedProd: map[string]int{
				"rocket-control-unit": 20,
			},
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%d %s", test.amount, test.item), func(tt *testing.T) {
			actualIng, actualProd := RecipeFullCost(data.GetRecipe(test.item), test.amount, nil)
			if !maps.Equal(actualIng, test.expectedIng) {
				tt.Errorf("wrong ingredients for test '%d %s': wanted %v but got %v", test.amount, test.item, test.expectedIng, actualIng)
			}
			if !maps.Equal(actualProd, test.expectedProd) {
				tt.Errorf("wrong products for test '%d %s': wanted %v but got %v", test.amount, test.item, test.expectedProd, actualProd)
			}
		})
	}
}

func TestRecipeAllIngredients(t *testing.T) {

	var tests = []struct {
		item        string
		amount      int
		result      data.Ingredients
		expectedErr error
	}{
		{
			item:   "electronic-circuit",
			amount: 3,
			result: data.Ingredients{
				{
					Name:   "copper-ore",
					Amount: 5,
				},
				{
					Name:   "copper-plate",
					Amount: 5,
				},
				{
					Name:   "iron-ore",
					Amount: 3,
				},
				{
					Name:   "copper-cable",
					Amount: 9,
				},
				{
					Name:   "iron-plate",
					Amount: 3,
				},
				{
					Name:   "electronic-circuit",
					Amount: 3,
				},
			},
		},
		{
			item:   "red-wire",
			amount: 1,
			result: data.Ingredients{
				{
					Name:   "copper-ore",
					Amount: 2,
				},
				{
					Name:   "iron-ore",
					Amount: 1,
				},
				{
					Name:   "copper-plate",
					Amount: 2,
				},
				{
					Name:   "iron-plate",
					Amount: 1,
				},
				{
					Name:   "copper-cable",
					Amount: 4,
				},
				{
					Name:   "electronic-circuit",
					Amount: 1,
				},
				{
					Name:   "red-wire",
					Amount: 1,
				},
			},
		},
		{
			item:   "logistic-science-pack",
			amount: 20,
			result: data.Ingredients{
				{
					Name:   "copper-ore",
					Amount: 30,
				},
				{
					Name:   "copper-plate",
					Amount: 30,
				},
				{
					Name:   "iron-ore",
					Amount: 110,
				},
				{
					Name:   "copper-cable",
					Amount: 60,
				},
				{
					Name:   "iron-plate",
					Amount: 110,
				},
				{
					Name:   "iron-gear-wheel",
					Amount: 30,
				},
				{
					Name:   "electronic-circuit",
					Amount: 20,
				},
				{
					Name:   "transport-belt",
					Amount: 20,
				},
				{
					Name:   "inserter",
					Amount: 20,
				},
				{
					Name:   "logistic-science-pack",
					Amount: 20,
				},
			},
		},
		{
			item:   "rocket-control-unit",
			amount: 20,
			result: data.Ingredients{
				{
					Name:   "copper-ore",
					Amount: 1450,
				},
				{
					Name:   "iron-ore",
					Amount: 782,
				},
				{
					Name:   "copper-plate",
					Amount: 1450,
				},
				{
					Name:   "coal",
					Amount: 140,
				},
				{
					Name:   "petroleum-gas",
					Amount: 2950,
				},
				{
					Name:   "water",
					Amount: 350,
				},
				{
					Name:   "iron-plate",
					Amount: 782,
				},
				{
					Name:   "sulfur",
					Amount: 10,
				},
				{
					Name:   "copper-cable",
					Amount: 2900,
				},
				{
					Name:   "plastic-bar",
					Amount: 280,
				},
				{
					Name:   "electronic-circuit",
					Amount: 780,
				},
				{
					Name:   "sulfuric-acid",
					Amount: 100,
				},
				{
					Name:   "advanced-circuit",
					Amount: 140,
				},
				{
					Name:   "speed-module",
					Amount: 20,
				},
				{
					Name:   "processing-unit",
					Amount: 20,
				},
				{
					Name:   "rocket-control-unit",
					Amount: 20,
				},
			},
		},
	}

	for _, test := range tests {
		actual, err := RecipeAllIngredients(data.GetRecipe(test.item), test.amount, nil)
		if !cmpErr(test.expectedErr, err) {
			t.Errorf("[%s] wrong error. Wanted %v but got %v", test.item, test.expectedErr, err)
			continue
		}

		if len(actual) != len(test.result) {
			t.Fatalf("[%s] wrong amount of ingredients. Wanted %d but got %d", test.item, len(test.result), len(actual))
		}

		for i, ing := range actual {
			if ing != test.result[i] {
				t.Errorf("[%s] wrong amount/item for index %d. Wanted %d %s but got %d %s",
					test.item, i, test.result[i].Amount, test.result[i].Name, ing.Amount, ing.Name)
			}
		}
	}
}

func TestHandcraft(t *testing.T) {
	var tests = []struct {
		item        string
		amount      uint
		inventory   map[string]uint
		expectedInv map[string]uint
		expectedErr error
	}{
		{
			item:   "iron-gear-wheel",
			amount: 5,
			inventory: map[string]uint{
				"iron-plate": 11,
			},
			expectedInv: map[string]uint{
				"iron-plate":      1,
				"iron-gear-wheel": 5,
			},
		},
		{
			item:   "advanced-circuit",
			amount: 1,
			inventory: map[string]uint{
				"iron-plate":   20,
				"copper-plate": 10,
				"plastic-bar":  10,
				"copper-cable": 3,
			},
			expectedInv: map[string]uint{
				"iron-plate":       18,
				"copper-plate":     6,
				"plastic-bar":      8,
				"copper-cable":     1,
				"advanced-circuit": 1,
			},
		},
		{
			item:   "engine-unit",
			amount: 1,

			expectedErr: ErrCantHandcraft,
		},
		{
			item:   "pistol",
			amount: 1,
			inventory: map[string]uint{
				"copper-plate": 5,
			},
			expectedErr: &ErrMissingIngredient{
				item: "iron-plate",
				n:    5,
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%d %s", test.amount, test.item), func(tt *testing.T) {
			inv, err := Handcraft(test.inventory, data.GetRecipe(test.item), test.amount)
			if !cmpErr(test.expectedErr, err) {
				tt.Fatalf("wrong error. Wanted %v but got %v", test.expectedErr, err)
			}

			for i, n := range inv {
				if n != test.expectedInv[i] {
					tt.Errorf("wrong number for inventory %q. Wanted %d but got %d", i, test.expectedInv[i], n)
				}
			}

		})
	}
}

func TestTechCost(t *testing.T) {

	var tests = []struct {
		tech     string
		expected map[string]int
	}{
		{
			tech: "automation",
			expected: map[string]int{
				"automation-science-pack": 10,
			},
		},
		{
			tech: "advanced-electronics",
			expected: map[string]int{
				"automation-science-pack": 200,
				"logistic-science-pack":   200,
			},
		},
	}

	for _, test := range tests {
		actual := TechCost(test.tech)
		if !maps.Equal(actual, test.expected) {
			t.Errorf("wrong cost for tech %q: wanted %v but got %v", test.tech, actual, test.expected)
		}
	}

}

func TestTechFullCost(t *testing.T) {
	var tests = []struct {
		tech       string
		researched map[string]bool
		expected   map[string]int
	}{
		{
			tech: "automation",
			expected: map[string]int{
				"automation-science-pack": 10,
			},
		},
		{
			tech: "advanced-electronics",
			researched: map[string]bool{
				"electronics": true,
			},
			expected: map[string]int{
				"automation-science-pack": 815,
				"logistic-science-pack":   690,
			},
		},
		{
			tech: "circuit-network",
			expected: map[string]int{
				"automation-science-pack": 215,
				"logistic-science-pack":   100,
			},
		},
	}

	for _, test := range tests {
		actual := TechFullCost(test.researched, test.tech)
		if !maps.Equal(actual, test.expected) {
			t.Errorf("wrong cost for tech %q: wanted %v but got %v", test.tech, test.expected, actual)
		}
	}
}
