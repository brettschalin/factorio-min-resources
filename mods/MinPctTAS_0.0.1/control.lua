require ("util")
local math2d = require("math2d")
local populate_tasks = require("tasks")
local queues = require("queues")
local l_ = require("locations")
local buildings, locations, resources = l_.buildings, l_.locations, l_.resources
local idle = 0
local pick = 0
local dropping = 0

populate_tasks(queues)

-- Set this to display debugging messages
local dbg = 1

-- Display debugging messages, if dbg is on
local function debug(p, msg)
	if dbg > 0 then
		p.print(msg)
	end
end


-- This keeps track of what we've mined (see the event hook at the bottom of this file)
-- and will be printed when the TAS is done. A more robust option in the future is to
-- grab the production stats for these resources directly but that's not implemented yet
local resources_used = {
	["iron-ore"] = 0,
	["copper-ore"] = 0,
	coal = 0,
	stone = 0
}

-- Helper functions. All of them will, in order,
-- * attempt to select the entity at the given location, throw an error if it's not found
-- * check if the character can reach the location, return false if not
-- * attempt to perform the action. Failure is treated as an error
-- * return true
-- The return value tells the main task runner if the action succeeded or if it needs to
-- walk the player closer to the location and try again next tick


-- Create an entity on the surface. In most cases this is building a structure/item/entity
-- The direction doesn't always work as you'd expect for fluids.
--   asms       - once the recipe gets set, the fluid input will always be north, requiring rotation
--   chems      - direction indicates the side where the fluids are input
--   refineries - direction indicates the side where the fluids output
--   pumps      - direction indicates the side where the fluid is input
local function build(p, position, item, direction)
	-- Check if we have the item

	local count = p.get_item_count(item) 
	-- debug(p, string.format("(%d) found %d %s in player inventory", game.tick, count, item))
	if count == 0 then
		error(string.format("(%d) build: missing %s", game.tick, item))
		return false
	end

	-- check if we can reach the destination. Note that character_build_distance_bonus
	-- affects this calculation in a way I haven't figured out but since vanilla doesn't use
	-- it I don't see the need to do anything with it yet
	if math2d.position.distance(p.position, position) > p.build_distance then
		return false
	end


	-- Check if we can actually place the item at this tile
	local canplace = p.can_place_entity{name = item, position = position, direction = direction}
	if not canplace then
		error(string.format("(%d) build: cannot place %s at %d %d", game.tick, item, position.x, position.y))
	end

	-- place the item
	p.surface.create_entity{name = item, position = position, direction = direction, force = "player"}
	p.remove_item({name = item, count = 1})
	local b = buildings.build(p, item, position)
	debug(p, string.format("(%d) placed building %s", game.tick, serpent.block(b)))

	return true
end

-- Mine the resource or building at this location
local function mine(p, position)

	-- check if we can reach. I haven't tested this but whether this works with
	-- multiple overlapping entities is unknown and likely random since it'd depend
	-- on what order the game returns them in
	local entity = p.surface.find_entities_filtered({position = position})
	if not entity or not entity[1] or not p.can_reach_entity(entity[1]) then
		return false
	end
	
	-- start/continue the mining
	p.update_selected_entity(position)
	p.mining_state = {mining = true, position = position}
	-- debug(p, string.format("(%d) mining (%d, %d)", game.tick, position.x, position.y))
	return true
	
end

-- Handcraft one or more of a recipe
local function craft(p, count, recipe)
	amt = p.begin_crafting({recipe = recipe, count = count})
	if amt ~= count then
		error(string.format("craft: needed to start %d %s but could only start %d", count, recipe, amt))
		return false
	end
	return true
end

-- Manually launch the rocket
local function launch(p, position)
	p.update_selected_entity(position)
	if not p.selected then
		error(string.format("(%d) launch: no silo placed", game.tick))
		return false
	end
	-- Check if we are in reach of this tile
	if not p.can_reach_entity(p.selected) then
		return false
	end
	p.selected.launch_rocket()
	return true
end

-- Place an item from the character's inventory into an entity's inventory
local function put(p, position, item, amount, slot)
	p.update_selected_entity(position)

	if not p.selected then
		error(string.format("(%d) put: no entity at (%d, %d)", game.tick, position.x, position.y))
		return false
	end

 	if not p.can_reach_entity(p.selected) then
 		return false
 	end

	local amountininventory = p.get_item_count(item)
	if amountininventory < amount then
		error(string.format("(%d) put: not enough %s (wanted %d but have %d)", game.tick, item, amount, amountininventory))
		return false
	end

	local otherinv = p.selected.get_inventory(slot)

	if not otherinv then
		error(string.format("(%d) put: no inventory %s on %s", game.tick, slot, p.selected.name))
		return false
	end

	local inserted = otherinv.insert({name=item, count=amount})
	if inserted < amount then
		error(string.format("(%d) put: could not insert %s (wanted %d but %d succeeded)", game.tick, item, amount, inserted))
		return false
	end

	p.remove_item({name=item, count=amount})
	return true
end

-- Set the recipe of an assembling machine, chemical plant, or oil refinery (anything I'm missing?)
-- Items still in the machine not used in the new recipe will be placed in the character's inventory
-- NOTE: There is a bug here. It is possible to set a recipe that is not yet available through
-- completed research. For now, go on the honor system.
local function recipe(p, position, recipe)
	p.update_selected_entity(position)
	if not p.selected then
		error(string.format("(%d) recipe: no entity at (%d, %d)", game.tick, position.x, position.y))
		return false
	end
	-- Check if we are in reach of this tile
	if not p.can_reach_entity(p.selected) then
		return false
	end
	if recipe == "none" then
		recipe = nil
	end
	local items = p.selected.set_recipe(recipe)
	if items then
		for name, count in pairs(items) do
			p.insert{name=name, count=count}
		end
	end
	return true
end

-- Rotate an entity one quarter turn
local function rotate(p, position, direction)
	local opts = {reverse = false}
	p.update_selected_entity(position)
	if not p.selected then
		error(string.format("(%d) rotate: no entity at (%d, %d)", game.tick, position.x, position.y))
		return false
	end
	-- Check if we are in reach of this tile
	if not p.can_reach_entity(p.selected) then
	 	return false
	end
	if direction == "ccw" then
		opts = {reverse = true}
	end
	p.selected.rotate(opts)
	-- Not sure this is a good idea. Rotating a belt 180 requires two rotations. But
	-- rotating an underground belt 180 requires only one rotation. So maybe allowing 180
	-- will cause some headaches.
	if direction == "180" then
		p.selected.rotate(opts)
	end
	return true
end

-- Set the gameplay speed. 1 is standard speed
local function speed(speed)
	game.speed = speed
	return true
end

-- Take an item from the entity's inventory into the character's inventory
local function take(p, position, item, amount, slot)
	p.update_selected_entity(position)

	if not p.selected then
		error(string.format("(%d) take: no entity at (%d, %d)", game.tick, position.x, position.y))
		return false
	end

	-- Check if we are in reach of this tile
	if not p.can_reach_entity(p.selected) then
		return false
	end

	local otherinv = p.selected.get_inventory(slot)

	if not otherinv then
		error(string.format("(%d) take: no inventory %s on %s", game.tick, slot, p.selected.name))
		return false
	end


	local amountinmachine = otherinv.get_item_count(item)
	if amountinmachine < amount then
		error(string.format("(%d) take: not enough %s (wanted %d but have %d)", game.tick, item, amount, amountinmachine))
		return false
	end

	local taken = p.insert({name=item, count=amount})
	if taken < amount then
		error(string.format("(%d) take: could not take %s (wanted %d but %d succeeded)", game.tick, item, amount, taken))
		return false
	end

	otherinv.remove({name=item, count=amount})
	return true
end

-- Set the current research
local function tech(p, research)
	p.force.research_queue_enabled = true
	p.force.add_research(research)
	return true
end

-- Walks the character in the direction of a coordinate
local function walk(delta_x, delta_y)
	if delta_x > 0.2 then
		-- Easterly
		if delta_y > 0.2 then
			return {walking = true, direction = defines.direction.southeast}
		elseif delta_y < -0.2 then
			return {walking = true, direction = defines.direction.northeast}
		else
			return {walking = true, direction = defines.direction.east}
		end
	elseif delta_x < -0.2 then
		-- Westerly
		if delta_y > 0.2 then
			return {walking = true, direction = defines.direction.southwest}
		elseif delta_y < -0.2 then
			return {walking = true, direction = defines.direction.northwest}
		else
			return {walking = true, direction = defines.direction.west}
		end
	else
		-- Vertically
		if delta_y > 0.2 then
			return {walking = true, direction = defines.direction.south}
		elseif delta_y < -0.2 then
			return {walking = true, direction = defines.direction.north}
		else
			return {walking = false, direction = defines.direction.north}
		end
	end
end


-- End of helper functions. Below this is the main task runner


local first_tick = true
local done = false
local can_reach = false
local current_craft = queues.pop("character_craft")
local current_action = queues.pop("character_action")
local current_tech = queues.pop("lab")
local destination = {x = 0, y = 0} -- destination can't be nil, so make sure it has some value

local function prereqs_done(task)
	if task == nil or task.prereqs == nil or #task.prereqs == 0 then
		return true
	end
	for i, t in pairs(task.prereqs) do
		if not queues.is_done(t) then return false end
	end
	return true
end

-- Main per-tick event handler
script.on_event(defines.events.on_tick, function(event)
	if done then return end
	
	local p = game.players[1]
	local pos = p.position
	
	-- we're finished when the queues are empty and the last tasks are done
	if queues.is_empty("character_action") and current_action == nil and
	   queues.is_empty("character_craft") and current_craft == nil and
	   queues.is_empty("lab") and current_tech == nil then
		
		time_str = ""
		seconds = p.online_time / 60
		minutes = seconds / 60
		hours = minutes / 60
		if hours >= 1 then
			time_str = time_str .. string.format("%d hours, ", hours)
		end
		if minutes >= 1 then
			time_str = time_str .. string.format("%d minutes, ", minutes % 60)
		end
		time_str = time_str .. string.format("%d seconds", seconds % 60)

		p.print(string.format("(%.2f, %.2f) Complete after %s (%d ticks)", pos.x, pos.y, time_str, p.online_time))	
		p.print("Resources used:"..game.table_to_json(resources_used))
		dbg = 0
		done = true
		return
	else
		-- Skip initial cutscene
		if first_tick then
			debug(p, "skipping initial cutscene")
			p.exit_cutscene()
			debug(p, "setting always daylight")
			p.surface.always_day = true
			debug(p, "increasing character inventory size")
			p.character_inventory_slots_bonus = 420 --500 slots total
			first_tick = false
		end
	end

	-- Handcrafting

	if current_craft ~= nil and current_craft.started and prereqs_done(current_craft) and current_craft.done(p) then
		debug(p, string.format("(%d): %s done", event.tick, current_craft.id))
		queues.mark_done(current_craft.id)
		current_craft = queues.pop("character_craft")
	end
	
	if current_craft ~= nil and (not current_craft.started) and prereqs_done(current_craft) then
		debug(p, string.format("(%d) starting craft: %d %s", event.tick, current_craft.args.amount, current_craft.args.item))
		craft(p, current_craft.args.amount, current_craft.args.item)
		current_craft.started = true
		return
	end

	-- Research

	if current_tech ~= nil and prereqs_done(current_tech) and current_tech.done(p) then
		debug(p, string.format("(%d): %s done", event.tick, current_tech.id))
		queues.mark_done(current_tech.id)
		current_tech = queues.pop("lab")
	end
	
	if current_tech ~= nil and (not current_tech.started) and prereqs_done(current_tech) then
		debug(p, string.format("(%d) researching tech %s", event.tick, current_tech.args.tech))
		tech(p, current_tech.args.tech)
		current_tech.started = true
		return
	end

	-- Actions

	if current_action == nil then
		return
	end

	if not prereqs_done(current_action) then
		return
	end

	task = current_action.task
	args = current_action.args
	
	-- the task definitions only provide the names of resources/buildings with the assumption
	-- that it can be searched for at runtime. This is runtime, let's figure it out
	if task == "walk" then
		destination = args.location
	elseif task == "put" or task == "take" or
		task == "recipe" then
			building = buildings.get(p, args.entity)
			destination = building.location
			args.location = building.location
	elseif task == "build" then
			loc = locations[args.entity]
			if not loc then
				error("don't know where to place " .. args.entity)
			end
			destination = loc
			args.location = loc
	elseif task == "mine" then
		if args.location == nil then
			if args.resource == nil then
				building = buildings.get(p, args.entity)
				destination = building.location
				args.location = building.location
			else
				resource = resources.find(p, args.resource)
				destination = resource.position
				args.location = resource.position
			end
		else
			destination = args.location
		end
	end

	-- now try to do the task
	if task == "build" then
		cr = build(p, args.location, args.entity, args.direction or defines.direction.north)
	elseif task == "recipe" then
		cr = recipe(p, args.location, args.recipe)
	elseif task == "mine" then
		cr = mine(p, args.location)
	elseif task == "put" then
		cr = put(p, args.location, args.item, args.amount, args.inventory)
	elseif task == "take" then
		cr = take(p, args.location, args.item, args.amount, args.inventory)
	-- elseif task == "launch" then
	-- 	cr = launch(p, args.location)
	elseif task == "speed" then
		cr = speed(args.n)
	end
	
	can_reach = can_reach or cr

	-- if we can reach, we've started doing the action. Stop walking and start checking if it's done
	if can_reach then
		destination = pos
	end


	if current_action.done(p) and (task == "walk" or can_reach) then
		debug(p, string.format("(%d): %s done", event.tick, current_action.id))
		queues.mark_done(current_action.id)
		current_action = queues.pop("character_action")
		debug(p, string.format("(%d) starting action %s", event.tick, serpent.block(current_action)))
		can_reach = false
	end

	local walking = walk(destination.x - pos.x, destination.y - pos.y)
	p.walking_state = walking

end)


-- Populates the resources_used table that's dumped at the end of the TAS
script.on_event(defines.events.on_player_mined_entity, function(event)
	local resource = event.entity.name

	-- only update the resources listed in the table definition.
	-- 0 is not a falsy value for some reason so this works. Thanks, Lua
	if resources_used[resource] then
		resources_used[resource] = resources_used[resource] + 1
	end
end)
