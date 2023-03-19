# TAS-Script

A simple language that will ease the generation of TAS Lua files. The code is very heavily based on what you can find at `$GOPATH/src/golang.org/x/tools/cmd/goyacc/testdata/expr`, as that's pretty much all of the documentation goyacc has. If you're familiar with the original yacc (or Bison), you probably are cringing at the absolute state of some of the hacks.

## Language Definition

Tas-Script is made up of a series of commands separated by newlines. The allowed commands are

* `START x y`:
    Where to start the TAS (defaults to `0 0`). If present, it must be the first command
* `HALT`:
    Optional, must be the last command if present
* `BUILD location item direction`:
    Constructs a building at the location, facing `direction`
* `CRAFT recipe n`:
    Handcraft `n` `recipe`s
* `IDLE n`:
    Wait for `n` ticks
* `LAUNCH location`:
    Launches a rocket from the silo at `location`
* `LOCATION name x y`:
    Defines a location later commands can use instead of always specifying XY coordinates
* `LOOP n`:
    repeat every statement until `ENDLOOP` `n` times
* `MINE location n`:
    Mines `n` resources from `location`. Make sure this is on a resource patch
* `PUT location item n inventory`:
    inserts `n` `item`s into the `inventory` slots of the building at `location` 
* `RECIPE location recipe`:
    Sets crafting recipe on the building at `location`
* `ROTATE location direction`:
    Rotates the building at `location` to face `direction`
* `SPEED f`:
    Sets the game speed. Factorio limitation is this must be at least `0.1`
* `TAKE location item n inventory`:
    Same as `PUT` but it takes from the inventory instead
* `TECH tech`:
    Start researching `tech`
* `WALK location`:
    Start walking to `location`. The character will continue doing other things when they're within reach

An example of all of this can be seen in the `testscript` file
___

Notes for the above commands:
- `direction` can be any of `NORTH`, `SOUTH`, `EAST`, or `WEST` (enforced by the parser, the rest are enforced by Lua or the game's logic)
- `location` can either be a name defined by a previous `LOCATION` command or a literal `x y` coordinate
- `item` must be a valid item name
- `x`, `y`, and `f` can be float values (like `0.5`), but `n` must be a positive integer
- `recipe` must be unlocked and available for the selected machine to craft
- `tech` must be unlocked and not already researched
- `inventory` must be one of the types defined [here](https://lua-api.factorio.com/1.1.77/defines.html#defines.inventory), and available for the selected building


***COMING SOON*** TM:

- Support for `#comments`
- Documentation: how to use `goyacc` - the learning curve is pretty close to vertical
