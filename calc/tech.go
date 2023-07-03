package calc

import "github.com/brettschalin/factorio-min-resources/data"

// TechCost returns the number of science packs required to research this technology
func TechCost(name string) map[string]int {

	t := data.GetTech(name)
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
	tech := data.GetTech(name)
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
