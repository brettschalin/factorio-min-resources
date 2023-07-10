## Factorio Minimum Resources TAS


Most speedruns try to complete their goal in the fastest time. Others minimize space or have other limitations for extra challenge, and they all use some really clever tricks to accomplish it. I, however, have problems, so I'm not doing any of that. I instead ask a different question: what is the absolute minimum amount of resources you can dig out of the ground and still beat the game? I saw the question going around r/factorio and the forums a few years back and it stuck with me for reasons I'm not sure I can explain even to myself. Those posts do give numerical answers but as far as I know they've never been proven with a run, probably because other people are still sane enough to know it's a dumb idea

## How-to

* create a `tas.TAS` object
* define various `tas.Task`s, `task.Prerequisites().Add()` if needed
* `tas.Add(tasks)` and check for errors
* `tas.Export(outFile)` to write the Lua code, save it to `mods/MinPctTAS_0.0.1/tasks.lua`
* `make start_factorio` and create a new map with the string in `SETUP.md`

## FAQ

### Definitions

- "beat the game": launch a rocket. Instead of time, we're minimizing the sum of the resources on the production tab
- "resource": anything you can mine. For vanilla, this means wood, stone, iron ore, copper ore, uranium ore, and coal. Water and crude oil are excluded primarily because they make the math far more difficult for little perceived benefit

### Does this work with mods?

Yes. I'm using Factorio's own mechanisms to dump the data it uses, so there's no practical reason it shouldn't work with your mod(s) of choice, although I will disclaim that by warning that recipe cycles are not something I'm prepared to handle yet so things will get weird if you try to use a mod that has them

### Does this work with another map?

Yes. There's nothing special about the seed I chose aside from it having a good layout. Just be sure to update `locations.lua` in the mod

### But how does it actually work?

Short answer: it's mod and Go command that takes some abstract tasks and turns them into a TAS.

Long answer: I've written a mod (based on https://mods.factorio.com/mod/AnyPctTAS but heavily modified to do what's described here) to take a series of tasks and perform them in order. Instead of the original mod hardcoding the order and which tick they're performed on, this allows you to set prerequisites that must be done before a task is started, so you can have something like "craft 10 iron-gear-wheels but not before you take 20 iron-plates from the furnace." The mod considers a TAS "done" when there's no more tasks left to perform, and it'll print out how many resources were used in the process.

The tasks are split into three queues:
- `character_craft` for handcrafting
- `lab` for research
- `character_action` for everything else, like mining or putting stuff in machines

On each tick, the queues are checked in the order listed above (only one is started on any given tick). If all of the task's prerequisites are `done()`, the task is started and marked as such. If a task is already running, we check if it's `done()`, and if it is, the next task is pulled off the queue. Some of the `character_action`s also need locations, but for the purpose of easier script generation they are not hardcoded anywhere in the Go code; we only ever build one of any given building so its name is used by the mod and the "real" location is found at runtime via the logic in `locations.lua`.

`tasks.lua` can be modified directly, but an easier option is to use the Go command I've also provided. That allows you to define goals more abstractly (and handles defining prerequisite actions in an automated manner). It works in a couple of phases

- define some Tasks using the provided functions (eg, `tas.Craft` or `tas.Tech`)
- create the dependencies using `task.Prerequisites().Add`
- create a `tas.TAS` and `Add` the tasks you just created. The returned error will tell you if the run is valid
    * this works by (a) checking if prerequisites appear in the right order, and (b) performing the state transformations provided by each task and verifying that they're possible in-game
- call `tas.Export` to write the generated Lua code, which should replace `mods/MinPctTAS_0.0.1/tasks.lua`


### What modifications are made to the game?

Aside from the obvious "character is being controlled by the script," I've also made a couple other changes for convenience. They don't affect the quantity of resources required but they do make the run faster to code/execute.
* night does not exist. This is because most of the run has machines powered by one solar panel
* character inventory size is greatly increased. This is also not necessary, but not doing it would require that I implement logic to drop things on the ground and pick them up and that seems like more trouble that it's worth

## LICENSE

All code in this repository is licensed under the terms of the GNU General Public License, version 3 (or at your choice, any later version). See the [LICENSE](./LICENSE) file for more details