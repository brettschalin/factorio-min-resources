package calc

import (
	"log"
	"os"
	"testing"

	"github.com/brettschalin/factorio-min-resources/constants"
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

func TestRecipeCost(t *testing.T) {

	var tests = []struct {
		item         string
		amount       int
		expectedIng  map[string]int
		expectedProd map[string]int
	}{
		{
			item:   "iron-plate",
			amount: 1,
			expectedIng: map[string]int{
				"iron-ore": 1,
			},
			expectedProd: map[string]int{
				"iron-plate": 1,
			},
		},

		{
			item:   "iron-gear-wheel",
			amount: 3,
			expectedIng: map[string]int{
				"iron-plate": 6,
			},
			expectedProd: map[string]int{
				"iron-gear-wheel": 3,
			},
		},

		{
			item:   "copper-cable",
			amount: 1,
			expectedIng: map[string]int{
				"copper-plate": 1,
			},
			expectedProd: map[string]int{
				"copper-cable": 2,
			},
		},

		{
			item:   "rocket-control-unit",
			amount: 10,
			expectedIng: map[string]int{
				"speed-module":    10,
				"processing-unit": 10,
			},
			expectedProd: map[string]int{
				"rocket-control-unit": 10,
			},
		},
		{
			item:   "utility-science-pack",
			amount: 12,
			expectedIng: map[string]int{
				"low-density-structure": 12,
				"processing-unit":       8,
				"flying-robot-frame":    4,
			},
			expectedProd: map[string]int{
				"utility-science-pack": 12,
			},
		},
		{
			item:   "this-item-does-not-exist",
			amount: 1,
		},
	}

	for _, test := range tests {
		actualIng, actualProd := RecipeCost(test.item, test.amount)
		if !maps.Equal(actualIng, test.expectedIng) {
			t.Errorf("wrong ingredients for test '%d %s': wanted %v but got %v", test.amount, test.item, test.expectedIng, actualIng)
		}
		if !maps.Equal(actualProd, test.expectedProd) {
			t.Errorf("wrong products for test '%d %s': wanted %v but got %v", test.amount, test.item, test.expectedProd, actualProd)
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
				"copper-cable":       1,
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
	}
	for _, test := range tests {
		actualIng, actualProd := RecipeFullCost(test.item, test.amount)
		if !maps.Equal(actualIng, test.expectedIng) {
			t.Errorf("wrong ingredients for test '%d %s': wanted %v but got %v", test.amount, test.item, test.expectedIng, actualIng)
		}
		if !maps.Equal(actualProd, test.expectedProd) {
			t.Errorf("wrong products for test '%d %s': wanted %v but got %v", test.amount, test.item, test.expectedProd, actualProd)
		}
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
			item:   "logistic-science-pack",
			amount: 20,
			result: data.Ingredients{
				{
					Name:   "iron-ore",
					Amount: 110,
				},
				{
					Name:   "iron-plate",
					Amount: 110,
				},
				{
					Name:   "copper-ore",
					Amount: 30,
				},
				{
					Name:   "copper-plate",
					Amount: 30,
				},
				{
					Name:   "copper-cable",
					Amount: 60,
				},
				{
					Name:   "electronic-circuit",
					Amount: 20,
				},
				{
					Name:   "iron-gear-wheel",
					Amount: 30,
				},
				{
					Name:   "inserter",
					Amount: 20,
				},
				{
					Name:   "transport-belt",
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
					Name:   "iron-ore",
					Amount: 782,
				},
				{
					Name:   "iron-plate",
					Amount: 782,
				},
				{
					Name:   "copper-ore",
					Amount: 1450,
				},
				{
					Name:   "copper-plate",
					Amount: 1450,
				},
				{
					Name:   "copper-cable",
					Amount: 2900,
				},
				{
					Name:   "electronic-circuit",
					Amount: 780,
				},
				{
					Name:   "petroleum-gas",
					Amount: 2950,
				},
				{
					Name:   constants.PreferredFuel,
					Amount: 140,
				},
				{
					Name:   "plastic-bar",
					Amount: 280,
				},
				{
					Name:   "advanced-circuit",
					Amount: 140,
				},
				{
					Name:   "water",
					Amount: 350,
				},
				{
					Name:   "sulfur",
					Amount: 10,
				},
				{
					Name:   "sulfuric-acid",
					Amount: 100,
				},
				{
					Name:   "processing-unit",
					Amount: 20,
				},
				{
					Name:   "speed-module",
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
		actual, err := RecipeAllIngredients(test.item, test.amount)
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
		inv, err := Handcraft(test.inventory, test.item, test.amount)
		if !cmpErr(test.expectedErr, err) {
			t.Errorf("wrong error. Wanted %v but got %v", test.expectedErr, err)
			continue
		}

		for i, n := range inv {
			if n != test.expectedInv[i] {
				t.Errorf("wrong number for inventory %q. Wanted %d but got %d", i, test.expectedInv[i], n)
			}
		}
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
