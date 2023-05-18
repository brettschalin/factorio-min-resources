package task

import (
	"fmt"
	"io"
	"strings"
)

// outputs the (Lua) tasks required to perform this action,
// and the ID of the last task to run
func (t *Task) Eval(lastID string) (string, string) {
	out := strings.Builder{}

	prereqIDs := make([]string, len(t.Prerequisites))

	var outLine string

	for i, p := range t.Prerequisites {
		prereqIDs[i] = p.GetID()
		outLine, lastID = p.Eval(lastID)
		out.WriteString(outLine)
		out.WriteByte('\n')
	}

	// meta tasks only appear for grouping purposes. Do not actually output them
	if t.Type == TaskMeta {
		return out.String(), lastID
	}

	var prereqs string

	if lastID != "" {
		prereqIDs = append(prereqIDs, lastID)
	}

	if len(prereqIDs) > 0 {
		prereqs = "{" + strings.Join(prereqIDs, ",") + "}"
	} else {
		prereqs = "nil"
	}

	var args string

	out.WriteString(t.GetID() + " = add_task(")
	typ := t.Type
	out.WriteString("\"" + typ.String() + "\"")
	switch typ {
	case TaskWalk:
		args = fmt.Sprintf("location = {x = %.2f, y = %.2f}", t.Location.X, t.Location.Y)

	case TaskWait:
		args = fmt.Sprintf(`done = %s(%q, %q, %d)`, t.WaitCondition, t.Entity, t.Item, t.Amount)

	case TaskCraft:
		args = fmt.Sprintf(`item = %q, amount = %d`, t.Item, t.Amount)

	case TaskBuild:
		if t.Direction != "" {
			args = fmt.Sprintf(`entity = %q, direction = %s`, t.Entity, t.Direction)
		} else {
			args = fmt.Sprintf(`entity = %q`, t.Entity)
		}
	case TaskTake, TaskPut:
		args = fmt.Sprintf(`entity = %q, inventory = %s, item = %q, amount = %d`,
			t.Entity, t.Slot, t.Item, t.Amount)

	case TaskTech:
		args = fmt.Sprintf(`tech = %q`, t.Tech)

	case TaskMine:
		if t.Location != nil {
			args = fmt.Sprintf(`location = {x = %.2f, y = %.2f`, t.Location.X, t.Location.Y)
		} else if t.Item != "" {
			args = fmt.Sprintf(`resource = %q, amount = %d`, t.Item, t.Amount)
		} else {
			args = fmt.Sprintf(`entity = %q`, t.Entity)
		}

	case TaskLaunch:
		args = `entity = "rocket-silo"`

	default:
		panic("unknown task type " + typ.String())
	}

	out.WriteString(", " + prereqs)
	out.WriteString(", {" + args + "}")

	out.WriteString(")\n")
	return out.String(), t.GetID()
}

func WriteTasksFile(w io.StringWriter, task *Task) error {

	var err error

	if _, err = w.WriteString(TasksLuaHeader); err != nil {
		return err
	}

	var out string

	out, _ = task.Eval("")
	if _, err = w.WriteString(out); err != nil {
		return err
	}

	_, err = w.WriteString(TasksLuaFooter)

	return err
}

const TasksLuaHeader = `-- Automatically generated. DO NOT EDIT this section
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
-- that are map dependentas the placement of the wreckage is somewhat random; a future improvement
-- will hopefully change that
add_task("mine", nil, {location = {x = -5, y = -6}})
add_task("mine", nil, {location = {x = -17.5, y = -3.5}})
add_task("mine", nil, {location = {x = -18.8, y = -0.3}})
add_task("mine", nil, {location = {x = -27.5, y = -3.8}})
add_task("mine", nil, {location = {x = -28, y = 1.9}})
add_task("mine", nil, {location = {x = -37.8, y = 1.5}})

`

const TasksLuaFooter = `
end -- populate_tasks

return populate_tasks
`
