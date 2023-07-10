package constants

type Inventory int

const (
	InventoryFuel Inventory = iota
	InventoryBurntResult
	InventoryChest
	InventoryFurnaceSource
	InventoryFurnaceResult
	InventoryFurnaceModules
	InventoryCharacterMain
	InventoryCharacterGuns
	InventoryCharacterAmmo
	InventoryCharacterArmor
	InventoryCharacterVehicle
	InventoryCharacterTrash
	InventoryGodMain
	InventoryEditorMain
	InventoryEditorGuns
	InventoryEditorAmmo
	InventoryEditorArmor
	InventoryRoboportRobot
	InventoryRoboportMaterial
	InventoryRobotCargo
	InventoryRobotRepair
	InventoryAssemblingMachineInput
	InventoryAssemblingMachineOutput
	InventoryAssemblingMachineModules
	InventoryLabInput
	InventoryLabModules
	InventoryItemMain
	InventoryRocketSiloRocket
	InventoryRocketSiloResult
	InventoryRocketSiloInput
	InventoryRocketSiloOutput
	InventoryRocketSiloModules
	InventoryRocket
	InventoryCarTrunk
	InventoryCarAmmo
	InventoryCargoWagon
	InventoryTurretAmmo
	InventoryBeaconModules
	InventoryCharacterCorpse
	InventoryArtilleryTurretAmmo
	InventoryArtilleryWagonAmmo
	InventorySpiderTrunk
	InventorySpiderAmmo
	InventorySpiderTrash
)

var inventoryNames = []string{
	"fuel",
	"burnt_result",
	"chest",
	"furnace_source",
	"furnace_result",
	"furnace_modules",
	"character_main",
	"character_guns",
	"character_ammo",
	"character_armor",
	"character_vehicle",
	"character_trash",
	"god_main",
	"editor_main",
	"editor_guns",
	"editor_ammo",
	"editor_armor",
	"roboport_robot",
	"roboport_material",
	"robot_cargo",
	"robot_repair",
	"assembling_machine_input",
	"assembling_machine_output",
	"assembling_machine_modules",
	"lab_input",
	"lab_modules",
	"mining_drill_modules",
	"item_main",
	"rocket_silo_rocket",
	"rocket_silo_result",
	"rocket_silo_input",
	"rocket_silo_output",
	"rocket_silo_modules",
	"rocket",
	"car_trunk",
	"car_ammo",
	"cargo_wagon",
	"turret_ammo",
	"beacon_modules",
	"character_corpse",
	"artillery_turret_ammo",
	"artillery_wagon_ammo",
	"spider_trunk",
	"spider_ammo",
	"spider_trash",
}

func (i Inventory) String() string {
	if i >= 0 && int(i) < len(inventoryNames) {
		return "defines.inventory." + inventoryNames[i]
	}
	return ""
}

type Direction int

const (
	DirectionNone Direction = iota
	DirectionNorth
	DirectionSouth
	DirectionEast
	DirectionWest
)

var directionNames = map[Direction]string{
	DirectionNone:  "",
	DirectionNorth: "defines.direction.north",
	DirectionSouth: "defines.direction.south",
	DirectionEast:  "defines.direction.east",
	DirectionWest:  "defines.direction.west",
}

func (d Direction) String() string {
	if d > 0 && int(d) < len(directionNames) {
		return directionNames[d]
	}
	return ""
}
