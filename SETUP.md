## Requirements

* A copy of Factorio version `1.1.77` or later - legal purchases are preferred for reasons I shouldn't need to explain
* The Go programming language - install from https://go.dev/dl/ or your package manager of choice (`1.18` or later is required)
* goyacc - run `go install golang.org/x/tools/cmd/goyacc@latest` and make sure `$(go env GOPATH)/bin` is in your `PATH`
* `make` and `patch` - these should come by default on Linux/Mac, but isn't hard to find either if you don't have them

## Get the data

As of version `1.1.77`, there's a command line option to dump the raw data the game uses as JSON - this is the reason this version is a hard requirement. Run `$FACTORIO_INSTALL_PATH/bin/x64/factorio --data-dump` and it'll dump a rather large (~35-40MB) JSON file into the `script-output` directory. Copy it to [`data`](./data).

We also need map-specific data, and since I don't feel like reverse-engineering the storage format, we'll get it from a running instance of the game. Start a new game with your chosen map (mine is using the exchange string below) and run the following with `/c ...`. When it's done another JSON file will appear in `script-output`. Copy it to `./data` 

```lua

local d = {}

local tiles = {}

for chunk in game.player.surface.get_chunks() do
    for x = chunk.area.left_top.x,chunk.area.right_bottom.x,1 do
        tiles[x] = tiles[x] or {}
        for y = chunk.area.left_top.y,chunk.area.right_bottom.y,1 do
            local tile = game.player.surface.get_tile(x, y)
            if tile.valid then
                tiles[x][y] = tile.name
            end
        end
    end
end

d.tile = tiles

local en = {}
local entities = game.player.surface.find_entities()
local i = 1
local e = entities[i]
while e ~= nil do
    if en[e.position.x] == nil then
        en[e.position.x] = {}
    end

    local entity = { name = e.name, position = e.position, type = e.type }
    if e.type == "resource" then
        entity.amount = e.amount
    end
    if e.type == "container" then
        local inv = e.get_inventory(defines.inventory.chest)
        if inv ~= nil then
            entity.contents = inv.get_contents()
        end
    end
    en[e.position.x][e.position.y] = entity
    i = i + 1
    e = entities[i]
end
d.entity = en
game.write_file("map-data.json", game.table_to_json(d))
```

This takes about 20 seconds on my computer on a completely new map. It loops over every single tile and entity in every generated chunk so be careful using it on any map where you've generated more than the starting area

## Map exchange string

This map has a good clustering of the starting resources that's reasonably close to water. If you find a better map please let me know

```
>>>eNpjZGBk8GVgYgCCBnsQ5mBJzk/MgfAOOIAwV3J+QUFqkW5+U
SqyMGdyUWlKqm5+Jqri1LzU3ErdpMRiqGKIyRyZRfl56CawFpfk5
6GKlBSlphbDnALC3KVFiXmZpbkIvVCnMi59/z+ioUWOAYT/1zMo/
P8PwkDWA6ACEGZgbICoBIrBAGtyTmZaGgODgiMQO4EVMTBWi6xzf
1g1xZ4RokbPAcr4ABU5kAQT8YQx/BxwSqnAGCZI5hiDwWckBsTSE
qAVUFUcDggGRLIFJMnI2Pt264Lvxy7YMf5Z+fGSb1KCPaOhq8i7D
0br7ICS7CAvMMGJWTNBYCfMKwwwMx/YQ6Vu2jOePQMCb+wZWUE6R
ECEgwWQOODNzMAowAdkLegBEgoyDDCn2cGMEXFgTAODbzCfPIYxL
tuj+wMYEDYgw+VAxAkQAbYQ7jJGCNOh34HRQR4mK4lQAtRvxIDsh
hSED0/CrD2MZD+aQzAjAtkfaCIqDliigQtkYQqceMEMdw0wPC+ww
3gO8x0YmUEMkKovQDEIDyQDMwpCCziAg5uZAQGAaUP20+XvAL9/o
5o=<<<
```

