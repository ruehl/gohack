// Package game is the simulation core. It knows nothing about input devices or
// rendering: actors submit Commands, the game applies the rules and returns
// Events, and frontends read a per-player View.
package game

import "gohack/internal/geom"

// Command is an action an entity asks to perform.
type Command interface{ isCommand() }

// MoveCmd moves one tile; walking into a closed door opens it.
type MoveCmd struct{ Dir geom.Dir }

// AttackCmd attacks whatever stands on the adjacent tile in Dir.
type AttackCmd struct{ Dir geom.Dir }

// UseStairsCmd takes the stairs on the actor's current tile.
type UseStairsCmd struct{}

func (MoveCmd) isCommand()      {}
func (AttackCmd) isCommand()    {}
func (UseStairsCmd) isCommand() {}
