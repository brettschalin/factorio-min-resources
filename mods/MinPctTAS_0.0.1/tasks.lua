local loc = require("locations")

local finished_research = {}

local NORTH = defines.direction.north
local SOUTH = defines.direction.south
local EAST = defines.direction.east
local WEST = defines.direction.west

script.on_event(defines.events.on_research_finished, function(event) 
    finished_research[event.research.name] = true
end)

-- the rest of the file is this function, which should be called from control.lua
function populate_tasks(queues)

function add_task(task, prereqs, done, args)
    if task == "walk" or task == "put" or
        task == "take" or task == "recipe" or
        task == "speed" or task == "idle" or
        task == "mine" or task == "build" or
        task == "wait" then
        q = "character_action"
    elseif task == "craft" then
        q = "character_craft"
    elseif task == "tech" then
        q = "lab"
    else
        error("unknown task "..task)
    end

    return queues.push(q, {
        task = task,
        prereqs = prereqs,
        args = args,
        done = done
    })
end

-- task done() checks. All should define or return a function that accepts a `game.player` object
-- and returns a boolean


-- Mining / inventory transfers. Checks that some amount of
-- item is in the machine
local function has_inventory(inv, item, amount)
    return function(p)
        if type(inv) == "function" then
            return inv().get_item_count(item) >= amount
        elseif inv == "player" then
            -- special case: player inventory
            return p.get_inventory(defines.inventory.character_main).get_item_count(item) >= amount
        else
            -- it's the name of a building. This will throw an error if the building isn't placed,
            -- which is a problem with the task dependencies rather than this code
            b = loc.buildings.get(p, inv).entity

            -- Search for any inventory that might have the item
            for k, inv in pairs({
                "character_main",
                "fuel",
                "chest",
                "furnace_source",
                "furnace_result",
                "furnace_modules",
                "assembling_machine_input",
                "assembling_machine_output",
                "assembling_machine_modules",
                "lab_input",
                "lab_modules",
                "item_main", -- what is this?
                "rocket_silo_input",
                "rocket_silo_output", -- is this different from result?
                "rocket_silo_modules",
                "rocket_silo_result" -- needed?
                
            }) do
                inventory = b.get_inventory(defines.inventory[inv])
                if inventory ~= nil and inventory.get_item_count(item) >= amount then
                    return true
                end
            end
            return false
        end
    end
end

-- is the building placed on the map?
local function is_built(entity)
    return function(p)
        return loc.buildings.is_placed(p, entity)
    end
end


-- is this tech researched?
local function research_done(tech)
    return function(p)
        return finished_research[tech]
    end
end


-- is the handcrafting queue empty?
local function handcrafting_done(player)
    return player.crafting_queue == nil or #player.crafting_queue == 0
end


-- Anything we know is finished after one tick
local function noop(p)
    return true
end

-- Task definitions. This is an example that mines and crafts the first basic research setup


-- Mine and smelt
task_0 = add_task("mine", nil, has_inventory("player", "copper-ore", 19), {resource = "copper-ore", amount = 19})
task_1 = add_task("mine", nil, has_inventory("player", "coal", 20), {resource = "coal", amount = 20})
task_2 = add_task("build", nil, is_built("stone-furnace"), {entity = "stone-furnace"})
task_3 = add_task("put", {task_0, task_1, task_2}, has_inventory("stone-furnace", "copper-ore", 19), {inventory = defines.inventory.furnace_source, entity = "stone-furnace", item = "copper-ore", amount = 19})
-- coal is one less because smelting starts instantly and consumes one
task_4 = add_task("put", {task_3}, has_inventory("stone-furnace", "coal", 19), {inventory = defines.inventory.fuel, entity = "stone-furnace", item = "coal", amount = 20})
task_5 = add_task("mine", nil, has_inventory("player", "iron-ore", 26), {resource = "iron-ore", amount = 26})
task_6 = add_task("wait", {task_4, task_5}, has_inventory("stone-furnace", "copper-plate", 19))
task_7 = add_task("take", {task_6}, has_inventory("player", "copper-plate", 19), {inventory = defines.inventory.furnace_result, entity = "stone-furnace", item = "copper-plate", amount = 19})

task_8 = add_task("put", {task_7}, has_inventory("stone-furnace", "iron-ore", 25), {inventory = defines.inventory.furnace_source, entity = "stone-furnace", item = "iron-ore", amount = 26})
task_9 = add_task("mine", {task_8}, has_inventory("player", "iron-ore", 50), {resource = "iron-ore", amount = 50})
task_10 = add_task("wait", {task_8}, has_inventory("stone-furnace", "iron-plate", 26))
task_11 = add_task("take", {task_10}, has_inventory("player", "iron-plate", 26), {inventory = defines.inventory.furnace_result, entity = "stone-furnace", item = "iron-plate", amount = 26})
task_12 = add_task("put", {task_11}, has_inventory("stone-furnace", "iron-ore", 49), {inventory = defines.inventory.furnace_source, entity = "stone-furnace", item = "iron-ore", amount = 50})
task_13 = add_task("mine", {task_12}, has_inventory("player", "stone", 5), {resource = "stone", amount = 5})
task_14 = add_task("wait", {task_12}, has_inventory("stone-furnace", "iron-plate", 50))
task_15 = add_task("take", {task_13}, has_inventory("player", "iron-plate", 50), {inventory = defines.inventory.furnace_result, entity = "stone-furnace", item = "iron-plate", amount = 50})

task_16 = add_task("craft", {task_15}, handcrafting_done, {item = "boiler", amount = 1})
task_17 = add_task("craft", {task_16}, handcrafting_done, {item = "small-electric-pole", amount = 1})
task_18 = add_task("craft", {task_17}, handcrafting_done, {item = "steam-engine", amount = 1})
task_19 = add_task("craft", {task_18}, handcrafting_done, {item = "offshore-pump", amount = 1})
task_20 = add_task("craft", {task_19}, handcrafting_done, {item = "lab", amount = 1})

end -- populate_tasks

return populate_tasks