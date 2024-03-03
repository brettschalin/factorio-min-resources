package calc

import (
	"errors"
	"fmt"
	"math"

	"github.com/brettschalin/factorio-min-resources/building"
	"github.com/brettschalin/factorio-min-resources/data"
	"github.com/brettschalin/factorio-min-resources/shims"
	"github.com/brettschalin/factorio-min-resources/shims/slices"
)

type Items[T shims.Ordered] map[string]T

func (i Items[T]) Merge(other Items[T]) {
	for k, v := range other {
		i[k] += v
	}
}

// RecipeCost returns the amount of ingredients required to craft the given recipe in the given machine
// and the products created.
// `nil` is returned if no recipe crafts the chosen items
func RecipeCost(recipe *data.Recipe, amount int, machine building.CraftingBuilding) (ingredients, products Items[int]) {

	if recipe == nil {
		return nil, nil
	}

	ingredients = make(Items[int])
	products = make(Items[int])

	var (
		bonus   = float64(1)
		recipes int
	)

	if machine != nil {
		bonus = 1 + machine.ProductivityBonus(recipe.Name)
	}

	recipes = int(math.Ceil(float64(amount) / bonus))

	for _, i := range recipe.Ingredients {
		ingredients[i.Name] = i.Amount * recipes
	}

	for _, p := range recipe.GetResults() {
		products[p.Name] = p.Amount * amount
	}

	return
}

// RecipeFullCost returns the amount of `baseItems` required to craft `amount` `item`s
// and the products created.
// Like RecipeCost, this calculates prerequisites, but unlike it, this performs the same algorithm recursively
// on the ingredients until `baseItems` are reached
// `nil“ is returned if no recipe crafts the given item
func RecipeFullCost(recipe *data.Recipe, amount int, machine building.CraftingBuilding) (ingredients, products Items[int]) {

	ingredients = make(Items[int])
	products = make(Items[int])

	ing, err := RecipeAllIngredients(recipe, amount, machine)

	if err != nil {
		return
	}

	for _, i := range ing {
		if BaseItems[i.Name] {
			ingredients[i.Name] += i.Amount
		}
	}

	for _, p := range recipe.GetResults() {
		products[p.Name] += p.Amount * amount
	}

	return
}

type recipeDependency struct {
	item    string
	recipe  *data.Recipe
	amount  int                 // items, not recipes
	deps    []*recipeDependency // what this recipe requires
	uses    []*recipeDependency // what requires this recipe
	visited bool
}

func (r *recipeDependency) reset() {
	r.visited = false
	for _, d := range r.deps {
		d.reset()
	}
}

func (r *recipeDependency) iter(f func(*recipeDependency)) {

	q := []*recipeDependency{r}

	for len(q) > 0 {
		rec := q[0]
		q = q[1:]
		if rec.visited {
			continue
		}
		ready := true
		for _, u := range rec.uses {
			if !u.visited {
				ready = false
				break
			}
		}
		if !ready {
			// we should be ready by the time the next loop comes around
			q = append(q, rec)
			continue
		}

		f(rec)
		for _, d := range rec.deps {
			if !d.visited {
				q = append(q, d)
			}
		}
		rec.visited = true
	}
	r.reset()
}

func buildRecDeps(item string, amount int) *recipeDependency {
	if item == "" {
		return nil
	}

	root := &recipeDependency{
		item:   item,
		amount: amount,
	}

	if BaseItems[item] {
		return root
	}

	recipe := data.GetRecipe(item)
	root.recipe = recipe

	deps := map[string]*recipeDependency{recipe.Name: root}

	q := []*recipeDependency{root}
	for len(q) > 0 {
		r := q[0]
		q = q[1:]
		if r.visited || r.recipe == nil {
			continue
		}
		r.visited = true

		if BaseItems[r.item] {
			continue
		}

		// build dependency graph
		for _, i := range r.recipe.Ingredients {
			rec, ok := deps[i.Name]
			if !ok {
				rec = &recipeDependency{
					item: i.Name,
				}

				if !BaseItems[i.Name] {
					rec.recipe = data.GetRecipe(i.Name)
				}
				deps[i.Name] = rec
			}
			rec.uses = append(rec.uses, r)
			r.deps = append(r.deps, rec)
			q = append(q, rec)
		}
	}

	// set amounts of each item
	root.reset()
	root.iter(func(r *recipeDependency) {
		if r.item == item {
			return
		}

		amt := 0

		for _, other := range r.uses {
			i := other.recipe.Ingredients.Amount(r.item)
			p := other.recipe.ProductCount(other.item)
			amt += int(math.Ceil(float64(other.amount) * float64(i) / float64(p)))
		}

		r.amount = amt
	})
	return root
}

// RecipeAllIngredients returns the list of items that need to be created in order
// to craft the final item.
func RecipeAllIngredients(recipe *data.Recipe, amount int, machine building.CraftingBuilding) (data.Ingredients, error) {
	if recipe == nil {
		return nil, errors.New("no recipe to craft")
	}

	out := data.Ingredients{}

	deps := buildRecDeps(recipe.Name, amount*recipe.ProductCount(recipe.Name))

	// TODO: deps.iter() to apply module bonuses

	deps.iter(func(r *recipeDependency) {
		out = append(out, data.Ingredient{
			Name:   r.item,
			Amount: r.amount,
		})
	})

	slices.Reverse(out)

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
func Handcraft(inventory Items[uint], recipe *data.Recipe, amount uint) (newInventory Items[uint], err error) {

	if recipe == nil || !recipe.CanHandcraft() {
		return nil, ErrCantHandcraft
	}

	newInventory = make(Items[uint])
	newInventory.Merge(inventory)

	ingredients, products := RecipeCost(recipe, int(math.Ceil(float64(amount)/float64(recipe.ProductCount(recipe.Name)))), nil)

	for ing, n := range ingredients {
		diff := n - int(newInventory[ing])
		if diff > 0 {
			if r := data.GetRecipe(ing); !r.CanHandcraft() {
				return nil, &ErrMissingIngredient{ing, diff}
			}
			// not enough in inventory. Try to craft it
			newInventory, err = Handcraft(newInventory, data.GetRecipe(ing), uint(diff))
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
