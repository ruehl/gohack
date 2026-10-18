# GoHack

A 2D dungeon crawler with RPG elements, written in Go. It currently renders in the terminal.

## Requirements

- Go 1.26 or newer
- A terminal with at least about 60×20 cells and 256-color support

Alternatively, open the folder in VS Code and choose **Dev Containers: Reopen in Container**. The devcontainer provides Go 1.26, the Go extension, and the Claude Code CLI.

If you already have the container open, run **Dev Containers: Rebuild Container** to install Claude Code. Then run `claude` in the container's terminal and follow the sign-in prompts. You can verify the installation with `claude --version`.

## Build and run

```sh
# run directly
go run ./cmd/gohack

# or build a binary
go build -o gohack ./cmd/gohack
./gohack
```

Run the game in a real terminal. It needs an interactive TTY, so it will not work when piped or inside most IDE output panes.

## Controls

| Key                  | Action                                      |
| -------------------- | ------------------------------------------- |
| `W` `A` `S` `D`      | Move (also arrow keys)                      |
| `F`, then a direction | Attack the adjacent tile in that direction  |
| `E`                  | Take the stairs you are standing on         |
| `I`                  | Open the gear screen (see below)            |
| `Q` / `Esc`          | Quit (`Esc` first cancels a pending attack) |

Walking into a closed door opens it. Walking into an NPC does not attack it.

### Gear screen

`W`/`S` or the arrows move the cursor, `Enter` picks a slot and then an item for it (`(nothing)` takes the slot's item off), `Esc` steps back and closes the screen, `I` closes it at once. Each change takes a turn. Slots: head, neck, two rings, body, hands, legs, feet, main hand and off hand. A two-handed weapon needs both hands, and a second one-handed weapon in the off hand needs the `dual_wielding` skill; a dual-wielder attacks once with each weapon.

## Test

```sh
go vet ./...
go test ./...
go test -bench . ./internal/fov/   # line-of-sight benchmark
```

## Project layout

```
cmd/gohack/          entry point
internal/geom/       points and directions
internal/world/      tiles, levels, portals, level loader
internal/world/levels/  ASCII level files (embedded in the binary)
internal/entity/     attributes, race, class, items, inventory, NPC templates
internal/rules/      combat rules and RNG
internal/fov/        symmetric shadowcasting line of sight
internal/game/       simulation: commands, events, turns, AI, per-player view
internal/ui/term/    terminal frontend (tcell)
```

The simulation in `internal/game` knows nothing about input or rendering. Frontends send `Command`s and draw a `View`, so the terminal renderer can be replaced later.

## Level files

Levels are text files in `internal/world/levels/`. A header, a `---` line, then the map:

```
id: level1
name: Goblin Cellar
portal: 33,3 -> level2:3,1
---
  ###########
  #....@....+
```

| Glyph   | Meaning                                  |
| ------- | ---------------------------------------- |
| (space) | void                                     |
| `#`     | wall                                     |
| `.`     | floor                                    |
| `+`     | closed door                              |
| `'`     | open door                                |
| `>`     | stairs                                   |
| `@`     | player start (exactly one across levels) |
| `g`     | goblin spawn                             |

A `portal: x,y -> level:x,y` line links a stair tile to a stair tile in another level. The link works in both directions, so declare it in only one of the two files. Levels can be at most 1024×1024 tiles.
