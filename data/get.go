// Automatically generated by go generate. DO NOT EDIT
package data

func GetRecipe(item string) *Recipe {
	return d.GetRecipe(item)
}

func GetTech(tech string) *Technology {
	return d.GetTech(tech)
}

func GetAssemblingMachine(name string) *AssemblingMachine {
	x := d.AssemblingMachine[name]
	return &x
}

func GetBoiler(name string) *Boiler {
	x := d.Boiler[name]
	return &x
}

func GetFurnace(name string) *Furnace {
	x := d.Furnace[name]
	return &x
}

func GetGenerator(name string) *Generator {
	x := d.Generator[name]
	return &x
}

func GetItem(name string) *Item {
	x := d.Item[name]
	return &x
}

func GetLab(name string) *Lab {
	x := d.Lab[name]
	return &x
}

func GetModule(name string) *Module {
	x := d.Module[name]
	return &x
}

func GetRocketSilo(name string) *RocketSilo {
	x := d.RocketSilo[name]
	return &x
}
