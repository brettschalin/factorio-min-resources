package calc

// These have a resource patch
var MinableResources = map[string]bool{
	"coal":        true,
	"copper-ore":  true,
	"iron-ore":    true,
	"stone":       true,
	"uranium-ore": true, // requires a drill and sulfuric acid
}

// These can be gathered with a pumpjack or offshore pump
var MineableFluids = map[string]bool{
	"crude-oil": true,
	"water":     true,
}

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
