math2d = require("math2d")

-- where the machines are placed. Change this if you use a different map or don't like the layout
local locations = {
    lab = {x = -15.5, y = -20.5},
    boiler = {x = -20, y = -25.5},
    ["stone-furnace"] = {x = 11, y = 44},
    ["steel-furnace"] = {x = 11, y = 44},
    ["small-electric-pole"] = {x = -16.5, y = -22.5},
    ["offshore-pump"] = {x = -20.5, y = -27.5},
    ["steam-engine"] = {x = -16.5, y = -25.5},
}

-- TODO:
-- * electric_furnace
-- * solar-panel
-- * assembling-machine-2
-- * oil-refinery
-- * chemical-plant
-- * assembline-machine-3
-- * rocket-silo
-- * a second small-electric-pole (possibly not needed, will require thought of how to make it work)


-- the first locations we mine. Map-specific 
local resources = {
    coal = {x = 11.5, y = 45.5},
    stone = {x = 28.5, y = 57.5},
    ["iron-ore"] = {x = 12.5, y = 43.5},
    ["copper-ore"] = {x = 9.5, y = 43.5},
}


-- finds location to mine the provided resource. To avoid
-- more lag than we need search is limited to 512 tiles away from
-- the starting position
function resources.find(p, name)

    local radius, end_radius = 8, 512
 
    start = resources[name]

    res = p.surface.find_entities_filtered({
        position = start,
        radius = radius,
        name = name
    })
    while res == nil or #res == 0 do
        radius = radius * 2
        if radius > end_radius then
            error("no "..name.." found on player surface")
        end
        res = p.surface.find_entities_filtered({
            position = start,
            radius = radius,
            name = name
        })
    end
    
    sort_func = function(i, j)
        d1 = math2d.position.distance_squared(i.position, start)
        d2 = math2d.position.distance_squared(j.position, start)
        return d1 < d2
    end

    table.sort(res, sort_func)

    -- Make this the new starting point of the search
    resources[name] = res[1].position

    return {
        resource = name,
        position = res[1].position,
        amount = res[1].amount,
        entity = res[1]
    }
end


-- Where the buildings are
local buildings = {}

function buildings.build(p, name, location)
    local loc = math2d.position.ensure_xy(location)
    buildings[name] = {
        name = name,
        location = loc
    }
    return buildings.get(p, name)
end

function buildings.is_placed(p, name)
    building = buildings.get(p, name)

    return building ~= nil and building.is_placed
end

function buildings.mine(p, name)
    buildings[name] = nil
end

function buildings.get(p, name)
    local building = buildings[name]
    if not building or not building.location then
        return nil
    end

    building.entity = p.surface.find_entity(building.name, building.location)
    building.is_placed = building.entity ~= nil

    return building
end

return {
    buildings = buildings,
    locations = locations,
    resources = resources,
}