package calc

import (
	"errors"
	"fmt"
	"math"

	"github.com/brettschalin/factorio-min-resources/data"
)

// Vanilla base items. These are items that are either mined or deemed too complicated to figure out automatically
var BaseItems = map[string]bool{
	// these are extracted from the world, and in all likelihood do not have a recipe that crafts them
	"water":       true,
	"crude-oil":   true,
	"stone":       true,
	"iron-ore":    true,
	"copper-ore":  true,
	"coal":        true,
	"uranium-ore": true,
	"wood":        true,

	// these can be crafted, but either there's multiple recipes (which makes it non-deterministic which is chosen)
	// or the recipes are cyclic (which makes them very annoying), or both. Best to leave them alone for now
	"heavy-oil":     true,
	"light-oil":     true,
	"petroleum-gas": true,
	"solid-fuel":    true,
	"uranium-235":   true,
	"uranium-238":   true,
}

// RecipeCost returns the amount of ingredients required to craft at least `amount` many `item`s
// and the products created.
// `nil` is returned if no recipe crafts the chosen items
func RecipeCost(item string, amount int) (ingredients, products map[string]int) {
	recipe := data.D.GetRecipe(item)

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
	rec := data.D.GetRecipe(item)
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
func Handcraft(inventory map[string]int, item string, amount int) (newInventory map[string]int, err error) {

	rec := data.D.GetRecipe(item)
	if rec == nil || !rec.CanHandcraft() {
		return nil, ErrCantHandcraft
	}

	newInventory = make(map[string]int)
	for k, v := range inventory {
		newInventory[k] = v
	}

	ingredients, products := RecipeCost(item, amount)

	for ing, n := range ingredients {
		diff := n - newInventory[ing]
		if diff > 0 {
			if r := data.D.GetRecipe(ing); !r.CanHandcraft() {
				return nil, &ErrMissingIngredient{ing, diff}
			}
			// not enough in inventory. Try to craft it
			newInventory, err = Handcraft(newInventory, ing, diff)
			if err != nil {
				return nil, err
			}
		}

		// now we definitely have enough. Take it out of the inventory
		newInventory[ing] -= n
	}

	for p, n := range products {
		newInventory[p] += n
	}

	return newInventory, nil
}

// TechCost returns the number of science packs required to research this technology
func TechCost(name string) map[string]int {

	t := data.D.GetTech(name)
	if t == nil {
		return nil
	}
	cost := map[string]int{}
	for _, c := range t.Unit.Ingredients {
		cost[c.Name] += t.Unit.Count * c.Amount
	}

	return cost
}

// TechFullCost returns the number of science packs required to research this technology and
// all unresearched prerequisites
func TechFullCost(researched map[string]bool, name string) map[string]int {
	res := make(map[string]bool)
	for k, v := range researched {
		res[k] = v
	}
	return techFullCost(res, name)
}

func techFullCost(researched map[string]bool, name string) map[string]int {

	if researched[name] {
		return nil
	}

	cost := TechCost(name)
	if cost == nil {
		return nil
	}
	tech := data.D.GetTech(name)
	if tech == nil {
		return nil
	}
	for _, p := range tech.Prerequisites {
		for pack, amount := range techFullCost(researched, p) {
			cost[pack] += amount
			researched[p] = true
		}
	}
	return cost
}
