## Factorio Minimum Resources TAS


Most speedruns try to complete their goal in the fastest time. Others minimize space or have other limitations for extra challenge, and they all use some really clever tricks to accomplish it. I, however, have problems, so I'm not doing any of that. I instead ask a different question: what is the absolute minimum amount of resources you can dig out of the ground and still beat the game? I saw the question going around r/factorio and the forums a few years back and it stuck with me for reasons I'm not sure I can explain even to myself. Those posts do give numerical answers but as far as I know they've never been proven with a run, probably because other people are still sane enough to know it's a dumb idea

## How-to

In progress, what I've built so far takes a simplified grammar and translates it into Lua code that can be ran with a (heavily) modified version of [this mod](https://github.com/gotyoke/Factorio-AnyPct-TAS). Follow the instructions in [SETUP.md](./SETUP.md), then
* compile with `make`, 
* run `./compile <infile> mods/MinPctTAS_0.0.1/tasks.lua`
* run `$FACTORIO_INSTALL_PATH/bin/x64/factorio --mod-directory mods`, start a new map and watch it run

## FAQ

### Definitions

- "beat the game": launch a rocket. Instead of time, we're minimizing the sum of the resources on the production tab
- "resource": anything you can mine. For vanilla, this means wood, stone, iron ore, copper ore, uranium ore, and coal. Water and crude oil are excluded primarily because they make the math far more difficult for little perceived benefit

### Does this work with mods?

Yes. I'm using Factorio's own mechanisms to dump the data it uses, so there's no practical reason it shouldn't work with your mod(s) of choice, although I will disclaim that by warning that recipe cycles are not something I'm prepared to handle yet so things will get weird if you try to use a mod that has them

### Does this work with another map?

Yes. There's nothing special about the seed I chose aside from it having a good layout. Just be sure to update `locations.lua` in the mod

### But how does it actually work?

Short answer: it's a series of scripts that takes some abstract goals and turns them into a TAS.

Long answer: I've written a mod (based on https://mods.factorio.com/mod/AnyPctTAS but heavily modified to do what's described here) to take a series of tasks and perform them in order. Instead of the original mod hardcoding the order and which tick they're performed on, this allows you to set prerequisites that must be done before a task is started, so you can have something like "craft 10 iron-gear-wheels but not before you take 20 iron-plates from the furnace." The mod considers a TAS "done" when there's no more tasks left to perform, and it'll print out how many resources were used in the process.

The tasks are split into three queues:
- `character_craft` for handcrafting
- `lab` for research
- `character_action` for everything else, like mining or putting stuff in machines

On each tick, the queues are checked in the order listed above (only one is started on any given tick). If all of the task's prerequisites are `done()`, the task is started and marked as such. If a task is already running, we check if it's `done()`, and if it is, the next task is pulled off the queue. Some of the `character_action`s also need locations, but for the purpose of easier script generation they are not hardcoded anywhere in the Go code; we only ever build one of any given building so its name is used by the mod and the "real" location is found at runtime via the logic in `locations.lua`.

`tasks.lua` can be modified directly, but an easier option is to use the Go command I've also provided. That allows you to define more abstract goals and it will do the hard work of turning them into actions to perform. The work is split into 4 phases, the last three of which are contained in `task.Optimize`:

- Phase "0": define the tasks. This is done at the point the `task.New...` functions are called. For "craft" tasks, use the recipe data to add prerequisite tasks to craft its ingredients first (or mine ore). For "tech" tasks, research every prerequisite task and craft the science packs

- Phase 1: Tasks are ran in order with a state object. "craft", "build", and "mine" all affect the inventory, and if they alter the number of crafts required then the pass will reverse and remove any "extra" crafts/mines that happened earlier. "tech" tasks for technologies that are already researched will likewise be removed along with their requirement to craft the associated science packs. When I get to that point module bonuses will also be applied on this pass

- Phase 2: Tasks are ran again with a new state object. Abstract "craft" tasks are replaced with either handcrafting or (more likely) placing ingredients into a machine, waiting, and grabbing the result. This pass also takes into account mining fuel to power machines and batching the crafts to one stack of material.

- Phase 3: Tasks are reordered so that actions can be done while waiting for something else to finish

After these optimizations are done, the tasks are converted to a lua formatted line and outputted to the new `tasks.lua`

### What modifications are made to the game?

Aside from the obvious "character is being controlled by the script," I've also made a couple other changes for convenience. They don't affect the quantity of resources required but they do make the run faster to code/execute.
* night does not exist. This is because most of the run has machines powered by one solar panel
* character inventory size is greatly increased. This is also not necessary, but not doing it would require that I implement logic to drop things on the ground and pick them up and that seems like more trouble that it's worth

## LICENSE

All code in this repository is licensed under the terms of the GNU General Public License, version 3 (or at your choice, any later version). See the [LICENSE](./LICENSE) file for more details