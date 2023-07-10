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

tech_0 = add_task("tech", nil, {tech = "automation"})
build_0 = add_task("build", nil, {entity = "stone-furnace", direction = defines.direction.north})
mine_0 = add_task("mine", nil, {resource = "coal", amount = 7})
put_0 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.fuel, item = "coal", amount = 7})
mine_1 = add_task("mine", nil, {resource = "iron-ore", amount = 50})
put_1 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 50})
mine_2 = add_task("mine", nil, {resource = "iron-ore", amount = 18})
mine_3 = add_task("mine", nil, {resource = "copper-ore", amount = 19})
wait_0 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 50, false, defines.inventory.furnace_result)})
take_0 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 50})
put_2 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 18})
wait_1 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 18, false, defines.inventory.furnace_result)})
take_1 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 18})
mine_4 = add_task("mine", nil, {resource = "stone", amount = 5})
put_3 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 19})
wait_2 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 19, false, defines.inventory.furnace_result)})
take_2 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 19})
craft_0 = add_task("craft", {take_1}, {item = "iron-gear-wheel", amount = 21})
craft_1 = add_task("craft", {take_2}, {item = "copper-cable", amount = 19})
craft_2 = add_task("craft", nil, {item = "stone-furnace", amount = 1})
craft_3 = add_task("craft", nil, {item = "pipe", amount = 10})
craft_4 = add_task("craft", nil, {item = "electronic-circuit", amount = 12})
craft_5 = add_task("craft", nil, {item = "transport-belt", amount = 2})
craft_6 = add_task("craft", nil, {item = "steam-engine", amount = 1})
craft_7 = add_task("craft", nil, {item = "offshore-pump", amount = 1})
craft_8 = add_task("craft", nil, {item = "lab", amount = 1})
craft_9 = add_task("craft", nil, {item = "small-electric-pole", amount = 1})
craft_10 = add_task("craft", nil, {item = "boiler", amount = 1})
build_1 = add_task("build", {craft_6}, {entity = "steam-engine", direction = defines.direction.east})
build_2 = add_task("build", {craft_7}, {entity = "offshore-pump", direction = defines.direction.north})
build_3 = add_task("build", {craft_8}, {entity = "lab", direction = defines.direction.north})
build_4 = add_task("build", {craft_9}, {entity = "small-electric-pole", direction = defines.direction.north})
build_5 = add_task("build", {craft_10}, {entity = "boiler", direction = defines.direction.east})
mine_5 = add_task("mine", nil, {resource = "coal", amount = 15})
put_4 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.fuel, item = "coal", amount = 5})
mine_6 = add_task("mine", nil, {resource = "iron-ore", amount = 20})
put_5 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "iron-ore", amount = 20})
mine_7 = add_task("mine", nil, {resource = "copper-ore", amount = 10})
wait_3 = add_task("wait", nil, {done = has_inventory("stone-furnace", "iron-plate", 20, false, defines.inventory.furnace_result)})
take_3 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "iron-plate", amount = 20})
craft_11 = add_task("craft", {take_3}, {item = "iron-gear-wheel", amount = 10})
put_6 = add_task("put", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_source, item = "copper-ore", amount = 10})
wait_4 = add_task("wait", nil, {done = has_inventory("stone-furnace", "copper-plate", 10, false, defines.inventory.furnace_result)})
take_4 = add_task("take", nil, {entity = "stone-furnace", inventory = defines.inventory.furnace_result, item = "copper-plate", amount = 10})
craft_12 = add_task("craft", {take_4}, {item = "automation-science-pack", amount = 10})
put_7 = add_task("put", nil, {entity = "boiler", inventory = defines.inventory.fuel, item = "coal", amount = 10})
put_8 = add_task("put", {craft_12}, {entity = "lab", inventory = defines.inventory.lab_input, item = "automation-science-pack", amount = 10})

end -- populate_tasks

return populate_tasks
