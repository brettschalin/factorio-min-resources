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

Yes. There's nothing special about the seed I chose aside from it having a good layout.

### But how does it actually work?

Short answer: it's a series of scripts that takes some abstract goals and turns them into a TAS.

Long answer: A lot of this is still TBD, so take it with a grain of salt. What I have now is a heavily, heavily modified version of a TAS mod, which rather than take a list of "what to do every tick," the things to do are structured as a DAG. The task list adds things like "mine 10 iron-ore" or "craft 2 electronic-circuits," then the mod will do each in order (and take prerequisite actions into account) until there's nothing left to do.

The task lists also don't hardcode any locations. This is by design. When every `character_action` task runs, it first looks up where to go (this logic is in `locations.lua`) and walks toward the target location if necessary, then performs the action, which in practice means the TAS is extremely resilient to changing the map seed. The only locations predetermined are for (1) where to construct the machines and (2) where to start searching for mining a specific resource; both are in `locations.lua` and can be changed easily for different maps.

In the future I have planned another command (written in Go) that will take a more abstract / human readable script and transform it into something the TAS can understand. You can find an early version in the `cmd/compile` and `tasscript` folders


## LICENSE

All code in this repository is licensed under the terms of the GNU General Public License, version 3 (or at your choice, any later version). See the [LICENSE](./LICENSE) file for more details