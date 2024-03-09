package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/brettschalin/factorio-min-resources/constants"
	"github.com/brettschalin/factorio-min-resources/geo"
	"github.com/brettschalin/factorio-min-resources/shims/slices"
)

//go:generate ./gen_get.bash

const (
	recipeTypeSmelt = "smelting"
)

var (
	d Data
)

func Init(dataFile string) error {
	df, err := os.Open(dataFile)
	if err != nil {
		return err
	}
	defer df.Close()

	dec := json.NewDecoder(df)
	return dec.Decode(&d)
}

type Data struct {
	AssemblingMachine map[string]AssemblingMachine `json:"assembling-machine"`
	Boiler            map[string]Boiler            `json:"boiler"`
	Character         struct {
		Character `json:"character"`
	} `json:"character"`
	Furnace    map[string]Furnace    `json:"furnace"`
	Generator  map[string]Generator  `json:"generator"`
	Item       map[string]Item       `json:"item"`
	Lab        map[string]Lab        `json:"lab"`
	Module     map[string]Module     `json:"module"`
	Recipe     map[string]Recipe     `json:"recipe"`
	RocketSilo map[string]RocketSilo `json:"rocket-silo"`
	Technology map[string]Technology `json:"technology"`

	recipeCache map[string]*Recipe
	techCache   map[string]*Technology
}

func (d *Data) GetRecipe(item string) *Recipe {

	if d.recipeCache == nil {
		d.recipeCache = make(map[string]*Recipe)
	}

	if r, ok := d.recipeCache[item]; ok {
		return r
	}

	for s, r := range d.Recipe {
		// barreling recipes don't really produce anything and can lead to infinite loops very easily
		if strings.HasSuffix(r.Subgroup, "-barrel") {
			continue
		}
		if s == item {
			rec := r.Get()
			d.recipeCache[item] = rec
			return rec
		}

		rec := r.Get()

		if rec.Result == item {
			d.recipeCache[item] = rec
			return rec
		}

		for _, p := range rec.Results {
			if p.Name == item {
				d.recipeCache[item] = rec
				return rec
			}
		}
	}
	return nil
}

func GetSmeltingRecipe(ore string) *Recipe {
	cacheKey := "SMELT_" + ore
	if d.recipeCache == nil {
		d.recipeCache = make(map[string]*Recipe)
	}

	if r, ok := d.recipeCache[cacheKey]; ok {
		return r
	}

	for _, r := range d.Recipe {
		rec := r.Get()

		if rec.Category != recipeTypeSmelt {
			continue
		}

		if len(rec.Ingredients) > 1 {
			continue
		}

		if rec.Ingredients.Amount(ore) > 0 {
			d.recipeCache[cacheKey] = rec
			return rec
		}
	}
	return nil
}

func (d *Data) GetTech(tech string) *Technology {
	if d.techCache == nil {
		d.techCache = make(map[string]*Technology)
	}
	if cached, ok := d.techCache[tech]; ok {
		return cached
	}
	for n, t := range d.Technology {
		if n == tech {
			d.techCache[n] = &t
			return &t
		}
	}
	return nil
}

type AssemblingMachine struct {
	CollisionBox        geo.Rectangle       `json:"collision_box"`
	CraftingCategories  []string            `json:"crafting_categories"`
	CraftingSpeed       float64             `json:"crafting_speed"`
	EnergySource        EnergySource        `json:"energy_source"`
	EnergyUsage         EnergyString        `json:"energy_usage"`
	Minable             Minable             `json:"minable"`
	ModuleSpecification ModuleSpecification `json:"module_specification"`
	Name                string              `json:"name"`
	SelectionBox        geo.Rectangle       `json:"selection_box"`
}

type Boiler struct {
	BurningCooldown   int           `json:"burning_cooldown"`
	CollisionBox      geo.Rectangle `json:"collision_box"`
	EnergyConsumption EnergyString  `json:"energy_consumption"`
	EnergySource      EnergySource  `json:"energy_source"`
	Minable           Minable       `json:"minable"`
	Name              string        `json:"name"`
	SelectionBox      geo.Rectangle `json:"selection_box"`
}

type EnergySource struct {
	// ignored if type == "electric"
	Effectivity  int                    `json:"effectivity"`
	FuelCategory constants.FuelCategory `json:"fuel_category"`

	Type string `json:"type"`
}

type EnergyString float64

// UnmarshalJSON implements the json.Unmarshal interface. In the data JSON energy / power
// quantities are given as human readable strings like "1.8KW" or "1.21GJ"; this converts them to
// numeric values
func (e *EnergyString) UnmarshalJSON(b []byte) error {

	b = bytes.Trim(b, `"`)

	numIdx := bytes.IndexFunc(b, func(r rune) bool {
		return r > '9' || r < '0'
	})

	suffix := bytes.ToLower(b[numIdx:])
	n, err := strconv.ParseFloat(string(b[:numIdx]), 64)
	if err != nil {
		return err
	}
	switch string(suffix) {
	case "kw", "kj":
		n *= 1000
	case "mw", "mj":
		n *= 1000_000
	case "gw", "gj":
		n *= 1000_000_000
	case "tw", "tj":
		n *= 1000_000_000_000
	}

	*e = EnergyString(n)
	return nil
}

type Character struct {
	BuildDistance         float64  `json:"build_distance"`
	CraftingCategories    []string `json:"crafting_categories"`
	DropItemDistance      float64  `json:"drop_item_distance"`
	InventorySize         int      `json:"inventory_size"`
	MiningSpeed           float64  `json:"mining_speed"`
	Name                  string   `json:"name"`
	ReachDistance         float64  `json:"reach_distance"`
	ReachResourceDistance float64  `json:"reach_resource_distance"`
	RunningSpeed          float64  `json:"running_speed"`
}

type Furnace struct {
	AllowedEffects      []string            `json:"allowed_effects"`
	CollisionBox        geo.Rectangle       `json:"collision_box"`
	CraftingCategories  []string            `json:"crafting_categories"`
	CraftingSpeed       float64             `json:"crafting_speed"`
	EnergySource        EnergySource        `json:"energy_source"`
	EnergyUsage         EnergyString        `json:"energy_usage"`
	Minable             Minable             `json:"minable"`
	Name                string              `json:"name"`
	ModuleSpecification ModuleSpecification `json:"module_specification"`
	SelectionBox        geo.Rectangle       `json:"selection_box"`
}

type Generator struct {
	BurnsFluid   bool          `json:"burns_fluid"`
	CollisionBox geo.Rectangle `json:"collision_box"`
	Effectivity  float64       `json:"effectivity"`
	FluidUsage   float64       `json:"fluid_usage_per_tick"`
	MaxTemp      int           `json:"maximum_temperature"`
	Minable      Minable       `json:"minable"`
	Name         string        `json:"name"`
	SelectionBox geo.Rectangle `json:"selection_box"`
}

type Item struct {
	Name      string       `json:"name"`
	StackSize int          `json:"stack_size"`
	Subgroup  string       `json:"subgroup"`
	FuelValue EnergyString `json:"fuel_value"`
}

type Lab struct {
	CollisionBox        geo.Rectangle       `json:"collision_box"`
	EnergyUsage         EnergyString        `json:"energy_usage"`
	Inputs              []string            `json:"inputs"` // what science packs this accepts
	Minable             Minable             `json:"minable"`
	ModuleSpecification ModuleSpecification `json:"module_specification"`
	Name                string              `json:"name"`
	ResearchingSpeed    float64             `json:"researching_speed"`
	SelectionBox        geo.Rectangle       `json:"selection_box"`
}

// Minable represents the result of mining a building
type Minable struct {
	MiningTime float64 `json:"mining_time"`
	Result     string  `json:"result"`
}

var (
	modLimitCache = map[string]map[string]bool{}
)

type Module struct {
	Category   string       `json:"category"`
	Effect     ModuleEffect `json:"effect"`
	Name       string       `json:"name"`
	Tier       int          `json:"tier"`
	Limitation []string     `json:"limitation"` // what recipes this can be used on
}

func (m *Module) ProductivityBonus() float64 {
	return m.Effect.Productivity.Bonus
}

func (m *Module) AppliesTo(recipe string) bool {
	b, ok := modLimitCache[m.Name]
	if !ok {
		b = map[string]bool{}
		modLimitCache[m.Name] = b
	}

	if a, ok := b[recipe]; ok {
		return a
	}

	a := slices.Contains(m.Limitation, recipe)
	b[recipe] = a
	return a
}

type ModuleEffect struct {
	Consumption  ModuleEffectBonus `json:"consumption"`
	Pollution    ModuleEffectBonus `json:"pollution"`
	Productivity ModuleEffectBonus `json:"productivity"`
	Speed        ModuleEffectBonus `json:"speed"`
}

type ModuleEffectBonus struct {
	Bonus float64 `json:"bonus"`
}

type Recipe struct {

	// crafting category. If not given, or if it's in Character.CraftingCategories, it's handcraftable
	Category string `json:"category"`

	// crafting time, in seconds
	// if not present, assume "0.5"
	EnergyRequired float64 `json:"energy_required"`

	Name     string `json:"name"`
	Subgroup string `json:"subgroup"`

	Ingredients Ingredients `json:"ingredients"`
	Result      string      `json:"result"`
	Results     Ingredients `json:"results"`
	ResultCount int         `json:"result_count"`

	// Some recipes have expensive variants. If these are
	// not nil, the other fields won't be populated
	Expensive *Recipe `json:"expensive"`
	Normal    *Recipe `json:"normal"`
}

func (r *Recipe) ProductCount(item string) int {
	if r.Result == item {
		if r.ResultCount == 0 {
			return 1
		}
		return r.ResultCount
	}

	for _, res := range r.Results {
		if res.Name == item {
			return res.Amount
		}
	}
	return 0
}

func (r *Recipe) Get() *Recipe {
	if e := r.Expensive; constants.UseExpensive && e != nil {
		e.Category = r.Category
		e.Name = r.Name
		return e
	}
	if n := r.Normal; n != nil {
		n.Category = r.Category
		n.Name = r.Name
		return n
	}
	return r
}

func (r *Recipe) GetResults() Ingredients {
	if r.Result != "" {
		count := r.ResultCount
		if count == 0 {
			count = 1
		}
		return Ingredients{
			{
				Name:   r.Result,
				Amount: count,
			},
		}
	}
	out := make(Ingredients, len(r.Results))
	copy(out, r.Results)
	return out
}

// CraftingTime returns the crafting time in seconds
func (r *Recipe) CraftingTime() float64 {
	craftingTime := r.EnergyRequired

	if craftingTime == 0 {
		craftingTime = 0.5
	}

	return craftingTime
}

func (r *Recipe) CanHandcraft() bool {

	if r.Category == "" {
		return true
	}

	return slices.Contains(d.Character.CraftingCategories, r.Category)
}

type Ingredients []Ingredient

func (i Ingredients) Amount(item string) int {
	for _, ing := range i {
		if ing.Name == item {
			return ing.Amount
		}
	}
	return 0
}

func (i *Ingredients) MergeDuplicates() {
	if len(*i) <= 1 {
		return
	}

	newIng := Ingredients{}
	seen := map[string]bool{}

	for j, ing := range *i {
		if seen[ing.Name] {
			continue
		}
		seen[ing.Name] = true
		newIng = append(newIng, ing)

		for k := j + 1; k < len(*i); k++ {
			name := (*i)[k].Name
			if name == ing.Name {
				for l, subIng := range newIng {
					if subIng.Name == name {
						newIng[l].Amount += (*i)[k].Amount
						break
					}
				}

			}
		}
	}
	*i = newIng

}

type Ingredient struct {
	Name    string
	Amount  int
	IsFluid bool
}

// Factorio's raw data defines two different schemas for an Ingredient.
// This will take into account both and ensure JSON unmarshaling works
func (i *Ingredient) UnmarshalJSON(b []byte) error {

	var in []any
	err := json.Unmarshal(b, &in)
	if err == nil {
		e := fmt.Errorf(`malformed ingredient %v`, in)
		if len(in) != 2 {
			return e
		}
		name, ok := in[0].(string)
		if !ok {
			return e
		}
		amount, err := strconv.ParseInt(fmt.Sprint(in[1]), 10, 64)
		if err != nil {
			return e
		}
		*i = Ingredient{
			Name:   name,
			Amount: int(amount),
		}
		return nil
	}

	var i2 struct {
		Amount int    `json:"amount"`
		Name   string `json:"name"`
		Type   string `json:"type"`
	}
	err = json.Unmarshal(b, &i2)
	if err != nil {
		return err
	}

	*i = Ingredient{
		Name:    i2.Name,
		Amount:  i2.Amount,
		IsFluid: i2.Type == "fluid",
	}
	return nil
}

type ModuleSpecification struct {
	ModuleSlots int `json:"module_slots"`
}

type RocketSilo struct {
	AssemblingMachine         `json:",inline"`
	FixedRecipe               string `json:"fixed_recipe"`
	RocketPartsRequired       int    `json:"rocket_parts_required"`
	RocketResultInventorySize int    `json:"rocket_result_inventory_size"`
}

type Technology struct {
	Effects       []TechEffect `json:"effects"`
	Name          string       `json:"name"`
	Prerequisites []string     `json:"prerequisites"`
	Unit          TechCost     `json:"unit"`
}

type TechEffect struct {
	Recipe string `json:"recipe"`
	Type   string `json:"type"`

	// the data defines a lot more possibilities but so far
	// "unlock-recipe" is the only one that's relevant to our goal.
	// I might account for the rest later if the need arises
}

type TechCost struct {
	Count       int         `json:"count"`
	Ingredients Ingredients `json:"ingredients"`
	Time        int         `json:"time"`
}

func canCraft(r *Recipe, categories []string) bool {
	if r == nil {
		return false
	}
	for _, c := range categories {
		if c == r.Category {
			return true
		}
	}
	return false
}

func (a *AssemblingMachine) CanCraft(r *Recipe) bool {
	return canCraft(r, append(a.CraftingCategories, ""))
}

func (f *Furnace) CanCraft(r *Recipe) bool {
	return canCraft(r, f.CraftingCategories)
}

func (a *AssemblingMachine) IsBurner() bool {
	return a.EnergySource.FuelCategory == constants.FuelCategoryChemical
}

func (f *Furnace) IsBurner() bool {
	return f.EnergySource.FuelCategory == constants.FuelCategoryChemical
}
