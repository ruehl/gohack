# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Working agreement

The owner wants tight control over game design. Build only what is asked. If a feature turns out to be strictly required but was not decided on, ask before adding it. Examples of things deliberately not built yet: inventory UI or pickup, NPC movement, weapons, other portal types, other races/classes/NPCs. Combat (`internal/rules/combat.go`) is a placeholder until the combat system is designed.

## Commands

```sh
go run ./cmd/gohack                  # play (needs a real interactive terminal)
go vet ./... && go test ./...        # check everything
go test ./internal/game -run TestDeath -v   # a single test
go test -bench . ./internal/fov/     # line-of-sight benchmark
```

The game cannot be driven through a plain pipe. To exercise the UI headlessly, run the built binary under a pty and read the screen with a terminal emulator.

## Architecture

The core rule: **the simulation knows nothing about input or rendering**, so the tcell frontend can be swapped for Ebitengine or a browser, and the game can later run on a server. Dependency direction is `geom` ← `world` ← `entity` ← `rules`/`game` ← `ui/term`; `fov` depends on nothing.

- **Turn flow** (`internal/game/game.go`): a frontend calls `Game.Submit(actorID, Command)`. The command is applied, and only if it *consumed a turn* do NPCs act (`ai.go`) and `Turn` advance. Blocked moves (walls, bumping an NPC) and attacking empty air cost no turn. `Submit` returns `[]Event`; the same events are turned into per-player message log lines in `messages.go`, written from the viewer's perspective ("You hit…" vs "The goblin hits you…").
- **Rendering contract**: frontends only call `Game.BuildView(playerID, w, h)` and draw the returned `View` (cells with tile, visibility state and optional actor, plus HUD data and recent messages). The player is always at cell `(W/2, H/2)`; the map scrolls. The terminal-specific glyphs and colors live only in `internal/ui/term/draw.go`.
- **Entities live in `Game`, not `World`.** `World` is static map data (tiles, portals, spawn points); `entity` imports `world` for `LevelID`. Position lookup goes through `Game.occ`, which must be kept in sync (`place`, `moveTo`, `remove`). Dead NPCs are removed; a dead player stays in the world.
- **Line of sight**: `internal/fov` is symmetric shadowcasting in exact integer math. Perception is per player (`perception.go`): a `visible` bitgrid for the current level and a `seen` bitgrid per level. Remembered tiles are drawn using the tile's *current* type, not what was last seen.
- **Levels and portals**: levels are ASCII files embedded from `internal/world/levels/`. A `portal:` header line is declared in one file only; `World.link` creates both ends and validates that both are stair tiles. Map glyphs for spawns (e.g. `g`) map to NPC template IDs in `spawnGlyphs` (`loader.go`), which must exist in `entity.NPCTemplates`. Max level size is 1024×1024. Exactly one `@` player start across all levels.
- **Rules are injected and deterministic**: `rules.RNG` is an interface; `Game` takes one. Tests use a scripted `seq` RNG that returns fixed values (see `game_test.go`, `combat_test.go`), and build worlds from in-memory level text with `testing/fstest.MapFS` and `world.Load`.
- **Stats**: modifiers are `(attr-10)/2` rounded down (`entity.Mod`); max HP is `10 + STA`; AC is `10 +` armor bonuses of equipped items. NPC disposition is Hostile/Neutral/Friendly towards the player; only hostile NPCs attack, and a player attacking any NPC turns it hostile.

## Environment

Go 1.26. The devcontainer (`.devcontainer/`) uses `mcr.microsoft.com/devcontainers/go:1.26-bookworm`; check that a tag exists on MCR before changing it.
