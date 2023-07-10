package calc

import (
	"errors"
	"fmt"
	"math"

	"github.com/brettschalin/factorio-min-resources/data"
)

// RecipeCost returns the amount of ingredients required to craft at least `amount` many `item`s
// and the products created.
// `nil` is returned if no recipe crafts the chosen items
func RecipeCost(item string, amount int) (ingredients, products map[string]int) {
	recipe := data.GetRecipe(item)

	if recipe == nil {
		return nil, nil
	}

	ingredients = make(map[string]int)
	products = make(map[string]int)

	var (
		mul int
		a   = 1
	)
	if recipe.ResultCount != 0 {
		a = recipe.ResultCount
	}
	if len(recipe.Results) > 0 {
		a = recipe.Results.Amount(item)
	}

	// You cannot craft fractional recipes, so do some math to round up amount.
	// For example: copper-cables are produced in pairs, so if you want 3 you must craft 4
	mul = int(math.Ceil(float64(amount) / float64(a)))

	if recipe.Result != "" {
		for _, i := range recipe.Ingredients {
			ingredients[i.Name] = i.Amount * mul
		}
		products[recipe.Result] = a * mul
	} else {
		for _, i := range recipe.Ingredients {
			ingredients[i.Name] = i.Amount * mul
		}
		for _, i := range recipe.Results {
			products[i.Name] = i.Amount * mul
		}
	}

	return
}

// RecipeFullCost returns the amount of `baseItems` required to craft `amount` `item`s
// and the products created.
// Like RecipeCost, this calculates prerequisites, but unlike it, this performs the same algorithm recursively
// on the ingredients until `baseItems` are reached
// `nil`` is returned if no recipe crafts the given item
func RecipeFullCost(item string, amount int) (ingredients, products map[string]int) {
	ingredients, products, leftovers := recipeFullCost(item, amount)

	for name, n := range leftovers {
		// find cost to craft at most `n` `name`s
		// (ie, decrease provided `n` until the produced result is <= n
		// subtract ingredients from `ingredients`

		for i := 0; i < n; i++ {
			ing, prod, _ := recipeFullCost(name, n-i)
			if prod[name] <= n {
				for sName, sAmt := range ing {
					ingredients[sName] -= sAmt
				}
				delete(leftovers, name)
				break
			}
		}
	}

	for name, n := range leftovers {
		products[name] += n
	}

	return
}

func recipeFullCost(item string, amount int) (ingredients, products, leftovers map[string]int) {
	ingredients = make(map[string]int)
	products = make(map[string]int)
	leftovers = make(map[string]int)

	ing, prod := RecipeCost(item, amount)
	leftover := prod[item] - amount

	if leftover > 0 {
		leftovers[item] = leftover
	}

	if ing == nil {
		return nil, nil, nil
	}

	for i, a := range prod {
		products[i] = a
	}

	for n, i := range ing {
		if BaseItems[n] {
			ingredients[n] = i
			continue
		}
		newIng, _, newLeftovers := recipeFullCost(n, i)
		if newIng == nil {
			return nil, nil, nil
		}

		for name, amount := range newIng {
			ingredients[name] += amount
		}

		for name, amount := range newLeftovers {
			leftovers[name] += amount
		}
	}
	return
}

// RecipeAllIngredients returns the list of items that need to be created in order
// to craft the final item.
func RecipeAllIngredients(item string, amount int) (data.Ingredients, error) {
	return recipeAllIngredients(item, amount, 0)
}

func recipeAllIngredients(item string, amount int, depth int) (data.Ingredients, error) {
	rec := data.GetRecipe(item)
	if rec == nil {
		return nil, errors.New("no recipe crafts " + item)
	}

	out := data.Ingredients{}

	ingredients, _ := RecipeCost(item, amount)

	for _, ing := range rec.Ingredients {
		amt := ingredients[ing.Name]
		if !BaseItems[ing.Name] {
			subIng, err := recipeAllIngredients(ing.Name, amt, depth+1)
			if err != nil {
				return nil, err
			}
			out = append(out, subIng...)
		}
		out = append(out, data.Ingredient{Name: ing.Name, Amount: amt})
	}

	if depth == 0 {
		out = append(out, data.Ingredient{Name: item, Amount: amount})
	}
	out.MergeDuplicates()

	return out, nil
}

var (
	ErrCantHandcraft = errors.New("cannot handcraft this recipe")
)

type ErrMissingIngredient struct {
	item string
	n    int
}

func (e *ErrMissingIngredient) Error() string {
	return fmt.Sprintf(`missing %d %q(s)`, e.n, e.item)
}

// Handcraft performs a handcrafting action
func Handcraft(inventory map[string]uint, item string, amount uint) (newInventory map[string]uint, err error) {

	rec := data.GetRecipe(item)
	if rec == nil || !rec.CanHandcraft() {
		return nil, ErrCantHandcraft
	}

	newInventory = make(map[string]uint)
	for k, v := range inventory {
		newInventory[k] = v
	}

	ingredients, products := RecipeCost(item, int(amount))

	for ing, n := range ingredients {
		diff := n - int(newInventory[ing])
		if diff > 0 {
			if r := data.GetRecipe(ing); !r.CanHandcraft() {
				return nil, &ErrMissingIngredient{ing, diff}
			}
			// not enough in inventory. Try to craft it
			newInventory, err = Handcraft(newInventory, ing, uint(diff))
			if err != nil {
				return nil, err
			}
		}

		// now we definitely have enough. Take it out of the inventory
		newInventory[ing] -= uint(n)
	}

	for p, n := range products {
		newInventory[p] += uint(n)
	}

	return newInventory, nil
}

// OneStackRecipe returns the number of crafts of a recipe that can be done if given one
// stack of each input item, and at most produces one stack of output.
// Fluid inputs are skipped since they're not subject to the same stacking constraints
func OneStackRecipe(r *data.Recipe) int {
	count := math.MaxInt

	for _, ing := range r.Ingredients {
		if ing.IsFluid {
			continue
		}
		item := data.GetItem(ing.Name)
		c := int(math.Floor(float64(item.StackSize) / float64(ing.Amount)))
		if c < count {
			count = c
		}
	}

	item := data.GetItem(r.Name)
	c := int(math.Floor(float64(item.StackSize) / float64(r.ProductCount(r.Name))))
	if c < count {
		count = c
	}

	return count
}
