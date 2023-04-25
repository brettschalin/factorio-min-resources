local loc = require("locations")
local math2d = require("math2d")

local finished_research = {}

script.on_event(defines.events.on_research_finished, function(event) 
    finished_research[event.research.name] = true
end)

-- the rest of the file is this function, which should be called from control.lua
function populate_tasks(queues)


-- task done() checks. All should define or return a function that accepts a `game.player` object
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
                inventory = b.get_inventory(defines.inventory[inv])
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

    -- ores are items, but they mess up the `mined` function
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
            done = has_inventory("player", args.entity, args.amount)
        else
            done = has_inventory("player", args.resource, args.amount)
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


-- Task definitions. This is an example that mines, crafts and builds the first basic research setup

add_task("speed", nil, {n = 10})
-- Mine and smelt
task_0 = add_task("mine", nil, {resource = "copper-ore", amount = 19})
task_1 = add_task("mine", nil, {resource = "coal", amount = 20})
task_2 = add_task("build", nil, {entity = "stone-furnace"})
task_3 = add_task("put", {task_0, task_1, task_2}, {inventory = defines.inventory.furnace_source, entity = "stone-furnace", item = "copper-ore", amount = 19})
task_4 = add_task("put", {task_3},{inventory = defines.inventory.fuel, entity = "stone-furnace", item = "coal", amount = 20, sub = 1})
task_5 = add_task("mine", nil, {resource = "iron-ore", amount = 26})
task_6 = add_task("wait", {task_4, task_5}, {done = has_inventory("stone-furnace", "copper-plate", 19)})
task_7 = add_task("take", {task_6}, {inventory = defines.inventory.furnace_result, entity = "stone-furnace", item = "copper-plate", amount = 19})

task_8 = add_task("put", {task_7}, {inventory = defines.inventory.furnace_source, entity = "stone-furnace", item = "iron-ore", amount = 26, sub = 1})
task_9 = add_task("mine", {task_8}, {resource = "iron-ore", amount = 50})
task_10 = add_task("wait", {task_8}, {done = has_inventory("stone-furnace", "iron-plate", 26)})
task_11 = add_task("take", {task_10}, {inventory = defines.inventory.furnace_result, entity = "stone-furnace", item = "iron-plate", amount = 26})
task_12 = add_task("put", {task_11}, {inventory = defines.inventory.furnace_source, entity = "stone-furnace", item = "iron-ore", amount = 50, sub = 1})
task_13 = add_task("mine", {task_12}, {resource = "stone", amount = 5})
task_14 = add_task("wait", {task_12}, {done = has_inventory("stone-furnace", "iron-plate", 50)})
task_15 = add_task("take", {task_13}, {inventory = defines.inventory.furnace_result, entity = "stone-furnace", item = "iron-plate", amount = 50})

add_task("speed", nil, {n = 1})

-- craft everything
task_16 = add_task("craft", {task_15}, {item = "boiler", amount = 1})
task_17 = add_task("craft", {task_16}, {item = "small-electric-pole", amount = 1})
task_18 = add_task("craft", {task_17}, {item = "steam-engine", amount = 1})
task_19 = add_task("craft", {task_18}, {item = "offshore-pump", amount = 1})
task_20 = add_task("craft", {task_19}, {item = "lab", amount = 1})

-- mine the crash site while we wait
task_21 = add_task("mine", {task_15}, {location = {x = -5, y = -6}})
task_22 = add_task("mine", nil, {location = {x = -17.5, y = -3.5}})
task_23 = add_task("mine", nil, {location = {x = -18.8, y = -0.3}})
task_24 = add_task("mine", nil, {location = {x = -27.5, y = -3.8}})
task_25 = add_task("mine", nil, {location = {x = -28, y = 1.9}})
task_26 = add_task("mine", nil, {location = {x = -37.8, y = 1.5}})

-- build the setup
task_27 = add_task("build", {task_18, task_26}, {entity = "offshore-pump", direction = defines.direction.north})
task_28 = add_task("build", {task_15, task_27}, {entity = "boiler", direction = defines.direction.east})
task_29 = add_task("build", {task_18, task_28}, {entity = "steam-engine", direction = defines.direction.east})
task_30 = add_task("build", {task_17, task_29}, {entity = "small-electric-pole"})
task_31 = add_task("build", {task_20, task_30}, {entity = "lab"})


end -- populate_tasks

return populate_tasks