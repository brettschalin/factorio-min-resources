## Factorio Minimum Resources TAS


Most speedruns try to complete their goal in the fastest time. Others minimize space or have other limitations for extra challenge, and they all use some really clever tricks to accomplish it. I, however, have problems, so I'm not doing any of that. I instead ask a different question: what is the absolute minimum amount of resources you can dig out of the ground and still beat the game? I saw the question going around r/factorio and the forums a few years back and it stuck with me for reasons I'm not sure I can explain even to myself. Those posts do give numerical answers but as far as I know they've never been proven with a run, probably because other people are still sane enough to know it's a dumb idea

## How-to

In progress, what I've built so far takes a simplified grammar and translates it into Lua code that can be ran with [this mod](https://github.com/gotyoke/Factorio-AnyPct-TAS). Follow the instructions in [SETUP.md](./SETUP.md), then
* compile with `make`, 
* run `fmin <infile> <outfile>`
* copy the result to `mods/AnyPctTAS_0.2.2/control.lua`
* run `$FACTORIO_INSTALL_PATH/bin/x64/factorio --mod-directory mods`, start a new map with your chosen exchange string and watch it run

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

Long answer: Details still TBD, but what I do have built is a language that compiles down to a series of tasks that can be ran with a modified https://github.com/gotyoke/Factorio-AnyPct-TAS. The exchange string for the map I'm using is in [SETUP.md](./SETUP.md#map-exchange-string), and the information found in the rest of the file should give idea of how everything pieces together

## LICENSE

All code in this repository is licensed under the terms of the GNU General Public License, version 3 (or at your choice, any later version). See the [LICENSE](./LICENSE) file for more details