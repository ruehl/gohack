package game

import (
	"gohack/internal/entity"
	"gohack/internal/geom"
	"gohack/internal/world"
)

// EventKind says what happened.
type EventKind uint8

const (
	EvMoved EventKind = iota
	EvDoorOpened
	EvBlocked // actor tried to walk into Target
	EvAttack
	EvDied
	EvLevelChanged
	EvMessage // Text is shown to Actor only
)

// Event describes something that happened during a turn.
type Event struct {
	Kind                  EventKind
	Actor, Target         entity.EntityID
	ActorName, TargetName string
	Level                 world.LevelID
	Pos                   geom.Point
	Hit                   bool
	Damage                int
	Text                  string
}
