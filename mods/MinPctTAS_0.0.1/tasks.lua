-- Automatically generated. DO NOT EDIT this section
local loc = require("locations")
local math2d = require("math2d")

local finished_research = {}

script.on_event(defines.events.on_research_finished, function(event) 
    finished_research[event.research.name] = true
end)

-- the rest of the file is this function, which should be called from control.lua
function populate_tasks(queues)


-- task done() checks. All should define or return a function that accepts a "game.player" object
-- and returns a boolean


-- Mining / inventory transfers. Checks that some amount of
-- item is in the machine
local function has_inventory(inv, item, amount, exact, slot)
    return function(p)
        if type(inv) == "function" then
            local cnt = inv().get_item_count(item)
            if exact then
                return cnt == amount
            else
                return cnt >= amount
            end
        elseif inv == "player" then
            -- special case: player inventory
            local cnt = p.get_inventory(defines.inventory.character_main).get_item_count(item)
            if exact then
                return cnt == amount
            else
                return cnt >= amount
            end
        else
            -- it's the name of a building. This will throw an error if the building isn't placed,
            -- which is a problem with the task dependencies rather than this code
            b = loc.buildings.get(p, inv).entity

            if slot then
                slots = {
                    slot
                }
            else
                slots = {
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
                    -- "item_main", -- what is this?
                    "rocket_silo_input",
                    "rocket_silo_output", -- is this different from result?
                    "rocket_silo_modules",
                    "rocket_silo_result" -- needed?
                }
            end
            -- Search for an inventory that might have the item
            for k, inv in pairs(slots) do
                inventory = b.get_inventory(inv)
                local cmp = false
                if inventory ~= nil then
                    local cnt = inventory.get_item_count(item) 
                    if exact then
                        cmp = cnt == amount
                    else
                        cmp = cnt >= amount
                    end
                end
                if cmp then
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
local function handcrafting_done(p)
    return p.crafting_queue == nil or #p.crafting_queue == 0
end

-- wait for a set number of ticks
local function idle(n)
    return function (p)
        if n > 0 then 
            n = n - 1
        end
        return n == 0
    end
end

-- finds the absolute value of the input number
local function abs(n)
    if n < 0 then
        return -n
    end
    return n
end

-- are we close enough to the target?
local function walking_done(pos)
    return function (p)
        dx = abs(p.position.x - pos.x)
        dy = abs(p.position.y - pos.y)
        return dx < 0.2 and dy < 0.2
    end
end

local mining_done = false

-- for location specific mining. This is mostly a special case for mining the
-- spaceship crash site
script.on_event(defines.events.on_player_mined_item, function(event)

    -- ores are items, but they mess up the "mined" function
    -- if they're allowed to do anything here
    if event.item_stack.name == "iron-ore" or
       event.item_stack.name == "copper-ore" or
       event.item_stack.name == "coal" or
       event.item_stack.name == "wood" or
       event.item_stack.name == "stone" then
        return
    end

	mining_done = true
end)

local function mined(p)
    if mining_done then
        mining_done = false
        return true
    end
    return false
end

-- add a task. This places it into the appropriate queues and sets the done() function
function add_task(task, prereqs, args)
    if task == "walk" then
        q = "character_action"
        done = walking_done(args.location)
    elseif task == "put" then
        q = "character_action"
        done = idle(1)
    elseif task == "take" then
        q = "character_action"
        done = idle(1)
    elseif task == "recipe" then
        q = "character_action"
        done = idle(1)
    elseif task == "speed" then
        q = "character_action" 
        done = idle(1)
    elseif task == "idle" then
        q = "character_action"
        done = idle(args.n)
    elseif task == "mine" then
        q = "character_action"
        if args.location ~= nil then
            done = mined
        elseif args.entity ~= nil then
            -- TODO: replace this with mined?
            done = has_inventory("player", args.entity, args.amount or 1)
        else
            done = has_inventory("player", args.resource, args.amount or 1)
        end
    elseif task == "build" then
        q = "character_action"
        done = is_built(args.entity)
    elseif task == "wait" then
        q = "character_action"
    elseif task == "craft" then
        q = "character_craft"
        done = handcrafting_done
    elseif task == "tech" then
        q = "lab"
        done = research_done(args.tech)
    else
        error("unknown task "..task)
    end

    -- override the default function if a custom one is provided
    if args.done then
        done = args.done
    end

    return queues.push(q, {
        task = task,
        prereqs = prereqs,
        args = args,
        done = done
    })
end

-- Task definitions. Feel free to modify anything below here

-- mine the crash site. The Go code assumes we have the 8 iron-plates
-- you get from this, so be careful about editing it. These locations are the few
-- that are map dependent as the placement of the wreckage is somewhat random; a future improvement
-- will hopefully change that
add_task("mine", nil, {location = {x = -5, y = -6}})
add_task("mine", nil, {location = {x = -17.5, y = -3.5}})
add_task("mine", nil, {location = {x = -18.8, y = -0.3}})
add_task("mine", nil, {location = {x = -27.5, y = -3.8}})
add_task("mine", nil, {location = {x = -28, y = 1.9}})
add_task("mine", nil, {location = {x = -37.8, y = 1.5}})

speed_0 = add_task("speed", nil, {n = 100.00})
tech_0 = add_task("tech", nil, {tech = "steel-processing"})
tech_1 = add_task("tech", nil, {tech = "logistic-science-pack"})
tech_2 = add_task("tech", nil, {tech = "automation"})
tech_3 = add_task("tech", nil, {tech = "electronics"})
tech_4 = add_task("tech", nil, {tech = "optics"})
tech_5 = add_task("tech", nil, {tech = "solar-energy"})
tech_6 = add_task("tech", nil, {tech = "advanced-material-processing"})
build_0 = add_task("build", nil, {entity = "stone-furnace", direction = defines.direction.north})
mine_0 = add_task("mine", nil, {resource = "coal", amount = 7})
put_0 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.fuel, item = "coal", amount = 7})
mine_1 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
put_1 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_2 = add_task("mine", nil, {resource = "iron-ore", amount = 18})
mine_3 = add_task("mine", nil, {resource = "copper-ore", amount = 19})
wait_0 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_0 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_2 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 18})
wait_1 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 18, true, defines.inventory.furnace_result)})
take_1 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 18})
mine_4 = add_task("mine", nil, {resource = "stone", amount = 5})
put_3 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 19})
wait_2 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 19, true, defines.inventory.furnace_result)})
take_2 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 19})
craft_0 = add_task("craft", {mine_4}, {item = "steam-engine", amount = 1})
craft_1 = add_task("craft", {take_2}, {item = "offshore-pump", amount = 1})
craft_2 = add_task("craft", nil, {item = "lab", amount = 1})
craft_3 = add_task("craft", nil, {item = "small-electric-pole", amount = 1})
craft_4 = add_task("craft", nil, {item = "boiler", amount = 1})
build_1 = add_task("build", {craft_0}, {entity = "steam-engine", direction = defines.direction.east})
build_2 = add_task("build", {craft_1}, {entity = "offshore-pump", direction = defines.direction.north})
build_3 = add_task("build", {craft_2}, {entity = "lab", direction = defines.direction.north})
build_4 = add_task("build", {craft_3}, {entity = "small-electric-pole", direction = defines.direction.north})
build_5 = add_task("build", {craft_4}, {entity = "boiler", direction = defines.direction.east})
mine_5 = add_task("mine", nil, {resource = "coal", amount = 4})
put_4 = add_task("put", nil, {entity = "boiler", inventory = defines.inventory.fuel, item = "coal", amount = 4})
mine_6 = add_task("mine", nil, {resource = "coal", amount = 11})
put_5 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.fuel, item = "coal", amount = 11})
mine_7 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
put_6 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_8 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_3 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_3 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_7 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
wait_4 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_4 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
mine_9 = add_task("mine", nil, {resource = "copper-ore", amount = 50})
put_8 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 50})
wait_5 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 50, true, defines.inventory.furnace_result)})
take_5 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 50})
craft_5 = add_task("craft", {take_5}, {item = "automation-science-pack", amount = 50})
put_9 = add_task("put", {craft_5}, {entity = "lab", inventory = defines.inventory.lab_input, item = "automation-science-pack", amount = 50})
mine_10 = add_task("mine", nil, {resource = "coal", amount = 6})
put_10 = add_task("put", nil, {entity = "boiler", inventory = defines.inventory.fuel, item = "coal", amount = 6})
mine_11 = add_task("mine", nil, {resource = "coal", amount = 17})
put_11 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.fuel, item = "coal", amount = 17})
mine_12 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
put_12 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_13 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_6 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_6 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_13 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_14 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_7 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_7 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_14 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
wait_8 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_8 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
mine_15 = add_task("mine", nil, {resource = "copper-ore", amount = 50})
put_15 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 50})
mine_16 = add_task("mine", nil, {resource = "copper-ore", amount = 25})
wait_9 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 50, true, defines.inventory.furnace_result)})
take_9 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 50})
put_16 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 25})
wait_10 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 25, true, defines.inventory.furnace_result)})
take_10 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 25})
craft_6 = add_task("craft", {take_10}, {item = "automation-science-pack", amount = 75})
put_17 = add_task("put", {craft_6}, {entity = "lab", inventory = defines.inventory.lab_input, item = "automation-science-pack", amount = 75})
mine_17 = add_task("mine", nil, {resource = "coal", amount = 49})
put_18 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.fuel, item = "coal", amount = 49})
mine_18 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
put_19 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_19 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_11 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_11 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_20 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_20 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_12 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_12 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_21 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_21 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_13 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_13 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_22 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_22 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_14 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_14 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_23 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_23 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_15 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_15 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_24 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_24 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_16 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_16 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_25 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_25 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_17 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_17 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_26 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_26 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_18 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_18 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_27 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_27 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_19 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_19 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_28 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_28 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_20 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_20 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_29 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_29 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_21 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_21 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_30 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_30 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_22 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_22 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_31 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_31 = add_task("mine", nil, {resource = "iron-ore", amount = 40})
wait_23 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_23 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_32 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 40})
wait_24 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 40, true, defines.inventory.furnace_result)})
take_24 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 40})
mine_32 = add_task("mine", nil, {resource = "coal", amount = 49})
put_33 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.fuel, item = "coal", amount = 49})
mine_33 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
put_34 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_34 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_25 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_25 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_35 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_35 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_26 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_26 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_36 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_36 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_27 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_27 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_37 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_37 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_28 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_28 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_38 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_38 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_29 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_29 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_39 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_39 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_30 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_30 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_40 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_40 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_31 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_31 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_41 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_41 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_32 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_32 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_42 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_42 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_33 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_33 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_43 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_43 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_34 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_34 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_44 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_44 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_35 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_35 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_45 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_45 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_36 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_36 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_46 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_46 = add_task("mine", nil, {resource = "iron-ore", amount = 40})
wait_37 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_37 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_47 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 40})
wait_38 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 40, true, defines.inventory.furnace_result)})
take_38 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 40})
mine_47 = add_task("mine", nil, {resource = "coal", amount = 43})
put_48 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.fuel, item = "coal", amount = 43})
mine_48 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
put_49 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_49 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_39 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_39 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_50 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_50 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_40 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_40 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_51 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_51 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_41 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_41 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_52 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_52 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_42 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_42 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_53 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_53 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_43 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_43 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_54 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_54 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_44 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_44 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_55 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_55 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_45 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_45 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_56 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_56 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_46 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_46 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_57 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_57 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_47 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_47 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_58 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_58 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_48 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_48 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_59 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_59 = add_task("mine", nil, {resource = "iron-ore", amount = 45})
wait_49 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_49 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_60 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 45})
wait_50 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 45, true, defines.inventory.furnace_result)})
take_50 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 45})
craft_7 = add_task("craft", {take_50}, {item = "iron-gear-wheel", amount = 675})
craft_8 = add_task("craft", nil, {item = "transport-belt", amount = 125})
mine_60 = add_task("mine", nil, {resource = "coal", amount = 49})
put_61 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.fuel, item = "coal", amount = 49})
mine_61 = add_task("mine", nil, {resource = "copper-ore", amount = 50})
put_62 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 50})
mine_62 = add_task("mine", nil, {resource = "copper-ore", amount = 50})
wait_51 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 50, true, defines.inventory.furnace_result)})
take_51 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 50})
put_63 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 50})
mine_63 = add_task("mine", nil, {resource = "copper-ore", amount = 50})
wait_52 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 50, true, defines.inventory.furnace_result)})
take_52 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 50})
put_64 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 50})
mine_64 = add_task("mine", nil, {resource = "copper-ore", amount = 50})
wait_53 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 50, true, defines.inventory.furnace_result)})
take_53 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 50})
put_65 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 50})
mine_65 = add_task("mine", nil, {resource = "copper-ore", amount = 50})
wait_54 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 50, true, defines.inventory.furnace_result)})
take_54 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 50})
put_66 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 50})
mine_66 = add_task("mine", nil, {resource = "copper-ore", amount = 50})
wait_55 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 50, true, defines.inventory.furnace_result)})
take_55 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 50})
put_67 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 50})
mine_67 = add_task("mine", nil, {resource = "copper-ore", amount = 50})
wait_56 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 50, true, defines.inventory.furnace_result)})
take_56 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 50})
put_68 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 50})
mine_68 = add_task("mine", nil, {resource = "copper-ore", amount = 50})
wait_57 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 50, true, defines.inventory.furnace_result)})
take_57 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 50})
put_69 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 50})
mine_69 = add_task("mine", nil, {resource = "copper-ore", amount = 50})
wait_58 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 50, true, defines.inventory.furnace_result)})
take_58 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 50})
put_70 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 50})
mine_70 = add_task("mine", nil, {resource = "copper-ore", amount = 50})
wait_59 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 50, true, defines.inventory.furnace_result)})
take_59 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 50})
put_71 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 50})
mine_71 = add_task("mine", nil, {resource = "copper-ore", amount = 50})
wait_60 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 50, true, defines.inventory.furnace_result)})
take_60 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 50})
put_72 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 50})
mine_72 = add_task("mine", nil, {resource = "copper-ore", amount = 50})
wait_61 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 50, true, defines.inventory.furnace_result)})
take_61 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 50})
put_73 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 50})
mine_73 = add_task("mine", nil, {resource = "copper-ore", amount = 50})
wait_62 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 50, true, defines.inventory.furnace_result)})
take_62 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 50})
put_74 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 50})
mine_74 = add_task("mine", nil, {resource = "copper-ore", amount = 25})
wait_63 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 50, true, defines.inventory.furnace_result)})
take_63 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 50})
put_75 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 25})
wait_64 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 25, true, defines.inventory.furnace_result)})
take_64 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 25})
craft_9 = add_task("craft", {take_56}, {item = "automation-science-pack", amount = 300})
craft_10 = add_task("craft", {take_64}, {item = "logistic-science-pack", amount = 50})
craft_11 = add_task("craft", nil, {item = "logistic-science-pack", amount = 200})
mine_75 = add_task("mine", nil, {resource = "coal", amount = 36})
put_76 = add_task("put", nil, {entity = "boiler", inventory = defines.inventory.fuel, item = "coal", amount = 36})
wait_65 = add_task("wait", nil, {done = has_inventory("player", "automation-science-pack", 10, false, defines.inventory.character_main)})
put_77 = add_task("put", nil, {entity = "lab", inventory = defines.inventory.lab_input, item = "automation-science-pack", amount = 10})
wait_66 = add_task("wait", {tech_2}, {done = has_inventory("player", "automation-science-pack", 30, false, defines.inventory.character_main)})
put_78 = add_task("put", nil, {entity = "lab", inventory = defines.inventory.lab_input, item = "automation-science-pack", amount = 30})
wait_67 = add_task("wait", {tech_3}, {done = has_inventory("player", "automation-science-pack", 10, false, defines.inventory.character_main)})
put_79 = add_task("put", nil, {entity = "lab", inventory = defines.inventory.lab_input, item = "automation-science-pack", amount = 10})
wait_68 = add_task("wait", {tech_4}, {done = has_inventory("player", "logistic-science-pack", 50, false, defines.inventory.character_main)})
put_80 = add_task("put", nil, {entity = "lab", inventory = defines.inventory.lab_input, item = "automation-science-pack", amount = 50})
put_81 = add_task("put", nil, {entity = "lab", inventory = defines.inventory.lab_input, item = "logistic-science-pack", amount = 50})
mine_76 = add_task("mine", nil, {resource = "coal", amount = 37})
wait_69 = add_task("wait", nil, {done = has_inventory("boiler", "coal", 0, true, defines.inventory.fuel)})
put_82 = add_task("put", nil, {entity = "boiler", inventory = defines.inventory.fuel, item = "coal", amount = 37})
wait_70 = add_task("wait", nil, {done = has_inventory("lab", "automation-science-pack", 0, true, defines.inventory.lab_input)})
put_83 = add_task("put", nil, {entity = "lab", inventory = defines.inventory.lab_input, item = "automation-science-pack", amount = 50})
wait_71 = add_task("wait", nil, {done = has_inventory("player", "logistic-science-pack", 50, false, defines.inventory.character_main)})
put_84 = add_task("put", nil, {entity = "lab", inventory = defines.inventory.lab_input, item = "logistic-science-pack", amount = 50})
wait_72 = add_task("wait", nil, {done = has_inventory("lab", "automation-science-pack", 0, true, defines.inventory.lab_input)})
put_85 = add_task("put", nil, {entity = "lab", inventory = defines.inventory.lab_input, item = "automation-science-pack", amount = 150})
wait_73 = add_task("wait", nil, {done = has_inventory("lab", "logistic-science-pack", 0, true, defines.inventory.lab_input)})
put_86 = add_task("put", nil, {entity = "lab", inventory = defines.inventory.lab_input, item = "logistic-science-pack", amount = 150})
mine_77 = add_task("mine", nil, {resource = "coal", amount = 50})
wait_74 = add_task("wait", nil, {done = has_inventory("boiler", "coal", 0, true, defines.inventory.fuel)})
put_87 = add_task("put", nil, {entity = "boiler", inventory = defines.inventory.fuel, item = "coal", amount = 50})
mine_78 = add_task("mine", nil, {resource = "coal", amount = 6})
put_88 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.fuel, item = "coal", amount = 6})
mine_79 = add_task("mine", nil, {resource = "iron-ore", amount = 40})
put_89 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 40})
wait_75 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 40, true, defines.inventory.furnace_result)})
take_65 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 40})
mine_80 = add_task("mine", nil, {resource = "copper-ore", amount = 28})
put_90 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 28})
wait_76 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 28, true, defines.inventory.furnace_result)})
take_66 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 28})
put_91 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-plate", amount = 25})
wait_77 = add_task("wait", nil, {done = has_inventory("stone-furnace", "steel-plate", 5, true, defines.inventory.furnace_result)})
take_67 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "steel-plate", amount = 5})
craft_12 = add_task("craft", {tech_5}, {item = "solar-panel", amount = 1})
mine_81 = add_task("mine", {tech_5}, {entity = "steam-engine"})
mine_82 = add_task("mine", nil, {entity = "boiler"})
build_6 = add_task("build", {craft_12}, {entity = "solar-panel"})
mine_83 = add_task("mine", nil, {resource = "coal", amount = 41})
put_92 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.fuel, item = "coal", amount = 41})
mine_84 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
put_93 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_85 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_78 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_68 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_94 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_86 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_79 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_69 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_95 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_87 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_80 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_70 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_96 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_88 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_81 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_71 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_97 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_89 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_82 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_72 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_98 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_90 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_83 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_73 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_99 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_91 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_84 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_74 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_100 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_92 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_85 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_75 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_101 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_93 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_86 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_76 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_102 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_94 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
wait_87 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_77 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_103 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_95 = add_task("mine", nil, {resource = "iron-ore", amount = 14})
wait_88 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, true, defines.inventory.furnace_result)})
take_78 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_104 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 14})
wait_89 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 14, true, defines.inventory.furnace_result)})
take_79 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 14})
craft_13 = add_task("craft", {take_79}, {item = "transport-belt", amount = 38})
mine_96 = add_task("mine", nil, {resource = "coal", amount = 13})
put_105 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.fuel, item = "coal", amount = 13})
mine_97 = add_task("mine", nil, {resource = "copper-ore", amount = 50})
put_106 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 50})
mine_98 = add_task("mine", nil, {resource = "copper-ore", amount = 50})
wait_90 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 50, true, defines.inventory.furnace_result)})
take_80 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 50})
put_107 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 50})
mine_99 = add_task("mine", nil, {resource = "copper-ore", amount = 50})
wait_91 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 50, true, defines.inventory.furnace_result)})
take_81 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 50})
put_108 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 50})
mine_100 = add_task("mine", nil, {resource = "copper-ore", amount = 38})
wait_92 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 50, true, defines.inventory.furnace_result)})
take_82 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 50})
put_109 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 38})
wait_93 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 38, true, defines.inventory.furnace_result)})
take_83 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 38})
craft_14 = add_task("craft", {take_81}, {item = "automation-science-pack", amount = 75})
craft_15 = add_task("craft", {take_83}, {item = "inserter", amount = 75})
craft_16 = add_task("craft", {tech_1}, {item = "logistic-science-pack", amount = 75})
put_110 = add_task("put", {craft_14}, {entity = "lab", inventory = defines.inventory.lab_input, item = "automation-science-pack", amount = 75})
put_111 = add_task("put", {craft_16}, {entity = "lab", inventory = defines.inventory.lab_input, item = "logistic-science-pack", amount = 75})
mine_101 = add_task("mine", nil, {resource = "coal", amount = 5})
put_112 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.fuel, item = "coal", amount = 5})
mine_102 = add_task("mine", nil, {resource = "iron-ore", amount = 30})
put_113 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 30})
mine_103 = add_task("mine", nil, {resource = "stone", amount = 20})
wait_94 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 30, true, defines.inventory.furnace_result)})
take_84 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 30})
put_114 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "stone", amount = 20})
wait_95 = add_task("wait", nil, {done = has_inventory("stone-furnace", "stone-brick", 10, true, defines.inventory.furnace_result)})
take_85 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "stone-brick", amount = 10})
put_115 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-plate", amount = 30})
wait_96 = add_task("wait", nil, {done = has_inventory("stone-furnace", "steel-plate", 6, true, defines.inventory.furnace_result)})
take_86 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "steel-plate", amount = 6})
speed_1 = add_task("speed", nil, {n = 1.00})
craft_17 = add_task("craft", {tech_6}, {item = "steel-furnace", amount = 1})
mine_104 = add_task("mine", nil, {entity = "stone-furnace"})
build_7 = add_task("build", {craft_17, mine_104}, {entity = "steel-furnace", direction = defines.direction.north})

end -- populate_tasks

return populate_tasks
