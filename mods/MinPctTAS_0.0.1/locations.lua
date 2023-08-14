math2d = require("math2d")

local DIR = defines.direction

-- where the machines are placed. Change this if you use a different map or don't like the layout
local locations = {
    lab = {x = 306.5, y = 172.5},
    boiler = {x = 300, y = 175.5, dir = DIR.east},

    ["stone-furnace"] = {x = 11, y = 45},
    ["steel-furnace"] = {x = 11, y = 45},

    ["electric-furnace"] = {x = 303.5, y = 175.5},
    
    ["small-electric-pole"] = {
        {x = 303.5, y = 172.5},
        {x = 303.5, y = 165.5}
    },
    ["offshore-pump"] = {x = 299.5, y = 177.5, dir = DIR.south},
    ["steam-engine"] = {x = 303.5, y = 175.5, dir = DIR.east},
    ["solar-panel"] = {x = 301.5, y = 171.5},
    ["oil-refinery"] = {x = 300.5, y = 166.5, dir = DIR.north},
    ["chemical-plant"] = {x = 306.5, y = 163.5, dir = DIR.west},
    ["assembling-machine-2"] = {x = 306.5, y = 175.5}, -- overlaps with steam-engine
    ["assembling-machine-3"] = {x = 306.5, y = 175.5},
    ["pumpjack"] = {x = 304.5, y = 168.5, dir = DIR.west},
    ["pipe"] = {
        -- boiler to refinery
        {x = 299.5, y = 173.5},
        {x = 299.5, y = 172.5},
        {x = 299.5, y = 171.5},
        {x = 299.5, y = 170.5},
        {x = 299.5, y = 169.5},

        -- boiler to chem plant
        {x = 300.5, y = 173.5},
        {x = 301.5, y = 173.5},
        {x = 302.5, y = 173.5},
        {x = 303.5, y = 173.5},
        {x = 304.5, y = 173.5},
        {x = 304.5, y = 172.5},
        {x = 304.5, y = 171.5},
        {x = 304.5, y = 170.5},
        {x = 305.5, y = 170.5},
        {x = 306.5, y = 170.5},
        {x = 306.5, y = 169.5},
        {x = 306.5, y = 168.5},
        {x = 306.5, y = 167.5},
        {x = 306.5, y = 166.5},
        {x = 306.5, y = 165.5},
        {x = 305.5, y = 165.5},
        {x = 304.5, y = 165.5},
        {x = 304.5, y = 164.5},

        -- pumpjack to refinery
        {x = 302.5, y = 169.5},
        {x = 301.5, y = 169.5},

        -- chem plant to assembler
        {x = 308.5, y = 163.5},
        {x = 308.5, y = 164.5},
        {x = 308.5, y = 165.5},
        {x = 308.5, y = 166.5},
        {x = 308.5, y = 167.5},
        {x = 308.5, y = 168.5},
        {x = 308.5, y = 169.5},
        {x = 308.5, y = 170.5},
        {x = 308.5, y = 171.5},
        {x = 308.5, y = 172.5},
        {x = 308.5, y = 173.5},
        {x = 308.5, y = 174.5},
        {x = 308.5, y = 175.5},

        -- TODO: refinery outputs will need to be selectively routed
        -- to the chem plant
    },
    -- TODO: will require mining something to make space near the power poles
    -- ["rocket-silo"] = {}
}

function locations.get(entity, n)
    loc = locations[entity]
    if loc.x then
        return loc
    end
    return loc[n]    
end

-- the first locations we mine. Map-specific 
local resources = {
    coal = {x = 12.5, y = 46.5},
    stone = {x = 28.5, y = 56.5},
    ["iron-ore"] = {x = 11.5, y = 43.5},
    ["copper-ore"] = {x = 8.5, y = 43.5},
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
        if radius > end_radius then
            error("no "..name.." found on player surface")
        end
        res = p.surface.find_entities_filtered({
            position = start,
            radius = radius,
            name = name
        })
        radius = radius * 2
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

local function format_n(n)
    if not n then
        return ""
    end
    return string.format("%d", n)
end

function buildings.build(p, name, location, n)
    local loc = math2d.position.ensure_xy(location)
    local b = {
        name = name,
        location = loc,
        n = n
    }
    if not n then
        buildings[name] = b
    else
        buildings[name][n] = b
    end

    return buildings.get(p, name)
end

function buildings.is_placed(p, name, n)
    building = buildings.get(p, name, n)
    return building ~= nil and building.is_placed
end

function buildings.mine(p, name, n)
    if not n then
        buildings[name] = nil
    else
        buildings[name][n] = nil
    end
end

function buildings.get(p, name, n)

    if not n then
        building = buildings[name]
    else
        building = buildings[name][n]
    end

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